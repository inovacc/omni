---
phase: 01-cmderr-migration-completion
plan: 16
type: execute
wave: Z
depends_on: [12, 13, 14, 15]
files_modified:
  - testing/golden/golden_tests.yaml
  - tools/golden/golden_tests.yaml
  - internal/cli/find/find.go
  - internal/cli/sed/sed.go
  - internal/cli/awk/awk.go
  - internal/cli/dd/dd.go
  - internal/cli/grep/grep.go
  - internal/cli/curl/curl.go
  - internal/cli/tar/tar.go
  - internal/cli/jq/jq.go
  - internal/cli/diff/diff.go
autonomous: true
requirements: [POLISH-02]
must_haves:
  truths:
    - "find, sed, awk, dd, grep, curl, tar, jq, diff each have a 3-5 snapshot rich error matrix per RESEARCH.md Risky Commands"
    - "All 9 commands' existing classification is verified correct (audit-then-add)"
  artifacts:
    - path: "testing/golden/golden_tests.yaml"
      provides: "Rich matrices for 9 risky commands (~34 new entries)"
      contains: "find_invalid_regex, sed_bad_regex, curl_unresolvable_host"
    - path: "tools/golden/golden_tests.yaml"
      provides: "Mirror of testing/golden/ rich matrix entries"
      contains: "find_invalid_regex, sed_bad_regex, curl_unresolvable_host"
  key_links:
    - from: "risky command goldens"
      to: "cmderr sentinels"
      via: "exit_code: assertion"
      pattern: "exit_code:"
---

# Plan 16 — Wave Z: Rich Error Matrices for Risky Commands

## Goal

Add the rich-matrix golden snapshots required by CONTEXT Decision 3 for `find`, `sed`, `awk`, `dd`, `grep`, `curl`, `tar`, `jq`, `diff`. These 9 commands are in the 84-already-migrated baseline but lack rich error matrices.

Per RESEARCH.md Risky Commands table, each gets 3–5 snapshots covering distinct failure classes.

## Wave

Wave Z.

## Requirements covered

POLISH-02.

## Depends on

Plans 12, 13, 14, 15 (all migration waves complete — Wave Z operates on the finished code).

## Parallelizable with

Plan 17 (docs/CLAUDE.md update). Plan 18 (CI enforcement) depends on this.

## Commands touched

9 commands, source files touched only if the audit (Task 1) reveals mis-classification. Most work is YAML editing.

## Context

@.planning/phases/01-cmderr-migration-completion/01-RESEARCH.md
@.planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md
@internal/cli/find/find.go
@internal/cli/grep/grep.go
@internal/cli/hash/hash.go

## Tasks

### Task 1: Audit existing classification for the 9 risky commands

**Files:** `internal/cli/{find,sed,awk,dd,grep,curl,tar,jq,diff}/*.go` (read-only unless audit finds bugs)

**Action:** For each command, produce an error-class inventory:

```bash
for cmd in find sed awk dd grep curl tar jq diff; do
  echo "=== $cmd ==="
  omni grep -n 'cmderr\\.Wrap\\|cmderr\\.Err' internal/cli/$cmd/
done
```

Check each `cmderr.Wrap` call against the matrix in RESEARCH.md Risky Commands. If a sentinel is wrong (e.g., `find` classifies a regex error as `ErrIO` instead of `ErrInvalidInput`), fix it in the source file and add a row to `EXIT-CODE-CHANGES.md`.

**Verify:**
```
<automated>go test -race ./internal/cli/find/... ./internal/cli/sed/... ./internal/cli/awk/... ./internal/cli/dd/... ./internal/cli/grep/... ./internal/cli/curl/... ./internal/cli/tar/... ./internal/cli/jq/... ./internal/cli/diff/...</automated>
```

**Done:** All 9 commands' existing classification verified correct; any corrections committed.

### Task 2: Add rich-matrix golden snapshots per RESEARCH.md table

**Files:** `testing/golden/golden_tests.yaml`, `tools/golden/golden_tests.yaml`

**Action:** Add the following entries to BOTH registries (35 total, listed by command):

**find (5):** invalid-regex, invalid-size, nonexistent-root, permission-denied-root (Linux only), invalid-type-flag

**sed (4):** bad-regex, bad-substitute-flags, missing-file, unsupported-command

**awk (3):** parse-error, runtime-divide-by-zero, missing-field

**dd (5):** missing-if, missing-of, bad-bs, permission-denied-of (skip if harness can't simulate), unsupported-conv

**grep (4):** file-not-found, pattern-compile, binary-file, no-match-silent-exit

**curl (5):** bad-url, unresolvable-host (use `.invalid` TLD per RESEARCH.md), connect-refused-port, http-404-with-f, http-404-without-f

**tar (4):** corrupt-header, nonexistent-archive, path-traversal-entry, unsupported-compression

**jq (4):** bad-filter-syntax, missing-field-with-`-e`, type-mismatch, invalid-json-input

**diff (3):** files-differ (exit 1 ErrConflict), missing-file, binary-files

Naming: `<cmd>_<scenario>` per RESEARCH.md §"Naming Convention".

Example entry (find, bad regex):
```yaml
- name: find_invalid_regex
  args: ["find", ".", "--regex", "[[bad"]
  exit_code: 2
```

After adding all entries: `task test:golden:update -- --filter 'find_|sed_|awk_|dd_|grep_|curl_|tar_|jq_|diff_'`.

**Verify:**
```
<automated>task test:golden -- --filter 'find_|sed_|awk_|dd_|grep_|curl_|tar_|jq_|diff_'</automated>
```

Expect 35+ passing snapshots (some may require platform gating; document skips).

**Done:** Both registries contain the full matrix; no unintended drift between them.

### Task 3: Drift check between registries

**Files:** none — verification step only

**Action:**
```bash
omni diff testing/golden/golden_tests.yaml tools/golden/golden_tests.yaml | omni head -100
```

Any diff outside the newly-added entries is a regression — investigate and fix in the same PR.

**Verify:**
```
<automated>task test:golden && task golden:compare 2>&1 | tee /tmp/golden-drift.log</automated>
```

**Done:** No unexpected drift.

## Golden test additions

~35 new entries spanning 9 risky commands (see Task 2 for complete breakdown).

## Verification

```bash
task test:golden -- --filter 'find_|sed_|awk_|dd_|grep_|curl_|tar_|jq_|diff_'
task test:golden
task lint:cmderr-coverage
```

## Out of scope

- Non-risky command rich matrices (single-snapshot minimum is sufficient per POLISH-02)
- Consolidating the two golden registries (post-1.0 per CONTEXT Decision 3)
- CI enforcement of the gate as a required check (Plan 18)
