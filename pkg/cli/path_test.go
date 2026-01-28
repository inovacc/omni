package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDirname(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "simple path",
			input:  filepath.Join("a", "b", "c"),
			expect: filepath.Join("a", "b"),
		},
		{
			name:   "single component",
			input:  "filename",
			expect: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunDirname(&buf, []string{tt.input})
			if err != nil {
				t.Fatalf("RunDirname() error = %v", err)
			}

			result := strings.TrimSpace(buf.String())
			if result != tt.expect {
				t.Errorf("RunDirname() = %v, want %v", result, tt.expect)
			}
		})
	}
}

func TestRunBasename(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		suffix string
		expect string
	}{
		{
			name:   "simple path",
			input:  filepath.Join("a", "b", "file.txt"),
			suffix: "",
			expect: "file.txt",
		},
		{
			name:   "with suffix removal",
			input:  filepath.Join("a", "b", "file.txt"),
			suffix: ".txt",
			expect: "file",
		},
		{
			name:   "no directory",
			input:  "filename.go",
			suffix: ".go",
			expect: "filename",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunBasename(&buf, []string{tt.input}, tt.suffix)
			if err != nil {
				t.Fatalf("RunBasename() error = %v", err)
			}

			result := strings.TrimSpace(buf.String())
			if result != tt.expect {
				t.Errorf("RunBasename() = %v, want %v", result, tt.expect)
			}
		})
	}

	t.Run("no arguments", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunBasename(&buf, []string{}, "")
		if err == nil {
			t.Error("RunBasename() expected error with no arguments")
		}
	})
}

func TestRunRealpath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "realpath_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("existing file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "testfile.txt")
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunRealpath(&buf, []string{file})
		if err != nil {
			t.Fatalf("RunRealpath() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		expected, _ := filepath.Abs(file)
		// Handle symlinks
		expectedReal, _ := filepath.EvalSymlinks(expected)
		if result != expected && result != expectedReal {
			t.Errorf("RunRealpath() = %v, want %v or %v", result, expected, expectedReal)
		}
	})

	t.Run("relative path", func(t *testing.T) {
		origDir, _ := os.Getwd()

		defer func() { _ = os.Chdir(origDir) }()

		_ = os.Chdir(tmpDir)

		file := "localfile.txt"
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunRealpath(&buf, []string{file})
		if err != nil {
			t.Fatalf("RunRealpath() error = %v", err)
		}

		result := strings.TrimSpace(buf.String())
		if !filepath.IsAbs(result) {
			t.Errorf("RunRealpath() = %v, want absolute path", result)
		}
	})
}

func TestRunReadlink(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "readlink_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("regular file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "regular.txt")
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunReadlink(&buf, []string{file}, ReadlinkOptions{})
		// Regular file is not a symlink, should handle gracefully
		if err != nil {
			t.Logf("RunReadlink() for regular file: %v", err)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunReadlink(&buf, []string{"/nonexistent/path"}, ReadlinkOptions{})
		if err == nil {
			t.Error("RunReadlink() expected error for nonexistent file")
		}
	})
}

func TestBasename(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{filepath.Join("a", "b", "c"), "c"},
		{"file.txt", "file.txt"},
		{"", "."},
	}

	for _, tt := range tests {
		result := Basename(tt.input)
		if result != tt.expect {
			t.Errorf("Basename(%q) = %q, want %q", tt.input, result, tt.expect)
		}
	}
}

func TestDirname(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{filepath.Join("a", "b", "c"), filepath.Join("a", "b")},
		{"file.txt", "."},
	}

	for _, tt := range tests {
		result := Dirname(tt.input)
		if result != tt.expect {
			t.Errorf("Dirname(%q) = %q, want %q", tt.input, result, tt.expect)
		}
	}
}
