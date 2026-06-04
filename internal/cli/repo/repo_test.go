package repo

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

func TestResolvePath(t *testing.T) {
	tests := []struct {
		name    string
		target  string
		wantErr bool
	}{
		{"empty uses cwd", "", false},
		{"dot uses cwd", ".", false},
		{"nonexistent", "/nonexistent/path/xyz", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := resolvePath(tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolvePath(%q) error = %v, wantErr %v", tt.target, err, tt.wantErr)
			}
		})
	}
}

func TestIsRemote(t *testing.T) {
	tests := []struct {
		target string
		want   bool
	}{
		{"https://github.com/owner/repo", true},
		{"git@github.com:owner/repo.git", true},
		{"github.com/owner/repo", true},
		{"owner/repo", true},
		{".", false},
		{"/path/to/dir", false},
		{"./relative", false},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			got := isRemote(tt.target)
			if got != tt.want {
				t.Errorf("isRemote(%q) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func TestNormalizeRemote(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"owner/repo", "owner/repo"},
		{"github.com/owner/repo", "owner/repo"},
		{"github.com/owner/repo.git", "owner/repo"},
		{"https://github.com/owner/repo", "https://github.com/owner/repo"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeRemote(tt.input)
			if got != tt.want {
				t.Errorf("normalizeRemote(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestCloneToTempRejectsLeadingDash verifies that a normalized clone target
// beginning with "-" is rejected as invalid input (CWE-88 argument injection)
// before any git/gh process is invoked. Without the guard, "-foo" would be
// passed to `git clone`/`gh repo clone` as a flag rather than a URL.
func TestCloneToTempRejectsLeadingDash(t *testing.T) {
	malicious := []string{
		"-malicious",
		"--upload-pack=touch /tmp/pwn",
		"-oProxyCommand=evil",
	}

	for _, target := range malicious {
		t.Run(target, func(t *testing.T) {
			dir, err := cloneToTemp(target)
			if dir != "" {
				_ = os.RemoveAll(dir)
				t.Errorf("cloneToTemp(%q) returned non-empty dir %q; expected rejection", target, dir)
			}

			if !errors.Is(err, cmderr.ErrInvalidInput) {
				t.Errorf("cloneToTemp(%q) error = %v, want cmderr.ErrInvalidInput", target, err)
			}
		})
	}
}

// TestGitCloneArgsTerminator verifies the git clone argv places a "--" option
// terminator immediately before the URL so a hostile URL cannot be parsed as a
// flag, and that a leading-dash URL is rejected.
func TestGitCloneArgsTerminator(t *testing.T) {
	args, err := gitCloneArgs("owner/repo", "/tmp/dest")
	if err != nil {
		t.Fatalf("gitCloneArgs returned unexpected error: %v", err)
	}

	// "--" must appear, and the URL + dest must come after it.
	idx := -1
	for i, a := range args {
		if a == "--" {
			idx = i
			break
		}
	}

	if idx == -1 {
		t.Fatalf("gitCloneArgs(%v) missing -- terminator", args)
	}

	if idx >= len(args)-2 || args[idx+2] != "/tmp/dest" {
		t.Errorf("gitCloneArgs = %v; want URL then dest after --", args)
	}

	if _, err := gitCloneArgs("-evil", "/tmp/dest"); !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Errorf("gitCloneArgs(-evil) error = %v, want cmderr.ErrInvalidInput", err)
	}
}

// TestGhCloneArgsTerminator verifies the gh repo clone argv places a "--"
// terminator immediately before the URL and rejects leading-dash URLs.
func TestGhCloneArgsTerminator(t *testing.T) {
	args, err := ghCloneArgs("owner/repo", "/tmp/dest")
	if err != nil {
		t.Fatalf("ghCloneArgs returned unexpected error: %v", err)
	}

	idx := -1
	for i, a := range args {
		if a == "--" {
			idx = i
			break
		}
	}

	if idx == -1 {
		t.Fatalf("ghCloneArgs(%v) missing -- terminator", args)
	}

	if idx >= len(args)-1 || args[idx+1] != "owner/repo" {
		t.Errorf("ghCloneArgs = %v; want URL immediately after --", args)
	}

	if _, err := ghCloneArgs("-evil", "/tmp/dest"); !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Errorf("ghCloneArgs(-evil) error = %v, want cmderr.ErrInvalidInput", err)
	}
}

func TestSectionFiltering(t *testing.T) {
	tests := []struct {
		sections string
		section  string
		want     bool
	}{
		{"", "tree", true},
		{"tree,deps", "tree", true},
		{"tree,deps", "deps", true},
		{"tree,deps", "git", false},
		{"api", "api", true},
		{"api", "tree", false},
	}

	for _, tt := range tests {
		t.Run(tt.sections+"/"+tt.section, func(t *testing.T) {
			opts := Options{Sections: tt.sections}
			got := opts.wantSection(tt.section)
			if got != tt.want {
				t.Errorf("wantSection(%q) with sections=%q = %v, want %v", tt.section, tt.sections, got, tt.want)
			}
		})
	}
}

func TestBuildTree(t *testing.T) {
	dir := t.TempDir()

	// Create a small directory structure
	_ = os.MkdirAll(filepath.Join(dir, "cmd", "app"), 0o755)
	_ = os.MkdirAll(filepath.Join(dir, "pkg", "util"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0o644)

	tree := buildTree(dir, 2)

	if tree == "" {
		t.Fatal("expected non-empty tree")
	}

	if !strings.Contains(tree, "cmd/") {
		t.Error("expected cmd/ in tree")
	}

	if !strings.Contains(tree, "main.go") {
		t.Error("expected main.go in tree")
	}
}

func TestCollectKeyFiles(t *testing.T) {
	dir := t.TempDir()

	_ = os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test Project\nLine 2"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.22"), 0o644)

	files := collectKeyFiles(dir, false)

	if len(files) != 2 {
		t.Fatalf("expected 2 key files, got %d", len(files))
	}

	names := make(map[string]bool)
	for _, f := range files {
		names[f.Name] = true
	}

	if !names["README.md"] {
		t.Error("expected README.md in key files")
	}

	if !names["go.mod"] {
		t.Error("expected go.mod in key files")
	}
}

func TestCollectKeyFilesCompact(t *testing.T) {
	dir := t.TempDir()

	// Create a long README
	var lines []string
	for i := range 200 {
		lines = append(lines, "line "+string(rune('0'+i%10)))
	}

	_ = os.WriteFile(filepath.Join(dir, "README.md"), []byte(strings.Join(lines, "\n")), 0o644)

	files := collectKeyFiles(dir, true)

	if len(files) != 1 {
		t.Fatalf("expected 1 key file, got %d", len(files))
	}

	if !strings.Contains(files[0].Content, "...(truncated)") {
		t.Error("expected truncated content in compact mode")
	}
}

func TestFindEntryPoints(t *testing.T) {
	dir := t.TempDir()

	_ = os.MkdirAll(filepath.Join(dir, "cmd", "myapp"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "cmd", "myapp", "main.go"), []byte("package main"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0o644)

	entries := findEntryPoints(dir)

	if len(entries) != 2 {
		t.Fatalf("expected 2 entry points, got %d", len(entries))
	}
}

func TestInferArchitecture(t *testing.T) {
	tests := []struct {
		name    string
		dirs    []string
		pattern string
	}{
		{"hexagonal", []string{"cmd", "internal", "pkg"}, "hexagonal/clean"},
		{"standard Go", []string{"cmd", "internal"}, "standard Go layout"},
		{"src-based", []string{"src"}, "src-based"},
		{"flat", []string{}, "flat"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			for _, d := range tt.dirs {
				_ = os.MkdirAll(filepath.Join(dir, d), 0o755)
			}

			arch := inferArchitecture(dir)
			if arch.Pattern != tt.pattern {
				t.Errorf("got pattern %q, want %q", arch.Pattern, tt.pattern)
			}
		})
	}
}

func TestDetectTestPatterns(t *testing.T) {
	dir := t.TempDir()

	_ = os.WriteFile(filepath.Join(dir, "foo_test.go"), []byte("package foo"), 0o644)
	_ = os.MkdirAll(filepath.Join(dir, "testdata"), 0o755)
	_ = os.MkdirAll(filepath.Join(dir, "golden"), 0o755)

	patterns := detectTestPatterns(dir)

	pset := make(map[string]bool)
	for _, p := range patterns {
		pset[p] = true
	}

	if !pset["table-driven (Go)"] {
		t.Error("expected table-driven (Go) pattern")
	}

	if !pset["fixtures/testdata"] {
		t.Error("expected fixtures/testdata pattern")
	}

	if !pset["golden master"] {
		t.Error("expected golden master pattern")
	}
}

func TestRunAnalyzeLocal(t *testing.T) {
	dir := t.TempDir()

	// Set up a minimal project
	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.22"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test"), 0o644)
	_ = os.MkdirAll(filepath.Join(dir, "cmd", "app"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "cmd", "app", "main.go"), []byte("package main"), 0o644)

	var buf bytes.Buffer

	err := RunAnalyze(&buf, []string{dir}, Options{})
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	if !strings.Contains(output, "# Repository Context:") {
		t.Error("expected markdown header in output")
	}

	if !strings.Contains(output, "Go") {
		t.Error("expected Go language in output")
	}
}

func TestRunAnalyzeJSON(t *testing.T) {
	dir := t.TempDir()

	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.22"), 0o644)

	var buf bytes.Buffer

	err := RunAnalyze(&buf, []string{dir}, Options{JSON: true})
	if err != nil {
		t.Fatal(err)
	}

	var report AnalyzeReport
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if report.Name == "" {
		t.Error("expected non-empty name")
	}
}

func TestRunAnalyzeCompact(t *testing.T) {
	dir := t.TempDir()

	_ = os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test\n"+strings.Repeat("line\n", 100)), 0o644)

	var buf bytes.Buffer

	err := RunAnalyze(&buf, []string{dir}, Options{Compact: true})
	if err != nil {
		t.Fatal(err)
	}

	// Should still produce valid output
	if buf.Len() == 0 {
		t.Error("expected non-empty output")
	}
}

func TestRunAnalyzeSections(t *testing.T) {
	dir := t.TempDir()

	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.22"), 0o644)

	var buf bytes.Buffer

	err := RunAnalyze(&buf, []string{dir}, Options{Sections: "tree"})
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	if !strings.Contains(output, "Directory Tree") {
		t.Error("expected tree section in output")
	}
}

func TestRunAnalyzeOutputFile(t *testing.T) {
	dir := t.TempDir()

	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.22"), 0o644)

	outFile := filepath.Join(dir, "context.md")

	var buf bytes.Buffer

	err := RunAnalyze(&buf, []string{dir}, Options{Output: outFile})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "# Repository Context:") {
		t.Error("expected markdown header in output file")
	}
}

func TestFindConfigFiles(t *testing.T) {
	dir := t.TempDir()

	_ = os.WriteFile(filepath.Join(dir, ".env.example"), []byte("KEY=val"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte("version: 3"), 0o644)

	found := findConfigFiles(dir)

	if len(found) != 2 {
		t.Errorf("expected 2 config files, got %d: %v", len(found), found)
	}
}
