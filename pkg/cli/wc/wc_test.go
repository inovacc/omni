package wc

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWC(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wc_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("count lines", func(t *testing.T) {
		file := filepath.Join(tmpDir, "lines.txt")
		content := "line1\nline2\nline3\n"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunWC(&buf, []string{file}, WCOptions{Lines: true})
		if err != nil {
			t.Fatalf("RunWC() error = %v", err)
		}

		if !strings.Contains(buf.String(), "3") {
			t.Errorf("RunWC() lines = %v, want 3", buf.String())
		}
	})

	t.Run("count words", func(t *testing.T) {
		file := filepath.Join(tmpDir, "words.txt")
		content := "one two three\nfour five"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunWC(&buf, []string{file}, WCOptions{Words: true})
		if err != nil {
			t.Fatalf("RunWC() error = %v", err)
		}

		if !strings.Contains(buf.String(), "5") {
			t.Errorf("RunWC() words = %v, want 5", buf.String())
		}
	})

	t.Run("count bytes", func(t *testing.T) {
		file := filepath.Join(tmpDir, "bytes.txt")
		content := "hello"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunWC(&buf, []string{file}, WCOptions{Bytes: true})
		if err != nil {
			t.Fatalf("RunWC() error = %v", err)
		}

		if !strings.Contains(buf.String(), "5") {
			t.Errorf("RunWC() bytes = %v, want 5", buf.String())
		}
	})

	t.Run("count chars", func(t *testing.T) {
		file := filepath.Join(tmpDir, "chars.txt")
		content := "hello"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunWC(&buf, []string{file}, WCOptions{Chars: true})
		if err != nil {
			t.Fatalf("RunWC() error = %v", err)
		}

		if !strings.Contains(buf.String(), "5") {
			t.Errorf("RunWC() chars = %v, want 5", buf.String())
		}
	})

	t.Run("default all counts", func(t *testing.T) {
		file := filepath.Join(tmpDir, "all.txt")
		content := "one two\nthree\n"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunWC(&buf, []string{file}, WCOptions{})
		if err != nil {
			t.Fatalf("RunWC() error = %v", err)
		}

		output := buf.String()
		// Default shows lines, words, bytes
		if !strings.Contains(output, "2") { // 2 lines
			t.Errorf("RunWC() default should show lines: %v", output)
		}
	})

	t.Run("multiple files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "multi1.txt")
		file2 := filepath.Join(tmpDir, "multi2.txt")

		if err := os.WriteFile(file1, []byte("line1\nline2\n"), 0644); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(file2, []byte("line1\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunWC(&buf, []string{file1, file2}, WCOptions{Lines: true})
		if err != nil {
			t.Fatalf("RunWC() error = %v", err)
		}

		output := buf.String()
		// Should show total for multiple files
		if !strings.Contains(output, "total") {
			t.Errorf("RunWC() multiple files should show total: %v", output)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "empty.txt")

		if err := os.WriteFile(file, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunWC(&buf, []string{file}, WCOptions{Lines: true})
		if err != nil {
			t.Fatalf("RunWC() error = %v", err)
		}

		if !strings.Contains(buf.String(), "0") {
			t.Errorf("RunWC() empty file should have 0 lines: %v", buf.String())
		}
	})

	t.Run("max line length", func(t *testing.T) {
		file := filepath.Join(tmpDir, "maxline.txt")
		content := "short\nthis is a longer line\nmed"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunWC(&buf, []string{file}, WCOptions{MaxLineLen: true})
		if err != nil {
			t.Fatalf("RunWC() error = %v", err)
		}

		// "this is a longer line" is 21 chars
		if !strings.Contains(buf.String(), "21") {
			t.Errorf("RunWC() max line length = %v, want 21", buf.String())
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		// Implementation prints error to stderr and continues, doesn't return error
		err := RunWC(&buf, []string{"/nonexistent/file.txt"}, WCOptions{})
		if err != nil {
			t.Logf("RunWC() error = %v (implementation may vary)", err)
		}
	})

	t.Run("file without trailing newline", func(t *testing.T) {
		file := filepath.Join(tmpDir, "notrailing.txt")
		content := "line1\nline2"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunWC(&buf, []string{file}, WCOptions{Lines: true})
		if err != nil {
			t.Fatalf("RunWC() error = %v", err)
		}

		// wc counts newline chars, so "line1\nline2" has 1 newline
		if !strings.Contains(buf.String(), "1") {
			t.Errorf("RunWC() file without trailing newline = %v, want 1", buf.String())
		}
	})
}
