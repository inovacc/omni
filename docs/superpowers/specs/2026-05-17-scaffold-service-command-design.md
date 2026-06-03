# omni scaffolder: `{app} service` command — Design Spec

Date: 2026-05-17
Status: Approved

## Purpose

Extend omni's Cobra scaffolder so a project generated with `--service` also
gets a cross-platform `{app} service` command group with
`install/uninstall/start/stop/restart/status` plus a `run` entrypoint, using
the pure-Go `github.com/kardianos/service` library (honors omni's no-exec
foundational rule; covers Windows SCM / systemd / launchd, including the macOS
case weaver lacked).

## Scope

- Modify: omni's scaffolder only (`internal/cli/scaffolding/cobra/`).
- The generated project gains one new file `cmd/{AppName}/cmd_service.go` and a
  `kardianos/service` require in its `go.mod` when `--service` is used.
- omni itself adds NO new runtime dependency and NO process exec.
- Out of scope: configurable service user/working-dir/log paths (kardianos
  defaults), per-OS template branching (kardianos abstracts OS), changes to
  non-service scaffold modes.

## Design

### Generated command surface (in the scaffolded project)

```
{app} service install     register OS service, auto-start at boot
{app} service uninstall    deregister
{app} service start        start service
{app} service stop         stop service
{app} service restart      stop then start
{app} service status       running | stopped | not installed
{app} service run          long-running entrypoint the OS invokes
```

### Generated implementation (`cmd/{AppName}/cmd_service.go`, package main)

- A `serviceProgram` type implementing kardianos `service.Interface`:
  - `Start(s service.Service) error` — launches existing `service.Handler`
    logic in a goroutine with a cancellable `context.Context`; returns
    immediately (kardianos contract).
  - `Stop(s service.Service) error` — cancels the context for graceful
    shutdown (mirrors weaver's SIGTERM intent via the library).
- `service.Config{Name: "{{.AppName}}", DisplayName: "{{.AppName}}",
  Description: "{{.AppName}} service", Arguments: []string{"service","run"}}`
  so the OS launches `{app} service run`.
- Subcommand RunE bodies call `svc.Install()/Uninstall()/Start()/Stop()/
  Restart()` and, for `status`, `svc.Status()` mapping
  `service.StatusRunning|StatusStopped|StatusUnknown` →
  `running|stopped|not installed`.
- `run` builds the kardianos service and calls `svc.Run()` (blocks; kardianos
  invokes `Start`/`Stop`). Reuses the existing `internal/service` Handler so
  service business logic is not duplicated.
- All subcommands registered under one `serviceCmd` parent; `serviceCmd`
  registered to `rootCmd` via this file's `init()` (matches omni's existing
  per-file `init()` AddCommand pattern in CommandTemplate).

### Root command change (MainTemplate, `--service` path)

Currently `--service` sets `rootCmd.RunE = service.Handler`. Change: under
`{{if .UseService}}` the root no longer hard-binds `RunE`; the long-running
behavior moves to `service run`. `{app}` with no args prints help (Cobra
default). This keeps root a pure command group and is backward-safe (no panic,
no behavior regression for `version`/`cmdtree`/etc.).

### Scaffolder wiring

1. `templates/templates.go`: add `ServiceCmdTemplate` string const
   (text/template, uses `{{.AppName}}` and `{{.Module}}` for the
   `internal/service` import path).
2. `cobra.go` `RunCobraInit`: add a write block, gated by the existing
   `opts.UseService`, calling `scaffolding.WriteTemplate(fs,
   "cmd/<AppName>/cmd_service.go", templates.ServiceCmdTemplate, data)`,
   placed immediately after the existing `internal/service/service.go` write.
3. `templates/templates.go` `GoModTemplate`: add
   `{{if .UseService}}\n\trequire github.com/kardianos/service v1.2.2{{end}}`
   to the generated project's require block (generated go.mod only).
4. MainTemplate: adjust the `{{if .UseService}}` branch per "Root command
   change" above.

## Template Variables

Uses omni's existing `TemplateData` (templates.go): `{{.AppName}}`,
`{{.Module}}`, `{{.Description}}`, `{{.UseService}}`. No new fields required.

## Error Handling

Generated subcommands return wrapped errors (`fmt.Errorf("service install:
%w", err)`) to stderr via Cobra; non-zero exit on failure. `status` is
non-error for stopped/not-installed (prints state, exits 0). Privilege errors
from kardianos (needs admin/root) surface verbatim with a hint line.

## Testing

`internal/cli/scaffolding/cobra/cobra_test.go`, afero in-memory FS, omni's
existing `strings.Contains` assertion style:

- `TestServiceCommandMode`: scaffold with `UseService=true`; assert
  `cmd/<app>/cmd_service.go` exists and contains: `serviceCmd`, all six verb
  subcommands (`install uninstall start stop restart status`), `run`,
  `kardianos/service`, `serviceProgram`, `func (p *serviceProgram) Start`,
  `Stop`. Assert generated `go.mod` contains
  `github.com/kardianos/service`. Assert MainTemplate output for service mode
  does NOT hard-bind `rootCmd.RunE = service.Handler`.
- Regression: existing `TestServiceMode` (internal/service Handler) still
  passes; `TestRunCobraInit`/`TestFullMode` unaffected (non-service path
  unchanged).

## Non-Goals / YAGNI

No service config flags, no log-path options, no Taskfile service targets, no
omni-side dependency, no weaver code ported.
