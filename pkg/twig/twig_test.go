package twig

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewTree(t *testing.T) {
	tree := NewTree()

	if tree == nil {
		t.Fatal("NewTree() returned nil")
	}

	if tree.scanConfig == nil {
		t.Error("NewTree().scanConfig should not be nil")
	}

	if tree.formatConfig == nil {
		t.Error("NewTree().formatConfig should not be nil")
	}

	if tree.buildConfig == nil {
		t.Error("NewTree().buildConfig should not be nil")
	}
}

func TestNewTree_WithOptions(t *testing.T) {
	tree := NewTree(
		WithMaxDepth(3),
		WithShowHidden(true),
		WithDirsOnly(true),
	)

	if tree.scanConfig.MaxDepth != 3 {
		t.Errorf("MaxDepth = %d, want 3", tree.scanConfig.MaxDepth)
	}

	if !tree.scanConfig.ShowHidden {
		t.Error("ShowHidden should be true")
	}

	if !tree.scanConfig.DirsOnly {
		t.Error("DirsOnly should be true")
	}
}

func TestTree_Generate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "twig_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test structure
	_ = os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "src", "main.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test"), 0644)

	tree := NewTree()

	output, err := tree.Generate(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if output == "" {
		t.Error("Generate() should produce output")
	}

	if !strings.Contains(output, "src") {
		t.Error("Generate() output should contain 'src'")
	}

	if !strings.Contains(output, "README.md") {
		t.Error("Generate() output should contain 'README.md'")
	}
}

func TestTree_GenerateWithStats(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "twig_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test structure
	_ = os.MkdirAll(filepath.Join(tmpDir, "dir1"), 0755)
	_ = os.MkdirAll(filepath.Join(tmpDir, "dir2"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "dir1", "file2.txt"), []byte("test"), 0644)

	tree := NewTree()

	result, err := tree.GenerateWithStats(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("GenerateWithStats() error = %v", err)
	}

	if result.Output == "" {
		t.Error("GenerateWithStats() should produce output")
	}

	if result.Stats == nil {
		t.Fatal("GenerateWithStats() should return stats")
	}

	// Root dir + 2 subdirs
	if result.Stats.TotalDirs != 3 {
		t.Errorf("TotalDirs = %d, want 3", result.Stats.TotalDirs)
	}

	if result.Stats.TotalFiles != 2 {
		t.Errorf("TotalFiles = %d, want 2", result.Stats.TotalFiles)
	}

	if result.Root == nil {
		t.Error("GenerateWithStats() should return root node")
	}
}

func TestTree_GenerateJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "twig_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	_ = os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644)

	tree := NewTree()

	output, err := tree.GenerateJSON(context.Background(), tmpDir, true)
	if err != nil {
		t.Fatalf("GenerateJSON() error = %v", err)
	}

	if !strings.Contains(output, "\"tree\"") {
		t.Error("GenerateJSON() should contain 'tree' key")
	}

	if !strings.Contains(output, "\"stats\"") {
		t.Error("GenerateJSON() should contain 'stats' key")
	}
}

func TestTree_Parse(t *testing.T) {
	tree := NewTree()

	treeFormat := `project/
├── src/
└── README.md`

	root, err := tree.Parse(treeFormat)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if root == nil {
		t.Fatal("Parse() should return root node")
	}

	if root.Name != "project" {
		t.Errorf("Root name = %q, want %q", root.Name, "project")
	}

	if !root.IsDir {
		t.Error("Root should be a directory")
	}

	if len(root.Children) != 2 {
		t.Errorf("Root should have 2 children, got %d", len(root.Children))
	}
}

func TestTree_Format(t *testing.T) {
	tree := NewTree()

	treeFormat := `project/
├── src/
│   └── main.go
└── README.md`

	root, err := tree.Parse(treeFormat)
	if err != nil {
		t.Fatal(err)
	}

	output := tree.Format(root)

	if output == "" {
		t.Error("Format() should produce output")
	}

	if !strings.Contains(output, "project") {
		t.Error("Format() output should contain 'project'")
	}
}

func TestTree_Scan(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "twig_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	_ = os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("test"), 0644)

	tree := NewTree()

	root, err := tree.Scan(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if root == nil {
		t.Fatal("Scan() should return root node")
	}

	if len(root.Children) != 2 {
		t.Errorf("Root should have 2 children, got %d", len(root.Children))
	}
}

func TestTree_CreateFromString(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "twig_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	tree := NewTree()

	treeFormat := `newproject/
├── src/
└── README.md`

	result, err := tree.CreateFromString(context.Background(), treeFormat, tmpDir)
	if err != nil {
		t.Fatalf("CreateFromString() error = %v", err)
	}

	if result == nil {
		t.Fatal("CreateFromString() should return result")
	}

	// Check that result has data
	if result.BuildResult == nil {
		t.Error("CreateFromString() should return build result")
	}

	if result.Root == nil {
		t.Error("CreateFromString() should return root node")
	}
}

func TestTree_Build(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "twig_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	tree := NewTree()

	// Parse a structure
	treeFormat := `buildtest/
└── subdir/`

	root, err := tree.Parse(treeFormat)
	if err != nil {
		t.Fatal(err)
	}

	// Build it
	result, err := tree.Build(context.Background(), root, tmpDir)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if result == nil {
		t.Fatal("Build() should return result")
	}

	// Verify result structure (builder creates children, not root)
	if result.DryRun {
		t.Error("Build() should not be in dry run mode by default")
	}
}

func TestTree_Generate_MaxDepth(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "twig_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create deep structure
	deepPath := filepath.Join(tmpDir, "a", "b", "c", "d")
	_ = os.MkdirAll(deepPath, 0755)
	_ = os.WriteFile(filepath.Join(deepPath, "deep.txt"), []byte("test"), 0644)

	tree := NewTree(WithMaxDepth(2))

	output, err := tree.Generate(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Generate() with MaxDepth error = %v", err)
	}

	// Should not contain deeply nested file
	if strings.Contains(output, "deep.txt") {
		t.Error("Generate() with MaxDepth=2 should not show files at depth 4")
	}

	// Should contain first levels
	if !strings.Contains(output, "a") {
		t.Error("Generate() should contain first level 'a'")
	}
}

func TestTree_Generate_DirsOnly(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "twig_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	_ = os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("test"), 0644)

	tree := NewTree(WithDirsOnly(true))

	output, err := tree.Generate(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Generate() with DirsOnly error = %v", err)
	}

	if strings.Contains(output, "file.txt") {
		t.Error("Generate() with DirsOnly should not show files")
	}

	if !strings.Contains(output, "subdir") {
		t.Error("Generate() with DirsOnly should show directories")
	}
}

func TestTree_Generate_IgnorePatterns(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "twig_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	_ = os.MkdirAll(filepath.Join(tmpDir, "node_modules"), 0755)
	_ = os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)

	tree := NewTree(WithIgnorePatterns("node_modules"))

	output, err := tree.Generate(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("Generate() with IgnorePatterns error = %v", err)
	}

	if strings.Contains(output, "node_modules") {
		t.Error("Generate() should ignore 'node_modules'")
	}

	if !strings.Contains(output, "src") {
		t.Error("Generate() should show 'src'")
	}
}

func TestTree_DryRun(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "twig_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	tree := NewTree(WithDryRun(true))

	treeFormat := `dryrun/
└── subdir/`

	result, err := tree.CreateFromString(context.Background(), treeFormat, tmpDir)
	if err != nil {
		t.Fatalf("CreateFromString() with DryRun error = %v", err)
	}

	if result == nil {
		t.Fatal("CreateFromString() should return result")
	}

	// Directory should NOT be created in dry-run mode
	dryrunDir := filepath.Join(tmpDir, "dryrun")
	if _, err := os.Stat(dryrunDir); !os.IsNotExist(err) {
		t.Error("CreateFromString() with DryRun should not create directory")
	}
}

func TestTree_Generate_NonexistentPath(t *testing.T) {
	tree := NewTree()

	_, err := tree.Generate(context.Background(), "/nonexistent/path/12345")
	if err == nil {
		t.Error("Generate() should error on nonexistent path")
	}
}

func TestTree_Parse_InvalidFormat(t *testing.T) {
	tree := NewTree()

	_, err := tree.Parse("")
	if err == nil {
		t.Error("Parse() should error on empty input")
	}
}
