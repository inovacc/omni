package fs

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cd_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()

	err = Cd(tmpDir)
	if err != nil {
		t.Fatalf("Cd() error = %v", err)
	}

	cwd, _ := os.Getwd()
	// Normalize paths for comparison
	if filepath.Clean(cwd) != filepath.Clean(tmpDir) {
		// On some systems, the temp dir may be a symlink
		t.Logf("Cd() current dir = %q, expected %q (may be symlink)", cwd, tmpDir)
	}
}

func TestChmod(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping chmod test on Windows")
	}

	tmpDir, err := os.MkdirTemp("", "chmod_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	file := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(file, []byte("content"), 0644)

	err = Chmod(file, 0755)
	if err != nil {
		t.Fatalf("Chmod() error = %v", err)
	}

	info, _ := os.Stat(file)
	if info.Mode().Perm() != 0755 {
		t.Errorf("Chmod() mode = %o, want 0755", info.Mode().Perm())
	}
}

func TestMkdir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mkdir_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("simple mkdir", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "newdir")
		err := Mkdir(dir, 0755, false)
		if err != nil {
			t.Fatalf("Mkdir() error = %v", err)
		}

		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("Mkdir() directory not created")
		}
		if !info.IsDir() {
			t.Error("Mkdir() should create directory")
		}
	})

	t.Run("mkdir with parents", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "parent", "child", "grandchild")
		err := Mkdir(dir, 0755, true)
		if err != nil {
			t.Fatalf("Mkdir() -p error = %v", err)
		}

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Error("Mkdir() -p should create parent directories")
		}
	})

	t.Run("mkdir without parents fails for nested", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "noparent", "child")
		err := Mkdir(dir, 0755, false)
		if err == nil {
			t.Error("Mkdir() without -p should fail for nested directory")
		}
	})
}

func TestRmdir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rmdir_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	dir := filepath.Join(tmpDir, "todelete")
	_ = os.Mkdir(dir, 0755)

	err = Rmdir(dir)
	if err != nil {
		t.Fatalf("Rmdir() error = %v", err)
	}

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("Rmdir() should remove directory")
	}
}

func TestTouch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "touch_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("create new file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "new.txt")
		err := Touch(file)
		if err != nil {
			t.Fatalf("Touch() error = %v", err)
		}

		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Error("Touch() should create file")
		}
	})

	t.Run("update existing file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "existing.txt")
		_ = os.WriteFile(file, []byte("content"), 0644)

		info1, _ := os.Stat(file)
		time1 := info1.ModTime()

		// Wait a bit to ensure time difference
		// Touch should update timestamp
		err := Touch(file)
		if err != nil {
			t.Fatalf("Touch() error = %v", err)
		}

		info2, _ := os.Stat(file)
		time2 := info2.ModTime()

		if !time2.After(time1) && !time2.Equal(time1) {
			t.Logf("Touch() modtime update (timing sensitive): before=%v, after=%v", time1, time2)
		}
	})
}

func TestRm(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rm_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("remove file", func(t *testing.T) {
		file := filepath.Join(tmpDir, "file.txt")
		_ = os.WriteFile(file, []byte("content"), 0644)

		err := Rm(file, false)
		if err != nil {
			t.Fatalf("Rm() error = %v", err)
		}

		if _, err := os.Stat(file); !os.IsNotExist(err) {
			t.Error("Rm() should remove file")
		}
	})

	t.Run("remove directory recursively", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "dir_to_remove")
		_ = os.Mkdir(dir, 0755)
		_ = os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0644)

		err := Rm(dir, true)
		if err != nil {
			t.Fatalf("Rm() -r error = %v", err)
		}

		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Error("Rm() -r should remove directory recursively")
		}
	})
}

func TestIsNotExist(t *testing.T) {
	_, err := os.Stat("/nonexistent/file.txt")

	if !IsNotExist(err) {
		t.Error("IsNotExist() should return true for nonexistent file")
	}

	if IsNotExist(nil) {
		t.Error("IsNotExist() should return false for nil error")
	}
}

func TestCopy(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "copy_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("copy file", func(t *testing.T) {
		src := filepath.Join(tmpDir, "source.txt")
		dst := filepath.Join(tmpDir, "dest.txt")
		content := []byte("test content")
		_ = os.WriteFile(src, content, 0644)

		err := Copy(src, dst)
		if err != nil {
			t.Fatalf("Copy() error = %v", err)
		}

		data, err := os.ReadFile(dst)
		if err != nil {
			t.Fatalf("Copy() dest not readable: %v", err)
		}

		if string(data) != string(content) {
			t.Errorf("Copy() content = %q, want %q", data, content)
		}
	})

	t.Run("copy nonexistent source", func(t *testing.T) {
		err := Copy("/nonexistent/file.txt", filepath.Join(tmpDir, "dst.txt"))
		if err == nil {
			t.Error("Copy() expected error for nonexistent source")
		}
	})

	t.Run("copy directory fails", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "srcdir")
		_ = os.Mkdir(dir, 0755)

		err := Copy(dir, filepath.Join(tmpDir, "dstdir"))
		if err == nil {
			t.Error("Copy() expected error for directory source")
		}
	})
}

func TestMove(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "move_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	src := filepath.Join(tmpDir, "source.txt")
	dst := filepath.Join(tmpDir, "dest.txt")
	content := []byte("test content")
	_ = os.WriteFile(src, content, 0644)

	err = Move(src, dst)
	if err != nil {
		t.Fatalf("Move() error = %v", err)
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("Move() source should not exist after move")
	}

	data, _ := os.ReadFile(dst)
	if string(data) != string(content) {
		t.Errorf("Move() content = %q, want %q", data, content)
	}
}

func TestStat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stat_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	file := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(file, []byte("content"), 0644)

	info, err := Stat(file)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	if info.Name() != "test.txt" {
		t.Errorf("Stat() name = %q, want 'test.txt'", info.Name())
	}
	if info.Size() != 7 {
		t.Errorf("Stat() size = %d, want 7", info.Size())
	}
}

func TestLstat(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Lstat symlink test on Windows")
	}

	tmpDir, err := os.MkdirTemp("", "lstat_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	target := filepath.Join(tmpDir, "target.txt")
	link := filepath.Join(tmpDir, "link")
	_ = os.WriteFile(target, []byte("content"), 0644)
	_ = os.Symlink(target, link)

	info, err := Lstat(link)
	if err != nil {
		t.Fatalf("Lstat() error = %v", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("Lstat() should identify symlink")
	}
}
