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
}
