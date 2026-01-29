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

		err := RunWC(&buf, nil, []string{file}, WCOptions{Lines: true})
		if err != nil {
			t.Fatalf("RunWC() error = %v", err)
		}

		if !strings.Contains(buf.String(), "3") {
			t.Errorf("RunWC() lines = %v, want 3", buf.String())
		}
	})

	t.Run("count lines no trailing newline", func(t *testing.T) {
		file := filepath.Join(tmpDir, "lines_notrail.txt")
		content := "line1\nline2"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunWC(&buf, nil, []string{file}, WCOptions{Lines: true})

		// wc counts newlines, so "line1\nline2" has 1 newline
		if !strings.Contains(buf.String(), "1") {
			t.Errorf("RunWC() lines no trailing = %v, want 1", buf.String())
		}
	})

	t.Run("count words", func(t *testing.T) {
		file := filepath.Join(tmpDir, "words.txt")
		content := "one two three\nfour five"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunWC(&buf, nil, []string{file}, WCOptions{Words: true})
		if err != nil {
			t.Fatalf("RunWC() error = %v", err)
		}

		if !strings.Contains(buf.String(), "5") {
			t.Errorf("RunWC() words = %v, want 5", buf.String())
		}
	})

	t.Run("count words with multiple spaces", func(t *testing.T) {
		file := filepath.Join(tmpDir, "words_spaces.txt")
		content := "one   two    three"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunWC(&buf, nil, []string{file}, WCOptions{Words: true})

		if !strings.Contains(buf.String(), "3") {
			t.Errorf("RunWC() words with spaces = %v, want 3", buf.String())
		}
	})

	t.Run("count words with tabs", func(t *testing.T) {
		file := filepath.Join(tmpDir, "words_tabs.txt")
		content := "one\ttwo\tthree"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunWC(&buf, nil, []string{file}, WCOptions{Words: true})

		if !strings.Contains(buf.String(), "3") {
			t.Errorf("RunWC() words with tabs = %v, want 3", buf.String())
		}
	})

	t.Run("count bytes", func(t *testing.T) {
		file := filepath.Join(tmpDir, "bytes.txt")
		content := "hello"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunWC(&buf, nil, []string{file}, WCOptions{Bytes: true})
		if err != nil {
			t.Fatalf("RunWC() error = %v", err)
		}

		if !strings.Contains(buf.String(), "5") {
			t.Errorf("RunWC() bytes = %v, want 5", buf.String())
		}
	})

	t.Run("count bytes with newline", func(t *testing.T) {
		file := filepath.Join(tmpDir, "bytes_nl.txt")
		content := "hello\n"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunWC(&buf, nil, []string{file}, WCOptions{Bytes: true})

		if !strings.Contains(buf.String(), "6") {
			t.Errorf("RunWC() bytes with newline = %v, want 6", buf.String())
		}
	})

	t.Run("count chars ASCII", func(t *testing.T) {
		file := filepath.Join(tmpDir, "chars.txt")
		content := "hello"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunWC(&buf, nil, []string{file}, WCOptions{Chars: true})
		if err != nil {
			t.Fatalf("RunWC() error = %v", err)
		}

		if !strings.Contains(buf.String(), "5") {
			t.Errorf("RunWC() chars = %v, want 5", buf.String())
		}
	})

	t.Run("count chars unicode", func(t *testing.T) {
		file := filepath.Join(tmpDir, "chars_unicode.txt")
		content := "日本語" // 3 characters, but 9 bytes in UTF-8

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunWC(&buf, nil, []string{file}, WCOptions{Chars: true})

		if !strings.Contains(buf.String(), "3") {
			t.Errorf("RunWC() chars unicode = %v, want 3", buf.String())
		}
	})

	t.Run("bytes vs chars unicode", func(t *testing.T) {
		file := filepath.Join(tmpDir, "unicode.txt")
		content := "日本語" // 3 chars, 9 bytes

		_ = os.WriteFile(file, []byte(content), 0644)

		var bufChars, bufBytes bytes.Buffer

		_ = RunWC(&bufChars, nil, []string{file}, WCOptions{Chars: true})
		_ = RunWC(&bufBytes, nil, []string{file}, WCOptions{Bytes: true})

		// Chars should be 3, bytes should be 9
		if !strings.Contains(bufChars.String(), "3") {
			t.Errorf("RunWC() chars unicode = %v, want 3", bufChars.String())
		}

		if !strings.Contains(bufBytes.String(), "9") {
			t.Errorf("RunWC() bytes unicode = %v, want 9", bufBytes.String())
		}
	})

	t.Run("default all counts", func(t *testing.T) {
		file := filepath.Join(tmpDir, "all.txt")
		content := "one two\nthree\n"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunWC(&buf, nil, []string{file}, WCOptions{})
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

		err := RunWC(&buf, nil, []string{file1, file2}, WCOptions{Lines: true})
		if err != nil {
			t.Fatalf("RunWC() error = %v", err)
		}

		output := buf.String()
		// Should show total for multiple files
		if !strings.Contains(output, "total") {
			t.Errorf("RunWC() multiple files should show total: %v", output)
		}
	})

	t.Run("multiple files total", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "total1.txt")
		file2 := filepath.Join(tmpDir, "total2.txt")

		_ = os.WriteFile(file1, []byte("a\nb\n"), 0644)    // 2 lines
		_ = os.WriteFile(file2, []byte("c\nd\ne\n"), 0644) // 3 lines

		var buf bytes.Buffer

		_ = RunWC(&buf, nil, []string{file1, file2}, WCOptions{Lines: true})

		output := buf.String()
		// Total should be 5 lines
		if !strings.Contains(output, "5") || !strings.Contains(output, "total") {
			t.Errorf("RunWC() total = %v, want 5 total", output)
		}
	})

	t.Run("three files", func(t *testing.T) {
		files := []string{
			filepath.Join(tmpDir, "t1.txt"),
			filepath.Join(tmpDir, "t2.txt"),
			filepath.Join(tmpDir, "t3.txt"),
		}

		for _, f := range files {
			_ = os.WriteFile(f, []byte("line\n"), 0644)
		}

		var buf bytes.Buffer

		_ = RunWC(&buf, nil, files, WCOptions{Lines: true})

		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")

		// Should have 4 lines: 3 files + 1 total
		if len(lines) != 4 {
			t.Errorf("RunWC() three files got %d output lines, want 4", len(lines))
		}
	})

	t.Run("empty file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "empty.txt")

		if err := os.WriteFile(file, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunWC(&buf, nil, []string{file}, WCOptions{Lines: true})
		if err != nil {
			t.Fatalf("RunWC() error = %v", err)
		}

		if !strings.Contains(buf.String(), "0") {
			t.Errorf("RunWC() empty file should have 0 lines: %v", buf.String())
		}
	})

	t.Run("empty file all counts", func(t *testing.T) {
		file := filepath.Join(tmpDir, "empty_all.txt")

		_ = os.WriteFile(file, []byte(""), 0644)

		var buf bytes.Buffer

		_ = RunWC(&buf, nil, []string{file}, WCOptions{})

		output := buf.String()
		// Should have 0 for all counts
		if !strings.Contains(output, "0") {
			t.Errorf("RunWC() empty all = %v", output)
		}
	})

	t.Run("max line length", func(t *testing.T) {
		file := filepath.Join(tmpDir, "maxline.txt")
		content := "short\nthis is a longer line\nmed"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunWC(&buf, nil, []string{file}, WCOptions{MaxLineLen: true})
		if err != nil {
			t.Fatalf("RunWC() error = %v", err)
		}

		// "this is a longer line" is 21 chars
		if !strings.Contains(buf.String(), "21") {
			t.Errorf("RunWC() max line length = %v, want 21", buf.String())
		}
	})

	t.Run("max line length single line", func(t *testing.T) {
		file := filepath.Join(tmpDir, "maxline_single.txt")
		content := "exactly ten"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunWC(&buf, nil, []string{file}, WCOptions{MaxLineLen: true})

		if !strings.Contains(buf.String(), "11") {
			t.Errorf("RunWC() max line single = %v, want 11", buf.String())
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		// Implementation prints error to stderr and continues, doesn't return error
		err := RunWC(&buf, nil, []string{"/nonexistent/file.txt"}, WCOptions{})
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

		err := RunWC(&buf, nil, []string{file}, WCOptions{Lines: true})
		if err != nil {
			t.Fatalf("RunWC() error = %v", err)
		}

		// wc counts newline chars, so "line1\nline2" has 1 newline
		if !strings.Contains(buf.String(), "1") {
			t.Errorf("RunWC() file without trailing newline = %v, want 1", buf.String())
		}
	})

	t.Run("single word", func(t *testing.T) {
		file := filepath.Join(tmpDir, "single_word.txt")
		content := "hello"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunWC(&buf, nil, []string{file}, WCOptions{Words: true})

		if !strings.Contains(buf.String(), "1") {
			t.Errorf("RunWC() single word = %v, want 1", buf.String())
		}
	})

	t.Run("only whitespace", func(t *testing.T) {
		file := filepath.Join(tmpDir, "whitespace.txt")
		content := "   \n\t\t\n   "

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunWC(&buf, nil, []string{file}, WCOptions{Words: true})

		if !strings.Contains(buf.String(), "0") {
			t.Errorf("RunWC() whitespace only = %v, want 0 words", buf.String())
		}
	})

	t.Run("combined options", func(t *testing.T) {
		file := filepath.Join(tmpDir, "combined.txt")
		content := "one two three\nfour five\n"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunWC(&buf, nil, []string{file}, WCOptions{Lines: true, Words: true, Bytes: true})

		output := buf.String()
		// Should have 2 lines, 5 words, and byte count
		if !strings.Contains(output, "2") || !strings.Contains(output, "5") {
			t.Errorf("RunWC() combined = %v", output)
		}
	})
}

func TestWC(t *testing.T) {
	t.Run("basic wc", func(t *testing.T) {
		data := []byte("hello world\nfoo bar\n")

		result := WC(data)

		if result.Lines != 2 {
			t.Errorf("WC() lines = %d, want 2", result.Lines)
		}

		if result.Words != 4 {
			t.Errorf("WC() words = %d, want 4", result.Words)
		}

		if result.Bytes != 20 {
			t.Errorf("WC() bytes = %d, want 20", result.Bytes)
		}
	})

	t.Run("empty data", func(t *testing.T) {
		data := []byte("")

		result := WC(data)

		if result.Lines != 0 || result.Words != 0 || result.Bytes != 0 {
			t.Errorf("WC() empty = %+v, want all zeros", result)
		}
	})

	t.Run("single line no newline", func(t *testing.T) {
		data := []byte("hello")

		result := WC(data)

		if result.Lines != 0 {
			t.Errorf("WC() lines = %d, want 0 (no newline)", result.Lines)
		}

		if result.Words != 1 {
			t.Errorf("WC() words = %d, want 1", result.Words)
		}
	})

	t.Run("unicode", func(t *testing.T) {
		data := []byte("日本語")

		result := WC(data)

		if result.Chars != 3 {
			t.Errorf("WC() chars = %d, want 3", result.Chars)
		}

		if result.Bytes != 9 {
			t.Errorf("WC() bytes = %d, want 9", result.Bytes)
		}
	})
}

func TestWCWithStats(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		data := []byte("test")

		result, err := WCWithStats(data)
		if err != nil {
			t.Fatalf("WCWithStats() error = %v", err)
		}

		if result.Words != 1 {
			t.Errorf("WCWithStats() words = %d, want 1", result.Words)
		}
	})

	t.Run("empty", func(t *testing.T) {
		data := []byte("")

		result, err := WCWithStats(data)
		if err != nil {
			t.Fatalf("WCWithStats() error = %v", err)
		}

		if result.Words != 0 || result.Lines != 0 {
			t.Errorf("WCWithStats() empty = %+v", result)
		}
	})
}
