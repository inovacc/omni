package fold

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunFold(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fold_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("fold long line", func(t *testing.T) {
		file := filepath.Join(tmpDir, "long.txt")

		longLine := strings.Repeat("a", 100) + "\n"
		if err := os.WriteFile(file, []byte(longLine), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunFold(&buf, []string{file}, FoldOptions{Width: 80})
		if err != nil {
			t.Fatalf("RunFold() error = %v", err)
		}

		lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
		if len(lines) != 2 {
			t.Errorf("RunFold() got %d lines, want 2", len(lines))
		}
	})

	t.Run("short line unchanged", func(t *testing.T) {
		file := filepath.Join(tmpDir, "short.txt")
		if err := os.WriteFile(file, []byte("short line\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunFold(&buf, []string{file}, FoldOptions{Width: 80})
		if err != nil {
			t.Fatalf("RunFold() error = %v", err)
		}

		if buf.String() != "short line\n" {
			t.Errorf("RunFold() = %q, want 'short line\\n'", buf.String())
		}
	})

	t.Run("custom width", func(t *testing.T) {
		file := filepath.Join(tmpDir, "width.txt")
		if err := os.WriteFile(file, []byte("1234567890\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunFold(&buf, []string{file}, FoldOptions{Width: 5})
		if err != nil {
			t.Fatalf("RunFold() error = %v", err)
		}

		lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
		if len(lines) != 2 {
			t.Errorf("RunFold() got %d lines, want 2", len(lines))
		}

		if lines[0] != "12345" {
			t.Errorf("RunFold() first line = %q, want '12345'", lines[0])
		}
	})

	t.Run("break at spaces", func(t *testing.T) {
		file := filepath.Join(tmpDir, "spaces.txt")
		if err := os.WriteFile(file, []byte("hello world foo bar\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunFold(&buf, []string{file}, FoldOptions{Width: 12, Spaces: true})
		if err != nil {
			t.Fatalf("RunFold() error = %v", err)
		}

		lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
		// Should break at space boundaries
		if len(lines) < 2 {
			t.Errorf("RunFold() got %d lines, want >= 2", len(lines))
		}
	})

	t.Run("fold by bytes", func(t *testing.T) {
		file := filepath.Join(tmpDir, "bytes.txt")
		if err := os.WriteFile(file, []byte("abcdefghij\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunFold(&buf, []string{file}, FoldOptions{Width: 5, Bytes: true})
		if err != nil {
			t.Fatalf("RunFold() error = %v", err)
		}

		lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
		if len(lines) != 2 {
			t.Errorf("RunFold() got %d lines, want 2", len(lines))
		}
	})

	t.Run("empty file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "empty.txt")
		if err := os.WriteFile(file, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunFold(&buf, []string{file}, FoldOptions{})
		if err != nil {
			t.Fatalf("RunFold() error = %v", err)
		}
	})

	t.Run("empty line", func(t *testing.T) {
		file := filepath.Join(tmpDir, "emptyline.txt")
		if err := os.WriteFile(file, []byte("\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunFold(&buf, []string{file}, FoldOptions{})
		if err != nil {
			t.Fatalf("RunFold() error = %v", err)
		}

		if buf.String() != "\n" {
			t.Errorf("RunFold() = %q, want '\\n'", buf.String())
		}
	})

	t.Run("multiple files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")

		_ = os.WriteFile(file1, []byte("line1\n"), 0644)
		_ = os.WriteFile(file2, []byte("line2\n"), 0644)

		var buf bytes.Buffer

		err := RunFold(&buf, []string{file1, file2}, FoldOptions{})
		if err != nil {
			t.Fatalf("RunFold() error = %v", err)
		}

		if buf.String() != "line1\nline2\n" {
			t.Errorf("RunFold() = %q, want 'line1\\nline2\\n'", buf.String())
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunFold(&buf, []string{"/nonexistent/file.txt"}, FoldOptions{})
		if err == nil {
			t.Error("RunFold() expected error for nonexistent file")
		}
	})

	t.Run("default width is 80", func(t *testing.T) {
		file := filepath.Join(tmpDir, "default.txt")

		line := strings.Repeat("x", 85) + "\n"
		if err := os.WriteFile(file, []byte(line), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunFold(&buf, []string{file}, FoldOptions{})
		if err != nil {
			t.Fatalf("RunFold() error = %v", err)
		}

		lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
		if len(lines) != 2 {
			t.Errorf("RunFold() got %d lines, want 2 (default width 80)", len(lines))
		}
	})

	t.Run("unicode characters", func(t *testing.T) {
		file := filepath.Join(tmpDir, "unicode.txt")
		// 5 unicode chars
		if err := os.WriteFile(file, []byte("世界你好啊\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunFold(&buf, []string{file}, FoldOptions{Width: 3})
		if err != nil {
			t.Fatalf("RunFold() error = %v", err)
		}

		lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
		if len(lines) != 2 {
			t.Errorf("RunFold() got %d lines, want 2", len(lines))
		}
	})
}

func TestFoldLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		opts     FoldOptions
		expected int // expected number of resulting lines
	}{
		{"empty line", "", FoldOptions{Width: 80}, 1},
		{"short line", "hello", FoldOptions{Width: 80}, 1},
		{"exact width", "12345", FoldOptions{Width: 5}, 1},
		{"needs fold", "1234567890", FoldOptions{Width: 5}, 2},
		{"triple fold", "123456789012345", FoldOptions{Width: 5}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := foldLine(tt.line, tt.opts)
			if len(result) != tt.expected {
				t.Errorf("foldLine(%q) = %d lines, want %d", tt.line, len(result), tt.expected)
			}
		})
	}
}
