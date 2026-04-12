---
phase: "02"
plan: "04"
subsystem: "pkg-api-tests"
tags: ["testing", "coverage", "api-baseline"]
dependency_graph:
  requires: []
  provides: ["pkg-api-baselines"]
  affects: ["coverage-gate"]
tech_stack:
  added: []
  patterns: ["black-box package tests", "table-driven tests", "short-mode skip for network"]
key_files:
  created:
    - pkg/userdirs/userdirs_api_test.go
    - pkg/video/video_api_test.go
    - pkg/video/format/format_api_test.go
    - pkg/video/utils/utils_api_test.go
    - pkg/video/nethttp/nethttp_api_test.go
    - pkg/video/cache/cache_api_test.go
    - pkg/twig/builder/builder_api_test.go
  modified: []
decisions:
  - "Used package foo_test (black-box) style for all new files"
  - "Skipped pkg/video/types (pure struct definitions, no functions to test)"
  - "Skipped pkg/video/extractor/generic (requires HTTP — network test, not suitable for -short)"
  - "video_api_test.go uses testing.Short() skip because New() may resolve cookie paths"
metrics:
  duration: "10m"
  completed: "2026-04-12"
  tasks_completed: 4
  files_created: 7
---

# Phase 2 Plan 04: API Baseline Tests for pkg/ Packages Summary

Added black-box `_api_test.go` files for 7 pkg/ sub-packages to establish API coverage baselines, with all tests passing under `go test -short`.

## What Was Done

### Coverage Before (relevant packages)

| Package | Before |
|---------|--------|
| pkg/userdirs | 42.9% (internal-style tests only) |
| pkg/video | 46.0% |
| pkg/video/cache | 73.3% |
| pkg/video/format | 50.0% |
| pkg/video/nethttp | 61.8% |
| pkg/video/utils | 58.4% |
| pkg/twig/builder | 58.9% |
| pkg/video/extractor/generic | 0.0% |
| pkg/video/types | 0.0% (no test files) |

### Packages Given New API Test Files

| Package | New Test File | Key Functions Tested |
|---------|--------------|----------------------|
| pkg/userdirs | userdirs_api_test.go | DownloadsDir, DocumentsDir |
| pkg/video | video_api_test.go | New, WithFormat/WithQuiet/etc options |
| pkg/video/format | format_api_test.go | SortFormats, NewSelector.Select |
| pkg/video/utils | utils_api_test.go | SanitizeFilename, ParseDuration, URLJoin |
| pkg/video/nethttp | nethttp_api_test.go | NewClient, DefaultCookiePath |
| pkg/video/cache | cache_api_test.go | New, Store, Load, Remove |
| pkg/twig/builder | builder_api_test.go | DefaultBuildConfig, NewBuilder, Build(DryRun) |

### Packages Skipped

- **pkg/video/types**: Pure struct definitions only — no exported functions, nothing to call.
- **pkg/video/extractor/generic**: Requires live HTTP to test (all methods need `*nethttp.Client` + actual URL). Not suitable for `-short` baseline tests.
- **pkg/video/extractor/all**: Blank-import registration package, no testable surface.

## Deviations from Plan

**1. [Rule 1 - Bug] Corrected SanitizeFilename signature**
- Found during: Task 1 (utils test creation)
- Issue: Plan implied `SanitizeFilename(name string)` but actual signature is `SanitizeFilename(name string, restrictMode bool)`
- Fix: Added `restrictMode bool` parameter in test calls
- Files modified: pkg/video/utils/utils_api_test.go

**2. [Rule 1 - Bug] Corrected Selector.Select return signature**
- Found during: Task 1 (format test creation)
- Issue: Plan implied `Select` returns single `*types.Format` but actual signature returns `([]types.Format, error)`
- Fix: Updated test to handle both return values
- Files modified: pkg/video/format/format_api_test.go

## Self-Check: PASSED

All 7 files created and verified by `go test -short` run (all packages `ok`, no compilation errors).

Commit: `e7cb0553`
