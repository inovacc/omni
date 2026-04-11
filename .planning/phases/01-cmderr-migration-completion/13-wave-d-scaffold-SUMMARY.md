---
phase: 01-cmderr-migration-completion
plan: 13
subsystem: scaffold, project, repo
tags: [cmderr, migration, wave-d]
dependency_graph:
  requires: [09, 10, 11]
  provides: [scaffold-cmderr, project-cmderr, repo-cmderr]
  affects: [cmd/scaffold.go, internal/cli/scaffolding, internal/cli/project, internal/cli/repo]
tech_stack:
  added: []
  patterns: [cmderr sentinel classification, Pattern 1/2 error wrapping]
key_files:
  created: []
  modified:
    - internal/cli/scaffolding/shared.go
    - internal/cli/scaffolding/cobra/cobra.go
    - internal/cli/scaffolding/handler/handler.go
    - internal/cli/scaffolding/repository/repository.go
    - internal/cli/scaffolding/testgen/testgen.go
    - internal/cli/scaffolding/mcp/mcp.go
    - internal/cli/project/project.go
    - internal/cli/repo/repo.go
    - internal/cli/repo/analyze.go
    - .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
decisions:
  - "Classified at sub-package Run* boundary (not cmd/scaffold.go) since sub-packages own the error paths"
  - "remote.go exec usage noted as pre-existing design-principle violation; deferred to deferred-items"
  - "resolvePath errors pass through classify-then-bubble pattern (already classified, no double-wrap)"
metrics:
  duration: ~25min
  completed: 2026-04-11
  tasks: 4
  files: 9
---

# Phase 1 Plan 13: Wave D scaffold, project, repo Summary

One-liner: cmderr sentinel classification for scaffold subcommands (cobra/handler/repository/testgen/mcp), project path resolution, and repo analyze path/clone/output errors.

## Tasks Completed

| # | Task | Status | Commit |
|---|------|--------|--------|
| 1 | Migrate scaffolding CLI entry (shared + sub-packages) | Done | ff607bfc |
| 2 | Migrate project + repo | Done | d82ba923 |
| 3 | Golden snapshots | DEFERRED per plan scope (Plan 16) | — |
| 4 | Log EXIT-CODE-CHANGES.md | Done | (metadata commit) |

## Commits

- `ff607bfc` — refactor(scaffold): adopt cmderr sentinels (6 files, 40 insertions)
- `d82ba923` — refactor(project,repo): adopt cmderr sentinels (3 files, 34 insertions)

## Error Classifications Applied

### scaffolding/shared.go
- `WriteTemplate`: template parse failure → `ErrInvalidInput`; file create/write failure → `ErrIO`
- `WriteLicense`: unknown license type → `ErrInvalidInput`

### scaffolding/cobra/cobra.go
- Missing `--module` flag → `ErrInvalidInput`
- `MkdirAll` failure → `ErrIO`
- `go.mod` not found → `ErrNotFound`
- Module name parse failure → `ErrInvalidInput`
- cmd dir not found → `ErrNotFound`
- File already exists (cmdtree, aicontext, command) → `ErrConflict`
- Empty command name → `ErrInvalidInput`

### scaffolding/handler/handler.go
- Empty name → `ErrInvalidInput`
- `MkdirAll` failure → `ErrIO`

### scaffolding/repository/repository.go
- Empty name → `ErrInvalidInput`
- `MkdirAll` failure → `ErrIO`

### scaffolding/testgen/testgen.go
- Empty source path → `ErrInvalidInput`
- Source file not found → `ErrNotFound`
- Go parse failure → `ErrInvalidInput`
- No exported functions → `ErrInvalidInput`

### scaffolding/mcp/mcp.go
- Empty name → `ErrInvalidInput`
- Invalid transport → `ErrInvalidInput`
- go.mod not found (module detection) → `ErrNotFound`
- `MkdirAll` failures → `ErrIO`

### project/project.go
- `resolvePath`: `os.ErrNotExist` → `ErrNotFound`; `os.ErrPermission` → `ErrPermission`; other stat → `ErrIO`; bad path → `ErrInvalidInput`

### repo/repo.go
- `resolvePath`: same classification pattern as project

### repo/analyze.go
- `cloneToTemp` failure → `ErrIO`
- `resolvePath` errors pass through (already classified)
- Output file `os.Create` failure → `ErrIO`

## Deviations from Plan

### Scope adjustment — Task 3 (golden snapshots) deferred
Per plan execution rules: "DEFER all golden snapshot work to Plan 16."
Task 3 was never in scope for this execution.

### Sub-package classification (not top-level dispatcher only)
The plan said "wrap at that file's exported Run* functions" referring to the top-level `scaffolding.go`. Since no top-level `scaffolding.go` with Run* functions exists (the dispatcher lives in sub-packages), classification was applied at the sub-package Run* level, which is the correct CLI boundary per Pattern 1/2.

### remote.go exec usage
`internal/cli/repo/remote.go` uses `os/exec` to run `gh` and `git clone`, violating the project's "No exec" design principle. This is pre-existing and out of scope for this plan. Logged to deferred-items.

## Known Stubs

None.

## Threat Flags

None — no new network endpoints, auth paths, or schema changes introduced.

## Self-Check: PASSED

- `ff607bfc` exists: confirmed via `git log`
- `d82ba923` exists: confirmed via `git log`
- All modified files contain `cmderr.` references
- `go build ./...` clean
- `go test -race ./internal/cli/scaffolding/... ./internal/cli/project/... ./internal/cli/repo/...` all pass
