package urlenc

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
			name: "simple text",
			args: []string{"hello world"},
			opts: Options{},
			want: "hello%20world",
		},
		{
			name: "special characters",
			args: []string{"a/b/c"},
			opts: Options{},
			want: "a%2Fb%2Fc",
		},
		{
			name: "component encoding",
			args: []string{"a=b&c=d"},
			opts: Options{Component: true},
			want: "a%3Db%26c%3Dd",
		},
		{
			name: "unicode",
			args: []string{"日本語"},
			opts: Options{},
			want: "%E6%97%A5%E6%9C%AC%E8%AA%9E",
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
			name: "simple encoded",
			args: []string{"hello%20world"},
			opts: Options{},
			want: "hello world",
		},
		{
			name: "path encoded",
			args: []string{"a%2Fb%2Fc"},
			opts: Options{},
			want: "a/b/c",
		},
		{
			name: "component decoding",
			args: []string{"a%3Db%26c%3Dd"},
			opts: Options{Component: true},
			want: "a=b&c=d",
		},
		{
			name: "unicode",
			args: []string{"%E6%97%A5%E6%9C%AC%E8%AA%9E"},
			opts: Options{},
			want: "日本語",
		},
		{
			name: "plus as space (component)",
			args: []string{"hello+world"},
			opts: Options{Component: true},
			want: "hello world",
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

	err := RunEncode(&buf, []string{"hello world"}, opts)
	if err != nil {
		t.Fatalf("RunEncode() error = %v", err)
	}

	var result Result
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result.Input != "hello world" {
		t.Errorf("Input = %q, want %q", result.Input, "hello world")
	}

	if result.Output != "hello%20world" {
		t.Errorf("Output = %q, want %q", result.Output, "hello%20world")
	}

	if result.Mode != "encode" {
		t.Errorf("Mode = %q, want %q", result.Mode, "encode")
	}
}

func TestRunDecodeJSON(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{JSON: true}

	err := RunDecode(&buf, []string{"hello%20world"}, opts)
	if err != nil {
		t.Fatalf("RunDecode() error = %v", err)
	}

	var result Result
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result.Input != "hello%20world" {
		t.Errorf("Input = %q, want %q", result.Input, "hello%20world")
	}

	if result.Output != "hello world" {
		t.Errorf("Output = %q, want %q", result.Output, "hello world")
	}

	if result.Mode != "decode" {
		t.Errorf("Mode = %q, want %q", result.Mode, "decode")
	}
}

func TestRoundTrip(t *testing.T) {
	original := "Hello, 世界! Special chars: a=b&c=d /path/to/file"

	// Encode
	var encodeBuf bytes.Buffer

	err := RunEncode(&encodeBuf, []string{original}, Options{Component: true})
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	encoded := strings.TrimSpace(encodeBuf.String())

	// Decode
	var decodeBuf bytes.Buffer

	err = RunDecode(&decodeBuf, []string{encoded}, Options{Component: true})
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}

	decoded := strings.TrimSpace(decodeBuf.String())

	if decoded != original {
		t.Errorf("Round trip failed: got %q, want %q", decoded, original)
	}
}
