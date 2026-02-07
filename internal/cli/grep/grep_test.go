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

		err := RunGrep(&buf, nil, "hello", []string{file}, GrepOptions{})
		if err != nil {
			t.Fatalf("RunGrep() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Errorf("RunGrep() got %d matches, want 2", len(lines))
		}
	})

	t.Run("regex pattern", func(t *testing.T) {
		file := filepath.Join(tmpDir, "regex.txt")
		content := "test1\ntest2\ntest3\nnotest"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, nil, "test[0-9]", []string{file}, GrepOptions{})
		if err != nil {
			t.Fatalf("RunGrep() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Errorf("RunGrep() regex got %d matches, want 3", len(lines))
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		file := filepath.Join(tmpDir, "case.txt")
		content := "Hello World\nHELLO AGAIN\nhello there"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, nil, "hello", []string{file}, GrepOptions{IgnoreCase: true})
		if err != nil {
			t.Fatalf("RunGrep() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Errorf("RunGrep() with IgnoreCase got %d matches, want 3", len(lines))
		}
	})

	t.Run("case sensitive default", func(t *testing.T) {
		file := filepath.Join(tmpDir, "case_sens.txt")
		content := "Hello\nhello\nHELLO"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, nil, "hello", []string{file}, GrepOptions{})
		if err != nil {
			t.Fatalf("RunGrep() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 1 {
			t.Errorf("RunGrep() case sensitive got %d matches, want 1", len(lines))
		}
	})

	t.Run("invert match", func(t *testing.T) {
		file := filepath.Join(tmpDir, "invert.txt")
		content := "line1\nmatch\nline2\nmatch again"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, nil, "match", []string{file}, GrepOptions{InvertMatch: true})
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
		content := "no match\nfind here\nno match\nagain"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, nil, "here", []string{file}, GrepOptions{LineNumber: true})
		if err != nil {
			t.Fatalf("RunGrep() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "2:") {
			t.Errorf("RunGrep() LineNumber should show line 2: %v", output)
		}
	})

	t.Run("line numbers multiple matches", func(t *testing.T) {
		file := filepath.Join(tmpDir, "linenum_multi.txt")
		content := "match1\nno\nmatch2\nno\nmatch3"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunGrep(&buf, nil, "match", []string{file}, GrepOptions{LineNumber: true})

		output := buf.String()
		if !strings.Contains(output, "1:") || !strings.Contains(output, "3:") || !strings.Contains(output, "5:") {
			t.Errorf("RunGrep() LineNumber should show lines 1, 3, 5: %v", output)
		}
	})

	t.Run("count matches", func(t *testing.T) {
		file := filepath.Join(tmpDir, "count.txt")
		content := "match1\nmatch2\nmatch3\nno match"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, nil, "match[0-9]", []string{file}, GrepOptions{Count: true})
		if err != nil {
			t.Fatalf("RunGrep() error = %v", err)
		}

		if !strings.Contains(buf.String(), "3") {
			t.Errorf("RunGrep() Count should show 3: %v", buf.String())
		}
	})

	t.Run("count zero matches", func(t *testing.T) {
		file := filepath.Join(tmpDir, "count_zero.txt")
		content := "nothing here"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunGrep(&buf, nil, "xyz", []string{file}, GrepOptions{Count: true})

		if !strings.Contains(buf.String(), "0") {
			t.Errorf("RunGrep() Count should show 0: %v", buf.String())
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

		err := RunGrep(&buf, nil, "pattern", []string{file1, file2}, GrepOptions{FilesWithMatch: true})
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

	t.Run("files without matches", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "has_match2.txt")
		file2 := filepath.Join(tmpDir, "no_match2.txt")

		_ = os.WriteFile(file1, []byte("has pattern"), 0644)
		_ = os.WriteFile(file2, []byte("nothing here"), 0644)

		var buf bytes.Buffer

		_ = RunGrep(&buf, nil, "pattern", []string{file1, file2}, GrepOptions{FilesNoMatch: true})

		output := strings.TrimSpace(buf.String())
		if strings.Contains(output, "has_match2.txt") {
			t.Errorf("RunGrep() FilesNoMatch should not list matching file")
		}

		if !strings.Contains(output, "no_match2.txt") {
			t.Errorf("RunGrep() FilesNoMatch should list non-matching file")
		}
	})

	t.Run("word regexp", func(t *testing.T) {
		file := filepath.Join(tmpDir, "word.txt")
		content := "test testing tester\nthe test passed"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, nil, "test", []string{file}, GrepOptions{WordRegexp: true})
		if err != nil {
			t.Fatalf("RunGrep() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		// Both lines contain "test" as a whole word
		if len(lines) != 2 {
			t.Errorf("RunGrep() WordRegexp got %d matches, want 2", len(lines))
		}
	})

	t.Run("word regexp excludes partial", func(t *testing.T) {
		file := filepath.Join(tmpDir, "word_partial.txt")
		content := "testing\ntester\ncontest"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunGrep(&buf, nil, "test", []string{file}, GrepOptions{WordRegexp: true})

		// None of these contain "test" as a whole word
		if strings.TrimSpace(buf.String()) != "" {
			t.Errorf("RunGrep() WordRegexp should not match partial words: %v", buf.String())
		}
	})

	t.Run("line regexp", func(t *testing.T) {
		file := filepath.Join(tmpDir, "line.txt")
		content := "exact\nmatch\nno exact"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunGrep(&buf, nil, "exact", []string{file}, GrepOptions{LineRegexp: true})

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 1 || lines[0] != "exact" {
			t.Errorf("RunGrep() LineRegexp should match only exact line: %v", buf.String())
		}
	})

	t.Run("fixed strings", func(t *testing.T) {
		file := filepath.Join(tmpDir, "fixed.txt")
		content := "foo.bar\nfoo bar\nfooXbar"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, nil, "foo.bar", []string{file}, GrepOptions{FixedStrings: true})
		if err != nil {
			t.Fatalf("RunGrep() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 1 {
			t.Errorf("RunGrep() FixedStrings should match only literal: %v", buf.String())
		}
	})

	t.Run("fixed strings special chars", func(t *testing.T) {
		file := filepath.Join(tmpDir, "fixed_special.txt")
		content := "a+b\na*b\na?b\na[b"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunGrep(&buf, nil, "a+b", []string{file}, GrepOptions{FixedStrings: true})

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 1 || lines[0] != "a+b" {
			t.Errorf("RunGrep() FixedStrings should escape regex chars: %v", buf.String())
		}
	})

	t.Run("max count", func(t *testing.T) {
		file := filepath.Join(tmpDir, "maxcount.txt")
		content := "match1\nmatch2\nmatch3\nmatch4\nmatch5"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, nil, "match", []string{file}, GrepOptions{MaxCount: 2})
		if err != nil {
			t.Fatalf("RunGrep() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Errorf("RunGrep() MaxCount got %d matches, want 2", len(lines))
		}
	})

	t.Run("only matching", func(t *testing.T) {
		file := filepath.Join(tmpDir, "only.txt")
		content := "foo123bar\nfoo456bar"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunGrep(&buf, nil, "[0-9]+", []string{file}, GrepOptions{OnlyMatching: true})

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Errorf("RunGrep() OnlyMatching got %d results", len(lines))
		}

		if lines[0] != "123" || lines[1] != "456" {
			t.Errorf("RunGrep() OnlyMatching should show only matched parts: %v", lines)
		}
	})

	t.Run("multiple files with filename", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")

		_ = os.WriteFile(file1, []byte("match here"), 0644)
		_ = os.WriteFile(file2, []byte("match there"), 0644)

		var buf bytes.Buffer

		_ = RunGrep(&buf, nil, "match", []string{file1, file2}, GrepOptions{})

		output := buf.String()
		// Multiple files should show filenames
		if !strings.Contains(output, "file1.txt") || !strings.Contains(output, "file2.txt") {
			t.Errorf("RunGrep() multiple files should show filenames: %v", output)
		}
	})

	t.Run("single file no filename", func(t *testing.T) {
		file := filepath.Join(tmpDir, "single.txt")

		_ = os.WriteFile(file, []byte("match here"), 0644)

		var buf bytes.Buffer

		_ = RunGrep(&buf, nil, "match", []string{file}, GrepOptions{})

		output := buf.String()
		// Single file should not show filename by default
		if strings.Contains(output, "single.txt") {
			t.Errorf("RunGrep() single file should not show filename by default: %v", output)
		}
	})

	t.Run("with filename option", func(t *testing.T) {
		file := filepath.Join(tmpDir, "withname.txt")

		_ = os.WriteFile(file, []byte("match here"), 0644)

		var buf bytes.Buffer

		_ = RunGrep(&buf, nil, "match", []string{file}, GrepOptions{WithFilename: true})

		output := buf.String()
		if !strings.Contains(output, "withname.txt") {
			t.Errorf("RunGrep() WithFilename should show filename: %v", output)
		}
	})

	t.Run("no filename option", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "noname1.txt")
		file2 := filepath.Join(tmpDir, "noname2.txt")

		_ = os.WriteFile(file1, []byte("match"), 0644)
		_ = os.WriteFile(file2, []byte("match"), 0644)

		var buf bytes.Buffer

		_ = RunGrep(&buf, nil, "match", []string{file1, file2}, GrepOptions{NoFilename: true})

		output := buf.String()
		if strings.Contains(output, "noname1.txt") || strings.Contains(output, "noname2.txt") {
			t.Errorf("RunGrep() NoFilename should hide filenames: %v", output)
		}
	})

	t.Run("context lines", func(t *testing.T) {
		file := filepath.Join(tmpDir, "context.txt")
		content := "line1\nline2\nmatch\nline4\nline5"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunGrep(&buf, nil, "match", []string{file}, GrepOptions{Context: 1})

		output := buf.String()
		if !strings.Contains(output, "line2") || !strings.Contains(output, "line4") {
			t.Errorf("RunGrep() Context should show surrounding lines: %v", output)
		}
	})

	t.Run("before context", func(t *testing.T) {
		file := filepath.Join(tmpDir, "before.txt")
		content := "line1\nline2\nmatch\nline4"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunGrep(&buf, nil, "match", []string{file}, GrepOptions{BeforeContext: 2})

		output := buf.String()
		if !strings.Contains(output, "line1") || !strings.Contains(output, "line2") {
			t.Errorf("RunGrep() BeforeContext should show lines before: %v", output)
		}
	})

	t.Run("after context", func(t *testing.T) {
		file := filepath.Join(tmpDir, "after.txt")
		content := "match\nline2\nline3\nline4"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunGrep(&buf, nil, "match", []string{file}, GrepOptions{AfterContext: 2})

		output := buf.String()
		if !strings.Contains(output, "line2") || !strings.Contains(output, "line3") {
			t.Errorf("RunGrep() AfterContext should show lines after: %v", output)
		}
	})

	t.Run("no match returns error", func(t *testing.T) {
		file := filepath.Join(tmpDir, "nomatch.txt")
		content := "nothing matches here"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunGrep(&buf, nil, "xyz123", []string{file}, GrepOptions{})
		// Like GNU grep, return error (exit code 1) when no matches found
		if err == nil {
			t.Errorf("RunGrep() should return error on no match")
		}
	})

	t.Run("quiet mode with match", func(t *testing.T) {
		file := filepath.Join(tmpDir, "quiet_match.txt")

		_ = os.WriteFile(file, []byte("has match"), 0644)

		var buf bytes.Buffer

		err := RunGrep(&buf, nil, "match", []string{file}, GrepOptions{Quiet: true})
		if err != nil {
			t.Errorf("RunGrep() quiet with match should not error: %v", err)
		}

		if buf.Len() != 0 {
			t.Errorf("RunGrep() quiet mode should produce no output: %v", buf.String())
		}
	})

	t.Run("quiet mode no match", func(t *testing.T) {
		file := filepath.Join(tmpDir, "quiet_nomatch.txt")

		_ = os.WriteFile(file, []byte("nothing"), 0644)

		var buf bytes.Buffer

		err := RunGrep(&buf, nil, "xyz", []string{file}, GrepOptions{Quiet: true})
		if err == nil {
			t.Errorf("RunGrep() quiet without match should error")
		}
	})

	t.Run("empty pattern", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunGrep(&buf, nil, "", []string{"file.txt"}, GrepOptions{})
		if err == nil {
			t.Error("RunGrep() should return error for empty pattern")
		}
	})

	t.Run("invalid regex pattern", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunGrep(&buf, nil, "[invalid", []string{"file.txt"}, GrepOptions{})
		if err == nil {
			t.Error("RunGrep() should return error for invalid regex")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		// Nonexistent file should return an error
		err := RunGrep(&buf, nil, "pattern", []string{"/nonexistent/file.txt"}, GrepOptions{})
		if err == nil {
			t.Errorf("RunGrep() should return error for nonexistent file")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "empty.txt")

		_ = os.WriteFile(file, []byte(""), 0644)

		var buf bytes.Buffer

		err := RunGrep(&buf, nil, "pattern", []string{file}, GrepOptions{})
		// Empty file has no matches, so should return "no match" error like GNU grep
		if err == nil {
			t.Fatalf("RunGrep() should return error for empty file with no matches")
		}

		if buf.Len() != 0 {
			t.Errorf("RunGrep() empty file should produce no output")
		}
	})

	t.Run("unicode pattern", func(t *testing.T) {
		file := filepath.Join(tmpDir, "unicode.txt")
		content := "Hello 世界\n你好 World\nこんにちは"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunGrep(&buf, nil, "世界", []string{file}, GrepOptions{})

		if !strings.Contains(buf.String(), "世界") {
			t.Errorf("RunGrep() should match unicode: %v", buf.String())
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

	t.Run("all match", func(t *testing.T) {
		lines := []string{"abc", "abcd", "abcde"}

		result := Grep(lines, "abc")
		if len(result) != 3 {
			t.Errorf("Grep() got %d matches, want 3", len(result))
		}
	})

	t.Run("empty lines", func(t *testing.T) {
		lines := []string{"", "hello", ""}

		result := Grep(lines, "hello")
		if len(result) != 1 {
			t.Errorf("Grep() got %d matches, want 1", len(result))
		}
	})

	t.Run("empty input", func(t *testing.T) {
		var lines []string

		result := Grep(lines, "pattern")
		if len(result) != 0 {
			t.Errorf("Grep() empty input should return empty")
		}
	})
}

func TestGrepWithOptions(t *testing.T) {
	t.Run("case insensitive", func(t *testing.T) {
		lines := []string{"Hello", "HELLO", "hello"}

		result := GrepWithOptions(lines, "hello", GrepOptions{IgnoreCase: true})
		if len(result) != 3 {
			t.Errorf("GrepWithOptions() IgnoreCase got %d matches, want 3", len(result))
		}
	})

	t.Run("invert match", func(t *testing.T) {
		lines := []string{"match", "no", "match", "no"}

		result := GrepWithOptions(lines, "match", GrepOptions{InvertMatch: true})
		if len(result) != 2 {
			t.Errorf("GrepWithOptions() InvertMatch got %d matches, want 2", len(result))
		}
	})

	t.Run("case insensitive invert", func(t *testing.T) {
		lines := []string{"Match", "no", "MATCH", "no"}

		result := GrepWithOptions(lines, "match", GrepOptions{IgnoreCase: true, InvertMatch: true})
		if len(result) != 2 {
			t.Errorf("GrepWithOptions() combined got %d matches, want 2", len(result))
		}
	})

	t.Run("regex pattern", func(t *testing.T) {
		lines := []string{"test1", "test2", "notest"}

		result := GrepWithOptions(lines, "test[0-9]", GrepOptions{})
		if len(result) != 2 {
			t.Errorf("GrepWithOptions() regex got %d matches, want 2", len(result))
		}
	})
}
