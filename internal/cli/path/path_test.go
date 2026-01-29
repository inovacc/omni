package path

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
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
