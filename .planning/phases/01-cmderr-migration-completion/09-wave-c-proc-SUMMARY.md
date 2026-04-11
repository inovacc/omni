---
phase: 01-cmderr-migration-completion
plan: 09
subsystem: internal/cli/ps, internal/cli/pkill
tags: [cmderr, migration, wave-c, proc, platform-split]
dependency_graph:
  requires: [06-wave-b-format, 07-wave-b-format2, 08-wave-b-format3]
  provides: [ps-cmderr, pkill-cmderr]
  affects: [EXIT-CODE-CHANGES.md, MIGRATION-LEDGER.md]
tech_stack:
  added: []
  patterns: [platform-split checkPlatformSupport, SilentExit pattern-4, ErrUnsupported locked-message]
key_files:
  created: []
  modified:
    - internal/cli/ps/ps.go
    - internal/cli/ps/ps_unix.go
    - internal/cli/ps/ps_windows.go
    - internal/cli/pkill/pkill.go
    - internal/cli/pkill/pkill_test.go
    - .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
decisions:
  - "SilentExit(1) for pkill no-match (Pattern 4): matches POSIX pkill behavior, no message on no-match"
  - "Locked ErrUnsupported message 'ps: field user not supported on windows' pins POLISH-09 audit format"
  - "checkPlatformSupport() defined in platform files (ps_unix.go no-op, ps_windows.go rejects -u)"
  - "ps invalid sort key validated before GetProcessList call — ErrInvalidInput → exit 2"
metrics:
  duration: ~15min
  completed: 2026-04-11
  tasks_completed: 3
  tasks_deferred: 1
  files_modified: 6
---

# Phase 1 Plan 09: Wave C proc (ps, pkill) Summary

## One-liner

ps + pkill cmderr migration with platform-split ErrUnsupported for Windows user-field parity gap and SilentExit(1) for pkill no-match.

## What Was Built

### Task 1: ps migration (commit c1b2fbb9)

**ps.go (shared):**
- Added `cmderr` import
- `validSortKeys` map — invalid `--sort` value returns `ErrInvalidInput → exit 2`
- `checkPlatformSupport(opts)` called before `GetProcessList` — platform-specific gate
- Both `Run()` and `RunTop()` propagate classified errors from `GetProcessList`

**ps_unix.go:**
- Added `errors`, `cmderr` imports
- `/proc` read: `os.ErrPermission` → `ErrPermission → exit 3`; other failures → `ErrIO → exit 4`
- `checkPlatformSupport` no-op (all options supported on Unix)

**ps_windows.go:**
- Added `cmderr` import
- `CreateToolhelp32Snapshot` failure → `ErrIO → exit 4`
- `Process32First` failure → `ErrIO → exit 4`
- `User` field changed from `"SYSTEM"` to `"?"` with locked comment: `"ps: field user not supported on windows"` (POLISH-09 audit pin)
- `checkPlatformSupport`: `-u` user filter → `ErrUnsupported → exit 6`

### Task 2: pkill migration (commit 63947dc4)

**pkill.go:**
- Added `errors`, `cmderr` imports
- Empty pattern → `ErrInvalidInput → exit 2`
- Invalid regex → `ErrInvalidInput → exit 2`
- Invalid signal name/number → `ErrInvalidInput → exit 2`
- `process.Processes()` failure → `ErrIO → exit 4`
- No match → `cmderr.SilentExit(1)` (Pattern 4 — pkill canonical behavior)
- `proc.Signal()` permission error → `ErrPermission` classified in `Result.Error`

**pkill_test.go:**
- Added `errors`, `cmderr` imports
- `TestRun_ExactMatch`: updated to assert `SilentExit(1)` on no-match
- `TestRun_NoMatch_JSON`: updated to assert `SilentExit(1)` + `"[]"` output
- `TestRun_SignalParsing`: valid signals accept `nil` or `SilentExit(1)`; invalid signals must NOT be `SilentExit`

### Task 3: Golden snapshots

**DEFERRED to Plan 16** per execution rules. No golden files touched.

### Task 4: EXIT-CODE-CHANGES.md

12 rows appended for Wave C ps/pkill.

## Deviations from Plan

### Auto-fixed Issues

None.

### Scope Adjustments

**1. [Deferred] Golden snapshots (Task 3)**
- Plan 09 included golden snapshot additions for `ps_invalid_field`, `pkill_invalid_regex`, `pkill_no_match_silent`
- Per execution rules: "DEFER all golden snapshot work to Plan 16"
- No golden files touched

**2. [Rule 2 - Missing functionality] pkill_test.go updated for SilentExit behavior**
- Found during: Task 2
- Issue: Existing tests `TestRun_ExactMatch` and `TestRun_NoMatch_JSON` expected `nil` on no-match; after SilentExit(1) change they would fail
- Fix: Updated both tests + `TestRun_SignalParsing` to assert correct `SilentExit(1)` behavior
- Files modified: `internal/cli/pkill/pkill_test.go`
- Commit: 63947dc4

## Known Stubs

None. No placeholder values introduced.

## Threat Flags

None. No new network endpoints, auth paths, or trust boundary changes.

## Self-Check: PASSED

- `internal/cli/ps/ps.go` — modified ✓
- `internal/cli/ps/ps_unix.go` — modified ✓
- `internal/cli/ps/ps_windows.go` — modified ✓
- `internal/cli/pkill/pkill.go` — modified ✓
- `internal/cli/pkill/pkill_test.go` — modified ✓
- Commit c1b2fbb9 (ps) — exists ✓
- Commit 63947dc4 (pkill) — exists ✓
- Tests: `go test -race ./internal/cli/ps/... ./internal/cli/pkill/...` — PASS ✓
- `go build ./...` — PASS ✓
