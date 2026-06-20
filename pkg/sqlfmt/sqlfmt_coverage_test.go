package sqlfmt

import (
	"strings"
	"testing"
)

func TestWithIndentOption(t *testing.T) {
	tests := []struct {
		name   string
		indent string
		input  string
		want   string // substring expected in output
	}{
		{"tab indent", "\t", "SELECT a FROM (SELECT b FROM t)", "\t"},
		{"four spaces", "    ", "SELECT a FROM (SELECT b FROM t)", "    "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Format(tt.input, WithIndent(tt.indent))
			if !strings.Contains(got, tt.want) {
				t.Errorf("Format() with indent %q = %q, want it to contain %q", tt.indent, got, tt.want)
			}
		})
	}
}

func TestTokenizeComments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string // token that must appear
	}{
		{"line comment", "SELECT a -- a comment\nFROM t", "-- a comment\n"},
		{"line comment at eof", "SELECT a -- trailing", "-- trailing"},
		{"block comment", "SELECT /* note */ a", "/* note */"},
		{"block comment multiline", "SELECT /* one\ntwo */ a", "/* one\ntwo */"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenizeSQL(tt.input)
			found := false
			for _, tok := range tokens {
				if tok == tt.want {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("tokenizeSQL(%q) = %v, want a token %q", tt.input, tokens, tt.want)
			}
		})
	}
}

func TestTokenizeOperators(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string // operator token expected
	}{
		{"less-equal", "a <= b", "<="},
		{"greater-equal", "a >= b", ">="},
		{"not-equal-angle", "a <> b", "<>"},
		{"concat", "a || b", "||"},
		{"single greater", "a > b", ">"},
		{"single less", "a < b", "<"},
		{"plus", "a + b", "+"},
		{"star", "a * b", "*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenizeSQL(tt.input)
			found := false
			for _, tok := range tokens {
				if tok == tt.want {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("tokenizeSQL(%q) = %v, want operator %q", tt.input, tokens, tt.want)
			}
		})
	}
}

func TestTokenizeStringEscaping(t *testing.T) {
	// Doubled quote inside a string literal must stay one token.
	tokens := tokenizeSQL("SELECT 'it''s' FROM t")
	found := false
	for _, tok := range tokens {
		if tok == "'it''s'" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("tokenizeSQL did not keep escaped-quote literal intact: %v", tokens)
	}
}

func TestCheckBalancedQuotesEscaped(t *testing.T) {
	tests := []struct {
		name  string
		input string
		quote rune
		want  bool
	}{
		{"backslash-escaped quote balanced", `'a\'b'`, '\'', true},
		{"double backslash then quote", `'a\\' 'b'`, '\'', true},
		{"unbalanced single", `'abc`, '\'', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkBalancedQuotes(tt.input, tt.quote); got != tt.want {
				t.Errorf("checkBalancedQuotes(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
