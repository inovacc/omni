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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunBasename(&buf, tt.args, tt.suffix)

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
