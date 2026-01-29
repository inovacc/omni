package paste

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPaste(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "paste_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("paste two files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")

		_ = os.WriteFile(file1, []byte("a\nb\nc\n"), 0644)
		_ = os.WriteFile(file2, []byte("1\n2\n3\n"), 0644)

		var buf bytes.Buffer

		err := RunPaste(&buf, nil, []string{file1, file2}, PasteOptions{})
		if err != nil {
			t.Fatalf("RunPaste() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "a\t1") {
			t.Errorf("RunPaste() = %q, missing expected tab-separated output", output)
		}
	})

	t.Run("custom delimiter", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "d1.txt")
		file2 := filepath.Join(tmpDir, "d2.txt")

		_ = os.WriteFile(file1, []byte("a\nb\n"), 0644)
		_ = os.WriteFile(file2, []byte("1\n2\n"), 0644)

		var buf bytes.Buffer

		err := RunPaste(&buf, nil, []string{file1, file2}, PasteOptions{Delimiters: ","})
		if err != nil {
			t.Fatalf("RunPaste() error = %v", err)
		}

		if !strings.Contains(buf.String(), "a,1") {
			t.Errorf("RunPaste() should use comma delimiter")
		}
	})

	t.Run("serial mode", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "s1.txt")
		file2 := filepath.Join(tmpDir, "s2.txt")

		_ = os.WriteFile(file1, []byte("a\nb\nc\n"), 0644)
		_ = os.WriteFile(file2, []byte("1\n2\n3\n"), 0644)

		var buf bytes.Buffer

		err := RunPaste(&buf, nil, []string{file1, file2}, PasteOptions{Serial: true})
		if err != nil {
			t.Fatalf("RunPaste() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Errorf("RunPaste() serial got %d lines, want 2", len(lines))
		}

		// First line should be a, b, c tab-separated
		if !strings.Contains(lines[0], "a\tb\tc") {
			t.Errorf("RunPaste() serial first line = %q", lines[0])
		}
	})

	t.Run("files with different lengths", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "long.txt")
		file2 := filepath.Join(tmpDir, "short.txt")

		_ = os.WriteFile(file1, []byte("a\nb\nc\nd\n"), 0644)
		_ = os.WriteFile(file2, []byte("1\n2\n"), 0644)

		var buf bytes.Buffer

		err := RunPaste(&buf, nil, []string{file1, file2}, PasteOptions{})
		if err != nil {
			t.Fatalf("RunPaste() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 4 {
			t.Errorf("RunPaste() got %d lines, want 4", len(lines))
		}
	})

	t.Run("cycling delimiters", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "c1.txt")
		file2 := filepath.Join(tmpDir, "c2.txt")
		file3 := filepath.Join(tmpDir, "c3.txt")

		_ = os.WriteFile(file1, []byte("a\n"), 0644)
		_ = os.WriteFile(file2, []byte("b\n"), 0644)
		_ = os.WriteFile(file3, []byte("c\n"), 0644)

		var buf bytes.Buffer

		err := RunPaste(&buf, nil, []string{file1, file2, file3}, PasteOptions{Delimiters: ",:"})
		if err != nil {
			t.Fatalf("RunPaste() error = %v", err)
		}

		// Should cycle: a,b:c
		output := strings.TrimSpace(buf.String())
		if output != "a,b:c" {
			t.Errorf("RunPaste() = %q, want 'a,b:c'", output)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "e1.txt")
		file2 := filepath.Join(tmpDir, "e2.txt")

		_ = os.WriteFile(file1, []byte("a\nb\n"), 0644)
		_ = os.WriteFile(file2, []byte(""), 0644)

		var buf bytes.Buffer

		err := RunPaste(&buf, nil, []string{file1, file2}, PasteOptions{})
		if err != nil {
			t.Fatalf("RunPaste() error = %v", err)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunPaste(&buf, nil, []string{"/nonexistent/file.txt"}, PasteOptions{})
		if err == nil {
			t.Error("RunPaste() expected error for nonexistent file")
		}
	})

	t.Run("single file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "single.txt")
		_ = os.WriteFile(file, []byte("line1\nline2\n"), 0644)

		var buf bytes.Buffer

		err := RunPaste(&buf, nil, []string{file}, PasteOptions{})
		if err != nil {
			t.Fatalf("RunPaste() error = %v", err)
		}

		if buf.String() != "line1\nline2\n" {
			t.Errorf("RunPaste() single file = %q", buf.String())
		}
	})
}

func TestExpandDelimiters(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"\\t", "\t"},
		{"\\n", "\n"},
		{"\\\\", "\\"},
		{"\\0", "\x00"},
		{",", ","},
		{",\\t:", ",\t:"},
		{"abc", "abc"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := expandDelimiters(tt.input)
			if result != tt.expected {
				t.Errorf("expandDelimiters(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
