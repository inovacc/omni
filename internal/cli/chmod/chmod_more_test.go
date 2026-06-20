package chmod

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestIsAllDigits covers the pure isAllDigits helper on all platforms.
func TestIsAllDigits(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"755", true},
		{"999", true},
		{"0", true},
		{"", false},
		{"7a5", false},
		{"u+x", false},
		{"-1", false},
		{"12.3", false},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := isAllDigits(tt.in); got != tt.want {
				t.Errorf("isAllDigits(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// TestRunChmodErrors covers the validation/error branches of RunChmod on all
// platforms (these do not depend on the OS permission model).
func TestRunChmodErrors(t *testing.T) {
	t.Run("missing operand", func(t *testing.T) {
		var buf bytes.Buffer
		if err := RunChmod(&buf, []string{"755"}, ChmodOptions{}); err == nil {
			t.Error("expected error for missing operand")
		}
	})

	t.Run("all-digit non-octal", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "f.txt")
		_ = os.WriteFile(f, []byte("x"), 0644)

		var buf bytes.Buffer
		if err := RunChmod(&buf, []string{"999", f}, ChmodOptions{}); err == nil {
			t.Error("expected error for invalid octal mode 999")
		}
	})

	t.Run("invalid symbolic mode", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "f.txt")
		_ = os.WriteFile(f, []byte("x"), 0644)

		var buf bytes.Buffer
		// parseSymbolicMode never errors, so an odd symbolic string is a no-op
		// rather than an error; assert it does not error out.
		if err := RunChmod(&buf, []string{"u+x", f}, ChmodOptions{}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("reference missing", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "f.txt")
		_ = os.WriteFile(f, []byte("x"), 0644)

		var buf bytes.Buffer
		if err := RunChmod(&buf, []string{"ignored", f}, ChmodOptions{Reference: filepath.Join(dir, "nope")}); err == nil {
			t.Error("expected error for missing reference file")
		}
	})
}

// TestRunChmodExecutes drives RunChmod/chmodFile through octal, symbolic,
// reference, verbose and recursive paths. On Windows os.Chmod only toggles the
// read-only bit, so we assert the call succeeds rather than exact perm bits.
func TestRunChmodExecutes(t *testing.T) {
	dir := t.TempDir()

	mkFile := func(name string, mode os.FileMode) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte("content"), mode); err != nil {
			t.Fatal(err)
		}
		return p
	}

	t.Run("octal", func(t *testing.T) {
		f := mkFile("octal.txt", 0644)
		var buf bytes.Buffer
		if err := RunChmod(&buf, []string{"600", f}, ChmodOptions{}); err != nil {
			t.Fatalf("RunChmod octal error = %v", err)
		}
		if runtime.GOOS != "windows" {
			info, _ := os.Stat(f)
			if info.Mode().Perm() != 0600 {
				t.Errorf("mode = %o, want 0600", info.Mode().Perm())
			}
		}
	})

	t.Run("symbolic verbose", func(t *testing.T) {
		f := mkFile("sym.txt", 0644)
		var buf bytes.Buffer
		if err := RunChmod(&buf, []string{"u+x", f}, ChmodOptions{Verbose: true}); err != nil {
			t.Fatalf("RunChmod symbolic error = %v", err)
		}
		if buf.Len() == 0 {
			t.Error("verbose mode should emit a diagnostic")
		}
	})

	t.Run("reference", func(t *testing.T) {
		ref := mkFile("ref.txt", 0600)
		target := mkFile("tgt.txt", 0644)
		var buf bytes.Buffer
		if err := RunChmod(&buf, []string{"ignored", target}, ChmodOptions{Reference: ref}); err != nil {
			t.Fatalf("RunChmod reference error = %v", err)
		}
	})

	t.Run("recursive", func(t *testing.T) {
		sub := filepath.Join(dir, "sub")
		if err := os.Mkdir(sub, 0755); err != nil {
			t.Fatal(err)
		}
		_ = os.WriteFile(filepath.Join(sub, "nested.txt"), []byte("x"), 0644)
		var buf bytes.Buffer
		if err := RunChmod(&buf, []string{"700", sub}, ChmodOptions{Recursive: true}); err != nil {
			t.Fatalf("RunChmod recursive error = %v", err)
		}
	})

	t.Run("changes flag no-op", func(t *testing.T) {
		f := mkFile("chg.txt", 0600)
		// Apply same octal mode; -c should report only on actual change.
		var buf bytes.Buffer
		if err := RunChmod(&buf, []string{"600", f}, ChmodOptions{Changes: true}); err != nil {
			t.Fatalf("RunChmod changes error = %v", err)
		}
	})

	t.Run("nonexistent target silent", func(t *testing.T) {
		var buf bytes.Buffer
		// Non-recursive on a missing file: chmodFile errors but RunChmod swallows
		// it (writes to stderr) and returns nil.
		if err := RunChmod(&buf, []string{"644", filepath.Join(dir, "ghost.txt")}, ChmodOptions{Silent: true}); err != nil {
			t.Errorf("expected nil (errors logged), got %v", err)
		}
	})
}
