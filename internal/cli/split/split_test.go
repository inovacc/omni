package split

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSplit(t *testing.T) {
	// Use current directory for output files
	origDir, _ := os.Getwd()

	tmpDir, err := os.MkdirTemp("", "split_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = os.Chdir(origDir)
		_ = os.RemoveAll(tmpDir)
	}()

	_ = os.Chdir(tmpDir)

	t.Run("split by lines", func(t *testing.T) {
		input := filepath.Join(tmpDir, "input.txt")
		_ = os.WriteFile(input, []byte("line1\nline2\nline3\nline4\nline5\n"), 0644)

		var buf bytes.Buffer

		err := RunSplit(&buf, []string{input}, SplitOptions{Lines: 2})
		if err != nil {
			t.Fatalf("RunSplit() error = %v", err)
		}

		// Check output files were created
		if _, err := os.Stat("xaa"); os.IsNotExist(err) {
			t.Error("RunSplit() did not create xaa")
		}

		if _, err := os.Stat("xab"); os.IsNotExist(err) {
			t.Error("RunSplit() did not create xab")
		}

		// Clean up
		_ = os.Remove("xaa")
		_ = os.Remove("xab")
		_ = os.Remove("xac")
	})

	t.Run("split by bytes", func(t *testing.T) {
		input := filepath.Join(tmpDir, "bytes.txt")
		_ = os.WriteFile(input, []byte("0123456789"), 0644)

		var buf bytes.Buffer

		err := RunSplit(&buf, []string{input}, SplitOptions{Bytes: "5"})
		if err != nil {
			t.Fatalf("RunSplit() error = %v", err)
		}

		// Check xaa has 5 bytes
		content, _ := os.ReadFile("xaa")
		if len(content) != 5 {
			t.Errorf("RunSplit() xaa has %d bytes, want 5", len(content))
		}

		// Clean up
		_ = os.Remove("xaa")
		_ = os.Remove("xab")
	})

	t.Run("custom prefix", func(t *testing.T) {
		input := filepath.Join(tmpDir, "prefix.txt")
		_ = os.WriteFile(input, []byte("line1\nline2\n"), 0644)

		var buf bytes.Buffer

		err := RunSplit(&buf, []string{input, "myprefix"}, SplitOptions{Lines: 1})
		if err != nil {
			t.Fatalf("RunSplit() error = %v", err)
		}

		if _, err := os.Stat("myprefixaa"); os.IsNotExist(err) {
			t.Error("RunSplit() did not create file with custom prefix")
		}

		// Clean up
		_ = os.Remove("myprefixaa")
		_ = os.Remove("myprefixab")
	})

	t.Run("numeric suffix", func(t *testing.T) {
		input := filepath.Join(tmpDir, "numeric.txt")
		_ = os.WriteFile(input, []byte("line1\nline2\n"), 0644)

		var buf bytes.Buffer

		err := RunSplit(&buf, []string{input, "num"}, SplitOptions{Lines: 1, NumericSufx: true})
		if err != nil {
			t.Fatalf("RunSplit() error = %v", err)
		}

		if _, err := os.Stat("num00"); os.IsNotExist(err) {
			t.Error("RunSplit() did not create file with numeric suffix")
		}

		// Clean up
		_ = os.Remove("num00")
		_ = os.Remove("num01")
	})

	t.Run("verbose mode", func(t *testing.T) {
		input := filepath.Join(tmpDir, "verbose.txt")
		_ = os.WriteFile(input, []byte("line1\nline2\n"), 0644)

		var buf bytes.Buffer

		err := RunSplit(&buf, []string{input}, SplitOptions{Lines: 1, Verbose: true})
		if err != nil {
			t.Fatalf("RunSplit() error = %v", err)
		}

		if !strings.Contains(buf.String(), "creating file") {
			t.Errorf("RunSplit() verbose output missing")
		}

		// Clean up
		_ = os.Remove("xaa")
		_ = os.Remove("xab")
	})

	t.Run("default 1000 lines", func(t *testing.T) {
		// Create file with 10 lines (less than default 1000)
		input := filepath.Join(tmpDir, "default.txt")

		var content strings.Builder
		for range 10 {
			content.WriteString("line\n")
		}

		_ = os.WriteFile(input, []byte(content.String()), 0644)

		var buf bytes.Buffer

		err := RunSplit(&buf, []string{input}, SplitOptions{})
		if err != nil {
			t.Fatalf("RunSplit() error = %v", err)
		}

		// Should create only one file since 10 < 1000
		if _, err := os.Stat("xaa"); os.IsNotExist(err) {
			t.Error("RunSplit() did not create xaa")
		}

		if _, err := os.Stat("xab"); !os.IsNotExist(err) {
			t.Error("RunSplit() should not create xab for small file")
		}

		_ = os.Remove("xaa")
	})

	t.Run("nonexistent input file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunSplit(&buf, []string{"/nonexistent/file.txt"}, SplitOptions{})
		if err == nil {
			t.Error("RunSplit() expected error for nonexistent file")
		}
	})

	t.Run("invalid byte size", func(t *testing.T) {
		input := filepath.Join(tmpDir, "invalid.txt")
		_ = os.WriteFile(input, []byte("data"), 0644)

		var buf bytes.Buffer

		err := RunSplit(&buf, []string{input}, SplitOptions{Bytes: "abc"})
		if err == nil {
			t.Error("RunSplit() expected error for invalid byte size")
		}
	})
}

func TestGenerateSuffix(t *testing.T) {
	tests := []struct {
		num      int
		length   int
		numeric  bool
		expected string
	}{
		{0, 2, false, "aa"},
		{1, 2, false, "ab"},
		{25, 2, false, "az"},
		{26, 2, false, "ba"},
		{0, 2, true, "00"},
		{1, 2, true, "01"},
		{99, 2, true, "99"},
		{0, 3, false, "aaa"},
		{0, 3, true, "000"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := generateSuffix(tt.num, tt.length, tt.numeric)
			if result != tt.expected {
				t.Errorf("generateSuffix(%d, %d, %v) = %q, want %q",
					tt.num, tt.length, tt.numeric, result, tt.expected)
			}
		})
	}
}

func TestParseByteSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		{"100", 100, false},
		{"1K", 1024, false},
		{"1k", 1024, false},
		{"1M", 1024 * 1024, false},
		{"1m", 1024 * 1024, false},
		{"1G", 1024 * 1024 * 1024, false},
		{"2K", 2048, false},
		{"", 0, true},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseByteSize(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseByteSize(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != tt.expected {
				t.Errorf("parseByteSize(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}
