package textutil

import (
	"testing"
)

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

	t.Run("empty slice", func(t *testing.T) {
		var lines []string

		Sort(lines)

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

func TestSortLines(t *testing.T) {
	t.Run("reverse", func(t *testing.T) {
		lines := []string{"apple", "banana", "cherry"}

		SortLines(lines, WithReverse())

		if lines[0] != "cherry" || lines[2] != "apple" {
			t.Errorf("SortLines() reverse = %v", lines)
		}
	})

	t.Run("numeric", func(t *testing.T) {
		lines := []string{"10", "2", "1", "20"}

		SortLines(lines, WithNumeric())

		if lines[0] != "1" || lines[1] != "2" || lines[2] != "10" || lines[3] != "20" {
			t.Errorf("SortLines() numeric = %v", lines)
		}
	})

	t.Run("numeric reverse", func(t *testing.T) {
		lines := []string{"1", "10", "2", "20"}

		SortLines(lines, WithNumeric(), WithReverse())

		if lines[0] != "20" || lines[3] != "1" {
			t.Errorf("SortLines() numeric reverse = %v", lines)
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		lines := []string{"Banana", "apple", "Cherry"}

		SortLines(lines, WithIgnoreCase())

		if lines[0] != "apple" {
			t.Errorf("SortLines() case insensitive first = %v", lines[0])
		}
	})

	t.Run("ignore leading", func(t *testing.T) {
		lines := []string{"  banana", "apple", "   cherry"}

		SortLines(lines, WithIgnoreLeading())

		if lines[0] != "apple" {
			t.Errorf("SortLines() ignore leading first = %v", lines[0])
		}
	})

	t.Run("stable", func(t *testing.T) {
		lines := []string{"b", "a", "b", "a"}

		SortLines(lines, WithStable())

		if lines[0] != "a" || lines[2] != "b" {
			t.Errorf("SortLines() stable = %v", lines)
		}
	})
}

func TestSortLinesWithOpts(t *testing.T) {
	t.Run("numeric", func(t *testing.T) {
		lines := []string{"10", "2", "1", "20"}

		SortLinesWithOpts(lines, SortOptions{Numeric: true})

		if lines[0] != "1" || lines[1] != "2" || lines[2] != "10" || lines[3] != "20" {
			t.Errorf("SortLinesWithOpts() numeric = %v", lines)
		}
	})
}

func TestCheckSorted(t *testing.T) {
	t.Run("sorted", func(t *testing.T) {
		lines := []string{"apple", "banana", "cherry"}

		result := CheckSorted(lines, SortOptions{})
		if result != "" {
			t.Errorf("CheckSorted() sorted should return empty, got %q", result)
		}
	})

	t.Run("unsorted", func(t *testing.T) {
		lines := []string{"cherry", "apple", "banana"}

		result := CheckSorted(lines, SortOptions{})
		if result == "" {
			t.Error("CheckSorted() unsorted should return disorder line")
		}
	})

	t.Run("numeric sorted", func(t *testing.T) {
		lines := []string{"1", "2", "10"}

		result := CheckSorted(lines, SortOptions{Numeric: true})
		if result != "" {
			t.Errorf("CheckSorted() numeric sorted should return empty, got %q", result)
		}
	})

	t.Run("reverse sorted", func(t *testing.T) {
		lines := []string{"cherry", "banana", "apple"}

		result := CheckSorted(lines, SortOptions{Reverse: true})
		if result != "" {
			t.Errorf("CheckSorted() reverse sorted should return empty, got %q", result)
		}
	})
}

func TestUniqueConsecutive(t *testing.T) {
	t.Run("consecutive duplicates", func(t *testing.T) {
		lines := []string{"a", "a", "b", "b", "c"}

		result := UniqueConsecutive(lines, false)
		if len(result) != 3 || result[0] != "a" || result[1] != "b" || result[2] != "c" {
			t.Errorf("UniqueConsecutive() = %v", result)
		}
	})

	t.Run("non-consecutive duplicates preserved", func(t *testing.T) {
		lines := []string{"a", "b", "a", "b"}

		result := UniqueConsecutive(lines, false)
		if len(result) != 4 {
			t.Errorf("UniqueConsecutive() non-consecutive = %v, want 4 elements", result)
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		lines := []string{"Apple", "apple", "APPLE", "banana"}

		result := UniqueConsecutive(lines, true)
		if len(result) != 2 {
			t.Errorf("UniqueConsecutive() case insensitive = %v, want 2 elements", result)
		}
	})

	t.Run("empty", func(t *testing.T) {
		var lines []string

		result := UniqueConsecutive(lines, false)
		if len(result) != 0 {
			t.Errorf("UniqueConsecutive() empty = %v", result)
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

	t.Run("non-consecutive duplicates", func(t *testing.T) {
		lines := []string{"a", "b", "a", "b"}

		result := Uniq(lines)

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
}
