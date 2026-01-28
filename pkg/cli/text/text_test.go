package text

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSort(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sort_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("basic sort", func(t *testing.T) {
		file := filepath.Join(tmpDir, "basic.txt")
		content := "banana\napple\ncherry"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunSort(&buf, []string{file}, SortOptions{})
		if err != nil {
			t.Fatalf("RunSort() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if lines[0] != "apple" || lines[1] != "banana" || lines[2] != "cherry" {
			t.Errorf("RunSort() = %v, want sorted", lines)
		}
	})

	t.Run("reverse sort", func(t *testing.T) {
		file := filepath.Join(tmpDir, "reverse.txt")
		content := "apple\nbanana\ncherry"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunSort(&buf, []string{file}, SortOptions{Reverse: true})
		if err != nil {
			t.Fatalf("RunSort() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if lines[0] != "cherry" || lines[2] != "apple" {
			t.Errorf("RunSort() reverse = %v", lines)
		}
	})

	t.Run("numeric sort", func(t *testing.T) {
		file := filepath.Join(tmpDir, "numeric.txt")
		content := "10\n2\n1\n20"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunSort(&buf, []string{file}, SortOptions{Numeric: true})
		if err != nil {
			t.Fatalf("RunSort() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if lines[0] != "1" || lines[1] != "2" || lines[2] != "10" || lines[3] != "20" {
			t.Errorf("RunSort() numeric = %v", lines)
		}
	})

	t.Run("unique sort", func(t *testing.T) {
		file := filepath.Join(tmpDir, "unique.txt")
		content := "apple\nbanana\napple\ncherry\nbanana"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunSort(&buf, []string{file}, SortOptions{Unique: true})
		if err != nil {
			t.Fatalf("RunSort() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Errorf("RunSort() unique got %d lines, want 3", len(lines))
		}
	})

	t.Run("case insensitive sort", func(t *testing.T) {
		file := filepath.Join(tmpDir, "case.txt")
		content := "Banana\napple\nCherry"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunSort(&buf, []string{file}, SortOptions{IgnoreCase: true})
		if err != nil {
			t.Fatalf("RunSort() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		// Case insensitive: apple, Banana, Cherry
		if !strings.EqualFold(lines[0], "apple") {
			t.Errorf("RunSort() case insensitive first = %v, want apple", lines[0])
		}
	})

	t.Run("ignore leading blanks", func(t *testing.T) {
		file := filepath.Join(tmpDir, "blanks.txt")
		content := "  banana\napple\n   cherry"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunSort(&buf, []string{file}, SortOptions{IgnoreLeading: true})
		if err != nil {
			t.Fatalf("RunSort() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if !strings.Contains(lines[0], "apple") {
			t.Errorf("RunSort() ignore blanks first = %v, want apple", lines[0])
		}
	})

	t.Run("stable sort", func(t *testing.T) {
		file := filepath.Join(tmpDir, "stable.txt")
		content := "b\na\nb\na"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunSort(&buf, []string{file}, SortOptions{Stable: true})
		if err != nil {
			t.Fatalf("RunSort() error = %v", err)
		}

		// Just verify it doesn't error
		if buf.Len() == 0 {
			t.Error("RunSort() stable should produce output")
		}
	})

	t.Run("check sorted", func(t *testing.T) {
		file := filepath.Join(tmpDir, "check.txt")
		content := "apple\nbanana\ncherry"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunSort(&buf, []string{file}, SortOptions{Check: true})
		// Should not error for sorted input
		if err != nil {
			t.Errorf("RunSort() check sorted file error = %v", err)
		}
	})

	t.Run("check unsorted fails", func(t *testing.T) {
		file := filepath.Join(tmpDir, "unsorted.txt")
		content := "cherry\napple\nbanana"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunSort(&buf, []string{file}, SortOptions{Check: true})
		// Should error for unsorted input
		if err == nil {
			t.Error("RunSort() check unsorted should error")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "empty.txt")

		if err := os.WriteFile(file, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunSort(&buf, []string{file}, SortOptions{})
		if err != nil {
			t.Fatalf("RunSort() error = %v", err)
		}

		if buf.Len() != 0 {
			t.Errorf("RunSort() empty file should be empty: %v", buf.String())
		}
	})
}

func TestRunUniq(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "uniq_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("basic uniq", func(t *testing.T) {
		file := filepath.Join(tmpDir, "basic.txt")
		content := "apple\napple\nbanana\nbanana\nbanana\ncherry"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunUniq(&buf, []string{file}, UniqOptions{})
		if err != nil {
			t.Fatalf("RunUniq() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Errorf("RunUniq() got %d lines, want 3", len(lines))
		}
	})

	t.Run("count occurrences", func(t *testing.T) {
		file := filepath.Join(tmpDir, "count.txt")
		content := "apple\napple\nbanana\ncherry\ncherry\ncherry"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunUniq(&buf, []string{file}, UniqOptions{Count: true})
		if err != nil {
			t.Fatalf("RunUniq() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "2") && !strings.Contains(output, "3") {
			t.Errorf("RunUniq() count should show numbers: %v", output)
		}
	})

	t.Run("repeated only", func(t *testing.T) {
		file := filepath.Join(tmpDir, "repeated.txt")
		content := "apple\napple\nbanana\ncherry\ncherry"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunUniq(&buf, []string{file}, UniqOptions{Repeated: true})
		if err != nil {
			t.Fatalf("RunUniq() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		// Only apple and cherry are repeated
		if len(lines) != 2 {
			t.Errorf("RunUniq() repeated got %d lines, want 2", len(lines))
		}
	})

	t.Run("unique only", func(t *testing.T) {
		file := filepath.Join(tmpDir, "unique.txt")
		content := "apple\napple\nbanana\ncherry\ncherry"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunUniq(&buf, []string{file}, UniqOptions{Unique: true})
		if err != nil {
			t.Fatalf("RunUniq() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		// Only banana is unique
		if len(lines) != 1 || lines[0] != "banana" {
			t.Errorf("RunUniq() unique = %v, want [banana]", lines)
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		file := filepath.Join(tmpDir, "case.txt")
		content := "Apple\napple\nAPPLE\nbanana"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunUniq(&buf, []string{file}, UniqOptions{IgnoreCase: true})
		if err != nil {
			t.Fatalf("RunUniq() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		// Case insensitive: Apple group and banana
		if len(lines) != 2 {
			t.Errorf("RunUniq() case insensitive got %d lines, want 2", len(lines))
		}
	})
}

func TestSort(t *testing.T) {
	t.Run("simple sort", func(t *testing.T) {
		lines := []string{"cherry", "apple", "banana"}

		Sort(lines)

		if lines[0] != "apple" || lines[1] != "banana" || lines[2] != "cherry" {
			t.Errorf("Sort() = %v", lines)
		}
	})
}

func TestUniq(t *testing.T) {
	t.Run("simple uniq", func(t *testing.T) {
		lines := []string{"apple", "apple", "banana", "cherry", "cherry"}

		result := Uniq(lines)

		if len(result) != 3 {
			t.Errorf("Uniq() got %d, want 3", len(result))
		}
	})

	t.Run("all unique", func(t *testing.T) {
		lines := []string{"a", "b", "c"}

		result := Uniq(lines)

		if len(result) != 3 {
			t.Errorf("Uniq() got %d, want 3", len(result))
		}
	})

	t.Run("all same", func(t *testing.T) {
		lines := []string{"a", "a", "a"}

		result := Uniq(lines)

		if len(result) != 1 {
			t.Errorf("Uniq() got %d, want 1", len(result))
		}
	})
}

func TestTrimLines(t *testing.T) {
	t.Run("trim whitespace", func(t *testing.T) {
		lines := []string{"  hello  ", "\tworld\t", "  foo bar  "}

		result := TrimLines(lines)

		if result[0] != "hello" || result[1] != "world" || result[2] != "foo bar" {
			t.Errorf("TrimLines() = %v", result)
		}
	})
}
