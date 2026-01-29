package copy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunCopy(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "copy_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("copy single file", func(t *testing.T) {
		src := filepath.Join(tmpDir, "source1.txt")
		dst := filepath.Join(tmpDir, "dest1.txt")
		content := []byte("hello world")
		_ = os.WriteFile(src, content, 0644)

		err := RunCopy([]string{src, dst}, CopyOptions{})
		if err != nil {
			t.Fatalf("RunCopy() error = %v", err)
		}

		data, err := os.ReadFile(dst)
		if err != nil {
			t.Fatalf("RunCopy() dest not created: %v", err)
		}

		if string(data) != string(content) {
			t.Errorf("RunCopy() content = %q, want %q", data, content)
		}
	})

	t.Run("copy to directory", func(t *testing.T) {
		src := filepath.Join(tmpDir, "source2.txt")
		destDir := filepath.Join(tmpDir, "destdir")
		_ = os.WriteFile(src, []byte("content"), 0644)
		_ = os.Mkdir(destDir, 0755)

		err := RunCopy([]string{src, destDir}, CopyOptions{})
		if err != nil {
			t.Fatalf("RunCopy() error = %v", err)
		}

		expected := filepath.Join(destDir, "source2.txt")
		if _, err := os.Stat(expected); os.IsNotExist(err) {
			t.Error("RunCopy() did not copy to directory")
		}
	})

	t.Run("copy multiple files to directory", func(t *testing.T) {
		src1 := filepath.Join(tmpDir, "multi1.txt")
		src2 := filepath.Join(tmpDir, "multi2.txt")
		destDir := filepath.Join(tmpDir, "multidir")

		_ = os.WriteFile(src1, []byte("1"), 0644)
		_ = os.WriteFile(src2, []byte("2"), 0644)
		_ = os.Mkdir(destDir, 0755)

		err := RunCopy([]string{src1, src2, destDir}, CopyOptions{})
		if err != nil {
			t.Fatalf("RunCopy() multiple error = %v", err)
		}

		if _, err := os.Stat(filepath.Join(destDir, "multi1.txt")); os.IsNotExist(err) {
			t.Error("RunCopy() did not copy first file")
		}

		if _, err := os.Stat(filepath.Join(destDir, "multi2.txt")); os.IsNotExist(err) {
			t.Error("RunCopy() did not copy second file")
		}
	})

	t.Run("fail multiple to non-directory", func(t *testing.T) {
		src1 := filepath.Join(tmpDir, "f1.txt")
		src2 := filepath.Join(tmpDir, "f2.txt")
		notDir := filepath.Join(tmpDir, "notdir.txt")

		_ = os.WriteFile(src1, []byte("1"), 0644)
		_ = os.WriteFile(src2, []byte("2"), 0644)
		_ = os.WriteFile(notDir, []byte("file"), 0644)

		err := RunCopy([]string{src1, src2, notDir}, CopyOptions{})
		if err == nil {
			t.Error("RunCopy() expected error for non-directory target")
		}
	})

	t.Run("missing operand", func(t *testing.T) {
		err := RunCopy([]string{"single"}, CopyOptions{})
		if err == nil {
			t.Error("RunCopy() expected error for missing operand")
		}
	})

	t.Run("nonexistent source", func(t *testing.T) {
		err := RunCopy([]string{"/nonexistent/file.txt", tmpDir}, CopyOptions{})
		if err == nil {
			t.Error("RunCopy() expected error for nonexistent source")
		}
	})
}

func TestRunMove(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "move_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("move single file", func(t *testing.T) {
		src := filepath.Join(tmpDir, "tomove.txt")
		dst := filepath.Join(tmpDir, "moved.txt")
		_ = os.WriteFile(src, []byte("content"), 0644)

		err := RunMove([]string{src, dst}, MoveOptions{})
		if err != nil {
			t.Fatalf("RunMove() error = %v", err)
		}

		if _, err := os.Stat(src); !os.IsNotExist(err) {
			t.Error("RunMove() source should not exist after move")
		}

		if _, err := os.Stat(dst); os.IsNotExist(err) {
			t.Error("RunMove() destination should exist after move")
		}
	})

	t.Run("move to directory", func(t *testing.T) {
		src := filepath.Join(tmpDir, "moveme.txt")
		destDir := filepath.Join(tmpDir, "movedir")
		_ = os.WriteFile(src, []byte("content"), 0644)
		_ = os.Mkdir(destDir, 0755)

		err := RunMove([]string{src, destDir}, MoveOptions{})
		if err != nil {
			t.Fatalf("RunMove() error = %v", err)
		}

		expected := filepath.Join(destDir, "moveme.txt")
		if _, err := os.Stat(expected); os.IsNotExist(err) {
			t.Error("RunMove() did not move to directory")
		}
	})

	t.Run("move multiple to directory", func(t *testing.T) {
		src1 := filepath.Join(tmpDir, "mv1.txt")
		src2 := filepath.Join(tmpDir, "mv2.txt")
		destDir := filepath.Join(tmpDir, "mvdir")

		_ = os.WriteFile(src1, []byte("1"), 0644)
		_ = os.WriteFile(src2, []byte("2"), 0644)
		_ = os.Mkdir(destDir, 0755)

		err := RunMove([]string{src1, src2, destDir}, MoveOptions{})
		if err != nil {
			t.Fatalf("RunMove() multiple error = %v", err)
		}

		if _, err := os.Stat(filepath.Join(destDir, "mv1.txt")); os.IsNotExist(err) {
			t.Error("RunMove() did not move first file")
		}
	})

	t.Run("fail multiple to non-directory", func(t *testing.T) {
		src1 := filepath.Join(tmpDir, "m1.txt")
		src2 := filepath.Join(tmpDir, "m2.txt")
		notDir := filepath.Join(tmpDir, "notadir.txt")

		_ = os.WriteFile(src1, []byte("1"), 0644)
		_ = os.WriteFile(src2, []byte("2"), 0644)
		_ = os.WriteFile(notDir, []byte("file"), 0644)

		err := RunMove([]string{src1, src2, notDir}, MoveOptions{})
		if err == nil {
			t.Error("RunMove() expected error for non-directory target")
		}
	})

	t.Run("missing operand", func(t *testing.T) {
		err := RunMove([]string{"single"}, MoveOptions{})
		if err == nil {
			t.Error("RunMove() expected error for missing operand")
		}
	})
}

func TestCopyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "copyfile_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("copy regular file", func(t *testing.T) {
		src := filepath.Join(tmpDir, "src.txt")
		dst := filepath.Join(tmpDir, "dst.txt")
		content := []byte("test content")
		_ = os.WriteFile(src, content, 0644)

		err := copyFile(src, dst)
		if err != nil {
			t.Fatalf("copyFile() error = %v", err)
		}

		data, _ := os.ReadFile(dst)
		if string(data) != string(content) {
			t.Errorf("copyFile() content mismatch")
		}
	})

	t.Run("fail for directory", func(t *testing.T) {
		dir := filepath.Join(tmpDir, "srcdir")
		_ = os.Mkdir(dir, 0755)

		err := copyFile(dir, filepath.Join(tmpDir, "dstdir"))
		if err == nil {
			t.Error("copyFile() expected error for directory")
		}
	})
}
