package dd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dd_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("basic copy", func(t *testing.T) {
		src := filepath.Join(tmpDir, "input.txt")
		dst := filepath.Join(tmpDir, "output.txt")
		content := []byte("hello world")
		_ = os.WriteFile(src, content, 0644)

		var buf bytes.Buffer

		err := RunDd(&buf, DdOptions{
			InputFile:    src,
			OutputFile:   dst,
			Status:       "none",
			StatusWriter: &bytes.Buffer{},
		})
		if err != nil {
			t.Fatalf("RunDd() error = %v", err)
		}

		data, _ := os.ReadFile(dst)
		if string(data) != string(content) {
			t.Errorf("RunDd() output = %q, want %q", data, content)
		}
	})

	t.Run("with block size", func(t *testing.T) {
		src := filepath.Join(tmpDir, "bs_input.txt")
		dst := filepath.Join(tmpDir, "bs_output.txt")
		content := []byte("abcdefghij")
		_ = os.WriteFile(src, content, 0644)

		var buf bytes.Buffer

		err := RunDd(&buf, DdOptions{
			InputFile:    src,
			OutputFile:   dst,
			BlockSize:    5,
			Status:       "none",
			StatusWriter: &bytes.Buffer{},
		})
		if err != nil {
			t.Fatalf("RunDd() error = %v", err)
		}

		data, _ := os.ReadFile(dst)
		if string(data) != string(content) {
			t.Errorf("RunDd() output = %q, want %q", data, content)
		}
	})

	t.Run("count blocks", func(t *testing.T) {
		src := filepath.Join(tmpDir, "count_input.txt")
		dst := filepath.Join(tmpDir, "count_output.txt")
		content := []byte("aaaabbbbcccc")
		_ = os.WriteFile(src, content, 0644)

		var buf bytes.Buffer

		err := RunDd(&buf, DdOptions{
			InputFile:    src,
			OutputFile:   dst,
			BlockSize:    4,
			Count:        2, // Only copy 2 blocks (8 bytes)
			Status:       "none",
			StatusWriter: &bytes.Buffer{},
		})
		if err != nil {
			t.Fatalf("RunDd() error = %v", err)
		}

		data, _ := os.ReadFile(dst)
		if string(data) != "aaaabbbb" {
			t.Errorf("RunDd() count output = %q, want 'aaaabbbb'", data)
		}
	})

	t.Run("skip blocks", func(t *testing.T) {
		src := filepath.Join(tmpDir, "skip_input.txt")
		dst := filepath.Join(tmpDir, "skip_output.txt")
		content := []byte("aaaabbbbcccc")
		_ = os.WriteFile(src, content, 0644)

		var buf bytes.Buffer

		err := RunDd(&buf, DdOptions{
			InputFile:    src,
			OutputFile:   dst,
			BlockSize:    4,
			Skip:         1, // Skip first block
			Status:       "none",
			StatusWriter: &bytes.Buffer{},
		})
		if err != nil {
			t.Fatalf("RunDd() error = %v", err)
		}

		data, _ := os.ReadFile(dst)
		if string(data) != "bbbbcccc" {
			t.Errorf("RunDd() skip output = %q, want 'bbbbcccc'", data)
		}
	})

	t.Run("conv lcase", func(t *testing.T) {
		src := filepath.Join(tmpDir, "lcase_input.txt")
		dst := filepath.Join(tmpDir, "lcase_output.txt")
		content := []byte("HELLO WORLD")
		_ = os.WriteFile(src, content, 0644)

		var buf bytes.Buffer

		err := RunDd(&buf, DdOptions{
			InputFile:    src,
			OutputFile:   dst,
			Conv:         "lcase",
			Status:       "none",
			StatusWriter: &bytes.Buffer{},
		})
		if err != nil {
			t.Fatalf("RunDd() error = %v", err)
		}

		data, _ := os.ReadFile(dst)
		if string(data) != "hello world" {
			t.Errorf("RunDd() lcase output = %q, want 'hello world'", data)
		}
	})

	t.Run("conv ucase", func(t *testing.T) {
		src := filepath.Join(tmpDir, "ucase_input.txt")
		dst := filepath.Join(tmpDir, "ucase_output.txt")
		content := []byte("hello world")
		_ = os.WriteFile(src, content, 0644)

		var buf bytes.Buffer

		err := RunDd(&buf, DdOptions{
			InputFile:    src,
			OutputFile:   dst,
			Conv:         "ucase",
			Status:       "none",
			StatusWriter: &bytes.Buffer{},
		})
		if err != nil {
			t.Fatalf("RunDd() error = %v", err)
		}

		data, _ := os.ReadFile(dst)
		if string(data) != "HELLO WORLD" {
			t.Errorf("RunDd() ucase output = %q, want 'HELLO WORLD'", data)
		}
	})

	t.Run("to stdout", func(t *testing.T) {
		src := filepath.Join(tmpDir, "stdout_input.txt")
		content := []byte("output to stdout")
		_ = os.WriteFile(src, content, 0644)

		var buf bytes.Buffer

		err := RunDd(&buf, DdOptions{
			InputFile:    src,
			Status:       "none",
			StatusWriter: &bytes.Buffer{},
		})
		if err != nil {
			t.Fatalf("RunDd() error = %v", err)
		}

		if buf.String() != string(content) {
			t.Errorf("RunDd() stdout = %q, want %q", buf.String(), content)
		}
	})

	t.Run("status output", func(t *testing.T) {
		src := filepath.Join(tmpDir, "status_input.txt")
		dst := filepath.Join(tmpDir, "status_output.txt")
		content := []byte("test content")
		_ = os.WriteFile(src, content, 0644)

		var buf, statusBuf bytes.Buffer

		err := RunDd(&buf, DdOptions{
			InputFile:    src,
			OutputFile:   dst,
			Status:       "noxfer", // Use noxfer to avoid transfer rate calculation that can hang
			StatusWriter: &statusBuf,
		})
		if err != nil {
			t.Fatalf("RunDd() error = %v", err)
		}

		if !strings.Contains(statusBuf.String(), "records") {
			t.Errorf("RunDd() status should contain 'records': %s", statusBuf.String())
		}
	})

	t.Run("nonexistent input", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDd(&buf, DdOptions{
			InputFile:    "/nonexistent/file.txt",
			Status:       "none",
			StatusWriter: &bytes.Buffer{},
		})
		if err == nil {
			t.Error("RunDd() expected error for nonexistent input")
		}
	})
}

func TestParseDdSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		hasErr   bool
	}{
		{"512", 512, false},
		{"1K", 1024, false},
		{"1k", 1024, false},
		{"1KB", 1024, false},
		{"1M", 1024 * 1024, false},
		{"1G", 1024 * 1024 * 1024, false},
		{"2K", 2048, false},
		{"", 0, true},
		{"abc", 0, true},
		{"1X", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseDdSize(tt.input)

			if tt.hasErr {
				if err == nil {
					t.Errorf("ParseDdSize(%q) expected error", tt.input)
				}

				return
			}

			if err != nil {
				t.Fatalf("ParseDdSize(%q) error = %v", tt.input, err)
			}

			if got != tt.expected {
				t.Errorf("ParseDdSize(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestApplyDdConversions(t *testing.T) {
	t.Run("lcase", func(t *testing.T) {
		data := []byte("HELLO")

		result := applyDdConversions(data, map[string]bool{"lcase": true})
		if string(result) != "hello" {
			t.Errorf("applyDdConversions lcase = %q, want 'hello'", result)
		}
	})

	t.Run("ucase", func(t *testing.T) {
		data := []byte("hello")

		result := applyDdConversions(data, map[string]bool{"ucase": true})
		if string(result) != "HELLO" {
			t.Errorf("applyDdConversions ucase = %q, want 'HELLO'", result)
		}
	})

	t.Run("swab", func(t *testing.T) {
		data := []byte("abcd")

		result := applyDdConversions(data, map[string]bool{"swab": true})
		if string(result) != "badc" {
			t.Errorf("applyDdConversions swab = %q, want 'badc'", result)
		}
	})
}

func TestFormatDdBytes(t *testing.T) {
	tests := []struct {
		bytes    float64
		expected string
	}{
		{100, "100 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := formatDdBytes(tt.bytes)
			if got != tt.expected {
				t.Errorf("formatDdBytes(%f) = %q, want %q", tt.bytes, got, tt.expected)
			}
		})
	}
}
