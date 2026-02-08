package pipeline

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// --- Streaming stages (line-by-line, constant memory) ---

// Grep filters lines matching a pattern.
type Grep struct {
	Pattern    string
	IgnoreCase bool
	Invert     bool
	re         *regexp.Regexp
}

func (s *Grep) Name() string {
	if s.Invert {
		return "grep-v"
	}
	return "grep"
}

func (s *Grep) Process(ctx context.Context, in io.Reader, out io.Writer) error {
	pattern := s.Pattern
	if s.IgnoreCase {
		pattern = "(?i)" + pattern
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("grep: invalid pattern %q: %w", s.Pattern, err)
	}

	s.re = re

	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		line := scanner.Text()
		matched := s.re.MatchString(line)

		if matched != s.Invert {
			if _, err := fmt.Fprintln(out, line); err != nil {
				return nil // downstream closed
			}
		}
	}

	return scanner.Err()
}

// Contains filters lines containing a literal substring.
type Contains struct {
	Substr     string
	IgnoreCase bool
}

func (s *Contains) Name() string { return "contains" }

func (s *Contains) Process(ctx context.Context, in io.Reader, out io.Writer) error {
	substr := s.Substr
	if s.IgnoreCase {
		substr = strings.ToLower(substr)
	}

	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		line := scanner.Text()
		check := line
		if s.IgnoreCase {
			check = strings.ToLower(line)
		}

		if strings.Contains(check, substr) {
			if _, err := fmt.Fprintln(out, line); err != nil {
				return nil
			}
		}
	}

	return scanner.Err()
}

// Replace replaces all occurrences of old with new in each line.
type Replace struct {
	Old string
	New string
}

func (s *Replace) Name() string { return "replace" }

func (s *Replace) Process(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		line := strings.ReplaceAll(scanner.Text(), s.Old, s.New)
		if _, err := fmt.Fprintln(out, line); err != nil {
			return nil
		}
	}

	return scanner.Err()
}

// Head outputs the first N lines.
type Head struct {
	N int
}

func (s *Head) Name() string { return "head" }

func (s *Head) Process(ctx context.Context, in io.Reader, out io.Writer) error {
	n := s.N
	if n <= 0 {
		n = 10
	}

	scanner := bufio.NewScanner(in)
	count := 0

	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if count >= n {
			break
		}

		if _, err := fmt.Fprintln(out, scanner.Text()); err != nil {
			return nil
		}

		count++
	}

	// Drain remaining input to avoid broken pipe
	for scanner.Scan() {
	}

	return scanner.Err()
}

// Skip skips the first N lines and outputs the rest.
type Skip struct {
	N int
}

func (s *Skip) Name() string { return "skip" }

func (s *Skip) Process(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	skipped := 0

	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if skipped < s.N {
			skipped++
			continue
		}

		if _, err := fmt.Fprintln(out, scanner.Text()); err != nil {
			return nil
		}
	}

	return scanner.Err()
}

// Uniq removes consecutive duplicate lines.
type Uniq struct {
	IgnoreCase bool
}

func (s *Uniq) Name() string { return "uniq" }

func (s *Uniq) Process(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	prev := ""
	first := true

	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		line := scanner.Text()
		cmp := line
		cmpPrev := prev

		if s.IgnoreCase {
			cmp = strings.ToLower(cmp)
			cmpPrev = strings.ToLower(cmpPrev)
		}

		if first || cmp != cmpPrev {
			if _, err := fmt.Fprintln(out, line); err != nil {
				return nil
			}
		}

		prev = line
		first = false
	}

	return scanner.Err()
}

// Cut extracts fields from each line.
type Cut struct {
	Delimiter string
	Fields    []int // 1-based field indices
}

func (s *Cut) Name() string { return "cut" }

func (s *Cut) Process(ctx context.Context, in io.Reader, out io.Writer) error {
	delim := s.Delimiter
	if delim == "" {
		delim = "\t"
	}

	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		line := scanner.Text()
		parts := strings.Split(line, delim)
		var selected []string

		for _, f := range s.Fields {
			if f >= 1 && f <= len(parts) {
				selected = append(selected, parts[f-1])
			}
		}

		if _, err := fmt.Fprintln(out, strings.Join(selected, delim)); err != nil {
			return nil
		}
	}

	return scanner.Err()
}

// Tr translates or deletes characters.
type Tr struct {
	From string
	To   string
}

func (s *Tr) Name() string { return "tr" }

func (s *Tr) Process(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		line := scanner.Text()
		result := translateChars(line, s.From, s.To)

		if _, err := fmt.Fprintln(out, result); err != nil {
			return nil
		}
	}

	return scanner.Err()
}

func translateChars(s, from, to string) string {
	fromRunes := []rune(from)
	toRunes := []rune(to)

	// Build mapping
	mapping := make(map[rune]rune)
	for i, r := range fromRunes {
		if i < len(toRunes) {
			mapping[r] = toRunes[i]
		} else if len(toRunes) > 0 {
			// Map to last char of 'to' if shorter
			mapping[r] = toRunes[len(toRunes)-1]
		}
	}

	var result strings.Builder
	for _, r := range s {
		if repl, ok := mapping[r]; ok {
			result.WriteRune(repl)
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// Sed performs regex substitution on each line.
type Sed struct {
	Pattern     string
	Replacement string
	Global      bool
}

func (s *Sed) Name() string { return "sed" }

func (s *Sed) Process(ctx context.Context, in io.Reader, out io.Writer) error {
	re, err := regexp.Compile(s.Pattern)
	if err != nil {
		return fmt.Errorf("sed: invalid pattern %q: %w", s.Pattern, err)
	}

	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		line := scanner.Text()
		if s.Global {
			line = re.ReplaceAllString(line, s.Replacement)
		} else {
			loc := re.FindStringIndex(line)
			if loc != nil {
				line = line[:loc[0]] + re.ReplaceAllString(line[loc[0]:loc[1]], s.Replacement) + line[loc[1]:]
			}
		}

		if _, err := fmt.Fprintln(out, line); err != nil {
			return nil
		}
	}

	return scanner.Err()
}

// Rev reverses each line character by character.
type Rev struct{}

func (s *Rev) Name() string { return "rev" }

func (s *Rev) Process(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		runes := []rune(scanner.Text())
		slices.Reverse(runes)

		if _, err := fmt.Fprintln(out, string(runes)); err != nil {
			return nil
		}
	}

	return scanner.Err()
}

// Nl numbers each line.
type Nl struct {
	Start int
}

func (s *Nl) Name() string { return "nl" }

func (s *Nl) Process(ctx context.Context, in io.Reader, out io.Writer) error {
	start := s.Start
	if start <= 0 {
		start = 1
	}

	scanner := bufio.NewScanner(in)
	n := start

	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if _, err := fmt.Fprintf(out, "%6d\t%s\n", n, scanner.Text()); err != nil {
			return nil
		}

		n++
	}

	return scanner.Err()
}

// Tee writes output to both a file and the next stage.
type Tee struct {
	Path string
}

func (s *Tee) Name() string { return "tee" }

func (s *Tee) Process(ctx context.Context, in io.Reader, out io.Writer) error {
	var writers []io.Writer
	writers = append(writers, out)

	if s.Path != "" {
		f, err := createFile(s.Path)
		if err != nil {
			return fmt.Errorf("tee: %w", err)
		}
		defer func() { _ = f.Close() }()

		writers = append(writers, f)
	}

	mw := io.MultiWriter(writers...)
	scanner := bufio.NewScanner(in)

	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if _, err := fmt.Fprintln(mw, scanner.Text()); err != nil {
			return nil
		}
	}

	return scanner.Err()
}

// Filter is a library-only stage that applies a Go function predicate.
type Filter struct {
	Fn   func(string) bool
	Desc string
}

func (s *Filter) Name() string {
	if s.Desc != "" {
		return "filter(" + s.Desc + ")"
	}
	return "filter"
}

func (s *Filter) Process(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		line := scanner.Text()
		if s.Fn(line) {
			if _, err := fmt.Fprintln(out, line); err != nil {
				return nil
			}
		}
	}

	return scanner.Err()
}

// Map is a library-only stage that applies a Go function to each line.
type Map struct {
	Fn   func(string) string
	Desc string
}

func (s *Map) Name() string {
	if s.Desc != "" {
		return "map(" + s.Desc + ")"
	}
	return "map"
}

func (s *Map) Process(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		line := s.Fn(scanner.Text())
		if _, err := fmt.Fprintln(out, line); err != nil {
			return nil
		}
	}

	return scanner.Err()
}

// --- Buffering stages (reads all input first) ---

// Sort sorts all lines.
type Sort struct {
	Reverse bool
	Numeric bool
}

func (s *Sort) Name() string { return "sort" }

func (s *Sort) Process(ctx context.Context, in io.Reader, out io.Writer) error {
	lines, err := readAllLines(in)
	if err != nil {
		return fmt.Errorf("sort: %w", err)
	}

	if s.Numeric {
		sort.SliceStable(lines, func(i, j int) bool {
			a, _ := strconv.ParseFloat(strings.TrimSpace(lines[i]), 64)
			b, _ := strconv.ParseFloat(strings.TrimSpace(lines[j]), 64)
			if s.Reverse {
				return a > b
			}
			return a < b
		})
	} else {
		sort.SliceStable(lines, func(i, j int) bool {
			if s.Reverse {
				return lines[i] > lines[j]
			}
			return lines[i] < lines[j]
		})
	}

	for _, line := range lines {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if _, err := fmt.Fprintln(out, line); err != nil {
			return nil
		}
	}

	return nil
}

// Tail outputs the last N lines.
type Tail struct {
	N int
}

func (s *Tail) Name() string { return "tail" }

func (s *Tail) Process(ctx context.Context, in io.Reader, out io.Writer) error {
	n := s.N
	if n <= 0 {
		n = 10
	}

	lines, err := readAllLines(in)
	if err != nil {
		return fmt.Errorf("tail: %w", err)
	}

	start := 0
	if len(lines) > n {
		start = len(lines) - n
	}

	for _, line := range lines[start:] {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if _, err := fmt.Fprintln(out, line); err != nil {
			return nil
		}
	}

	return nil
}

// Tac reverses line order.
type Tac struct{}

func (s *Tac) Name() string { return "tac" }

func (s *Tac) Process(ctx context.Context, in io.Reader, out io.Writer) error {
	lines, err := readAllLines(in)
	if err != nil {
		return fmt.Errorf("tac: %w", err)
	}

	for i := len(lines) - 1; i >= 0; i-- {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if _, err := fmt.Fprintln(out, lines[i]); err != nil {
			return nil
		}
	}

	return nil
}

// Wc counts lines, words, and characters.
type Wc struct {
	Lines bool
	Words bool
	Chars bool
}

func (s *Wc) Name() string { return "wc" }

func (s *Wc) Process(ctx context.Context, in io.Reader, out io.Writer) error {
	lines, err := readAllLines(in)
	if err != nil {
		return fmt.Errorf("wc: %w", err)
	}

	lineCount := len(lines)
	wordCount := 0
	charCount := 0

	for _, line := range lines {
		wordCount += len(strings.Fields(line))
		charCount += len(line) + 1 // +1 for newline
	}

	// If no flags set, show all
	showAll := !s.Lines && !s.Words && !s.Chars

	var parts []string
	if showAll || s.Lines {
		parts = append(parts, strconv.Itoa(lineCount))
	}

	if showAll || s.Words {
		parts = append(parts, strconv.Itoa(wordCount))
	}

	if showAll || s.Chars {
		parts = append(parts, strconv.Itoa(charCount))
	}

	_, _ = fmt.Fprintln(out, strings.Join(parts, "\t"))

	return nil
}

// readAllLines reads all lines from a reader.
func readAllLines(r io.Reader) ([]string, error) {
	var lines []string

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}
