package xmlfmt

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		opts    Options
		want    string
		wantErr bool
	}{
		{
			name: "simple format",
			args: []string{"<root><item>value</item></root>"},
			opts: Options{},
			want: "<root>\n  <item>value</item>\n</root>",
		},
		{
			name: "nested elements",
			args: []string{"<root><parent><child>text</child></parent></root>"},
			opts: Options{},
			want: "<root>\n  <parent>\n    <child>text</child>\n  </parent>\n</root>",
		},
		{
			name: "custom indent",
			args: []string{"<root><item>value</item></root>"},
			opts: Options{Indent: "    "},
			want: "<root>\n    <item>value</item>\n</root>",
		},
		{
			name: "tab indent",
			args: []string{"<root><item>value</item></root>"},
			opts: Options{Indent: "\t"},
			want: "<root>\n\t<item>value</item>\n</root>",
		},
		{
			name: "with attributes",
			args: []string{`<root attr="val"><item id="1">value</item></root>`},
			opts: Options{},
			want: `<root attr="val">` + "\n  " + `<item id="1">value</item>` + "\n</root>",
		},
		{
			name: "minify",
			args: []string{"<root>\n  <item>value</item>\n</root>"},
			opts: Options{Minify: true},
			want: "<root><item>value</item></root>",
		},
		{
			name: "empty element",
			args: []string{"<root><empty/></root>"},
			opts: Options{},
			want: "<root>\n  <empty></empty>\n</root>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := Run(&buf, tt.args, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got := strings.TrimSpace(buf.String())
			if got != tt.want {
				t.Errorf("Run() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRunInvalidXML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{
			name:    "unclosed tag",
			input:   "<root><item>value</root>",
			wantErr: "xml:",
		},
		{
			name:    "mismatched tags",
			input:   "<root><item>value</other></root>",
			wantErr: "xml:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := Run(&buf, []string{tt.input}, Options{})
			if err == nil {
				t.Errorf("Run() expected error, got nil")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Run() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestMinifyPreservesContent(t *testing.T) {
	input := `<root>
  <item>
    some text content
  </item>
  <other>more content</other>
</root>`

	var buf bytes.Buffer

	err := Run(&buf, []string{input}, Options{Minify: true})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := strings.TrimSpace(buf.String())

	// Should contain text content
	if !strings.Contains(output, "some text content") {
		t.Errorf("Output should contain text content")
	}

	if !strings.Contains(output, "more content") {
		t.Errorf("Output should contain more content")
	}

	// Should not contain newlines
	if strings.Contains(output, "\n") {
		t.Errorf("Minified output should not contain newlines")
	}
}

func TestFormatXMLDeclaration(t *testing.T) {
	input := `<?xml version="1.0" encoding="UTF-8"?><root><item>value</item></root>`

	var buf bytes.Buffer

	err := Run(&buf, []string{input}, Options{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := buf.String()

	// Should contain XML declaration
	if !strings.Contains(output, "<?xml") {
		t.Errorf("Output should contain XML declaration")
	}
}

func TestFormatNamespaces(t *testing.T) {
	input := `<root xmlns="http://example.com" xmlns:ns="http://ns.example.com"><ns:item>value</ns:item></root>`

	var buf bytes.Buffer

	err := Run(&buf, []string{input}, Options{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := buf.String()

	// Should preserve namespace
	if !strings.Contains(output, "xmlns") {
		t.Errorf("Output should contain namespace declaration")
	}
}

func TestRoundTrip(t *testing.T) {
	original := `<root><parent><child attr="value">text content</child></parent></root>`

	// Format
	var formatBuf bytes.Buffer

	err := Run(&formatBuf, []string{original}, Options{})
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	formatted := formatBuf.String()

	// Minify
	var minifyBuf bytes.Buffer

	err = Run(&minifyBuf, []string{formatted}, Options{Minify: true})
	if err != nil {
		t.Fatalf("Minify error: %v", err)
	}

	minified := strings.TrimSpace(minifyBuf.String())

	// Should be equivalent to original (without whitespace)
	if minified != original {
		t.Errorf("Round trip failed:\ngot:  %q\nwant: %q", minified, original)
	}
}

func TestRunValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		opts    ValidateOptions
		wantErr bool
	}{
		{
			name:    "valid simple xml",
			input:   "<root><item>value</item></root>",
			opts:    ValidateOptions{},
			wantErr: false,
		},
		{
			name:    "valid with attributes",
			input:   `<root attr="value"><item id="1">text</item></root>`,
			opts:    ValidateOptions{},
			wantErr: false,
		},
		{
			name:    "valid with declaration",
			input:   `<?xml version="1.0"?><root/>`,
			opts:    ValidateOptions{},
			wantErr: false,
		},
		{
			name:    "valid with namespace",
			input:   `<root xmlns="http://example.com"><item>value</item></root>`,
			opts:    ValidateOptions{},
			wantErr: false,
		},
		{
			name:    "invalid - unclosed tag",
			input:   "<root><item>value</root>",
			opts:    ValidateOptions{},
			wantErr: true,
		},
		{
			name:    "invalid - mismatched tags",
			input:   "<root><item>value</other></root>",
			opts:    ValidateOptions{},
			wantErr: true,
		},
		{
			name:    "invalid - bad attribute",
			input:   `<root attr=value></root>`,
			opts:    ValidateOptions{},
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   "",
			opts:    ValidateOptions{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunValidate(&buf, []string{tt.input}, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunValidate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunValidateJSON(t *testing.T) {
	var buf bytes.Buffer

	opts := ValidateOptions{JSON: true}

	err := RunValidate(&buf, []string{"<root><item>value</item></root>"}, opts)
	if err != nil {
		t.Fatalf("RunValidate() error = %v", err)
	}

	var result ValidateResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if !result.Valid {
		t.Errorf("Valid = false, want true")
	}
}

func TestRunValidateJSONInvalid(t *testing.T) {
	var buf bytes.Buffer

	opts := ValidateOptions{JSON: true}

	_ = RunValidate(&buf, []string{"<root><item>value</root>"}, opts)

	var result ValidateResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result.Valid {
		t.Errorf("Valid = true, want false")
	}

	if result.Error == "" {
		t.Errorf("Error should not be empty")
	}
}
