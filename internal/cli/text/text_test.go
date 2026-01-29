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

		err := RunSort(&buf, nil, []string{file}, SortOptions{})
		if err != nil {
			t.Fatalf("RunSort() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if lines[0] != "apple" || lines[1] != "banana" || lines[2] != "cherry" {
			t.Errorf("RunSort() = %v, want sorted", lines)
		}
	})

	t.Run("already sorted", func(t *testing.T) {
		file := filepath.Join(tmpDir, "already.txt")
		content := "apple\nbanana\ncherry"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunSort(&buf, nil, []string{file}, SortOptions{})

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if lines[0] != "apple" || lines[1] != "banana" || lines[2] != "cherry" {
			t.Errorf("RunSort() already sorted = %v", lines)
		}
	})

	t.Run("reverse sort", func(t *testing.T) {
		file := filepath.Join(tmpDir, "reverse.txt")
		content := "apple\nbanana\ncherry"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunSort(&buf, nil, []string{file}, SortOptions{Reverse: true})
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

		err := RunSort(&buf, nil, []string{file}, SortOptions{Numeric: true})
		if err != nil {
			t.Fatalf("RunSort() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if lines[0] != "1" || lines[1] != "2" || lines[2] != "10" || lines[3] != "20" {
			t.Errorf("RunSort() numeric = %v", lines)
		}
	})

	t.Run("numeric sort with negatives", func(t *testing.T) {
		file := filepath.Join(tmpDir, "negative.txt")
		content := "-5\n10\n-10\n5\n0"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunSort(&buf, nil, []string{file}, SortOptions{Numeric: true})

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		// Should be: -10, -5, 0, 5, 10
		if len(lines) != 5 {
			t.Errorf("RunSort() numeric negatives got %d lines", len(lines))
		}
	})

	t.Run("unique sort", func(t *testing.T) {
		file := filepath.Join(tmpDir, "unique.txt")
		content := "apple\nbanana\napple\ncherry\nbanana"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunSort(&buf, nil, []string{file}, SortOptions{Unique: true})
		if err != nil {
			t.Fatalf("RunSort() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Errorf("RunSort() unique got %d lines, want 3", len(lines))
		}
	})

	t.Run("unique with all duplicates", func(t *testing.T) {
		file := filepath.Join(tmpDir, "all_dup.txt")
		content := "same\n"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunSort(&buf, nil, []string{file}, SortOptions{Unique: true})

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 1 {
			t.Errorf("RunSort() all duplicates got %d lines, want 1", len(lines))
		}
	})

	t.Run("case insensitive sort", func(t *testing.T) {
		file := filepath.Join(tmpDir, "case.txt")
		content := "Banana\napple\nCherry"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunSort(&buf, nil, []string{file}, SortOptions{IgnoreCase: true})
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

		err := RunSort(&buf, nil, []string{file}, SortOptions{IgnoreLeading: true})
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

		err := RunSort(&buf, nil, []string{file}, SortOptions{Stable: true})
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

		err := RunSort(&buf, nil, []string{file}, SortOptions{Check: true})
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

		err := RunSort(&buf, nil, []string{file}, SortOptions{Check: true})
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

		err := RunSort(&buf, nil, []string{file}, SortOptions{})
		if err != nil {
			t.Fatalf("RunSort() error = %v", err)
		}

		if buf.Len() != 0 {
			t.Errorf("RunSort() empty file should be empty: %v", buf.String())
		}
	})

	t.Run("single line", func(t *testing.T) {
		file := filepath.Join(tmpDir, "single.txt")
		content := "only"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunSort(&buf, nil, []string{file}, SortOptions{})

		if strings.TrimSpace(buf.String()) != "only" {
			t.Errorf("RunSort() single line = %v", buf.String())
		}
	})

	t.Run("unicode content", func(t *testing.T) {
		file := filepath.Join(tmpDir, "unicode.txt")
		content := "日本語\n中文\n한국어"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		err := RunSort(&buf, nil, []string{file}, SortOptions{})
		if err != nil {
			t.Fatalf("RunSort() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Errorf("RunSort() unicode got %d lines", len(lines))
		}
	})

	t.Run("numeric reverse", func(t *testing.T) {
		file := filepath.Join(tmpDir, "numrev.txt")
		content := "1\n10\n2\n20"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunSort(&buf, nil, []string{file}, SortOptions{Numeric: true, Reverse: true})

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if lines[0] != "20" || lines[3] != "1" {
			t.Errorf("RunSort() numeric reverse = %v", lines)
		}
	})

	t.Run("output ends with newline", func(t *testing.T) {
		file := filepath.Join(tmpDir, "newline.txt")
		content := "b\na"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunSort(&buf, nil, []string{file}, SortOptions{})

		if !strings.HasSuffix(buf.String(), "\n") {
			t.Error("RunSort() output should end with newline")
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
		content := "apple\nbanana\ncherry"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunUniq(&buf, nil, []string{file}, UniqOptions{})
		if err != nil {
			t.Fatalf("RunUniq() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Errorf("RunUniq() got %d lines, want 3", len(lines))
		}
	})

	t.Run("no consecutive duplicates", func(t *testing.T) {
		file := filepath.Join(tmpDir, "no_dup.txt")
		content := "a\nb\nc\na\nb"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunUniq(&buf, nil, []string{file}, UniqOptions{})

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		// uniq only removes consecutive duplicates
		if len(lines) != 5 {
			t.Errorf("RunUniq() no consecutive got %d lines, want 5", len(lines))
		}
	})

	t.Run("count occurrences", func(t *testing.T) {
		file := filepath.Join(tmpDir, "count.txt")
		// uniq counts consecutive duplicates
		content := "apple\nbanana\n"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunUniq(&buf, nil, []string{file}, UniqOptions{Count: true})
		if err != nil {
			t.Fatalf("RunUniq() error = %v", err)
		}

		output := buf.String()
		// Should show count 2 for apple, 3 for banana, 1 for cherry
		if !strings.Contains(output, "2") || !strings.Contains(output, "3") {
			t.Errorf("RunUniq() count should show numbers 2 and 3: %v", output)
		}
	})

	t.Run("count format", func(t *testing.T) {
		file := filepath.Join(tmpDir, "count_fmt.txt")
		content := "a\n"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunUniq(&buf, nil, []string{file}, UniqOptions{Count: true})

		output := strings.TrimSpace(buf.String())
		// Should show count 3 for 'a'
		if !strings.Contains(output, "3") {
			t.Errorf("RunUniq() count format = %v, want count of 3", output)
		}
	})

	t.Run("repeated only", func(t *testing.T) {
		file := filepath.Join(tmpDir, "repeated.txt")
		// apple appears twice (consecutive), banana once, cherry twice (consecutive)
		content := "apple\nbanana\ncherry"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunUniq(&buf, nil, []string{file}, UniqOptions{Repeated: true})
		if err != nil {
			t.Fatalf("RunUniq() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		// apple and cherry are repeated (appear >1 consecutively)
		if len(lines) != 2 {
			t.Errorf("RunUniq() repeated got %d lines, want 2: %v", len(lines), lines)
		}
	})

	t.Run("unique only", func(t *testing.T) {
		file := filepath.Join(tmpDir, "unique.txt")
		// apple appears twice, banana once, cherry twice
		content := "apple\nbanana\ncherry"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunUniq(&buf, nil, []string{file}, UniqOptions{Unique: true})
		if err != nil {
			t.Fatalf("RunUniq() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		// Only banana is unique (appears exactly once)
		if len(lines) != 1 || lines[0] != "banana" {
			t.Errorf("RunUniq() unique = %v, want [banana]", lines)
		}
	})

	t.Run("unique with all duplicates", func(t *testing.T) {
		file := filepath.Join(tmpDir, "all_dup.txt")
		// All lines are duplicated consecutively
		content := "same\nsame\n"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunUniq(&buf, nil, []string{file}, UniqOptions{Unique: true})

		output := strings.TrimSpace(buf.String())
		// No unique lines (same appears twice)
		if output != "" {
			t.Errorf("RunUniq() all duplicates unique = %v, want empty", output)
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		file := filepath.Join(tmpDir, "case.txt")
		content := "Apple\napple\nAPPLE\nbanana"

		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunUniq(&buf, nil, []string{file}, UniqOptions{IgnoreCase: true})
		if err != nil {
			t.Fatalf("RunUniq() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		// Case insensitive: Apple group and banana
		if len(lines) != 2 {
			t.Errorf("RunUniq() case insensitive got %d lines, want 2", len(lines))
		}
	})

	t.Run("empty file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "empty.txt")

		_ = os.WriteFile(file, []byte(""), 0644)

		var buf bytes.Buffer

		err := RunUniq(&buf, nil, []string{file}, UniqOptions{})
		if err != nil {
			t.Fatalf("RunUniq() error = %v", err)
		}

		if buf.Len() != 0 {
			t.Errorf("RunUniq() empty file should be empty")
		}
	})

	t.Run("single line", func(t *testing.T) {
		file := filepath.Join(tmpDir, "single.txt")
		content := "only"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunUniq(&buf, nil, []string{file}, UniqOptions{})

		if strings.TrimSpace(buf.String()) != "only" {
			t.Errorf("RunUniq() single = %v", buf.String())
		}
	})

	t.Run("output ends with newline", func(t *testing.T) {
		file := filepath.Join(tmpDir, "newline.txt")
		content := "a\na"

		_ = os.WriteFile(file, []byte(content), 0644)

		var buf bytes.Buffer

		_ = RunUniq(&buf, nil, []string{file}, UniqOptions{})

		if !strings.HasSuffix(buf.String(), "\n") {
			t.Error("RunUniq() output should end with newline")
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

	t.Run("already sorted", func(t *testing.T) {
		lines := []string{"a", "b", "c"}

		Sort(lines)

		if lines[0] != "a" || lines[1] != "b" || lines[2] != "c" {
			t.Errorf("Sort() already sorted = %v", lines)
		}
	})

	t.Run("reverse sorted", func(t *testing.T) {
		lines := []string{"c", "b", "a"}

		Sort(lines)

		if lines[0] != "a" || lines[1] != "b" || lines[2] != "c" {
			t.Errorf("Sort() reverse = %v", lines)
		}
	})

	t.Run("single element", func(t *testing.T) {
		lines := []string{"only"}

		Sort(lines)

		if lines[0] != "only" {
			t.Errorf("Sort() single = %v", lines)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		var lines []string

		Sort(lines) // Should not panic

		if len(lines) != 0 {
			t.Errorf("Sort() empty should remain empty")
		}
	})

	t.Run("duplicates", func(t *testing.T) {
		lines := []string{"b", "a", "b", "a"}

		Sort(lines)

		if lines[0] != "a" || lines[1] != "a" || lines[2] != "b" || lines[3] != "b" {
			t.Errorf("Sort() duplicates = %v", lines)
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

	t.Run("empty slice", func(t *testing.T) {
		var lines []string

		result := Uniq(lines)

		if len(result) != 0 {
			t.Errorf("Uniq() empty got %d, want 0", len(result))
		}
	})

	t.Run("single element", func(t *testing.T) {
		lines := []string{"only"}

		result := Uniq(lines)

		if len(result) != 1 || result[0] != "only" {
			t.Errorf("Uniq() single = %v", result)
		}
	})

	t.Run("non-consecutive duplicates", func(t *testing.T) {
		lines := []string{"a", "b", "a", "b"}

		result := Uniq(lines)

		// This implementation removes ALL duplicates (not just consecutive like Unix uniq)
		if len(result) != 2 {
			t.Errorf("Uniq() non-consecutive got %d, want 2", len(result))
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

	t.Run("no whitespace", func(t *testing.T) {
		lines := []string{"hello", "world"}

		result := TrimLines(lines)

		if result[0] != "hello" || result[1] != "world" {
			t.Errorf("TrimLines() no whitespace = %v", result)
		}
	})

	t.Run("empty lines", func(t *testing.T) {
		lines := []string{"", "  ", "\t"}

		result := TrimLines(lines)

		if result[0] != "" || result[1] != "" || result[2] != "" {
			t.Errorf("TrimLines() empty = %v", result)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		var lines []string

		result := TrimLines(lines)

		if len(result) != 0 {
			t.Errorf("TrimLines() empty slice = %v", result)
		}
	})

	t.Run("mixed content", func(t *testing.T) {
		lines := []string{"  a", "b  ", "  c  ", "d"}

		result := TrimLines(lines)

		if result[0] != "a" || result[1] != "b" || result[2] != "c" || result[3] != "d" {
			t.Errorf("TrimLines() mixed = %v", result)
		}
	})
}
