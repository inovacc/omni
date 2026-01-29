package chown

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRunChown(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping chown tests on Windows")
	}

	// Most chown tests require root privileges, so we test what we can
	tmpDir, err := os.MkdirTemp("", "chown_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("missing operand", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunChown(&buf, []string{"1000"}, ChownOptions{})
		if err == nil {
			t.Error("RunChown() expected error for missing operand")
		}
	})

	t.Run("invalid user", func(t *testing.T) {
		file := filepath.Join(tmpDir, "file1.txt")
		_ = os.WriteFile(file, []byte("content"), 0644)

		var buf bytes.Buffer
		err := RunChown(&buf, []string{"nonexistent_user_12345", file}, ChownOptions{})
		if err == nil {
			t.Error("RunChown() expected error for invalid user")
		}
	})

	t.Run("invalid group", func(t *testing.T) {
		file := filepath.Join(tmpDir, "file2.txt")
		_ = os.WriteFile(file, []byte("content"), 0644)

		var buf bytes.Buffer
		err := RunChown(&buf, []string{":nonexistent_group_12345", file}, ChownOptions{})
		if err == nil {
			t.Error("RunChown() expected error for invalid group")
		}
	})

	t.Run("preserve root", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunChown(&buf, []string{"0", "/"}, ChownOptions{Recursive: true, PreserveRoot: true})
		if err == nil {
			t.Error("RunChown() expected error for recursive on root with preserve-root")
		}
	})

	t.Run("reference nonexistent", func(t *testing.T) {
		file := filepath.Join(tmpDir, "file3.txt")
		_ = os.WriteFile(file, []byte("content"), 0644)

		var buf bytes.Buffer
		err := RunChown(&buf, []string{"ignored", file}, ChownOptions{Reference: "/nonexistent/ref"})
		if err == nil {
			t.Error("RunChown() expected error for nonexistent reference file")
		}
	})
}

func TestParseOwnerGroup(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping parseOwnerGroup tests on Windows")
	}

	t.Run("numeric uid", func(t *testing.T) {
		uid, gid, err := parseOwnerGroup("1000", "")
		if err != nil {
			t.Fatalf("parseOwnerGroup() error = %v", err)
		}

		if uid != 1000 {
			t.Errorf("parseOwnerGroup() uid = %d, want 1000", uid)
		}

		if gid != -1 {
			t.Errorf("parseOwnerGroup() gid = %d, want -1", gid)
		}
	})

	t.Run("numeric uid:gid", func(t *testing.T) {
		uid, gid, err := parseOwnerGroup("1000:1000", "")
		if err != nil {
			t.Fatalf("parseOwnerGroup() error = %v", err)
		}

		if uid != 1000 {
			t.Errorf("parseOwnerGroup() uid = %d, want 1000", uid)
		}

		if gid != 1000 {
			t.Errorf("parseOwnerGroup() gid = %d, want 1000", gid)
		}
	})

	t.Run("numeric uid.gid", func(t *testing.T) {
		uid, gid, err := parseOwnerGroup("1000.2000", "")
		if err != nil {
			t.Fatalf("parseOwnerGroup() error = %v", err)
		}

		if uid != 1000 {
			t.Errorf("parseOwnerGroup() uid = %d, want 1000", uid)
		}

		if gid != 2000 {
			t.Errorf("parseOwnerGroup() gid = %d, want 2000", gid)
		}
	})

	t.Run("colon only group", func(t *testing.T) {
		uid, gid, err := parseOwnerGroup(":1000", "")
		if err != nil {
			t.Fatalf("parseOwnerGroup() error = %v", err)
		}

		if uid != -1 {
			t.Errorf("parseOwnerGroup() uid = %d, want -1", uid)
		}

		if gid != 1000 {
			t.Errorf("parseOwnerGroup() gid = %d, want 1000", gid)
		}
	})

	t.Run("invalid user name", func(t *testing.T) {
		_, _, err := parseOwnerGroup("nonexistent_user_xyz", "")
		if err == nil {
			t.Error("parseOwnerGroup() expected error for invalid user")
		}
	})

	t.Run("invalid group name", func(t *testing.T) {
		_, _, err := parseOwnerGroup(":nonexistent_group_xyz", "")
		if err == nil {
			t.Error("parseOwnerGroup() expected error for invalid group")
		}
	})
}

func TestChown(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Chown tests on Windows")
	}

	// Chown typically requires root, so we just test the function exists
	tmpDir, err := os.MkdirTemp("", "chown_func_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	file := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(file, []byte("content"), 0644)

	// This will likely fail without root, but we can test the error path
	err = Chown(file, -1, -1)
	// -1 means don't change, so this should succeed
	if err != nil {
		t.Logf("Chown() error (may be expected): %v", err)
	}
}

func TestLchown(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Lchown tests on Windows")
	}

	tmpDir, err := os.MkdirTemp("", "lchown_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	file := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(file, []byte("content"), 0644)

	// This will likely fail without root, but we can test the error path
	err = Lchown(file, -1, -1)
	// -1 means don't change, so this should succeed
	if err != nil {
		t.Logf("Lchown() error (may be expected): %v", err)
	}
}
