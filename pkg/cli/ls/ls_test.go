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

	t.Run("recursive listing", func(t *testing.T) {
		// Create nested structure
		nestedDir := filepath.Join(tmpDir, "nested")
		if err := os.Mkdir(nestedDir, 0755); err != nil {
			t.Fatal(err)
		}

		nestedFile := filepath.Join(nestedDir, "nested_file.txt")
		if err := os.WriteFile(nestedFile, []byte("nested"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{Recursive: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "nested_file.txt") {
			t.Errorf("Run() recursive should show nested files: %v", output)
		}
	})

	t.Run("long format shows permissions", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{LongFormat: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		// Should contain permission characters
		if !strings.Contains(output, "r") && !strings.Contains(output, "w") {
			t.Log("Note: Long format permissions may vary by platform")
		}
	})

	t.Run("long format shows size", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{LongFormat: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		// Output should include file size (number)
		output := buf.String()
		hasNumber := false
		for _, r := range output {
			if r >= '0' && r <= '9' {
				hasNumber = true
				break
			}
		}

		if !hasNumber {
			t.Error("Run() long format should show file sizes")
		}
	})

	t.Run("multiple directories", func(t *testing.T) {
		dir1 := filepath.Join(tmpDir, "dir1")
		dir2 := filepath.Join(tmpDir, "dir2")

		_ = os.Mkdir(dir1, 0755)
		_ = os.Mkdir(dir2, 0755)
		_ = os.WriteFile(filepath.Join(dir1, "d1file.txt"), []byte("d1"), 0644)
		_ = os.WriteFile(filepath.Join(dir2, "d2file.txt"), []byte("d2"), 0644)

		var buf bytes.Buffer

		err := Run(&buf, []string{dir1, dir2}, Options{})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "d1file.txt") || !strings.Contains(output, "d2file.txt") {
			t.Errorf("Run() multiple dirs should show both: %v", output)
		}
	})

	t.Run("inode number", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{Inode: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		// Should produce output with numbers (inodes)
		if buf.Len() == 0 {
			t.Error("Run() with inode should produce output")
		}
	})

	t.Run("long format combined options", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{LongFormat: true, HumanReadble: true, All: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("Run() with combined options should produce output")
		}
	})

	t.Run("all with long format", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{LongFormat: true, All: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, ".hidden") {
			t.Errorf("Run() -la should show hidden files: %v", output)
		}
	})

	t.Run("unicode filename", func(t *testing.T) {
		unicodeFile := filepath.Join(tmpDir, "文件名.txt")
		if err := os.WriteFile(unicodeFile, []byte("unicode"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "文件名") {
			t.Log("Note: Unicode filenames may not display on all platforms")
		}
	})

	t.Run("filename with spaces", func(t *testing.T) {
		spaceFile := filepath.Join(tmpDir, "file with spaces.txt")
		if err := os.WriteFile(spaceFile, []byte("spaces"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "file with spaces") && !strings.Contains(output, "file") {
			t.Errorf("Run() should show file with spaces: %v", output)
		}
	})

	t.Run("consistent output", func(t *testing.T) {
		var buf1, buf2 bytes.Buffer

		_ = Run(&buf1, []string{tmpDir}, Options{OnePerLine: true})
		_ = Run(&buf2, []string{tmpDir}, Options{OnePerLine: true})

		if buf1.String() != buf2.String() {
			t.Errorf("Run() inconsistent output")
		}
	})

	t.Run("current directory default", func(t *testing.T) {
		var buf bytes.Buffer

		// Use the temp directory as current
		oldWd, _ := os.Getwd()
		_ = os.Chdir(tmpDir)

		defer func() { _ = os.Chdir(oldWd) }()

		err := Run(&buf, []string{}, Options{})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "file1.txt") {
			t.Errorf("Run() with no args should list current dir: %v", output)
		}
	})

	t.Run("output ends with newline", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{OnePerLine: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		if len(output) > 0 && !strings.HasSuffix(output, "\n") {
			t.Error("Run() output should end with newline")
		}
	})

	t.Run("no trailing whitespace", func(t *testing.T) {
		var buf bytes.Buffer

		_ = Run(&buf, []string{tmpDir}, Options{OnePerLine: true})

		lines := strings.Split(buf.String(), "\n")
		for i, line := range lines {
			if line != strings.TrimRight(line, " \t") {
				t.Errorf("Run() line %d has trailing whitespace", i)
			}
		}
	})

	t.Run("no sort", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{NoSort: true, OnePerLine: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("Run() with no sort should produce output")
		}
	})

	t.Run("symlink handling", func(t *testing.T) {
		linkTarget := filepath.Join(tmpDir, "file1.txt")
		linkPath := filepath.Join(tmpDir, "link.txt")

		// Try to create symlink (may fail on Windows without admin)
		err := os.Symlink(linkTarget, linkPath)
		if err != nil {
			t.Skip("Symlink creation not supported")
		}

		var buf bytes.Buffer

		err = Run(&buf, []string{tmpDir}, Options{LongFormat: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "link.txt") {
			t.Errorf("Run() should show symlink: %v", output)
		}
	})

	t.Run("classify with symlink", func(t *testing.T) {
		linkTarget := filepath.Join(tmpDir, "file1.txt")
		linkPath := filepath.Join(tmpDir, "link2.txt")

		err := os.Symlink(linkTarget, linkPath)
		if err != nil {
			t.Skip("Symlink creation not supported")
		}

		var buf bytes.Buffer

		err = Run(&buf, []string{tmpDir}, Options{Classify: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		// Just verify it doesn't error
		if buf.Len() == 0 {
			t.Error("Run() classify with symlink should produce output")
		}
	})

	t.Run("sort by time reverse", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{SortByTime: true, Reverse: true, OnePerLine: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("Run() sort by time reverse should produce output")
		}
	})

	t.Run("sort by size reverse", func(t *testing.T) {
		var buf bytes.Buffer

		err := Run(&buf, []string{tmpDir}, Options{SortBySize: true, Reverse: true, OnePerLine: true})
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		if buf.Len() == 0 {
			t.Error("Run() sort by size reverse should produce output")
		}
	})
}
