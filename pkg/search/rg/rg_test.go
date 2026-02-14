package rg

import (
	"testing"
)

func TestParsePattern(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		negation bool
		dirOnly  bool
		anchored bool
	}{
		{"simple", "*.go", false, false, false},
		{"negation", "!*.go", true, false, false},
		{"directory", "build/", false, true, false},
		{"anchored slash prefix", "/build", false, false, true},
		{"anchored contains slash", "src/test", false, false, true},
		{"double glob", "**/test.go", false, false, true}, // contains / so anchored
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ParsePattern(tt.input)
			if p == nil {
				t.Fatal("ParsePattern() returned nil")
			}

			if p.Negation != tt.negation {
				t.Errorf("Negation = %v, want %v", p.Negation, tt.negation)
			}

			if p.DirOnly != tt.dirOnly {
				t.Errorf("DirOnly = %v, want %v", p.DirOnly, tt.dirOnly)
			}

			if p.Anchored != tt.anchored {
				t.Errorf("Anchored = %v, want %v", p.Anchored, tt.anchored)
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
		{"*.go", "main.go", true},
		{"*.go", "main.py", false},
		{"test?", "test1", true},
		{"test?", "test", false},
		{"[abc]", "a", true},
		{"[abc]", "d", false},
		{"**/*.go", "src/main.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"/"+tt.input, func(t *testing.T) {
			p := ParsePattern(tt.pattern)
			if p == nil {
				t.Fatal("ParsePattern() returned nil")
			}

			got := p.Regex.MatchString(tt.input)
			if got != tt.match {
				t.Errorf("MatchString(%q) = %v, want %v (regex: %s)", tt.input, got, tt.match, p.Regex.String())
			}
		})
	}
}

func TestParseGitignore(t *testing.T) {
	content := `# Comment
*.log
!important.log
build/
node_modules
`

	gi := ParseGitignore(content, "/project")
	if gi == nil {
		t.Fatal("ParseGitignore() returned nil")
	}

	if len(gi.Patterns) != 4 {
		t.Errorf("ParseGitignore() got %d patterns, want 4", len(gi.Patterns))
	}
}

func TestParseGitignoreEmpty(t *testing.T) {
	gi := ParseGitignore("", "/project")
	if gi != nil {
		t.Error("ParseGitignore() should return nil for empty content")
	}

	gi = ParseGitignore("# only comments\n# nothing else\n", "/project")
	if gi != nil {
		t.Error("ParseGitignore() should return nil for comments-only content")
	}
}

func TestGitignoreSetMatch(t *testing.T) {
	gs := NewGitignoreSet(".")
	gi := ParseGitignore("*.log\n!important.log\nbuild/\n", ".")
	gs.AddGitignore(gi)

	tests := []struct {
		path  string
		isDir bool
		want  MatchResult
	}{
		{"test.log", false, Ignore},
		{"important.log", false, Include},
		{"build", true, Ignore},
		{"build", false, NoMatch}, // build/ only matches dirs
		{"main.go", false, NoMatch},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := gs.Match(tt.path, tt.isDir)
			if got != tt.want {
				t.Errorf("Match(%q, %v) = %v, want %v", tt.path, tt.isDir, got, tt.want)
			}
		})
	}
}

func TestGitignoreSetShouldIgnore(t *testing.T) {
	gs := NewGitignoreSet(".")
	gi := ParseGitignore("*.log\n", ".")
	gs.AddGitignore(gi)

	if !gs.ShouldIgnore("test.log", false) {
		t.Error("ShouldIgnore() should return true for *.log")
	}

	if gs.ShouldIgnore("main.go", false) {
		t.Error("ShouldIgnore() should return false for *.go")
	}
}

func TestAddCommonIgnores(t *testing.T) {
	gs := NewGitignoreSet(".")
	gs.AddCommonIgnores()

	common := []string{".git", "node_modules", "__pycache__", ".idea", ".vscode"}
	for _, name := range common {
		if !gs.ShouldIgnore(name, true) {
			t.Errorf("ShouldIgnore(%q) = false, want true after AddCommonIgnores()", name)
		}
	}
}

func TestMatchesFileType(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		include []string
		exclude []string
		want    bool
	}{
		{"no filters", "main.go", nil, nil, true},
		{"include go", "main.go", []string{"go"}, nil, true},
		{"include go miss", "main.py", []string{"go"}, nil, false},
		{"exclude go", "main.go", nil, []string{"go"}, false},
		{"exclude go miss", "main.py", nil, []string{"go"}, true},
		{"include py", "test.py", []string{"py"}, nil, true},
		{"include json", "data.json", []string{"json"}, nil, true},
		{"Makefile", "Makefile", []string{"make"}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchesFileType(tt.path, tt.include, tt.exclude)
			if got != tt.want {
				t.Errorf("MatchesFileType(%q, %v, %v) = %v, want %v", tt.path, tt.include, tt.exclude, got, tt.want)
			}
		})
	}
}

func TestMatchesGlob(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		patterns []string
		want     bool
	}{
		{"no patterns", "main.go", nil, true},
		{"match", "main.go", []string{"*.go"}, true},
		{"no match", "main.py", []string{"*.go"}, false},
		{"negation only", "main.go", []string{"!*.py"}, true},
		{"negation excludes", "main.py", []string{"!*.py"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchesGlob(tt.path, tt.patterns)
			if got != tt.want {
				t.Errorf("MatchesGlob(%q, %v) = %v, want %v", tt.path, tt.patterns, got, tt.want)
			}
		})
	}
}

func TestIsBinary(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"text", []byte("hello world"), false},
		{"binary", []byte{0x00, 0x01, 0x02}, true},
		{"mixed", []byte("hello\x00world"), true},
		{"empty", []byte{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBinary(tt.data)
			if got != tt.want {
				t.Errorf("IsBinary() = %v, want %v", got, tt.want)
			}
		})
	}
}
