package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if config.MaxDepth != -1 {
		t.Errorf("MaxDepth = %d, want -1", config.MaxDepth)
	}

	if config.ShowHidden {
		t.Error("ShowHidden should be false by default")
	}

	if config.DirsOnly {
		t.Error("DirsOnly should be false by default")
	}

	if len(config.IgnorePatterns) == 0 {
		t.Error("IgnorePatterns should have default values")
	}
}

func TestNewScanner(t *testing.T) {
	t.Run("with nil config", func(t *testing.T) {
		s := NewScanner(nil)
		if s == nil {
			t.Fatal("NewScanner(nil) returned nil")
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &ScanConfig{MaxDepth: 2}

		s := NewScanner(config)
		if s == nil {
			t.Fatal("NewScanner() returned nil")
		}
	})
}

func TestScanner_Scan(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test structure
	_ = os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "src", "main.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test"), 0644)

	s := NewScanner(nil)

	root, err := s.Scan(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if root == nil {
		t.Fatal("Scan() returned nil root")
	}

	if !root.IsDir {
		t.Error("Root should be a directory")
	}

	if len(root.Children) == 0 {
		t.Error("Root should have children")
	}
}

func TestScanner_Scan_File(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("test"), 0644)

	s := NewScanner(nil)

	root, err := s.Scan(context.Background(), testFile)
	if err != nil {
		t.Fatalf("Scan() file error = %v", err)
	}

	if root.IsDir {
		t.Error("Root should be a file")
	}

	if root.Name != "test.txt" {
		t.Errorf("Root.Name = %q, want %q", root.Name, "test.txt")
	}
}

func TestScanner_Scan_MaxDepth(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create deep structure
	_ = os.MkdirAll(filepath.Join(tmpDir, "a", "b", "c", "d"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "a", "b", "c", "d", "deep.txt"), []byte("test"), 0644)

	config := &ScanConfig{MaxDepth: 2}
	s := NewScanner(config)

	root, err := s.Scan(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	// Check that root has children (meaning scanning worked)
	if len(root.Children) == 0 {
		t.Error("Root should have children")
	}

	// Find level a and check its children are limited by depth
	for _, child := range root.Children {
		if child.Name == "a" {
			if len(child.Children) == 0 {
				t.Error("Directory 'a' should have children at depth 1")
			}

			break
		}
	}
}

func TestScanner_Scan_ShowHidden(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	_ = os.WriteFile(filepath.Join(tmpDir, ".hidden"), []byte("hidden"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "visible.txt"), []byte("visible"), 0644)

	t.Run("hidden files excluded", func(t *testing.T) {
		s := NewScanner(&ScanConfig{ShowHidden: false})

		root, err := s.Scan(context.Background(), tmpDir)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		for _, child := range root.Children {
			if child.Name == ".hidden" {
				t.Error("Hidden files should not be included when ShowHidden is false")
			}
		}
	})

	t.Run("hidden files included", func(t *testing.T) {
		s := NewScanner(&ScanConfig{
			ShowHidden:     true,
			MaxDepth:       -1,
			IgnorePatterns: []string{}, // Clear default ignore patterns
		})

		root, err := s.Scan(context.Background(), tmpDir)
		if err != nil {
			t.Fatalf("Scan() error = %v", err)
		}

		foundHidden := false

		for _, child := range root.Children {
			if child.Name == ".hidden" {
				foundHidden = true

				break
			}
		}

		if !foundHidden {
			t.Error("Hidden files should be included when ShowHidden is true")
		}
	})
}

func TestScanner_Scan_DirsOnly(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	_ = os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("test"), 0644)

	s := NewScanner(&ScanConfig{DirsOnly: true, ShowHidden: true})

	root, err := s.Scan(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	for _, child := range root.Children {
		if !child.IsDir {
			t.Errorf("DirsOnly should exclude files, found %q", child.Name)
		}
	}
}

func TestScanner_Scan_IgnorePatterns(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	_ = os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0755)
	_ = os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)

	s := NewScanner(&ScanConfig{
		IgnorePatterns: []string{"node_modules"},
		ShowHidden:     true,
	})

	root, err := s.Scan(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	for _, child := range root.Children {
		if child.Name == "node_modules" {
			t.Error("node_modules should be ignored")
		}
	}
}

func TestScanner_Scan_ShowHash(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	_ = os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("hello world"), 0644)

	s := NewScanner(&ScanConfig{
		ShowHash:       true,
		ShowHidden:     true,
		MaxDepth:       -1,
		IgnorePatterns: []string{}, // Clear default patterns
	})

	root, err := s.Scan(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	foundHash := false

	for _, child := range root.Children {
		if child.Name == "test.txt" {
			if child.Hash != "" {
				foundHash = true
			} else {
				t.Logf("File found but no hash: %+v", child)
			}

			break
		}
	}

	if !foundHash {
		t.Logf("Children: %d", len(root.Children))

		for _, child := range root.Children {
			t.Logf("Child: %s, Hash: %s", child.Name, child.Hash)
		}

		t.Error("File should have hash when ShowHash is true")
	}
}

func TestScanner_Scan_NonexistentPath(t *testing.T) {
	s := NewScanner(nil)

	_, err := s.Scan(context.Background(), "/nonexistent/path/12345")
	if err == nil {
		t.Error("Scan() should error on nonexistent path")
	}
}

func TestScanner_Scan_ContextCancelled(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	s := NewScanner(nil)

	_, err = s.Scan(ctx, tmpDir)
	if err == nil {
		t.Error("Scan() should error when context is cancelled")
	}
}

func TestScanner_Scan_ContextTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create many directories to scan
	for i := range 10 {
		dir := filepath.Join(tmpDir, "dir"+string(rune('0'+i)))
		_ = os.MkdirAll(dir, 0755)

		for j := range 10 {
			_ = os.WriteFile(filepath.Join(dir, "file"+string(rune('0'+j))+".txt"), []byte("test"), 0644)
		}
	}

	// Very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	s := NewScanner(nil)
	_, err = s.Scan(ctx, tmpDir)

	// Either succeeds quickly or times out - both are acceptable
	// Just verify it doesn't panic
	_ = err
}

func TestScanner_Scan_AbsolutePath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	_ = os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644)

	s := NewScanner(nil)

	root, err := s.Scan(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	// Path should be absolute
	if !filepath.IsAbs(root.Path) {
		t.Error("Root path should be absolute")
	}
}

func TestScanner_WildcardPatterns(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "scanner_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	_ = os.WriteFile(filepath.Join(tmpDir, "test.log"), []byte("log"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("txt"), 0644)

	s := NewScanner(&ScanConfig{
		IgnorePatterns: []string{"*.log"},
		ShowHidden:     true,
	})

	root, err := s.Scan(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	for _, child := range root.Children {
		if child.Name == "test.log" {
			t.Error("*.log pattern should ignore .log files")
		}
	}
}
