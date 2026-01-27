package tests

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/inovacc/goshell/pkg/cli"
)

func TestPhase1Commands(t *testing.T) {
	// Setup: Create a temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "goshell_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	origWd, _ := os.Getwd()
	defer os.Chdir(origWd)
	os.Chdir(tmpDir)

	// 1. Test pwd
	t.Run("pwd", func(t *testing.T) {
		out := captureStdout(t, func() error {
			return cli.RunPwd()
		})
		expected, _ := filepath.Abs(".")
		if strings.TrimSpace(out) != expected {
			t.Errorf("expected %s, got %s", expected, out)
		}
	})

	// 2. Test date
	t.Run("date", func(t *testing.T) {
		out := captureStdout(t, func() error {
			return cli.RunDate()
		})
		_, err := time.Parse(time.RFC3339, strings.TrimSpace(out))
		if err != nil {
			t.Errorf("expected RFC3339 date, got %s: %v", out, err)
		}
	})

	// 3. Test ls
	t.Run("ls", func(t *testing.T) {
		// Create a file
		fname := "testfile.txt"
		if err := os.WriteFile(fname, []byte("hello"), 0644); err != nil {
			t.Fatal(err)
		}
		out := captureStdout(t, func() error {
			return cli.RunLs([]string{"."}, false)
		})
		if !strings.Contains(out, fname) {
			t.Errorf("expected output to contain %s, got %s", fname, out)
		}

		// Test JSON mode
		outJson := captureStdout(t, func() error {
			return cli.RunLs([]string{"."}, true)
		})
		if !strings.Contains(outJson, `"`+fname+`"`) {
			t.Errorf("expected JSON output to contain %s, got %s", `"`+fname+`"`, outJson)
		}
	})

	// 4. Test cat
	t.Run("cat", func(t *testing.T) {
		fname := "cat-test.txt"
		content := "cat content"
		if err := os.WriteFile(fname, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		out := captureStdout(t, func() error {
			return cli.RunCat([]string{fname})
		})
		if strings.TrimSpace(out) != content {
			t.Errorf("expected %s, got %s", content, out)
		}
	})

	// 5. Test dirname
	t.Run("dirname", func(t *testing.T) {
		path := filepath.Join("a", "b", "c")
		out := captureStdout(t, func() error {
			return cli.RunDirname([]string{path})
		})
		expected := filepath.Dir(path)
		if strings.TrimSpace(out) != expected {
			t.Errorf("expected %s, got %s", expected, out)
		}
	})

	// 6. Test basename
	t.Run("basename", func(t *testing.T) {
		path := filepath.Join("a", "b", "c.txt")
		out := captureStdout(t, func() error {
			return cli.RunBasename([]string{path, ".txt"})
		})
		if strings.TrimSpace(out) != "c" {
			t.Errorf("expected c, got %s", out)
		}
	})

	// 7. Test realpath
	t.Run("realpath", func(t *testing.T) {
		fname := "real-test.txt"
		if err := os.WriteFile(fname, []byte("data"), 0644); err != nil {
			t.Fatal(err)
		}
		out := captureStdout(t, func() error {
			return cli.RunRealpath([]string{fname})
		})
		expected, _ := filepath.Abs(fname)
		// filepath.EvalSymlinks might be needed if tmpDir has symlinks,
		// but let's see if simple Abs matches.
		if strings.TrimSpace(out) != expected {
			// On some systems (like macOS), /tmp is a symlink to /private/tmp
			realExpected, _ := filepath.EvalSymlinks(expected)
			if strings.TrimSpace(out) != realExpected {
				t.Errorf("expected %s or %s, got %s", expected, realExpected, out)
			}
		}
	})
}

func captureStdout(t *testing.T, f func() error) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := f()

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}
