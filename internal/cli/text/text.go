package text

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/input"
	"github.com/inovacc/omni/pkg/textutil"
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
	JSON          bool   // --json: output as JSON
}

// SortResult represents sort output for JSON
type SortResult struct {
	Lines []string `json:"lines"`
	Count int      `json:"count"`
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
	JSON          bool // --json: output as JSON
}

// UniqResult represents uniq output for JSON
type UniqResult struct {
	Lines []UniqLine `json:"lines"`
	Count int        `json:"count"`
}

// UniqLine represents a line with its count for JSON
type UniqLine struct {
	Line  string `json:"line"`
	Count int    `json:"count"`
}

// RunSort executes the sort command
// r is the default input reader (used when args is empty or contains "-")
func RunSort(w io.Writer, r io.Reader, args []string, opts SortOptions) error {
	sources, err := input.Open(args, r)
	if err != nil {
		return fmt.Errorf("sort: %w", err)
	}
	defer input.CloseAll(sources)

	var lines []string

	for _, src := range sources {
		scanner := bufio.NewScanner(src.Reader)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("sort: %w", err)
		}
	}

	// Check mode
	if opts.Check {
		pkgOpts := toPkgSortOptions(opts)
		disorder := textutil.CheckSorted(lines, pkgOpts)
		if disorder != "" {
			return fmt.Errorf("sort: disorder: %s", disorder)
		}
		return nil
	}

	// Sort the lines
	textutil.SortLinesWithOpts(lines, toPkgSortOptions(opts))

	// Remove duplicates if -u
	if opts.Unique {
		lines = textutil.UniqueConsecutive(lines, opts.IgnoreCase)
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(SortResult{Lines: lines, Count: len(lines)})
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

func toPkgSortOptions(opts SortOptions) textutil.SortOptions {
	return textutil.SortOptions{
		Reverse:       opts.Reverse,
		Numeric:       opts.Numeric,
		Unique:        opts.Unique,
		IgnoreCase:    opts.IgnoreCase,
		IgnoreLeading: opts.IgnoreLeading,
		Stable:        opts.Stable,
	}
}

// RunUniq executes the uniq command
// r is the default input reader (used when args is empty or contains "-")
func RunUniq(w io.Writer, r io.Reader, args []string, opts UniqOptions) error {
	src, err := input.OpenOne(args, r)
	if err != nil {
		return fmt.Errorf("uniq: %w", err)
	}
	defer input.MustClose(&src)

	scanner := bufio.NewScanner(src.Reader)

	var (
		prevLine string
		prevKey  string
	)

	count := 0
	first := true

	var jsonLines []UniqLine

	outputLine := func(line string, cnt int) {
		if opts.Unique && cnt > 1 {
			return
		}

		if opts.Repeated && cnt <= 1 {
			return
		}

		if opts.JSON {
			jsonLines = append(jsonLines, UniqLine{Line: line, Count: cnt})
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

			if opts.AllRepeated && !opts.JSON {
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

	if opts.JSON {
		return json.NewEncoder(w).Encode(UniqResult{Lines: jsonLines, Count: len(jsonLines)})
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
	textutil.Sort(lines)
}

// Uniq returns unique lines from a slice (for compatibility)
func Uniq(lines []string) []string {
	return textutil.Uniq(lines)
}

// TrimLines trims whitespace from all lines (for compatibility)
func TrimLines(lines []string) []string {
	return textutil.TrimLines(lines)
}
