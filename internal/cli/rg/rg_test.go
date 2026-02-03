package rg

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	// Create temp directory with test files
	dir := t.TempDir()

	// Create test files
	files := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`,
		"lib.go": `package main

func helper() string {
	return "helper function"
}
`,
		"README.md": `# Test Project

This is a test project.
Hello from markdown!
`,
		"config.json": `{
	"name": "test",
	"hello": "world"
}
`,
		"subdir/nested.go": `package subdir

func Nested() {
	println("nested hello")
}
`,
	}

	for name, content := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name       string
		pattern    string
		paths      []string
		opts       Options
		wantMatch  bool
		wantInOut  []string
		wantNotOut []string
	}{
		{
			name:      "simple pattern",
			pattern:   "Hello",
			paths:     []string{dir},
			opts:      Options{LineNumber: true},
			wantMatch: true,
			wantInOut: []string{"Hello"},
		},
		{
			name:      "case insensitive",
			pattern:   "hello",
			paths:     []string{dir},
			opts:      Options{IgnoreCase: true, LineNumber: true},
			wantMatch: true,
			wantInOut: []string{"Hello", "hello"},
		},
		{
			name:      "file type filter go",
			pattern:   "func",
			paths:     []string{dir},
			opts:      Options{Types: []string{"go"}, LineNumber: true},
			wantMatch: true,
			wantInOut: []string{"func main", "func helper", "func Nested"},
		},
		{
			name:       "file type filter md",
			pattern:    "Hello",
			paths:      []string{dir},
			opts:       Options{Types: []string{"md"}, LineNumber: true},
			wantMatch:  true,
			wantInOut:  []string{"Hello from markdown"},
			wantNotOut: []string{"Hello, World"},
		},
		{
			name:      "count mode",
			pattern:   "hello",
			paths:     []string{dir},
			opts:      Options{Count: true, IgnoreCase: true},
			wantMatch: true,
			wantInOut: []string{":"},
		},
		{
			name:      "files with match",
			pattern:   "func",
			paths:     []string{dir},
			opts:      Options{FilesWithMatch: true},
			wantMatch: true,
			wantInOut: []string{".go"},
		},
		{
			name:      "word regexp",
			pattern:   "main",
			paths:     []string{dir},
			opts:      Options{WordRegexp: true, LineNumber: true},
			wantMatch: true,
			wantInOut: []string{"main"},
		},
		{
			name:       "invert match",
			pattern:    "func",
			paths:      []string{filepath.Join(dir, "main.go")},
			opts:       Options{InvertMatch: true, LineNumber: true},
			wantMatch:  true,
			wantInOut:  []string{"package", "import"},
			wantNotOut: []string{"func main"},
		},
		{
			name:      "fixed string",
			pattern:   "fmt.Println",
			paths:     []string{dir},
			opts:      Options{Fixed: true, LineNumber: true},
			wantMatch: true,
			wantInOut: []string{"fmt.Println"},
		},
		{
			name:      "no pattern error",
			pattern:   "",
			paths:     []string{dir},
			opts:      Options{},
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := Run(&buf, tt.pattern, tt.paths, tt.opts)

			if tt.pattern == "" {
				if err == nil {
					t.Error("expected error for empty pattern")
				}

				return
			}

			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}

			output := buf.String()

			if tt.wantMatch && output == "" && !tt.opts.JSON {
				t.Error("expected matches but got none")
			}

			for _, want := range tt.wantInOut {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got:\n%s", want, output)
				}
			}

			for _, notWant := range tt.wantNotOut {
				if strings.Contains(output, notWant) {
					t.Errorf("output should NOT contain %q, got:\n%s", notWant, output)
				}
			}
		})
	}
}

func TestMatchesFileType(t *testing.T) {
	tests := []struct {
		path    string
		include []string
		exclude []string
		want    bool
	}{
		{"main.go", []string{"go"}, nil, true},
		{"main.go", []string{"js"}, nil, false},
		{"app.js", []string{"js"}, nil, true},
		{"app.ts", []string{"ts"}, nil, true},
		{"main.go", nil, []string{"go"}, false},
		{"app.js", nil, []string{"go"}, true},
		{"main.go", nil, nil, true},
		{"README.md", []string{"md"}, nil, true},
		{"config.json", []string{"json"}, nil, true},
		{"Dockerfile", []string{"dockerfile"}, nil, true},
		{"Makefile", []string{"make"}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := matchesFileType(tt.path, tt.include, tt.exclude)
			if got != tt.want {
				t.Errorf("matchesFileType(%q, %v, %v) = %v, want %v",
					tt.path, tt.include, tt.exclude, got, tt.want)
			}
		})
	}
}

func TestMatchesGlob(t *testing.T) {
	tests := []struct {
		path     string
		patterns []string
		want     bool
	}{
		{"main.go", []string{"*.go"}, true},
		{"main.go", []string{"*.js"}, false},
		{"main.go", nil, true},
		{"main_test.go", []string{"*_test.go"}, true},
		{"main.go", []string{"!*.go"}, false},
		{"main.js", []string{"!*.go"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := matchesGlob(tt.path, tt.patterns)
			if got != tt.want {
				t.Errorf("matchesGlob(%q, %v) = %v, want %v",
					tt.path, tt.patterns, got, tt.want)
			}
		})
	}
}

func TestIsIgnored(t *testing.T) {
	patterns := []string{".git", "node_modules", "*.log", "vendor"}

	tests := []struct {
		path string
		want bool
	}{
		{".git", true},
		{"node_modules", true},
		{"src/main.go", false},
		{"debug.log", true},
		{"vendor/lib.go", true},
		{"src/vendor/lib.go", true},
		{"main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isIgnored(tt.path, patterns)
			if got != tt.want {
				t.Errorf("isIgnored(%q) = %v, want %v", tt.path, got, tt.want)
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
		{"text", []byte("Hello, World!"), false},
		{"binary with null", []byte{0x48, 0x65, 0x00, 0x6c}, true},
		{"empty", []byte{}, false},
		{"utf8", []byte("Hello, 世界!"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBinary(tt.data)
			if got != tt.want {
				t.Errorf("isBinary() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJSONOutput(t *testing.T) {
	dir := t.TempDir()

	// Create test file
	testFile := filepath.Join(dir, "test.go")

	content := `package main

func hello() {
	println("hello world")
}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	opts := Options{JSON: true}

	err := Run(&buf, "hello", []string{dir}, opts)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"path"`) {
		t.Error("JSON output should contain path field")
	}

	if !strings.Contains(output, `"total_matches"`) {
		t.Error("JSON output should contain total_matches field")
	}
}

func TestContextLines(t *testing.T) {
	dir := t.TempDir()

	// Create test file with multiple lines
	testFile := filepath.Join(dir, "test.txt")

	content := `line 1
line 2
line 3
MATCH HERE
line 5
line 6
line 7
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	opts := Options{
		Context:    1,
		LineNumber: true,
	}

	err := Run(&buf, "MATCH", []string{testFile}, opts)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	output := buf.String()

	// Should contain the match
	if !strings.Contains(output, "MATCH HERE") {
		t.Error("output should contain the match")
	}

	// Should contain context lines
	if !strings.Contains(output, "line 3") {
		t.Error("output should contain before context (line 3)")
	}

	if !strings.Contains(output, "line 5") {
		t.Error("output should contain after context (line 5)")
	}
}
