package collect_test

import (
	"runtime/debug"
	"testing"

	"github.com/inovacc/omni/pkg/sbom/collect"
	"github.com/inovacc/omni/pkg/sbom/model"
)

func TestBinaryFromBuildInfo(t *testing.T) {
	bi := &debug.BuildInfo{
		GoVersion: "go1.25.0",
		Path:      "github.com/inovacc/omni",
		Main:      debug.Module{Path: "github.com/inovacc/omni", Version: "v1.0.0"},
		Deps: []*debug.Module{
			{Path: "github.com/spf13/cobra", Version: "v1.10.2"},
			{
				Path:    "github.com/old/mod",
				Version: "v1.0.0",
				Replace: &debug.Module{Path: "github.com/new/mod", Version: "v2.0.0"},
			},
		},
	}
	sb := collect.Binary(bi)
	if sb.Root.Path != "github.com/inovacc/omni" || sb.Root.Version != "v1.0.0" {
		t.Errorf("root = %+v", sb.Root)
	}
	var sawToolchain, sawReplaced bool
	for _, c := range sb.Components {
		if c.Kind == model.KindToolchain && c.Path == "std" && c.Version == "go1.25.0" {
			sawToolchain = true
		}
		if c.Path == "github.com/new/mod" && c.Version == "v2.0.0" && c.OriginalPath == "github.com/old/mod" {
			sawReplaced = true
		}
	}
	if !sawToolchain {
		t.Error("toolchain component (std@go1.25.0) missing")
	}
	if !sawReplaced {
		t.Error("replace directive not resolved to effective module")
	}
}
