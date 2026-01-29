package text

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

// SortOptions configures the sort command behavior
type SortOptions struct {
	Reverse       bool   // -r: reverse the result of comparisons
	Numeric       bool   // -n: compare according to string numerical value
	Unique        bool   // -u: output only unique lines
	IgnoreCase    bool   // -f: fold lower case to upper case characters
	IgnoreLeading bool   // -b: ignore leading blanks
	Dictionary    bool   // -d: consider only blanks and alphanumeric characters
	Key           string // -k: sort via a key
	FieldSep      string // -t: use SEP as field separator
	Check         bool   // -c: check for sorted input
	Stable        bool   // -s: stabilize sort by disabling last-resort comparison
	Output        string // -o: write result to FILE
}

// UniqOptions configures the uniq command behavior
type UniqOptions struct {
	Count         bool // -c: prefix lines by the number of occurrences
	Repeated      bool // -d: only print duplicate lines
	AllRepeated   bool // -D: print all duplicate lines
	IgnoreCase    bool // -i: ignore differences in case
	Unique        bool // -u: only print unique lines
	SkipFields    int  // -f: avoid comparing the first N fields
	SkipChars     int  // -s: avoid comparing the first N characters
	CheckChars    int  // -w: compare no more than N characters
	ZeroTerminate bool // -z: line delimiter is NUL, not newline
}

// RunSort executes the sort command
func RunSort(w io.Writer, args []string, opts SortOptions) error {
	var lines []string

	if len(args) == 0 {
		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("sort: %w", err)
		}
	} else {
		// Read from files
		for _, file := range args {
			f, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("sort: cannot read: %s: %w", file, err)
			}

			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}

			_ = f.Close()

			if err := scanner.Err(); err != nil {
				return fmt.Errorf("sort: %w", err)
			}
		}
	}

	// Check mode
	if opts.Check {
		return checkSorted(lines, opts)
	}

	// Sort the lines
	sortLines(lines, opts)

	// Remove duplicates if -u
	if opts.Unique {
		lines = uniqueLines(lines, opts.IgnoreCase)
	}

	// Write output
	var output = w

	if opts.Output != "" {
		f, err := os.Create(opts.Output)
		if err != nil {
			return fmt.Errorf("sort: %w", err)
		}

		defer func() {
			_ = f.Close()
		}()

		output = f
	}

	for _, line := range lines {
		_, _ = fmt.Fprintln(output, line)
	}

	return nil
}

func sortLines(lines []string, opts SortOptions) {
	comparator := func(i, j int) bool {
		a, b := lines[i], lines[j]

		// Ignore leading blanks
		if opts.IgnoreLeading {
			a = strings.TrimLeft(a, " \t")
			b = strings.TrimLeft(b, " \t")
		}

		// Case insensitive
		if opts.IgnoreCase {
			a = strings.ToLower(a)
			b = strings.ToLower(b)
		}

		// Numeric comparison
		if opts.Numeric {
			na, _ := strconv.ParseFloat(strings.TrimSpace(a), 64)

			nb, _ := strconv.ParseFloat(strings.TrimSpace(b), 64)
			if opts.Reverse {
				return na > nb
			}

			return na < nb
		}

		// String comparison
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

func checkSorted(lines []string, opts SortOptions) error {
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
			return fmt.Errorf("sort: disorder: %s", lines[i])
		}
	}

	return nil
}

func uniqueLines(lines []string, ignoreCase bool) []string {
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

// RunUniq executes the uniq command
func RunUniq(w io.Writer, args []string, opts UniqOptions) error {
	var r io.Reader = os.Stdin

	if len(args) > 0 && args[0] != "-" {
		f, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("uniq: %w", err)
		}

		defer func() {
			_ = f.Close()
		}()

		r = f
	}

	scanner := bufio.NewScanner(r)

	var (
		prevLine string
		prevKey  string
	)

	count := 0
	first := true

	outputLine := func(line string, cnt int) {
		if opts.Unique && cnt > 1 {
			return
		}

		if opts.Repeated && cnt <= 1 {
			return
		}

		if opts.Count {
			_, _ = fmt.Fprintf(w, "%7d %s\n", cnt, line)
		} else {
			_, _ = fmt.Fprintln(w, line)
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		key := getUniqKey(line, opts)

		if first {
			prevLine = line
			prevKey = key
			count = 1
			first = false

			continue
		}

		if keysEqual(prevKey, key, opts) {
			count++

			if opts.AllRepeated {
				_, _ = fmt.Fprintln(w, line)
			}
		} else {
			outputLine(prevLine, count)
			prevLine = line
			prevKey = key
			count = 1
		}
	}

	// Output last line
	if !first {
		outputLine(prevLine, count)
	}

	return scanner.Err()
}

func getUniqKey(line string, opts UniqOptions) string {
	key := line

	// Skip fields
	if opts.SkipFields > 0 {
		fields := strings.Fields(line)
		if opts.SkipFields < len(fields) {
			key = strings.Join(fields[opts.SkipFields:], " ")
		} else {
			key = ""
		}
	}

	// Skip characters
	if opts.SkipChars > 0 && len(key) > opts.SkipChars {
		key = key[opts.SkipChars:]
	}

	// Check only first N characters
	if opts.CheckChars > 0 && len(key) > opts.CheckChars {
		key = key[:opts.CheckChars]
	}

	return key
}

func keysEqual(a, b string, opts UniqOptions) bool {
	if opts.IgnoreCase {
		return strings.EqualFold(a, b)
	}

	return a == b
}

// Sort sorts a slice of strings in place (for compatibility)
func Sort(lines []string) {
	sort.Strings(lines)
}

// Uniq returns unique lines from a slice (for compatibility)
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

// TrimLines trims whitespace from all lines (for compatibility)
func TrimLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		out = append(out, strings.TrimSpace(l))
	}

	return out
}
