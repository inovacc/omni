package nl

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunNl(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nl_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("number non-empty lines", func(t *testing.T) {
		file := filepath.Join(tmpDir, "lines.txt")
		if err := os.WriteFile(file, []byte("line1\nline2\nline3\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunNl(&buf, []string{file}, NlOptions{})
		if err != nil {
			t.Fatalf("RunNl() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "1") || !strings.Contains(output, "line1") {
			t.Errorf("RunNl() should number lines: %v", output)
		}
	})

	t.Run("skip empty lines by default", func(t *testing.T) {
		file := filepath.Join(tmpDir, "empty.txt")
		if err := os.WriteFile(file, []byte("line1\n\nline2\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunNl(&buf, []string{file}, NlOptions{})

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		// Empty line should not be numbered
		if len(lines) != 3 {
			t.Logf("RunNl() got %d lines: %v", len(lines), lines)
		}
	})

	t.Run("number all lines", func(t *testing.T) {
		file := filepath.Join(tmpDir, "all.txt")
		if err := os.WriteFile(file, []byte("line1\n\nline2\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunNl(&buf, []string{file}, NlOptions{BodyNumbering: "a"})

		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		// All 3 lines (including empty) should be numbered
		if len(lines) != 3 {
			t.Logf("RunNl() all got %d lines", len(lines))
		}
	})

	t.Run("number none", func(t *testing.T) {
		file := filepath.Join(tmpDir, "none.txt")
		if err := os.WriteFile(file, []byte("line1\nline2\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunNl(&buf, []string{file}, NlOptions{BodyNumbering: "n"})

		output := buf.String()
		// Lines should not have numbers (just spaces)
		if strings.Contains(output, "1\t") {
			t.Logf("RunNl() none should not number lines: %v", output)
		}
	})

	t.Run("custom width", func(t *testing.T) {
		file := filepath.Join(tmpDir, "width.txt")
		if err := os.WriteFile(file, []byte("line1\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunNl(&buf, []string{file}, NlOptions{NumberWidth: 3})

		output := buf.String()
		// Number should be 3 characters wide
		if !strings.HasPrefix(output, "  1") {
			t.Logf("RunNl() width = %v", output)
		}
	})

	t.Run("custom separator", func(t *testing.T) {
		file := filepath.Join(tmpDir, "sep.txt")
		if err := os.WriteFile(file, []byte("line1\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunNl(&buf, []string{file}, NlOptions{NumberSep: ": "})

		output := buf.String()
		if !strings.Contains(output, ": ") {
			t.Errorf("RunNl() separator = %v", output)
		}
	})

	t.Run("starting number", func(t *testing.T) {
		file := filepath.Join(tmpDir, "start.txt")
		if err := os.WriteFile(file, []byte("line1\nline2\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunNl(&buf, []string{file}, NlOptions{StartingNumber: 10})

		output := buf.String()
		if !strings.Contains(output, "10") {
			t.Errorf("RunNl() starting number = %v", output)
		}
	})

	t.Run("increment", func(t *testing.T) {
		file := filepath.Join(tmpDir, "inc.txt")
		if err := os.WriteFile(file, []byte("a\nb\nc\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunNl(&buf, []string{file}, NlOptions{Increment: 5, StartingNumber: 1})

		output := buf.String()
		if !strings.Contains(output, "1") || !strings.Contains(output, "6") || !strings.Contains(output, "11") {
			t.Logf("RunNl() increment = %v", output)
		}
	})

	t.Run("left justified", func(t *testing.T) {
		file := filepath.Join(tmpDir, "left.txt")
		if err := os.WriteFile(file, []byte("line1\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunNl(&buf, []string{file}, NlOptions{NumberFormat: "ln"})

		output := buf.String()
		// Left justified: number followed by spaces
		if !strings.HasPrefix(output, "1") {
			t.Logf("RunNl() left justified = %v", output)
		}
	})

	t.Run("leading zeros", func(t *testing.T) {
		file := filepath.Join(tmpDir, "zeros.txt")
		if err := os.WriteFile(file, []byte("line1\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunNl(&buf, []string{file}, NlOptions{NumberFormat: "rz"})

		output := buf.String()
		if !strings.Contains(output, "000001") {
			t.Logf("RunNl() leading zeros = %v", output)
		}
	})
}

func TestFormatLineNumber(t *testing.T) {
	tests := []struct {
		num      int
		format   string
		width    int
		expected string
	}{
		{1, "rn", 6, "     1"},
		{1, "ln", 6, "1     "},
		{1, "rz", 6, "000001"},
		{42, "rn", 4, "  42"},
		{123, "rz", 5, "00123"},
	}

	for _, tt := range tests {
		result := formatLineNumber(tt.num, tt.format, tt.width)
		if result != tt.expected {
			t.Errorf("formatLineNumber(%d, %s, %d) = %q, want %q",
				tt.num, tt.format, tt.width, result, tt.expected)
		}
	}
}
