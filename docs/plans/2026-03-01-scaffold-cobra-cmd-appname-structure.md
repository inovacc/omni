# Scaffold Cobra: cmd/{appName} Structure

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Change scaffold cobra to generate `cmd/{appName}/` structure instead of `main.go` + `cmd/root.go`, with `package main` entry point at `cmd/{appName}/{appName}.go` and commands as `cmd/{appName}/cmd_{name}.go`.

**Architecture:** Merge the current `main.go` and `cmd/root.go` into a single `cmd/{appName}/{appName}.go` file that contains both `func main()` and the root command. All generated commands (version, cmdtree, aicontext, user-added) use `cmd_` prefix. The `scaffold cobra add` command places new files at `cmd/{appName}/cmd_{name}.go`.

**Tech Stack:** Go, Cobra, afero (filesystem abstraction), text/template

---

### Task 1: Update Templates — MainTemplate + RootTemplate merge

**Files:**
- Modify: `internal/cli/scaffolding/cobra/templates/templates.go:33-110`

**Step 1: Write the failing test**

Add to `internal/cli/scaffolding/cobra/cobra_test.go`:

```go
func TestRunCobraInitNewStructure(t *testing.T) {
	t.Run("creates cmd/appName/appName.go instead of main.go and cmd/root.go", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer

		err := RunCobraInit(&buf, fs, "/myapp", CobraInitOptions{
			Module:      "github.com/test/myapp",
			AppName:     "myapp",
			Description: "Test application",
		}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunCobraInit() error = %v", err)
		}

		// New structure: cmd/myapp/myapp.go should exist
		if _, err := fs.Stat("/myapp/cmd/myapp/myapp.go"); err != nil {
			t.Error("Expected cmd/myapp/myapp.go to be created")
		}

		// Old structure: main.go and cmd/root.go should NOT exist
		if _, err := fs.Stat("/myapp/main.go"); err == nil {
			t.Error("main.go should NOT be created in new structure")
		}
		if _, err := fs.Stat("/myapp/cmd/root.go"); err == nil {
			t.Error("cmd/root.go should NOT be created in new structure")
		}

		// Verify content
		content, _ := afero.ReadFile(fs, "/myapp/cmd/myapp/myapp.go")
		mainStr := string(content)

		if !strings.Contains(mainStr, "package main") {
			t.Error("cmd/myapp/myapp.go should have package main")
		}
		if !strings.Contains(mainStr, "func main()") {
			t.Error("cmd/myapp/myapp.go should have func main()")
		}
		if !strings.Contains(mainStr, `Use:   "myapp"`) {
			t.Error("cmd/myapp/myapp.go should define root command with app name")
		}
	})
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestRunCobraInitNewStructure ./internal/cli/scaffolding/cobra/ -v`
Expected: FAIL — still creates main.go + cmd/root.go

**Step 3: Update MainTemplate to be the merged entry point**

In `internal/cli/scaffolding/cobra/templates/templates.go`, replace `MainTemplate` with:

```go
// MainTemplate generates cmd/{appName}/{appName}.go — the entry point with root command
const MainTemplate = `package main

import (
{{if .UseService}}
	"fmt"
	"os"

	"{{.Module}}/internal/parameters"
	"{{.Module}}/internal/service"

	"github.com/inovacc/config"
{{else if .UseViper}}
	"{{.Module}}/internal/config"
{{end}}
	"github.com/spf13/cobra"
)
{{if or .UseViper .UseService}}
var cfgFile string
{{end}}
var rootCmd = &cobra.Command{
	Use:   "{{.AppName}}",
	Short: "{{.Description}}",
	Long: ` + "`" + `{{.Description}}

This is a CLI application built with Cobra.` + "`" + `,
{{if .UseService}}
	RunE: service.Handler,
{{end}}
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func main() {
	Execute()
}

func init() {
{{if or .UseViper .UseService}}
	cobra.OnInitialize(initConfig)
{{end}}
{{if .Full}}
	rootCmd.Version = GetVersionJSON()
	rootCmd.CompletionOptions.DisableDefaultCmd = true
{{end}}
{{if or .UseViper .UseService}}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "config file (default is config.yaml)")
{{else}}
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
{{end}}
}
{{if or .UseViper .UseService}}
// initConfig reads in config file and ENV variables if set.
func initConfig() {
{{if .UseService}}
	if cfgFile == "" {
		_, _ = fmt.Fprint(os.Stdout, "Using default config file: config.yaml")
	}

	// Load configuration from a file, applying defaults if needed
	if err := config.InitServiceConfig(&parameters.Service{}, cfgFile); err != nil {
		_, _ = fmt.Fprint(os.Stdout, "failed to load config: %w", err)
	}
{{else if .UseViper}}
	config.InitConfig("{{.AppName}}")
{{end}}
}
{{end}}
`
```

Delete `RootTemplate` entirely (it's now merged into `MainTemplate`).

**Step 4: Update RunCobraInit to use new paths**

In `internal/cli/scaffolding/cobra/cobra.go`, update the directory creation and file generation:

- Change `filepath.Join(dir, "cmd")` → `filepath.Join(dir, "cmd", opts.AppName)`
- Change `mainPath := filepath.Join(dir, "main.go")` → `mainPath := filepath.Join(dir, "cmd", opts.AppName, opts.AppName+".go")`
- Remove the `cmd/root.go` generation block entirely
- Update all `cmd/*.go` paths to `cmd/{appName}/cmd_*.go`

**Step 5: Run test to verify it passes**

Run: `go test -run TestRunCobraInitNewStructure ./internal/cli/scaffolding/cobra/ -v`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/cli/scaffolding/cobra/templates/templates.go internal/cli/scaffolding/cobra/cobra.go internal/cli/scaffolding/cobra/cobra_test.go
git commit -m "feat(scaffold): merge main.go+root.go into cmd/{appName}/{appName}.go"
```

---

### Task 2: Update VersionTemplate, CmdtreeTemplate, AIContextTemplate, CommandTemplate packages

**Files:**
- Modify: `internal/cli/scaffolding/cobra/templates/templates.go`
- Modify: `internal/cli/scaffolding/cobra/templates/cmdtree.go`
- Modify: `internal/cli/scaffolding/cobra/templates/aicontext.go`

**Step 1: Write the failing test**

Add to `cobra_test.go`:

```go
func TestNewStructureCommandFiles(t *testing.T) {
	fs := afero.NewMemMapFs()
	var buf bytes.Buffer

	err := RunCobraInit(&buf, fs, "/myapp", CobraInitOptions{
		Module:    "github.com/test/myapp",
		AppName:   "myapp",
		AIContext: true,
	}, scaffolding.Options{})
	if err != nil {
		t.Fatalf("RunCobraInit() error = %v", err)
	}

	// All command files should be cmd_{name}.go inside cmd/myapp/
	files := map[string]string{
		"/myapp/cmd/myapp/cmd_version.go":   "package main",
		"/myapp/cmd/myapp/cmd_cmdtree.go":   "package main",
		"/myapp/cmd/myapp/cmd_aicontext.go": "package main",
	}

	for path, expectedPkg := range files {
		content, err := afero.ReadFile(fs, path)
		if err != nil {
			t.Errorf("Expected %s to exist", path)
			continue
		}
		if !strings.Contains(string(content), expectedPkg) {
			t.Errorf("%s should contain %q", path, expectedPkg)
		}
	}

	// Old paths should NOT exist
	oldPaths := []string{
		"/myapp/cmd/version.go",
		"/myapp/cmd/cmdtree.go",
		"/myapp/cmd/aicontext.go",
	}
	for _, p := range oldPaths {
		if _, err := fs.Stat(p); err == nil {
			t.Errorf("%s should NOT exist in new structure", p)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestNewStructureCommandFiles ./internal/cli/scaffolding/cobra/ -v`
Expected: FAIL

**Step 3: Change all template `package cmd` to `package main`**

In `templates.go`:
- `VersionTemplate`: change `package cmd` → `package main`
- `CommandTemplate`: change `package cmd` → `package main`

In `cmdtree.go`:
- `CmdtreeTemplate`: change `package cmd` → `package main`

In `aicontext.go`:
- `AIContextTemplate`: change `package cmd` → `package main`

**Step 4: Update file paths in RunCobraInit**

In `cobra.go`, update all generated file paths:
- `cmd/version.go` → `cmd/{appName}/cmd_version.go`
- `cmd/cmdtree.go` → `cmd/{appName}/cmd_cmdtree.go`
- `cmd/aicontext.go` → `cmd/{appName}/cmd_aicontext.go`

**Step 5: Run test to verify it passes**

Run: `go test -run TestNewStructureCommandFiles ./internal/cli/scaffolding/cobra/ -v`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/cli/scaffolding/cobra/
git commit -m "feat(scaffold): move all command templates to cmd/{appName}/cmd_{name}.go"
```

---

### Task 3: Update RunCobraAdd to use new structure

**Files:**
- Modify: `internal/cli/scaffolding/cobra/cobra.go:424-482`

**Step 1: Write the failing test**

Add to `cobra_test.go`:

```go
func TestRunCobraAddNewStructure(t *testing.T) {
	setupProject := func(t *testing.T) (afero.Fs, string) {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer
		err := RunCobraInit(&buf, fs, "/addtest", CobraInitOptions{
			Module: "github.com/test/addtest",
		}, scaffolding.Options{})
		if err != nil {
			t.Fatal(err)
		}
		return fs, "/addtest"
	}

	t.Run("add command creates cmd/{appName}/cmd_{name}.go", func(t *testing.T) {
		fs, appDir := setupProject(t)
		var buf bytes.Buffer

		err := RunCobraAdd(&buf, fs, appDir, CobraAddOptions{
			Name:        "serve",
			Parent:      "root",
			Description: "Start the server",
		}, scaffolding.Options{})
		if err != nil {
			t.Fatalf("RunCobraAdd() error = %v", err)
		}

		// New path
		content, err := afero.ReadFile(fs, "/addtest/cmd/addtest/cmd_serve.go")
		if err != nil {
			t.Fatal("cmd/addtest/cmd_serve.go should be created")
		}

		serveStr := string(content)
		if !strings.Contains(serveStr, "package main") {
			t.Error("cmd_serve.go should have package main")
		}
		if !strings.Contains(serveStr, "serveCmd") {
			t.Error("cmd_serve.go should define serveCmd")
		}

		// Old path should NOT exist
		if _, err := fs.Stat("/addtest/cmd/serve.go"); err == nil {
			t.Error("cmd/serve.go should NOT exist")
		}
	})
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestRunCobraAddNewStructure ./internal/cli/scaffolding/cobra/ -v`
Expected: FAIL

**Step 3: Update RunCobraAdd**

In `cobra.go`, modify `RunCobraAdd`:
- Detect app name from go.mod (reuse `parseModuleName`)
- Change `cmdDir` from `filepath.Join(dir, "cmd")` to `filepath.Join(dir, "cmd", appName)`
- Change `cmdPath` from `filepath.Join(cmdDir, opts.Name+".go")` to `filepath.Join(cmdDir, "cmd_"+opts.Name+".go")`
- Update the result `File` field accordingly

```go
func RunCobraAdd(w io.Writer, fs afero.Fs, dir string, opts CobraAddOptions, genOpts scaffolding.Options) error {
	if opts.Name == "" {
		return fmt.Errorf("scaffold: command name is required")
	}

	if opts.Parent == "" {
		opts.Parent = "root"
	}

	if opts.Description == "" {
		opts.Description = fmt.Sprintf("%s command", opts.Name)
	}

	// Read go.mod to get app name
	goModData, err := afero.ReadFile(fs, filepath.Join(dir, "go.mod"))
	if err != nil {
		return fmt.Errorf("scaffold: failed to read go.mod: %w", err)
	}

	moduleName := parseModuleName(goModData)
	if moduleName == "" {
		return fmt.Errorf("scaffold: failed to parse module name from go.mod")
	}

	appName := moduleName
	if parts := strings.Split(moduleName, "/"); len(parts) > 0 {
		appName = parts[len(parts)-1]
	}

	// Check if cmd/{appName} directory exists
	cmdDir := filepath.Join(dir, "cmd", appName)
	if _, err := fs.Stat(cmdDir); err != nil {
		return fmt.Errorf("scaffold: cmd/%s directory not found, is this a Cobra project?", appName)
	}

	// Generate the command file with cmd_ prefix
	cmdPath := filepath.Join(cmdDir, "cmd_"+opts.Name+".go")
	if _, err := fs.Stat(cmdPath); err == nil {
		return fmt.Errorf("scaffold: command %s already exists", opts.Name)
	}

	data := struct {
		Name        string
		Parent      string
		Description string
		NameTitle   string
	}{
		Name:        opts.Name,
		Parent:      opts.Parent,
		Description: opts.Description,
		NameTitle:   strings.Title(opts.Name), //nolint:staticcheck
	}

	if err := scaffolding.WriteTemplate(fs, cmdPath, cobratpl.CommandTemplate, data); err != nil {
		return fmt.Errorf("scaffold: failed to create cmd_%s.go: %w", opts.Name, err)
	}

	relPath := filepath.Join("cmd", appName, "cmd_"+opts.Name+".go")

	if genOpts.JSON {
		result := AddResult{
			Status:  "created",
			Command: opts.Name,
			Parent:  opts.Parent,
			File:    relPath,
		}
		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintf(w, "Created command: %s\n", opts.Name)
	_, _ = fmt.Fprintf(w, "Parent: %s\n", opts.Parent)
	_, _ = fmt.Fprintf(w, "File: %s\n", relPath)

	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test -run TestRunCobraAddNewStructure ./internal/cli/scaffolding/cobra/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/cli/scaffolding/cobra/cobra.go internal/cli/scaffolding/cobra/cobra_test.go
git commit -m "feat(scaffold): update cobra add to use cmd/{appName}/cmd_{name}.go"
```

---

### Task 4: Update RunCobraAddTools for new structure

**Files:**
- Modify: `internal/cli/scaffolding/cobra/cobra.go:337-410`

**Step 1: Write the failing test**

Add to `cobra_test.go`:

```go
func TestRunCobraAddToolsNewStructure(t *testing.T) {
	fs := afero.NewMemMapFs()
	appDir := "/toolsapp"

	// Create minimal project with new structure
	if err := fs.MkdirAll(appDir+"/cmd/toolsapp", 0755); err != nil {
		t.Fatal(err)
	}
	goMod := []byte("module github.com/test/toolsapp\n\ngo 1.21\n")
	if err := afero.WriteFile(fs, appDir+"/go.mod", goMod, 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err := RunCobraAddTools(&buf, fs, appDir, AddToolsOptions{AIContext: true}, scaffolding.Options{})
	if err != nil {
		t.Fatalf("RunCobraAddTools() error = %v", err)
	}

	if _, err := fs.Stat(appDir + "/cmd/toolsapp/cmd_cmdtree.go"); err != nil {
		t.Error("cmd/toolsapp/cmd_cmdtree.go should be created")
	}
	if _, err := fs.Stat(appDir + "/cmd/toolsapp/cmd_aicontext.go"); err != nil {
		t.Error("cmd/toolsapp/cmd_aicontext.go should be created")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestRunCobraAddToolsNewStructure ./internal/cli/scaffolding/cobra/ -v`
Expected: FAIL

**Step 3: Update RunCobraAddTools**

Change `cmdDir` to `filepath.Join(dir, "cmd", appName)` and file paths to use `cmd_` prefix:
- `cmd/cmdtree.go` → `cmd/{appName}/cmd_cmdtree.go`
- `cmd/aicontext.go` → `cmd/{appName}/cmd_aicontext.go`

**Step 4: Run test to verify it passes**

Run: `go test -run TestRunCobraAddToolsNewStructure ./internal/cli/scaffolding/cobra/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/cli/scaffolding/cobra/cobra.go internal/cli/scaffolding/cobra/cobra_test.go
git commit -m "feat(scaffold): update add-tools to use cmd/{appName}/cmd_{name}.go"
```

---

### Task 5: Update Taskfile, README, goreleaser templates for new paths

**Files:**
- Modify: `internal/cli/scaffolding/cobra/templates/templates.go`

**Step 1: Write the failing test**

```go
func TestTaskfileNewStructure(t *testing.T) {
	fs := afero.NewMemMapFs()
	var buf bytes.Buffer

	_ = RunCobraInit(&buf, fs, "/tasktest", CobraInitOptions{
		Module:  "github.com/test/tasktest",
		AppName: "tasktest",
	}, scaffolding.Options{})

	content, _ := afero.ReadFile(fs, "/tasktest/Taskfile.yml")
	taskStr := string(content)

	if !strings.Contains(taskStr, "./cmd/tasktest") {
		t.Error("Taskfile should reference ./cmd/tasktest as MAIN_PACKAGE")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestTaskfileNewStructure ./internal/cli/scaffolding/cobra/ -v`
Expected: FAIL

**Step 3: Update templates**

In `TaskfileTemplate`:
- Change `MAIN_PACKAGE: .` → `MAIN_PACKAGE: ./cmd/{{.AppName}}`

In `ReadmeTemplate`:
- Update project structure section to show `cmd/{appName}/` layout

In `GoreleaserTemplate`:
- Add `main: ./cmd/{{.AppName}}` under builds

In build ldflags:
- Change `-X {{.Module}}/cmd.Version` → `-X main.Version` (since it's now package main)

In `TaskfileTemplate` install task:
- Same ldflags change

The `aicontext.go` template's project structure section also needs updating:
```go
structure := []string{
    "cmd/{appName}/     # CLI entry point and commands",
    "internal/          # Private application code",
    "pkg/               # Public reusable libraries",
}
```

**Step 4: Run test to verify it passes**

Run: `go test -run TestTaskfileNewStructure ./internal/cli/scaffolding/cobra/ -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/cli/scaffolding/cobra/templates/
git commit -m "feat(scaffold): update Taskfile/README/goreleaser for cmd/{appName} structure"
```

---

### Task 6: Fix all existing tests

**Files:**
- Modify: `internal/cli/scaffolding/cobra/cobra_test.go`

**Step 1: Update all existing test assertions**

Go through every test in `cobra_test.go` and update path expectations:
- `"/myapp/main.go"` → `"/myapp/cmd/myapp/myapp.go"`
- `"/myapp/cmd/root.go"` → removed (merged)
- `"/myapp/cmd/version.go"` → `"/myapp/cmd/myapp/cmd_version.go"`
- `"/myapp/cmd/cmdtree.go"` → `"/myapp/cmd/myapp/cmd_cmdtree.go"`
- `"/myapp/cmd/aicontext.go"` → `"/myapp/cmd/myapp/cmd_aicontext.go"`
- `"package cmd"` checks → `"package main"` checks
- `RunCobraAdd` tests: `"/addtest/cmd/serve.go"` → `"/addtest/cmd/addtest/cmd_serve.go"`
- `RunCobraAddTools` tests: update `cmd/` → `cmd/{appName}/cmd_` paths

**Step 2: Run full test suite**

Run: `go test ./internal/cli/scaffolding/cobra/ -v`
Expected: ALL PASS

**Step 3: Commit**

```bash
git add internal/cli/scaffolding/cobra/cobra_test.go
git commit -m "test(scaffold): update all tests for cmd/{appName} structure"
```

---

### Task 7: Update cmd/scaffold.go CLI help text

**Files:**
- Modify: `cmd/scaffold.go:65-109`

**Step 1: Update help text**

Change the `scaffoldCobraInitCmd` Long description:
- `main.go          Entry point` → remove
- `cmd/root.go      Root command` → `cmd/{name}/{name}.go  Entry point + root command`
- `cmd/version.go   Version command` → `cmd/{name}/cmd_version.go   Version command`
- `cmd/cmdtree.go   Command tree utility` → `cmd/{name}/cmd_cmdtree.go   Command tree utility`
- `cmd/aicontext.go AI context generator` → `cmd/{name}/cmd_aicontext.go AI context generator`

Update `scaffoldCobraAddCmd` Long description to mention `cmd_{name}.go` prefix.

Update `scaffoldCobraAddToolsCmd` Long description similarly.

**Step 2: Run omni build to verify**

Run: `go build ./...`
Expected: SUCCESS

**Step 3: Commit**

```bash
git add cmd/scaffold.go
git commit -m "docs(scaffold): update CLI help text for cmd/{appName} structure"
```

---

### Task 8: Update CLAUDE.md documentation

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Update the scaffold cobra section**

Update the "Command Implementation Pattern" to show the new structure. Update the `scaffoldCobraInitCmd` documentation to reflect new paths.

No test needed — documentation only.

**Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md for scaffold cobra cmd/{appName} structure"
```
