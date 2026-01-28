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

	createTestFile := func(name string, numLines int) string {
		file := filepath.Join(tmpDir, name)
		var content strings.Builder

		for i := 1; i <= numLines; i++ {
			content.WriteString("line")
			content.WriteString(string(rune('0' + i%10)))
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

	t.Run("lines 1", func(t *testing.T) {
		file := filepath.Join(tmpDir, "one.txt")
		content := "first\nsecond\nthird"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunTail(&buf, []string{file}, TailOptions{Lines: 1})

		if strings.TrimSpace(buf.String()) != "third" {
			t.Errorf("RunTail() Lines=1 = %v, want 'third'", buf.String())
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

	t.Run("zero lines defaults to 10", func(t *testing.T) {
		file := createTestFile("zero.txt", 15)

		var buf bytes.Buffer

		err := RunTail(&buf, []string{file}, TailOptions{Lines: 0})
		if err != nil {
			t.Fatalf("RunTail() error = %v", err)
		}

		// Lines: 0 defaults to 10
		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 10 {
			t.Errorf("RunTail() with Lines=0 got %d lines, want 10", len(lines))
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

	t.Run("multiple files content", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "m1.txt")
		file2 := filepath.Join(tmpDir, "m2.txt")

		_ = os.WriteFile(file1, []byte("file1line1\nfile1line2\n"), 0644)
		_ = os.WriteFile(file2, []byte("file2line1\nfile2line2\n"), 0644)

		var buf bytes.Buffer

		_ = RunTail(&buf, []string{file1, file2}, TailOptions{Lines: 1})

		output := buf.String()
		if !strings.Contains(output, "file1line2") || !strings.Contains(output, "file2line2") {
			t.Errorf("RunTail() multiple files should show last line of each: %v", output)
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

	t.Run("verbose mode single file", func(t *testing.T) {
		file := createTestFile("verbose.txt", 5)

		var buf bytes.Buffer

		_ = RunTail(&buf, []string{file}, TailOptions{Lines: 2, Verbose: true})

		output := buf.String()
		if !strings.Contains(output, "==>") {
			t.Errorf("RunTail() verbose mode should show header even for single file: %v", output)
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

	t.Run("bytes mode more than file size", func(t *testing.T) {
		file := filepath.Join(tmpDir, "bytes_more.txt")
		content := "short"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunTail(&buf, []string{file}, TailOptions{Bytes: 100})

		if buf.String() != "short" {
			t.Errorf("RunTail() bytes more than size = %v, want 'short'", buf.String())
		}
	})

	t.Run("bytes mode exact size", func(t *testing.T) {
		file := filepath.Join(tmpDir, "bytes_exact.txt")
		content := "12345"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunTail(&buf, []string{file}, TailOptions{Bytes: 5})

		if buf.String() != "12345" {
			t.Errorf("RunTail() bytes exact size = %v, want '12345'", buf.String())
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

	t.Run("file with no trailing newline", func(t *testing.T) {
		file := filepath.Join(tmpDir, "notrailing.txt")
		content := "line1\nline2\nline3"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunTail(&buf, []string{file}, TailOptions{Lines: 2})

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Errorf("RunTail() no trailing newline got %d lines, want 2", len(lines))
		}

		if lines[0] != "line2" || lines[1] != "line3" {
			t.Errorf("RunTail() = %v, want [line2, line3]", lines)
		}
	})

	t.Run("file with only newlines", func(t *testing.T) {
		file := filepath.Join(tmpDir, "newlines.txt")
		content := "\n\n\n\n\n"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunTail(&buf, []string{file}, TailOptions{Lines: 3})

		lines := strings.Split(buf.String(), "\n")
		// Should have 3 empty lines (plus potential trailing from split)
		count := 0
		for _, l := range lines {
			if l == "" {
				count++
			}
		}

		if count < 3 {
			t.Errorf("RunTail() newlines file got %d empty lines", count)
		}
	})

	t.Run("unicode content", func(t *testing.T) {
		file := filepath.Join(tmpDir, "unicode.txt")
		content := "Hello\n世界\n🌍\nこんにちは"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunTail(&buf, []string{file}, TailOptions{Lines: 2})

		output := buf.String()
		if !strings.Contains(output, "🌍") || !strings.Contains(output, "こんにちは") {
			t.Errorf("RunTail() should preserve unicode: %v", output)
		}
	})

	t.Run("very long lines", func(t *testing.T) {
		file := filepath.Join(tmpDir, "longlines.txt")
		longLine := strings.Repeat("x", 10000)
		content := "short\n" + longLine + "\n"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunTail(&buf, []string{file}, TailOptions{Lines: 1})

		if strings.TrimSpace(buf.String()) != longLine {
			t.Errorf("RunTail() should handle long lines")
		}
	})

	t.Run("large file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "large.txt")
		var content strings.Builder

		for i := 0; i < 10000; i++ {
			content.WriteString("line")
			content.WriteString(string(rune('0' + i%10)))
			content.WriteString("\n")
		}

		_ = os.WriteFile(file, []byte(content.String()), 0644)

		var buf bytes.Buffer

		err := RunTail(&buf, []string{file}, TailOptions{Lines: 5})
		if err != nil {
			t.Fatalf("RunTail() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 5 {
			t.Errorf("RunTail() large file got %d lines, want 5", len(lines))
		}
	})
}

func TestTail(t *testing.T) {
	t.Run("basic tail", func(t *testing.T) {
		lines := []string{"a", "b", "c", "d", "e"}

		result := Tail(lines, 3)
		if len(result) != 3 {
			t.Errorf("Tail() got %d lines, want 3", len(result))
		}

		if result[0] != "c" || result[1] != "d" || result[2] != "e" {
			t.Errorf("Tail() = %v, want [c, d, e]", result)
		}
	})

	t.Run("tail more than length", func(t *testing.T) {
		lines := []string{"a", "b", "c"}

		result := Tail(lines, 10)
		if len(result) != 3 {
			t.Errorf("Tail() got %d lines, want 3", len(result))
		}
	})

	t.Run("tail zero", func(t *testing.T) {
		lines := []string{"a", "b", "c"}

		result := Tail(lines, 0)
		if len(result) != 0 {
			t.Errorf("Tail() zero got %d lines, want 0", len(result))
		}
	})

	t.Run("tail negative", func(t *testing.T) {
		lines := []string{"a", "b", "c"}

		result := Tail(lines, -1)
		if len(result) != 0 {
			t.Errorf("Tail() negative got %d lines, want 0", len(result))
		}
	})

	t.Run("tail empty slice", func(t *testing.T) {
		var lines []string

		result := Tail(lines, 5)
		if len(result) != 0 {
			t.Errorf("Tail() empty slice got %d lines, want 0", len(result))
		}
	})

	t.Run("tail one", func(t *testing.T) {
		lines := []string{"a", "b", "c", "d", "e"}

		result := Tail(lines, 1)
		if len(result) != 1 || result[0] != "e" {
			t.Errorf("Tail() one = %v, want [e]", result)
		}
	})

	t.Run("tail all", func(t *testing.T) {
		lines := []string{"a", "b", "c"}

		result := Tail(lines, 3)
		if len(result) != 3 {
			t.Errorf("Tail() all got %d lines, want 3", len(result))
		}
	})
}
