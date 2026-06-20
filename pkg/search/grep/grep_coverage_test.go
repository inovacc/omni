package grep

import (
	"testing"
)

// TestSearchWithOptionsFallback exercises the fallback string-matching branch
// that runs when the pattern fails to compile as a regexp.
func TestSearchWithOptionsFallback(t *testing.T) {
	tests := []struct {
		name    string
		lines   []string
		pattern string
		opt     Options
		want    int
	}{
		{
			// "[invalid" is not a valid regex; ExtendedRegexp avoids BRE rewriting
			// so the literal bracket reaches regexp.Compile and fails -> fallback.
			name:    "invalid regex falls back to substring",
			lines:   []string{"a[invalid", "ok", "x[invalidy"},
			pattern: "[invalid",
			opt:     Options{ExtendedRegexp: true},
			want:    2,
		},
		{
			name:    "invalid regex fallback ignore case",
			lines:   []string{"A[INVALID", "ok", "[invalid"},
			pattern: "[Invalid",
			opt:     Options{ExtendedRegexp: true, IgnoreCase: true},
			want:    2,
		},
		{
			name:    "invalid regex fallback invert",
			lines:   []string{"a[invalid", "ok", "fine"},
			pattern: "[invalid",
			opt:     Options{ExtendedRegexp: true, InvertMatch: true},
			want:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := searchWithOptions(tt.lines, tt.pattern, tt.opt)
			if len(got) != tt.want {
				t.Errorf("searchWithOptions(%q) = %v (len %d), want len %d", tt.pattern, got, len(got), tt.want)
			}
		})
	}
}

// TestConvertBREtoERE covers the escape-handling branches directly, including
// the default branch where a non-special escaped char is preserved.
func TestConvertBREtoERE(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    string
	}{
		{"unescape pipe", `cat\|dog`, `cat|dog`},
		{"unescape parens", `\(foo\)`, `(foo)`},
		{"unescape braces", `a\{2\}`, `a{2}`},
		{"unescape plus and question", `a\+b\?`, `a+b?`},
		{"preserve digit escape", `\d`, `\d`},
		{"preserve word boundary", `\bword\b`, `\bword\b`},
		{"trailing backslash kept", `abc\`, `abc\`},
		{"no escapes unchanged", `plain`, `plain`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertBREtoERE(tt.pattern); got != tt.want {
				t.Errorf("convertBREtoERE(%q) = %q, want %q", tt.pattern, got, tt.want)
			}
		})
	}
}
