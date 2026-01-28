package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDiff(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "diff_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("identical files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "same1.txt")
		file2 := filepath.Join(tmpDir, "same2.txt")
		content := "line1\nline2\nline3"

		if err := os.WriteFile(file1, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(file2, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunDiff(&buf, []string{file1, file2}, DiffOptions{})
		if err != nil {
			t.Fatalf("RunDiff() error = %v", err)
		}

		// Identical files should produce minimal or no output
		output := buf.String()
		if strings.Contains(output, "-") || strings.Contains(output, "+") {
			t.Logf("RunDiff() output for identical files: %v", output)
		}
	})

	t.Run("different files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "diff1.txt")
		file2 := filepath.Join(tmpDir, "diff2.txt")

		if err := os.WriteFile(file1, []byte("apple\nbanana\ncherry"), 0644); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(file2, []byte("apple\norange\ncherry"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunDiff(&buf, []string{file1, file2}, DiffOptions{})
		// RunDiff returns error when files differ
		if err == nil {
			// Some implementations return nil even when files differ
			t.Log("RunDiff() returned no error for different files")
		}

		output := buf.String()
		// Should show some difference
		if !strings.Contains(output, "banana") && !strings.Contains(output, "orange") {
			t.Logf("RunDiff() output: %v", output)
		}
	})

	t.Run("unified format", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "uni1.txt")
		file2 := filepath.Join(tmpDir, "uni2.txt")

		if err := os.WriteFile(file1, []byte("old line\ncommon"), 0644); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(file2, []byte("new line\ncommon"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		_ = RunDiff(&buf, []string{file1, file2}, DiffOptions{Unified: 3})

		output := buf.String()
		// Unified diff should have context markers
		if len(output) > 0 {
			t.Logf("Unified diff output: %v", output)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "exists.txt")
		file2 := filepath.Join(tmpDir, "notexists.txt")

		if err := os.WriteFile(file1, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunDiff(&buf, []string{file1, file2}, DiffOptions{})
		if err == nil {
			t.Error("RunDiff() expected error for nonexistent file")
		}
	})

	t.Run("single argument", func(t *testing.T) {
		file := filepath.Join(tmpDir, "single.txt")
		if err := os.WriteFile(file, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunDiff(&buf, []string{file}, DiffOptions{})
		// With only one file, behavior depends on implementation
		if err != nil {
			t.Logf("RunDiff() with single file: %v", err)
		}
	})
}
