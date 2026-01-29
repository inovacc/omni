package echo

import (
	"bytes"
	"testing"
)

func TestRunEcho(t *testing.T) {
	t.Run("simple string", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunEcho(&buf, []string{"hello", "world"}, EchoOptions{})
		if err != nil {
			t.Fatalf("RunEcho() error = %v", err)
		}

		if buf.String() != "hello world\n" {
			t.Errorf("RunEcho() = %q, want %q", buf.String(), "hello world\n")
		}
	})

	t.Run("no newline", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunEcho(&buf, []string{"hello"}, EchoOptions{NoNewline: true})
		if err != nil {
			t.Fatalf("RunEcho() error = %v", err)
		}

		if buf.String() != "hello" {
			t.Errorf("RunEcho() = %q, want %q", buf.String(), "hello")
		}
	})

	t.Run("empty args", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunEcho(&buf, []string{}, EchoOptions{})
		if err != nil {
			t.Fatalf("RunEcho() error = %v", err)
		}

		if buf.String() != "\n" {
			t.Errorf("RunEcho() = %q, want %q", buf.String(), "\n")
		}
	})

	t.Run("escape sequences enabled", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunEcho(&buf, []string{"hello\\tworld"}, EchoOptions{EnableEscapes: true})
		if err != nil {
			t.Fatalf("RunEcho() error = %v", err)
		}

		if buf.String() != "hello\tworld\n" {
			t.Errorf("RunEcho() = %q, want %q", buf.String(), "hello\tworld\n")
		}
	})

	t.Run("escape sequences disabled by default", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunEcho(&buf, []string{"hello\\tworld"}, EchoOptions{})
		if err != nil {
			t.Fatalf("RunEcho() error = %v", err)
		}

		if buf.String() != "hello\\tworld\n" {
			t.Errorf("RunEcho() = %q, want %q", buf.String(), "hello\\tworld\n")
		}
	})

	t.Run("escape disabled overrides enabled", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunEcho(&buf, []string{"hello\\tworld"}, EchoOptions{EnableEscapes: true, DisableEscapes: true})
		if err != nil {
			t.Fatalf("RunEcho() error = %v", err)
		}

		if buf.String() != "hello\\tworld\n" {
			t.Errorf("RunEcho() = %q, want %q", buf.String(), "hello\\tworld\n")
		}
	})
}

func TestInterpretEscapes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"newline", "hello\\nworld", "hello\nworld"},
		{"tab", "hello\\tworld", "hello\tworld"},
		{"backslash", "hello\\\\world", "hello\\world"},
		{"carriage return", "hello\\rworld", "hello\rworld"},
		{"bell", "hello\\aworld", "hello\aworld"},
		{"backspace", "hello\\bworld", "hello\bworld"},
		{"form feed", "hello\\fworld", "hello\fworld"},
		{"vertical tab", "hello\\vworld", "hello\vworld"},
		{"escape", "hello\\eworld", "hello\x1bworld"},
		{"stop output", "hello\\cworld", "hello"},
		{"octal", "hello\\0101world", "helloAworld"},
		{"hex", "hello\\x41world", "helloAworld"},
		{"unknown escape", "hello\\qworld", "hello\\qworld"},
		{"trailing backslash", "hello\\", "hello\\"},
		{"multiple escapes", "\\t\\n\\t", "\t\n\t"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpretEscapes(tt.input)
			if result != tt.expected {
				t.Errorf("interpretEscapes(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseOctal(t *testing.T) {
	tests := []struct {
		input       string
		maxDigits   int
		expectedVal int
		expectedLen int
	}{
		{"101", 3, 65, 3},  // 'A' in octal
		{"77", 3, 63, 2},   // Max 2-digit octal
		{"7abc", 3, 7, 1},  // Stops at non-octal
		{"", 3, 0, 0},      // Empty
		{"89", 3, 0, 0},    // Invalid octal digits
		{"1234", 3, 83, 3}, // Max 3 digits: 123 octal = 83
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			val, consumed := parseOctal(tt.input, tt.maxDigits)
			if val != tt.expectedVal || consumed != tt.expectedLen {
				t.Errorf("parseOctal(%q, %d) = (%d, %d), want (%d, %d)",
					tt.input, tt.maxDigits, val, consumed, tt.expectedVal, tt.expectedLen)
			}
		})
	}
}

func TestParseHex(t *testing.T) {
	tests := []struct {
		input       string
		maxDigits   int
		expectedVal int
		expectedLen int
	}{
		{"41", 2, 65, 2},  // 'A' in hex
		{"ff", 2, 255, 2}, // Max 2-digit hex lowercase
		{"FF", 2, 255, 2}, // Max 2-digit hex uppercase
		{"aB", 2, 171, 2}, // Mixed case
		{"1g", 2, 1, 1},   // Stops at non-hex
		{"", 2, 0, 0},     // Empty
		{"xyz", 2, 0, 0},  // Invalid hex digits
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			val, consumed := parseHex(tt.input, tt.maxDigits)
			if val != tt.expectedVal || consumed != tt.expectedLen {
				t.Errorf("parseHex(%q, %d) = (%d, %d), want (%d, %d)",
					tt.input, tt.maxDigits, val, consumed, tt.expectedVal, tt.expectedLen)
			}
		})
	}
}
