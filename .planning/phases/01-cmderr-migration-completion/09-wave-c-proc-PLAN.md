---
phase: 01-cmderr-migration-completion
plan: 09
type: execute
wave: C
depends_on: [06, 07, 08]
files_modified:
  - internal/cli/ps/ps.go
  - internal/cli/ps/ps_unix.go
  - internal/cli/ps/ps_windows.go
  - internal/cli/pkill/pkill.go
  - testing/golden/golden_tests.yaml
  - tools/golden/golden_tests.yaml
  - .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
autonomous: true
requirements: [POLISH-01, POLISH-02]
must_haves:
  truths:
    - "ps classifies Windows parity gap as ErrUnsupported with a locked message"
    - "pkill classifies bad patterns as ErrInvalidInput and permission errors as ErrPermission"
    - "Both have error-path goldens"
  artifacts:
    - path: "internal/cli/ps/ps_windows.go"
      provides: "ErrUnsupported for unsupported fields on Windows"
      contains: "cmderr.ErrUnsupported"
    - path: "internal/cli/pkill/pkill.go"
      provides: "cmderr classification"
      contains: "cmderr.Wrap"
  key_links:
    - from: "ps_windows.go"
      to: "cmderr.ErrUnsupported"
      via: "standardized message"
      pattern: "not supported on windows"
---

# Plan 09 — Wave C: ps, pkill (Risky / POLISH-09 preview)

## Goal

Migrate `ps` and `pkill`. Both touch process inspection — `ps` has a Windows parity gap per POLISH-09 and needs a locked `ErrUnsupported` message per Pitfall 6.

## Wave

Wave C.

## Requirements covered

POLISH-01, POLISH-02, anticipates POLISH-09.

## Depends on

Plans 06, 07, 08 (Wave B done).

## Parallelizable with

Plans 10, 11.

## Commands touched

- `internal/cli/ps/` (all platform files)
- `internal/cli/pkill/`

## Context

@.planning/phases/01-cmderr-migration-completion/01-RESEARCH.md
@.planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md
@internal/cli/kill/kill_windows.go  # Pattern 6 reference from Plan 05

RESEARCH.md Pattern 6 (system/syscall) and Pitfall 6 (Windows stderr stability).

## Tasks

### Task 1: Migrate `ps` (shared + platform files)

**Files:** `internal/cli/ps/ps.go`, `internal/cli/ps/ps_unix.go`, `internal/cli/ps/ps_windows.go`, `internal/cli/ps/ps_test.go`

**Action:**
- Unix: syscall errors → `cmderr.ErrIO` for read failures; permission errors → `ErrPermission`.
- Windows: fields that can't be resolved (per POLISH-09 parity gap) → `cmderr.Wrap(cmderr.ErrUnsupported, "ps: field %s not supported on windows")`. Same exact message format as kill/df/free so Wave Z Phase-3 can audit them together.
- Invalid `-o <field>` → `ErrInvalidInput`.

**Verify:**
```
<automated>go test -race ./internal/cli/ps/... && GOOS=windows go vet ./internal/cli/ps/...</automated>
```

**Done:** All branches classified with locked messages.

### Task 2: Migrate `pkill`

**Files:** `internal/cli/pkill/pkill.go`, `internal/cli/pkill/pkill_test.go`

**Action:**
- Bad regex pattern → `ErrInvalidInput`.
- No match → `cmderr.SilentExit(1)` per Pattern 4 (pkill canonical behavior).
- Permission denied killing a process → `ErrPermission`.
- Signal errors mirror kill (Plan 05).

**Verify:**
```
<automated>go test -race ./internal/cli/pkill/...</automated>
```

**Done:** Classified; tests pass.

### Task 3: Golden error snapshots

**Files:** `testing/golden/golden_tests.yaml`, `tools/golden/golden_tests.yaml`

**Action:**

```yaml
- name: ps_invalid_field
  args: ["ps", "-o", "definitely_not_a_field"]
  exit_code: 2
- name: pkill_invalid_regex
  args: ["pkill", "[[bad-regex"]
  exit_code: 2
- name: pkill_no_match_silent
  args: ["pkill", "definitely-no-process-xyz"]
  exit_code: 1
```

**Verify:**
```
<automated>task test:golden -- --filter 'ps_invalid_field|pkill_'</automated>
```

**Done:** Snapshots green.

### Task 4: Log exit-code changes

**Files:** `.planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md`

## Golden test additions

- `ps_invalid_field`
- `pkill_invalid_regex`
- `pkill_no_match_silent`

## Verification

```bash
go test -race ./internal/cli/ps/... ./internal/cli/pkill/...
task test:golden -- --filter 'ps_|pkill_'
task lint:cmderr-coverage
```

## Out of scope

- Full POLISH-09 parity fix (Phase 3)
- df, du, free (Plan 10)
- lsof, ss, uname, yes (Plan 11)
