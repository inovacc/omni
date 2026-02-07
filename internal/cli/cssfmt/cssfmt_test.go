package cssfmt

import (
	"bytes"
	"strings"
	"testing"

	pkgcss "github.com/inovacc/omni/pkg/cssfmt"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		opts    Options
		want    string
		wantErr bool
	}{
		{
			name:  "simple rule",
			input: "body{margin:0;padding:0}",
			opts:  Options{},
			want: `body {
  margin: 0;
  padding: 0;
}`,
		},
		{
			name:  "multiple rules",
			input: "body{margin:0}h1{color:red}",
			opts:  Options{},
			want: `body {
  margin: 0;
}
h1 {
  color: red;
}`,
		},
		{
			name:  "nested selectors",
			input: ".container .item{color:blue}",
			opts:  Options{},
			want: `.container .item {
  color: blue;
}`,
		},
		{
			name:  "multiple selectors",
			input: "h1, h2, h3{font-weight:bold}",
			opts:  Options{},
			want: `h1, h2, h3 {
  font-weight: bold;
}`,
		},
		{
			name:  "sort properties",
			input: "body{z-index:1;color:red;background:white}",
			opts:  Options{SortProps: true},
			want: `body {
  background: white;
  color: red;
  z-index: 1;
}`,
		},
		{
			name:  "custom indent",
			input: "body{margin:0}",
			opts:  Options{Indent: "    "},
			want: `body {
    margin: 0;
}`,
		},
		{
			name:  "with comments removed",
			input: "body{/* comment */margin:0}",
			opts:  Options{},
			want: `body {
  margin: 0;
}`,
		},
		{
			name:  "complex values",
			input: "body{font-family:'Arial', sans-serif;background:url('test.png')}",
			opts:  Options{},
			want: `body {
  font-family: 'Arial', sans-serif;
  background: url('test.png');
}`,
		},
		{
			name:  "media query",
			input: "@media (max-width: 768px){body{margin:0}}",
			opts:  Options{},
			want: `@media (max-width: 768px) {
  body {
    margin: 0;
  }
}`,
		},
		{
			name:  "import at-rule",
			input: "@import url('style.css');body{margin:0}",
			opts:  Options{},
			want: `@import url('style.css');
body {
  margin: 0;
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			r := strings.NewReader(tt.input)

			err := Run(&buf, r, nil, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got := strings.TrimSpace(buf.String())
			if got != tt.want {
				t.Errorf("Run() =\n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestRunMinify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name: "simple rule",
			input: `body {
  margin: 0;
  padding: 0;
}`,
			want: "body{margin:0;padding:0}",
		},
		{
			name: "multiple rules",
			input: `body {
  margin: 0;
}

h1 {
  color: red;
}`,
			want: "body{margin:0}h1{color:red}",
		},
		{
			name:  "removes whitespace",
			input: "body    {    margin:    0;    }",
			want:  "body{margin:0}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			r := strings.NewReader(tt.input)

			err := RunMinify(&buf, r, nil, Options{})
			if err != nil {
				t.Errorf("RunMinify() error = %v", err)
				return
			}

			got := strings.TrimSpace(buf.String())
			if got != tt.want {
				t.Errorf("RunMinify() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRunValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid css",
			input:   "body { margin: 0; }",
			wantErr: false,
		},
		{
			name:    "valid complex",
			input:   "@media (max-width: 768px) { body { margin: 0; } }",
			wantErr: false,
		},
		{
			name:    "unbalanced braces open",
			input:   "body { margin: 0;",
			wantErr: true,
		},
		{
			name:    "unbalanced braces close",
			input:   "body margin: 0; }",
			wantErr: true,
		},
		{
			name:    "unbalanced parens",
			input:   "body { background: url('test.png'; }",
			wantErr: true,
		},
		{
			name:    "unclosed string",
			input:   "body { content: 'hello; }",
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			r := strings.NewReader(tt.input)

			err := RunValidate(&buf, r, nil, ValidateOptions{})
			if (err != nil) != tt.wantErr {
				t.Errorf("RunValidate() error = %v, wantErr %v", err, tt.wantErr)
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
			got := pkgcss.RemoveComments(tt.input)
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
		{"background: url('test;image.png')", 1}, // semicolon in URL
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := pkgcss.ParseDeclarations(tt.input)
			if len(got) != tt.want {
				t.Errorf("ParseDeclarations(%q) returned %d declarations, want %d", tt.input, len(got), tt.want)
			}
		})
	}
}
