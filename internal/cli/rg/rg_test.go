package rg

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	outputpkg "github.com/inovacc/omni/internal/cli/output"
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

			if tt.wantMatch && output == "" && tt.opts.OutputFormat != outputpkg.FormatJSON {
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

func TestLiteralSearch(t *testing.T) {
	dir := t.TempDir()

	// Create test file with regex special characters
	content := `package main

func test() {
	if x := foo(); x != nil {
		println("match: foo()")
	}
}
`

	testFile := filepath.Join(dir, "test.go")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		pattern string
		opts    Options
		want    bool
	}{
		{
			name:    "literal parentheses",
			pattern: "foo()",
			opts:    Options{Fixed: true},
			want:    true,
		},
		{
			name:    "literal with special chars",
			pattern: "x != nil",
			opts:    Options{Fixed: true},
			want:    true,
		},
		{
			name:    "literal case insensitive",
			pattern: "FOO()",
			opts:    Options{Fixed: true, IgnoreCase: true},
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := Run(&buf, tt.pattern, []string{dir}, tt.opts)
			if err != nil {
				t.Fatal(err)
			}

			hasMatch := buf.Len() > 0
			if hasMatch != tt.want {
				t.Errorf("Run() hasMatch = %v, want %v, output: %s", hasMatch, tt.want, buf.String())
			}
		})
	}
}

func TestParallelSearch(t *testing.T) {
	dir := t.TempDir()

	// Create multiple test files
	for i := range 10 {
		content := "package main\n\nfunc test() {\n\tprintln(\"hello\")\n}\n"
		path := filepath.Join(dir, "file"+string(rune('0'+i))+".go")

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Run with single thread
	var buf1 bytes.Buffer

	err := Run(&buf1, "hello", []string{dir}, Options{Threads: 1, FilesWithMatch: true})
	if err != nil {
		t.Fatal(err)
	}

	// Run with multiple threads
	var buf2 bytes.Buffer

	err = Run(&buf2, "hello", []string{dir}, Options{Threads: 4, FilesWithMatch: true})
	if err != nil {
		t.Fatal(err)
	}

	// Both should have the same number of matches
	count1 := strings.Count(buf1.String(), ".go")
	count2 := strings.Count(buf2.String(), ".go")

	if count1 != count2 {
		t.Errorf("parallel search gave different results: single=%d, parallel=%d", count1, count2)
	}

	if count1 != 10 {
		t.Errorf("expected 10 matches, got %d", count1)
	}
}

func TestJSONStreamOutput(t *testing.T) {
	dir := t.TempDir()

	content := `package main

func hello() {
	println("hello world")
}
`
	if err := os.WriteFile(filepath.Join(dir, "test.go"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	opts := Options{JSONStream: true}

	err := Run(&buf, "hello", []string{dir}, opts)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have: begin, match(es), end, summary
	if len(lines) < 4 {
		t.Errorf("expected at least 4 NDJSON lines, got %d: %s", len(lines), output)
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		if line == "" {
			continue
		}

		var msg StreamMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			t.Errorf("line %d is not valid JSON: %v", i, err)
		}
	}

	// Check for expected message types
	if !strings.Contains(output, `"type":"begin"`) {
		t.Error("missing begin message")
	}

	if !strings.Contains(output, `"type":"match"`) {
		t.Error("missing match message")
	}

	if !strings.Contains(output, `"type":"end"`) {
		t.Error("missing end message")
	}

	if !strings.Contains(output, `"type":"summary"`) {
		t.Error("missing summary message")
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

	opts := Options{OutputFormat: outputpkg.FormatJSON}

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

func TestGitignoreIntegration(t *testing.T) {
	dir := t.TempDir()

	// Create .gitignore
	gitignore := `*.log
!important.log
build/
`
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignore), 0644); err != nil {
		t.Fatal(err)
	}

	// Create test files
	files := map[string]string{
		"main.go":       "package main\n\nfunc main() { println(\"hello\") }",
		"debug.log":     "DEBUG: hello world",
		"important.log": "IMPORTANT: hello critical",
		"build/out.go":  "package build\n\nfunc Build() { println(\"hello\") }",
		"src/lib.go":    "package lib\n\nfunc Lib() { println(\"hello\") }",
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

	// Search with gitignore enabled
	var buf bytes.Buffer

	opts := Options{FilesWithMatch: true}

	err := Run(&buf, "hello", []string{dir}, opts)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	// Should include main.go and src/lib.go
	if !strings.Contains(output, "main.go") {
		t.Error("should include main.go")
	}

	if !strings.Contains(output, "lib.go") {
		t.Error("should include src/lib.go")
	}

	// Should NOT include debug.log (ignored)
	if strings.Contains(output, "debug.log") {
		t.Error("should NOT include debug.log (ignored by *.log)")
	}

	// SHOULD include important.log (negation pattern)
	if !strings.Contains(output, "important.log") {
		t.Error("should include important.log (negation pattern !important.log)")
	}

	// Should NOT include build/out.go (ignored directory)
	if strings.Contains(output, "build") {
		t.Error("should NOT include files in build/ (ignored directory)")
	}
}

func TestIgnoreFileSupport(t *testing.T) {
	dir := t.TempDir()

	// Create .ignore file (ripgrep-specific)
	ignoreContent := `*.generated.go
testdata/
`
	if err := os.WriteFile(filepath.Join(dir, ".ignore"), []byte(ignoreContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create test files
	files := map[string]string{
		"main.go":            "package main\n\nfunc main() { println(\"hello\") }",
		"types.generated.go": "package main\n\n// hello generated code",
		"testdata/test.go":   "package testdata\n\n// hello test data",
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

	var buf bytes.Buffer

	opts := Options{FilesWithMatch: true}

	err := Run(&buf, "hello", []string{dir}, opts)
	if err != nil {
		t.Fatal(err)
	}

	output := buf.String()

	// Should include main.go
	if !strings.Contains(output, "main.go") {
		t.Error("should include main.go")
	}

	// Should NOT include *.generated.go
	if strings.Contains(output, "generated") {
		t.Error("should NOT include types.generated.go")
	}

	// Should NOT include testdata/
	if strings.Contains(output, "testdata") {
		t.Error("should NOT include files in testdata/")
	}
}
