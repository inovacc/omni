# Scaffolding Refactor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Split `internal/cli/generate/` into domain-specific subpackages under `internal/cli/scaffolding/`, rename CLI from `generate` to `scaffold`, add normalized cmdtree + aicontext templates to cobra scaffold.

**Architecture:** Each generator domain (cobra, handler, repository, testgen) becomes its own Go package under `internal/cli/scaffolding/<domain>/`. Shared helpers live in the parent `scaffolding` package. Templates stay nested under each domain. Two new template files (cmdtree.go, aicontext.go) get added to cobra templates, generated conditionally via flags.

**Tech Stack:** Go, Cobra, text/template, gopkg.in/yaml.v3

---

### Task 1: Create scaffolding package with shared helpers

**Files:**
- Create: `internal/cli/scaffolding/shared.go`
- Create: `internal/cli/scaffolding/shared_test.go`

**Step 1: Write the failing test**

```go
// internal/cli/scaffolding/shared_test.go
package scaffolding

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.txt")

	err := WriteTemplate(path, "Hello {{.Name}}", struct{ Name string }{"World"})
	if err != nil {
		t.Fatalf("WriteTemplate() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if string(data) != "Hello World" {
		t.Errorf("got %q, want %q", string(data), "Hello World")
	}
}

func TestWriteLicense(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "LICENSE")

	err := WriteLicense(path, "MIT", "Test Author")
	if err != nil {
		t.Fatalf("WriteLicense() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("LICENSE file is empty")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/cli/scaffolding/ -v`
Expected: FAIL — package doesn't exist yet

**Step 3: Write minimal implementation**

Extract from `internal/cli/generate/generate.go` lines 375-408 into:

```go
// internal/cli/scaffolding/shared.go
package scaffolding

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	cobratpl "github.com/inovacc/omni/internal/cli/scaffolding/cobra/templates"
)

// Options configures scaffold command behavior
type Options struct {
	JSON bool // --json: output as JSON
}

// WriteTemplate renders a Go text/template to a file.
func WriteTemplate(path string, tmpl string, data any) error {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	defer func() { _ = f.Close() }()

	return t.Execute(f, data)
}

// WriteLicense writes a LICENSE file for the given type and author.
func WriteLicense(path, licenseType, author string) error {
	year := time.Now().Year()

	var content string

	switch strings.ToLower(licenseType) {
	case "mit":
		content = fmt.Sprintf(cobratpl.MITLicense, year, author)
	case "apache-2.0", "apache":
		content = fmt.Sprintf(cobratpl.ApacheLicense, year, author)
	case "bsd-3", "bsd":
		content = fmt.Sprintf(cobratpl.BSDLicense, year, author)
	default:
		return fmt.Errorf("unknown license type: %s", licenseType)
	}

	return os.WriteFile(path, []byte(content), 0644)
}
```

Note: `shared.go` will have a circular import on cobratpl initially. To break it, the license constants should stay in cobra/templates but WriteLicense can reference them. If circular, move license constants to `scaffolding/licenses.go` instead — the implementer should check and adapt.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/cli/scaffolding/ -v`
Expected: PASS

**Step 5: Commit**

```
feat(scaffold): add shared helpers package
```

---

### Task 2: Move cobra domain to scaffolding/cobra

**Files:**
- Create: `internal/cli/scaffolding/cobra/cobra.go` (from `generate/generate.go` lines 22-373)
- Create: `internal/cli/scaffolding/cobra/config.go` (copy `generate/config.go`, change package name)
- Move: `internal/cli/generate/templates/cobra/templates.go` → `internal/cli/scaffolding/cobra/templates/templates.go`
- Create: `internal/cli/scaffolding/cobra/cobra_test.go` (from `generate/generate_test.go` — TestRunCobraInit, TestRunCobraAdd, TestLicenses, TestTaskfileContent, TestGitignoreContent, TestFullMode, TestServiceMode, TestEditorConfigContent, TestCobraConfig)

**Step 1: Create the cobra package**

```go
// internal/cli/scaffolding/cobra/cobra.go
package cobra

// Package name is "cobra" — importers use scaffoldcobra or similar alias.
// Contains: CobraInitOptions, CobraAddOptions, InitResult, AddResult,
// RunCobraInit, RunCobraAdd
//
// Copy from generate.go lines 22-373.
// Change import of cobratpl from:
//   "github.com/inovacc/omni/internal/cli/generate/templates/cobra"
// To:
//   "github.com/inovacc/omni/internal/cli/scaffolding/cobra/templates"
//
// Change calls to writeTemplate/writeLicense to use scaffolding.WriteTemplate/scaffolding.WriteLicense:
//   import "github.com/inovacc/omni/internal/cli/scaffolding"
//   scaffolding.WriteTemplate(...)
//   scaffolding.WriteLicense(...)
//
// Move Options type reference to scaffolding.Options
```

**Step 2: Move config.go**

Copy `internal/cli/generate/config.go` → `internal/cli/scaffolding/cobra/config.go`, change `package generate` to `package cobra`.

**Step 3: Move templates**

Move `internal/cli/generate/templates/cobra/` → `internal/cli/scaffolding/cobra/templates/`

**Step 4: Create cobra_test.go**

Copy tests from `generate_test.go` (lines 12-1053, only cobra-related tests: TestRunCobraInit through TestCobraConfig). Update package to `package cobra`, update function references from `RunCobraInit` → same (now in local package), update `CobraInitOptions` → same, etc.

**Step 5: Run tests**

Run: `go test ./internal/cli/scaffolding/cobra/ -v`
Expected: PASS — all cobra tests pass

**Step 6: Commit**

```
refactor(scaffold): move cobra domain to scaffolding/cobra
```

---

### Task 3: Move handler domain to scaffolding/handler

**Files:**
- Create: `internal/cli/scaffolding/handler/handler.go` (from `generate/generate.go` lines 410-532)
- Move: `internal/cli/generate/templates/handler/` → `internal/cli/scaffolding/handler/templates/`
- Create: `internal/cli/scaffolding/handler/handler_test.go`

**Step 1: Extract handler code**

Copy `RunHandlerInit`, `HandlerOptions`, `HandlerResult` from `generate.go` lines 410-532. Change package to `handler`. Update template import path. Use `scaffolding.WriteTemplate`.

**Step 2: Write a basic test**

```go
package handler

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/inovacc/omni/internal/cli/scaffolding"
)

func TestRunHandlerInit(t *testing.T) {
	tmpDir := t.TempDir()
	dir := filepath.Join(tmpDir, "handlers")

	var buf bytes.Buffer
	err := RunHandlerInit(&buf, "user", HandlerOptions{
		Dir:       dir,
		Framework: "stdlib",
	}, scaffolding.Options{})
	if err != nil {
		t.Fatalf("RunHandlerInit() error = %v", err)
	}

	handlerPath := filepath.Join(dir, "user.go")
	if _, err := os.Stat(handlerPath); os.IsNotExist(err) {
		t.Error("Expected user.go to be created")
	}
}
```

**Step 3: Run tests**

Run: `go test ./internal/cli/scaffolding/handler/ -v`
Expected: PASS

**Step 4: Commit**

```
refactor(scaffold): move handler domain to scaffolding/handler
```

---

### Task 4: Move repository domain to scaffolding/repository

**Files:**
- Create: `internal/cli/scaffolding/repository/repository.go` (from `generate/generate.go` lines 534-666)
- Move: `internal/cli/generate/templates/repository/` → `internal/cli/scaffolding/repository/templates/`
- Create: `internal/cli/scaffolding/repository/repository_test.go`

**Step 1: Extract repository code**

Copy `RunRepositoryInit`, `RepositoryOptions`, `RepositoryResult`. Change package, imports, use `scaffolding.WriteTemplate`.

**Step 2: Write basic test**

```go
package repository

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/inovacc/omni/internal/cli/scaffolding"
)

func TestRunRepositoryInit(t *testing.T) {
	tmpDir := t.TempDir()
	dir := filepath.Join(tmpDir, "repo")

	var buf bytes.Buffer
	err := RunRepositoryInit(&buf, "user", RepositoryOptions{
		Dir: dir,
		DB:  "postgres",
	}, scaffolding.Options{})
	if err != nil {
		t.Fatalf("RunRepositoryInit() error = %v", err)
	}

	repoPath := filepath.Join(dir, "user.go")
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		t.Error("Expected user.go to be created")
	}
}
```

**Step 3: Run tests**

Run: `go test ./internal/cli/scaffolding/repository/ -v`
Expected: PASS

**Step 4: Commit**

```
refactor(scaffold): move repository domain to scaffolding/repository
```

---

### Task 5: Move testgen domain to scaffolding/testgen

**Files:**
- Create: `internal/cli/scaffolding/testgen/testgen.go` (from `generate/generate.go` lines 668-859)
- Move: `internal/cli/generate/templates/test/` → `internal/cli/scaffolding/testgen/templates/`
- Create: `internal/cli/scaffolding/testgen/testgen_test.go`

**Step 1: Extract testgen code**

Copy `RunTestInit`, `TestOptions`, `TestResult`, `exprToString`. Change package to `testgen`, update imports.

**Step 2: Write basic test**

Create a temporary `.go` file, run `RunTestInit` against it, verify `_test.go` output exists.

**Step 3: Run tests**

Run: `go test ./internal/cli/scaffolding/testgen/ -v`
Expected: PASS

**Step 4: Commit**

```
refactor(scaffold): move testgen domain to scaffolding/testgen
```

---

### Task 6: Add cmdtree template to cobra scaffold

**Files:**
- Create: `internal/cli/scaffolding/cobra/templates/cmdtree.go`

**Step 1: Write the cmdtree template constant**

Normalize from glix's `cmd/cmdtree.go` (442 lines, most complete version). The template must:
- Use `{{.AppName}}` nowhere (cmdtree is app-agnostic, reads from rootCmd)
- Include `FlagDetail`, `CommandDetail` with `GlobalFlags`
- Include `collectPersistentFlags` (glix pattern)
- Use `LocalFlags()` to avoid persistent flag duplication
- Include all 4 modes: verbose (default), brief, JSON, single command
- Use ASCII tree chars (`+--`, `\--`)
- Use `cmd.OutOrStdout()` (not `os.Stdout`) for testability
- Template renders as a complete `cmd/cmdtree.go` file

```go
// internal/cli/scaffolding/cobra/templates/cmdtree.go
package templates

// CmdtreeTemplate generates cmd/cmdtree.go for scaffolded projects.
// This is the normalized, most-complete version derived from glix.
const CmdtreeTemplate = `package cmd
...
`
```

The full template content is the glix cmdtree.go source with these changes:
- Remove glix-specific imports (only `bytes`, `encoding/json`, `fmt`, `io`, `os`, `strings`, `cobra`, `pflag`)
- Use `cmd.OutOrStdout()` instead of `os.Stdout` in `printJSONTree` and `printSingleCommand`
- Keep `GlobalFlags` on `CommandDetail`
- Keep `collectPersistentFlags`

**Step 2: Run existing tests (no new tests needed yet — template is a const string)**

Run: `go test ./internal/cli/scaffolding/cobra/... -v`
Expected: PASS (compiles)

**Step 3: Commit**

```
feat(scaffold): add normalized cmdtree template
```

---

### Task 7: Add aicontext template to cobra scaffold

**Files:**
- Create: `internal/cli/scaffolding/cobra/templates/aicontext.go`

**Step 1: Write the aicontext template constant**

Normalize from glix's `cmd/aicontext.go` (360 lines, most complete). The template must:
- Use `{{.AppName}}` and `{{.Description}}` for the app name/desc in output
- Include 3 output modes: markdown (default), JSON (`--json`), compact (`--compact`)
- Include global persistent flags section
- Use `cmd.OutOrStdout()` for testability
- Include stub category map with `// TODO: customize for your app` comment
- Include stub structure map with `// TODO: customize for your app` comment
- Template renders as a complete `cmd/aicontext.go` file

```go
// internal/cli/scaffolding/cobra/templates/aicontext.go
package templates

// AIContextTemplate generates cmd/aicontext.go for scaffolded projects.
// This is the normalized version derived from glix (most complete).
// The generated file includes TODO stubs for category and structure maps.
const AIContextTemplate = `package cmd
...
`
```

Key differences from glix original:
- Replace hardcoded `"glix"` with template `{{.AppName}}`
- Replace hardcoded description with `{{.Description}}`
- Replace hardcoded `Version` var reference with a placeholder or `"dev"` default
- Replace glix-specific usage patterns section with generic placeholder
- Category map has 3-4 example entries + TODO comment
- Structure map has generic entries (`cmd/`, `internal/`, `pkg/`) + TODO comment

**Step 2: Verify compilation**

Run: `go test ./internal/cli/scaffolding/cobra/... -v`
Expected: PASS

**Step 3: Commit**

```
feat(scaffold): add normalized aicontext template
```

---

### Task 8: Wire cmdtree + aicontext into RunCobraInit

**Files:**
- Modify: `internal/cli/scaffolding/cobra/cobra.go` — add AIContext field to CobraInitOptions, generate cmd/cmdtree.go always, generate cmd/aicontext.go when AIContext=true
- Modify: `internal/cli/scaffolding/cobra/templates/templates.go` — add AIContext to TemplateData

**Step 1: Update TemplateData**

Add to `TemplateData` struct in `templates/templates.go`:

```go
AIContext bool // Include aicontext command
```

**Step 2: Update CobraInitOptions**

Add to `CobraInitOptions` in `cobra.go`:

```go
AIContext bool // --aicontext: include aicontext command (default: false)
```

**Step 3: Add generation logic in RunCobraInit**

After generating `cmd/version.go`, add:

```go
// Generate cmd/cmdtree.go (always included)
cmdtreePath := filepath.Join(dir, "cmd", "cmdtree.go")
if err := scaffolding.WriteTemplate(cmdtreePath, cobratpl.CmdtreeTemplate, tplData); err != nil {
    return fmt.Errorf("scaffold: failed to create cmd/cmdtree.go: %w", err)
}
filesCreated = append(filesCreated, "cmd/cmdtree.go")

// Generate cmd/aicontext.go (optional)
if opts.AIContext {
    aicontextPath := filepath.Join(dir, "cmd", "aicontext.go")
    if err := scaffolding.WriteTemplate(aicontextPath, cobratpl.AIContextTemplate, tplData); err != nil {
        return fmt.Errorf("scaffold: failed to create cmd/aicontext.go: %w", err)
    }
    filesCreated = append(filesCreated, "cmd/aicontext.go")
}
```

**Step 4: Write failing test**

```go
func TestRunCobraInitWithCmdtree(t *testing.T) {
    tmpDir := t.TempDir()
    appDir := filepath.Join(tmpDir, "myapp")
    var buf bytes.Buffer

    err := RunCobraInit(&buf, appDir, CobraInitOptions{
        Module: "github.com/test/myapp",
    }, scaffolding.Options{})
    if err != nil {
        t.Fatal(err)
    }

    // cmdtree.go should always exist
    if _, err := os.Stat(filepath.Join(appDir, "cmd", "cmdtree.go")); os.IsNotExist(err) {
        t.Error("cmd/cmdtree.go not created")
    }

    // aicontext.go should NOT exist by default
    if _, err := os.Stat(filepath.Join(appDir, "cmd", "aicontext.go")); err == nil {
        t.Error("cmd/aicontext.go should not be created without --aicontext flag")
    }
}

func TestRunCobraInitWithAIContext(t *testing.T) {
    tmpDir := t.TempDir()
    appDir := filepath.Join(tmpDir, "myapp")
    var buf bytes.Buffer

    err := RunCobraInit(&buf, appDir, CobraInitOptions{
        Module:    "github.com/test/myapp",
        AIContext: true,
    }, scaffolding.Options{})
    if err != nil {
        t.Fatal(err)
    }

    if _, err := os.Stat(filepath.Join(appDir, "cmd", "aicontext.go")); os.IsNotExist(err) {
        t.Error("cmd/aicontext.go not created with --aicontext flag")
    }
}
```

**Step 5: Run tests**

Run: `go test ./internal/cli/scaffolding/cobra/ -v`
Expected: PASS

**Step 6: Commit**

```
feat(scaffold): wire cmdtree/aicontext into cobra init
```

---

### Task 9: Add `scaffold cobra add-tools` subcommand

**Files:**
- Modify: `internal/cli/scaffolding/cobra/cobra.go` — add RunCobraAddTools function
- Add test in: `internal/cli/scaffolding/cobra/cobra_test.go`

**Step 1: Write failing test**

```go
func TestRunCobraAddTools(t *testing.T) {
    tmpDir := t.TempDir()

    // Create minimal Cobra project structure
    cmdDir := filepath.Join(tmpDir, "cmd")
    _ = os.MkdirAll(cmdDir, 0755)
    _ = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module github.com/test/existing\n\ngo 1.21\n"), 0644)
    _ = os.WriteFile(filepath.Join(cmdDir, "root.go"), []byte("package cmd\n"), 0644)

    var buf bytes.Buffer

    err := RunCobraAddTools(&buf, tmpDir, AddToolsOptions{AIContext: false}, scaffolding.Options{})
    if err != nil {
        t.Fatal(err)
    }

    // cmdtree.go should exist
    if _, err := os.Stat(filepath.Join(cmdDir, "cmdtree.go")); os.IsNotExist(err) {
        t.Error("cmd/cmdtree.go not created")
    }

    // aicontext.go should NOT exist
    if _, err := os.Stat(filepath.Join(cmdDir, "aicontext.go")); err == nil {
        t.Error("cmd/aicontext.go should not be created without --aicontext")
    }
}

func TestRunCobraAddToolsWithAIContext(t *testing.T) {
    tmpDir := t.TempDir()

    cmdDir := filepath.Join(tmpDir, "cmd")
    _ = os.MkdirAll(cmdDir, 0755)
    _ = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module github.com/test/existing\n\ngo 1.21\n"), 0644)
    _ = os.WriteFile(filepath.Join(cmdDir, "root.go"), []byte("package cmd\n"), 0644)

    var buf bytes.Buffer

    err := RunCobraAddTools(&buf, tmpDir, AddToolsOptions{AIContext: true}, scaffolding.Options{})
    if err != nil {
        t.Fatal(err)
    }

    if _, err := os.Stat(filepath.Join(cmdDir, "aicontext.go")); os.IsNotExist(err) {
        t.Error("cmd/aicontext.go not created with --aicontext")
    }
}
```

**Step 2: Implement RunCobraAddTools**

```go
// AddToolsOptions configures add-tools behavior
type AddToolsOptions struct {
    AIContext bool // --aicontext: include aicontext command
}

// AddToolsResult represents the result of add-tools
type AddToolsResult struct {
    Status       string   `json:"status"`
    FilesCreated []string `json:"files_created"`
}

// RunCobraAddTools adds cmdtree and optionally aicontext to an existing Cobra project.
func RunCobraAddTools(w io.Writer, dir string, opts AddToolsOptions, genOpts scaffolding.Options) error {
    // Verify this is a Cobra project
    cmdDir := filepath.Join(dir, "cmd")
    if _, err := os.Stat(cmdDir); os.IsNotExist(err) {
        return fmt.Errorf("scaffold: cmd directory not found, is this a Cobra project?")
    }

    // Read module from go.mod
    goModPath := filepath.Join(dir, "go.mod")
    modData, err := os.ReadFile(goModPath)
    if err != nil {
        return fmt.Errorf("scaffold: failed to read go.mod: %w", err)
    }

    module := parseModuleName(modData)
    if module == "" {
        return fmt.Errorf("scaffold: could not parse module from go.mod")
    }

    appName := filepath.Base(module)
    tplData := cobratpl.TemplateData{
        Module:  module,
        AppName: appName,
    }

    var filesCreated []string

    // Always generate cmdtree
    cmdtreePath := filepath.Join(cmdDir, "cmdtree.go")
    if _, err := os.Stat(cmdtreePath); err == nil {
        return fmt.Errorf("scaffold: cmd/cmdtree.go already exists")
    }
    if err := scaffolding.WriteTemplate(cmdtreePath, cobratpl.CmdtreeTemplate, tplData); err != nil {
        return fmt.Errorf("scaffold: failed to create cmd/cmdtree.go: %w", err)
    }
    filesCreated = append(filesCreated, "cmd/cmdtree.go")

    // Optionally generate aicontext
    if opts.AIContext {
        aiPath := filepath.Join(cmdDir, "aicontext.go")
        if _, err := os.Stat(aiPath); err == nil {
            return fmt.Errorf("scaffold: cmd/aicontext.go already exists")
        }
        if err := scaffolding.WriteTemplate(aiPath, cobratpl.AIContextTemplate, tplData); err != nil {
            return fmt.Errorf("scaffold: failed to create cmd/aicontext.go: %w", err)
        }
        filesCreated = append(filesCreated, "cmd/aicontext.go")
    }

    if genOpts.JSON {
        return json.NewEncoder(w).Encode(AddToolsResult{
            Status:       "created",
            FilesCreated: filesCreated,
        })
    }

    _, _ = fmt.Fprintln(w, "Added tools to existing project:")
    for _, f := range filesCreated {
        _, _ = fmt.Fprintf(w, "  - %s\n", f)
    }
    return nil
}

// parseModuleName extracts the module name from go.mod content.
func parseModuleName(data []byte) string {
    for _, line := range strings.Split(string(data), "\n") {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "module ") {
            return strings.TrimSpace(strings.TrimPrefix(line, "module"))
        }
    }
    return ""
}
```

**Step 3: Run tests**

Run: `go test ./internal/cli/scaffolding/cobra/ -v`
Expected: PASS

**Step 4: Commit**

```
feat(scaffold): add cobra add-tools subcommand
```

---

### Task 10: Create cmd/scaffold.go (Cobra wiring)

**Files:**
- Create: `cmd/scaffold.go`

**Step 1: Write cmd/scaffold.go**

Copy the full content of `cmd/generate.go`, then:

1. Rename all variables: `generateCmd` → `scaffoldCmd`, `generateCobraCmd` → `scaffoldCobraCmd`, etc.
2. Change `Use: "generate"` → `Use: "scaffold"`
3. Update all import paths from `"github.com/inovacc/omni/internal/cli/generate"` to the new packages:
   - `scaffoldcobra "github.com/inovacc/omni/internal/cli/scaffolding/cobra"`
   - `"github.com/inovacc/omni/internal/cli/scaffolding/handler"`
   - `"github.com/inovacc/omni/internal/cli/scaffolding/repository"`
   - `"github.com/inovacc/omni/internal/cli/scaffolding/testgen"`
   - `"github.com/inovacc/omni/internal/cli/scaffolding"`
4. Update all type references: `generate.CobraInitOptions` → `scaffoldcobra.CobraInitOptions`, etc.
5. Update `generate.Options{JSON: jsonOutput}` → `scaffolding.Options{JSON: jsonOutput}`
6. Add `--aicontext` flag to `scaffoldCobraInitCmd`
7. Add `scaffoldCobraAddToolsCmd` for `add-tools` subcommand
8. In `init()`: `rootCmd.AddCommand(scaffoldCmd)` and wire all subcommands

Add the `--aicontext` flag:

```go
scaffoldCobraInitCmd.Flags().Bool("aicontext", false, "include aicontext command for AI coding agents")
```

And in the RunE, read it:

```go
aicontext, _ := cmd.Flags().GetBool("aicontext")
opts := scaffoldcobra.CobraInitOptions{
    // ... existing fields ...
    AIContext: aicontext,
}
```

Add the add-tools command:

```go
var scaffoldCobraAddToolsCmd = &cobra.Command{
    Use:   "add-tools",
    Short: "Add cmdtree and aicontext to an existing Cobra project",
    Long: `Add developer tools (cmdtree, aicontext) to an existing Cobra CLI project.

cmdtree is always added. aicontext is added with --aicontext flag.

Examples:
  omni scaffold cobra add-tools
  omni scaffold cobra add-tools --aicontext
  omni scaffold cobra add-tools --dir /path/to/project`,
    RunE: func(cmd *cobra.Command, args []string) error {
        jsonOutput, _ := cmd.Flags().GetBool("json")
        aicontext, _ := cmd.Flags().GetBool("aicontext")
        dir, _ := cmd.Flags().GetString("dir")

        if dir == "" {
            dir, _ = os.Getwd()
        }

        return scaffoldcobra.RunCobraAddTools(cmd.OutOrStdout(), dir, scaffoldcobra.AddToolsOptions{
            AIContext: aicontext,
        }, scaffolding.Options{JSON: jsonOutput})
    },
}
```

In `init()` add:

```go
scaffoldCobraCmd.AddCommand(scaffoldCobraAddToolsCmd)
scaffoldCobraAddToolsCmd.Flags().Bool("aicontext", false, "include aicontext command")
scaffoldCobraAddToolsCmd.Flags().String("dir", "", "project directory (defaults to current directory)")
```

**Step 2: Verify build**

Run: `go build ./...`
Expected: Compiles without errors

**Step 3: Commit**

```
feat(scaffold): create cmd/scaffold.go with full CLI wiring
```

---

### Task 11: Remove old generate package and cmd/generate.go

**Files:**
- Delete: `cmd/generate.go`
- Delete: `internal/cli/generate/generate.go`
- Delete: `internal/cli/generate/generate_test.go`
- Delete: `internal/cli/generate/config.go`
- Delete: `internal/cli/generate/templates/cobra/templates.go`
- Delete: `internal/cli/generate/templates/handler/templates.go`
- Delete: `internal/cli/generate/templates/repository/templates.go`
- Delete: `internal/cli/generate/templates/test/templates.go`

**Step 1: Search for any remaining references to the old package**

Run: `grep -r "internal/cli/generate" --include="*.go" .`

If any references remain, update them to point to the new scaffolding packages.

**Step 2: Delete old files**

Remove the entire `internal/cli/generate/` directory and `cmd/generate.go`.

**Step 3: Run full test suite**

Run: `go test ./... 2>&1 | head -50`
Expected: All tests pass, no import errors

**Step 4: Run build**

Run: `go build ./...`
Expected: Compiles

**Step 5: Commit**

```
refactor(scaffold): remove old generate package
```

---

### Task 12: Update CLAUDE.md and docs

**Files:**
- Modify: `CLAUDE.md` — update command categories table, code patterns section, directory structure
- Modify: `cmd/cmdtree.go` — update category map if aicontext references `generate`

**Step 1: Update CLAUDE.md**

- Change `"generate"` references to `"scaffold"` in command categories
- Update directory structure to show `internal/cli/scaffolding/` layout
- Update "Add New Command" section if it references generate
- Update code generation examples: `omni generate` → `omni scaffold`

**Step 2: Update golden test YAML if applicable**

Check `testing/golden/golden_tests.yaml` and `tools/golden/golden_tests.yaml` for any `generate` test cases and update to `scaffold`.

**Step 3: Run golden tests**

Run: `task test:golden` (if golden tests reference generate commands)

**Step 4: Commit**

```
docs: update CLAUDE.md and golden tests for scaffold rename
```

---

### Task 13: Update omni's own aicontext category map

**Files:**
- Modify: `internal/cli/aicontext/aicontext.go` — change `"generate": "codegen"` to `"scaffold": "codegen"` in categoryMap

**Step 1: Update the category map**

Line 82: change `"generate": "codegen"` to `"scaffold": "codegen"`

**Step 2: Run tests**

Run: `go test ./internal/cli/aicontext/ -v`
Expected: PASS

**Step 3: Commit**

```
fix: update aicontext category map for scaffold rename
```

---

### Task 14: Final verification

**Step 1: Full test suite**

Run: `go test -race ./... 2>&1 | tail -20`
Expected: All PASS

**Step 2: Build**

Run: `go build -o /dev/null .`
Expected: Success

**Step 3: Smoke test CLI**

Run: `go run . scaffold cobra init /tmp/test-scaffold --module github.com/test/app`
Verify: Creates project with `cmd/cmdtree.go`, without `cmd/aicontext.go`

Run: `go run . scaffold cobra init /tmp/test-scaffold2 --module github.com/test/app2 --aicontext`
Verify: Creates project with both `cmd/cmdtree.go` and `cmd/aicontext.go`

Run: `go run . scaffold cobra add-tools --dir /tmp/test-scaffold`
Expected: Error — cmdtree.go already exists

**Step 4: Lint**

Run: `golangci-lint run --fix ./... --timeout=5m`
Expected: Clean

**Step 5: Commit (if lint fixes anything)**

```
chore: lint fixes for scaffold refactor
```
