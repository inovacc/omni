package basename

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunBasename(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		suffix   string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple filename",
			args:     []string{"file.txt"},
			suffix:   "",
			expected: "file.txt",
		},
		{
			name:     "path with directory",
			args:     []string{filepath.Join("a", "b", "c.txt")},
			suffix:   "",
			expected: "c.txt",
		},
		{
			name:     "remove suffix",
			args:     []string{filepath.Join("path", "to", "file.txt")},
			suffix:   ".txt",
			expected: "file",
		},
		{
			name:     "suffix not matching",
			args:     []string{"file.txt"},
			suffix:   ".go",
			expected: "file.txt",
		},
		{
			name:     "multiple paths",
			args:     []string{filepath.Join("a", "file1.txt"), filepath.Join("b", "file2.txt")},
			suffix:   "",
			expected: "file1.txt\nfile2.txt",
		},
		{
			name:     "trailing slash",
			args:     []string{filepath.Join("a", "b") + string(filepath.Separator)},
			suffix:   "",
			expected: "b",
		},
		{
			name:     "root path",
			args:     []string{string(filepath.Separator)},
			suffix:   "",
			expected: string(filepath.Separator),
		},
		{
			name:     "dot path",
			args:     []string{"."},
			suffix:   "",
			expected: ".",
		},
		{
			name:     "double dot path",
			args:     []string{".."},
			suffix:   "",
			expected: "..",
		},
		{
			name:     "hidden file",
			args:     []string{filepath.Join("dir", ".hidden")},
			suffix:   "",
			expected: ".hidden",
		},
		{
			name:    "no arguments",
			args:    []string{},
			suffix:  "",
			wantErr: true,
		},
		{
			name:     "remove extension suffix",
			args:     []string{"archive.tar.gz"},
			suffix:   ".tar.gz",
			expected: "archive",
		},
		{
			name:     "suffix same as filename",
			args:     []string{".txt"},
			suffix:   ".txt",
			expected: ".txt",
		},
		{
			name:     "just filename no extension",
			args:     []string{"filename"},
			suffix:   "",
			expected: "filename",
		},
		{
			name:     "multiple extensions",
			args:     []string{"file.backup.tar.gz"},
			suffix:   ".gz",
			expected: "file.backup.tar",
		},
		{
			name:     "suffix partial match",
			args:     []string{"testing"},
			suffix:   "ing",
			expected: "test",
		},
		{
			name:     "deeply nested path",
			args:     []string{filepath.Join("a", "b", "c", "d", "e", "file.txt")},
			suffix:   "",
			expected: "file.txt",
		},
		{
			name:     "path with spaces",
			args:     []string{filepath.Join("path", "to my", "file name.txt")},
			suffix:   "",
			expected: "file name.txt",
		},
		{
			name:     "remove suffix with spaces",
			args:     []string{"my file.txt"},
			suffix:   ".txt",
			expected: "my file",
		},
		{
			name:     "suffix longer than filename",
			args:     []string{"a.txt"},
			suffix:   "verylongsuffix",
			expected: "a.txt",
		},
		{
			name:     "three paths",
			args:     []string{"a.txt", "b.txt", "c.txt"},
			suffix:   "",
			expected: "a.txt\nb.txt\nc.txt",
		},
		{
			name:     "three paths with suffix",
			args:     []string{"a.txt", "b.txt", "c.go"},
			suffix:   ".txt",
			expected: "a\nb\nc.go",
		},
		{
			name:     "relative path dot",
			args:     []string{filepath.Join(".", "file.txt")},
			suffix:   "",
			expected: "file.txt",
		},
		{
			name:     "relative path dotdot",
			args:     []string{filepath.Join("..", "dir", "file.txt")},
			suffix:   "",
			expected: "file.txt",
		},
		{
			name:     "empty suffix with extension",
			args:     []string{"file.txt"},
			suffix:   "",
			expected: "file.txt",
		},
		{
			name:     "case sensitive suffix",
			args:     []string{"file.TXT"},
			suffix:   ".txt",
			expected: "file.TXT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunBasename(&buf, tt.args, BasenameOptions{Suffix: tt.suffix})

			if (err != nil) != tt.wantErr {
				t.Errorf("RunBasename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				got := strings.TrimSpace(buf.String())
				if got != tt.expected {
					t.Errorf("RunBasename() = %v, want %v", got, tt.expected)
				}
			}
		})
	}
}

func TestRunBasename_OutputFormat(t *testing.T) {
	t.Run("ends with newline", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunBasename(&buf, []string{"file.txt"}, BasenameOptions{})

		if !strings.HasSuffix(buf.String(), "\n") {
			t.Error("RunBasename() output should end with newline")
		}
	})

	t.Run("multiple outputs each on new line", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunBasename(&buf, []string{"a.txt", "b.txt", "c.txt"}, BasenameOptions{})

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Errorf("RunBasename() got %d lines, want 3", len(lines))
		}
	})

	t.Run("error message format", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunBasename(&buf, []string{}, BasenameOptions{})
		if err == nil {
			t.Error("RunBasename() should return error for no args")
			return
		}

		if !strings.Contains(err.Error(), "basename") {
			t.Errorf("RunBasename() error should mention basename: %v", err)
		}
	})

	t.Run("no extra whitespace", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunBasename(&buf, []string{"file.txt"}, BasenameOptions{})

		output := buf.String()
		if strings.HasPrefix(output, " ") || strings.HasPrefix(output, "\t") {
			t.Error("RunBasename() output should not have leading whitespace")
		}
	})
}

func TestRunBasename_EdgeCases(t *testing.T) {
	t.Run("unicode filename", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunBasename(&buf, []string{filepath.Join("path", "文件.txt")}, BasenameOptions{Suffix: ".txt"})

		got := strings.TrimSpace(buf.String())
		if got != "文件" {
			t.Errorf("RunBasename() unicode = %v, want 文件", got)
		}
	})

	t.Run("filename with special characters", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunBasename(&buf, []string{"file-name_v1.2.3.txt"}, BasenameOptions{Suffix: ".txt"})

		got := strings.TrimSpace(buf.String())
		if got != "file-name_v1.2.3" {
			t.Errorf("RunBasename() special chars = %v", got)
		}
	})

	t.Run("very long filename", func(t *testing.T) {
		longName := strings.Repeat("a", 200) + ".txt"

		var buf bytes.Buffer

		_ = RunBasename(&buf, []string{filepath.Join("path", longName)}, BasenameOptions{Suffix: ".txt"})

		got := strings.TrimSpace(buf.String())
		expected := strings.Repeat("a", 200)

		if got != expected {
			t.Errorf("RunBasename() long filename length = %d, want %d", len(got), len(expected))
		}
	})

	t.Run("suffix with special regex characters", func(t *testing.T) {
		var buf bytes.Buffer

		// Suffix should be treated literally, not as regex
		_ = RunBasename(&buf, []string{"file.a+b.txt"}, BasenameOptions{Suffix: ".a+b.txt"})

		got := strings.TrimSpace(buf.String())
		if got != "file" {
			t.Errorf("RunBasename() special suffix = %v, want file", got)
		}
	})
}
