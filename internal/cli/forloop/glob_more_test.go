package forloop

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// setupGlobTree builds a small directory tree under t.TempDir and returns its root.
func setupGlobTree(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	files := []string{
		"a.txt",
		"b.txt",
		"c.go",
		filepath.Join("sub", "d.txt"),
		filepath.Join("sub", "e.go"),
		filepath.Join("sub", "deep", "f.txt"),
	}

	for _, f := range files {
		full := filepath.Join(root, f)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	return root
}

// TestFilepathGlobStandard covers the non-recursive glob branch.
func TestFilepathGlobStandard(t *testing.T) {
	root := setupGlobTree(t)

	matches, err := filepathGlob(filepath.Join(root, "*.txt"))
	if err != nil {
		t.Fatalf("filepathGlob error = %v", err)
	}

	if len(matches) != 2 {
		t.Errorf("got %d matches, want 2: %v", len(matches), matches)
	}
}

// TestFilepathGlobRecursive covers globRecursive via the ** pattern (with and
// without a suffix).
func TestFilepathGlobRecursive(t *testing.T) {
	root := setupGlobTree(t)

	t.Run("suffix txt", func(t *testing.T) {
		matches, err := filepathGlob(filepath.Join(root, "**", "*.txt"))
		if err != nil {
			t.Fatalf("filepathGlob ** error = %v", err)
		}

		var txt int
		for _, m := range matches {
			if strings.HasSuffix(m, ".txt") {
				txt++
			}
		}

		if txt < 3 {
			t.Errorf("expected >=3 .txt matches across tree, got %d: %v", txt, matches)
		}
	})

	t.Run("no suffix walks everything", func(t *testing.T) {
		matches, err := filepathGlob(filepath.Join(root, "**"))
		if err != nil {
			t.Fatalf("filepathGlob ** (no suffix) error = %v", err)
		}

		if len(matches) == 0 {
			t.Error("expected matches walking the whole tree")
		}
	})
}

// TestGlobFiles covers the globFiles wrapper.
func TestGlobFiles(t *testing.T) {
	root := setupGlobTree(t)

	matches, err := globFiles(filepath.Join(root, "*.go"))
	if err != nil {
		t.Fatalf("globFiles error = %v", err)
	}

	if len(matches) != 1 {
		t.Errorf("got %d matches, want 1: %v", len(matches), matches)
	}
}

// TestRunGlobDryRun drives RunGlob through the dry-run path (no exec) so it is
// fully offline.
func TestRunGlobDryRun(t *testing.T) {
	root := setupGlobTree(t)

	var buf bytes.Buffer
	err := RunGlob(&buf, filepath.Join(root, "*.txt"), "echo ${file}", Options{DryRun: true})
	if err != nil {
		t.Fatalf("RunGlob dry-run error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	sort.Strings(lines)

	if len(lines) != 2 {
		t.Errorf("expected 2 dry-run lines, got %d: %q", len(lines), buf.String())
	}

	for _, l := range lines {
		if !strings.HasPrefix(l, "echo ") || !strings.HasSuffix(l, ".txt") {
			t.Errorf("unexpected dry-run line: %q", l)
		}
	}
}

// TestRunGlobNoMatches covers RunGlob when the pattern matches nothing.
func TestRunGlobNoMatches(t *testing.T) {
	root := setupGlobTree(t)

	var buf bytes.Buffer
	err := RunGlob(&buf, filepath.Join(root, "*.none"), "echo ${file}", Options{DryRun: true})
	if err != nil {
		t.Fatalf("RunGlob no-match error = %v", err)
	}

	if strings.TrimSpace(buf.String()) != "" {
		t.Errorf("expected no output for zero matches, got %q", buf.String())
	}
}

// TestRunGlobBadPattern covers the invalid-pattern error branch.
func TestRunGlobBadPattern(t *testing.T) {
	var buf bytes.Buffer
	// An unterminated character class is a malformed glob pattern.
	err := RunGlob(&buf, "[", "echo ${file}", Options{DryRun: true})
	if err == nil {
		t.Error("expected error for malformed glob pattern")
	}
}
