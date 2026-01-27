package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunMkdir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mkdir_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("create single directory", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "newdir")
		err := RunMkdir([]string{dir}, false)
		if err != nil {
			t.Fatalf("RunMkdir() error = %v", err)
		}

		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("Directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("Created path is not a directory")
		}
	})

	t.Run("create nested directories with parents", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "a", "b", "c")
		err := RunMkdir([]string{dir}, true)
		if err != nil {
			t.Fatalf("RunMkdir() error = %v", err)
		}

		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("Directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("Created path is not a directory")
		}
	})

	t.Run("fail without parents flag", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "x", "y", "z")
		err := RunMkdir([]string{dir}, false)
		if err == nil {
			t.Error("RunMkdir() expected error without parents flag")
		}
	})

	t.Run("no arguments", func(t *testing.T) {
		err := RunMkdir([]string{}, false)
		if err == nil {
			t.Error("RunMkdir() expected error with no arguments")
		}
	})
}

func TestRunRmdir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rmdir_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("remove empty directory", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "emptydir")
		if err := os.Mkdir(dir, 0755); err != nil {
			t.Fatal(err)
		}

		err := RunRmdir([]string{dir})
		if err != nil {
			t.Fatalf("RunRmdir() error = %v", err)
		}

		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Error("Directory still exists")
		}
	})

	t.Run("fail on non-empty directory", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "nonempty")
		if err := os.Mkdir(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		err := RunRmdir([]string{dir})
		if err == nil {
			t.Error("RunRmdir() expected error on non-empty directory")
		}
	})

	t.Run("no arguments", func(t *testing.T) {
		err := RunRmdir([]string{})
		if err == nil {
			t.Error("RunRmdir() expected error with no arguments")
		}
	})
}

func TestRunRm(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rm_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("remove file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "testfile.txt")
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		err := RunRm([]string{file}, false, false)
		if err != nil {
			t.Fatalf("RunRm() error = %v", err)
		}

		if _, err := os.Stat(file); !os.IsNotExist(err) {
			t.Error("File still exists")
		}
	})

	t.Run("remove directory recursively", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "subdir")
		if err := os.MkdirAll(filepath.Join(dir, "nested"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		err := RunRm([]string{dir}, true, false)
		if err != nil {
			t.Fatalf("RunRm() error = %v", err)
		}

		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Error("Directory still exists")
		}
	})

	t.Run("fail on non-empty directory without recursive", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "anotherdir")
		if err := os.Mkdir(dir, 0755); err != nil {
			t.Fatal(err)
		}
		// Create a file inside to make it non-empty
		if err := os.WriteFile(filepath.Join(dir, "inner.txt"), []byte("data"), 0644); err != nil {
			t.Fatal(err)
		}

		err := RunRm([]string{dir}, false, false)
		if err == nil {
			t.Error("RunRm() expected error on non-empty directory without recursive flag")
		}
	})

	t.Run("force mode ignores nonexistent", func(t *testing.T) {
		err := RunRm([]string{"/nonexistent/file"}, false, true)
		if err != nil {
			t.Errorf("RunRm() with force should not error: %v", err)
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
		err := RunTouch([]string{file})
		if err != nil {
			t.Fatalf("RunTouch() error = %v", err)
		}

		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Error("File not created")
		}
	})

	t.Run("update existing file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "existing.txt")
		if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}

		err := RunTouch([]string{file})
		if err != nil {
			t.Fatalf("RunTouch() error = %v", err)
		}
	})

	t.Run("no arguments", func(t *testing.T) {
		err := RunTouch([]string{})
		if err == nil {
			t.Error("RunTouch() expected error with no arguments")
		}
	})
}

func TestRunStat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stat_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	file := filepath.Join(tmpDir, "statfile.txt")
	if err := os.WriteFile(file, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("stat file", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunStat(&buf, []string{file}, false)
		if err != nil {
			t.Fatalf("RunStat() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "statfile.txt") {
			t.Errorf("RunStat() output missing filename: %v", output)
		}
	})

	t.Run("stat json mode", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunStat(&buf, []string{file}, true)
		if err != nil {
			t.Fatalf("RunStat() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "statfile.txt") || !strings.Contains(output, "{") {
			t.Errorf("RunStat() JSON output invalid: %v", output)
		}
	})

	t.Run("stat nonexistent", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunStat(&buf, []string{"/nonexistent/path"}, false)
		if err == nil {
			t.Error("RunStat() expected error for nonexistent file")
		}
	})

	t.Run("no arguments", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunStat(&buf, []string{}, false)
		if err == nil {
			t.Error("RunStat() expected error with no arguments")
		}
	})
}

func TestRunCopy(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cp_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("copy file", func(t *testing.T) {
		src := filepath.Join(tmpDir, "source.txt")
		dst := filepath.Join(tmpDir, "dest.txt")
		content := []byte("copy test content")

		if err := os.WriteFile(src, content, 0644); err != nil {
			t.Fatal(err)
		}

		err := RunCopy([]string{src, dst})
		if err != nil {
			t.Fatalf("RunCopy() error = %v", err)
		}

		dstContent, err := os.ReadFile(dst)
		if err != nil {
			t.Fatalf("Failed to read destination: %v", err)
		}
		if !bytes.Equal(dstContent, content) {
			t.Errorf("RunCopy() content mismatch")
		}
	})

	t.Run("missing arguments", func(t *testing.T) {
		err := RunCopy([]string{})
		if err == nil {
			t.Error("RunCopy() expected error with no arguments")
		}
	})
}

func TestRunMove(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mv_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("move file", func(t *testing.T) {
		src := filepath.Join(tmpDir, "moveme.txt")
		dst := filepath.Join(tmpDir, "moved.txt")

		if err := os.WriteFile(src, []byte("move test"), 0644); err != nil {
			t.Fatal(err)
		}

		err := RunMove([]string{src, dst})
		if err != nil {
			t.Fatalf("RunMove() error = %v", err)
		}

		if _, err := os.Stat(src); !os.IsNotExist(err) {
			t.Error("Source file still exists")
		}
		if _, err := os.Stat(dst); err != nil {
			t.Error("Destination file not created")
		}
	})

	t.Run("missing arguments", func(t *testing.T) {
		err := RunMove([]string{})
		if err == nil {
			t.Error("RunMove() expected error with no arguments")
		}
	})
}
