package ln

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunLn(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping symlink tests on Windows")
	}

	tmpDir, err := os.MkdirTemp("", "ln_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("create hard link", func(t *testing.T) {
		target := filepath.Join(tmpDir, "target1.txt")
		link := filepath.Join(tmpDir, "link1")
		_ = os.WriteFile(target, []byte("content"), 0644)

		var buf bytes.Buffer

		err := RunLn(&buf, []string{target, link}, LnOptions{})
		if err != nil {
			t.Fatalf("RunLn() error = %v", err)
		}

		if _, err := os.Stat(link); os.IsNotExist(err) {
			t.Error("RunLn() did not create link")
		}
	})

	t.Run("create symbolic link", func(t *testing.T) {
		target := filepath.Join(tmpDir, "target2.txt")
		link := filepath.Join(tmpDir, "link2")
		_ = os.WriteFile(target, []byte("content"), 0644)

		var buf bytes.Buffer

		err := RunLn(&buf, []string{target, link}, LnOptions{Symbolic: true})
		if err != nil {
			t.Fatalf("RunLn() error = %v", err)
		}

		info, err := os.Lstat(link)
		if err != nil {
			t.Fatalf("RunLn() link not created: %v", err)
		}

		if info.Mode()&os.ModeSymlink == 0 {
			t.Error("RunLn() -s should create symbolic link")
		}
	})

	t.Run("force overwrite", func(t *testing.T) {
		target := filepath.Join(tmpDir, "target3.txt")
		link := filepath.Join(tmpDir, "link3")
		_ = os.WriteFile(target, []byte("content"), 0644)
		_ = os.WriteFile(link, []byte("existing"), 0644)

		var buf bytes.Buffer

		err := RunLn(&buf, []string{target, link}, LnOptions{Symbolic: true, Force: true})
		if err != nil {
			t.Fatalf("RunLn() -f error = %v", err)
		}
	})

	t.Run("backup existing", func(t *testing.T) {
		target := filepath.Join(tmpDir, "target4.txt")
		link := filepath.Join(tmpDir, "link4")
		_ = os.WriteFile(target, []byte("content"), 0644)
		_ = os.WriteFile(link, []byte("existing"), 0644)

		var buf bytes.Buffer

		err := RunLn(&buf, []string{target, link}, LnOptions{Symbolic: true, Backup: true})
		if err != nil {
			t.Fatalf("RunLn() -b error = %v", err)
		}

		if _, err := os.Stat(link + "~"); os.IsNotExist(err) {
			t.Error("RunLn() -b should create backup")
		}
	})

	t.Run("verbose mode", func(t *testing.T) {
		target := filepath.Join(tmpDir, "target5.txt")
		link := filepath.Join(tmpDir, "link5")
		_ = os.WriteFile(target, []byte("content"), 0644)

		var buf bytes.Buffer

		err := RunLn(&buf, []string{target, link}, LnOptions{Symbolic: true, Verbose: true})
		if err != nil {
			t.Fatalf("RunLn() -v error = %v", err)
		}

		if !strings.Contains(buf.String(), "->") {
			t.Errorf("RunLn() -v should show link creation: %s", buf.String())
		}
	})

	t.Run("multiple sources to directory", func(t *testing.T) {
		target1 := filepath.Join(tmpDir, "multi1.txt")
		target2 := filepath.Join(tmpDir, "multi2.txt")
		destDir := filepath.Join(tmpDir, "linkdir")

		_ = os.WriteFile(target1, []byte("1"), 0644)
		_ = os.WriteFile(target2, []byte("2"), 0644)
		_ = os.Mkdir(destDir, 0755)

		var buf bytes.Buffer

		err := RunLn(&buf, []string{target1, target2, destDir}, LnOptions{Symbolic: true})
		if err != nil {
			t.Fatalf("RunLn() multiple error = %v", err)
		}

		if _, err := os.Lstat(filepath.Join(destDir, "multi1.txt")); os.IsNotExist(err) {
			t.Error("RunLn() did not create first link")
		}

		if _, err := os.Lstat(filepath.Join(destDir, "multi2.txt")); os.IsNotExist(err) {
			t.Error("RunLn() did not create second link")
		}
	})

	t.Run("fail without force on existing", func(t *testing.T) {
		target := filepath.Join(tmpDir, "target6.txt")
		link := filepath.Join(tmpDir, "link6")
		_ = os.WriteFile(target, []byte("content"), 0644)
		_ = os.WriteFile(link, []byte("existing"), 0644)

		var buf bytes.Buffer

		err := RunLn(&buf, []string{target, link}, LnOptions{Symbolic: true})
		if err == nil {
			t.Error("RunLn() expected error for existing file")
		}
	})

	t.Run("missing operand", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunLn(&buf, []string{"single"}, LnOptions{})
		if err == nil {
			t.Error("RunLn() expected error for missing operand")
		}
	})

	t.Run("multiple to non-directory", func(t *testing.T) {
		target1 := filepath.Join(tmpDir, "m1.txt")
		target2 := filepath.Join(tmpDir, "m2.txt")
		notDir := filepath.Join(tmpDir, "notdir.txt")

		_ = os.WriteFile(target1, []byte("1"), 0644)
		_ = os.WriteFile(target2, []byte("2"), 0644)
		_ = os.WriteFile(notDir, []byte("file"), 0644)

		var buf bytes.Buffer

		err := RunLn(&buf, []string{target1, target2, notDir}, LnOptions{})
		if err == nil {
			t.Error("RunLn() expected error for non-directory target")
		}
	})
}

func TestSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping symlink tests on Windows")
	}

	tmpDir, err := os.MkdirTemp("", "symlink_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	target := filepath.Join(tmpDir, "target")
	link := filepath.Join(tmpDir, "symlink")
	_ = os.WriteFile(target, []byte("content"), 0644)

	err = Symlink(target, link)
	if err != nil {
		t.Fatalf("Symlink() error = %v", err)
	}

	info, _ := os.Lstat(link)
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("Symlink() should create symbolic link")
	}
}

func TestLink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping hard link tests on Windows")
	}

	tmpDir, err := os.MkdirTemp("", "link_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	target := filepath.Join(tmpDir, "target")
	link := filepath.Join(tmpDir, "hardlink")
	_ = os.WriteFile(target, []byte("content"), 0644)

	err = Link(target, link)
	if err != nil {
		t.Fatalf("Link() error = %v", err)
	}

	if _, err := os.Stat(link); os.IsNotExist(err) {
		t.Error("Link() should create hard link")
	}
}
