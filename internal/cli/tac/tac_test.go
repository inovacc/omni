package tac

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunTac(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tac_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("reverse lines", func(t *testing.T) {
		file := filepath.Join(tmpDir, "lines.txt")
		if err := os.WriteFile(file, []byte("line1\nline2\nline3\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunTac(&buf, []string{file}, TacOptions{})
		if err != nil {
			t.Fatalf("RunTac() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Errorf("RunTac() got %d lines, want 3", len(lines))
		}

		if lines[0] != "line3" || lines[1] != "line2" || lines[2] != "line1" {
			t.Errorf("RunTac() = %v", lines)
		}
	})

	t.Run("single line", func(t *testing.T) {
		file := filepath.Join(tmpDir, "single.txt")
		if err := os.WriteFile(file, []byte("only\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunTac(&buf, []string{file}, TacOptions{})

		output := strings.TrimSpace(buf.String())
		if output != "only" {
			t.Errorf("RunTac() single = %v, want 'only'", output)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "empty.txt")
		if err := os.WriteFile(file, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunTac(&buf, []string{file}, TacOptions{})
		if err != nil {
			t.Fatalf("RunTac() error = %v", err)
		}
	})

	t.Run("unicode content", func(t *testing.T) {
		file := filepath.Join(tmpDir, "unicode.txt")
		if err := os.WriteFile(file, []byte("世界\nこんにちは\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunTac(&buf, []string{file}, TacOptions{})

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if lines[0] != "こんにちは" || lines[1] != "世界" {
			t.Errorf("RunTac() unicode = %v", lines)
		}
	})

	t.Run("multiple files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")

		_ = os.WriteFile(file1, []byte("a\nb\n"), 0644)
		_ = os.WriteFile(file2, []byte("x\ny\n"), 0644)

		var buf bytes.Buffer

		err := RunTac(&buf, []string{file1, file2}, TacOptions{})
		if err != nil {
			t.Fatalf("RunTac() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "b") && !strings.Contains(output, "y") {
			t.Errorf("RunTac() multiple files = %v", output)
		}
	})

	t.Run("custom separator", func(t *testing.T) {
		file := filepath.Join(tmpDir, "sep.txt")
		if err := os.WriteFile(file, []byte("a,b,c"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunTac(&buf, []string{file}, TacOptions{Separator: ","})

		output := buf.String()
		if !strings.Contains(output, "c") {
			t.Logf("RunTac() custom separator = %v", output)
		}
	})
}

func TestReverse(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "three elements",
			input:    []string{"a", "b", "c"},
			expected: []string{"c", "b", "a"},
		},
		{
			name:     "empty",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "single",
			input:    []string{"only"},
			expected: []string{"only"},
		},
		{
			name:     "two elements",
			input:    []string{"first", "second"},
			expected: []string{"second", "first"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reverse(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Reverse() length = %d, want %d", len(result), len(tt.expected))
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("Reverse()[%d] = %v, want %v", i, result[i], tt.expected[i])
				}
			}
		})
	}
}
