// Package grep provides pattern matching on string slices.
package grep

import (
	"regexp"
	"strings"
)

// Options configures grep behavior.
type Options struct {
	IgnoreCase     bool // Case insensitive matching
	InvertMatch    bool // Select non-matching lines
	FixedStrings   bool // Interpret pattern as fixed string (no regex)
	WordRegexp     bool // Match whole words only
	LineRegexp     bool // Match whole lines only
	ExtendedRegexp bool // Interpret pattern as ERE (skip BRE conversion)
}

// Option is a functional option for Search.
type Option func(*Options)

// WithIgnoreCase enables case-insensitive matching.
func WithIgnoreCase() Option {
	return func(o *Options) { o.IgnoreCase = true }
}

// WithInvertMatch selects non-matching lines.
func WithInvertMatch() Option {
	return func(o *Options) { o.InvertMatch = true }
}

// WithFixedStrings interprets the pattern as a fixed string.
func WithFixedStrings() Option {
	return func(o *Options) { o.FixedStrings = true }
}

// WithWordRegexp matches whole words only.
func WithWordRegexp() Option {
	return func(o *Options) { o.WordRegexp = true }
}

// WithLineRegexp matches whole lines only.
func WithLineRegexp() Option {
	return func(o *Options) { o.LineRegexp = true }
}

// Search filters lines containing the pattern (simple string matching).
func Search(lines []string, pattern string) []string {
	var out []string

	for _, l := range lines {
		if strings.Contains(l, pattern) {
			out = append(out, l)
		}
	}

	return out
}

// SearchWithOptions filters lines using the given options.
func SearchWithOptions(lines []string, pattern string, opts ...Option) []string {
	o := Options{}
	for _, opt := range opts {
		opt(&o)
	}

	return searchWithOptions(lines, pattern, o)
}

// SearchWithOptionsStruct filters lines using the Options struct directly.
func SearchWithOptionsStruct(lines []string, pattern string, opt Options) []string {
	return searchWithOptions(lines, pattern, opt)
}

// CompilePattern compiles a grep pattern with the given options into a regexp.
func CompilePattern(pattern string, opts Options) (*regexp.Regexp, error) {
	return compilePattern(pattern, opts)
}

func searchWithOptions(lines []string, pattern string, opt Options) []string {
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

// convertBREtoERE converts a BRE (Basic Regular Expression) pattern to ERE
// (Extended Regular Expression). In BRE mode (default grep), metacharacters
// like |, (, ), {, }, +, ? must be backslash-escaped to have special meaning.
// Go's regexp package uses ERE, so we convert by unescaping these.
func convertBREtoERE(pattern string) string {
	var b strings.Builder
	b.Grow(len(pattern))

	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '\\' && i+1 < len(pattern) {
			next := pattern[i+1]
			switch next {
			case '|', '(', ')', '{', '}', '+', '?':
				b.WriteByte(next)
				i++ // skip the backslash
			default:
				b.WriteByte('\\')
				b.WriteByte(next)
				i++
			}
		} else {
			b.WriteByte(pattern[i])
		}
	}

	return b.String()
}

func compilePattern(pattern string, opts Options) (*regexp.Regexp, error) {
	if opts.FixedStrings {
		pattern = regexp.QuoteMeta(pattern)
	} else if !opts.ExtendedRegexp {
		// Default grep mode is BRE; convert to ERE for Go's regexp engine
		pattern = convertBREtoERE(pattern)
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
