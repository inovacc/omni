package seq

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunSeq(t *testing.T) {
	t.Run("single argument", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunSeq(&buf, []string{"5"}, SeqOptions{})
		if err != nil {
			t.Fatalf("RunSeq() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 5 {
			t.Errorf("RunSeq(5) got %d lines, want 5", len(lines))
		}

		if lines[0] != "1" || lines[4] != "5" {
			t.Errorf("RunSeq(5) = %v, want 1..5", lines)
		}
	})

	t.Run("two arguments ascending", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunSeq(&buf, []string{"3", "7"}, SeqOptions{})
		if err != nil {
			t.Fatalf("RunSeq() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 5 {
			t.Errorf("RunSeq(3,7) got %d lines, want 5", len(lines))
		}

		if lines[0] != "3" || lines[4] != "7" {
			t.Errorf("RunSeq(3,7) = %v", lines)
		}
	})

	t.Run("two arguments descending", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunSeq(&buf, []string{"5", "1"}, SeqOptions{})
		if err != nil {
			t.Fatalf("RunSeq() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 5 {
			t.Errorf("RunSeq(5,1) got %d lines, want 5", len(lines))
		}

		if lines[0] != "5" || lines[4] != "1" {
			t.Errorf("RunSeq(5,1) = %v", lines)
		}
	})

	t.Run("three arguments with increment", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunSeq(&buf, []string{"1", "2", "10"}, SeqOptions{})
		if err != nil {
			t.Fatalf("RunSeq() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 5 {
			t.Errorf("RunSeq(1,2,10) got %d lines, want 5", len(lines))
		}

		expected := []string{"1", "3", "5", "7", "9"}
		for i, exp := range expected {
			if lines[i] != exp {
				t.Errorf("RunSeq(1,2,10)[%d] = %v, want %v", i, lines[i], exp)
			}
		}
	})

	t.Run("negative increment", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunSeq(&buf, []string{"10", "-2", "2"}, SeqOptions{})
		if err != nil {
			t.Fatalf("RunSeq() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 5 {
			t.Errorf("RunSeq(10,-2,2) got %d lines, want 5", len(lines))
		}
	})

	t.Run("decimal values", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunSeq(&buf, []string{"1", "0.5", "3"}, SeqOptions{})
		if err != nil {
			t.Fatalf("RunSeq() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "1.5") || !strings.Contains(output, "2.5") {
			t.Errorf("RunSeq() should handle decimals: %v", output)
		}
	})

	t.Run("custom separator", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunSeq(&buf, []string{"3"}, SeqOptions{Separator: ","})
		if err != nil {
			t.Fatalf("RunSeq() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output != "1,2,3" {
			t.Errorf("RunSeq() with separator = %v, want '1,2,3'", output)
		}
	})

	t.Run("equal width", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunSeq(&buf, []string{"8", "10"}, SeqOptions{EqualWidth: true})
		if err != nil {
			t.Fatalf("RunSeq() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if lines[0] != "08" || lines[1] != "09" || lines[2] != "10" {
			t.Errorf("RunSeq() equal width = %v", lines)
		}
	})

	t.Run("no arguments error", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunSeq(&buf, []string{}, SeqOptions{})
		if err == nil {
			t.Error("RunSeq() expected error with no arguments")
		}
	})

	t.Run("too many arguments error", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunSeq(&buf, []string{"1", "2", "3", "4"}, SeqOptions{})
		if err == nil {
			t.Error("RunSeq() expected error with too many arguments")
		}
	})

	t.Run("invalid argument error", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunSeq(&buf, []string{"abc"}, SeqOptions{})
		if err == nil {
			t.Error("RunSeq() expected error with invalid argument")
		}
	})

	t.Run("zero increment error", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunSeq(&buf, []string{"1", "0", "5"}, SeqOptions{})
		if err == nil {
			t.Error("RunSeq() expected error with zero increment")
		}
	})

	t.Run("single value", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunSeq(&buf, []string{"1"}, SeqOptions{})
		if err != nil {
			t.Fatalf("RunSeq() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output != "1" {
			t.Errorf("RunSeq(1) = %v, want '1'", output)
		}
	})

	t.Run("output ends with newline", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunSeq(&buf, []string{"3"}, SeqOptions{})

		if !strings.HasSuffix(buf.String(), "\n") {
			t.Error("RunSeq() output should end with newline")
		}
	})
}

func TestHasDecimalPart(t *testing.T) {
	tests := []struct {
		input    float64
		expected bool
	}{
		{1.0, false},
		{1.5, true},
		{0.0, false},
		{-1.5, true},
		{100.0, false},
		{0.001, true},
	}

	for _, tt := range tests {
		result := hasDecimalPart(tt.input)
		if result != tt.expected {
			t.Errorf("hasDecimalPart(%v) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestMaxPrecision(t *testing.T) {
	tests := []struct {
		nums     []float64
		expected int
	}{
		{[]float64{1.0, 2.0, 3.0}, 0},
		{[]float64{1.5, 2.0, 3.0}, 1},
		{[]float64{1.0, 2.55, 3.0}, 2},
		{[]float64{1.123, 2.12, 3.1}, 3},
	}

	for _, tt := range tests {
		result := maxPrecision(tt.nums...)
		if result != tt.expected {
			t.Errorf("maxPrecision(%v) = %v, want %v", tt.nums, result, tt.expected)
		}
	}
}
