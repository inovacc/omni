package rg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePattern(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantNeg   bool
		wantDir   bool
		wantAnch  bool
		wantMatch []string
		wantNoM   []string
	}{
		{
			name:      "simple pattern",
			input:     "*.log",
			wantMatch: []string{"debug.log", "error.log"},
			wantNoM:   []string{"log", "main.go"},
		},
		{
			name:      "negation pattern",
			input:     "!important.log",
			wantNeg:   true,
			wantMatch: []string{"important.log"},
			wantNoM:   []string{"other.log"},
		},
		{
			name:      "directory only pattern",
			input:     "build/",
			wantDir:   true,
			wantAnch:  false, // build/ is not anchored, it matches any directory named build
			wantMatch: []string{"build"},
			wantNoM:   []string{"build.go"},
		},
		{
			name:      "anchored pattern with leading slash",
			input:     "/root.txt",
			wantAnch:  true,
			wantMatch: []string{"root.txt"},
			wantNoM:   []string{"sub/root.txt"},
		},
		{
			name:      "anchored pattern with internal slash",
			input:     "foo/bar",
			wantAnch:  true,
			wantMatch: []string{"foo/bar"},
			wantNoM:   []string{"baz/foo/bar"},
		},
		{
			name:      "double glob at start",
			input:     "**/test.go",
			wantAnch:  true, // Contains **/ which makes it match anywhere in path
			wantMatch: []string{"test.go", "pkg/test.go", "a/b/c/test.go"},
			wantNoM:   []string{"test.py"},
		},
		{
			name:      "double glob in middle",
			input:     "src/**/test.go",
			wantAnch:  true,
			wantMatch: []string{"src/test.go", "src/pkg/test.go"},
			wantNoM:   []string{"lib/test.go"},
		},
		{
			name:      "double glob at end",
			input:     "logs/**",
			wantAnch:  true,
			wantMatch: []string{"logs/a", "logs/a/b"},
		},
		{
			name:      "question mark wildcard",
			input:     "file?.txt",
			wantMatch: []string{"file1.txt", "fileA.txt"},
			wantNoM:   []string{"file10.txt", "file.txt"},
		},
		{
			name:      "character class",
			input:     "file[0-9].txt",
			wantMatch: []string{"file0.txt", "file5.txt"},
			wantNoM:   []string{"fileA.txt"},
		},
		{
			name:      "negated character class",
			input:     "file[!0-9].txt",
			wantMatch: []string{"fileA.txt", "filez.txt"},
			wantNoM:   []string{"file5.txt"},
		},
		{
			name:      "escaped special chars",
			input:     "\\!important",
			wantMatch: []string{"!important"},
			wantNoM:   []string{"important"},
		},
		{
			name:      "escaped hash",
			input:     "\\#readme",
			wantMatch: []string{"#readme"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parsePattern(tt.input)
			if p == nil {
				t.Fatal("parsePattern returned nil")
			}

			if p.Negation != tt.wantNeg {
				t.Errorf("Negation = %v, want %v", p.Negation, tt.wantNeg)
			}

			if p.DirOnly != tt.wantDir {
				t.Errorf("DirOnly = %v, want %v", p.DirOnly, tt.wantDir)
			}

			if p.Anchored != tt.wantAnch {
				t.Errorf("Anchored = %v, want %v", p.Anchored, tt.wantAnch)
			}

			for _, path := range tt.wantMatch {
				// For directory-only patterns, test with isDir=true
				isDir := tt.wantDir
				if !p.matchPath(path, path, isDir) {
					t.Errorf("pattern should match %q", path)
				}
			}

			for _, path := range tt.wantNoM {
				if p.matchPath(path, path, false) {
					t.Errorf("pattern should NOT match %q", path)
				}
			}
		})
	}
}

func TestGitignoreNegation(t *testing.T) {
	// Create temp directory with gitignore
	dir := t.TempDir()

	gitignore := `# Ignore all log files
*.log
# But keep important.log
!important.log
# Ignore build directory
build/
# But keep build/release
!build/release/
`

	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignore), 0644); err != nil {
		t.Fatal(err)
	}

	gs := NewGitignoreSet(dir)

	tests := []struct {
		path   string
		isDir  bool
		expect MatchResult
	}{
		{"debug.log", false, Ignore},
		{"error.log", false, Ignore},
		{"important.log", false, Include}, // Negated
		{"main.go", false, NoMatch},
		{"build", true, Ignore},
		{"build/output", true, NoMatch},  // build/ only matches the dir itself, not its contents
		{"build/release", true, Include}, // Negated explicitly
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := gs.Match(tt.path, tt.isDir)
			if result != tt.expect {
				t.Errorf("Match(%q, %v) = %v, want %v", tt.path, tt.isDir, result, tt.expect)
			}
		})
	}
}

func TestGitignoreDirectoryOnly(t *testing.T) {
	dir := t.TempDir()

	gitignore := `# Ignore directories named "cache"
cache/
# Ignore all tmp files
*.tmp
`

	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignore), 0644); err != nil {
		t.Fatal(err)
	}

	gs := NewGitignoreSet(dir)

	tests := []struct {
		path   string
		isDir  bool
		expect MatchResult
	}{
		{"cache", true, Ignore},           // Directory - should be ignored
		{"cache", false, NoMatch},         // File named cache - should NOT be ignored
		{"data.tmp", false, Ignore},       // File matching *.tmp
		{"backup.tmp", false, Ignore},     // File matching *.tmp
		{"tmpdir", true, NoMatch},         // Directory not matching pattern
		{"src/cache", true, Ignore},       // Nested cache directory
		{"src/cache.txt", false, NoMatch}, // File with cache in name
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := gs.Match(tt.path, tt.isDir)
			if result != tt.expect {
				t.Errorf("Match(%q, isDir=%v) = %v, want %v", tt.path, tt.isDir, result, tt.expect)
			}
		})
	}
}

func TestGitignoreDoubleGlob(t *testing.T) {
	dir := t.TempDir()

	gitignore := `# Ignore all test files anywhere
**/test_*.py
# Ignore vendor directories
vendor/**/
# Ignore all .bak files in docs
docs/**/*.bak
`

	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignore), 0644); err != nil {
		t.Fatal(err)
	}

	gs := NewGitignoreSet(dir)

	tests := []struct {
		path  string
		isDir bool
		want  bool // true = should ignore
	}{
		{"test_main.py", false, true},
		{"src/test_main.py", false, true},
		{"a/b/c/test_main.py", false, true},
		{"main.py", false, false},
		{"docs/file.bak", false, true},
		{"docs/sub/file.bak", false, true},
		{"other/file.bak", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := gs.ShouldIgnore(tt.path, tt.isDir)
			if result != tt.want {
				t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.path, result, tt.want)
			}
		})
	}
}

func TestIgnoreFile(t *testing.T) {
	// Test .ignore file (ripgrep-specific)
	dir := t.TempDir()

	// Create .ignore file
	ignoreContent := `# ripgrep-specific ignore
*.generated.go
testdata/
`

	if err := os.WriteFile(filepath.Join(dir, ".ignore"), []byte(ignoreContent), 0644); err != nil {
		t.Fatal(err)
	}

	gs := NewGitignoreSet(dir)

	tests := []struct {
		path  string
		isDir bool
		want  bool
	}{
		{"main.generated.go", false, true},
		{"main.go", false, false},
		{"testdata", true, true},
		{"testdata/file.txt", false, false}, // Only dir is ignored
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := gs.ShouldIgnore(tt.path, tt.isDir)
			if result != tt.want {
				t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.path, result, tt.want)
			}
		})
	}
}

func TestGitignoreHierarchy(t *testing.T) {
	// Create a directory hierarchy with multiple gitignore files
	root := t.TempDir()

	// Root .gitignore
	rootIgnore := `*.log
build/
`
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte(rootIgnore), 0644); err != nil {
		t.Fatal(err)
	}

	// Create subdirectory with its own .gitignore
	subDir := filepath.Join(root, "src")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Subdirectory .gitignore (overrides root for *.log)
	subIgnore := `!debug.log
*.tmp
`
	if err := os.WriteFile(filepath.Join(subDir, ".gitignore"), []byte(subIgnore), 0644); err != nil {
		t.Fatal(err)
	}

	gs := NewGitignoreSet(subDir)

	tests := []struct {
		path  string
		isDir bool
		want  MatchResult
	}{
		{"error.log", false, Ignore},  // From root
		{"debug.log", false, Include}, // Negated in subdirectory
		{"data.tmp", false, Ignore},   // From subdirectory
		{"build", true, Ignore},       // From root
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := gs.Match(tt.path, tt.isDir)
			if result != tt.want {
				t.Errorf("Match(%q) = %v, want %v", tt.path, result, tt.want)
			}
		})
	}
}

func TestCommonIgnores(t *testing.T) {
	gs := &GitignoreSet{gitignores: make([]*Gitignore, 0)}
	gs.AddCommonIgnores()

	tests := []struct {
		path string
		want bool
	}{
		{".git", true},
		{"node_modules", true},
		{"__pycache__", true},
		{".idea", true},
		{".vscode", true},
		{"src", false},
		{"main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := gs.ShouldIgnore(tt.path, true)
			if result != tt.want {
				t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.path, result, tt.want)
			}
		})
	}
}

func TestPatternToRegex(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		match   bool
	}{
		{"*.txt", "file.txt", true},
		{"*.txt", "file.go", false},
		{"test*", "testing", true},
		{"test*", "contest", false},
		{"file?.go", "file1.go", true},
		{"file?.go", "file12.go", false},
		{"[abc].txt", "a.txt", true},
		{"[abc].txt", "d.txt", false},
		{"[!abc].txt", "d.txt", true},
		{"[!abc].txt", "a.txt", false},
		{"foo.bar", "foo.bar", true},
		{"foo\\.bar", "foo.bar", true}, // Escaped dot
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.input, func(t *testing.T) {
			p := parsePattern(tt.pattern)
			if p == nil {
				t.Fatal("parsePattern returned nil")
			}

			got := p.Regex.MatchString(tt.input)
			if got != tt.match {
				t.Errorf("pattern %q matching %q: got %v, want %v (regex: %s)",
					tt.pattern, tt.input, got, tt.match, p.Regex.String())
			}
		})
	}
}
