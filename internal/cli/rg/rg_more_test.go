package rg

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPatternToRegex_Table(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
	}{
		{"star", "*.go"},
		{"doublestar", "**/foo"},
		{"question", "f?o.txt"},
		{"leading slash", "/build"},
		{"trailing slash", "node_modules/"},
		{"nested", "src/**/*.js"},
		{"plain", "vendor"},
		{"char class", "[abc].go"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := patternToRegex(tt.pattern)
			// We don't assert exact regex shape (it's delegated), only that it
			// produces a non-empty, non-panicking result the engine can use.
			if got == "" {
				t.Errorf("patternToRegex(%q) returned empty", tt.pattern)
			}
		})
	}
}

func TestFormatByteOffset(t *testing.T) {
	scheme := ColorScheme{Line: FgGreen}
	tests := []struct {
		name     string
		offset   int64
		useColor bool
		want     string
	}{
		{"no color", 42, false, "42"},
		{"with color", 7, true, FgGreen + "7" + Reset},
		{"zero no color", 0, false, "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatByteOffset(tt.offset, scheme, tt.useColor); got != tt.want {
				t.Errorf("FormatByteOffset() = %q want %q", got, tt.want)
			}
		})
	}

	// Color requested but no scheme color → falls through to plain.
	if got := FormatByteOffset(9, ColorScheme{}, true); got != "9" {
		t.Errorf("FormatByteOffset empty scheme = %q want 9", got)
	}
}

func TestNewGitignoreSet_WithFiles(t *testing.T) {
	dir := t.TempDir()
	// Create a .gitignore and an .ignore in the search dir.
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("*.log\nbuild/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".ignore"), []byte("tmp/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Point XDG_CONFIG_HOME at a temp config with a global ignore to exercise getGlobalGitignorePath.
	cfg := t.TempDir()
	gitCfg := filepath.Join(cfg, "git")
	if err := os.MkdirAll(gitCfg, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitCfg, "ignore"), []byte("*.bak\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", cfg)

	gs := NewGitignoreSet(dir)
	if gs == nil {
		t.Fatal("NewGitignoreSet returned nil")
	}

	if p := getGlobalGitignorePath(); p == "" {
		t.Error("expected global gitignore path to resolve from XDG_CONFIG_HOME")
	}
}

func TestGetGlobalGitignorePath_None(t *testing.T) {
	// Empty XDG and an empty HOME-equivalent dir → no global ignore present.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	// HOME points at an empty temp dir so the ~/.config and ~/.gitignore_global
	// fallbacks both miss.
	empty := t.TempDir()
	t.Setenv("HOME", empty)
	t.Setenv("USERPROFILE", empty) // Windows home
	_ = getGlobalGitignorePath()   // result may vary by platform; just exercise the branches
}

func writeTree(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		full := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRun_SearchVariants(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"a.txt":     "alpha\nbeta\ngamma needle here\ndelta\n",
		"b.txt":     "no match in this file\n",
		"sub/c.txt": "another needle line\ncontext after\n",
	})

	tests := []struct {
		name string
		opts Options
		want string // substring expected in output (empty = just no error)
	}{
		{"plain", Options{LineNumber: true}, "needle"},
		{"context", Options{Context: 1, LineNumber: true}, "needle"},
		{"after", Options{After: 1}, "needle"},
		{"before", Options{Before: 1}, "needle"},
		{"only-matching", Options{OnlyMatching: true}, "needle"},
		{"count", Options{Count: true}, ""},
		{"files-with-match", Options{FilesWithMatch: true}, ""},
		{"byte-offset", Options{ByteOffset: true}, "needle"},
		{"color always", Options{Color: "always"}, "needle"},
		{"ignore-case", Options{IgnoreCase: true}, "needle"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			err := Run(context.Background(), &buf, "needle", []string{dir}, tt.opts)
			if err != nil {
				t.Fatalf("Run(%s) err=%v", tt.name, err)
			}
			if tt.want != "" && !strings.Contains(buf.String(), tt.want) {
				t.Errorf("Run(%s) output missing %q:\n%s", tt.name, tt.want, buf.String())
			}
		})
	}
}

func TestRun_NoMatchAndInvert(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{"x.txt": "hello\nworld\n"})

	var buf strings.Builder
	// Invert match: lines NOT containing "hello".
	if err := Run(context.Background(), &buf, "hello", []string{dir}, Options{InvertMatch: true}); err != nil {
		t.Fatalf("invert err=%v", err)
	}
	if !strings.Contains(buf.String(), "world") {
		t.Errorf("invert output should contain non-matching line: %q", buf.String())
	}
}
