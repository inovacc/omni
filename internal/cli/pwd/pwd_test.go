package pwd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPwd(t *testing.T) {
	t.Run("returns current directory", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunPwd(&buf)
		if err != nil {
			t.Fatalf("RunPwd() error = %v", err)
		}

		expected, _ := os.Getwd()
		got := strings.TrimSpace(buf.String())

		if got != expected {
			t.Errorf("RunPwd() = %v, want %v", got, expected)
		}
	})

	t.Run("returns absolute path", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunPwd(&buf)
		if err != nil {
			t.Fatalf("RunPwd() error = %v", err)
		}

		got := strings.TrimSpace(buf.String())

		if !filepath.IsAbs(got) {
			t.Errorf("RunPwd() returned non-absolute path: %v", got)
		}
	})

	t.Run("after chdir", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "pwd_test")
		if err != nil {
			t.Fatal(err)
		}

		defer func() { _ = os.RemoveAll(tmpDir) }()

		origWd, _ := os.Getwd()

		defer func() { _ = os.Chdir(origWd) }()

		_ = os.Chdir(tmpDir)

		var buf bytes.Buffer

		err = RunPwd(&buf)
		if err != nil {
			t.Fatalf("RunPwd() error = %v", err)
		}

		got := strings.TrimSpace(buf.String())
		expected, _ := filepath.Abs(tmpDir)

		// Handle symlinks (e.g., macOS /tmp -> /private/tmp)
		expectedReal, _ := filepath.EvalSymlinks(expected)
		gotReal, _ := filepath.EvalSymlinks(got)

		if gotReal != expectedReal {
			t.Errorf("RunPwd() = %v, want %v", got, expected)
		}
	})

	t.Run("output ends with newline", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunPwd(&buf)
		if err != nil {
			t.Fatalf("RunPwd() error = %v", err)
		}

		output := buf.String()
		if !strings.HasSuffix(output, "\n") {
			t.Error("RunPwd() output should end with newline")
		}
	})

	t.Run("consistent results", func(t *testing.T) {
		var buf1, buf2 bytes.Buffer

		_ = RunPwd(&buf1)
		_ = RunPwd(&buf2)

		if buf1.String() != buf2.String() {
			t.Errorf("RunPwd() should be consistent: %v vs %v", buf1.String(), buf2.String())
		}
	})

	t.Run("single line output", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunPwd(&buf)

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 1 {
			t.Errorf("RunPwd() should output exactly one line, got %d", len(lines))
		}
	})

	t.Run("no extra whitespace", func(t *testing.T) {
		var buf bytes.Buffer

		_ = RunPwd(&buf)

		output := buf.String()
		trimmed := strings.TrimRight(output, "\n")

		if strings.HasPrefix(trimmed, " ") || strings.HasSuffix(trimmed, " ") {
			t.Error("RunPwd() should not have extra whitespace")
		}
	})

	t.Run("matches os.Getwd", func(t *testing.T) {
		expected, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunPwd(&buf)

		got := strings.TrimSpace(buf.String())
		if got != expected {
			t.Errorf("RunPwd() = %v, want %v", got, expected)
		}
	})

	t.Run("after multiple chdirs", func(t *testing.T) {
		tmpDir1, _ := os.MkdirTemp("", "pwd_test1")
		tmpDir2, _ := os.MkdirTemp("", "pwd_test2")

		defer func() {
			_ = os.RemoveAll(tmpDir1)
			_ = os.RemoveAll(tmpDir2)
		}()

		origWd, _ := os.Getwd()

		defer func() { _ = os.Chdir(origWd) }()

		// First chdir
		_ = os.Chdir(tmpDir1)

		var buf1 bytes.Buffer

		_ = RunPwd(&buf1)

		// Second chdir
		_ = os.Chdir(tmpDir2)

		var buf2 bytes.Buffer

		_ = RunPwd(&buf2)

		got1 := strings.TrimSpace(buf1.String())
		got2 := strings.TrimSpace(buf2.String())

		// Results should be different
		real1, _ := filepath.EvalSymlinks(tmpDir1)
		real2, _ := filepath.EvalSymlinks(tmpDir2)
		gotReal1, _ := filepath.EvalSymlinks(got1)
		gotReal2, _ := filepath.EvalSymlinks(got2)

		if gotReal1 != real1 {
			t.Errorf("RunPwd() after first chdir = %v, want %v", got1, tmpDir1)
		}

		if gotReal2 != real2 {
			t.Errorf("RunPwd() after second chdir = %v, want %v", got2, tmpDir2)
		}
	})
}

func TestPwd(t *testing.T) {
	t.Run("returns current directory", func(t *testing.T) {
		expected, _ := os.Getwd()

		got, err := Pwd()
		if err != nil {
			t.Fatalf("Pwd() error = %v", err)
		}

		if got != expected {
			t.Errorf("Pwd() = %v, want %v", got, expected)
		}
	})

	t.Run("returns absolute path", func(t *testing.T) {
		got, err := Pwd()
		if err != nil {
			t.Fatalf("Pwd() error = %v", err)
		}

		if !filepath.IsAbs(got) {
			t.Errorf("Pwd() should return absolute path: %v", got)
		}
	})

	t.Run("consistent results", func(t *testing.T) {
		result1, _ := Pwd()
		result2, _ := Pwd()

		if result1 != result2 {
			t.Errorf("Pwd() should be consistent: %v vs %v", result1, result2)
		}
	})

	t.Run("no trailing newline", func(t *testing.T) {
		got, _ := Pwd()

		if strings.HasSuffix(got, "\n") {
			t.Error("Pwd() should not have trailing newline")
		}
	})

	t.Run("after chdir", func(t *testing.T) {
		tmpDir, _ := os.MkdirTemp("", "pwd_func_test")

		defer func() { _ = os.RemoveAll(tmpDir) }()

		origWd, _ := os.Getwd()

		defer func() { _ = os.Chdir(origWd) }()

		_ = os.Chdir(tmpDir)

		got, err := Pwd()
		if err != nil {
			t.Fatalf("Pwd() error = %v", err)
		}

		realTmpDir, _ := filepath.EvalSymlinks(tmpDir)
		realGot, _ := filepath.EvalSymlinks(got)

		if realGot != realTmpDir {
			t.Errorf("Pwd() after chdir = %v, want %v", got, tmpDir)
		}
	})

	t.Run("no error normally", func(t *testing.T) {
		_, err := Pwd()
		if err != nil {
			t.Errorf("Pwd() should not error in normal conditions: %v", err)
		}
	})
}
