package tests

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/inovacc/omni/pkg/cli"
)

//nolint:maintidx // Test function has expected high complexity with many subtest cases
func TestPhase1Commands(t *testing.T) {
	// Setup: Create a temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "omni_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	origWd, _ := os.Getwd()

	defer func() {
		_ = os.Chdir(origWd)
	}()

	_ = os.Chdir(tmpDir)

	// 1. Test pwd
	t.Run("pwd", func(t *testing.T) {
		var buf bytes.Buffer

		err := cli.RunPwd(&buf)
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}

		expected, _ := filepath.Abs(".")
		if strings.TrimSpace(buf.String()) != expected {
			t.Errorf("expected %s, got %s", expected, buf.String())
		}
	})

	// 2. Test date
	t.Run("date", func(t *testing.T) {
		var buf bytes.Buffer

		err := cli.RunDate(&buf, cli.DateOptions{})
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}

		_, err = time.Parse(time.RFC3339, strings.TrimSpace(buf.String()))
		if err != nil {
			t.Errorf("expected RFC3339 date, got %s: %v", buf.String(), err)
		}
	})

	// 3. Test ls
	t.Run("ls", func(t *testing.T) {
		// Create a file
		fname := "testfile.txt"
		if err := os.WriteFile(fname, []byte("hello"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := cli.RunLs(&buf, []string{"."}, cli.LsOptions{})
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}

		if !strings.Contains(buf.String(), fname) {
			t.Errorf("expected output to contain %s, got %s", fname, buf.String())
		}

		// Test JSON mode
		var jsonBuf bytes.Buffer

		err = cli.RunLs(&jsonBuf, []string{"."}, cli.LsOptions{JSON: true})
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}

		if !strings.Contains(jsonBuf.String(), `"`+fname+`"`) {
			t.Errorf("expected JSON output to contain %s, got %s", `"`+fname+`"`, jsonBuf.String())
		}
	})

	// 4. Test cat
	t.Run("cat", func(t *testing.T) {
		fname := "cat-test.txt"

		content := "cat content"
		if err := os.WriteFile(fname, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := cli.RunCat(&buf, []string{fname}, cli.CatOptions{})
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}

		if strings.TrimSpace(buf.String()) != content {
			t.Errorf("expected %s, got %s", content, buf.String())
		}
	})

	// 5. Test dirname
	t.Run("dirname", func(t *testing.T) {
		path := filepath.Join("a", "b", "c")

		var buf bytes.Buffer

		err := cli.RunDirname(&buf, []string{path})
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}

		expected := filepath.Dir(path)
		if strings.TrimSpace(buf.String()) != expected {
			t.Errorf("expected %s, got %s", expected, buf.String())
		}
	})

	// 6. Test basename
	t.Run("basename", func(t *testing.T) {
		path := filepath.Join("a", "b", "c.txt")

		var buf bytes.Buffer

		err := cli.RunBasename(&buf, []string{path}, ".txt")
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}

		if strings.TrimSpace(buf.String()) != "c" {
			t.Errorf("expected c, got %s", buf.String())
		}
	})

	// 7. Test realpath
	t.Run("realpath", func(t *testing.T) {
		fname := "real-test.txt"
		if err := os.WriteFile(fname, []byte("data"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := cli.RunRealpath(&buf, []string{fname})
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}

		expected, _ := filepath.Abs(fname)
		// filepath.EvalSymlinks might be needed if tmpDir has symlinks,
		// but let's see if simple Abs matches.
		if strings.TrimSpace(buf.String()) != expected {
			// On some systems (like macOS), /tmp is a symlink to /private/tmp
			realExpected, _ := filepath.EvalSymlinks(expected)
			if strings.TrimSpace(buf.String()) != realExpected {
				t.Errorf("expected %s or %s, got %s", expected, realExpected, buf.String())
			}
		}
	})

	// 8. Test grep
	t.Run("grep", func(t *testing.T) {
		fname := "grep-test.txt"

		content := "hello world\nfoo bar\nhello again"
		if err := os.WriteFile(fname, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := cli.RunGrep(&buf, "hello", []string{fname}, cli.GrepOptions{})
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Errorf("expected 2 matching lines, got %d: %s", len(lines), buf.String())
		}
	})

	// 9. Test head
	t.Run("head", func(t *testing.T) {
		fname := "head-test.txt"

		content := "line1\nline2\nline3\nline4\nline5"
		if err := os.WriteFile(fname, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := cli.RunHead(&buf, []string{fname}, cli.HeadOptions{Lines: 2})
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Errorf("expected 2 lines, got %d: %s", len(lines), buf.String())
		}
	})

	// 10. Test tail
	t.Run("tail", func(t *testing.T) {
		fname := "tail-test.txt"

		content := "line1\nline2\nline3\nline4\nline5"
		if err := os.WriteFile(fname, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := cli.RunTail(&buf, []string{fname}, cli.TailOptions{Lines: 2})
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) != 2 {
			t.Errorf("expected 2 lines, got %d: %s", len(lines), buf.String())
		}

		if lines[0] != "line4" || lines[1] != "line5" {
			t.Errorf("expected line4 and line5, got %v", lines)
		}
	})

	// 11. Test wc
	t.Run("wc", func(t *testing.T) {
		fname := "wc-test.txt"
		// Linux wc -l counts newline characters, so "hello\nworld\n" = 2 newlines
		content := "hello world\nfoo bar\n"
		if err := os.WriteFile(fname, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := cli.RunWC(&buf, []string{fname}, cli.WCOptions{Lines: true})
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}

		if !strings.Contains(buf.String(), "2") {
			t.Errorf("expected 2 lines, got %s", buf.String())
		}
	})

	// 12. Test sort
	t.Run("sort", func(t *testing.T) {
		fname := "sort-test.txt"

		content := "banana\napple\ncherry"
		if err := os.WriteFile(fname, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := cli.RunSort(&buf, []string{fname}, cli.SortOptions{})
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if lines[0] != "apple" || lines[1] != "banana" || lines[2] != "cherry" {
			t.Errorf("expected sorted order, got %v", lines)
		}
	})

	// 13. Test whoami
	t.Run("whoami", func(t *testing.T) {
		var buf bytes.Buffer

		err := cli.RunWhoami(&buf)
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		if output == "" {
			t.Error("expected username, got empty string")
		}
	})

	// 14. Test env
	t.Run("env", func(t *testing.T) {
		// Set a test variable
		if err := os.Setenv("omni_TEST_VAR", "test_value"); err != nil {
			t.Fatal(err)
		}

		defer func() {
			_ = os.Unsetenv("omni_TEST_VAR")
		}()

		var buf bytes.Buffer

		err := cli.RunEnv(&buf, []string{"omni_TEST_VAR"}, cli.EnvOptions{})
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}

		if !strings.Contains(buf.String(), "omni_TEST_VAR=test_value") {
			t.Errorf("expected env var, got %s", buf.String())
		}
	})

	// 15. Test uname
	t.Run("uname", func(t *testing.T) {
		var buf bytes.Buffer

		err := cli.RunUname(&buf, cli.UnameOptions{})
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}

		output := strings.TrimSpace(buf.String())
		// Default should be kernel name (Linux, Darwin, Windows_NT, etc.)
		if output == "" {
			t.Error("expected kernel name, got empty string")
		}
	})

	// 16. Test uname -a
	t.Run("uname-all", func(t *testing.T) {
		var buf bytes.Buffer

		err := cli.RunUname(&buf, cli.UnameOptions{All: true})
		if err != nil {
			t.Fatalf("command failed: %v", err)
		}

		output := buf.String()
		// Should contain multiple space-separated fields
		fields := strings.Fields(output)
		if len(fields) < 5 {
			t.Errorf("expected multiple fields in uname -a, got %d: %s", len(fields), output)
		}
	})
}
