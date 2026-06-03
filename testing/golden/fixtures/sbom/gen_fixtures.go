//go:build ignore

// Command gen_fixtures documents (and, if ever needed, re-materializes) the
// committed golden-master fixture for the `sbom` category.
//
// The fixture is a FROZEN Go module directory at:
//
//	testing/golden/fixtures/sbom/mod/go.mod
//
// containing a fixed module path and a pinned require set. It is committed and
// must NOT drift: the `sbom` golden snapshots encode the exact purls derived
// from this go.mod (e.g. pkg:golang/github.com/spf13/cobra@v1.10.2), so any
// change to the require set would change the deterministic SBOM bytes and break
// the golden masters by design.
//
// Run by hand (never in CI) only when intentionally rebaselining the fixture:
//
//	go run testing/golden/fixtures/sbom/gen_fixtures.go
//
// after which `task test:golden:update` + `task golden:record` must be re-run.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// frozenGoMod is the canonical, pinned module definition the sbom golden
// snapshots are recorded against. Keep in sync with mod/go.mod (this tool
// writes it; the committed file is the source of truth for the test path).
const frozenGoMod = `module github.com/example/golden-app

go 1.25.0

require (
	github.com/spf13/cobra v1.10.2
	golang.org/x/mod v0.36.0
)
`

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(thisFile)
	modDir := filepath.Join(dir, "mod")
	if err := os.MkdirAll(modDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", modDir, err)
		os.Exit(1)
	}
	path := filepath.Join(modDir, "go.mod")
	if err := os.WriteFile(path, []byte(frozenGoMod), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", path, err)
		os.Exit(1)
	}
	fmt.Println("frozen sbom fixture written to", path)
}
