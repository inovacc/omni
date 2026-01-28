package sort

import (
	"testing"
)

func TestPipeline(t *testing.T) {
	t.Run("no operations", func(t *testing.T) {
		input := []string{"c", "a", "b"}
		result := Pipeline(input, "", false, false)

		if len(result) != 3 {
			t.Errorf("Pipeline() = %v, want 3 items", result)
		}
		// Order should be unchanged
		if result[0] != "c" || result[1] != "a" || result[2] != "b" {
			t.Errorf("Pipeline() = %v, want [c a b]", result)
		}
	})

	t.Run("sort only", func(t *testing.T) {
		input := []string{"c", "a", "b"}
		result := Pipeline(input, "", true, false)

		if result[0] != "a" || result[1] != "b" || result[2] != "c" {
			t.Errorf("Pipeline() sort = %v, want [a b c]", result)
		}
	})

	t.Run("grep only", func(t *testing.T) {
		input := []string{"apple", "banana", "apricot", "berry"}
		result := Pipeline(input, "ap", false, false)

		if len(result) != 2 {
			t.Errorf("Pipeline() grep = %v, want 2 items", result)
		}
	})

	t.Run("uniq only", func(t *testing.T) {
		input := []string{"a", "b", "a", "c", "b"}
		result := Pipeline(input, "", false, true)

		if len(result) != 3 {
			t.Errorf("Pipeline() uniq = %v, want 3 unique items", result)
		}
	})

	t.Run("grep and sort", func(t *testing.T) {
		input := []string{"cat", "car", "dog", "cap"}
		result := Pipeline(input, "ca", true, false)

		if len(result) != 3 {
			t.Errorf("Pipeline() grep+sort = %v, want 3 items", result)
		}
		if result[0] != "cap" || result[1] != "car" || result[2] != "cat" {
			t.Errorf("Pipeline() grep+sort = %v, want [cap car cat]", result)
		}
	})

	t.Run("sort and uniq", func(t *testing.T) {
		input := []string{"b", "a", "b", "c", "a"}
		result := Pipeline(input, "", true, true)

		if len(result) != 3 {
			t.Errorf("Pipeline() sort+uniq = %v, want 3 items", result)
		}
		if result[0] != "a" || result[1] != "b" || result[2] != "c" {
			t.Errorf("Pipeline() sort+uniq = %v, want [a b c]", result)
		}
	})

	t.Run("all operations", func(t *testing.T) {
		input := []string{"hello", "world", "hello", "help", "world"}
		result := Pipeline(input, "hel", true, true)

		if len(result) != 2 {
			t.Errorf("Pipeline() all = %v, want 2 items", result)
		}
		if result[0] != "hello" || result[1] != "help" {
			t.Errorf("Pipeline() all = %v, want [hello help]", result)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		result := Pipeline([]string{}, "pattern", true, true)

		if len(result) != 0 {
			t.Errorf("Pipeline() empty = %v, want empty", result)
		}
	})

	t.Run("grep no matches", func(t *testing.T) {
		input := []string{"apple", "banana", "cherry"}
		result := Pipeline(input, "xyz", false, false)

		if len(result) != 0 {
			t.Errorf("Pipeline() no match = %v, want empty", result)
		}
	})
}

func TestGrepLines(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		pattern  string
		expected int
	}{
		{"match all", []string{"abc", "abcd", "abcde"}, "abc", 3},
		{"match some", []string{"abc", "def", "abcdef"}, "abc", 2},
		{"match none", []string{"hello", "world"}, "xyz", 0},
		{"empty pattern", []string{"hello", "world"}, "", 2},
		{"case sensitive", []string{"Hello", "hello"}, "hello", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := grepLines(tt.lines, tt.pattern)
			if len(result) != tt.expected {
				t.Errorf("grepLines() = %d items, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestUniqLines(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected int
	}{
		{"all unique", []string{"a", "b", "c"}, 3},
		{"all same", []string{"a", "a", "a"}, 1},
		{"some duplicates", []string{"a", "b", "a", "c", "b"}, 3},
		{"empty", []string{}, 0},
		{"single", []string{"only"}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := uniqLines(tt.lines)
			if len(result) != tt.expected {
				t.Errorf("uniqLines() = %d items, want %d", len(result), tt.expected)
			}
		})
	}
}
