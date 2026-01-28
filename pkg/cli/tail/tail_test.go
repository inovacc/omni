package tail

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunTail(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tail_test")
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

		err := RunTail(&buf, []string{file}, TailOptions{})
		if err != nil {
			t.Fatalf("RunTail() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 10 {
			t.Errorf("RunTail() default got %d lines, want 10", len(lines))
		}
	})

	t.Run("custom line count", func(t *testing.T) {
		file := filepath.Join(tmpDir, "custom.txt")
		content := "line1\nline2\nline3\nline4\nline5"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunTail(&buf, []string{file}, TailOptions{Lines: 2})
		if err != nil {
			t.Fatalf("RunTail() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Errorf("RunTail() got %d lines, want 2", len(lines))
		}

		if lines[0] != "line4" || lines[1] != "line5" {
			t.Errorf("RunTail() = %v, want [line4, line5]", lines)
		}
	})

	t.Run("file shorter than requested lines", func(t *testing.T) {
		file := createTestFile("short.txt", 3)

		var buf bytes.Buffer

		err := RunTail(&buf, []string{file}, TailOptions{Lines: 10})
		if err != nil {
			t.Fatalf("RunTail() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Errorf("RunTail() got %d lines, want 3", len(lines))
		}
	})

	t.Run("zero lines", func(t *testing.T) {
		file := createTestFile("zero.txt", 5)

		var buf bytes.Buffer

		err := RunTail(&buf, []string{file}, TailOptions{Lines: 0})
		if err != nil {
			t.Fatalf("RunTail() error = %v", err)
		}

		if strings.TrimSpace(buf.String()) != "" {
			t.Errorf("RunTail() with 0 lines should output nothing")
		}
	})

	t.Run("multiple files", func(t *testing.T) {
		file1 := createTestFile("multi1.txt", 5)
		file2 := createTestFile("multi2.txt", 5)

		var buf bytes.Buffer

		err := RunTail(&buf, []string{file1, file2}, TailOptions{Lines: 2})
		if err != nil {
			t.Fatalf("RunTail() error = %v", err)
		}

		output := buf.String()
		// Should contain headers for multiple files
		if !strings.Contains(output, "==>") {
			t.Errorf("RunTail() multiple files should have headers: %v", output)
		}
	})

	t.Run("quiet mode no headers", func(t *testing.T) {
		file1 := createTestFile("quiet1.txt", 5)
		file2 := createTestFile("quiet2.txt", 5)

		var buf bytes.Buffer

		err := RunTail(&buf, []string{file1, file2}, TailOptions{Lines: 2, Quiet: true})
		if err != nil {
			t.Fatalf("RunTail() error = %v", err)
		}

		output := buf.String()
		if strings.Contains(output, "==>") {
			t.Errorf("RunTail() quiet mode should not have headers: %v", output)
		}
	})

	t.Run("bytes mode", func(t *testing.T) {
		file := filepath.Join(tmpDir, "bytes.txt")
		content := "0123456789abcdef"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunTail(&buf, []string{file}, TailOptions{Bytes: 5})
		if err != nil {
			t.Fatalf("RunTail() error = %v", err)
		}

		if buf.String() != "bcdef" {
			t.Errorf("RunTail() bytes mode = %v, want bcdef", buf.String())
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunTail(&buf, []string{"/nonexistent/file.txt"}, TailOptions{})
		if err == nil {
			t.Error("RunTail() should return error for nonexistent file")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "empty.txt")
		if err := os.WriteFile(file, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunTail(&buf, []string{file}, TailOptions{Lines: 10})
		if err != nil {
			t.Fatalf("RunTail() error = %v", err)
		}

		if buf.Len() != 0 {
			t.Errorf("RunTail() empty file should produce no output")
		}
	})

	t.Run("single line file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "single.txt")
		content := "only line"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunTail(&buf, []string{file}, TailOptions{Lines: 5})
		if err != nil {
			t.Fatalf("RunTail() error = %v", err)
		}

		if strings.TrimSpace(buf.String()) != "only line" {
			t.Errorf("RunTail() = %v, want 'only line'", buf.String())
		}
	})
}
