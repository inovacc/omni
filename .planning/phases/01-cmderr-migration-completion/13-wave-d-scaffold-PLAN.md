---
phase: 01-cmderr-migration-completion
plan: 13
type: execute
wave: D
depends_on: [09, 10, 11]
files_modified:
  - internal/cli/scaffolding/scaffolding.go
  - internal/cli/project/project.go
  - internal/cli/repo/repo.go
  - testing/golden/golden_tests.yaml
  - tools/golden/golden_tests.yaml
  - .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
autonomous: true
requirements: [POLISH-01, POLISH-02]
must_haves:
  truths:
    - "scaffold, project, repo subcommands classify all error paths via Patterns 1/2"
    - "Each has ≥1 error-path golden"
  artifacts:
    - path: "internal/cli/scaffolding/scaffolding.go"
      provides: "cmderr classification for cobra/handler/repository/testgen subcommands"
      contains: "cmderr."
    - path: "internal/cli/project/project.go"
      provides: "cmderr classification"
      contains: "cmderr."
    - path: "internal/cli/repo/repo.go"
      provides: "cmderr classification"
      contains: "cmderr."
  key_links:
    - from: "scaffolding CLI wrapper"
      to: "pkg/*/ libraries (afero filesystem)"
      via: "Pattern 1/2 classification"
      pattern: "cmderr\\."
---

# Plan 13 — Wave D: scaffold, project, repo

## Goal

Migrate the `scaffold` subcommand family, `project`, and `repo analyze`. These are template/generator heavy and use afero for filesystem abstraction.

## Wave

Wave D.

## Requirements covered

POLISH-01, POLISH-02.

## Depends on

Plans 09, 10, 11.

## Parallelizable with

Plans 12, 14, 15.

## Commands touched

- `internal/cli/scaffolding/` (CLI wrapper only — NOT the sub-libs `cobra/handler/repository/testgen` which are pkg-like per CONTEXT Decision 2)
- `internal/cli/project/`
- `internal/cli/repo/`

## Context

@.planning/phases/01-cmderr-migration-completion/01-RESEARCH.md
@.planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md

Per RESEARCH.md §"Non-Command Exclusions": scaffolding's subcommand dispatchers ARE user-facing (`omni scaffold cobra init`), so the top-level `internal/cli/scaffolding/scaffolding.go` CLI entry wraps them. The internal sub-packages (`cobra/`, `handler/`, `repository/`, `testgen/`) stay raw-error per CONTEXT Decision 2.

## Tasks

### Task 1: Migrate scaffolding CLI entry

**Files:** `internal/cli/scaffolding/scaffolding.go`, `internal/cli/scaffolding/scaffolding_test.go`

**Action:**
- Target directory already exists → `ErrConflict` (scaffold refuses to overwrite by default).
- Template rendering errors → `ErrInvalidInput`.
- Filesystem write errors → `ErrIO`.
- Do NOT touch `internal/cli/scaffolding/{cobra,handler,repository,testgen}/*.go` internals — only the top-level dispatcher.
- If the dispatcher lives in a file named `scaffolding.go` or `scaffold.go`, wrap at that file's exported `Run*` functions.

**Verify:**
```
<automated>go test -race ./internal/cli/scaffolding/...</automated>
```

**Done:** Dispatcher-level error paths classified.

### Task 2: Migrate project + repo

**Files:** `internal/cli/project/project.go`, `internal/cli/project/project_test.go`, `internal/cli/repo/repo.go`, `internal/cli/repo/repo_test.go`

**Action:**
- `project`: directory doesn't exist → `ErrNotFound`, parse error on `go.mod` etc. → `ErrInvalidInput`, health check mismatch → `ErrConflict`.
- `repo analyze`: path resolution failures → `ErrNotFound`, remote detection with no git remote → use whatever the existing code does; classify accordingly; invalid output format flag → `ErrInvalidInput`.

**Verify:**
```
<automated>go test -race ./internal/cli/project/... ./internal/cli/repo/...</automated>
```

**Done:** Classified; tests pass.

### Task 3: Golden error snapshots

**Files:** `testing/golden/golden_tests.yaml`, `tools/golden/golden_tests.yaml`

```yaml
- name: scaffold_target_exists_conflict
  args: ["scaffold", "cobra", "init", "{dir}/existing"]
  fixtures_dir: "scaffold_existing"
  exit_code: 1
- name: project_not_a_project
  args: ["project", "info", "nonexistent_dir_xyz"]
  exit_code: 1
- name: repo_nonexistent_path
  args: ["repo", "analyze", "nonexistent_dir_xyz"]
  exit_code: 1
```

**Verify:**
```
<automated>task test:golden -- --filter 'scaffold_|project_|repo_'</automated>
```

**Done:** Snapshots green.

### Task 4: Log exit-code changes

**Files:** `.planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md`

## Golden test additions

- `scaffold_target_exists_conflict`
- `project_not_a_project`
- `repo_nonexistent_path`

## Verification

```bash
go test -race ./internal/cli/scaffolding/... ./internal/cli/project/... ./internal/cli/repo/...
task test:golden -- --filter 'scaffold_|project_|repo_'
task lint:cmderr-coverage
```

## Out of scope

- `internal/cli/scaffolding/{cobra,handler,repository,testgen}/*.go` internals (library-like code)
- pkg/* of any kind
