---
phase: 01-cmderr-migration-completion
plan: 14
type: execute
wave: D
depends_on: [09, 10, 11]
files_modified:
  - internal/cli/buf/buf.go
  - internal/cli/pipe/pipe.go
  - internal/cli/tree/tree.go
  - internal/cli/task/task.go
  - internal/cli/exec/exec.go
  - internal/cli/lint/lint.go
  - internal/cli/loc/loc.go
  - internal/cli/testcheck/testcheck.go
  - testing/golden/golden_tests.yaml
  - tools/golden/golden_tests.yaml
  - .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
autonomous: true
requirements: [POLISH-01, POLISH-02]
must_haves:
  truths:
    - "tree (Risky) returns classified errors with single first-error-wins semantics per RESEARCH.md"
    - "pipe (Risky) preserves stage classification via fmt.Errorf passthrough — no re-classification"
    - "exec is audited for spawning; any violation is logged to EXIT-CODE-CHANGES.md for backlog"
    - "buf verify+topup (Research Open Q2)"
    - "Each command has ≥1 error-path golden"
  artifacts:
    - path: "internal/cli/tree/tree.go"
      provides: "Risky: first-error-wins classification"
      contains: "cmderr."
    - path: "internal/cli/pipe/pipe.go"
      provides: "Risky: stage passthrough classification"
      contains: "fmt.Errorf"
    - path: "internal/cli/exec/exec.go"
      provides: "audited + classified (or flagged pre-existing spawn violation)"
      contains: "cmderr."
  key_links:
    - from: "pipe.go stage dispatch"
      to: "inner stage errors"
      via: "passthrough fmt.Errorf per Pitfall 3"
      pattern: "fmt.Errorf.*stage"
    - from: "tree.go"
      to: "pkg/twig/scanner errors"
      via: "wrapper classifies first error"
      pattern: "cmderr\\.Wrap"
---

# Plan 14 — Wave D: Dev tools + Risky commands (tree, pipe, exec, buf)

## Goal

Migrate the 8 developer-tool commands. Four of them are on the Risky list or need verify+topup:

- **tree** — Risky: parallel scanner, first-error-wins classification
- **pipe** — Risky: must not double-classify stage errors
- **exec** — Risky: audit for pre-existing spawn violation
- **buf** — Research Open Q2: verify+topup (claimed migrated in CLAUDE.md, may have gaps)

The remaining four (task, lint, loc, testcheck) are mechanical.

## Wave

Wave D.

## Requirements covered

POLISH-01, POLISH-02.

## Depends on

Plans 09, 10, 11.

## Parallelizable with

Plans 12, 13, 15.

## Commands touched

- `internal/cli/tree/`, `internal/cli/pipe/`, `internal/cli/exec/`, `internal/cli/buf/` (verify+topup)
- `internal/cli/task/`, `internal/cli/lint/`, `internal/cli/loc/`, `internal/cli/testcheck/`

## Context

@.planning/phases/01-cmderr-migration-completion/01-RESEARCH.md
@.planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md
@internal/cli/head/head.go
@internal/cli/find/find.go

## Tasks

### Task 1: Migrate tree (Risky) + pipe (Risky)

**Files:** `internal/cli/tree/tree.go`, `internal/cli/tree/tree_test.go`, `internal/cli/pipe/pipe.go`, `internal/cli/pipe/pipe_test.go`

**Action:**

**tree (per RESEARCH.md Risky row):**
- Root directory doesn't exist → `ErrNotFound` (Pattern 1).
- Root permission denied → `ErrPermission`.
- Scanner (`pkg/twig/scanner`) returns a first error — classify it at the CLI boundary; do NOT touch `pkg/twig/*`.
- Invalid `--compare` JSON → `ErrInvalidInput`.
- Compare with missing input file → `ErrNotFound`.
- Rich matrix optional but not required — single error-path golden is enough per CONTEXT Decision 3.

**pipe (per RESEARCH.md Risky row and Pitfall 3):**
- Stage dispatch errors → `fmt.Errorf("pipe stage %d: %w", i, err)` passthrough. The inner stage already classified; do NOT re-wrap with cmderr.
- Unknown command in pipe expression → `cmderr.ErrInvalidInput`.
- Variable substitution errors (undefined var) → `cmderr.ErrInvalidInput`.
- Verify Research Open Q2: `omni grep -n cmderr internal/cli/pipe/` — pipe claims migrated in CLAUDE but may have gaps. Top up.

**Verify:**
```
<automated>go test -race ./internal/cli/tree/... ./internal/cli/pipe/...</automated>
```

**Done:** Classification matches Risky-Commands mitigations; no double-wraps.

### Task 2: Migrate exec (Risky) + buf (verify+topup) + task/lint/loc/testcheck

**Files:**
- `internal/cli/exec/exec.go`, `internal/cli/exec/exec_test.go`
- `internal/cli/buf/buf.go`, `internal/cli/buf/buf_test.go`
- `internal/cli/task/task.go`, `internal/cli/task/task_test.go`
- `internal/cli/lint/lint.go`, `internal/cli/lint/lint_test.go`
- `internal/cli/loc/loc.go`, `internal/cli/loc/loc_test.go`
- `internal/cli/testcheck/testcheck.go`, `internal/cli/testcheck/testcheck_test.go`

**Action:**

**exec (per Risky + CLAUDE.md "no exec" rule):**
- `omni grep -n 'exec\\.Command\\|syscall.Exec' internal/cli/exec/`. If it spawns, log as a pre-existing CLAUDE.md violation in `EXIT-CODE-CHANGES.md` and classify only what doesn't spawn. If it doesn't spawn, migration is trivial (Pattern 2).

**buf:**
- `omni grep -n 'cmderr\\.' internal/cli/buf/` — CLAUDE says `buf build/format/lint` migrated; top up any subcommand that isn't.
- Build errors → `ErrInvalidInput`, lint failures → `ErrConflict`, breaking-change mismatch → `ErrConflict`.

**task/lint/loc/testcheck (mechanical Pattern 1/2):**
- File-not-found → `ErrNotFound`, bad flag / invalid input → `ErrInvalidInput`, write failure → `ErrIO`.

**Verify:**
```
<automated>go test -race ./internal/cli/exec/... ./internal/cli/buf/... ./internal/cli/task/... ./internal/cli/lint/... ./internal/cli/loc/... ./internal/cli/testcheck/...</automated>
```

**Done:** Classified; exec audit logged if violation found.

### Task 3: Golden error snapshots

**Files:** `testing/golden/golden_tests.yaml`, `tools/golden/golden_tests.yaml`

```yaml
- name: tree_root_not_found
  args: ["tree", "nonexistent_dir_xyz"]
  exit_code: 1
- name: pipe_unknown_command
  args: ["pipe", "{definitely-not-a-command}"]
  exit_code: 2
- name: exec_invalid_flag
  args: ["exec", "--definitely-not-a-flag"]
  exit_code: 2
- name: buf_lint_failure
  args: ["buf", "lint", "-"]
  stdin: "syntax = \"proto3\";\nmessage X {}"
  exit_code: 1
- name: task_file_not_found
  args: ["task", "--file", "nonexistent_taskfile.yml"]
  exit_code: 1
- name: lint_file_not_found
  args: ["lint", "nonexistent_file_xyz.go"]
  exit_code: 1
- name: loc_nonexistent_path
  args: ["loc", "nonexistent_dir_xyz"]
  exit_code: 1
- name: testcheck_nonexistent
  args: ["testcheck", "nonexistent_dir_xyz"]
  exit_code: 1
```

**Verify:**
```
<automated>task test:golden -- --filter 'tree_|pipe_|exec_|buf_|task_|lint_|loc_|testcheck_'</automated>
```

**Done:** Snapshots green.

### Task 4: Log exit-code changes + exec-spawn audit

**Files:** `.planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md`

## Golden test additions

8 snapshots listed above.

## Verification

```bash
go test -race ./internal/cli/tree/... ./internal/cli/pipe/... ./internal/cli/exec/... ./internal/cli/buf/... ./internal/cli/task/... ./internal/cli/lint/... ./internal/cli/loc/... ./internal/cli/testcheck/...
task test:golden -- --filter 'tree_|pipe_|exec_|buf_|task_|lint_|loc_|testcheck_'
task lint:cmderr-coverage
```

## Out of scope

- Fixing any exec-spawn violation (backlog)
- `pkg/twig/scanner` (library, stays raw-error per CONTEXT Decision 2)
- pipeline (already migrated per CLAUDE.md; confirm in Plan 03 ledger)
- misc-tail (Plan 15)
