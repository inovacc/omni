package readlink

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunReadlink(t *testing.T) {
	// Skip symlink tests on Windows if not running as admin
	if runtime.GOOS == "windows" {
		t.Skip("Skipping symlink tests on Windows")
	}

	tmpDir, err := os.MkdirTemp("", "readlink_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a test file and symlink
	file := filepath.Join(tmpDir, "target.txt")
	_ = os.WriteFile(file, []byte("content"), 0644)

	link := filepath.Join(tmpDir, "link")
	_ = os.Symlink(file, link)

	t.Run("read symlink", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunReadlink(&buf, []string{link}, ReadlinkOptions{})
		if err != nil {
			t.Fatalf("RunReadlink() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output != file {
			t.Errorf("RunReadlink() = %q, want %q", output, file)
		}
	})

	t.Run("canonicalize", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunReadlink(&buf, []string{link}, ReadlinkOptions{Canonicalize: true})
		if err != nil {
			t.Fatalf("RunReadlink() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if !filepath.IsAbs(output) {
			t.Errorf("RunReadlink() -f should return absolute path: %q", output)
		}
	})

	t.Run("canonicalize existing", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunReadlink(&buf, []string{file}, ReadlinkOptions{CanonicalizeExisting: true})
		if err != nil {
			t.Fatalf("RunReadlink() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if !filepath.IsAbs(output) {
			t.Errorf("RunReadlink() -e should return absolute path: %q", output)
		}
	})

	t.Run("canonicalize missing", func(t *testing.T) {
		var buf bytes.Buffer

		nonexistent := filepath.Join(tmpDir, "nonexistent")
		err := RunReadlink(&buf, []string{nonexistent}, ReadlinkOptions{CanonicalizeMissing: true})
		if err != nil {
			t.Fatalf("RunReadlink() error = %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if !filepath.IsAbs(output) {
			t.Errorf("RunReadlink() -m should return absolute path: %q", output)
		}
	})

	t.Run("no newline", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunReadlink(&buf, []string{link}, ReadlinkOptions{NoNewline: true})
		if err != nil {
			t.Fatalf("RunReadlink() error = %v", err)
		}

		output := buf.String()
		if strings.HasSuffix(output, "\n") {
			t.Errorf("RunReadlink() -n should not have trailing newline")
		}
	})

	t.Run("zero terminated", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunReadlink(&buf, []string{link}, ReadlinkOptions{Zero: true})
		if err != nil {
			t.Fatalf("RunReadlink() error = %v", err)
		}

		if !strings.HasSuffix(buf.String(), "\x00") {
			t.Errorf("RunReadlink() -z should have null terminator")
		}
	})

	t.Run("missing operand", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunReadlink(&buf, []string{}, ReadlinkOptions{})
		if err == nil {
			t.Error("RunReadlink() expected error for missing operand")
		}
	})

	t.Run("not a symlink", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunReadlink(&buf, []string{file}, ReadlinkOptions{})
		if err == nil {
			t.Error("RunReadlink() expected error for non-symlink")
		}
	})

	t.Run("quiet mode", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunReadlink(&buf, []string{file}, ReadlinkOptions{Quiet: true})
		// Should still return error but not print it
		if err == nil {
			t.Error("RunReadlink() expected error even with quiet mode")
		}
	})
}

func TestCanonicalPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "canonical_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	file := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(file, []byte("content"), 0644)

	result, err := CanonicalPath(file)
	if err != nil {
		t.Fatalf("CanonicalPath() error = %v", err)
	}

	if !filepath.IsAbs(result) {
		t.Errorf("CanonicalPath() = %q, want absolute path", result)
	}
}

func TestReadlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping symlink tests on Windows")
	}

	tmpDir, err := os.MkdirTemp("", "readlink_helper_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	target := filepath.Join(tmpDir, "target")
	_ = os.WriteFile(target, []byte("content"), 0644)

	link := filepath.Join(tmpDir, "link")
	_ = os.Symlink(target, link)

	result, err := Readlink(link)
	if err != nil {
		t.Fatalf("Readlink() error = %v", err)
	}

	if result != target {
		t.Errorf("Readlink() = %q, want %q", result, target)
	}
}
