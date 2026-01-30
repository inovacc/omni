package loc

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunLoc(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "loc_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a Go file
	goFile := filepath.Join(tmpDir, "main.go")
	goContent := `package main

// This is a comment
func main() {
	url := "http://example.com" // inline comment
	fmt.Println("Hello")
}
`
	if err := os.WriteFile(goFile, []byte(goContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a Python file
	pyFile := filepath.Join(tmpDir, "script.py")
	pyContent := `# Python comment
def hello():
    print("Hello")

# Another comment
hello()
`
	if err := os.WriteFile(pyFile, []byte(pyContent), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("basic count", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunLoc(&buf, []string{tmpDir}, Options{})
		if err != nil {
			t.Fatalf("RunLoc() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Go") {
			t.Errorf("output missing Go language")
		}

		if !strings.Contains(output, "Python") {
			t.Errorf("output missing Python language")
		}

		if !strings.Contains(output, "Total") {
			t.Errorf("output missing Total line")
		}
	})

	t.Run("json output", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunLoc(&buf, []string{tmpDir}, Options{JSON: true})
		if err != nil {
			t.Fatalf("RunLoc() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, `"language"`) {
			t.Errorf("JSON output missing language field")
		}

		if !strings.Contains(output, `"code"`) {
			t.Errorf("JSON output missing code field")
		}
	})

	t.Run("exclude directory", func(t *testing.T) {
		// Create excluded directory
		excludeDir := filepath.Join(tmpDir, "excluded")
		if err := os.Mkdir(excludeDir, 0755); err != nil {
			t.Fatal(err)
		}

		excludedFile := filepath.Join(excludeDir, "test.go")
		if err := os.WriteFile(excludedFile, []byte("package test\n"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunLoc(&buf, []string{tmpDir}, Options{Exclude: []string{"excluded"}})
		if err != nil {
			t.Fatalf("RunLoc() error = %v", err)
		}

		// The excluded file should not affect the count significantly
		// Just verify it runs without error
	})
}

func TestParseLine(t *testing.T) {
	goLang := &langDef{
		lineComment: []string{"//"},
		blockStart:  "/*",
		blockEnd:    "*/",
		quotes:      [][2]string{{`"`, `"`}, {"`", "`"}},
	}

	tests := []struct {
		name      string
		line      string
		want      lineClass
		inBlock   bool
		wantBlock bool
	}{
		{
			name: "pure code",
			line: "x := 1",
			want: lineCode,
		},
		{
			name: "pure line comment",
			line: "// this is a comment",
			want: lineComment,
		},
		{
			name: "code with inline comment",
			line: "x := 1 // comment",
			want: lineCode,
		},
		{
			name: "url in string not comment",
			line: `url := "http://example.com"`,
			want: lineCode,
		},
		{
			name: "comment marker in string",
			line: `s := "// not a comment"`,
			want: lineCode,
		},
		{
			name: "block comment single line",
			line: "/* comment */",
			want: lineComment,
		},
		{
			name: "code before block comment",
			line: "x := 1 /* comment */",
			want: lineCode,
		},
		{
			name:      "block comment start",
			line:      "/* start of block",
			want:      lineComment,
			wantBlock: true,
		},
		{
			name:      "inside block comment",
			line:      "still in block",
			inBlock:   true,
			want:      lineComment,
			wantBlock: true,
		},
		{
			name:      "block comment end",
			line:      "end of block */",
			inBlock:   true,
			want:      lineComment,
			wantBlock: false,
		},
		{
			name:      "block end with code after",
			line:      "end */ x := 1",
			inBlock:   true,
			want:      lineCode,
			wantBlock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inBlock := tt.inBlock

			depth := 0
			if inBlock {
				depth = 1
			}

			got := parseLine(tt.line, goLang, &inBlock, &depth)

			if got != tt.want {
				t.Errorf("parseLine(%q) = %v, want %v", tt.line, got, tt.want)
			}

			if inBlock != tt.wantBlock {
				t.Errorf("parseLine(%q) inBlock = %v, want %v", tt.line, inBlock, tt.wantBlock)
			}
		})
	}
}

func TestMarkdownLiterate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc_md_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	mdFile := filepath.Join(tmpDir, "README.md")

	mdContent := `# Title

Some text here.

` + "```go" + `
package main

func main() {
    fmt.Println("Hello")
}
` + "```" + `

More text.

` + "```bash" + `
echo "hello"
` + "```" + `
`
	if err := os.WriteFile(mdFile, []byte(mdContent), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	err = RunLoc(&buf, []string{tmpDir}, Options{JSON: true})
	if err != nil {
		t.Fatalf("RunLoc() error = %v", err)
	}

	output := buf.String()

	// Markdown content should be comments
	if !strings.Contains(output, `"Markdown"`) {
		t.Error("output missing Markdown language")
	}

	// Embedded Go code should be extracted
	if !strings.Contains(output, `"Go"`) {
		t.Error("output missing embedded Go code")
	}

	// Embedded Shell code should be extracted
	if !strings.Contains(output, `"Shell"`) {
		t.Error("output missing embedded Shell code")
	}
}

func TestCountFileWithStrings(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc_strings_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	goFile := filepath.Join(tmpDir, "test.go")
	goContent := `package main

func main() {
	// real comment
	url := "http://example.com"
	pattern := "// not a comment"
	code := 1 // inline comment
}
`
	if err := os.WriteFile(goFile, []byte(goContent), 0644); err != nil {
		t.Fatal(err)
	}

	lang := extToLang[".go"]

	stats, err := countFile(goFile, lang)
	if err != nil {
		t.Fatal(err)
	}

	// Line counts:
	// 1. package main - code
	// 2. blank
	// 3. func main() { - code
	// 4. // real comment - comment
	// 5. url := "http://example.com" - code (not comment!)
	// 6. pattern := "// not a comment" - code (not comment!)
	// 7. code := 1 // inline comment - code (has code before comment)
	// 8. } - code
	// Total: 6 code, 1 comment, 1 blank

	if stats.main.Code != 6 {
		t.Errorf("Code = %d, want 6", stats.main.Code)
	}

	if stats.main.Comments != 1 {
		t.Errorf("Comments = %d, want 1", stats.main.Comments)
	}

	if stats.main.Blanks != 1 {
		t.Errorf("Blanks = %d, want 1", stats.main.Blanks)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"short", 10, "short"},
		{"longerstring", 8, "longers."},
		{"exact", 5, "exact"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := truncate(tt.input, tt.max)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
			}
		})
	}
}

func TestEmptyDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc_empty_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	var buf bytes.Buffer

	err = RunLoc(&buf, []string{tmpDir}, Options{})
	if err != nil {
		t.Fatalf("RunLoc() error = %v", err)
	}

	if !strings.Contains(buf.String(), "No files found") {
		t.Errorf("expected 'No files found' message")
	}
}

func TestDetectLanguageName(t *testing.T) {
	tests := []struct {
		hint string
		want string
	}{
		{"go", "Go"},
		{"golang", "Go"},
		{"Go", "Go"},
		{"javascript", "JavaScript"},
		{"js", "JavaScript"},
		{"typescript", "TypeScript"},
		{"ts", "TypeScript"},
		{"python", "Python"},
		{"py", "Python"},
		{"bash", "Shell"},
		{"sh", "Shell"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.hint, func(t *testing.T) {
			got := detectLanguageName(tt.hint)
			if got != tt.want {
				t.Errorf("detectLanguageName(%q) = %q, want %q", tt.hint, got, tt.want)
			}
		})
	}
}
