package builder

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/inovacc/omni/pkg/twig/models"
)

func TestDefaultBuildConfig(t *testing.T) {
	cfg := DefaultBuildConfig()
	if cfg.DryRun || cfg.Overwrite || cfg.SkipExisting || cfg.AbortOnConflict || cfg.Verbose {
		t.Error("DefaultBuildConfig should have all bools false")
	}
}

func TestNewBuilder_NilConfig(t *testing.T) {
	b := NewBuilder(nil)
	if b == nil {
		t.Fatal("NewBuilder(nil) returned nil")
	}
}

func TestBuild_DryRun(t *testing.T) {
	root := models.NewNode("project", "project", true)
	root.AddChild(models.NewNode("src", "project/src", true))
	root.AddChild(models.NewNode("README.md", "project/README.md", false))

	b := NewBuilder(&BuildConfig{DryRun: true})
	target := filepath.Join(t.TempDir(), "project")

	result, err := b.Build(context.Background(), root, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.DryRun {
		t.Error("result.DryRun should be true")
	}
	if len(result.Created) == 0 {
		t.Error("result.Created should not be empty in dry run")
	}

	// Nothing should exist on disk
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Error("target should not exist on disk in dry run")
	}
}

func TestBuild_CreatesFilesAndDirs(t *testing.T) {
	root := models.NewNode("project", "project", true)
	src := models.NewNode("src", "project/src", true)
	root.AddChild(src)
	src.AddChild(models.NewNode("main.go", "project/src/main.go", false))
	root.AddChild(models.NewNode("go.mod", "project/go.mod", false))

	target := filepath.Join(t.TempDir(), "project")
	b := NewBuilder(&BuildConfig{})

	result, err := b.Build(context.Background(), root, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Created) == 0 {
		t.Error("should have created items")
	}

	// Verify directory exists
	info, err := os.Stat(filepath.Join(target, "src"))
	if err != nil {
		t.Fatalf("src dir should exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("src should be a directory")
	}

	// Verify file exists
	if _, err := os.Stat(filepath.Join(target, "go.mod")); err != nil {
		t.Errorf("go.mod should exist: %v", err)
	}
}

func TestBuild_FileWithComment(t *testing.T) {
	root := models.NewNode("proj", "proj", true)
	child := models.NewNode("config.yaml", "proj/config.yaml", false)
	child.Comment = "main configuration"
	root.AddChild(child)

	target := filepath.Join(t.TempDir(), "proj")
	b := NewBuilder(&BuildConfig{})

	_, err := b.Build(context.Background(), root, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(target, "config.yaml"))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	expected := "# main configuration\n"
	if string(data) != expected {
		t.Errorf("file content = %q, want %q", string(data), expected)
	}
}

func TestBuild_AbortOnConflict(t *testing.T) {
	root := models.NewNode("existing", "existing", true)
	target := filepath.Join(t.TempDir(), "existing")

	// Pre-create the target directory
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}

	b := NewBuilder(&BuildConfig{AbortOnConflict: true})
	_, err := b.Build(context.Background(), root, target)
	if !errors.Is(err, ErrTargetExists) {
		t.Errorf("err = %v, want ErrTargetExists", err)
	}
}

func TestBuild_SkipExisting(t *testing.T) {
	root := models.NewNode("proj", "proj", true)
	root.AddChild(models.NewNode("existing.txt", "proj/existing.txt", false))

	target := filepath.Join(t.TempDir(), "proj")
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "existing.txt"), []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}

	b := NewBuilder(&BuildConfig{SkipExisting: true})
	result, err := b.Build(context.Background(), root, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Skipped) == 0 {
		t.Error("should have skipped existing file")
	}

	// Content should be unchanged
	data, err := os.ReadFile(filepath.Join(target, "existing.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "original" {
		t.Errorf("file was modified, content = %q", string(data))
	}
}

func TestBuild_ContextCancelled(t *testing.T) {
	root := models.NewNode("proj", "proj", true)
	root.AddChild(models.NewNode("file.txt", "proj/file.txt", false))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	target := filepath.Join(t.TempDir(), "proj")
	b := NewBuilder(&BuildConfig{})

	_, err := b.Build(ctx, root, target)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestPrintResult_NoPanic(t *testing.T) {
	// Just verify it doesn't panic
	PrintResult(&BuildResult{
		Created: []string{"a", "b"},
		Skipped: []string{"c"},
		DryRun:  true,
	})
}

func TestPrintResult_WithErrors(t *testing.T) {
	// Just verify it doesn't panic with errors
	PrintResult(&BuildResult{
		Created: []string{"a"},
		Errors:  []error{errors.New("something went wrong"), errors.New("another error")},
		DryRun:  false,
	})
}

func TestBuild_Overwrite(t *testing.T) {
	root := models.NewNode("proj", "proj", true)
	root.AddChild(models.NewNode("file.txt", "proj/file.txt", false))

	target := filepath.Join(t.TempDir(), "proj")
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "file.txt"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	b := NewBuilder(&BuildConfig{Overwrite: true})
	result, err := b.Build(context.Background(), root, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// File existed so not added to Created, but no error either
	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
}

func TestBuild_VerboseOutput(t *testing.T) {
	root := models.NewNode("proj", "proj", true)
	root.AddChild(models.NewNode("subdir", "proj/subdir", true))
	root.AddChild(models.NewNode("file.txt", "proj/file.txt", false))

	target := filepath.Join(t.TempDir(), "proj")
	b := NewBuilder(&BuildConfig{Verbose: true})

	result, err := b.Build(context.Background(), root, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Created) == 0 {
		t.Error("should have created items with verbose mode")
	}
}

func TestBuild_DryRunVerbose(t *testing.T) {
	root := models.NewNode("proj", "proj", true)
	root.AddChild(models.NewNode("subdir", "proj/subdir", true))
	root.AddChild(models.NewNode("file.txt", "proj/file.txt", false))

	target := filepath.Join(t.TempDir(), "proj")
	b := NewBuilder(&BuildConfig{DryRun: true, Verbose: true})

	result, err := b.Build(context.Background(), root, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.DryRun {
		t.Error("result.DryRun should be true")
	}
}

func TestBuild_AbortOnConflictChild(t *testing.T) {
	root := models.NewNode("proj", "proj", true)
	root.AddChild(models.NewNode("existing.txt", "proj/existing.txt", false))

	target := filepath.Join(t.TempDir(), "proj")
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "existing.txt"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	b := NewBuilder(&BuildConfig{AbortOnConflict: true})
	_, err := b.Build(context.Background(), root, target)
	if err == nil {
		t.Error("expected error when AbortOnConflict and child exists")
	}
}

func TestBuild_TypeMismatch(t *testing.T) {
	// Create a file where the tree expects a directory
	root := models.NewNode("proj", "proj", true)
	dirNode := models.NewNode("item", "proj/item", true) // expect dir
	root.AddChild(dirNode)

	target := filepath.Join(t.TempDir(), "proj")
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}
	// Create a file where builder expects a directory
	if err := os.WriteFile(filepath.Join(target, "item"), []byte("file content"), 0644); err != nil {
		t.Fatal(err)
	}

	b := NewBuilder(&BuildConfig{})
	result, err := b.Build(context.Background(), root, target)
	if err != nil {
		t.Fatalf("unexpected fatal error: %v", err)
	}
	if len(result.Errors) == 0 {
		t.Error("expected type mismatch error in result.Errors")
	}
	if !errors.Is(result.Errors[0], ErrTypeMismatch) {
		t.Errorf("expected ErrTypeMismatch, got %v", result.Errors[0])
	}
}

func TestBuild_NilNode(t *testing.T) {
	b := NewBuilder(&BuildConfig{})
	target := t.TempDir()
	result, err := b.Build(context.Background(), nil, target)
	if err != nil {
		t.Fatalf("unexpected error for nil root: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
}

func TestBuild_SkipExistingWithVerbose(t *testing.T) {
	root := models.NewNode("proj", "proj", true)
	root.AddChild(models.NewNode("existing.txt", "proj/existing.txt", false))

	target := filepath.Join(t.TempDir(), "proj")
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "existing.txt"), []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}

	b := NewBuilder(&BuildConfig{SkipExisting: true, Verbose: true})
	result, err := b.Build(context.Background(), root, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Skipped) == 0 {
		t.Error("should have skipped existing file")
	}
}
