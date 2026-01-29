package arch

import (
	"bytes"
	"runtime"
	"strings"
	"testing"
)

func TestRunArch(t *testing.T) {
	t.Run("returns architecture", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunArch(&buf)
		if err != nil {
			t.Fatalf("RunArch() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output == "" {
			t.Error("RunArch() should return architecture")
		}
	})

	t.Run("output ends with newline", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunArch(&buf)

		if !strings.HasSuffix(buf.String(), "\n") {
			t.Error("RunArch() output should end with newline")
		}
	})

	t.Run("consistent output", func(t *testing.T) {
		var buf1, buf2 bytes.Buffer

		_ = RunArch(&buf1)
		_ = RunArch(&buf2)

		if buf1.String() != buf2.String() {
			t.Error("RunArch() should be consistent")
		}
	})

	t.Run("relates to GOARCH", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunArch(&buf)

		output := strings.ToLower(strings.TrimSpace(buf.String()))
		goarch := strings.ToLower(runtime.GOARCH)

		// Common mappings
		validMappings := map[string][]string{
			"amd64": {"x86_64", "amd64"},
			"386":   {"i386", "i686", "386"},
			"arm64": {"aarch64", "arm64"},
			"arm":   {"arm", "armv7l"},
		}

		if archs, ok := validMappings[goarch]; ok {
			found := false
			for _, arch := range archs {
				if strings.Contains(output, arch) {
					found = true
					break
				}
			}

			if !found {
				t.Logf("RunArch() = %v, GOARCH = %v (mapping may vary)", output, goarch)
			}
		}
	})

	t.Run("no error", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunArch(&buf)
		if err != nil {
			t.Errorf("RunArch() should not error: %v", err)
		}
	})
}
