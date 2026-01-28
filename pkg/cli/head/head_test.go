package head

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunHead(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "head_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	createTestFile := func(name string, lines int) string {
		file := filepath.Join(tmpDir, name)
		var content strings.Builder

		for i := 1; i <= lines; i++ {
			content.WriteString("line")
			content.WriteString(string(rune('0' + i)))
			content.WriteString("\n")
		}

		_ = os.WriteFile(file, []byte(content.String()), 0644)

		return file
	}

	t.Run("default 10 lines", func(t *testing.T) {
		file := createTestFile("default.txt", 15)

		var buf bytes.Buffer

		err := RunHead(&buf, []string{file}, HeadOptions{})
		if err != nil {
			t.Fatalf("RunHead() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 10 {
			t.Errorf("RunHead() default got %d lines, want 10", len(lines))
		}
	})

	t.Run("custom line count", func(t *testing.T) {
		file := createTestFile("custom.txt", 10)

		var buf bytes.Buffer

		err := RunHead(&buf, []string{file}, HeadOptions{Lines: 3})
		if err != nil {
			t.Fatalf("RunHead() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Errorf("RunHead() got %d lines, want 3", len(lines))
		}

		if lines[0] != "line1" || lines[2] != "line3" {
			t.Errorf("RunHead() wrong lines: %v", lines)
		}
	})

	t.Run("file shorter than requested lines", func(t *testing.T) {
		file := createTestFile("short.txt", 3)

		var buf bytes.Buffer

		err := RunHead(&buf, []string{file}, HeadOptions{Lines: 10})
		if err != nil {
			t.Fatalf("RunHead() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Errorf("RunHead() got %d lines, want 3", len(lines))
		}
	})

	t.Run("zero lines defaults to 10", func(t *testing.T) {
		file := createTestFile("zero.txt", 15)

		var buf bytes.Buffer

		err := RunHead(&buf, []string{file}, HeadOptions{Lines: 0})
		if err != nil {
			t.Fatalf("RunHead() error = %v", err)
		}

		// Lines: 0 is treated as default (10 lines)
		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 10 {
			t.Errorf("RunHead() with Lines=0 got %d lines, want 10 (default)", len(lines))
		}
	})

	t.Run("multiple files", func(t *testing.T) {
		file1 := createTestFile("multi1.txt", 5)
		file2 := createTestFile("multi2.txt", 5)

		var buf bytes.Buffer

		err := RunHead(&buf, []string{file1, file2}, HeadOptions{Lines: 2})
		if err != nil {
			t.Fatalf("RunHead() error = %v", err)
		}

		output := buf.String()
		// Should contain headers for multiple files
		if !strings.Contains(output, "==>") {
			t.Errorf("RunHead() multiple files should have headers: %v", output)
		}
	})

	t.Run("quiet mode no headers", func(t *testing.T) {
		file1 := createTestFile("quiet1.txt", 5)
		file2 := createTestFile("quiet2.txt", 5)

		var buf bytes.Buffer

		err := RunHead(&buf, []string{file1, file2}, HeadOptions{Lines: 2, Quiet: true})
		if err != nil {
			t.Fatalf("RunHead() error = %v", err)
		}

		output := buf.String()
		if strings.Contains(output, "==>") {
			t.Errorf("RunHead() quiet mode should not have headers: %v", output)
		}
	})

	t.Run("bytes mode", func(t *testing.T) {
		file := filepath.Join(tmpDir, "bytes.txt")
		content := "0123456789abcdef"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunHead(&buf, []string{file}, HeadOptions{Bytes: 5})
		if err != nil {
			t.Fatalf("RunHead() error = %v", err)
		}

		if buf.String() != "01234" {
			t.Errorf("RunHead() bytes mode = %v, want 01234", buf.String())
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunHead(&buf, []string{"/nonexistent/file.txt"}, HeadOptions{})
		if err == nil {
			t.Error("RunHead() should return error for nonexistent file")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "empty.txt")
		if err := os.WriteFile(file, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunHead(&buf, []string{file}, HeadOptions{Lines: 10})
		if err != nil {
			t.Fatalf("RunHead() error = %v", err)
		}

		if buf.Len() != 0 {
			t.Errorf("RunHead() empty file should produce no output")
		}
	})
}
