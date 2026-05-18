---
phase: "02"
plan: "07"
subsystem: internal/cli
tags: [testing, coverage, baseline]
dependency_graph:
  requires: []
  provides: [baseline-tests-cli-stragglers]
  affects: [internal/cli/gh/hacks, internal/cli/hash, internal/cli/pager, internal/cli/readlink, internal/cli/vault, internal/cli/video]
tech_stack:
  added: []
  patterns: [table-driven tests, io.Writer pattern, temp file fixtures]
key_files:
  created:
    - internal/cli/gh/hacks/hacks_extra_test.go
    - internal/cli/hash/hash_extra_test.go
    - internal/cli/pager/pager_extra_test.go
    - internal/cli/readlink/readlink_extra_test.go
    - internal/cli/vault/vault_extra_test.go
    - internal/cli/video/auth_extra_test.go
  modified: []
decisions:
  - "Fixed hash_extra_test.go: replaced untyped string literal \"json\" with output.FormatJSON constant (Rule 1 auto-fix)"
metrics:
  duration: "~10 minutes"
  completed: "2026-04-12"
  tasks_completed: 1
  files_created: 6
---

# Phase 2 Plan 07: CLI Straggler Baseline Tests Summary

Verified and committed 6 extra test files created by a previous agent. One compilation error was auto-fixed before committing.

## Packages with New Tests

| Package | Test File | Key Coverage Added |
|---------|-----------|-------------------|
| `internal/cli/gh/hacks` | `hacks_extra_test.go` | `runGhCommand`, `runGhCommandOutput`, `PRCheckout`, `ActionsRerun`, `IssueMine`, `PRDiff`, `PRApprove`, `RepoCloneOrg` (limit default + positive) |
| `internal/cli/hash` | `hash_extra_test.go` | Binary mode, recursive directory hashing, JSON output, checksum verify (ok/quiet/status/corrupt/missing/malformed), SHA1Sum, recursive+JSON |
| `internal/cli/pager` | `pager_extra_test.go` | `RunLess`/`RunMore` not-found, `pagerModel.View()` with content/search/message/END/line-numbers/chop/highlight/scroll-percent, invalid regex in `highlightSearchMatches` and `findMatches` |
| `internal/cli/readlink` | `readlink_extra_test.go` | Missing operand, `-m` canonicalize-missing, `-e` canonicalize-existing (found/missing), `-f` canonicalize, quiet/silent mode, no-newline, zero-terminator multi-arg, `Readlink()` helper, `CanonicalPath()`, JSON output |
| `internal/cli/vault` | `vault_extra_test.go` | `classifyVaultError` (nil/401/403/404/400/500/plain), `New()` with namespace/TLS-skip/address+token, `SetToken`, `NewKV` (default/custom mount), `getTokenFile` |
| `internal/cli/video` | `auth_extra_test.go` | `copyFileIfExists` (missing/existing), `wsEncodeTextFrame`/`wsExtractPayload` (empty/too-short/unmasked/126/127/roundtrip), `isYouTubeDomain`, `isGoogleDomain`, `freePort` |

## Auto-fixed Issues

**1. [Rule 1 - Bug] hash_extra_test.go type error**
- **Found during:** compilation
- **Issue:** `OutputFormat: "json"` used untyped string constant where `output.Format` type expected
- **Fix:** Changed to `output.FormatJSON`
- **Files modified:** `internal/cli/hash/hash_extra_test.go` line 268
- **Commit:** included in `73d36047`

## Commits

| Hash | Message |
|------|---------|
| `73d36047` | `test(cli): add baseline tests for internal/cli/ stragglers [02-07]` |

## Self-Check: PASSED

All 6 test files exist and `go test -short` passes for all 6 packages.
