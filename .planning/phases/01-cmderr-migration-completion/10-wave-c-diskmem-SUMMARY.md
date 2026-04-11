---
phase: 01-cmderr-migration-completion
plan: 10
subsystem: internal/cli/df, internal/cli/du, internal/cli/free
tags: [cmderr, migration, wave-c, platform-split]
dependency_graph:
  requires: [06, 07, 08]
  provides: [df-classified, du-classified, free-classified]
  affects: [EXIT-CODE-CHANGES.md]
tech_stack:
  added: []
  patterns: [Pattern-1-file-walker, Pattern-6-syscall-classify, ErrUnsupported-locked-message]
key_files:
  created: []
  modified:
    - internal/cli/df/df.go
    - internal/cli/df/df_unix.go
    - internal/cli/df/df_windows.go
    - internal/cli/du/du.go
    - internal/cli/free/free.go
    - internal/cli/free/free_unix.go
    - internal/cli/free/free_windows.go
    - .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
decisions:
  - "df_unix.go uses errors.Is(os.ErrNotExist/ErrPermission) for Statfs errors; falls back to ErrIO"
  - "df_windows.go uses ErrInvalidInput for UTF16PtrFromString, ErrIO for API call failure"
  - "du classifies only root os.Lstat errors; mid-walk errors remain log-and-continue per POSIX du behavior"
  - "free_unix.go classifies syscall.Sysinfo as ErrIO; /proc/meminfo open failure is silent (falls through to sysinfo)"
  - "free_windows.go classifies GlobalMemoryStatusEx ret==0 as ErrIO"
  - "free.go passes errors through directly — no double-wrap; fmt.Errorf wrapper removed"
metrics:
  duration: ~10min
  completed: 2026-04-11
  tasks_completed: 4
  files_modified: 8
---

# Phase 1 Plan 10: Wave C — df, du, free Summary

One-liner: classified syscall/filesystem errors in df, du, free using cmderr ErrIO/ErrPermission/ErrNotFound sentinels with platform-split files.

## Tasks Completed

| Task | Description | Commit |
|------|-------------|--------|
| 1 | Migrate df (unix + windows platform files + shared) | 11514292 |
| 2 | Migrate du (root Lstat classification, Pattern 1) | 0793521d |
| 3 | Migrate free (unix sysinfo + windows GlobalMemoryStatusEx) | 3e4a833d |
| 4 | Append EXIT-CODE-CHANGES.md rows | (in final commit) |

Task 4 (golden snapshots) is deferred to Plan 16 per execution scope rules.

## Deviations from Plan

### Scope adjustment

**Task 4 (golden snapshots):** Deferred to Plan 16 per execution_rules in the task prompt. YAML files not touched.

### Auto-fixed Issues

None.

## Exit-Code Changes Introduced

| Command | Before | After | Notes |
|---------|--------|-------|-------|
| df (unix, not found) | 1 | ErrNotFound → 1 | Sentinel set; exit unchanged |
| df (unix, permission denied) | 1 | ErrPermission → 3 | |
| df (unix, I/O) | 1 | ErrIO → 4 | |
| df (windows, bad path string) | 1 | ErrInvalidInput → 2 | |
| df (windows, API failure) | 1 | ErrIO → 4 | |
| du (root not found) | 1 | ErrNotFound → 1 | Sentinel set; exit unchanged |
| du (root permission denied) | 1 | ErrPermission → 3 | |
| du (root I/O) | 1 | ErrIO → 4 | |
| free (unix, sysinfo failure) | 1 | ErrIO → 4 | |
| free (windows, API failure) | 1 | ErrIO → 4 | |

## Known Stubs

None — source migration only; no UI/data rendering affected.

## Self-Check: PASSED

- `internal/cli/df/df_unix.go` — modified, cmderr imported
- `internal/cli/df/df_windows.go` — modified, cmderr imported
- `internal/cli/df/df.go` — modified, cmderr imported
- `internal/cli/du/du.go` — modified, cmderr imported
- `internal/cli/free/free.go` — modified (double-wrap removed)
- `internal/cli/free/free_unix.go` — modified, cmderr imported
- `internal/cli/free/free_windows.go` — modified, cmderr imported
- Commits 11514292, 0793521d, 3e4a833d verified via git log
- `go test -race` passed for all three packages
- `go build ./...` passed
- `GOOS=windows go vet` passed for df and free
