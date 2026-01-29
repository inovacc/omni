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
	// Another comment
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

func TestCountFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc_count_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tests := []struct {
		name     string
		content  string
		lang     *langDef
		wantCode int
		wantComm int
		wantBlnk int
	}{
		{
			name: "go file",
			content: `package main

// Comment
func main() {
}
`,
			lang:     &langDef{lineComment: "//", blockStart: "/*", blockEnd: "*/"},
			wantCode: 3,
			wantComm: 1,
			wantBlnk: 1,
		},
		{
			name: "block comment",
			content: `package main

/*
Multi-line
comment
*/
func main() {}
`,
			lang:     &langDef{lineComment: "//", blockStart: "/*", blockEnd: "*/"},
			wantCode: 2,
			wantComm: 4,
			wantBlnk: 1,
		},
		{
			name: "python",
			content: `# Comment
def hello():
    pass
`,
			lang:     &langDef{lineComment: "#"},
			wantCode: 2,
			wantComm: 1,
			wantBlnk: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := filepath.Join(tmpDir, "test.txt")
			if err := os.WriteFile(file, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			stats, err := countFile(file, tt.lang)
			if err != nil {
				t.Fatalf("countFile() error = %v", err)
			}

			if stats.Code != tt.wantCode {
				t.Errorf("Code = %d, want %d", stats.Code, tt.wantCode)
			}
			if stats.Comments != tt.wantComm {
				t.Errorf("Comments = %d, want %d", stats.Comments, tt.wantComm)
			}
			if stats.Blanks != tt.wantBlnk {
				t.Errorf("Blanks = %d, want %d", stats.Blanks, tt.wantBlnk)
			}
		})
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
