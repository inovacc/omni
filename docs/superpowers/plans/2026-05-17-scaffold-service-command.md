# omni scaffolder `{app} service` command — Implementation Plan

**Status: ✅ COMPLETE** — executed and e2e-verified on `main`. Task→commit map:
T2 `c58fbf68` (failing test) · T3 `db673222` (ServiceCmdTemplate) · T4 `2a4918df`
(wire into `--service` mode) · T5 `1cbe1e1c` (e2e scaffold+build+smoke verified).
Follow-on `c568db8e` added `--platform-split`/`--daemon`. Reconciled via `/steps:next` 2026-06-03.

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a cross-platform `{app} service` (install/uninstall/start/stop/restart/status + run) command group to omni's `--service` scaffold output, using pure-Go `kardianos/service`.

**Architecture:** New `ServiceCmdTemplate` text/template constant rendered into `cmd/{AppName}/cmd_service.go`; wired into `RunCobraInit` behind the existing `opts.UseService`; `GoModTemplate` + MainTemplate adjusted. omni itself gains no deps and no exec.

**Tech Stack:** Go, Cobra, `text/template`, `afero` (scaffold I/O), `github.com/kardianos/service` (in *generated* project only).

**Working dir:** `C:\Users\dyamm\My Drive\acer\projects\omni`

**Repo hygiene:** omni may have a dirty/staged `.planning` tree. Use explicit Go-file pathspecs in every commit (`git commit -- <files>`), never `git add -A`. No AI attribution. Conventional commits. Branch: `feat/scaffold-service-command`.

---

### Task 1: Branch + locate exact insertion points

**Files:** none (investigation + branch)

- [x] **Step 1: Branch**

```bash
cd "C:/Users/dyamm/My Drive/acer/projects/omni"
git status   # note pre-existing .planning dirt, do not touch
git checkout -b feat/scaffold-service-command
```

- [x] **Step 2: Record exact anchors** (read, do not edit)

Read and note line numbers in:
- `internal/cli/scaffolding/cobra/templates/templates.go`: `TemplateData` struct, `MainTemplate` (the `{{if .UseService}}` block that sets `rootCmd.RunE = service.Handler`), `GoModTemplate` require block, end of file (where to append `ServiceCmdTemplate`).
- `internal/cli/scaffolding/cobra/cobra.go`: `RunCobraInit`, specifically the `WriteTemplate(... "internal/service/service.go", templates.ServiceTemplate ...)` call (the new block goes immediately after it, same `if opts.UseService` guard).
- `internal/cli/scaffolding/cobra/cobra_test.go`: `TestServiceMode` (for the regression assertion style + how `UseService` opts are constructed).

Record these as a short note in the task (file:line) for use in later tasks. No commit.

---

### Task 2: Failing test for the new service command output

**Files:**
- Test: `internal/cli/scaffolding/cobra/cobra_test.go`

- [x] **Step 1: Add failing test**

Add this test (adapt the options-construction to match how `TestServiceMode` builds opts — use the SAME opts struct/fields, only ensuring `UseService` is true and module/appname set):

```go
func TestServiceCommandMode(t *testing.T) {
	fs := afero.NewMemMapFs()
	opts := newTestInitOpts("svcapp", "example.com/svcapp") // mirror TestServiceMode's setup
	opts.UseService = true

	if err := RunCobraInit(fs, opts); err != nil {
		t.Fatalf("RunCobraInit: %v", err)
	}

	svcFile := "cmd/svcapp/cmd_service.go"
	b, err := afero.ReadFile(fs, svcFile)
	if err != nil {
		t.Fatalf("expected %s: %v", svcFile, err)
	}
	src := string(b)
	for _, want := range []string{
		"serviceCmd", "serviceProgram",
		"func (p *serviceProgram) Start", "func (p *serviceProgram) Stop",
		"kardianos/service",
		`Use:   "install"`, `Use:   "uninstall"`, `Use:   "start"`,
		`Use:   "stop"`, `Use:   "restart"`, `Use:   "status"`, `Use:   "run"`,
		"rootCmd.AddCommand(serviceCmd)",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("cmd_service.go missing %q", want)
		}
	}

	gomod, _ := afero.ReadFile(fs, "go.mod")
	if !strings.Contains(string(gomod), "github.com/kardianos/service") {
		t.Errorf("go.mod missing kardianos/service require")
	}

	mainB, _ := afero.ReadFile(fs, "cmd/svcapp/svcapp.go")
	if strings.Contains(string(mainB), "rootCmd.RunE = service.Handler") {
		t.Errorf("service mode must not hard-bind rootCmd.RunE; run owns it")
	}
}
```

If `newTestInitOpts` does not exist, replace with the literal opts construction copied from `TestServiceMode` (read it in Task 1) — do not invent a helper.

- [x] **Step 2: Run, verify it fails**

Run: `go test ./internal/cli/scaffolding/cobra/ -run TestServiceCommandMode -v`
Expected: FAIL (file `cmd/svcapp/cmd_service.go` not found; go.mod lacks kardianos).

- [x] **Step 3: Commit the failing test**

```bash
git commit -- internal/cli/scaffolding/cobra/cobra_test.go -m "test(scaffold): failing test for {app} service command"
```

---

### Task 3: Add `ServiceCmdTemplate` constant

**Files:**
- Modify: `internal/cli/scaffolding/cobra/templates/templates.go` (append new const)

- [x] **Step 1: Append the template constant**

Append at end of `templates.go` (Go raw-string; note `{{`/`}}` are template
actions, literal Go braces are written as-is):

```go
// ServiceCmdTemplate renders cmd/{AppName}/cmd_service.go — a cross-platform
// service lifecycle command group backed by pure-Go kardianos/service.
const ServiceCmdTemplate = `package main

import (
	"context"
	"fmt"

	"{{.Module}}/internal/service"

	"github.com/kardianos/service"
	"github.com/spf13/cobra"
)

type serviceProgram struct {
	cancel context.CancelFunc
}

func (p *serviceProgram) Start(s service.Service) error {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	go func() { _ = svcrun.Handler(ctx) }()
	return nil
}

func (p *serviceProgram) Stop(s service.Service) error {
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

func newService() (service.Service, error) {
	cfg := &service.Config{
		Name:        "{{.AppName}}",
		DisplayName: "{{.AppName}}",
		Description: "{{.AppName}} service",
		Arguments:   []string{"service", "run"},
	}
	return service.New(&serviceProgram{}, cfg)
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage the {{.AppName}} OS service",
}

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install {{.AppName}} as an OS service",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := newService()
		if err != nil {
			return fmt.Errorf("service install: %w", err)
		}
		if err := s.Install(); err != nil {
			return fmt.Errorf("service install (needs admin/root?): %w", err)
		}
		fmt.Println("installed")
		return nil
	},
}

var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove the {{.AppName}} OS service",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := newService()
		if err != nil {
			return fmt.Errorf("service uninstall: %w", err)
		}
		if err := s.Uninstall(); err != nil {
			return fmt.Errorf("service uninstall: %w", err)
		}
		fmt.Println("uninstalled")
		return nil
	},
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the {{.AppName}} service",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := newService()
		if err != nil {
			return fmt.Errorf("service start: %w", err)
		}
		if err := s.Start(); err != nil {
			return fmt.Errorf("service start: %w", err)
		}
		fmt.Println("started")
		return nil
	},
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the {{.AppName}} service",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := newService()
		if err != nil {
			return fmt.Errorf("service stop: %w", err)
		}
		if err := s.Stop(); err != nil {
			return fmt.Errorf("service stop: %w", err)
		}
		fmt.Println("stopped")
		return nil
	},
}

var serviceRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the {{.AppName}} service",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := newService()
		if err != nil {
			return fmt.Errorf("service restart: %w", err)
		}
		if err := s.Restart(); err != nil {
			return fmt.Errorf("service restart: %w", err)
		}
		fmt.Println("restarted")
		return nil
	},
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show {{.AppName}} service status",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := newService()
		if err != nil {
			return fmt.Errorf("service status: %w", err)
		}
		st, err := s.Status()
		if err != nil {
			fmt.Println("not installed")
			return nil
		}
		switch st {
		case service.StatusRunning:
			fmt.Println("running")
		case service.StatusStopped:
			fmt.Println("stopped")
		default:
			fmt.Println("not installed")
		}
		return nil
	},
}

var serviceRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run {{.AppName}} as a service (invoked by the OS)",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := newService()
		if err != nil {
			return fmt.Errorf("service run: %w", err)
		}
		return s.Run()
	},
}

func init() {
	serviceCmd.AddCommand(
		serviceInstallCmd,
		serviceUninstallCmd,
		serviceStartCmd,
		serviceStopCmd,
		serviceRestartCmd,
		serviceStatusCmd,
		serviceRunCmd,
	)
	rootCmd.AddCommand(serviceCmd)
}
`
```

NOTE on the `svcrun.Handler` reference: the existing `internal/service`
package's exported entrypoint may not be named `Handler` or may not accept a
`context.Context`. In Task 1 you read `ServiceTemplate`/`TestServiceMode`;
use the ACTUAL exported function name and signature from
`internal/service/service.go` as generated. If it is `service.Handler()` with
no ctx, adapt the `Start` goroutine to `go func() { _ = service.Handler() }()`
and drop the ctx wiring for the call while still using ctx for `Stop`
signaling via a package the handler observes — but DO NOT invent an API.
If the handler cannot be cleanly cancelled, implement `Stop` to call
`s.Stop()`'s default (return nil) and add a `// TODO` comment; report this as
DONE_WITH_CONCERNS. Prefer the simplest correct wiring that compiles in the
generated project.

- [x] **Step 2: Build omni (template is just a string — must still compile)**

Run: `go build ./...`
Expected: success.

- [x] **Step 3: Commit**

```bash
git commit -- internal/cli/scaffolding/cobra/templates/templates.go -m "feat(scaffold): add ServiceCmdTemplate constant"
```

---

### Task 4: Wire template into RunCobraInit + go.mod + MainTemplate

**Files:**
- Modify: `internal/cli/scaffolding/cobra/cobra.go`
- Modify: `internal/cli/scaffolding/cobra/templates/templates.go` (GoModTemplate, MainTemplate)

- [x] **Step 1: Add the write block in `RunCobraInit`**

Immediately after the existing `internal/service/service.go` WriteTemplate
call (located in Task 1), inside the same `if opts.UseService { ... }` guard,
add (match the surrounding code's variable names for `fs`, data struct, and
the AppName path segment exactly as the service.go line uses them):

```go
if err := scaffolding.WriteTemplate(fs,
	filepath.Join("cmd", data.AppName, "cmd_service.go"),
	templates.ServiceCmdTemplate, data); err != nil {
	return err
}
```

(Use the SAME path-join idiom and `data` value the adjacent service.go write
uses — if it uses `fmt.Sprintf("cmd/%s/...", ...)` copy that form instead.)

- [x] **Step 2: Add kardianos require to GoModTemplate**

In `GoModTemplate`, inside the require section, add (matching existing
require-line indentation/syntax in that template):

```
{{if .UseService}}	github.com/kardianos/service v1.2.2
{{end}}
```

Place it so non-service projects render an unchanged go.mod (the `{{if}}`
must not leave a blank require line when false — mirror how other optional
requires in this template are gated; if none exist, ensure the `{{if}}`/`{{end}}`
sit on their own lines so output stays valid).

- [x] **Step 3: Adjust MainTemplate service branch**

In `MainTemplate`, find the `{{if .UseService}}` block that sets
`rootCmd.RunE = service.Handler` (Task 1 anchor). Remove the `RunE` binding
so root stays a command group (the `service run` subcommand now owns the
long-running path). Keep any service-related import only if still used; if
removing the binding makes the `service` import unused in the main file,
remove that import line too (the generated file must compile). Leave all
non-service template content untouched.

- [x] **Step 4: Run the Task 2 test — must pass**

Run: `go test ./internal/cli/scaffolding/cobra/ -run TestServiceCommandMode -v`
Expected: PASS

- [x] **Step 5: Full scaffolder suite (regression)**

Run: `go test ./internal/cli/scaffolding/...`
Expected: PASS (including existing `TestServiceMode`, `TestRunCobraInit`,
`TestFullMode`, `TestTaskfileContent`). If `TestServiceMode` asserted the old
`rootCmd.RunE = service.Handler` string, update THAT assertion to reflect the
new behavior (root is a group; `service run` owns the handler) — this is an
intended behavior change, not a test break to paper over.

- [x] **Step 6: Commit**

```bash
git commit -- internal/cli/scaffolding/cobra/cobra.go internal/cli/scaffolding/cobra/templates/templates.go -m "feat(scaffold): wire {app} service command into --service mode"
```

---

### Task 5: End-to-end — scaffold a project and prove it builds

**Files:** none (verification; uses a temp dir outside the repo)

- [x] **Step 1: Scaffold a real project with --service**

```bash
cd "C:/Users/dyamm/My Drive/acer/projects/omni"
mkdir -p "$TEMP/svcdemo" && cd "$TEMP/svcdemo"
go run "C:/Users/dyamm/My Drive/acer/projects/omni" scaffold cobra init svcdemo --module example.com/svcdemo --service
```

(Use the actual omni invocation form discovered in Task 1 — adjust the
command path / subcommand to match `cmd/scaffold.go`.)

- [x] **Step 2: Build the generated project**

```bash
cd "$TEMP/svcdemo"
go mod tidy
go build ./...
```

Expected: success — `cmd/svcdemo/cmd_service.go` compiles, kardianos resolved.

- [x] **Step 3: Smoke the command surface**

```bash
go run ./cmd/svcdemo service --help
go run ./cmd/svcdemo service status
```

Expected: help lists install/uninstall/start/stop/restart/status/run;
`status` prints `not installed` (exit 0) since nothing is registered.

- [x] **Step 4: Record result; final commit**

If anything failed, fix in the template/wiring (re-commit scoped), re-run.
When green:

```bash
cd "C:/Users/dyamm/My Drive/acer/projects/omni"
git commit --allow-empty -m "test(scaffold): verified {app} service e2e (scaffold+build+smoke)"
```

---

## Self-Review

**Spec coverage:** ServiceCmdTemplate (T3) ↔ spec "Generated implementation";
RunCobraInit wiring + GoModTemplate + MainTemplate (T4) ↔ spec "Scaffolder
wiring" + "Root command change"; tests T2/T4-step5 ↔ spec "Testing"; e2e T5
↔ spec intent (pure-Go, builds on the dev box). kardianos-only, no omni dep,
no exec — honored (omni only emits a template string + a generated-project
require). All spec sections mapped.

**Placeholder scan:** No TBD/TODO except the explicit, bounded contingency in
T3 about the real `internal/service` entrypoint name — this is a
read-the-actual-code instruction with a defined fallback, not a vague
placeholder.

**Type/name consistency:** `serviceCmd`, `serviceProgram`,
`serviceInstallCmd…serviceRunCmd`, `newService()` consistent between T3
template and T2 test assertions. Test checks `rootCmd.AddCommand(serviceCmd)`
which T3's `init()` emits. go.mod assertion (`github.com/kardianos/service`)
matches T4 GoModTemplate addition. MainTemplate negative assertion (no
`rootCmd.RunE = service.Handler`) matches T4 Step 3.

**Known risk:** the generated `internal/service` handler API name/signature is
not knowable from the spec alone; T1 mandates reading it and T3 gives a
bounded adaptation rule. Flag as DONE_WITH_CONCERNS if the handler is
un-cancellable.
