---
phase: 01-cmderr-migration-completion
plan: 11
type: execute
wave: C
depends_on: [06, 07, 08]
files_modified:
  - internal/cli/lsof/lsof.go
  - internal/cli/ss/ss.go
  - internal/cli/uname/uname.go
  - internal/cli/yes/yes.go
  - testing/golden/golden_tests.yaml
  - tools/golden/golden_tests.yaml
  - .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
autonomous: true
requirements: [POLISH-01, POLISH-02]
must_haves:
  truths:
    - "lsof, ss classify permission and unsupported-platform errors via Pattern 6"
    - "uname, yes use Pattern 5 (write failures → ErrIO only, per CONTEXT Decision 7)"
    - "Each command has one error-path golden (yes/uname goldens stress the write path)"
  artifacts:
    - path: "internal/cli/lsof/lsof.go"
      provides: "cmderr classification"
      contains: "cmderr."
    - path: "internal/cli/ss/ss.go"
      provides: "cmderr classification"
      contains: "cmderr."
    - path: "internal/cli/uname/uname.go"
      provides: "Pattern 5 write classification"
      contains: "cmderr.ErrIO"
    - path: "internal/cli/yes/yes.go"
      provides: "Pattern 5 write classification per Pitfall 7"
      contains: "cmderr.ErrIO"
  key_links:
    - from: "yes.go write-loop"
      to: "cmderr.ErrIO"
      via: "single classification after loop per Pitfall 7"
      pattern: "cmderr.Wrap\\(cmderr.ErrIO"
---

# Plan 11 — Wave C: lsof, ss, uname, yes

## Goal

Finish Wave C by migrating the remaining system/info commands. `uname` and `yes` are Pattern 5 "no-classifiable-error" commands per CONTEXT Decision 7; `lsof` and `ss` are Pattern 6 syscall wrappers.

## Wave

Wave C.

## Requirements covered

POLISH-01, POLISH-02.

## Depends on

Plans 06, 07, 08.

## Parallelizable with

Plans 09, 10.

## Commands touched

- `internal/cli/lsof/`
- `internal/cli/ss/`
- `internal/cli/uname/`
- `internal/cli/yes/`

## Context

@.planning/phases/01-cmderr-migration-completion/01-RESEARCH.md
@.planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md

RESEARCH.md Patterns 5 and 6. Pitfall 7 (hot-loop write classification for `yes`).

## Tasks

### Task 1: Migrate `lsof` + `ss`

**Files:** `internal/cli/lsof/lsof.go`, `internal/cli/lsof/lsof_test.go`, `internal/cli/ss/ss.go`, `internal/cli/ss/ss_test.go`

**Action:**
- Permission errors reading `/proc` or netlink → `ErrPermission`.
- Unsupported platform (e.g., `ss` on macOS) → `ErrUnsupported` with locked message `<cmd>: not supported on <GOOS>`.
- Invalid filter expression → `ErrInvalidInput`.

**Verify:**
```
<automated>go test -race ./internal/cli/lsof/... ./internal/cli/ss/...</automated>
```

**Done:** Classified; tests pass.

### Task 2: Migrate `uname` (Pattern 5)

**Files:** `internal/cli/uname/uname.go`, `internal/cli/uname/uname_test.go`

**Action:** Apply Pattern 5 exactly per CONTEXT Decision 7. `uname` has no meaningful input validation (Cobra handles flags); only stdout write errors exist:

```go
if _, err := fmt.Fprintln(w, output); err != nil {
    return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("uname: write: %s", err))
}
return nil
```

**Verify:**
```
<automated>go test -race ./internal/cli/uname/...</automated>
```

**Done:** Exactly one classified branch; everything else returns nil.

### Task 3: Migrate `yes` (Pattern 5 + Pitfall 7)

**Files:** `internal/cli/yes/yes.go`, `internal/cli/yes/yes_test.go`

**Action:** Apply Pattern 5 **with Pitfall 7 mitigation** — classify after the loop, not per iteration:

```go
for {
    if _, err := fmt.Fprintln(w, s); err != nil {
        // Broken pipe is the expected termination path for `omni yes | head`.
        // Classify once; do not allocate per iteration.
        return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("yes: write: %s", err))
    }
}
```

Benchmark sanity: `omni yes | omni head -n 10` should terminate cleanly with exit 0 (broken pipe is expected downstream termination, not an error). Audit `cmd/root.go` to confirm EPIPE is mapped to exit 0 for streaming commands, or accept the `ErrIO` classification and add a golden verifying the exit code. Document the decision in `EXIT-CODE-CHANGES.md`.

**Verify:**
```
<automated>go test -race -bench=. ./internal/cli/yes/...</automated>
```

**Done:** Classified; no benchmark regression; EPIPE behavior documented.

### Task 4: Golden error snapshots

**Files:** `testing/golden/golden_tests.yaml`, `tools/golden/golden_tests.yaml`

```yaml
- name: lsof_permission_denied
  args: ["lsof", "/proc/1/mem"]
  exit_code: 3
  platforms: [linux]
- name: ss_unsupported_platform
  args: ["ss", "-tlnp"]
  exit_code: 6
  platforms: [darwin, windows]
- name: uname_basic   # happy path — uname has no error path short of broken stdout
  args: ["uname"]
  exit_code: 0
- name: yes_broken_pipe
  args: ["yes", "test"]
  exit_code: 0   # or 4 — document actual behavior in EXIT-CODE-CHANGES.md
  stdin_close_after: 0    # harness-dependent; if unsupported, skip and document
```

If platform/pipe gating isn't supported by the harness, substitute with the closest achievable snapshot and document the skip in the YAML comment.

**Verify:**
```
<automated>task test:golden -- --filter 'lsof_|ss_|uname_|yes_'</automated>
```

**Done:** Snapshots green per available platform gating.

### Task 5: Log exit-code changes

**Files:** `.planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md`

## Golden test additions

- `lsof_permission_denied` (Linux only)
- `ss_unsupported_platform` (darwin/windows only)
- `uname_basic` (happy path for completeness)
- `yes_broken_pipe` (pending harness support)

## Verification

```bash
go test -race ./internal/cli/lsof/... ./internal/cli/ss/... ./internal/cli/uname/... ./internal/cli/yes/...
task test:golden -- --filter 'lsof_|ss_|uname_|yes_'
task lint:cmderr-coverage
```

## Out of scope

- Any `pkg/*` changes
- Rich matrix for yes (Pattern 5 is minimal by design)
