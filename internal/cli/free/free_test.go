package free

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunFree(t *testing.T) {
	t.Run("default output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunFree(&buf, FreeOptions{})
		if err != nil {
			t.Fatalf("RunFree() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Mem:") {
			t.Errorf("RunFree() should contain 'Mem:': %s", output)
		}

		if !strings.Contains(output, "Swap:") {
			t.Errorf("RunFree() should contain 'Swap:': %s", output)
		}
	})

	t.Run("human readable", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunFree(&buf, FreeOptions{Human: true})
		if err != nil {
			t.Fatalf("RunFree() error = %v", err)
		}

		output := buf.String()
		// Human readable should have units like K, M, G
		if !strings.Contains(output, "i") { // Ki, Mi, Gi
			t.Logf("RunFree() human readable = %s", output)
		}
	})

	t.Run("bytes mode", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunFree(&buf, FreeOptions{Bytes: true})
		if err != nil {
			t.Fatalf("RunFree() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Mem:") {
			t.Errorf("RunFree() bytes mode should work: %s", output)
		}
	})

	t.Run("mebibytes mode", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunFree(&buf, FreeOptions{Mebibytes: true})
		if err != nil {
			t.Fatalf("RunFree() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Mem:") {
			t.Errorf("RunFree() mebibytes mode should work: %s", output)
		}
	})

	t.Run("gibibytes mode", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunFree(&buf, FreeOptions{Gibibytes: true})
		if err != nil {
			t.Fatalf("RunFree() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Mem:") {
			t.Errorf("RunFree() gibibytes mode should work: %s", output)
		}
	})

	t.Run("with total", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunFree(&buf, FreeOptions{Total: true})
		if err != nil {
			t.Fatalf("RunFree() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Total:") {
			t.Errorf("RunFree() with total should contain 'Total:': %s", output)
		}
	})
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected string
	}{
		{0, "0B"},
		{100, "100B"},
		{1024, "1.0Ki"},
		{1536, "1.5Ki"},
		{1024 * 1024, "1.0Mi"},
		{1024 * 1024 * 1024, "1.0Gi"},
		{1024 * 1024 * 1024 * 1024, "1.0Ti"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestGetMemInfo(t *testing.T) {
	info, err := GetMemInfo()
	if err != nil {
		t.Fatalf("GetMemInfo() error = %v", err)
	}

	// Basic sanity checks
	if info.MemTotal == 0 {
		t.Error("GetMemInfo() MemTotal should not be 0")
	}

	if info.MemFree > info.MemTotal {
		t.Error("GetMemInfo() MemFree should not exceed MemTotal")
	}
}
