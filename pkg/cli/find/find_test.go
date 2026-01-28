package find

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunFind(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "find_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test structure
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("hello"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file2.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "empty.txt"), []byte(""), 0644)
	subDir := filepath.Join(tmpDir, "subdir")
	_ = os.Mkdir(subDir, 0755)
	_ = os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("nested"), 0644)
	emptyDir := filepath.Join(tmpDir, "emptydir")
	_ = os.Mkdir(emptyDir, 0755)

	t.Run("default find all", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunFind(&buf, []string{tmpDir}, FindOptions{})
		if err != nil {
			t.Fatalf("RunFind() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "file1.txt") {
			t.Errorf("RunFind() should find file1.txt: %s", output)
		}
	})

	t.Run("find by name", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunFind(&buf, []string{tmpDir}, FindOptions{Name: "*.txt"})
		if err != nil {
			t.Fatalf("RunFind() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "file1.txt") {
			t.Errorf("RunFind() -name should find txt files: %s", output)
		}
		if strings.Contains(output, "file2.go") {
			t.Errorf("RunFind() -name should not find go files: %s", output)
		}
	})

	t.Run("find by iname", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunFind(&buf, []string{tmpDir}, FindOptions{IName: "*.TXT"})
		if err != nil {
			t.Fatalf("RunFind() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "file1.txt") {
			t.Errorf("RunFind() -iname should find txt files case-insensitive: %s", output)
		}
	})

	t.Run("find by type file", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunFind(&buf, []string{tmpDir}, FindOptions{Type: "f"})
		if err != nil {
			t.Fatalf("RunFind() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "file1.txt") {
			t.Errorf("RunFind() -type f should find files: %s", output)
		}
		// Should not include directories in the matches (though root may appear)
	})

	t.Run("find by type directory", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunFind(&buf, []string{tmpDir}, FindOptions{Type: "d"})
		if err != nil {
			t.Fatalf("RunFind() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "subdir") {
			t.Errorf("RunFind() -type d should find directories: %s", output)
		}
	})

	t.Run("find empty files", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunFind(&buf, []string{tmpDir}, FindOptions{Empty: true, Type: "f"})
		if err != nil {
			t.Fatalf("RunFind() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "empty.txt") {
			t.Errorf("RunFind() -empty should find empty files: %s", output)
		}
	})

	t.Run("find empty directories", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunFind(&buf, []string{tmpDir}, FindOptions{Empty: true, Type: "d"})
		if err != nil {
			t.Fatalf("RunFind() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "emptydir") {
			t.Errorf("RunFind() -empty -type d should find empty dirs: %s", output)
		}
	})

	t.Run("maxdepth", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunFind(&buf, []string{tmpDir}, FindOptions{MaxDepth: 1})
		if err != nil {
			t.Fatalf("RunFind() error = %v", err)
		}

		output := buf.String()
		if strings.Contains(output, "nested.txt") {
			t.Errorf("RunFind() -maxdepth 1 should not find nested files: %s", output)
		}
	})

	t.Run("mindepth", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunFind(&buf, []string{tmpDir}, FindOptions{MinDepth: 2})
		if err != nil {
			t.Fatalf("RunFind() error = %v", err)
		}

		output := buf.String()
		if strings.Contains(output, "file1.txt") {
			t.Errorf("RunFind() -mindepth 2 should not find top level files: %s", output)
		}
	})

	t.Run("print0", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunFind(&buf, []string{tmpDir}, FindOptions{Print0: true})
		if err != nil {
			t.Fatalf("RunFind() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "\x00") {
			t.Error("RunFind() -print0 should use null separator")
		}
	})

	t.Run("regex", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunFind(&buf, []string{tmpDir}, FindOptions{Regex: ".*\\.go$"})
		if err != nil {
			t.Fatalf("RunFind() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "file2.go") {
			t.Errorf("RunFind() -regex should match .go files: %s", output)
		}
	})

	t.Run("readable", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunFind(&buf, []string{tmpDir}, FindOptions{Readable: true, Type: "f"})
		if err != nil {
			t.Fatalf("RunFind() error = %v", err)
		}

		// Should find readable files
		if buf.Len() == 0 {
			t.Error("RunFind() -readable should find readable files")
		}
	})

	t.Run("default path", func(t *testing.T) {
		// Change to temp dir
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		var buf bytes.Buffer
		err := RunFind(&buf, []string{}, FindOptions{})
		if err != nil {
			t.Fatalf("RunFind() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, ".") {
			t.Errorf("RunFind() default should use current dir: %s", output)
		}
	})
}

func TestGlobToRegex(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		match   bool
	}{
		{"*.txt", "file.txt", true},
		{"*.txt", "file.go", false},
		{"file?.txt", "file1.txt", true},
		{"file?.txt", "file12.txt", false},
		{"[abc].txt", "a.txt", true},
		{"[abc].txt", "d.txt", false},
		{"[!abc].txt", "d.txt", true},
		{"[!abc].txt", "a.txt", false},
		{"*.go", "main.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.input, func(t *testing.T) {
			re, err := globToRegex(tt.pattern, false)
			if err != nil {
				t.Fatalf("globToRegex() error = %v", err)
			}

			got := re.MatchString(tt.input)
			if got != tt.match {
				t.Errorf("globToRegex(%q).MatchString(%q) = %v, want %v", tt.pattern, tt.input, got, tt.match)
			}
		})
	}
}

func TestGlobToRegexCaseInsensitive(t *testing.T) {
	re, err := globToRegex("*.TXT", true)
	if err != nil {
		t.Fatalf("globToRegex() error = %v", err)
	}

	if !re.MatchString("file.txt") {
		t.Error("globToRegex with caseInsensitive should match lowercase")
	}
	if !re.MatchString("FILE.TXT") {
		t.Error("globToRegex with caseInsensitive should match uppercase")
	}
}

func TestParseSizeFilter(t *testing.T) {
	tests := []struct {
		sizeStr  string
		fileSize int64
		match    bool
	}{
		{"+1k", 2048, true},
		{"+1k", 512, false},
		{"-1k", 0, true},    // 0 bytes rounds to 0 units, 0 < 1
		{"-1k", 512, false}, // 512 bytes rounds up to 1 unit, 1 < 1 = false
		{"-1k", 2048, false},
		{"1c", 1, true},
		{"1c", 2, false},
		{"+0c", 1, true},
		{"-2c", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.sizeStr, func(t *testing.T) {
			filter, err := parseSizeFilter(tt.sizeStr)
			if err != nil {
				t.Fatalf("parseSizeFilter() error = %v", err)
			}

			got := filter(tt.fileSize)
			if got != tt.match {
				t.Errorf("parseSizeFilter(%q)(%d) = %v, want %v", tt.sizeStr, tt.fileSize, got, tt.match)
			}
		})
	}
}

func TestParseTimeFilter(t *testing.T) {
	// Test with minutes
	filter, err := parseTimeFilter("-5", time.Minute)
	if err != nil {
		t.Fatalf("parseTimeFilter() error = %v", err)
	}

	// A time 2 minutes ago should match "-5" (less than 5 minutes ago)
	recentTime := time.Now().Add(-2 * time.Minute)
	if !filter(recentTime) {
		t.Error("parseTimeFilter('-5' minutes) should match time 2 minutes ago")
	}

	// A time 10 minutes ago should not match "-5"
	oldTime := time.Now().Add(-10 * time.Minute)
	if filter(oldTime) {
		t.Error("parseTimeFilter('-5' minutes) should not match time 10 minutes ago")
	}
}

func TestIsEmpty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "isempty_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create empty file
	emptyFile := filepath.Join(tmpDir, "empty.txt")
	_ = os.WriteFile(emptyFile, []byte(""), 0644)

	// Create non-empty file
	nonEmptyFile := filepath.Join(tmpDir, "nonempty.txt")
	_ = os.WriteFile(nonEmptyFile, []byte("content"), 0644)

	// Create empty directory
	emptyDir := filepath.Join(tmpDir, "emptydir")
	_ = os.Mkdir(emptyDir, 0755)

	// Create non-empty directory
	nonEmptyDir := filepath.Join(tmpDir, "nonemptydir")
	_ = os.Mkdir(nonEmptyDir, 0755)
	_ = os.WriteFile(filepath.Join(nonEmptyDir, "file.txt"), []byte("content"), 0644)

	tests := []struct {
		path  string
		isDir bool
		want  bool
	}{
		{emptyFile, false, true},
		{nonEmptyFile, false, false},
		{emptyDir, true, true},
		{nonEmptyDir, true, false},
	}

	for _, tt := range tests {
		t.Run(filepath.Base(tt.path), func(t *testing.T) {
			entries, _ := os.ReadDir(filepath.Dir(tt.path))
			for _, e := range entries {
				if filepath.Join(filepath.Dir(tt.path), e.Name()) == tt.path {
					got := isEmpty(tt.path, e)
					if got != tt.want {
						t.Errorf("isEmpty(%q) = %v, want %v", tt.path, got, tt.want)
					}
					break
				}
			}
		})
	}
}

func TestIsReadable(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "readable_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	file := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(file, []byte("content"), 0644)

	if !isReadable(file) {
		t.Error("isReadable() should return true for readable file")
	}

	if isReadable("/nonexistent/file.txt") {
		t.Error("isReadable() should return false for nonexistent file")
	}
}
