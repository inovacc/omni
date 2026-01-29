package cmp

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCmp(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cmp_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("identical files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "same1.txt")
		file2 := filepath.Join(tmpDir, "same2.txt")

		_ = os.WriteFile(file1, []byte("hello world"), 0644)
		_ = os.WriteFile(file2, []byte("hello world"), 0644)

		var buf bytes.Buffer

		result, err := RunCmp(&buf, []string{file1, file2}, CmpOptions{})
		if err != nil {
			t.Fatalf("RunCmp() error = %v", err)
		}

		if result != CmpEqual {
			t.Errorf("RunCmp() = %v, want CmpEqual", result)
		}
	})

	t.Run("different files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "diff1.txt")
		file2 := filepath.Join(tmpDir, "diff2.txt")

		_ = os.WriteFile(file1, []byte("hello world"), 0644)
		_ = os.WriteFile(file2, []byte("hello World"), 0644)

		var buf bytes.Buffer

		result, err := RunCmp(&buf, []string{file1, file2}, CmpOptions{})
		if err != nil {
			t.Fatalf("RunCmp() error = %v", err)
		}

		if result != CmpDiffer {
			t.Errorf("RunCmp() = %v, want CmpDiffer", result)
		}

		if !strings.Contains(buf.String(), "differ") {
			t.Errorf("RunCmp() output should indicate difference")
		}
	})

	t.Run("silent mode", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "silent1.txt")
		file2 := filepath.Join(tmpDir, "silent2.txt")

		_ = os.WriteFile(file1, []byte("aaa"), 0644)
		_ = os.WriteFile(file2, []byte("bbb"), 0644)

		var buf bytes.Buffer

		result, _ := RunCmp(&buf, []string{file1, file2}, CmpOptions{Silent: true})

		if result != CmpDiffer {
			t.Errorf("RunCmp() = %v, want CmpDiffer", result)
		}

		if buf.Len() != 0 {
			t.Errorf("RunCmp() silent mode should not produce output")
		}
	})

	t.Run("verbose mode", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "verbose1.txt")
		file2 := filepath.Join(tmpDir, "verbose2.txt")

		_ = os.WriteFile(file1, []byte("abc"), 0644)
		_ = os.WriteFile(file2, []byte("axc"), 0644)

		var buf bytes.Buffer

		_, _ = RunCmp(&buf, []string{file1, file2}, CmpOptions{Verbose: true})

		// Verbose mode prints byte number and values in octal
		output := buf.String()
		if output == "" {
			t.Errorf("RunCmp() verbose should produce output")
		}
	})

	t.Run("print bytes mode", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "bytes1.txt")
		file2 := filepath.Join(tmpDir, "bytes2.txt")

		_ = os.WriteFile(file1, []byte("abc"), 0644)
		_ = os.WriteFile(file2, []byte("aXc"), 0644)

		var buf bytes.Buffer

		_, _ = RunCmp(&buf, []string{file1, file2}, CmpOptions{PrintBytes: true})

		output := buf.String()
		if !strings.Contains(output, "differ") {
			t.Errorf("RunCmp() print bytes output = %q", output)
		}
	})

	t.Run("different lengths", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "len1.txt")
		file2 := filepath.Join(tmpDir, "len2.txt")

		_ = os.WriteFile(file1, []byte("hello"), 0644)
		_ = os.WriteFile(file2, []byte("hello world"), 0644)

		var buf bytes.Buffer

		result, _ := RunCmp(&buf, []string{file1, file2}, CmpOptions{})

		if result != CmpDiffer {
			t.Errorf("RunCmp() = %v, want CmpDiffer", result)
		}

		if !strings.Contains(buf.String(), "EOF") {
			t.Errorf("RunCmp() should report EOF difference")
		}
	})

	t.Run("max bytes limit", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "max1.txt")
		file2 := filepath.Join(tmpDir, "max2.txt")

		_ = os.WriteFile(file1, []byte("hello DIFF"), 0644)
		_ = os.WriteFile(file2, []byte("hello XXXX"), 0644)

		var buf bytes.Buffer

		// Only compare first 5 bytes
		result, _ := RunCmp(&buf, []string{file1, file2}, CmpOptions{MaxBytes: 5})

		if result != CmpEqual {
			t.Errorf("RunCmp() with max bytes = %v, want CmpEqual", result)
		}
	})

	t.Run("missing operand", func(t *testing.T) {
		var buf bytes.Buffer

		result, err := RunCmp(&buf, []string{}, CmpOptions{})
		if err == nil {
			t.Error("RunCmp() expected error for missing operand")
		}

		if result != CmpError {
			t.Errorf("RunCmp() = %v, want CmpError", result)
		}
	})

	t.Run("both stdin", func(t *testing.T) {
		var buf bytes.Buffer

		result, err := RunCmp(&buf, []string{"-", "-"}, CmpOptions{})
		if err == nil {
			t.Error("RunCmp() expected error when both files are stdin")
		}

		if result != CmpError {
			t.Errorf("RunCmp() = %v, want CmpError", result)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "exists.txt")
		_ = os.WriteFile(file1, []byte("data"), 0644)

		var buf bytes.Buffer

		result, err := RunCmp(&buf, []string{file1, "/nonexistent/file.txt"}, CmpOptions{})
		if err == nil {
			t.Error("RunCmp() expected error for nonexistent file")
		}

		if result != CmpError {
			t.Errorf("RunCmp() = %v, want CmpError", result)
		}
	})

	t.Run("files with newlines", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "nl1.txt")
		file2 := filepath.Join(tmpDir, "nl2.txt")

		_ = os.WriteFile(file1, []byte("line1\nline2\nline3"), 0644)
		_ = os.WriteFile(file2, []byte("line1\nlineX\nline3"), 0644)

		var buf bytes.Buffer

		result, _ := RunCmp(&buf, []string{file1, file2}, CmpOptions{})

		if result != CmpDiffer {
			t.Errorf("RunCmp() = %v, want CmpDiffer", result)
		}

		// Should report line number
		output := buf.String()
		if !strings.Contains(output, "line 2") {
			t.Errorf("RunCmp() should report line number: %s", output)
		}
	})
}

func TestPrintableChar(t *testing.T) {
	tests := []struct {
		input    byte
		expected byte
	}{
		{'a', 'a'},
		{'Z', 'Z'},
		{'5', '5'},
		{0, ' '},
		{127, ' '},
		{'\n', ' '},
		{32, 32},   // space
		{126, 126}, // tilde
	}

	for _, tt := range tests {
		result := printableChar(tt.input)
		if result != tt.expected {
			t.Errorf("printableChar(%d) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}
