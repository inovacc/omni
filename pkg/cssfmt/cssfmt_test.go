package cssfmt

import (
	"testing"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
		opts  []Option
		want  string
	}{
		{
			name:  "simple rule",
			input: "body{margin:0;padding:0}",
			want:  "body {\n  margin: 0;\n  padding: 0;\n}",
		},
		{
			name:  "multiple rules",
			input: "body{margin:0}h1{color:red}",
			want:  "body {\n  margin: 0;\n}\nh1 {\n  color: red;\n}",
		},
		{
			name:  "sort properties",
			input: "body{z-index:1;color:red;background:white}",
			opts:  []Option{WithSortProps()},
			want:  "body {\n  background: white;\n  color: red;\n  z-index: 1;\n}",
		},
		{
			name:  "custom indent",
			input: "body{margin:0}",
			opts:  []Option{WithIndent("    ")},
			want:  "body {\n    margin: 0;\n}",
		},
		{
			name:  "with comments removed",
			input: "body{/* comment */margin:0}",
			want:  "body {\n  margin: 0;\n}",
		},
		{
			name:  "media query",
			input: "@media (max-width: 768px){body{margin:0}}",
			want:  "@media (max-width: 768px) {\n  body {\n    margin: 0;\n  }\n}",
		},
		{
			name:  "import at-rule",
			input: "@import url('style.css');body{margin:0}",
			want:  "@import url('style.css');\nbody {\n  margin: 0;\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Format(tt.input, tt.opts...)
			if got != tt.want {
				t.Errorf("Format() =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}

func TestMinify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple rule",
			input: "body {\n  margin: 0;\n  padding: 0;\n}",
			want:  "body{margin:0;padding:0}",
		},
		{
			name:  "multiple rules",
			input: "body {\n  margin: 0;\n}\n\nh1 {\n  color: red;\n}",
			want:  "body{margin:0}h1{color:red}",
		},
		{
			name:  "removes whitespace",
			input: "body    {    margin:    0;    }",
			want:  "body{margin:0}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Minify(tt.input)
			if got != tt.want {
				t.Errorf("Minify() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValid bool
	}{
		{
			name:      "valid css",
			input:     "body { margin: 0; }",
			wantValid: true,
		},
		{
			name:      "valid complex",
			input:     "@media (max-width: 768px) { body { margin: 0; } }",
			wantValid: true,
		},
		{
			name:      "unbalanced braces open",
			input:     "body { margin: 0;",
			wantValid: false,
		},
		{
			name:      "unbalanced braces close",
			input:     "body margin: 0; }",
			wantValid: false,
		},
		{
			name:      "unbalanced parens",
			input:     "body { background: url('test.png'; }",
			wantValid: false,
		},
		{
			name:      "unclosed string",
			input:     "body { content: 'hello; }",
			wantValid: false,
		},
		{
			name:      "empty input",
			input:     "",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Validate(tt.input)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v (error: %s)", result.Valid, tt.wantValid, result.Error)
			}
		})
	}
}

func TestRemoveComments(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"body { margin: 0; }", "body { margin: 0; }"},
		{"/* comment */ body { margin: 0; }", " body { margin: 0; }"},
		{"body { /* comment */ margin: 0; }", "body {  margin: 0; }"},
		{"body { margin: 0; } /* end */", "body { margin: 0; } "},
		{"/* a */ body /* b */ { margin: 0; } /* c */", " body  { margin: 0; } "},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := RemoveComments(tt.input)
			if got != tt.want {
				t.Errorf("RemoveComments(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseDeclarations(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"margin: 0", 1},
		{"margin: 0; padding: 0", 2},
		{"margin: 0; padding: 0;", 2},
		{"font-family: 'Arial', sans-serif", 1},
		{"background: url('test;image.png')", 1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseDeclarations(tt.input)
			if len(got) != tt.want {
				t.Errorf("ParseDeclarations(%q) returned %d declarations, want %d", tt.input, len(got), tt.want)
			}
		})
	}
}
