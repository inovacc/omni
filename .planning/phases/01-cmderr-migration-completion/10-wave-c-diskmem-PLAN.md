---
phase: 01-cmderr-migration-completion
plan: 10
type: execute
wave: C
depends_on: [06, 07, 08]
files_modified:
  - internal/cli/df/df.go
  - internal/cli/df/df_unix.go
  - internal/cli/df/df_windows.go
  - internal/cli/du/du.go
  - internal/cli/free/free.go
  - internal/cli/free/free_unix.go
  - internal/cli/free/free_windows.go
  - testing/golden/golden_tests.yaml
  - tools/golden/golden_tests.yaml
  - .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
autonomous: true
requirements: [POLISH-01, POLISH-02]
must_haves:
  truths:
    - "df, du, free classify syscall failures via Pattern 6"
    - "df and free use the standardized 'not supported on windows' message for parity gaps"
    - "Each has ≥1 error-path golden"
  artifacts:
    - path: "internal/cli/df/df_windows.go"
      provides: "ErrUnsupported message"
      contains: "not supported on windows"
    - path: "internal/cli/du/du.go"
      provides: "cmderr classification"
      contains: "cmderr.Wrap"
    - path: "internal/cli/free/free_windows.go"
      provides: "ErrUnsupported message"
      contains: "not supported on windows"
  key_links:
    - from: "platform files"
      to: "cmderr.ErrUnsupported"
      via: "locked message string for Phase 3 audit"
      pattern: "not supported on windows"
---

# Plan 10 — Wave C: df, du, free

## Goal

Migrate disk/memory introspection commands. `df` and `free` have Windows parity gaps per POLISH-09 and require the locked `ErrUnsupported` message. `du` is a Pattern 1 file-walker.

## Wave

Wave C.

## Requirements covered

POLISH-01, POLISH-02, anticipates POLISH-09.

## Depends on

Plans 06, 07, 08.

## Parallelizable with

Plans 09, 11.

## Commands touched

- `internal/cli/df/`
- `internal/cli/du/`
- `internal/cli/free/`

## Context

@.planning/phases/01-cmderr-migration-completion/01-RESEARCH.md
@.planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md
@internal/cli/head/head.go
@internal/cli/kill/kill_windows.go  # from Plan 05

## Tasks

### Task 1: Migrate `df` (all platform files)

**Files:** `internal/cli/df/df.go`, `internal/cli/df/df_unix.go`, `internal/cli/df/df_windows.go`, `internal/cli/df/df_test.go`

**Action:**
- Unix syscall errors → `ErrIO` / `ErrPermission` per Pattern 6.
- Windows: if any field is not supported per POLISH-09, return `cmderr.Wrap(cmderr.ErrUnsupported, "df: not supported on windows")` — same literal as Pitfall 6 standard.
- Invalid `-t <type>` filter → `ErrInvalidInput`.

**Verify:**
```
<automated>go test -race ./internal/cli/df/... && GOOS=windows go vet ./internal/cli/df/...</automated>
```

**Done:** Classified; message locked.

### Task 2: Migrate `du`

**Files:** `internal/cli/du/du.go`, `internal/cli/du/du_test.go`

**Action:**
- Walk errors on the root → `ErrNotFound` / `ErrPermission` (Pattern 1).
- Mid-walk transient errors: log-and-continue is acceptable (match POSIX `du` behavior). Final exit code reflects the worst error via single classification at the end.
- Invalid size unit flag → `ErrInvalidInput`.

**Verify:**
```
<automated>go test -race ./internal/cli/du/...</automated>
```

**Done:** Classified; tests pass.

### Task 3: Migrate `free` (all platform files)

**Files:** `internal/cli/free/free.go`, `internal/cli/free/free_unix.go`, `internal/cli/free/free_windows.go`, `internal/cli/free/free_test.go`

**Action:** Mirror `df` — Unix syscall classification; Windows → `cmderr.Wrap(cmderr.ErrUnsupported, "free: not supported on windows")`.

**Verify:**
```
<automated>go test -race ./internal/cli/free/... && GOOS=windows go vet ./internal/cli/free/...</automated>
```

**Done:** Classified; locked message matches Pitfall 6 convention.

### Task 4: Golden error snapshots

**Files:** `testing/golden/golden_tests.yaml`, `tools/golden/golden_tests.yaml`

```yaml
- name: df_file_not_found
  args: ["df", "nonexistent_path_xyz"]
  exit_code: 1
- name: du_file_not_found
  args: ["du", "nonexistent_path_xyz"]
  exit_code: 1
- name: free_invalid_unit
  args: ["free", "-u", "petabyte"]
  exit_code: 2
```

**Verify:**
```
<automated>task test:golden -- --filter 'df_|du_|free_'</automated>
```

**Done:** Snapshots green.

### Task 5: Log exit-code changes

**Files:** `.planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md`

## Golden test additions

- `df_file_not_found`, `du_file_not_found`, `free_invalid_unit`

## Verification

```bash
go test -race ./internal/cli/df/... ./internal/cli/du/... ./internal/cli/free/...
GOOS=windows go vet ./internal/cli/df/... ./internal/cli/free/...
task test:golden -- --filter 'df_|du_|free_'
task lint:cmderr-coverage
```

## Out of scope

- Full POLISH-09 Windows parity implementation (Phase 3)
- ps, pkill (Plan 09); lsof, ss, uname, yes (Plan 11)
