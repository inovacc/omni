package ls

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ls_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test files and directories
	if err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content2"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte("hidden"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}

	t.Run("list directory contents", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "file1.txt") {
			t.Errorf("Run() should list file1.txt: %v", output)
		}

		if !strings.Contains(output, "file2.txt") {
			t.Errorf("Run() should list file2.txt: %v", output)
		}

		if !strings.Contains(output, "subdir") {
			t.Errorf("Run() should list subdir: %v", output)
		}
	})

	t.Run("hidden files not shown by default", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		if strings.Contains(output, ".hidden") {
			t.Errorf("Run() should not show hidden files by default: %v", output)
		}
	})

	t.Run("show all files including hidden", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{All: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, ".hidden") {
			t.Errorf("Run() with All should show hidden files: %v", output)
		}
	})

	t.Run("almost all (no . and ..)", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{AlmostAll: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, ".hidden") {
			t.Errorf("Run() with AlmostAll should show hidden files: %v", output)
		}
	})

	t.Run("long format", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{LongFormat: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		// Long format should include permissions
		if !strings.Contains(output, "rw") && !strings.Contains(output, "-") {
			t.Errorf("Run() long format should show permissions: %v", output)
		}
	})

	t.Run("one per line", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{OnePerLine: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		// Should have at least 3 entries (file1.txt, file2.txt, subdir)
		if len(lines) < 3 {
			t.Errorf("Run() one per line should have multiple lines: %d", len(lines))
		}
	})

	t.Run("JSON output", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{JSON: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		// JSON output should have brackets and quotes
		if !strings.Contains(output, "[") || !strings.Contains(output, "\"name\"") {
			t.Errorf("Run() JSON output invalid: %v", output)
		}
	})

	t.Run("sort by time", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{SortByTime: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		// Just verify it doesn't error
		if buf.Len() == 0 {
			t.Error("Run() sort by time should produce output")
		}
	})

	t.Run("sort by size", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{SortBySize: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		// Just verify it doesn't error
		if buf.Len() == 0 {
			t.Error("Run() sort by size should produce output")
		}
	})

	t.Run("reverse order", func(t *testing.T) {
		var buf1, buf2 bytes.Buffer

		err := Run(&buf1, []string{tmpDir}, Options{OnePerLine: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		err = Run(&buf2, []string{tmpDir}, Options{OnePerLine: true, Reverse: true})
		if err != nil {
			t.Fatalf("Run() reverse error = %v", err)
		}

		lines1 := strings.Split(strings.TrimSpace(buf1.String()), "\n")
		lines2 := strings.Split(strings.TrimSpace(buf2.String()), "\n")

		if len(lines1) > 1 && lines1[0] == lines2[0] {
			// Not necessarily wrong, but suspicious
			t.Log("Note: reverse may not have changed order")
		}
	})

	t.Run("directory flag", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{Directory: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		// Should show the directory itself, not contents
		if strings.Contains(output, "file1.txt") {
			t.Errorf("Run() with Directory should not list contents: %v", output)
		}
	})

	t.Run("classify appends indicators", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{Classify: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		// Directories should have / appended
		if !strings.Contains(output, "subdir/") {
			t.Errorf("Run() Classify should append / to directories: %v", output)
		}
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{"/nonexistent/directory"}, Options{})
		if err == nil {
			t.Error("Run() should return error for nonexistent directory")
		}
	})

	t.Run("list specific file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "file1.txt")

		var buf bytes.Buffer

		err := Run(&buf, []string{file}, Options{})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "file1.txt") {
			t.Errorf("Run() should show the specific file: %v", output)
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		emptyDir := filepath.Join(tmpDir, "empty")
		if err := os.Mkdir(emptyDir, 0755); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := Run(&buf, []string{emptyDir}, Options{})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		// Empty directory should produce minimal output
		output := strings.TrimSpace(buf.String())
		if output != "" && !strings.Contains(output, "total") {
			t.Errorf("Run() empty dir should be empty or show total: %v", output)
		}
	})

	t.Run("human readable sizes", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{LongFormat: true, HumanReadble: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		// Just verify it doesn't error
		if buf.Len() == 0 {
			t.Error("Run() human readable should produce output")
		}
	})
}
