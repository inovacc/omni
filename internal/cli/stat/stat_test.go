package stat

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunStat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stat_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("stat file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "test.txt")
		_ = os.WriteFile(file, []byte("hello world"), 0644)

		var buf bytes.Buffer

		err := RunStat(&buf, []string{file}, StatOptions{})
		if err != nil {
			t.Fatalf("RunStat() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "File:") {
			t.Errorf("RunStat() output missing File:")
		}
		if !strings.Contains(output, "Size:") {
			t.Errorf("RunStat() output missing Size:")
		}
		if !strings.Contains(output, "regular file") {
			t.Errorf("RunStat() should indicate regular file")
		}
	})

	t.Run("stat directory", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "testdir")
		_ = os.Mkdir(dir, 0755)

		var buf bytes.Buffer

		err := RunStat(&buf, []string{dir}, StatOptions{})
		if err != nil {
			t.Fatalf("RunStat() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "directory") {
			t.Errorf("RunStat() should indicate directory")
		}
	})

	t.Run("stat JSON output", func(t *testing.T) {
		file := filepath.Join(tmpDir, "json.txt")
		_ = os.WriteFile(file, []byte("test content"), 0644)

		var buf bytes.Buffer

		err := RunStat(&buf, []string{file}, StatOptions{JSON: true})
		if err != nil {
			t.Fatalf("RunStat() error = %v", err)
		}

		var results []StatInfo
		if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
			t.Errorf("RunStat() JSON output invalid: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("RunStat() JSON got %d results, want 1", len(results))
		}

		if results[0].Name != "json.txt" {
			t.Errorf("RunStat() JSON name = %q, want 'json.txt'", results[0].Name)
		}
	})

	t.Run("stat multiple files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "multi1.txt")
		file2 := filepath.Join(tmpDir, "multi2.txt")

		_ = os.WriteFile(file1, []byte("a"), 0644)
		_ = os.WriteFile(file2, []byte("bb"), 0644)

		var buf bytes.Buffer

		err := RunStat(&buf, []string{file1, file2}, StatOptions{})
		if err != nil {
			t.Fatalf("RunStat() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "multi1.txt") || !strings.Contains(output, "multi2.txt") {
			t.Errorf("RunStat() should show both files")
		}
	})

	t.Run("stat missing operand", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunStat(&buf, []string{}, StatOptions{})
		if err == nil {
			t.Error("RunStat() expected error for missing operand")
		}
	})

	t.Run("stat nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunStat(&buf, []string{"/nonexistent/file.txt"}, StatOptions{})
		if err == nil {
			t.Error("RunStat() expected error for nonexistent file")
		}
	})
}

func TestRunTouch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "touch_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("create new file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "newfile.txt")

		err := RunTouch([]string{file}, TouchOptions{})
		if err != nil {
			t.Fatalf("RunTouch() error = %v", err)
		}

		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Error("RunTouch() did not create file")
		}
	})

	t.Run("update existing file timestamp", func(t *testing.T) {
		file := filepath.Join(tmpDir, "existing.txt")
		_ = os.WriteFile(file, []byte("content"), 0644)

		// Set old modification time
		oldTime := time.Now().Add(-24 * time.Hour)
		_ = os.Chtimes(file, oldTime, oldTime)

		info1, _ := os.Stat(file)
		oldMod := info1.ModTime()

		// Touch the file
		err := RunTouch([]string{file}, TouchOptions{})
		if err != nil {
			t.Fatalf("RunTouch() error = %v", err)
		}

		info2, _ := os.Stat(file)
		newMod := info2.ModTime()

		if !newMod.After(oldMod) {
			t.Errorf("RunTouch() did not update modification time")
		}
	})

	t.Run("touch multiple files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "touch1.txt")
		file2 := filepath.Join(tmpDir, "touch2.txt")

		err := RunTouch([]string{file1, file2}, TouchOptions{})
		if err != nil {
			t.Fatalf("RunTouch() error = %v", err)
		}

		if _, err := os.Stat(file1); os.IsNotExist(err) {
			t.Error("RunTouch() did not create file1")
		}
		if _, err := os.Stat(file2); os.IsNotExist(err) {
			t.Error("RunTouch() did not create file2")
		}
	})

	t.Run("touch missing operand", func(t *testing.T) {
		err := RunTouch([]string{}, TouchOptions{})
		if err == nil {
			t.Error("RunTouch() expected error for missing operand")
		}
	})
}
