package htmlfmt

import (
	"strings"
	"testing"
)

func TestMinifyNodePaths(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantContain []string
		wantAbsent  []string
	}{
		{
			name:        "doctype preserved in minify",
			input:       "<!DOCTYPE html><html><body><p>hi</p></body></html>",
			wantContain: []string{"<!DOCTYPE html>", "<p>hi</p>"},
		},
		{
			name:        "comments stripped in minify",
			input:       "<div><!-- a comment -->text</div>",
			wantContain: []string{"<div>", "text", "</div>"},
			wantAbsent:  []string{"a comment", "<!--"},
		},
		{
			name:        "self-closing tag in minify",
			input:       "<div><br><img src=\"x.png\"></div>",
			wantContain: []string{"<br/>", "<img", "src=\"x.png\"", "/>"},
		},
		{
			name:        "whitespace collapsed in minify",
			input:       "<p>a     b\n\tc</p>",
			wantContain: []string{"<p>a b c</p>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Minify(tt.input)
			if err != nil {
				t.Fatalf("Minify(%q) error: %v", tt.input, err)
			}
			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("Minify() = %q, want it to contain %q", got, want)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(got, absent) {
					t.Errorf("Minify() = %q, want it to NOT contain %q", got, absent)
				}
			}
		})
	}
}

func TestFormatNodePaths(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		opts        []Option
		wantContain []string
	}{
		{
			name:        "doctype and comment kept in format",
			input:       "<!DOCTYPE html><html><body><!-- note --><p>hi</p></body></html>",
			wantContain: []string{"<!DOCTYPE html>", "<!-- note -->", "<p>hi</p>"},
		},
		{
			name:        "self-closing tag formatted",
			input:       "<div><br><hr></div>",
			wantContain: []string{"<br />", "<hr />"},
		},
		{
			name:        "attributes sorted",
			input:       "<a zeta=\"1\" alpha=\"2\" mid=\"3\"></a>",
			opts:        []Option{WithSortAttrs()},
			wantContain: []string{`alpha="2" mid="3" zeta="1"`},
		},
		{
			name:        "attribute value escaped",
			input:       `<a title="a &amp; b"></a>`,
			wantContain: []string{"a &amp; b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Format(tt.input, tt.opts...)
			if err != nil {
				t.Fatalf("Format(%q) error: %v", tt.input, err)
			}
			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("Format() =\n%q\nwant it to contain %q", got, want)
				}
			}
		})
	}
}

func TestValidateHTMLEmpty(t *testing.T) {
	res := Validate("   ")
	if res.Valid {
		t.Errorf("Validate(blank) = valid, want invalid")
	}
	if res.Error == "" {
		t.Errorf("Validate(blank) expected an error message")
	}
}
