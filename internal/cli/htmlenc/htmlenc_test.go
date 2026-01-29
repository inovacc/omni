package htmlenc

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunEncode(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		opts    Options
		want    string
		wantErr bool
	}{
		{
			name: "angle brackets",
			args: []string{"<div>"},
			opts: Options{},
			want: "&lt;div&gt;",
		},
		{
			name: "ampersand",
			args: []string{"Tom & Jerry"},
			opts: Options{},
			want: "Tom &amp; Jerry",
		},
		{
			name: "quotes",
			args: []string{`"hello"`},
			opts: Options{},
			want: "&#34;hello&#34;",
		},
		{
			name: "script tag",
			args: []string{"<script>alert('xss')</script>"},
			opts: Options{},
			want: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name: "plain text",
			args: []string{"hello world"},
			opts: Options{},
			want: "hello world",
		},
		{
			name: "empty string",
			args: []string{""},
			opts: Options{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunEncode(&buf, tt.args, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunEncode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got := strings.TrimSpace(buf.String())
			if got != tt.want {
				t.Errorf("RunEncode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRunDecode(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		opts    Options
		want    string
		wantErr bool
	}{
		{
			name: "angle brackets",
			args: []string{"&lt;div&gt;"},
			opts: Options{},
			want: "<div>",
		},
		{
			name: "ampersand",
			args: []string{"Tom &amp; Jerry"},
			opts: Options{},
			want: "Tom & Jerry",
		},
		{
			name: "quotes",
			args: []string{"&#34;hello&#34;"},
			opts: Options{},
			want: `"hello"`,
		},
		{
			name: "numeric entities",
			args: []string{"&#60;&#62;"},
			opts: Options{},
			want: "<>",
		},
		{
			name: "hex entities",
			args: []string{"&#x3C;&#x3E;"},
			opts: Options{},
			want: "<>",
		},
		{
			name: "named entities",
			args: []string{"&copy; &reg; &trade;"},
			opts: Options{},
			want: "© ® ™",
		},
		{
			name: "already decoded",
			args: []string{"hello world"},
			opts: Options{},
			want: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunDecode(&buf, tt.args, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunDecode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got := strings.TrimSpace(buf.String())
			if got != tt.want {
				t.Errorf("RunDecode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRunEncodeJSON(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{JSON: true}

	err := RunEncode(&buf, []string{"<div>"}, opts)
	if err != nil {
		t.Fatalf("RunEncode() error = %v", err)
	}

	var result Result
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result.Input != "<div>" {
		t.Errorf("Input = %q, want %q", result.Input, "<div>")
	}

	if result.Output != "&lt;div&gt;" {
		t.Errorf("Output = %q, want %q", result.Output, "&lt;div&gt;")
	}

	if result.Mode != "encode" {
		t.Errorf("Mode = %q, want %q", result.Mode, "encode")
	}
}

func TestRunDecodeJSON(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{JSON: true}

	err := RunDecode(&buf, []string{"&lt;div&gt;"}, opts)
	if err != nil {
		t.Fatalf("RunDecode() error = %v", err)
	}

	var result Result
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result.Input != "&lt;div&gt;" {
		t.Errorf("Input = %q, want %q", result.Input, "&lt;div&gt;")
	}

	if result.Output != "<div>" {
		t.Errorf("Output = %q, want %q", result.Output, "<div>")
	}

	if result.Mode != "decode" {
		t.Errorf("Mode = %q, want %q", result.Mode, "decode")
	}
}

func TestRoundTrip(t *testing.T) {
	original := `<script>alert("XSS & stuff")</script>`

	// Encode
	var encodeBuf bytes.Buffer

	err := RunEncode(&encodeBuf, []string{original}, Options{})
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	encoded := strings.TrimSpace(encodeBuf.String())

	// Decode
	var decodeBuf bytes.Buffer

	err = RunDecode(&decodeBuf, []string{encoded}, Options{})
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	decoded := strings.TrimSpace(decodeBuf.String())

	if decoded != original {
		t.Errorf("Round trip failed: got %q, want %q", decoded, original)
	}
}
