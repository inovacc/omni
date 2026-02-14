package htmlfmt

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
			name:  "simple element",
			input: "<div><p>text</p></div>",
			want:  "<html>\n  <head>\n  </head>\n  <body>\n    <div>\n      <p>text</p>\n    </div>\n  </body>\n</html>",
		},
		{
			name:  "with attributes",
			input: `<div class="container" id="main"><span>text</span></div>`,
			want:  "<html>\n  <head>\n  </head>\n  <body>\n    <div class=\"container\" id=\"main\">\n      <span>text</span>\n    </div>\n  </body>\n</html>",
		},
		{
			name:  "sort attributes",
			input: `<div z="1" a="2" m="3"></div>`,
			opts:  []Option{WithSortAttrs()},
			want:  "<html>\n  <head>\n  </head>\n  <body>\n    <div a=\"2\" m=\"3\" z=\"1\">\n    </div>\n  </body>\n</html>",
		},
		{
			name:  "self-closing tags",
			input: "<br><img src='test.jpg'><hr>",
			want:  "<html>\n  <head>\n  </head>\n  <body>\n    <br />\n    <img src=\"test.jpg\" />\n    <hr />\n  </body>\n</html>",
		},
		{
			name:  "custom indent",
			input: "<div><p>text</p></div>",
			opts:  []Option{WithIndent("    ")},
			want:  "<html>\n    <head>\n    </head>\n    <body>\n        <div>\n            <p>text</p>\n        </div>\n    </body>\n</html>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Format(tt.input, tt.opts...)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("Format() =\n%s\nwant\n%s", got, tt.want)
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
			name:  "simple element",
			input: "<div>\n  <p>text</p>\n</div>",
			want:  "<html><head></head><body><div> <p>text</p> </div></body></html>",
		},
		{
			name:  "removes comments",
			input: "<div><!-- comment --><p>text</p></div>",
			want:  "<html><head></head><body><div><p>text</p></div></body></html>",
		},
		{
			name:  "collapses whitespace",
			input: "<div>    multiple    spaces    </div>",
			want:  "<html><head></head><body><div> multiple spaces </div></body></html>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Minify(tt.input)
			if err != nil {
				t.Fatalf("Minify() error = %v", err)
			}

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
			name:      "valid html",
			input:     "<div><p>text</p></div>",
			wantValid: true,
		},
		{
			name:      "valid with attributes",
			input:     `<div class="test">content</div>`,
			wantValid: true,
		},
		{
			name:      "empty input",
			input:     "",
			wantValid: false,
		},
		{
			name:      "whitespace only",
			input:     "   ",
			wantValid: false,
		},
		{
			name:      "self-closing",
			input:     "<br><hr><img>",
			wantValid: true,
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

func TestIsSelfClosing(t *testing.T) {
	tests := []struct {
		tag  string
		want bool
	}{
		{"br", true},
		{"hr", true},
		{"img", true},
		{"input", true},
		{"meta", true},
		{"link", true},
		{"div", false},
		{"span", false},
		{"p", false},
		{"BR", true},
		{"IMG", true},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			got := IsSelfClosing(tt.tag)
			if got != tt.want {
				t.Errorf("IsSelfClosing(%q) = %v, want %v", tt.tag, got, tt.want)
			}
		})
	}
}

func TestCollapseWhitespace(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello world", "hello world"},
		{"hello  world", "hello world"},
		{"hello\nworld", "hello world"},
		{"hello\t\n  world", "hello world"},
		{"  hello  ", " hello "},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := CollapseWhitespace(tt.input)
			if got != tt.want {
				t.Errorf("CollapseWhitespace(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
