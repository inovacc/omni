package rg

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

// OutputFormat represents the output format type
type OutputFormat int

const (
	FormatDefault  OutputFormat = iota // Default ripgrep-style output
	FormatNoHeading                    // No file name headings (all on one line)
	FormatJSON                         // Full JSON output
	FormatJSONStream                   // Streaming NDJSON output
)

// Formatter handles output formatting with color support
type Formatter struct {
	w              io.Writer
	scheme         ColorScheme
	useColor       bool
	format         OutputFormat
	showLineNumber bool
	showColumn     bool
	onlyMatching   bool
	trim           bool
	replace        string
	re             *regexp.Regexp
	pattern        string
	caseInsensitive bool
	useLiteral     bool
}

// FormatterOptions configures the formatter
type FormatterOptions struct {
	UseColor        bool
	Scheme          ColorScheme
	Format          OutputFormat
	ShowLineNumber  bool
	ShowColumn      bool
	OnlyMatching    bool
	Trim            bool
	Replace         string
	Regex           *regexp.Regexp
	Pattern         string
	CaseInsensitive bool
	UseLiteral      bool
}

// NewFormatter creates a new output formatter
func NewFormatter(w io.Writer, opts FormatterOptions) *Formatter {
	return &Formatter{
		w:              w,
		scheme:         opts.Scheme,
		useColor:       opts.UseColor,
		format:         opts.Format,
		showLineNumber: opts.ShowLineNumber,
		showColumn:     opts.ShowColumn,
		onlyMatching:   opts.OnlyMatching,
		trim:           opts.Trim,
		replace:        opts.Replace,
		re:             opts.Regex,
		pattern:        opts.Pattern,
		caseInsensitive: opts.CaseInsensitive,
		useLiteral:     opts.UseLiteral,
	}
}

// PrintFileHeader prints a file header (only used in heading mode)
func (f *Formatter) PrintFileHeader(path string) {
	if f.format == FormatNoHeading {
		return
	}
	_, _ = fmt.Fprintln(f.w, FormatPath(path, f.scheme, f.useColor))
}

// PrintMatch prints a single match line
func (f *Formatter) PrintMatch(path string, lineNum, column int, line string, isContext bool) {
	// Handle trim
	if f.trim {
		line = strings.TrimSpace(line)
	}

	// Handle replacement
	if f.replace != "" && !isContext {
		if f.re != nil {
			line = f.re.ReplaceAllString(line, f.replace)
		}
	}

	// Handle only-matching
	if f.onlyMatching && !isContext {
		matches := f.findMatches(line)
		for _, m := range matches {
			f.printOnlyMatch(path, lineNum, m)
		}
		return
	}

	// Build output line
	var sb strings.Builder

	// Determine separator (: for matches, - for context)
	sep := ":"
	if isContext {
		sep = "-"
	}

	// Highlight matches in the line
	highlightedLine := line
	if !isContext && f.useColor {
		if f.useLiteral {
			highlightedLine = HighlightLiteralMatches(line, f.pattern, f.caseInsensitive, f.scheme, f.useColor)
		} else if f.re != nil {
			highlightedLine = HighlightMatches(line, f.re, f.scheme, f.useColor)
		}
	}

	switch f.format {
	case FormatNoHeading:
		// path:linenum:column:line or path:linenum:line or path:line
		sb.WriteString(FormatPath(path, f.scheme, f.useColor))
		sb.WriteString(FormatSeparator(sep, f.scheme, f.useColor))

		if f.showLineNumber && lineNum > 0 {
			sb.WriteString(FormatLineNumber(lineNum, f.scheme, f.useColor))
			sb.WriteString(FormatSeparator(sep, f.scheme, f.useColor))
		}

		if f.showColumn && column > 0 {
			sb.WriteString(FormatColumn(column, f.scheme, f.useColor))
			sb.WriteString(FormatSeparator(sep, f.scheme, f.useColor))
		}

		sb.WriteString(highlightedLine)

	default: // FormatDefault - grouped by file
		if f.showLineNumber && lineNum > 0 {
			sb.WriteString(FormatLineNumber(lineNum, f.scheme, f.useColor))
			sb.WriteString(FormatSeparator(sep, f.scheme, f.useColor))
		}

		if f.showColumn && column > 0 {
			sb.WriteString(FormatColumn(column, f.scheme, f.useColor))
			sb.WriteString(FormatSeparator(sep, f.scheme, f.useColor))
		}

		sb.WriteString(highlightedLine)
	}

	_, _ = fmt.Fprintln(f.w, sb.String())
}

// PrintContextSeparator prints the context separator ("--")
func (f *Formatter) PrintContextSeparator() {
	sep := "--"
	if f.useColor && f.scheme.Separator != "" {
		sep = f.scheme.Separator + "--" + Reset
	}
	_, _ = fmt.Fprintln(f.w, sep)
}

// PrintFilesWithMatch prints just the filename (for -l mode)
func (f *Formatter) PrintFilesWithMatch(path string) {
	_, _ = fmt.Fprintln(f.w, FormatPath(path, f.scheme, f.useColor))
}

// PrintCount prints the count for a file (for -c mode)
func (f *Formatter) PrintCount(path string, count int) {
	_, _ = fmt.Fprintf(f.w, "%s%s%d\n",
		FormatPath(path, f.scheme, f.useColor),
		FormatSeparator(":", f.scheme, f.useColor),
		count)
}

// findMatches finds all match substrings in the line
func (f *Formatter) findMatches(line string) []string {
	if f.useLiteral {
		return f.findLiteralMatches(line)
	}
	if f.re == nil {
		return nil
	}
	return f.re.FindAllString(line, -1)
}

// findLiteralMatches finds all literal matches in the line
func (f *Formatter) findLiteralMatches(line string) []string {
	var matches []string
	searchLine := line
	searchPattern := f.pattern
	if f.caseInsensitive {
		searchLine = strings.ToLower(line)
		searchPattern = strings.ToLower(f.pattern)
	}

	offset := 0
	for {
		idx := strings.Index(searchLine[offset:], searchPattern)
		if idx == -1 {
			break
		}
		start := offset + idx
		end := start + len(f.pattern)
		matches = append(matches, line[start:end])
		offset = end
	}

	return matches
}

// printOnlyMatch prints a single match in only-matching mode
func (f *Formatter) printOnlyMatch(path string, lineNum int, match string) {
	var sb strings.Builder

	if f.format == FormatNoHeading {
		sb.WriteString(FormatPath(path, f.scheme, f.useColor))
		sb.WriteString(FormatSeparator(":", f.scheme, f.useColor))
	}

	if f.showLineNumber && lineNum > 0 {
		sb.WriteString(FormatLineNumber(lineNum, f.scheme, f.useColor))
		sb.WriteString(FormatSeparator(":", f.scheme, f.useColor))
	}

	// Colorize the match itself
	if f.useColor && f.scheme.Match != "" {
		sb.WriteString(f.scheme.Match)
		sb.WriteString(match)
		sb.WriteString(Reset)
	} else {
		sb.WriteString(match)
	}

	_, _ = fmt.Fprintln(f.w, sb.String())
}

// ByteOffset calculates the byte offset of a match in a file
type ByteOffset struct {
	fileOffset int64 // Cumulative bytes from previous lines
}

// NewByteOffset creates a new byte offset tracker
func NewByteOffset() *ByteOffset {
	return &ByteOffset{}
}

// AddLine adds a line's length to the running offset
func (b *ByteOffset) AddLine(line string) {
	b.fileOffset += int64(len(line)) + 1 // +1 for newline
}

// GetMatchOffset returns the absolute byte offset of a match
func (b *ByteOffset) GetMatchOffset(column int) int64 {
	return b.fileOffset + int64(column-1) // column is 1-indexed
}

// Stats tracks search statistics
type Stats struct {
	FilesSearched int
	FilesMatched  int
	TotalMatches  int
	BytesSearched int64
}

// PrintStats prints search statistics
func (s *Stats) PrintStats(w io.Writer) {
	_, _ = fmt.Fprintf(w, "\n%d matches\n", s.TotalMatches)
	_, _ = fmt.Fprintf(w, "%d files contained matches\n", s.FilesMatched)
	_, _ = fmt.Fprintf(w, "%d files searched\n", s.FilesSearched)
}
