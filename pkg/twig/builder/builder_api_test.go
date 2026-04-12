package builder_test

import (
	"context"
	"testing"

	"github.com/inovacc/omni/pkg/twig/builder"
	"github.com/inovacc/omni/pkg/twig/models"
)

func TestDefaultBuildConfig_API(t *testing.T) {
	cfg := builder.DefaultBuildConfig()
	if cfg == nil {
		t.Fatal("DefaultBuildConfig() returned nil")
	}
	if cfg.DryRun {
		t.Error("DefaultBuildConfig().DryRun should be false")
	}
}

func TestNewBuilder_API(t *testing.T) {
	b := builder.NewBuilder(nil)
	if b == nil {
		t.Fatal("NewBuilder(nil) returned nil")
	}
}

func TestBuild_DryRun_API(t *testing.T) {
	tmp := t.TempDir()

	root := models.NewNode("project", "project", true)
	root.AddChild(models.NewNode("README.md", "project/README.md", false))

	cfg := builder.DefaultBuildConfig()
	cfg.DryRun = true

	b := builder.NewBuilder(cfg)
	result, err := b.Build(context.Background(), root, tmp)
	if err != nil {
		t.Fatalf("Build(DryRun) error = %v", err)
	}
	if result == nil {
		t.Fatal("Build(DryRun) returned nil result")
	}
	if !result.DryRun {
		t.Error("Build(DryRun) result.DryRun should be true")
	}
}
