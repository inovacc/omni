---
phase: 01-cmderr-migration-completion
plan: 11
subsystem: internal/cli
tags: [cmderr, migration, wave-c, yes, uname, lsof, ss]
dependency_graph:
  requires: [06, 07, 08]
  provides: [wave-c-host-cmderr]
  affects: [EXIT-CODE-CHANGES.md]
tech_stack:
  added: []
  patterns: [Pattern 5 (write-only ErrIO), Pattern 6 (syscall wrapper ErrPermission/ErrIO)]
key_files:
  created: []
  modified:
    - internal/cli/yes/yes.go
    - internal/cli/uname/uname.go
    - internal/cli/lsof/lsof.go
    - internal/cli/ss/ss.go
    - .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
decisions:
  - "yes: write errors classify as ErrIO (exit 4), not silent nil — broken-pipe is expected termination for pipes but now surfaced as classified error"
  - "lsof/ss: errors.Is(os.ErrPermission) -> ErrPermission; all other gopsutil failures -> ErrIO"
metrics:
  duration: ~10 minutes
  completed: 2026-04-11
  tasks: 4 (source only; Task 4 golden deferred to Plan 16)
  files: 5
---

# Phase 1 Plan 11: Wave C Host Commands Summary

One-liner: cmderr Pattern 5 applied to yes/uname (write-only ErrIO) and Pattern 6 to lsof/ss (permission/IO classification via errors.Is).

## Tasks Completed

| Task | Command | Pattern | Commit |
|------|---------|---------|--------|
| 1a | yes | Pattern 5 + Pitfall 7 | 8246fb44 |
| 1b | uname | Pattern 5 | 47535d30 |
| 2 | lsof | Pattern 6 (ErrPermission + ErrIO) | 66543386 |
| 3 | ss | Pattern 6 (ErrPermission + ErrIO) | 24d6bb87 |

Task 4 (golden snapshots) and Task 5 (EXIT-CODE-CHANGES logging) completed inline.

## Deviations from Plan

None — plan executed exactly as written. Golden snapshot work (Task 4) deferred to Plan 16 per execution_rules scope limit.

## Exit-Code Changes Logged

| Command | Before | After |
|---------|--------|-------|
| yes (write failure) | 0 (silent) | ErrIO → 4 |
| uname (write failure) | 0 (silent) | ErrIO → 4 |
| lsof (permission denied) | 1 | ErrPermission → 3 |
| lsof (I/O failure) | 1 | ErrIO → 4 |
| ss (JSON write failure) | 0 (silent) | ErrIO → 4 |
| ss (permission denied) | 1 | ErrPermission → 3 |
| ss (I/O failure) | 1 | ErrIO → 4 |

## Known Stubs

None.

## Threat Flags

None — no new network endpoints, auth paths, or file access patterns introduced.

## Self-Check: PASSED

- internal/cli/yes/yes.go: modified, committed 8246fb44
- internal/cli/uname/uname.go: modified, committed 47535d30
- internal/cli/lsof/lsof.go: modified, committed 66543386
- internal/cli/ss/ss.go: modified, committed 24d6bb87
- All tests: PASSED (go test -race ./internal/cli/yes/... ./internal/cli/uname/... ./internal/cli/lsof/... ./internal/cli/ss/...)
- go build ./...: PASSED
