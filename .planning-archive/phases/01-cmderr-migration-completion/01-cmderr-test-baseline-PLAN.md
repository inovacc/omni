---
phase: 01-cmderr-migration-completion
plan: 01
type: execute
wave: 0
depends_on: []
files_modified:
  - internal/cli/cmderr/cmderr_test.go
  - .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
autonomous: true
requirements: [POLISH-03]
must_haves:
  truths:
    - "internal/cli/cmderr has a table-driven test suite exercising every sentinel, SilentExit, ExitError, and ExitCodeFor"
    - "go test -cover ./internal/cli/cmderr/... reports ≥90% coverage"
    - "EXIT-CODE-CHANGES.md tracking file exists and is wired for wave append"
  artifacts:
    - path: "internal/cli/cmderr/cmderr_test.go"
      provides: "≥90% coverage of cmderr package"
      contains: "ExitCodeFor, SilentExit, Wrap, ErrNotFound, ErrPermission, ErrInvalidInput, ErrIO, ErrConflict, ErrTimeout, ErrUnsupported"
    - path: ".planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md"
      provides: "Append-only log of exit-code changes across waves"
  key_links:
    - from: "cmderr_test.go"
      to: "cmderr.go"
      via: "table-driven assertions on every exported symbol"
      pattern: "ExitCodeFor.*ErrNotFound|SilentExit"
---

# Plan 01 — cmderr Test Baseline (Wave 0)

## Goal

Add `cmderr_test.go` covering ≥90% of `internal/cli/cmderr` so the future CI gate (Plan 02) doesn't fail on first run. Also create the `EXIT-CODE-CHANGES.md` tracker.

## Wave

Wave 0 — must land before any migration wave (A/B/C/D).

## Requirements covered

- POLISH-03 (partial — adds the tests the gate will measure)

## Depends on

Nothing (first plan in the phase).

## Parallelizable with

Plan 02 (Taskfile gate) — both Wave 0, independent files.

## Commands touched

None. This plan touches the cmderr package itself + the tracking file.

## Context

@.planning/phases/01-cmderr-migration-completion/01-CONTEXT.md
@.planning/phases/01-cmderr-migration-completion/01-RESEARCH.md
@internal/cli/cmderr/cmderr.go

### Interfaces (from cmderr.go per RESEARCH.md)

```go
var ErrNotFound, ErrInvalidInput, ErrPermission, ErrIO, ErrConflict, ErrTimeout, ErrUnsupported error
type SilentError struct{ Code int }
type ExitError struct{ Err error; Code int }
func SilentExit(code int) error
func WithExitCode(err error, code int) error
func Wrap(sentinel error, msg string) error
func ExitCodeFor(err error) int
```

## Tasks

### Task 1: Create `internal/cli/cmderr/cmderr_test.go`

**Files:** `internal/cli/cmderr/cmderr_test.go`

**Action:** Write a table-driven test file in `package cmderr_test` covering:

1. `TestExitCodeFor` — table with one entry per sentinel asserting the exit code per CLAUDE.md cmderr section: `ErrNotFound=1, ErrConflict=1, ErrInvalidInput=2, ErrPermission=3, ErrIO=4, ErrTimeout=5, ErrUnsupported=6`. Also assert `nil → 0` and an unclassified error → 1.
2. `TestWrap` — asserts the returned error satisfies `errors.Is(err, sentinel)` for every sentinel and `err.Error()` contains the provided message prefix.
3. `TestSilentExit` — asserts `SilentExit(n)` returns an error satisfying `errors.As(err, &cmderr.SilentError{})` with `.Code == n`, and `ExitCodeFor(SilentExit(7)) == 7`.
4. `TestExitError` / `TestWithExitCode` — asserts wrapping an error preserves the inner sentinel (`errors.Is`) and uses the explicit code from the wrapper.
5. `TestWrapDoubleWrap` — guards Pitfall 3: `Wrap(ErrIO, Wrap(ErrNotFound, "x").Error())` returns a top-level `ErrIO` error (documents current behavior so a future refactor doesn't silently change it).

Use `errors.Is` / `errors.As` everywhere; no `==` checks.

**Verify:**
```
<automated>go test -race -cover ./internal/cli/cmderr/...</automated>
```
Expect `coverage: >=90.0% of statements`.

**Done:** Test file compiles; all subtests green; `go tool cover -func=<profile>` shows ≥90% for cmderr.go.

### Task 2: Create `EXIT-CODE-CHANGES.md` tracking file

**Files:** `.planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md`

**Action:** Create a markdown file with the header:

```
# Exit-code changes introduced by Phase 1

Each wave appends rows documenting commands whose exit code changed from
un-classified (exit 1) to a classified cmderr sentinel. Used as input to
v1.0 release notes per CONTEXT.md Decision 6.

| Wave | Command | Before | After (sentinel → code) | Notes |
|------|---------|--------|-------------------------|-------|
```

**Verify:**
```
<automated>test -f .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md && head -1 .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md | grep -q "Exit-code changes"</automated>
```

**Done:** File exists with header; waves A-D can append.

## Golden test additions

None (no new commands touched). Wave A+ plans add goldens.

## Verification

```bash
go test -race -cover ./internal/cli/cmderr/...
# Expect: coverage: >=90.0% of statements
test -f .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
```

## Out of scope

- Adding the Taskfile gate target (Plan 02)
- Wiring CI (Plan 02)
- Migrating any command (Wave A+)
