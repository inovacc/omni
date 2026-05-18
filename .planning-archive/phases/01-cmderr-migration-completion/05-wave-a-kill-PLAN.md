---
phase: 01-cmderr-migration-completion
plan: 05
type: execute
wave: A
depends_on: [01, 02, 03]
files_modified:
  - internal/cli/kill/kill.go
  - internal/cli/kill/kill_unix.go
  - internal/cli/kill/kill_windows.go
  - internal/cli/kill/kill_test.go
  - testing/golden/golden_tests.yaml
  - tools/golden/golden_tests.yaml
  - .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
autonomous: true
requirements: [POLISH-01, POLISH-02]
must_haves:
  truths:
    - "kill returns ErrInvalidInput for bad PID, ErrUnsupported for Windows-unsupported signals, ErrPermission for EPERM"
    - "Windows-unsupported signal message is stable and locked by a platform-agnostic golden"
    - "3+ error-path golden snapshots exist per Risky-Commands matrix"
  artifacts:
    - path: "internal/cli/kill/kill.go"
      provides: "classified error returns with platform-split"
      contains: "cmderr.ErrInvalidInput, cmderr.ErrUnsupported, cmderr.ErrPermission"
    - path: "internal/cli/kill/kill_windows.go"
      provides: "ErrUnsupported for signals outside INT/KILL/TERM"
      contains: "cmderr.ErrUnsupported"
  key_links:
    - from: "kill_windows.go"
      to: "cmderr.ErrUnsupported"
      via: "not supported on windows message"
      pattern: "not supported on windows"
---

# Plan 05 — Wave A: kill (Risky / Platform-Sensitive)

## Goal

Migrate `kill` with the 3-snapshot risky-command matrix from RESEARCH.md §"Risky-Commands List" and standardize the Windows `ErrUnsupported` message per Pitfall 6.

## Wave

Wave A. Split from Plan 04 because (a) it is on the Risky list and (b) it touches platform-specific files that the mechanical commands don't.

## Requirements covered

POLISH-01, POLISH-02, and anticipates POLISH-09 (Windows parity) which lands fully in Phase 3 — this plan only pins the `ErrUnsupported` message format.

## Depends on

Plans 01, 02, 03.

## Parallelizable with

Plan 04.

## Commands touched

- `internal/cli/kill/` (all three platform files)

## Context

@.planning/phases/01-cmderr-migration-completion/01-RESEARCH.md
@.planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md
@internal/cli/find/find.go  # Pattern 2 reference

### Canonical wrapping patterns (by reference)

RESEARCH.md Patterns 2 (validate) and 6 (system/syscall). Risky-Commands mitigation row for `kill`:
- bad-pid → `ErrInvalidInput`
- unsupported-signal-windows → `ErrUnsupported` with fixed message `kill: signal %s not supported on windows`
- permission-denied-pid → `ErrPermission`

## Tasks

### Task 1: Migrate `kill.go` (shared entry)

**Files:** `internal/cli/kill/kill.go`, `internal/cli/kill/kill_test.go`

**Action:**
1. Wrap PID parse failures (`strconv.Atoi`) as `cmderr.ErrInvalidInput` with message `kill: invalid pid: <value>`.
2. Wrap unknown signal-name lookups as `cmderr.ErrInvalidInput` with message `kill: unknown signal: <name>`.
3. Delegate actual signal dispatch to the platform-specific files and pass-through their errors with `fmt.Errorf("kill: %w", err)` — the platform files already classify.

**Verify:**
```
<automated>go test -race ./internal/cli/kill/...</automated>
```

**Done:** Shared-code error paths classified; tests pass on current platform.

### Task 2: Wrap `kill_unix.go`

**Files:** `internal/cli/kill/kill_unix.go`

**Action:** Convert raw `syscall` / `os.Process.Signal` errors to cmderr:
- `errors.Is(err, os.ErrPermission)` or `errors.Is(err, syscall.EPERM)` → `cmderr.ErrPermission` with message `kill: permission denied: pid %d`
- `errors.Is(err, syscall.ESRCH)` (no such process) → `cmderr.ErrNotFound` with message `kill: no such process: pid %d`
- Other syscall errors → `fmt.Errorf("kill: %w", err)` passthrough

**Verify:**
```
<automated>go test -race -tags unix ./internal/cli/kill/...</automated>
```

**Done:** Unix branch classified.

### Task 3: Wrap `kill_windows.go`

**Files:** `internal/cli/kill/kill_windows.go`

**Action:** Per CLAUDE.md "Windows signals (INT, KILL, TERM only)" and Pitfall 6 standardization:

```go
if !isSupportedSignal(sig) {
    return cmderr.Wrap(cmderr.ErrUnsupported,
        fmt.Sprintf("kill: signal %s not supported on windows (INT/KILL/TERM only)", sig))
}
```

Bake the message literal into `kill_windows.go` exactly — the golden snapshot pins it.

**Verify:**
```
<automated>go test -race -tags windows ./internal/cli/kill/...</automated>
```
(Or cross-compile sanity check: `GOOS=windows go build ./internal/cli/kill/...`)

**Done:** Windows branch returns `ErrUnsupported` with locked message for any signal outside the supported 3.

### Task 4: Risky-matrix golden snapshots

**Files:** `testing/golden/golden_tests.yaml`, `tools/golden/golden_tests.yaml`

**Action:** Add 3 golden entries to both registries per the Risky Commands matrix:

```yaml
- name: kill_bad_pid
  args: ["kill", "not-a-pid"]
  exit_code: 2
- name: kill_nonexistent_pid
  args: ["kill", "999999999"]
  exit_code: 1        # ErrNotFound (ESRCH)
- name: kill_unsupported_signal_windows
  args: ["kill", "-USR1", "1"]
  exit_code: 6        # ErrUnsupported
  platforms: [windows]   # or use platform gating if harness supports it
```

If the golden harness lacks platform gating, fall back to skipping the Windows-only entry on Linux/macOS CI and document the skip in the snapshot comment.

**Verify:**
```
<automated>task test:golden -- --filter 'kill_'</automated>
```

**Done:** Snapshots green on each supported platform.

### Task 5: Log exit-code changes

**Files:** `.planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md`

**Action:** Append kill rows for any observable exit-code change.

**Verify:**
```
<automated>grep kill .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md</automated>
```

**Done:** Rows appended.

## Golden test additions

- `kill_bad_pid`
- `kill_nonexistent_pid`
- `kill_unsupported_signal_windows` (platform-gated)

## Verification

```bash
go test -race ./internal/cli/kill/...
GOOS=windows go vet ./internal/cli/kill/...
task test:golden -- --filter 'kill_'
task lint:cmderr-coverage
```

## Out of scope

- Full POLISH-09 Windows parity work (Phase 3)
- `pkill`, `ps`, `df`, `free` (Wave C plans)
- Any change to `cmd/kill.go` beyond surfacing new errors
