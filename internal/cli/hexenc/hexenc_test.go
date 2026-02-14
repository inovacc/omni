package hexenc

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/output"
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
			name: "simple text",
			args: []string{"hello"},
			opts: Options{},
			want: "68656c6c6f",
		},
		{
			name: "uppercase",
			args: []string{"hello"},
			opts: Options{Uppercase: true},
			want: "68656C6C6F",
		},
		{
			name: "with spaces",
			args: []string{"hello world"},
			opts: Options{},
			want: "68656c6c6f20776f726c64",
		},
		{
			name: "special characters",
			args: []string{"!@#$%"},
			opts: Options{},
			want: "2140232425",
		},
		{
			name: "unicode",
			args: []string{"日本"},
			opts: Options{},
			want: "e697a5e69cac",
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
			name: "simple hex",
			args: []string{"68656c6c6f"},
			opts: Options{},
			want: "hello",
		},
		{
			name: "uppercase hex",
			args: []string{"68656C6C6F"},
			opts: Options{},
			want: "hello",
		},
		{
			name: "with colons",
			args: []string{"68:65:6c:6c:6f"},
			opts: Options{},
			want: "hello",
		},
		{
			name: "with spaces",
			args: []string{"68 65 6c 6c 6f"},
			opts: Options{},
			want: "hello",
		},
		{
			name: "with dashes",
			args: []string{"68-65-6c-6c-6f"},
			opts: Options{},
			want: "hello",
		},
		{
			name: "unicode",
			args: []string{"e697a5e69cac"},
			opts: Options{},
			want: "日本",
		},
		{
			name: "empty string",
			args: []string{""},
			opts: Options{},
			want: "",
		},
		{
			name:    "invalid hex",
			args:    []string{"xyz"},
			opts:    Options{},
			wantErr: true,
		},
		{
			name:    "odd length",
			args:    []string{"abc"},
			opts:    Options{},
			wantErr: true,
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

			if tt.wantErr {
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

	opts := Options{OutputFormat: output.FormatJSON}

	err := RunEncode(&buf, []string{"hello"}, opts)
	if err != nil {
		t.Fatalf("RunEncode() error = %v", err)
	}

	var result Result
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result.Input != "hello" {
		t.Errorf("Input = %q, want %q", result.Input, "hello")
	}

	if result.Output != "68656c6c6f" {
		t.Errorf("Output = %q, want %q", result.Output, "68656c6c6f")
	}

	if result.Mode != "encode" {
		t.Errorf("Mode = %q, want %q", result.Mode, "encode")
	}
}

func TestRunDecodeJSON(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{OutputFormat: output.FormatJSON}

	err := RunDecode(&buf, []string{"68656c6c6f"}, opts)
	if err != nil {
		t.Fatalf("RunDecode() error = %v", err)
	}

	var result Result
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result.Input != "68656c6c6f" {
		t.Errorf("Input = %q, want %q", result.Input, "68656c6c6f")
	}

	if result.Output != "hello" {
		t.Errorf("Output = %q, want %q", result.Output, "hello")
	}

	if result.Mode != "decode" {
		t.Errorf("Mode = %q, want %q", result.Mode, "decode")
	}
}

func TestRoundTrip(t *testing.T) {
	original := "Hello, 世界! Special: !@#$%^&*()"

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
