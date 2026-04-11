---
phase: 01-cmderr-migration-completion
plan: 04
type: execute
wave: A
depends_on: [01, 02, 03]
files_modified:
  - internal/cli/sort/sort.go
  - internal/cli/sort/sort_test.go
  - internal/cli/env/env.go
  - internal/cli/env/env_test.go
  - internal/cli/date/date.go
  - internal/cli/date/date_test.go
  - testing/golden/golden_tests.yaml
  - tools/golden/golden_tests.yaml
  - .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
autonomous: true
requirements: [POLISH-01, POLISH-02]
must_haves:
  truths:
    - "sort, env, date return classified cmderr sentinels on every error path"
    - "Each command has at least one golden error-path snapshot in both registries"
    - "EXIT-CODE-CHANGES.md logs any exit-code shifts from the migration"
  artifacts:
    - path: "internal/cli/sort/sort.go"
      provides: "cmderr-wrapped error returns"
      contains: "cmderr.Wrap"
    - path: "internal/cli/env/env.go"
      provides: "cmderr-wrapped error returns (verify + top-up per Research Q2)"
      contains: "cmderr."
    - path: "internal/cli/date/date.go"
      provides: "cmderr-wrapped error returns"
      contains: "cmderr.Wrap"
  key_links:
    - from: "sort.go / env.go / date.go"
      to: "cmderr sentinels"
      via: "errors.Is classification at CLI boundary"
      pattern: "cmderr\\.Wrap\\(cmderr\\.Err"
    - from: "testing/golden/golden_tests.yaml"
      to: "tools/golden/golden_tests.yaml"
      via: "dual-write per CONTEXT Decision 3"
      pattern: "sort_|env_|date_"
---

# Plan 04 — Wave A: sort, env, date

## Goal

Migrate Wave A mechanical commands (`sort`, `env`, `date`) to cmderr classification per RESEARCH.md §"Wave A". `env` is verify+top-up per Research Open Q2.

## Wave

Wave A.

## Requirements covered

POLISH-01, POLISH-02.

## Depends on

Plans 01, 02, 03 (coverage gate live and ledger computed). MUST NOT start until Plan 03 confirms these three commands are still in the "needs migration" set.

## Parallelizable with

Plan 05 (Wave A kill).

## Commands touched

- `internal/cli/sort/`
- `internal/cli/env/` (verify + top-up only)
- `internal/cli/date/`

## Context

@.planning/phases/01-cmderr-migration-completion/01-RESEARCH.md
@.planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md
@internal/cli/head/head.go  # Pattern 1 reference
@internal/cli/find/find.go  # Pattern 2 reference

### Canonical wrapping patterns (by reference)

Use RESEARCH.md §"Architecture Patterns" Patterns 1, 2, and 5 verbatim:
- **Pattern 1** (file I/O): `os.ErrNotExist → ErrNotFound`, `os.ErrPermission → ErrPermission`
- **Pattern 2** (parse/validate): invalid flags/formats → `ErrInvalidInput`
- **Pattern 5** (no-classifiable): write failures → `ErrIO`, otherwise nil

Do NOT re-derive classification logic; lift from `head.go` / `find.go`.

## Tasks

### Task 1: Migrate `internal/cli/sort/sort.go`

**Files:** `internal/cli/sort/sort.go`, `internal/cli/sort/sort_test.go`

**Action:** Wrap every error return at the CLI boundary:
- File open failures → Pattern 1 (ErrNotFound / ErrPermission)
- `-c` check-sorted failures (file not sorted) → `ErrConflict` per RESEARCH.md Pattern 3
- Invalid `-k` / `-t` flag combos → `ErrInvalidInput` (only if not already gated by Cobra)
- Stdout write failures → `ErrIO`

Do NOT touch `pkg/textutil/` (CONTEXT Decision 2).

Add/extend `sort_test.go` to assert sentinel on each wrapped path with `errors.Is`.

**Verify:**
```
<automated>go test -race ./internal/cli/sort/...</automated>
```

**Done:** Every `return err` in `sort.go` is classified; tests pass; `go vet` clean.

### Task 2: Verify + top-up `internal/cli/env/env.go`

**Files:** `internal/cli/env/env.go`, `internal/cli/env/env_test.go`

**Action:** Per Research Open Q2 — env may already be partially migrated.

1. `omni grep -n 'cmderr\\.' internal/cli/env/` — record current state.
2. For every unwrapped error return, apply Patterns 1/2/5.
3. `env -i <cmd>` subprocess spawning: **verify env does NOT spawn** per CLAUDE.md "no exec" rule. If spawning is found, flag as a pre-existing bug (Risky Commands table — same disposition as `exec`) and defer the fix; classify what can be classified.

**Verify:**
```
<automated>go test -race ./internal/cli/env/...</automated>
```

**Done:** All error paths wrapped or documented as spawning-related deferrals.

### Task 3: Migrate `internal/cli/date/date.go`

**Files:** `internal/cli/date/date.go`, `internal/cli/date/date_test.go`

**Action:** Apply Patterns 2 and 5:
- Invalid format string / unparseable input → `ErrInvalidInput`
- Stdout write failure → `ErrIO`
- No file I/O → Pattern 1 not applicable

**Verify:**
```
<automated>go test -race ./internal/cli/date/...</automated>
```

**Done:** All paths classified; tests pass.

### Task 4: Add golden error snapshots (sort, env, date)

**Files:** `testing/golden/golden_tests.yaml`, `tools/golden/golden_tests.yaml`

**Action:** Add one error snapshot per command to BOTH files (dual-write rule, CONTEXT Decision 3). Naming: `<cmd>_<failure>` per RESEARCH.md §"Naming Convention".

Minimum entries:

```yaml
- name: sort_file_not_found
  args: ["sort", "nonexistent_file_xyz"]
  exit_code: 1
- name: date_invalid_format
  args: ["date", "+%INVALID_FORMAT"]
  exit_code: 2
- name: env_missing_var
  args: ["env", "-u", "DEFINITELY_UNSET_VAR_XYZ", "--"]
  exit_code: <whatever classification env uses — likely 0 if unset is ok, or 2 if strict>
```

Tune exit codes to match the actual migrated behavior. If the command change alters a previously-observed exit code, append a row to `EXIT-CODE-CHANGES.md`.

Run `task test:golden:update -- --update sort_ date_ env_` to regenerate; review diffs.

**Verify:**
```
<automated>task test:golden -- --filter 'sort_file_not_found|date_invalid_format|env_missing_var'</automated>
```

**Done:** Snapshots exist in both registries; `omni diff testing/golden/golden_tests.yaml tools/golden/golden_tests.yaml` shows no drift outside the new entries.

### Task 5: Log exit-code changes

**Files:** `.planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md`

**Action:** Append one row per command whose exit code observably changed. Columns per Plan 01 schema.

**Verify:**
```
<automated>grep -E 'sort|env|date' .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md</automated>
```

**Done:** Rows present for every changed exit code.

## Golden test additions

- `sort_file_not_found`
- `date_invalid_format`
- `env_missing_var` (scenario TBD per actual env behavior)

## Verification

```bash
go test -race ./internal/cli/sort/... ./internal/cli/env/... ./internal/cli/date/...
task test:golden -- --filter 'sort_|env_|date_'
task lint:cmderr-coverage   # must still be green
```

## Out of scope

- `kill` — Plan 05 (separate because of Windows parity / risky)
- Any Wave B/C/D command
- Touching `pkg/textutil/` or any `pkg/*` file
