package model_test

import (
	"testing"

	"github.com/inovacc/omni/pkg/sbom/model"
)

func TestSortNormalizesOrder(t *testing.T) {
	m := &model.SBOM{
		Root: model.Component{Path: "example.com/app", Version: "v1.0.0"},
		Components: []model.Component{
			{Path: "github.com/z/z", Version: "v1.0.0"},
			{Path: "github.com/a/a", Version: "v2.0.0"},
			{Path: "github.com/a/a", Version: "v1.0.0"}, // duplicate path, lower version
		},
	}
	m.Normalize()
	if len(m.Components) != 3 {
		t.Fatalf("len = %d, want 3", len(m.Components))
	}
	// Sorted by (Path, Version): a@v1, a@v2, z@v1.
	want := []string{"github.com/a/a@v1.0.0", "github.com/a/a@v2.0.0", "github.com/z/z@v1.0.0"}
	for i, c := range m.Components {
		if got := c.Path + "@" + c.Version; got != want[i] {
			t.Errorf("Components[%d] = %q, want %q", i, got, want[i])
		}
	}
}

func TestSlugSanitizes(t *testing.T) {
	if got := model.Slug("github.com/spf13/cobra"); got != "github.com-spf13-cobra" {
		t.Errorf("Slug = %q", got)
	}
}
