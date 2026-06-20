package repo

import (
	"bytes"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// scaffoldGoModuleWithPkg builds a minimal Go module containing a pkg/ tree with
// exported functions so the API-surface walkers have something to count.
func scaffoldGoModuleWithPkg(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	write := func(rel, content string) {
		p := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("go.mod", "module example.com/scaffold\n\ngo 1.22\n")
	write("main.go", "package main\n\nfunc main() {}\n")

	// pkg/mathx with two exported funcs and one unexported.
	write("pkg/mathx/mathx.go", `package mathx

// Add sums two ints.
func Add(a, b int) int { return a + b }

// Sub subtracts.
func Sub(a, b int) int { return a - b }

func internalHelper() int { return 0 }
`)

	// pkg/strutil with one exported func + a test file that must be ignored.
	write("pkg/strutil/strutil.go", `package strutil

// Upper does nothing useful.
func Upper(s string) string { return s }
`)
	write("pkg/strutil/strutil_test.go", `package strutil

import "testing"

func TestNothing(t *testing.T) {}
`)

	// A pkg dir with no exported funcs should be skipped from the API surface.
	write("pkg/empty/empty.go", "package empty\n\nvar x = 1\n")

	return dir
}

func TestCountExportedFuncs(t *testing.T) {
	dir := scaffoldGoModuleWithPkg(t)
	fset := token.NewFileSet()

	t.Run("counts exported only, ignores tests", func(t *testing.T) {
		got := countExportedFuncs(fset, filepath.Join(dir, "pkg", "mathx"))
		if got != 2 {
			t.Errorf("countExportedFuncs(mathx) = %d, want 2", got)
		}
	})

	t.Run("single exported func", func(t *testing.T) {
		got := countExportedFuncs(fset, filepath.Join(dir, "pkg", "strutil"))
		if got != 1 {
			t.Errorf("countExportedFuncs(strutil) = %d, want 1 (test funcs ignored)", got)
		}
	})

	t.Run("no exported funcs", func(t *testing.T) {
		got := countExportedFuncs(fset, filepath.Join(dir, "pkg", "empty"))
		if got != 0 {
			t.Errorf("countExportedFuncs(empty) = %d, want 0", got)
		}
	})

	t.Run("nonexistent dir returns 0", func(t *testing.T) {
		got := countExportedFuncs(fset, filepath.Join(dir, "pkg", "does-not-exist"))
		if got != 0 {
			t.Errorf("countExportedFuncs(missing) = %d, want 0", got)
		}
	})
}

func TestAnalyzeAPISurface(t *testing.T) {
	dir := scaffoldGoModuleWithPkg(t)

	t.Run("aggregates packages with exported funcs", func(t *testing.T) {
		apis := analyzeAPISurface(dir)

		byPkg := make(map[string]int)
		for _, a := range apis {
			byPkg[a.Package] = a.ExportedFuncs
		}

		if byPkg["pkg/mathx"] != 2 {
			t.Errorf("pkg/mathx exported = %d, want 2 (%+v)", byPkg["pkg/mathx"], apis)
		}

		if byPkg["pkg/strutil"] != 1 {
			t.Errorf("pkg/strutil exported = %d, want 1 (%+v)", byPkg["pkg/strutil"], apis)
		}

		if _, ok := byPkg["pkg/empty"]; ok {
			t.Errorf("pkg/empty should be omitted (0 exported funcs): %+v", apis)
		}
	})

	t.Run("no pkg dir returns nil", func(t *testing.T) {
		bare := t.TempDir()
		if apis := analyzeAPISurface(bare); apis != nil {
			t.Errorf("analyzeAPISurface(no pkg) = %v, want nil", apis)
		}
	})
}

// TestRunAnalyze_APISection drives RunAnalyze with the api section so
// analyzeAPISurface + formatAnalyzeMarkdown's API-surface table both execute.
func TestRunAnalyze_APISection(t *testing.T) {
	dir := scaffoldGoModuleWithPkg(t)

	var buf bytes.Buffer
	if err := RunAnalyze(&buf, []string{dir}, Options{Sections: "api"}); err != nil {
		t.Fatalf("RunAnalyze(api) error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "## API Surface") {
		t.Errorf("expected API Surface section:\n%s", out)
	}

	if !strings.Contains(out, "pkg/mathx") {
		t.Errorf("expected pkg/mathx row in API surface:\n%s", out)
	}
}

// TestRunAnalyze_DepsSection drives the deps markdown branch (writeDeps) for a
// Go module.
func TestRunAnalyze_DepsSection(t *testing.T) {
	dir := scaffoldGoModuleWithPkg(t)

	var buf bytes.Buffer
	if err := RunAnalyze(&buf, []string{dir}, Options{Sections: "deps"}); err != nil {
		t.Fatalf("RunAnalyze(deps) error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "## Dependencies") {
		t.Errorf("expected Dependencies section:\n%s", out)
	}

	// The Go module name should be rendered by writeDeps.
	if !strings.Contains(out, "example.com/scaffold") {
		t.Errorf("expected module name in deps output:\n%s", out)
	}
}

// TestRunAnalyze_AllSectionsMarkdown exercises a broad markdown render to push
// formatAnalyzeMarkdown coverage (overview/tree/keys/api/tests/ci branches).
func TestRunAnalyze_AllSectionsMarkdown(t *testing.T) {
	dir := scaffoldGoModuleWithPkg(t)

	// Add files that trigger keys, tests, and ci sections.
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Scaffold\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "foo_test.go"), []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte("version: 3\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := RunAnalyze(&buf, []string{dir}, Options{
		Sections: "overview,tree,keys,deps,api,tests,ci",
	}); err != nil {
		t.Fatalf("RunAnalyze(all) error = %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"# Repository Context:",
		"## Overview",
		"## Directory Tree",
		"## Key Files",
		"## API Surface",
		"## Test Patterns",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("markdown output missing %q:\n%s", want, out)
		}
	}
}

func TestResolvePath_Variants(t *testing.T) {
	dir := t.TempDir()

	t.Run("existing dir returns abs", func(t *testing.T) {
		got, err := resolvePath(dir)
		if err != nil {
			t.Fatalf("resolvePath(dir) error = %v", err)
		}

		if !filepath.IsAbs(got) {
			t.Errorf("resolvePath(dir) = %q, want absolute", got)
		}
	})

	t.Run("file returns parent dir", func(t *testing.T) {
		file := filepath.Join(dir, "file.txt")
		if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}

		got, err := resolvePath(file)
		if err != nil {
			t.Fatalf("resolvePath(file) error = %v", err)
		}

		absDir, _ := filepath.Abs(dir)
		if got != absDir {
			t.Errorf("resolvePath(file) = %q, want parent %q", got, absDir)
		}
	})
}
