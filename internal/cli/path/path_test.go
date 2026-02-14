package path

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/output"
)

func TestRealpath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "realpath_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("absolute path", func(t *testing.T) {
		result, err := Realpath(tmpDir)
		if err != nil {
			t.Fatalf("Realpath() error = %v", err)
		}

		if !filepath.IsAbs(result) {
			t.Errorf("Realpath() = %q, want absolute path", result)
		}
	})

	t.Run("relative path", func(t *testing.T) {
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)

		defer func() { _ = os.Chdir(origDir) }()

		result, err := Realpath(".")
		if err != nil {
			t.Fatalf("Realpath() error = %v", err)
		}

		if !filepath.IsAbs(result) {
			t.Errorf("Realpath('.') = %q, want absolute path", result)
		}
	})

	t.Run("symlink", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping symlink test on Windows")
		}

		target := filepath.Join(tmpDir, "target")
		link := filepath.Join(tmpDir, "link")
		_ = os.WriteFile(target, []byte("content"), 0644)
		_ = os.Symlink(target, link)

		result, err := Realpath(link)
		if err != nil {
			t.Fatalf("Realpath() error = %v", err)
		}

		if result != target {
			t.Errorf("Realpath(symlink) = %q, want %q", result, target)
		}
	})

	t.Run("nonexistent", func(t *testing.T) {
		_, err := Realpath("/nonexistent/path/12345")
		if err == nil {
			t.Error("Realpath() expected error for nonexistent path")
		}
	})
}

func TestDirname(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/foo/bar/baz.txt", "/foo/bar"},
		{"/foo/bar/", "/foo/bar"},
		{"/foo", "/"},
		{"foo/bar", "foo"},
		{"foo", "."},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			// Normalize expected for the platform
			expected := filepath.FromSlash(tt.expected)

			got := Dirname(filepath.FromSlash(tt.path))
			if got != expected {
				t.Errorf("Dirname(%q) = %q, want %q", tt.path, got, expected)
			}
		})
	}
}

func TestBasename(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/foo/bar/baz.txt", "baz.txt"},
		{"/foo/bar/", "bar"},
		{"/foo", "foo"},
		{"foo/bar", "bar"},
		{"foo", "foo"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := Basename(filepath.FromSlash(tt.path))
			if got != tt.expected {
				t.Errorf("Basename(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		paths    []string
		expected string
	}{
		{[]string{"foo", "bar"}, "foo/bar"},
		{[]string{"/foo", "bar"}, "/foo/bar"},
		{[]string{"foo", "bar", "baz"}, "foo/bar/baz"},
		{[]string{"foo", "", "bar"}, "foo/bar"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			expected := filepath.FromSlash(tt.expected)

			got := Join(tt.paths...)
			if got != expected {
				t.Errorf("Join(%v) = %q, want %q", tt.paths, got, expected)
			}
		})
	}
}

func TestClean(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"foo/bar/../baz", "foo/baz"},
		{"foo//bar", "foo/bar"},
		{"./foo", "foo"},
		{"foo/./bar", "foo/bar"},
		{"/foo/bar/./baz/../qux", "/foo/bar/qux"},
		{".", "."},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			expected := filepath.FromSlash(tt.expected)

			got := Clean(filepath.FromSlash(tt.path))
			if got != expected {
				t.Errorf("Clean(%q) = %q, want %q", tt.path, got, expected)
			}
		})
	}
}

func TestAbs(t *testing.T) {
	t.Run("relative path", func(t *testing.T) {
		result, err := Abs(".")
		if err != nil {
			t.Fatalf("Abs() error = %v", err)
		}

		if !filepath.IsAbs(result) {
			t.Errorf("Abs('.') = %q, want absolute path", result)
		}
	})

	t.Run("relative subdir", func(t *testing.T) {
		cwd, _ := os.Getwd()

		result, err := Abs("./test")
		if err != nil {
			t.Fatalf("Abs() error = %v", err)
		}

		expected := filepath.Join(cwd, "test")
		if result != expected {
			t.Errorf("Abs('./test') = %q, want %q", result, expected)
		}
	})

	t.Run("absolute path unchanged", func(t *testing.T) {
		var input string
		if runtime.GOOS == "windows" {
			input = `C:\Users\test`
		} else {
			input = "/usr/local/bin"
		}

		result, err := Abs(input)
		if err != nil {
			t.Fatalf("Abs() error = %v", err)
		}

		if result != input {
			t.Errorf("Abs(%q) = %q, want %q", input, result, input)
		}
	})
}

func TestRunClean(t *testing.T) {
	t.Run("single path", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunClean(&buf, []string{"foo/bar/../baz"}, CleanOptions{})
		if err != nil {
			t.Fatal(err)
		}

		got := strings.TrimSpace(buf.String())

		expected := filepath.FromSlash("foo/baz")
		if got != expected {
			t.Errorf("RunClean() = %q, want %q", got, expected)
		}
	})

	t.Run("multiple paths", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunClean(&buf, []string{"./foo", "bar//baz"}, CleanOptions{})
		if err != nil {
			t.Fatal(err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Fatalf("expected 2 lines, got %d", len(lines))
		}

		if lines[0] != "foo" {
			t.Errorf("line 0 = %q, want %q", lines[0], "foo")
		}

		expected := filepath.FromSlash("bar/baz")
		if lines[1] != expected {
			t.Errorf("line 1 = %q, want %q", lines[1], expected)
		}
	})

	t.Run("json output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunClean(&buf, []string{"foo//bar"}, CleanOptions{OutputFormat: output.FormatJSON})
		if err != nil {
			t.Fatal(err)
		}

		var results []CleanResult
		if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
			t.Fatalf("JSON decode error: %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}

		if results[0].Original != "foo//bar" {
			t.Errorf("original = %q, want %q", results[0].Original, "foo//bar")
		}

		expected := filepath.FromSlash("foo/bar")
		if results[0].Cleaned != expected {
			t.Errorf("cleaned = %q, want %q", results[0].Cleaned, expected)
		}
	})

	t.Run("missing operand", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunClean(&buf, nil, CleanOptions{})
		if err == nil {
			t.Error("expected error for missing operand")
		}
	})
}

func TestRunAbs(t *testing.T) {
	t.Run("relative path", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunAbs(&buf, []string{"./test"}, AbsOptions{})
		if err != nil {
			t.Fatal(err)
		}

		got := strings.TrimSpace(buf.String())
		if !filepath.IsAbs(got) {
			t.Errorf("RunAbs() = %q, want absolute path", got)
		}

		cwd, _ := os.Getwd()

		expected := filepath.Join(cwd, "test")
		if got != expected {
			t.Errorf("RunAbs() = %q, want %q", got, expected)
		}
	})

	t.Run("dot resolves to cwd", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunAbs(&buf, []string{"."}, AbsOptions{})
		if err != nil {
			t.Fatal(err)
		}

		got := strings.TrimSpace(buf.String())

		cwd, _ := os.Getwd()
		if got != cwd {
			t.Errorf("RunAbs('.') = %q, want %q", got, cwd)
		}
	})

	t.Run("json output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunAbs(&buf, []string{"./test"}, AbsOptions{OutputFormat: output.FormatJSON})
		if err != nil {
			t.Fatal(err)
		}

		var results []AbsResult
		if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
			t.Fatalf("JSON decode error: %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("expected 1 result, got %d", len(results))
		}

		if results[0].Original != "./test" {
			t.Errorf("original = %q, want %q", results[0].Original, "./test")
		}

		if !filepath.IsAbs(results[0].Absolute) {
			t.Errorf("absolute = %q, want absolute path", results[0].Absolute)
		}
	})

	t.Run("missing operand", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunAbs(&buf, nil, AbsOptions{})
		if err == nil {
			t.Error("expected error for missing operand")
		}
	})
}
