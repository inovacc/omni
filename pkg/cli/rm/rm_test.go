package rm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunRm(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rm_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("remove single file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "file1.txt")
		_ = os.WriteFile(file, []byte("content"), 0644)

		err := RunRm([]string{file}, RmOptions{})
		if err != nil {
			t.Fatalf("RunRm() error = %v", err)
		}

		if _, err := os.Stat(file); !os.IsNotExist(err) {
			t.Error("RunRm() did not remove file")
		}
	})

	t.Run("remove multiple files", func(t *testing.T) {
		file1 := filepath.Join(tmpDir, "multi1.txt")
		file2 := filepath.Join(tmpDir, "multi2.txt")
		_ = os.WriteFile(file1, []byte("1"), 0644)
		_ = os.WriteFile(file2, []byte("2"), 0644)

		err := RunRm([]string{file1, file2}, RmOptions{})
		if err != nil {
			t.Fatalf("RunRm() error = %v", err)
		}

		if _, err := os.Stat(file1); !os.IsNotExist(err) {
			t.Error("RunRm() did not remove file1")
		}
		if _, err := os.Stat(file2); !os.IsNotExist(err) {
			t.Error("RunRm() did not remove file2")
		}
	})

	t.Run("remove directory recursively", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "dir_to_remove")
		_ = os.Mkdir(dir, 0755)
		_ = os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644)

		err := RunRm([]string{dir}, RmOptions{Recursive: true})
		if err != nil {
			t.Fatalf("RunRm() -r error = %v", err)
		}

		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Error("RunRm() -r did not remove directory")
		}
	})

	t.Run("fail removing non-empty directory without recursive", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "dir_no_r")
		_ = os.Mkdir(dir, 0755)
		_ = os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644)

		err := RunRm([]string{dir}, RmOptions{Recursive: false})
		if err == nil {
			t.Error("RunRm() expected error removing non-empty directory without -r")
		}

		// Clean up
		_ = os.RemoveAll(dir)
	})

	t.Run("force ignore nonexistent", func(t *testing.T) {
		err := RunRm([]string{"/nonexistent/file.txt"}, RmOptions{Force: true})
		if err != nil {
			t.Errorf("RunRm() -f should ignore nonexistent: %v", err)
		}
	})

	t.Run("fail on nonexistent without force", func(t *testing.T) {
		err := RunRm([]string{"/nonexistent/file.txt"}, RmOptions{})
		if err == nil {
			t.Error("RunRm() expected error for nonexistent file")
		}
	})

	t.Run("missing operand without force", func(t *testing.T) {
		err := RunRm([]string{}, RmOptions{})
		if err == nil {
			t.Error("RunRm() expected error for missing operand")
		}
	})

	t.Run("missing operand with force", func(t *testing.T) {
		err := RunRm([]string{}, RmOptions{Force: true})
		if err != nil {
			t.Error("RunRm() -f should not error with no args")
		}
	})
}

func TestRunRmdir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rmdir_rm_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("remove empty directory", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "empty_dir")
		_ = os.Mkdir(dir, 0755)

		err := RunRmdir([]string{dir}, RmdirOptions{})
		if err != nil {
			t.Fatalf("RunRmdir() error = %v", err)
		}

		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Error("RunRmdir() did not remove directory")
		}
	})

	t.Run("fail removing non-empty directory", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "nonempty_dir")
		_ = os.Mkdir(dir, 0755)
		_ = os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644)

		err := RunRmdir([]string{dir}, RmdirOptions{})
		if err == nil {
			t.Error("RunRmdir() expected error for non-empty directory")
		}

		// Clean up
		_ = os.RemoveAll(dir)
	})

	t.Run("missing operand", func(t *testing.T) {
		err := RunRmdir([]string{}, RmdirOptions{})
		if err == nil {
			t.Error("RunRmdir() expected error for missing operand")
		}
	})
}
