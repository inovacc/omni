package rev

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// nilReader returns nil for stdin when testing with files
var nilReader = bytes.NewReader(nil)

func TestRunRev(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rev_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("reverse single line", func(t *testing.T) {
		file := filepath.Join(tmpDir, "single.txt")
		if err := os.WriteFile(file, []byte("hello\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunRev(&buf, nilReader, []string{file}, RevOptions{})
		if err != nil {
			t.Fatalf("RunRev() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output != "olleh" {
			t.Errorf("RunRev() = %v, want 'olleh'", output)
		}
	})

	t.Run("reverse multiple lines", func(t *testing.T) {
		file := filepath.Join(tmpDir, "multi.txt")
		if err := os.WriteFile(file, []byte("abc\n123\nxyz\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunRev(&buf, nilReader, []string{file}, RevOptions{})
		if err != nil {
			t.Fatalf("RunRev() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Errorf("RunRev() got %d lines, want 3", len(lines))
		}

		if lines[0] != "cba" || lines[1] != "321" || lines[2] != "zyx" {
			t.Errorf("RunRev() = %v", lines)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "empty.txt")
		if err := os.WriteFile(file, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunRev(&buf, nilReader, []string{file}, RevOptions{})
		if err != nil {
			t.Fatalf("RunRev() error = %v", err)
		}

		if buf.Len() != 0 {
			t.Errorf("RunRev() empty file should produce no output")
		}
	})

	t.Run("unicode content", func(t *testing.T) {
		file := filepath.Join(tmpDir, "unicode.txt")
		if err := os.WriteFile(file, []byte("世界\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunRev(&buf, nilReader, []string{file}, RevOptions{})
		if err != nil {
			t.Fatalf("RunRev() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output != "界世" {
			t.Errorf("RunRev() unicode = %v, want '界世'", output)
		}
	})

	t.Run("multiple files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")

		_ = os.WriteFile(file1, []byte("abc\n"), 0644)
		_ = os.WriteFile(file2, []byte("xyz\n"), 0644)

		var buf bytes.Buffer

		err := RunRev(&buf, nilReader, []string{file1, file2}, RevOptions{})
		if err != nil {
			t.Fatalf("RunRev() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Errorf("RunRev() got %d lines, want 2", len(lines))
		}

		if lines[0] != "cba" || lines[1] != "zyx" {
			t.Errorf("RunRev() = %v", lines)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunRev(&buf, nilReader, []string{"/nonexistent/file.txt"}, RevOptions{})
		if err == nil {
			t.Error("RunRev() expected error for nonexistent file")
		}
	})

	t.Run("palindrome", func(t *testing.T) {
		file := filepath.Join(tmpDir, "palindrome.txt")
		if err := os.WriteFile(file, []byte("racecar\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunRev(&buf, nilReader, []string{file}, RevOptions{})

		output := strings.TrimSpace(buf.String())
		if output != "racecar" {
			t.Errorf("RunRev() palindrome = %v, want 'racecar'", output)
		}
	})

	t.Run("line with spaces", func(t *testing.T) {
		file := filepath.Join(tmpDir, "spaces.txt")
		if err := os.WriteFile(file, []byte("hello world\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunRev(&buf, nilReader, []string{file}, RevOptions{})

		output := strings.TrimSpace(buf.String())
		if output != "dlrow olleh" {
			t.Errorf("RunRev() spaces = %v, want 'dlrow olleh'", output)
		}
	})
}

func TestReverseString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "olleh"},
		{"", ""},
		{"a", "a"},
		{"ab", "ba"},
		{"abc", "cba"},
		{"世界", "界世"},
		{"12345", "54321"},
		{"  ", "  "},
		{"a b", "b a"},
	}

	for _, tt := range tests {
		result := reverseString(tt.input)
		if result != tt.expected {
			t.Errorf("reverseString(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
