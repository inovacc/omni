---
phase: 01-cmderr-migration-completion
plan: 17
subsystem: docs
tags: [cmderr, documentation, backlog, cleanup]
dependency_graph:
  requires: [plan-12, plan-13, plan-14, plan-15]
  provides: [stable-exit-code-contract-docs, backlog-deferrals]
  affects: [CLAUDE.md, docs/BACKLOG.md, docs/ISSUES.md, EXIT-CODE-CHANGES.md]
tech_stack:
  added: []
  patterns: []
key_files:
  modified:
    - CLAUDE.md
    - .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
    - docs/BACKLOG.md
    - docs/ISSUES.md
decisions:
  - "cmderr adoption statement changed from count-based (84) to 100% blanket statement to eliminate maintenance burden"
  - "EXIT-CODE-CHANGES.md frozen as Phase 1 artifact; release-notes template added inline"
metrics:
  duration: 15m
  completed: 2026-04-11
  tasks: 3
  files: 4
---

# Phase 1 Plan 17: Wave Z — Documentation Update + Backlog Sync Summary

One-liner: Updated CLAUDE.md to declare 100% cmderr adoption, finalized EXIT-CODE-CHANGES.md with release-notes template, and recorded all Phase 1 deferrals in BACKLOG.md.

## Tasks Completed

| # | Task | Commit | Files |
|---|------|--------|-------|
| 1 | Update CLAUDE.md cmderr section | `777741d1` | CLAUDE.md |
| 2 | Finalize EXIT-CODE-CHANGES.md | `472d8782` | EXIT-CODE-CHANGES.md |
| 3 | Add backlog rows + close ISSUES entry | `6cae8ef6` | docs/BACKLOG.md, docs/ISSUES.md |

## Changes Made

**CLAUDE.md:** Replaced `Commands adopted (84): <long list>` + `Commands NOT yet adopted: ~76 remaining` with a single 100% adoption statement plus the stable exit-code contract notice and link to EXIT-CODE-CHANGES.md.

**EXIT-CODE-CHANGES.md:** Added `## Summary: 84 commands changed exit codes during Phase 1` and `## Release-notes template` sections. File is now publication-ready for v1.0 release notes.

**docs/BACKLOG.md:**
- Marked cmderr P0 item complete
- Added Phase 2 follow-ups: golangci-lint no-raw-os-err rule, cross-command exit-code golden matrix
- Added Phase 3 deferrals: `cmderr.Is<Class>()` helpers, `docs/EXIT-CODES.md` generation
- Added no-exec violation backlog rows for `exec/exec.go` and `repo/remote.go`
- Added flaky Windows test backlog row for `detector_test.go::TestDetectGo_PrivateWithNetrc`

**docs/ISSUES.md:** Closed the cmderr adoption tracking issue (was P1, now resolved Apr 2026).

## Deviations from Plan

None — plan executed exactly as written.

## Known Stubs

None.

## Self-Check: PASSED

- CLAUDE.md: "Commands NOT yet adopted" removed (0 occurrences)
- EXIT-CODE-CHANGES.md: "Release-notes template" present (1 occurrence)
- docs/BACKLOG.md: contains "exit code" and "Phase 1"
- Commits 777741d1, 472d8782, 6cae8ef6 exist
