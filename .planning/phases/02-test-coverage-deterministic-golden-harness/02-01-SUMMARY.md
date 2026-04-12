---
phase: "02"
plan: "01"
subsystem: "tooling/covgate"
tags: [coverage, gate, ci, quality]
dependency_graph:
  requires: []
  provides: [coverage-gate]
  affects: [Taskfile.yml, .github/workflows/test.yml]
tech_stack:
  added: []
  patterns: [stdlib-only profile parsing, dual-rule enforcement, weighted average]
key_files:
  created:
    - tools/covgate/main.go
  modified:
    - Taskfile.yml
    - .github/workflows/test.yml
decisions:
  - "Parse coverage profile directly (bufio.Scanner) — no go tool cover subprocess, preserving no-exec principle"
  - "Weighted average uses total statement count as weight, not package count"
  - "pkg/private/ excluded by default via -exclude=private flag"
metrics:
  duration: "~10 minutes"
  completed: "2026-04-12"
  tasks_completed: 3
  files_modified: 3
---

# Phase 2 Plan 01: tools/covgate — Dual-Rule Coverage Gate Summary

**One-liner:** stdlib-only Go coverage gate enforcing per-package floor and weighted average minimum without spawning subprocesses.

## What Was Built

`tools/covgate/main.go` — a single-file Go program that:
1. Accepts flags: `-profile`, `-pkg-prefix`, `-avg-min`, `-floor`, `-exclude`
2. Parses Go coverage profile format directly with `bufio.Scanner` (no `go tool cover` subprocess)
3. Groups statements by package (strips filename from import path)
4. Applies **floor rule**: any package below `-floor`% prints `FAIL pkg/name: X% < Y% floor`
5. Applies **average rule**: weighted average (weighted by statement count) below `-avg-min`% prints `FAIL weighted avg X% < Y% minimum`
6. Exits 1 on any violation, 0 on clean pass

## Files Created/Modified

| File | Change |
|------|--------|
| `tools/covgate/main.go` | Created — 160-line stdlib-only gate binary |
| `Taskfile.yml` | Added `lint:coverage` task (3 steps: test + two gate invocations) |
| `.github/workflows/test.yml` | Added `lint-coverage` CI job after `quality-check` |

## Verification Results

- `go build ./tools/covgate/` — compiled with zero errors
- Functional run against `pkg/hashutil` + `pkg/encoding`: reported `OK weighted avg 94.3% >= 1.0% minimum`, EXIT 0
- Profile parsing correctly groups by package, counts covered/total statements, applies exclusions

## Taskfile lint:coverage

```
go test -coverprofile=coverage.out -short ./pkg/... ./internal/cli/...
go run ./tools/covgate -profile=coverage.out -pkg-prefix=.../pkg/ -avg-min=75 -floor=40
go run ./tools/covgate -profile=coverage.out -pkg-prefix=.../internal/cli/ -avg-min=60 -floor=30
```

## CI Job

`lint-coverage` job runs on `ubuntu-latest`, depends on `quality-check`, timeout 15 min. Calls `task lint:coverage`.

## Deviations from Plan

None — plan executed exactly as written. The profile parsing handles the standard Go coverage line format:
`file:startLine.startCol,endLine.endCol stmts count`
where position info is on the left of the last space-separated stmts/count fields.

## Known Stubs

None.

## Threat Flags

None — tooling-only change, no new network endpoints or auth paths.

## Self-Check: PASSED

- `tools/covgate/main.go` exists and compiles
- Commit `73d1c680` confirmed in git log
