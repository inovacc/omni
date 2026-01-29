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

	createTestFile := func(name string, numLines int) string {
		file := filepath.Join(tmpDir, name)

		var content strings.Builder

		for i := 1; i <= numLines; i++ {
			content.WriteString("line")
			content.WriteRune(rune('0' + i%10))
			content.WriteString("\n")
		}

		_ = os.WriteFile(file, []byte(content.String()), 0644)

		return file
	}

	t.Run("default 10 lines", func(t *testing.T) {
		file := createTestFile("default.txt", 15)

		var buf bytes.Buffer

		err := RunHead(&buf, nil, []string{file}, HeadOptions{})
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

		err := RunHead(&buf, nil, []string{file}, HeadOptions{Lines: 3})
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

	t.Run("lines 1", func(t *testing.T) {
		file := filepath.Join(tmpDir, "one.txt")
		content := "first\nsecond\nthird"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunHead(&buf, nil, []string{file}, HeadOptions{Lines: 1})

		if strings.TrimSpace(buf.String()) != "first" {
			t.Errorf("RunHead() Lines=1 = %v, want 'first'", buf.String())
		}
	})

	t.Run("lines 5", func(t *testing.T) {
		file := createTestFile("five.txt", 10)

		var buf bytes.Buffer

		_ = RunHead(&buf, nil, []string{file}, HeadOptions{Lines: 5})

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 5 {
			t.Errorf("RunHead() Lines=5 got %d lines", len(lines))
		}
	})

	t.Run("file shorter than requested lines", func(t *testing.T) {
		file := createTestFile("short.txt", 3)

		var buf bytes.Buffer

		err := RunHead(&buf, nil, []string{file}, HeadOptions{Lines: 10})
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

		err := RunHead(&buf, nil, []string{file}, HeadOptions{Lines: 0})
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

		err := RunHead(&buf, nil, []string{file1, file2}, HeadOptions{Lines: 2})
		if err != nil {
			t.Fatalf("RunHead() error = %v", err)
		}

		output := buf.String()
		// Should contain headers for multiple files
		if !strings.Contains(output, "==>") {
			t.Errorf("RunHead() multiple files should have headers: %v", output)
		}
	})

	t.Run("multiple files content", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "m1.txt")
		file2 := filepath.Join(tmpDir, "m2.txt")

		_ = os.WriteFile(file1, []byte("file1line1\nfile1line2\n"), 0644)
		_ = os.WriteFile(file2, []byte("file2line1\nfile2line2\n"), 0644)

		var buf bytes.Buffer

		_ = RunHead(&buf, nil, []string{file1, file2}, HeadOptions{Lines: 1})

		output := buf.String()
		if !strings.Contains(output, "file1line1") || !strings.Contains(output, "file2line1") {
			t.Errorf("RunHead() multiple files should show first line of each: %v", output)
		}
	})

	t.Run("three files", func(t *testing.T) {
		files := []string{
			filepath.Join(tmpDir, "a.txt"),
			filepath.Join(tmpDir, "b.txt"),
			filepath.Join(tmpDir, "c.txt"),
		}

		for i, f := range files {
			_ = os.WriteFile(f, []byte(strings.Repeat("x", i+1)+"\n"), 0644)
		}

		var buf bytes.Buffer

		_ = RunHead(&buf, nil, files, HeadOptions{Lines: 1})

		output := buf.String()
		// Should have 3 headers
		count := strings.Count(output, "==>")
		if count != 3 {
			t.Errorf("RunHead() three files got %d headers, want 3", count)
		}
	})

	t.Run("quiet mode no headers", func(t *testing.T) {
		file1 := createTestFile("quiet1.txt", 5)
		file2 := createTestFile("quiet2.txt", 5)

		var buf bytes.Buffer

		err := RunHead(&buf, nil, []string{file1, file2}, HeadOptions{Lines: 2, Quiet: true})
		if err != nil {
			t.Fatalf("RunHead() error = %v", err)
		}

		output := buf.String()
		if strings.Contains(output, "==>") {
			t.Errorf("RunHead() quiet mode should not have headers: %v", output)
		}
	})

	t.Run("verbose mode single file", func(t *testing.T) {
		file := createTestFile("verbose.txt", 5)

		var buf bytes.Buffer

		_ = RunHead(&buf, nil, []string{file}, HeadOptions{Lines: 2, Verbose: true})

		output := buf.String()
		if !strings.Contains(output, "==>") {
			t.Errorf("RunHead() verbose should show header for single file: %v", output)
		}
	})

	t.Run("bytes mode", func(t *testing.T) {
		file := filepath.Join(tmpDir, "bytes.txt")
		content := "0123456789abcdef"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunHead(&buf, nil, []string{file}, HeadOptions{Bytes: 5})
		if err != nil {
			t.Fatalf("RunHead() error = %v", err)
		}

		if buf.String() != "01234" {
			t.Errorf("RunHead() bytes mode = %v, want 01234", buf.String())
		}
	})

	t.Run("bytes mode more than file size", func(t *testing.T) {
		file := filepath.Join(tmpDir, "bytes_more.txt")
		content := "short"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunHead(&buf, nil, []string{file}, HeadOptions{Bytes: 100})

		if buf.String() != "short" {
			t.Errorf("RunHead() bytes more than size = %v, want 'short'", buf.String())
		}
	})

	t.Run("bytes mode exact size", func(t *testing.T) {
		file := filepath.Join(tmpDir, "bytes_exact.txt")
		content := "12345"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunHead(&buf, nil, []string{file}, HeadOptions{Bytes: 5})

		if buf.String() != "12345" {
			t.Errorf("RunHead() bytes exact = %v, want '12345'", buf.String())
		}
	})

	t.Run("bytes 1", func(t *testing.T) {
		file := filepath.Join(tmpDir, "bytes1.txt")
		content := "hello"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunHead(&buf, nil, []string{file}, HeadOptions{Bytes: 1})

		if buf.String() != "h" {
			t.Errorf("RunHead() bytes=1 = %v, want 'h'", buf.String())
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunHead(&buf, nil, []string{"/nonexistent/file.txt"}, HeadOptions{})
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

		err := RunHead(&buf, nil, []string{file}, HeadOptions{Lines: 10})
		if err != nil {
			t.Fatalf("RunHead() error = %v", err)
		}

		if buf.Len() != 0 {
			t.Errorf("RunHead() empty file should produce no output")
		}
	})

	t.Run("single line file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "single.txt")
		content := "only line"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunHead(&buf, nil, []string{file}, HeadOptions{Lines: 10})

		if strings.TrimSpace(buf.String()) != "only line" {
			t.Errorf("RunHead() single = %v", buf.String())
		}
	})

	t.Run("file with no trailing newline", func(t *testing.T) {
		file := filepath.Join(tmpDir, "notrailing.txt")
		content := "line1\nline2\nline3"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunHead(&buf, nil, []string{file}, HeadOptions{Lines: 2})

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Errorf("RunHead() no trailing got %d lines, want 2", len(lines))
		}
	})

	t.Run("unicode content", func(t *testing.T) {
		file := filepath.Join(tmpDir, "unicode.txt")
		content := "Hello\n世界\n🌍\nこんにちは"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunHead(&buf, nil, []string{file}, HeadOptions{Lines: 2})

		output := buf.String()
		if !strings.Contains(output, "Hello") || !strings.Contains(output, "世界") {
			t.Errorf("RunHead() should preserve unicode: %v", output)
		}
	})

	t.Run("very long lines", func(t *testing.T) {
		file := filepath.Join(tmpDir, "longlines.txt")
		longLine := strings.Repeat("x", 10000)
		content := longLine + "\nshort\n"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunHead(&buf, nil, []string{file}, HeadOptions{Lines: 1})

		if strings.TrimSpace(buf.String()) != longLine {
			t.Errorf("RunHead() should handle long lines")
		}
	})

	t.Run("large file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "large.txt")

		var content strings.Builder

		for i := range 10000 {
			content.WriteString("line")
			content.WriteRune(rune('0' + i%10))
			content.WriteString("\n")
		}

		_ = os.WriteFile(file, []byte(content.String()), 0644)

		var buf bytes.Buffer

		err := RunHead(&buf, nil, []string{file}, HeadOptions{Lines: 5})
		if err != nil {
			t.Fatalf("RunHead() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 5 {
			t.Errorf("RunHead() large file got %d lines, want 5", len(lines))
		}
	})

	t.Run("output ends with newline", func(t *testing.T) {
		file := createTestFile("newline.txt", 5)

		var buf bytes.Buffer

		_ = RunHead(&buf, nil, []string{file}, HeadOptions{Lines: 2})

		if !strings.HasSuffix(buf.String(), "\n") {
			t.Error("RunHead() output should end with newline")
		}
	})

	t.Run("consistent results", func(t *testing.T) {
		file := createTestFile("consistent.txt", 10)

		var buf1, buf2 bytes.Buffer

		_ = RunHead(&buf1, nil, []string{file}, HeadOptions{Lines: 5})
		_ = RunHead(&buf2, nil, []string{file}, HeadOptions{Lines: 5})

		if buf1.String() != buf2.String() {
			t.Errorf("RunHead() should be consistent")
		}
	})
}

func TestHead(t *testing.T) {
	t.Run("basic head", func(t *testing.T) {
		lines := []string{"a", "b", "c", "d", "e"}

		result := Head(lines, 3)
		if len(result) != 3 {
			t.Errorf("Head() got %d lines, want 3", len(result))
		}

		if result[0] != "a" || result[1] != "b" || result[2] != "c" {
			t.Errorf("Head() = %v, want [a, b, c]", result)
		}
	})

	t.Run("head more than length", func(t *testing.T) {
		lines := []string{"a", "b", "c"}

		result := Head(lines, 10)
		if len(result) != 3 {
			t.Errorf("Head() got %d lines, want 3", len(result))
		}
	})

	t.Run("head zero", func(t *testing.T) {
		lines := []string{"a", "b", "c"}

		result := Head(lines, 0)
		if len(result) != 0 {
			t.Errorf("Head() zero got %d lines, want 0", len(result))
		}
	})

	t.Run("head negative", func(t *testing.T) {
		lines := []string{"a", "b", "c"}

		result := Head(lines, -1)
		if len(result) != 0 {
			t.Errorf("Head() negative got %d lines, want 0", len(result))
		}
	})

	t.Run("head empty slice", func(t *testing.T) {
		var lines []string

		result := Head(lines, 5)
		if len(result) != 0 {
			t.Errorf("Head() empty slice got %d lines, want 0", len(result))
		}
	})

	t.Run("head one", func(t *testing.T) {
		lines := []string{"a", "b", "c", "d", "e"}

		result := Head(lines, 1)
		if len(result) != 1 || result[0] != "a" {
			t.Errorf("Head() one = %v, want [a]", result)
		}
	})

	t.Run("head all", func(t *testing.T) {
		lines := []string{"a", "b", "c"}

		result := Head(lines, 3)
		if len(result) != 3 {
			t.Errorf("Head() all got %d lines, want 3", len(result))
		}
	})
}
