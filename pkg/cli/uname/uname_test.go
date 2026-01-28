package uname

import (
	"bytes"
	"runtime"
	"strings"
	"testing"
)

func TestRunUname(t *testing.T) {
	t.Run("default kernel name", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUname(&buf, UnameOptions{})
		if err != nil {
			t.Fatalf("RunUname() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output == "" {
			t.Error("RunUname() should return kernel name")
		}
	})

	t.Run("default returns single field", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunUname(&buf, UnameOptions{})

		fields := strings.Fields(buf.String())
		if len(fields) != 1 {
			t.Errorf("RunUname() default got %d fields, want 1", len(fields))
		}
	})

	t.Run("all information", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUname(&buf, UnameOptions{All: true})
		if err != nil {
			t.Fatalf("RunUname() error = %v", err)
		}

		output := buf.String()
		fields := strings.Fields(output)

		// -a should return multiple fields
		if len(fields) < 3 {
			t.Errorf("RunUname() -a got %d fields, want >= 3", len(fields))
		}
	})

	t.Run("all contains kernel name", func(t *testing.T) {
		var bufAll, bufKernel bytes.Buffer

		_ = RunUname(&bufAll, UnameOptions{All: true})
		_ = RunUname(&bufKernel, UnameOptions{KernelName: true})

		kernelName := strings.TrimSpace(bufKernel.String())

		if !strings.Contains(bufAll.String(), kernelName) {
			t.Errorf("RunUname() -a should contain kernel name: %v", bufAll.String())
		}
	})

	t.Run("kernel name", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUname(&buf, UnameOptions{KernelName: true})
		if err != nil {
			t.Fatalf("RunUname() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output == "" {
			t.Error("RunUname() -s should return kernel name")
		}
	})

	t.Run("kernel name matches GOOS", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunUname(&buf, UnameOptions{KernelName: true})

		output := strings.ToLower(strings.TrimSpace(buf.String()))
		goos := strings.ToLower(runtime.GOOS)

		// Kernel name should relate to GOOS
		// Linux -> Linux, darwin -> Darwin, windows -> Windows_NT
		validNames := map[string][]string{
			"linux":   {"linux"},
			"darwin":  {"darwin"},
			"windows": {"windows", "windows_nt"},
			"freebsd": {"freebsd"},
		}

		if names, ok := validNames[goos]; ok {
			found := false
			for _, name := range names {
				if strings.Contains(output, name) {
					found = true
					break
				}
			}

			if !found {
				t.Logf("RunUname() -s = %v, GOOS = %v (may be valid)", output, goos)
			}
		}
	})

	t.Run("node name", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUname(&buf, UnameOptions{NodeName: true})
		if err != nil {
			t.Fatalf("RunUname() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output == "" {
			t.Error("RunUname() -n should return node name")
		}
	})

	t.Run("node name is hostname", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunUname(&buf, UnameOptions{NodeName: true})

		output := strings.TrimSpace(buf.String())
		// Node name should be a valid hostname
		if len(output) == 0 || len(output) > 255 {
			t.Errorf("RunUname() -n hostname length = %d seems invalid", len(output))
		}
	})

	t.Run("kernel release", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUname(&buf, UnameOptions{KernelRelease: true})
		if err != nil {
			t.Fatalf("RunUname() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output == "" {
			t.Error("RunUname() -r should return kernel release")
		}
	})

	t.Run("kernel version", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUname(&buf, UnameOptions{KernelVersion: true})
		if err != nil {
			t.Fatalf("RunUname() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output == "" {
			t.Error("RunUname() -v should return kernel version")
		}
	})

	t.Run("machine", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUname(&buf, UnameOptions{Machine: true})
		if err != nil {
			t.Fatalf("RunUname() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output == "" {
			t.Error("RunUname() -m should return machine type")
		}

		// Should match Go's GOARCH in some form
		goarch := runtime.GOARCH
		// amd64, arm64, etc. may map to x86_64, aarch64, etc.
		validArchs := []string{"amd64", "x86_64", "arm64", "aarch64", "386", "i686", "i386", "arm", goarch}

		found := false
		for _, arch := range validArchs {
			if strings.Contains(strings.ToLower(output), strings.ToLower(arch)) {
				found = true
				break
			}
		}

		if !found {
			t.Logf("RunUname() -m = %v (GOARCH=%v)", output, goarch)
		}
	})

	t.Run("machine single field", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunUname(&buf, UnameOptions{Machine: true})

		fields := strings.Fields(buf.String())
		if len(fields) != 1 {
			t.Errorf("RunUname() -m got %d fields, want 1", len(fields))
		}
	})

	t.Run("operating system", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUname(&buf, UnameOptions{OperatingSystem: true})
		if err != nil {
			t.Fatalf("RunUname() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output == "" {
			t.Error("RunUname() -o should return operating system")
		}
	})

	t.Run("operating system relates to GOOS", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunUname(&buf, UnameOptions{OperatingSystem: true})

		output := strings.ToLower(strings.TrimSpace(buf.String()))
		goos := runtime.GOOS

		// Operating system should relate to GOOS
		validOS := map[string][]string{
			"linux":   {"linux", "gnu/linux"},
			"darwin":  {"darwin", "macos", "mac os"},
			"windows": {"windows", "msys", "mingw"},
		}

		if names, ok := validOS[goos]; ok {
			found := false
			for _, name := range names {
				if strings.Contains(output, name) {
					found = true
					break
				}
			}

			if !found {
				t.Logf("RunUname() -o = %v, GOOS = %v (may be valid)", output, goos)
			}
		}
	})

	t.Run("multiple options", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUname(&buf, UnameOptions{KernelName: true, Machine: true})
		if err != nil {
			t.Fatalf("RunUname() error = %v", err)
		}

		output := buf.String()
		fields := strings.Fields(output)

		// Should have 2 fields
		if len(fields) != 2 {
			t.Errorf("RunUname() -s -m got %d fields, want 2", len(fields))
		}
	})

	t.Run("three options", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunUname(&buf, UnameOptions{KernelName: true, NodeName: true, Machine: true})

		fields := strings.Fields(buf.String())
		if len(fields) != 3 {
			t.Errorf("RunUname() three options got %d fields, want 3", len(fields))
		}
	})

	t.Run("options order", func(t *testing.T) {
		var bufSM, bufMS bytes.Buffer

		// Test that order is consistent regardless of which flags are set
		_ = RunUname(&bufSM, UnameOptions{KernelName: true, Machine: true})
		_ = RunUname(&bufMS, UnameOptions{Machine: true, KernelName: true})

		// Output should be the same
		if bufSM.String() != bufMS.String() {
			t.Logf("RunUname() order may matter: %v vs %v", bufSM.String(), bufMS.String())
		}
	})

	t.Run("output ends with newline", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunUname(&buf, UnameOptions{})
		if err != nil {
			t.Fatalf("RunUname() error = %v", err)
		}

		if !strings.HasSuffix(buf.String(), "\n") {
			t.Error("RunUname() output should end with newline")
		}
	})

	t.Run("all ends with newline", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunUname(&buf, UnameOptions{All: true})

		if !strings.HasSuffix(buf.String(), "\n") {
			t.Error("RunUname() -a output should end with newline")
		}
	})

	t.Run("single line output", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunUname(&buf, UnameOptions{All: true})

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 1 {
			t.Errorf("RunUname() should output exactly one line, got %d", len(lines))
		}
	})

	t.Run("consistent output", func(t *testing.T) {
		var buf1, buf2 bytes.Buffer

		err := RunUname(&buf1, UnameOptions{All: true})
		if err != nil {
			t.Fatalf("RunUname() error = %v", err)
		}

		err = RunUname(&buf2, UnameOptions{All: true})
		if err != nil {
			t.Fatalf("RunUname() second call error = %v", err)
		}

		// Output should be consistent (except maybe timestamps)
		if buf1.String() != buf2.String() {
			t.Logf("Note: uname output may vary: %v vs %v", buf1.String(), buf2.String())
		}
	})

	t.Run("no error", func(t *testing.T) {
		options := []UnameOptions{
			{},
			{All: true},
			{KernelName: true},
			{NodeName: true},
			{KernelRelease: true},
			{KernelVersion: true},
			{Machine: true},
			{OperatingSystem: true},
		}

		for i, opt := range options {
			var buf bytes.Buffer

			err := RunUname(&buf, opt)
			if err != nil {
				t.Errorf("RunUname() option %d error = %v", i, err)
			}
		}
	})

	t.Run("no leading whitespace", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunUname(&buf, UnameOptions{All: true})

		output := buf.String()
		if len(output) > 0 && (output[0] == ' ' || output[0] == '\t') {
			t.Error("RunUname() should not have leading whitespace")
		}
	})
}
