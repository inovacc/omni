package xxd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunDump(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		opts     Options
		contains []string
	}{
		{
			name:  "basic dump",
			input: "Hello, World!",
			opts:  DefaultOptions(),
			contains: []string{
				"00000000:",
				"4865 6c6c 6f2c",
				"Hello, World!",
			},
		},
		{
			name:  "uppercase",
			input: "hello",
			opts: Options{
				Columns:   16,
				Groups:    2,
				Uppercase: true,
			},
			contains: []string{
				"6865 6C6C 6F",
			},
		},
		{
			name:  "custom columns",
			input: "Hello, World!",
			opts: Options{
				Columns: 8,
				Groups:  2,
			},
			contains: []string{
				"00000000:",
				"Hello, W",
				"00000008:",
				"orld!",
			},
		},
		{
			name:  "binary data",
			input: "\x00\x01\x02\x03\xff\xfe",
			opts:  DefaultOptions(),
			contains: []string{
				"0001 0203 fffe",
				"......",
			},
		},
		{
			name:  "single byte groups",
			input: "hello",
			opts: Options{
				Columns: 16,
				Groups:  1,
			},
			contains: []string{
				"68 65 6c 6c 6f",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			r := strings.NewReader(tt.input)

			err := Run(&buf, r, nil, tt.opts)
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}

			output := buf.String()
			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("output does not contain %q\nGot:\n%s", s, output)
				}
			}
		})
	}
}

func TestRunPlain(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		opts     Options
		expected string
	}{
		{
			name:  "basic plain",
			input: "hello",
			opts: Options{
				Plain: true,
			},
			expected: "68656c6c6f",
		},
		{
			name:  "plain uppercase",
			input: "hello",
			opts: Options{
				Plain:     true,
				Uppercase: true,
			},
			expected: "68656C6C6F",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			r := strings.NewReader(tt.input)

			err := Run(&buf, r, nil, tt.opts)
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}

			output := strings.TrimSpace(buf.String())
			if output != tt.expected {
				t.Errorf("got %q, want %q", output, tt.expected)
			}
		})
	}
}

func TestRunInclude(t *testing.T) {
	var buf bytes.Buffer

	r := strings.NewReader("Hi")
	opts := Options{Include: true}

	err := Run(&buf, r, []string{"-"}, opts)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := buf.String()

	// Should contain C array syntax
	if !strings.Contains(output, "unsigned char") {
		t.Error("output should contain 'unsigned char'")
	}

	if !strings.Contains(output, "0x48") || !strings.Contains(output, "0x69") {
		t.Error("output should contain hex values for 'Hi'")
	}

	if !strings.Contains(output, "_len = 2") {
		t.Error("output should contain length declaration")
	}
}

func TestRunBits(t *testing.T) {
	var buf bytes.Buffer

	r := strings.NewReader("A")
	opts := Options{Bits: true, Columns: 6}

	err := Run(&buf, r, nil, opts)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := buf.String()

	// 'A' = 0x41 = 01000001
	if !strings.Contains(output, "01000001") {
		t.Errorf("output should contain binary for 'A' (01000001)\nGot: %s", output)
	}
}

func TestRunReverse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		opts     Options
		expected string
	}{
		{
			name:  "reverse plain",
			input: "68656c6c6f",
			opts: Options{
				Reverse: true,
				Plain:   true,
			},
			expected: "hello",
		},
		{
			name:  "reverse xxd format",
			input: "00000000: 4865 6c6c 6f2c 2057 6f72 6c64 21         Hello, World!",
			opts: Options{
				Reverse: true,
			},
			expected: "Hello, World!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			r := strings.NewReader(tt.input)

			err := Run(&buf, r, nil, tt.opts)
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}

			output := buf.String()
			if output != tt.expected {
				t.Errorf("got %q, want %q", output, tt.expected)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	original := "Hello, World! This is a test of the xxd command."

	// Dump to hex
	var dumpBuf bytes.Buffer

	err := Run(&dumpBuf, strings.NewReader(original), nil, DefaultOptions())
	if err != nil {
		t.Fatalf("dump error: %v", err)
	}

	// Reverse back to original
	var reverseBuf bytes.Buffer

	err = Run(&reverseBuf, strings.NewReader(dumpBuf.String()), nil, Options{Reverse: true})
	if err != nil {
		t.Fatalf("reverse error: %v", err)
	}

	if reverseBuf.String() != original {
		t.Errorf("round trip failed\nOriginal: %q\nGot: %q", original, reverseBuf.String())
	}
}

func TestRoundTripPlain(t *testing.T) {
	original := "Hello, World!"

	// Dump to plain hex
	var dumpBuf bytes.Buffer

	err := Run(&dumpBuf, strings.NewReader(original), nil, Options{Plain: true})
	if err != nil {
		t.Fatalf("dump error: %v", err)
	}

	// Reverse back to original
	var reverseBuf bytes.Buffer

	err = Run(&reverseBuf, strings.NewReader(dumpBuf.String()), nil, Options{Reverse: true, Plain: true})
	if err != nil {
		t.Fatalf("reverse error: %v", err)
	}

	if reverseBuf.String() != original {
		t.Errorf("round trip failed\nOriginal: %q\nGot: %q", original, reverseBuf.String())
	}
}

func TestLengthLimit(t *testing.T) {
	var buf bytes.Buffer

	r := strings.NewReader("Hello, World! This is a long string.")
	opts := Options{
		Columns: 16,
		Groups:  2,
		Length:  5,
	}

	err := Run(&buf, r, nil, opts)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := buf.String()

	// Should only contain "Hello"
	if !strings.Contains(output, "Hello") {
		t.Error("output should contain 'Hello'")
	}

	if strings.Contains(output, "World") {
		t.Error("output should not contain 'World' (limited to 5 bytes)")
	}
}

func TestSanitizeVarName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test.txt", "test_txt"},
		{"my-file.bin", "my_file_bin"},
		{"/path/to/file.dat", "file_dat"},
		{"123start", "_123start"},
		{"hello world", "hello_world"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeVarName(tt.input)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseXxdLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected []byte
	}{
		{
			name:     "standard line",
			line:     "00000000: 4865 6c6c 6f                             Hello",
			expected: []byte("Hello"),
		},
		{
			name:     "with different spacing",
			line:     "00000000: 48 65 6c 6c 6f  Hello",
			expected: []byte("Hello"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseXxdLine(tt.line)
			if err != nil {
				t.Fatalf("parseXxdLine() error = %v", err)
			}

			if !bytes.Equal(result, tt.expected) {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEmptyInput(t *testing.T) {
	var buf bytes.Buffer

	r := strings.NewReader("")

	err := Run(&buf, r, nil, DefaultOptions())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected empty output for empty input, got: %q", buf.String())
	}
}

func TestBinaryInput(t *testing.T) {
	// Test with all byte values 0-255
	input := make([]byte, 256)
	for i := range input {
		input[i] = byte(i)
	}

	var buf bytes.Buffer

	r := bytes.NewReader(input)

	err := Run(&buf, r, nil, DefaultOptions())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := buf.String()

	// Check first line contains 00
	if !strings.Contains(output, "00000000:") {
		t.Error("should have first offset")
	}

	// Check it contains ff somewhere
	if !strings.Contains(output, "ff") {
		t.Error("should contain 0xff byte")
	}
}
