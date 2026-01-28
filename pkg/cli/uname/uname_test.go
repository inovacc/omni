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
		validArchs := []string{"amd64", "x86_64", "arm64", "aarch64", "386", "i686", "arm", goarch}

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
}
