package grep

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGrep(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grep_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("basic pattern match", func(t *testing.T) {
		file := filepath.Join(tmpDir, "basic.txt")
		content := "hello world\nfoo bar\nhello again"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, "hello", []string{file}, GrepOptions{})
		if err != nil {
			t.Fatalf("RunGrep() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Errorf("RunGrep() got %d matches, want 2", len(lines))
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		file := filepath.Join(tmpDir, "case.txt")
		content := "Hello World\nHELLO AGAIN\nhello there"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, "hello", []string{file}, GrepOptions{IgnoreCase: true})
		if err != nil {
			t.Fatalf("RunGrep() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Errorf("RunGrep() with IgnoreCase got %d matches, want 3", len(lines))
		}
	})

	t.Run("invert match", func(t *testing.T) {
		file := filepath.Join(tmpDir, "invert.txt")
		content := "line1\nmatch\nline2\nmatch again"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, "match", []string{file}, GrepOptions{InvertMatch: true})
		if err != nil {
			t.Fatalf("RunGrep() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Errorf("RunGrep() with InvertMatch got %d lines, want 2", len(lines))
		}
	})

	t.Run("line numbers", func(t *testing.T) {
		file := filepath.Join(tmpDir, "linenum.txt")
		content := "no match\nmatch here\nno match\nmatch again"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, "match here", []string{file}, GrepOptions{LineNumber: true})
		if err != nil {
			t.Fatalf("RunGrep() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "2") {
			t.Errorf("RunGrep() LineNumber should show line numbers: %v", output)
		}
	})

	t.Run("count matches", func(t *testing.T) {
		file := filepath.Join(tmpDir, "count.txt")
		content := "match1\nmatch2\nmatch3\nno match"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, "match[0-9]", []string{file}, GrepOptions{Count: true})
		if err != nil {
			t.Fatalf("RunGrep() error = %v", err)
		}

		if !strings.Contains(buf.String(), "3") {
			t.Errorf("RunGrep() Count should show 3: %v", buf.String())
		}
	})

	t.Run("files with matches", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "has_match.txt")
		file2 := filepath.Join(tmpDir, "no_match.txt")

		if err := os.WriteFile(file1, []byte("has pattern"), 0644); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(file2, []byte("nothing here"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, "pattern", []string{file1, file2}, GrepOptions{FilesWithMatch: true})
		if err != nil {
			t.Fatalf("RunGrep() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if !strings.Contains(output, "has_match.txt") {
			t.Errorf("RunGrep() FilesWithMatch should list matching file: %v", output)
		}

		if strings.Contains(output, "no_match.txt") {
			t.Errorf("RunGrep() FilesWithMatch should not list non-matching file: %v", output)
		}
	})

	t.Run("word regexp", func(t *testing.T) {
		file := filepath.Join(tmpDir, "word.txt")
		content := "test testing tester\nthe test passed"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, "test", []string{file}, GrepOptions{WordRegexp: true})
		if err != nil {
			t.Fatalf("RunGrep() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		// Both lines contain "test" as a whole word
		if len(lines) != 2 {
			t.Errorf("RunGrep() WordRegexp got %d matches, want 2", len(lines))
		}
	})

	t.Run("fixed strings", func(t *testing.T) {
		file := filepath.Join(tmpDir, "fixed.txt")
		content := "foo.bar\nfoo bar\nfooXbar"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, "foo.bar", []string{file}, GrepOptions{FixedStrings: true})
		if err != nil {
			t.Fatalf("RunGrep() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 1 {
			t.Errorf("RunGrep() FixedStrings should match only literal: %v", buf.String())
		}
	})

	t.Run("max count", func(t *testing.T) {
		file := filepath.Join(tmpDir, "maxcount.txt")
		content := "match1\nmatch2\nmatch3\nmatch4\nmatch5"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, "match", []string{file}, GrepOptions{MaxCount: 2})
		if err != nil {
			t.Fatalf("RunGrep() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Errorf("RunGrep() MaxCount got %d matches, want 2", len(lines))
		}
	})

	t.Run("no match returns no error", func(t *testing.T) {
		file := filepath.Join(tmpDir, "nomatch.txt")
		content := "nothing matches here"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, "xyz123", []string{file}, GrepOptions{})
		// Non-quiet mode should not return error for no matches
		if err != nil {
			t.Errorf("RunGrep() should not error on no match: %v", err)
		}
	})

	t.Run("empty pattern", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunGrep(&buf, "", []string{"file.txt"}, GrepOptions{})
		if err == nil {
			t.Error("RunGrep() should return error for empty pattern")
		}
	})
}

func TestGrep(t *testing.T) {
	t.Run("simple grep", func(t *testing.T) {
		lines := []string{"hello world", "foo bar", "hello again"}

		result := Grep(lines, "hello")
		if len(result) != 2 {
			t.Errorf("Grep() got %d matches, want 2", len(result))
		}
	})

	t.Run("no matches", func(t *testing.T) {
		lines := []string{"hello", "world"}

		result := Grep(lines, "xyz")
		if len(result) != 0 {
			t.Errorf("Grep() got %d matches, want 0", len(result))
		}
	})
}
