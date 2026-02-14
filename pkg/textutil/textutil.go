// Package textutil provides text processing utilities including sorting, deduplication, and trimming.
package textutil

import (
	"sort"
	"strconv"
	"strings"
)

// SortOptions configures sorting behavior.
type SortOptions struct {
	Reverse       bool // Reverse the result of comparisons
	Numeric       bool // Compare according to string numerical value
	Unique        bool // Output only unique lines
	IgnoreCase    bool // Fold lower case to upper case characters
	IgnoreLeading bool // Ignore leading blanks
	Stable        bool // Stabilize sort by disabling last-resort comparison
}

// SortOption is a functional option for SortLines.
type SortOption func(*SortOptions)

// WithReverse reverses the sort order.
func WithReverse() SortOption {
	return func(o *SortOptions) { o.Reverse = true }
}

// WithNumeric enables numerical comparison.
func WithNumeric() SortOption {
	return func(o *SortOptions) { o.Numeric = true }
}

// WithUnique outputs only unique lines.
func WithUnique() SortOption {
	return func(o *SortOptions) { o.Unique = true }
}

// WithIgnoreCase enables case-insensitive sorting.
func WithIgnoreCase() SortOption {
	return func(o *SortOptions) { o.IgnoreCase = true }
}

// WithIgnoreLeading ignores leading blanks.
func WithIgnoreLeading() SortOption {
	return func(o *SortOptions) { o.IgnoreLeading = true }
}

// WithStable enables stable sorting.
func WithStable() SortOption {
	return func(o *SortOptions) { o.Stable = true }
}

// SortLines sorts a slice of strings in place with options.
func SortLines(lines []string, opts ...SortOption) {
	o := SortOptions{}
	for _, opt := range opts {
		opt(&o)
	}

	sortLines(lines, o)
}

// SortLinesWithOpts sorts a slice of strings in place using the options struct.
func SortLinesWithOpts(lines []string, opts SortOptions) {
	sortLines(lines, opts)
}

// Sort sorts a slice of strings in place alphabetically.
func Sort(lines []string) {
	sort.Strings(lines)
}

// CheckSorted verifies that lines are sorted according to options.
// Returns an error string if not sorted, or empty string if sorted.
func CheckSorted(lines []string, opts SortOptions) string {
	for i := 1; i < len(lines); i++ {
		a, b := lines[i-1], lines[i]
		if opts.IgnoreCase {
			a = strings.ToLower(a)
			b = strings.ToLower(b)
		}

		var outOfOrder bool

		if opts.Numeric {
			na, _ := strconv.ParseFloat(strings.TrimSpace(a), 64)

			nb, _ := strconv.ParseFloat(strings.TrimSpace(b), 64)
			if opts.Reverse {
				outOfOrder = na < nb
			} else {
				outOfOrder = na > nb
			}
		} else {
			if opts.Reverse {
				outOfOrder = a < b
			} else {
				outOfOrder = a > b
			}
		}

		if outOfOrder {
			return lines[i]
		}
	}

	return ""
}

// UniqueConsecutive removes consecutive duplicate lines (like Unix uniq).
func UniqueConsecutive(lines []string, ignoreCase bool) []string {
	if len(lines) == 0 {
		return lines
	}

	result := []string{lines[0]}
	for i := 1; i < len(lines); i++ {
		prev := result[len(result)-1]

		curr := lines[i]
		if ignoreCase {
			if !strings.EqualFold(prev, curr) {
				result = append(result, curr)
			}
		} else {
			if prev != curr {
				result = append(result, curr)
			}
		}
	}

	return result
}

// Uniq returns unique lines from a slice (removes all duplicates, not just consecutive).
func Uniq(lines []string) []string {
	seen := make(map[string]bool)
	out := make([]string, 0, len(lines))

	for _, l := range lines {
		if !seen[l] {
			seen[l] = true
			out = append(out, l)
		}
	}

	return out
}

// TrimLines trims whitespace from all lines.
func TrimLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		out = append(out, strings.TrimSpace(l))
	}

	return out
}

func sortLines(lines []string, opts SortOptions) {
	comparator := func(i, j int) bool {
		a, b := lines[i], lines[j]

		if opts.IgnoreLeading {
			a = strings.TrimLeft(a, " \t")
			b = strings.TrimLeft(b, " \t")
		}

		if opts.IgnoreCase {
			a = strings.ToLower(a)
			b = strings.ToLower(b)
		}

		if opts.Numeric {
			na, _ := strconv.ParseFloat(strings.TrimSpace(a), 64)

			nb, _ := strconv.ParseFloat(strings.TrimSpace(b), 64)
			if opts.Reverse {
				return na > nb
			}

			return na < nb
		}

		if opts.Reverse {
			return a > b
		}

		return a < b
	}

	if opts.Stable {
		sort.SliceStable(lines, comparator)
	} else {
		sort.Slice(lines, comparator)
	}
}
