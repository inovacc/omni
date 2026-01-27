package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// GrepOptions configures the grep command behavior
type GrepOptions struct {
	IgnoreCase     bool // -i: ignore case distinctions
	InvertMatch    bool // -v: select non-matching lines
	LineNumber     bool // -n: prefix each line with line number
	Count          bool // -c: only print count of matching lines
	FilesWithMatch bool // -l: only print file names with matches
	FilesNoMatch   bool // -L: only print file names without matches
	OnlyMatching   bool // -o: show only the matched part
	Quiet          bool // -q: quiet mode, exit immediately with status 0 if match found
	WithFilename   bool // -H: print file name with output lines
	NoFilename     bool // -h: suppress file name prefix
	ExtendedRegexp bool // -E: interpret pattern as extended regexp
	FixedStrings   bool // -F: interpret pattern as fixed strings
	WordRegexp     bool // -w: match whole words only
	LineRegexp     bool // -x: match whole lines only
	Context        int  // -C: print NUM lines of context
	BeforeContext  int  // -B: print NUM lines of leading context
	AfterContext   int  // -A: print NUM lines of trailing context
	MaxCount       int  // -m: stop after NUM matches
	Recursive      bool // -r/-R: search recursively
}

// GrepResult represents the result of a grep operation
type GrepResult struct {
	Filename    string
	LineNumber  int
	Line        string
	MatchCount  int
	HasMatch    bool
	MatchedPart string
}

// RunGrep executes the grep command
func RunGrep(w io.Writer, pattern string, args []string, opts GrepOptions) error {
	if pattern == "" {
		return fmt.Errorf("grep: no pattern specified")
	}

	// Compile the pattern
	re, err := compilePattern(pattern, opts)
	if err != nil {
		return fmt.Errorf("grep: %w", err)
	}

	// Determine input sources
	files := args
	if len(files) == 0 {
		files = []string{"-"} // stdin
	}

	// Auto-enable filename display for multiple files
	showFilename := opts.WithFilename || (len(files) > 1 && !opts.NoFilename)

	totalMatches := 0
	anyMatch := false

	for _, file := range files {
		var r io.Reader

		filename := file

		if file == "-" {
			r = os.Stdin
			filename = "(standard input)"
		} else {
			f, err := os.Open(file)
			if err != nil {
				if !opts.Quiet {
					_, _ = fmt.Fprintf(os.Stderr, "grep: %s: %v\n", file, err)
				}

				continue
			}

			r = f

			defer func() {
				_ = f.Close()
			}()
		}

		matches, hasMatch, err := grepReader(w, r, filename, re, opts, showFilename)
		if err != nil {
			return err
		}

		totalMatches += matches

		if hasMatch {
			anyMatch = true
		}

		if opts.Quiet && anyMatch {
			return nil // Exit immediately on match in quiet mode
		}
	}

	if opts.Quiet {
		if anyMatch {
			return nil
		}

		return fmt.Errorf("no match")
	}

	return nil
}

func compilePattern(pattern string, opts GrepOptions) (*regexp.Regexp, error) {
	if opts.FixedStrings {
		pattern = regexp.QuoteMeta(pattern)
	}

	if opts.WordRegexp {
		pattern = `\b` + pattern + `\b`
	}

	if opts.LineRegexp {
		pattern = "^" + pattern + "$"
	}

	flags := ""
	if opts.IgnoreCase {
		flags = "(?i)"
	}

	return regexp.Compile(flags + pattern)
}

func grepReader(w io.Writer, r io.Reader, filename string, re *regexp.Regexp, opts GrepOptions, showFilename bool) (int, bool, error) {
	scanner := bufio.NewScanner(r)
	lineNum := 0
	matchCount := 0
	hasMatch := false

	// Context handling
	var beforeLines []string

	afterRemaining := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		matches := re.MatchString(line)

		if opts.InvertMatch {
			matches = !matches
		}

		if matches {
			hasMatch = true
			matchCount++

			if opts.MaxCount > 0 && matchCount > opts.MaxCount {
				break
			}

			if opts.Quiet {
				return matchCount, true, nil
			}

			if opts.Count || opts.FilesWithMatch || opts.FilesNoMatch {
				// Don't print lines, just track
				continue
			}

			// Print before-context lines
			if opts.BeforeContext > 0 || opts.Context > 0 {
				for _, ctxLine := range beforeLines {
					printGrepLine(w, filename, 0, ctxLine, opts, showFilename, true)
				}

				beforeLines = nil
			}

			// Print matching line
			if opts.OnlyMatching {
				for _, match := range re.FindAllString(line, -1) {
					printGrepLine(w, filename, lineNum, match, opts, showFilename, false)
				}
			} else {
				printGrepLine(w, filename, lineNum, line, opts, showFilename, false)
			}

			// Set up after-context
			if opts.AfterContext > 0 || opts.Context > 0 {
				afterRemaining = max(opts.Context, opts.AfterContext)
			}
		} else {
			// Handle after-context
			if afterRemaining > 0 {
				printGrepLine(w, filename, 0, line, opts, showFilename, true)

				afterRemaining--
			}

			// Track before-context
			if opts.BeforeContext > 0 || opts.Context > 0 {
				contextLines := max(opts.Context, opts.BeforeContext)

				beforeLines = append(beforeLines, line)
				if len(beforeLines) > contextLines {
					beforeLines = beforeLines[1:]
				}
			}
		}
	}

	// Handle file-level output options
	switch {
	case opts.FilesWithMatch && hasMatch:
		_, _ = fmt.Fprintln(w, filename)
	case opts.FilesNoMatch && !hasMatch:
		_, _ = fmt.Fprintln(w, filename)
	case opts.Count:
		if showFilename {
			_, _ = fmt.Fprintf(w, "%s:%d\n", filename, matchCount)
		} else {
			_, _ = fmt.Fprintf(w, "%d\n", matchCount)
		}
	}

	return matchCount, hasMatch, scanner.Err()
}

func printGrepLine(w io.Writer, filename string, lineNum int, line string, opts GrepOptions, showFilename bool, isContext bool) {
	var prefix string

	separator := ":"
	if isContext {
		separator = "-"
	}

	if showFilename {
		prefix = filename + separator
	}

	if opts.LineNumber && lineNum > 0 {
		prefix += fmt.Sprintf("%d%s", lineNum, separator)
	}

	_, _ = fmt.Fprintf(w, "%s%s\n", prefix, line)
}

// Grep filters lines containing the pattern (simple version for compatibility)
func Grep(lines []string, pattern string) []string {
	var out []string

	for _, l := range lines {
		if strings.Contains(l, pattern) {
			out = append(out, l)
		}
	}

	return out
}

// GrepWithOptions filters lines with options (compatibility wrapper)
func GrepWithOptions(lines []string, pattern string, opt GrepOptions) []string {
	out := []string{}

	re, err := compilePattern(pattern, opt)
	if err != nil {
		// Fall back to simple string matching
		if opt.IgnoreCase {
			pattern = strings.ToLower(pattern)
		}

		for _, l := range lines {
			line := l
			if opt.IgnoreCase {
				line = strings.ToLower(l)
			}

			match := strings.Contains(line, pattern)
			if opt.InvertMatch {
				match = !match
			}

			if match {
				out = append(out, l)
			}
		}

		return out
	}

	for _, l := range lines {
		match := re.MatchString(l)
		if opt.InvertMatch {
			match = !match
		}

		if match {
			out = append(out, l)
		}
	}

	return out
}
