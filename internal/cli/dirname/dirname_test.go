package dirname

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDirname(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple path",
			args:     []string{filepath.Join("a", "b", "c")},
			expected: filepath.Join("a", "b"),
		},
		{
			name:     "file in current directory",
			args:     []string{"file.txt"},
			expected: ".",
		},
		{
			name:     "absolute path",
			args:     []string{filepath.Join(string(filepath.Separator), "usr", "local", "bin")},
			expected: filepath.Join(string(filepath.Separator), "usr", "local"),
		},
		{
			name:     "multiple paths",
			args:     []string{filepath.Join("a", "b"), filepath.Join("x", "y", "z")},
			expected: "a\n" + filepath.Join("x", "y"),
		},
		{
			name:     "root path",
			args:     []string{string(filepath.Separator)},
			expected: string(filepath.Separator),
		},
		{
			name:     "single component",
			args:     []string{"dirname"},
			expected: ".",
		},
		{
			name:     "trailing slash",
			args:     []string{filepath.Join("a", "b", "c") + string(filepath.Separator)},
			expected: filepath.Join("a", "b", "c"), // With trailing slash, dirname returns the path itself
		},
		{
			name:     "dot path",
			args:     []string{"."},
			expected: ".",
		},
		{
			name:     "double dot path",
			args:     []string{".."},
			expected: ".",
		},
		{
			name:     "hidden directory",
			args:     []string{filepath.Join(".hidden", "file")},
			expected: ".hidden",
		},
		{
			name:    "no arguments",
			args:    []string{},
			wantErr: true,
		},
		{
			name:     "path with extension",
			args:     []string{filepath.Join("dir", "subdir", "file.tar.gz")},
			expected: filepath.Join("dir", "subdir"),
		},
		{
			name:     "deeply nested path",
			args:     []string{filepath.Join("a", "b", "c", "d", "e", "f")},
			expected: filepath.Join("a", "b", "c", "d", "e"),
		},
		{
			name:     "path with spaces",
			args:     []string{filepath.Join("path", "to my", "file.txt")},
			expected: filepath.Join("path", "to my"),
		},
		{
			name:     "relative dot path",
			args:     []string{filepath.Join(".", "dir", "file.txt")},
			expected: filepath.Join(".", "dir"),
		},
		{
			name:     "relative dotdot path",
			args:     []string{filepath.Join("..", "dir", "file.txt")},
			expected: filepath.Join("..", "dir"),
		},
		{
			name:     "three paths",
			args:     []string{filepath.Join("a", "1"), filepath.Join("b", "2"), filepath.Join("c", "3")},
			expected: "a\nb\nc",
		},
		{
			name:     "single file in subdir",
			args:     []string{filepath.Join("dir", "file")},
			expected: "dir",
		},
		{
			name:     "hidden file in hidden dir",
			args:     []string{filepath.Join(".config", ".settings")},
			expected: ".config",
		},
		{
			name:     "file with multiple extensions",
			args:     []string{filepath.Join("dir", "archive.tar.gz")},
			expected: "dir",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunDirname(&buf, tt.args, DirnameOptions{})

			if (err != nil) != tt.wantErr {
				t.Errorf("RunDirname() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got := strings.TrimSpace(buf.String())
				if got != tt.expected {
					t.Errorf("RunDirname() = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

func TestRunDirname_OutputFormat(t *testing.T) {
	t.Run("ends with newline", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunDirname(&buf, []string{filepath.Join("path", "file.txt")}, DirnameOptions{})

		if !strings.HasSuffix(buf.String(), "\n") {
			t.Error("RunDirname() output should end with newline")
		}
	})

	t.Run("multiple outputs each on new line", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunDirname(&buf, []string{
			filepath.Join("a", "1.txt"),
			filepath.Join("b", "2.txt"),
			filepath.Join("c", "3.txt"),
		}, DirnameOptions{})

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Errorf("RunDirname() got %d lines, want 3", len(lines))
		}
	})

	t.Run("error message format", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunDirname(&buf, []string{}, DirnameOptions{})

		if err == nil {
			t.Error("RunDirname() should return error for no args")
			return
		}

		if !strings.Contains(err.Error(), "dirname") {
			t.Errorf("RunDirname() error should mention dirname: %v", err)
		}
	})

	t.Run("no extra whitespace", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunDirname(&buf, []string{filepath.Join("path", "file.txt")}, DirnameOptions{})

		output := buf.String()
		if strings.HasPrefix(output, " ") || strings.HasPrefix(output, "\t") {
			t.Error("RunDirname() output should not have leading whitespace")
		}
	})
}

func TestRunDirname_EdgeCases(t *testing.T) {
	t.Run("unicode path", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunDirname(&buf, []string{filepath.Join("путь", "файл.txt")}, DirnameOptions{})

		got := strings.TrimSpace(buf.String())
		if got != "путь" {
			t.Errorf("RunDirname() unicode = %v, want путь", got)
		}
	})

	t.Run("path with special characters", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunDirname(&buf, []string{filepath.Join("dir-name_v1", "file.txt")}, DirnameOptions{})

		got := strings.TrimSpace(buf.String())
		if got != "dir-name_v1" {
			t.Errorf("RunDirname() special chars = %v", got)
		}
	})

	t.Run("very long path", func(t *testing.T) {
		longDir := strings.Repeat("a", 100)
		path := filepath.Join(longDir, "file.txt")

		var buf bytes.Buffer

		_ = RunDirname(&buf, []string{path}, DirnameOptions{})

		got := strings.TrimSpace(buf.String())
		if got != longDir {
			t.Errorf("RunDirname() long path = %v", got)
		}
	})

	t.Run("many path components", func(t *testing.T) {
		parts := make([]string, 20)
		for i := range parts {
			parts[i] = "dir"
		}

		path := filepath.Join(parts...)

		var buf bytes.Buffer

		_ = RunDirname(&buf, []string{path}, DirnameOptions{})

		got := strings.TrimSpace(buf.String())
		expected := filepath.Join(parts[:19]...)

		if got != expected {
			t.Errorf("RunDirname() many components = %v, want %v", got, expected)
		}
	})
}
