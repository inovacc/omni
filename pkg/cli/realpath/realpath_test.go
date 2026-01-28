package realpath

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

		got := strings.TrimSpace(buf.String())
		expected, _ := filepath.Abs(file)
		expectedReal, _ := filepath.EvalSymlinks(expected)

		if got != expected && got != expectedReal {
			t.Errorf("RunRealpath() = %v, want %v or %v", got, expected, expectedReal)
		}
	})

	t.Run("returns absolute path", func(t *testing.T) {
		file := filepath.Join(tmpDir, "abs_test.txt")
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunRealpath(&buf, []string{file})
		if err != nil {
			t.Fatalf("RunRealpath() error = %v", err)
		}

		got := strings.TrimSpace(buf.String())
		if !filepath.IsAbs(got) {
			t.Errorf("RunRealpath() should return absolute path: %v", got)
		}
	})

	t.Run("directory path", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "subdir")
		if err := os.Mkdir(dir, 0755); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunRealpath(&buf, []string{dir})
		if err != nil {
			t.Fatalf("RunRealpath() error = %v", err)
		}

		got := strings.TrimSpace(buf.String())
		expected, _ := filepath.Abs(dir)
		expectedReal, _ := filepath.EvalSymlinks(expected)

		if got != expected && got != expectedReal {
			t.Errorf("RunRealpath() = %v, want %v or %v", got, expected, expectedReal)
		}
	})

	t.Run("multiple paths", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "file1.txt")
		file2 := filepath.Join(tmpDir, "file2.txt")

		if err := os.WriteFile(file1, []byte("1"), 0644); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(file2, []byte("2"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunRealpath(&buf, []string{file1, file2})
		if err != nil {
			t.Fatalf("RunRealpath() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Errorf("RunRealpath() should return 2 lines, got %d", len(lines))
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunRealpath(&buf, []string{filepath.Join(tmpDir, "nonexistent.txt")})
		if err == nil {
			t.Error("RunRealpath() should return error for nonexistent file")
		}
	})

	t.Run("no arguments", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunRealpath(&buf, []string{})
		if err == nil {
			t.Error("RunRealpath() should return error with no arguments")
		}
	})

	t.Run("dot path", func(t *testing.T) {
		origWd, _ := os.Getwd()

		defer func() { _ = os.Chdir(origWd) }()

		_ = os.Chdir(tmpDir)

		var buf bytes.Buffer

		err := RunRealpath(&buf, []string{"."})
		if err != nil {
			t.Fatalf("RunRealpath() error = %v", err)
		}

		got := strings.TrimSpace(buf.String())
		expected, _ := filepath.Abs(".")
		expectedReal, _ := filepath.EvalSymlinks(expected)

		if got != expected && got != expectedReal {
			t.Errorf("RunRealpath() = %v, want %v or %v", got, expected, expectedReal)
		}
	})
}
