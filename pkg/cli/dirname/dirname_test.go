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
			args:     []string{filepath.Join("a", "b") + string(filepath.Separator)},
			expected: "a",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunDirname(&buf, tt.args)

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
