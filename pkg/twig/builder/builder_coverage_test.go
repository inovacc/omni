package builder

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/inovacc/omni/pkg/twig/models"
)

func TestValidatePath(t *testing.T) {
	base := t.TempDir()

	tests := []struct {
		name    string
		path    string
		wantErr error // nil means no error expected
	}{
		{
			name:    "valid path under base",
			path:    filepath.Join(base, "sub", "file.txt"),
			wantErr: nil,
		},
		{
			name:    "path traversal outside base",
			path:    filepath.Join(base, "..", "escape.txt"),
			wantErr: ErrPathTraversal,
		},
		{
			name:    "invalid characters in name",
			path:    filepath.Join(base, "bad<name>.txt"),
			wantErr: ErrInvalidCharacters,
		},
	}

	b := &Builder{config: DefaultBuildConfig()}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := b.validatePath(tt.path, base)
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("validatePath(%q) = %v, want nil", tt.path, err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("validatePath(%q) = %v, want %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

// TestBuild_InvalidChildNameRecordsError drives the validatePath-failure branch
// inside buildNode: a child whose name contains invalid characters is recorded
// in result.Errors rather than aborting the whole build.
func TestBuild_InvalidChildNameRecordsError(t *testing.T) {
	root := models.NewNode("proj", "proj", true)
	good := models.NewNode("ok.txt", "proj/ok.txt", false)
	bad := models.NewNode("bad?name.txt", "proj/bad?name.txt", false)
	root.AddChild(good)
	root.AddChild(bad)

	target := filepath.Join(t.TempDir(), "proj")
	b := NewBuilder(&BuildConfig{})

	result, err := b.Build(context.Background(), root, target)
	if err != nil {
		t.Fatalf("Build returned fatal error: %v", err)
	}

	foundInvalid := false
	for _, e := range result.Errors {
		if errors.Is(e, ErrInvalidCharacters) {
			foundInvalid = true
			break
		}
	}
	if !foundInvalid {
		t.Errorf("expected ErrInvalidCharacters in result.Errors, got %v", result.Errors)
	}
}

// TestBuild_NestedDirsAndFiles exercises the recursive directory build path
// and the parent-directory creation for nested files.
func TestBuild_NestedDirsAndFiles(t *testing.T) {
	root := models.NewNode("app", "app", true)
	pkg := models.NewNode("pkg", "app/pkg", true)
	sub := models.NewNode("util", "app/pkg/util", true)
	pkg.AddChild(sub)
	sub.AddChild(models.NewNode("util.go", "app/pkg/util/util.go", false))
	root.AddChild(pkg)

	target := filepath.Join(t.TempDir(), "app")
	b := NewBuilder(&BuildConfig{Verbose: true})

	result, err := b.Build(context.Background(), root, target)
	if err != nil {
		t.Fatalf("Build error: %v", err)
	}
	if len(result.Created) == 0 {
		t.Error("expected created items for nested structure")
	}
}

func TestNodeType(t *testing.T) {
	b := &Builder{config: DefaultBuildConfig()}
	if got := b.nodeType(true); got != "directory" {
		t.Errorf("nodeType(true) = %q, want directory", got)
	}
	if got := b.nodeType(false); got != "file" {
		t.Errorf("nodeType(false) = %q, want file", got)
	}
}
