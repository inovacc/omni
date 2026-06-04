package copy

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestCopyFilePreservesMode is a regression test for finding cp-no-mode-preserve
// (CWE-732): copying a 0600 secret must not yield a world/group-readable 0644
// copy. The destination mode must NARROW to match the source, never widen.
//
// POSIX-only: Windows does not model Unix permission bits, so os.Chmod /
// FileMode.Perm() are not meaningful there.
func TestCopyFilePreservesMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not meaningful on Windows")
	}

	tmpDir := t.TempDir()

	t.Run("copyFile preserves 0600 source mode", func(t *testing.T) {
		src := filepath.Join(tmpDir, "secret.txt")
		dst := filepath.Join(tmpDir, "secret_copy.txt")
		if err := os.WriteFile(src, []byte("top secret"), 0600); err != nil {
			t.Fatal(err)
		}
		// WriteFile applies umask; force the exact source mode.
		if err := os.Chmod(src, 0600); err != nil {
			t.Fatal(err)
		}

		if err := copyFile(src, dst); err != nil {
			t.Fatalf("copyFile() error = %v", err)
		}

		info, err := os.Stat(dst)
		if err != nil {
			t.Fatalf("stat dest: %v", err)
		}
		if got := info.Mode().Perm(); got != 0600 {
			t.Errorf("copyFile() dest mode = %#o, want %#o", got, 0600)
		}
	})

	t.Run("RunCopy preserves 0600 source mode", func(t *testing.T) {
		src := filepath.Join(tmpDir, "secret2.txt")
		dst := filepath.Join(tmpDir, "secret2_copy.txt")
		if err := os.WriteFile(src, []byte("top secret"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(src, 0600); err != nil {
			t.Fatal(err)
		}

		if err := RunCopy([]string{src, dst}, CopyOptions{}); err != nil {
			t.Fatalf("RunCopy() error = %v", err)
		}

		info, err := os.Stat(dst)
		if err != nil {
			t.Fatalf("stat dest: %v", err)
		}
		if got := info.Mode().Perm(); got != 0600 {
			t.Errorf("RunCopy() dest mode = %#o, want %#o", got, 0600)
		}
	})
}
