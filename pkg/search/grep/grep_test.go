package grep

import (
	"testing"
)

func TestSearch(t *testing.T) {
	t.Run("simple grep", func(t *testing.T) {
		lines := []string{"hello world", "foo bar", "hello again"}

		result := Search(lines, "hello")
		if len(result) != 2 {
			t.Errorf("Search() got %d matches, want 2", len(result))
		}
	})

	t.Run("no matches", func(t *testing.T) {
		lines := []string{"hello", "world"}

		result := Search(lines, "xyz")
		if len(result) != 0 {
			t.Errorf("Search() got %d matches, want 0", len(result))
		}
	})

	t.Run("all match", func(t *testing.T) {
		lines := []string{"abc", "abcd", "abcde"}

		result := Search(lines, "abc")
		if len(result) != 3 {
			t.Errorf("Search() got %d matches, want 3", len(result))
		}
	})

	t.Run("empty lines", func(t *testing.T) {
		lines := []string{"", "hello", ""}

		result := Search(lines, "hello")
		if len(result) != 1 {
			t.Errorf("Search() got %d matches, want 1", len(result))
		}
	})

	t.Run("empty input", func(t *testing.T) {
		var lines []string

		result := Search(lines, "pattern")
		if len(result) != 0 {
			t.Errorf("Search() empty input should return empty")
		}
	})
}

func TestSearchWithOptions(t *testing.T) {
	t.Run("case insensitive", func(t *testing.T) {
		lines := []string{"Hello", "HELLO", "hello"}

		result := SearchWithOptions(lines, "hello", WithIgnoreCase())
		if len(result) != 3 {
			t.Errorf("SearchWithOptions() IgnoreCase got %d matches, want 3", len(result))
		}
	})

	t.Run("invert match", func(t *testing.T) {
		lines := []string{"match", "no", "match", "no"}

		result := SearchWithOptions(lines, "match", WithInvertMatch())
		if len(result) != 2 {
			t.Errorf("SearchWithOptions() InvertMatch got %d matches, want 2", len(result))
		}
	})

	t.Run("case insensitive invert", func(t *testing.T) {
		lines := []string{"Match", "no", "MATCH", "no"}

		result := SearchWithOptions(lines, "match", WithIgnoreCase(), WithInvertMatch())
		if len(result) != 2 {
			t.Errorf("SearchWithOptions() combined got %d matches, want 2", len(result))
		}
	})

	t.Run("regex pattern", func(t *testing.T) {
		lines := []string{"test1", "test2", "notest"}

		result := SearchWithOptions(lines, "test[0-9]")
		if len(result) != 2 {
			t.Errorf("SearchWithOptions() regex got %d matches, want 2", len(result))
		}
	})

	t.Run("fixed strings", func(t *testing.T) {
		lines := []string{"a+b", "a*b", "ab"}

		result := SearchWithOptions(lines, "a+b", WithFixedStrings())
		if len(result) != 1 || result[0] != "a+b" {
			t.Errorf("SearchWithOptions() FixedStrings got %v, want [a+b]", result)
		}
	})

	t.Run("word regexp", func(t *testing.T) {
		lines := []string{"testing", "tester", "the test passed"}

		result := SearchWithOptions(lines, "test", WithWordRegexp())
		if len(result) != 1 {
			t.Errorf("SearchWithOptions() WordRegexp got %d matches, want 1", len(result))
		}
	})

	t.Run("line regexp", func(t *testing.T) {
		lines := []string{"exact", "not exact", "exact match"}

		result := SearchWithOptions(lines, "exact", WithLineRegexp())
		if len(result) != 1 || result[0] != "exact" {
			t.Errorf("SearchWithOptions() LineRegexp got %v, want [exact]", result)
		}
	})
}

func TestSearchWithOptionsStruct(t *testing.T) {
	t.Run("case insensitive", func(t *testing.T) {
		lines := []string{"Hello", "HELLO", "hello"}

		result := SearchWithOptionsStruct(lines, "hello", Options{IgnoreCase: true})
		if len(result) != 3 {
			t.Errorf("SearchWithOptionsStruct() got %d matches, want 3", len(result))
		}
	})

	t.Run("invert match", func(t *testing.T) {
		lines := []string{"match", "no", "match", "no"}

		result := SearchWithOptionsStruct(lines, "match", Options{InvertMatch: true})
		if len(result) != 2 {
			t.Errorf("SearchWithOptionsStruct() got %d matches, want 2", len(result))
		}
	})
}

func TestCompilePattern(t *testing.T) {
	t.Run("simple pattern", func(t *testing.T) {
		re, err := CompilePattern("hello", Options{})
		if err != nil {
			t.Fatalf("CompilePattern() error = %v", err)
		}

		if !re.MatchString("hello world") {
			t.Error("CompilePattern() should match 'hello world'")
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		re, err := CompilePattern("hello", Options{IgnoreCase: true})
		if err != nil {
			t.Fatalf("CompilePattern() error = %v", err)
		}

		if !re.MatchString("HELLO") {
			t.Error("CompilePattern() case insensitive should match 'HELLO'")
		}
	})

	t.Run("invalid pattern", func(t *testing.T) {
		_, err := CompilePattern("[invalid", Options{})
		if err == nil {
			t.Error("CompilePattern() should error on invalid regex")
		}
	})

	t.Run("fixed strings escapes", func(t *testing.T) {
		re, err := CompilePattern("a+b", Options{FixedStrings: true})
		if err != nil {
			t.Fatalf("CompilePattern() error = %v", err)
		}

		if re.MatchString("aaab") {
			t.Error("CompilePattern() fixed strings should not match regex")
		}

		if !re.MatchString("a+b") {
			t.Error("CompilePattern() fixed strings should match literal")
		}
	})
}
