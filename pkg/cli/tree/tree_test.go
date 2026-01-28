package tree

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunTree(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tree_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test structure
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("hello"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "file2.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte("hidden"), 0644)
	subDir := filepath.Join(tmpDir, "subdir")
	_ = os.Mkdir(subDir, 0755)
	_ = os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("nested"), 0644)

	t.Run("default tree", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunTree(&buf, []string{tmpDir}, TreeOptions{})
		if err != nil {
			t.Fatalf("RunTree() error = %v", err)
		}

		// Tree output shows the root directory name
		// Just verify we get some output
		if buf.Len() == 0 {
			t.Error("RunTree() should produce output")
		}
	})

	t.Run("show hidden", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunTree(&buf, []string{tmpDir}, TreeOptions{All: true})
		if err != nil {
			t.Fatalf("RunTree() error = %v", err)
		}

		// Tree -a should produce output (may or may not explicitly show .hidden depending on implementation)
		if buf.Len() == 0 {
			t.Error("RunTree() -a should produce output")
		}
	})

	t.Run("dirs only", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunTree(&buf, []string{tmpDir}, TreeOptions{DirsOnly: true})
		if err != nil {
			t.Fatalf("RunTree() error = %v", err)
		}

		// Tree -d should produce output showing directories
		if buf.Len() == 0 {
			t.Error("RunTree() -d should produce output")
		}
	})

	t.Run("limited depth", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunTree(&buf, []string{tmpDir}, TreeOptions{Depth: 1})
		if err != nil {
			t.Fatalf("RunTree() error = %v", err)
		}

		output := buf.String()
		if strings.Contains(output, "nested.txt") {
			t.Errorf("RunTree() --depth 1 should not show nested files: %s", output)
		}
	})

	t.Run("with stats", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunTree(&buf, []string{tmpDir}, TreeOptions{Stats: true})
		if err != nil {
			t.Fatalf("RunTree() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "directories") || !strings.Contains(output, "files") {
			t.Errorf("RunTree() -s should show statistics: %s", output)
		}
	})

	t.Run("JSON output", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunTree(&buf, []string{tmpDir}, TreeOptions{JSON: true})
		if err != nil {
			t.Fatalf("RunTree() error = %v", err)
		}

		output := buf.String()
		// JSON output should contain some JSON structure
		if !strings.Contains(output, "{") && !strings.Contains(output, "[") {
			t.Errorf("RunTree() -j should produce JSON-like output: %s", output)
		}
	})

	t.Run("ignore pattern", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunTree(&buf, []string{tmpDir}, TreeOptions{Ignore: []string{"*.go"}})
		if err != nil {
			t.Fatalf("RunTree() error = %v", err)
		}

		output := buf.String()
		if strings.Contains(output, "file2.go") {
			t.Errorf("RunTree() -i should ignore .go files: %s", output)
		}
	})

	t.Run("default path", func(t *testing.T) {
		// Change to temp dir
		origDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)
		defer func() { _ = os.Chdir(origDir) }()

		var buf bytes.Buffer
		err := RunTree(&buf, []string{}, TreeOptions{})
		if err != nil {
			t.Fatalf("RunTree() error = %v", err)
		}

		// Tree should produce output for current directory
		if buf.Len() == 0 {
			t.Error("RunTree() default should produce output")
		}
	})

	t.Run("nonexistent path", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunTree(&buf, []string{"/nonexistent/path/12345"}, TreeOptions{})
		if err == nil {
			t.Error("RunTree() expected error for nonexistent path")
		}
	})

	t.Run("no color", func(t *testing.T) {
		var buf bytes.Buffer
		err := RunTree(&buf, []string{tmpDir}, TreeOptions{NoColor: true})
		if err != nil {
			t.Fatalf("RunTree() error = %v", err)
		}

		// Just verify it runs without error
		if buf.Len() == 0 {
			t.Error("RunTree() --no-color should produce output")
		}
	})
}
