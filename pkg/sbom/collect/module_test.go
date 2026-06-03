package collect_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/inovacc/omni/pkg/sbom/collect"
	"github.com/inovacc/omni/pkg/sbom/model"
)

const goMod = `module github.com/example/app

go 1.25.0

require (
	github.com/spf13/cobra v1.10.2
	golang.org/x/mod v0.36.0 // indirect
)

require github.com/single/dep v1.2.3
`

func writeMod(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestModuleDir(t *testing.T) {
	dir := t.TempDir()
	writeMod(t, dir, goMod)
	sb, err := collect.ModuleDir(dir)
	if err != nil {
		t.Fatalf("ModuleDir: %v", err)
	}
	if sb.Root.Path != "github.com/example/app" || sb.Root.Kind != model.KindRoot {
		t.Errorf("root = %+v", sb.Root)
	}
	sb.Normalize()
	got := map[string]string{}
	for _, c := range sb.Components {
		got[c.Path] = c.Version
	}
	for path, ver := range map[string]string{
		"github.com/spf13/cobra": "v1.10.2",
		"golang.org/x/mod":       "v0.36.0",
		"github.com/single/dep":  "v1.2.3",
	} {
		if got[path] != ver {
			t.Errorf("dep %s = %q, want %q", path, got[path], ver)
		}
	}
}

func TestModuleDirMissing(t *testing.T) {
	if _, err := collect.ModuleDir(t.TempDir()); err == nil {
		t.Error("expected error for missing go.mod")
	}
}
