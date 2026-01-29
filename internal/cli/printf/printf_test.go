package printf

import (
	"bytes"
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
			name: "simple string",
			args: []string{"Hello, %s!", "World"},
			opts: Options{},
			want: "Hello, World!\n",
		},
		{
			name: "integer",
			args: []string{"Number: %d", "42"},
			opts: Options{},
			want: "Number: 42\n",
		},
		{
			name: "hex lowercase",
			args: []string{"Hex: %x", "255"},
			opts: Options{},
			want: "Hex: ff\n",
		},
		{
			name: "hex uppercase",
			args: []string{"Hex: %X", "255"},
			opts: Options{},
			want: "Hex: FF\n",
		},
		{
			name: "float",
			args: []string{"Pi: %.2f", "3.14159"},
			opts: Options{},
			want: "Pi: 3.14\n",
		},
		{
			name: "multiple args",
			args: []string{"%s is %d years old", "Alice", "25"},
			opts: Options{},
			want: "Alice is 25 years old\n",
		},
		{
			name: "width padding",
			args: []string{"|%10s|", "test"},
			opts: Options{},
			want: "|      test|\n",
		},
		{
			name: "left align",
			args: []string{"|%-10s|", "test"},
			opts: Options{},
			want: "|test      |\n",
		},
		{
			name: "zero padding",
			args: []string{"%08d", "42"},
			opts: Options{},
			want: "00000042\n",
		},
		{
			name: "no newline",
			args: []string{"Hello"},
			opts: Options{NoNewline: true},
			want: "Hello",
		},
		{
			name: "escape newline",
			args: []string{"Line1\\nLine2"},
			opts: Options{NoNewline: true},
			want: "Line1\nLine2",
		},
		{
			name: "escape tab",
			args: []string{"Col1\\tCol2"},
			opts: Options{NoNewline: true},
			want: "Col1\tCol2",
		},
		{
			name: "percent escape",
			args: []string{"100%%"},
			opts: Options{NoNewline: true},
			want: "100%",
		},
		{
			name: "octal",
			args: []string{"%o", "8"},
			opts: Options{},
			want: "10\n",
		},
		{
			name: "binary",
			args: []string{"%b", "5"},
			opts: Options{},
			want: "101\n",
		},
		{
			name: "quoted string",
			args: []string{"%q", "hello world"},
			opts: Options{},
			want: "\"hello world\"\n",
		},
		{
			name: "character",
			args: []string{"%c", "A"},
			opts: Options{NoNewline: true},
			want: "A",
		},
		{
			name: "scientific notation",
			args: []string{"%.2e", "1234.5"},
			opts: Options{},
			want: "1.23e+03\n",
		},
		{
			name: "no format specifiers",
			args: []string{"plain text"},
			opts: Options{},
			want: "plain text\n",
		},
		{
			name: "missing args uses empty/zero",
			args: []string{"%s %d"},
			opts: Options{},
			want: " 0\n",
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

			if got := buf.String(); got != tt.want {
				t.Errorf("Run() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRunError(t *testing.T) {
	var buf bytes.Buffer

	err := Run(&buf, []string{}, Options{})
	if err == nil {
		t.Error("Run() expected error for empty args")
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		format string
		args   []string
		want   string
	}{
		{"Hello, %s!", []string{"World"}, "Hello, World!"},
		{"%d + %d = %d", []string{"1", "2", "3"}, "1 + 2 = 3"},
		{"%%", nil, "%"},
		{"no specifiers", nil, "no specifiers"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			got, err := Format(tt.format, tt.args)
			if err != nil {
				t.Errorf("Format() error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProcessEscapes(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`\n`, "\n"},
		{`\t`, "\t"},
		{`\r`, "\r"},
		{`\\`, "\\"},
		{`\x41`, "A"},
		{`\101`, "A"},
		{`Hello\nWorld`, "Hello\nWorld"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := processEscapes(tt.input)
			if got != tt.want {
				t.Errorf("processEscapes(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestInvalidNumber(t *testing.T) {
	// Invalid numbers should default to 0
	var buf bytes.Buffer

	_ = Run(&buf, []string{"%d", "not-a-number"}, Options{})
	if !strings.Contains(buf.String(), "0") {
		t.Errorf("Invalid number should default to 0")
	}
}
