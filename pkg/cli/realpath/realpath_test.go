package realpath

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
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

	t.Run("nested directory", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "a", "b", "c")
		if err := os.MkdirAll(dir, 0755); err != nil {
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
			t.Errorf("RunRealpath() nested = %v, want %v or %v", got, expected, expectedReal)
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

		// Both should be absolute
		for i, line := range lines {
			if !filepath.IsAbs(line) {
				t.Errorf("RunRealpath() line %d not absolute: %v", i, line)
			}
		}
	})

	t.Run("three paths", func(t *testing.T) {
		files := []string{
			filepath.Join(tmpDir, "a.txt"),
			filepath.Join(tmpDir, "b.txt"),
			filepath.Join(tmpDir, "c.txt"),
		}

		for _, f := range files {
			_ = os.WriteFile(f, []byte("x"), 0644)
		}

		var buf bytes.Buffer

		_ = RunRealpath(&buf, files)

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 3 {
			t.Errorf("RunRealpath() should return 3 lines, got %d", len(lines))
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

		if !strings.Contains(err.Error(), "realpath") {
			t.Errorf("RunRealpath() error should mention realpath: %v", err)
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

	t.Run("double dot path", func(t *testing.T) {
		subDir := filepath.Join(tmpDir, "subdir")
		_ = os.MkdirAll(subDir, 0755)

		origWd, _ := os.Getwd()

		defer func() { _ = os.Chdir(origWd) }()

		_ = os.Chdir(subDir)

		var buf bytes.Buffer

		err := RunRealpath(&buf, []string{".."})
		if err != nil {
			t.Fatalf("RunRealpath() error = %v", err)
		}

		got := strings.TrimSpace(buf.String())
		expected, _ := filepath.Abs("..")
		expectedReal, _ := filepath.EvalSymlinks(expected)

		if got != expected && got != expectedReal {
			t.Errorf("RunRealpath() .. = %v, want %v or %v", got, expected, expectedReal)
		}
	})

	t.Run("relative path from subdir", func(t *testing.T) {
		file := filepath.Join(tmpDir, "relative_test.txt")
		_ = os.WriteFile(file, []byte("test"), 0644)

		subDir := filepath.Join(tmpDir, "sub")
		_ = os.MkdirAll(subDir, 0755)

		origWd, _ := os.Getwd()

		defer func() { _ = os.Chdir(origWd) }()

		_ = os.Chdir(subDir)

		var buf bytes.Buffer

		err := RunRealpath(&buf, []string{filepath.Join("..", "relative_test.txt")})
		if err != nil {
			t.Fatalf("RunRealpath() error = %v", err)
		}

		got := strings.TrimSpace(buf.String())
		if !filepath.IsAbs(got) {
			t.Errorf("RunRealpath() relative should return absolute: %v", got)
		}
	})

	t.Run("symlink resolution", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Symlinks may require admin on Windows")
		}

		realFile := filepath.Join(tmpDir, "real.txt")
		linkFile := filepath.Join(tmpDir, "link.txt")

		_ = os.WriteFile(realFile, []byte("real"), 0644)

		if err := os.Symlink(realFile, linkFile); err != nil {
			t.Skip("Cannot create symlink:", err)
		}

		var buf bytes.Buffer

		err := RunRealpath(&buf, []string{linkFile})
		if err != nil {
			t.Fatalf("RunRealpath() error = %v", err)
		}

		got := strings.TrimSpace(buf.String())
		expected, _ := filepath.EvalSymlinks(realFile)

		if got != expected {
			t.Errorf("RunRealpath() symlink = %v, want %v", got, expected)
		}
	})

	t.Run("output ends with newline", func(t *testing.T) {
		file := filepath.Join(tmpDir, "newline_test.txt")
		_ = os.WriteFile(file, []byte("test"), 0644)

		var buf bytes.Buffer

		_ = RunRealpath(&buf, []string{file})

		if !strings.HasSuffix(buf.String(), "\n") {
			t.Error("RunRealpath() output should end with newline")
		}
	})

	t.Run("consistent results", func(t *testing.T) {
		file := filepath.Join(tmpDir, "consistent.txt")
		_ = os.WriteFile(file, []byte("test"), 0644)

		var buf1, buf2 bytes.Buffer

		_ = RunRealpath(&buf1, []string{file})
		_ = RunRealpath(&buf2, []string{file})

		if buf1.String() != buf2.String() {
			t.Errorf("RunRealpath() should be consistent: %v vs %v", buf1.String(), buf2.String())
		}
	})

	t.Run("file with spaces in name", func(t *testing.T) {
		file := filepath.Join(tmpDir, "file with spaces.txt")
		_ = os.WriteFile(file, []byte("test"), 0644)

		var buf bytes.Buffer

		err := RunRealpath(&buf, []string{file})
		if err != nil {
			t.Fatalf("RunRealpath() error = %v", err)
		}

		got := strings.TrimSpace(buf.String())
		if !strings.Contains(got, "file with spaces") {
			t.Errorf("RunRealpath() should preserve spaces: %v", got)
		}
	})

	t.Run("hidden file", func(t *testing.T) {
		file := filepath.Join(tmpDir, ".hidden")
		_ = os.WriteFile(file, []byte("hidden"), 0644)

		var buf bytes.Buffer

		err := RunRealpath(&buf, []string{file})
		if err != nil {
			t.Fatalf("RunRealpath() error = %v", err)
		}

		got := strings.TrimSpace(buf.String())
		if !strings.Contains(got, ".hidden") {
			t.Errorf("RunRealpath() should handle hidden files: %v", got)
		}
	})

	t.Run("unicode filename", func(t *testing.T) {
		file := filepath.Join(tmpDir, "文件.txt")
		_ = os.WriteFile(file, []byte("unicode"), 0644)

		var buf bytes.Buffer

		err := RunRealpath(&buf, []string{file})
		if err != nil {
			t.Fatalf("RunRealpath() error = %v", err)
		}

		got := strings.TrimSpace(buf.String())
		if !strings.Contains(got, "文件") {
			t.Errorf("RunRealpath() should handle unicode: %v", got)
		}
	})
}
