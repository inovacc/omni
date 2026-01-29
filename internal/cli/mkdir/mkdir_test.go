package mkdir

import (
	"os"
	"path/filepath"
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

		err := RunMkdir([]string{dir}, Options{})
		if err != nil {
			t.Fatalf("RunMkdir() error = %v", err)
		}

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Error("RunMkdir() did not create directory")
		}
	})

	t.Run("create multiple directories", func(t *testing.T) {
		dir1 := filepath.Join(tmpDir, "multi1")
		dir2 := filepath.Join(tmpDir, "multi2")

		err := RunMkdir([]string{dir1, dir2}, Options{})
		if err != nil {
			t.Fatalf("RunMkdir() error = %v", err)
		}

		if _, err := os.Stat(dir1); os.IsNotExist(err) {
			t.Error("RunMkdir() did not create dir1")
		}

		if _, err := os.Stat(dir2); os.IsNotExist(err) {
			t.Error("RunMkdir() did not create dir2")
		}
	})

	t.Run("create with parents", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "parent", "child", "grandchild")

		err := RunMkdir([]string{dir}, Options{Parents: true})
		if err != nil {
			t.Fatalf("RunMkdir() error = %v", err)
		}

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Error("RunMkdir() -p did not create nested directories")
		}
	})

	t.Run("fail without parents", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "noparent", "child")

		err := RunMkdir([]string{dir}, Options{Parents: false})
		if err == nil {
			t.Error("RunMkdir() expected error without -p for nested path")
		}
	})

	t.Run("directory already exists", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "existing")
		_ = os.Mkdir(dir, 0755)

		err := RunMkdir([]string{dir}, Options{})
		if err == nil {
			t.Error("RunMkdir() expected error for existing directory")
		}
	})

	t.Run("parents with existing", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "existingparent")
		_ = os.Mkdir(dir, 0755)

		// Should succeed with -p even if already exists
		err := RunMkdir([]string{dir}, Options{Parents: true})
		if err != nil {
			t.Errorf("RunMkdir() -p should not fail for existing: %v", err)
		}
	})

	t.Run("missing operand", func(t *testing.T) {
		err := RunMkdir([]string{}, Options{})
		if err == nil {
			t.Error("RunMkdir() expected error for missing operand")
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
		dir := filepath.Join(tmpDir, "todelete")
		_ = os.Mkdir(dir, 0755)

		err := RunRmdir([]string{dir}, RmdirOptions{})
		if err != nil {
			t.Fatalf("RunRmdir() error = %v", err)
		}

		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Error("RunRmdir() did not remove directory")
		}
	})

	t.Run("remove multiple directories", func(t *testing.T) {
		dir1 := filepath.Join(tmpDir, "rm1")
		dir2 := filepath.Join(tmpDir, "rm2")
		_ = os.Mkdir(dir1, 0755)
		_ = os.Mkdir(dir2, 0755)

		err := RunRmdir([]string{dir1, dir2}, RmdirOptions{})
		if err != nil {
			t.Fatalf("RunRmdir() error = %v", err)
		}

		if _, err := os.Stat(dir1); !os.IsNotExist(err) {
			t.Error("RunRmdir() did not remove dir1")
		}

		if _, err := os.Stat(dir2); !os.IsNotExist(err) {
			t.Error("RunRmdir() did not remove dir2")
		}
	})

	t.Run("fail for non-empty directory", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "nonempty")
		_ = os.Mkdir(dir, 0755)
		_ = os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644)

		err := RunRmdir([]string{dir}, RmdirOptions{})
		if err == nil {
			t.Error("RunRmdir() expected error for non-empty directory")
		}
	})

	t.Run("fail for nonexistent", func(t *testing.T) {
		err := RunRmdir([]string{"/nonexistent/dir"}, RmdirOptions{})
		if err == nil {
			t.Error("RunRmdir() expected error for nonexistent directory")
		}
	})

	t.Run("missing operand", func(t *testing.T) {
		err := RunRmdir([]string{}, RmdirOptions{})
		if err == nil {
			t.Error("RunRmdir() expected error for missing operand")
		}
	})
}
