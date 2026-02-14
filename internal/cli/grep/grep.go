package grep

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/inovacc/omni/internal/cli/input"
	"github.com/inovacc/omni/internal/cli/output"
	pkggrep "github.com/inovacc/omni/pkg/search/grep"
)

// GrepOptions configures the grep command behavior
type GrepOptions struct {
	IgnoreCase     bool          // -i: ignore case distinctions
	InvertMatch    bool          // -v: select non-matching lines
	LineNumber     bool          // -n: prefix each line with line number
	Count          bool          // -c: only print count of matching lines
	FilesWithMatch bool          // -l: only print file names with matches
	FilesNoMatch   bool          // -L: only print file names without matches
	OnlyMatching   bool          // -o: show only the matched part
	Quiet          bool          // -q: quiet mode, exit immediately with status 0 if match found
	WithFilename   bool          // -H: print file name with output lines
	NoFilename     bool          // -h: suppress file name prefix
	ExtendedRegexp bool          // -E: interpret pattern as extended regexp
	FixedStrings   bool          // -F: interpret pattern as fixed strings
	WordRegexp     bool          // -w: match whole words only
	LineRegexp     bool          // -x: match whole lines only
	Context        int           // -C: print NUM lines of context
	BeforeContext  int           // -B: print NUM lines of leading context
	AfterContext   int           // -A: print NUM lines of trailing context
	MaxCount       int           // -m: stop after NUM matches
	Recursive      bool          // -r/-R: search recursively
	OutputFormat   output.Format // output format
}

// GrepResult represents the result of a grep operation
type GrepResult struct {
	Filename    string `json:"filename,omitempty"`
	LineNumber  int    `json:"line_number,omitempty"`
	Line        string `json:"line,omitempty"`
	MatchCount  int    `json:"match_count,omitempty"`
	HasMatch    bool   `json:"has_match"`
	MatchedPart string `json:"matched_part,omitempty"`
}

// GrepOutput represents the complete grep output for JSON
type GrepOutput struct {
	Pattern        string       `json:"pattern"`
	Files          []string     `json:"files"`
	Matches        []GrepResult `json:"matches"`
	TotalCount     int          `json:"total_count"`
	FilesWithMatch int          `json:"files_with_match"`
}

// RunGrep executes the grep command
// r is the default input reader (used when args is empty or contains "-")
func RunGrep(w io.Writer, r io.Reader, pattern string, args []string, opts GrepOptions) error {
	if pattern == "" {
		return fmt.Errorf("grep: no pattern specified")
	}

	// Compile the pattern
	re, err := compilePattern(pattern, opts)
	if err != nil {
		return fmt.Errorf("grep: %w", err)
	}

	// Determine input sources
	sources, err := input.Open(args, r)
	if err != nil {
		if !opts.Quiet {
			_, _ = fmt.Fprintf(os.Stderr, "grep: %v\n", err)
		}

		return err
	}
	defer input.CloseAll(sources)

	// Auto-enable filename display for multiple sources
	showFilename := opts.WithFilename || (len(sources) > 1 && !opts.NoFilename)

	f := output.New(w, opts.OutputFormat)
	jsonMode := f.IsJSON()

	totalMatches := 0
	anyMatch := false
	filesWithMatch := 0

	var allResults []GrepResult

	for _, src := range sources {
		filename := src.Name
		if filename == "standard input" {
			filename = "(standard input)"
		}

		matches, hasMatch, results, err := grepReader(w, src.Reader, filename, re, opts, showFilename, jsonMode)
		if err != nil {
			return err
		}

		totalMatches += matches

		if hasMatch {
			anyMatch = true
			filesWithMatch++
		}

		if jsonMode {
			allResults = append(allResults, results...)
		}

		if opts.Quiet && anyMatch {
			return nil // Exit immediately on match in quiet mode
		}
	}

	if jsonMode {
		// Build file list for JSON output
		fileList := make([]string, len(sources))
		for i, src := range sources {
			fileList[i] = src.Name
		}

		grepOut := GrepOutput{
			Pattern:        pattern,
			Files:          fileList,
			Matches:        allResults,
			TotalCount:     totalMatches,
			FilesWithMatch: filesWithMatch,
		}

		return f.Print(grepOut)
	}

	if opts.Quiet {
		if anyMatch {
			return nil
		}

		return fmt.Errorf("no match")
	}

	if !anyMatch {
		return fmt.Errorf("no match")
	}

	return nil
}

func compilePattern(pattern string, opts GrepOptions) (*regexp.Regexp, error) {
	pkgOpts := pkggrep.Options{
		IgnoreCase:   opts.IgnoreCase,
		InvertMatch:  false, // not used for pattern compilation
		FixedStrings: opts.FixedStrings,
		WordRegexp:   opts.WordRegexp,
		LineRegexp:   opts.LineRegexp,
	}

	return pkggrep.CompilePattern(pattern, pkgOpts)
}

func grepReader(w io.Writer, r io.Reader, filename string, re *regexp.Regexp, opts GrepOptions, showFilename bool, jsonMode bool) (int, bool, []GrepResult, error) {
	scanner := bufio.NewScanner(r)
	lineNum := 0
	matchCount := 0
	hasMatch := false

	var results []GrepResult

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
				return matchCount, true, results, nil
			}

			if jsonMode {
				matchedPart := ""
				if matched := re.FindString(line); matched != "" {
					matchedPart = matched
				}

				results = append(results, GrepResult{
					Filename:    filename,
					LineNumber:  lineNum,
					Line:        line,
					HasMatch:    true,
					MatchedPart: matchedPart,
				})

				continue
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

	return matchCount, hasMatch, results, scanner.Err()
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
	return pkggrep.Search(lines, pattern)
}

// GrepWithOptions filters lines with options (compatibility wrapper)
func GrepWithOptions(lines []string, pattern string, opt GrepOptions) []string {
	return pkggrep.SearchWithOptionsStruct(lines, pattern, pkggrep.Options{
		IgnoreCase:   opt.IgnoreCase,
		InvertMatch:  opt.InvertMatch,
		FixedStrings: opt.FixedStrings,
		WordRegexp:   opt.WordRegexp,
		LineRegexp:   opt.LineRegexp,
	})
}
