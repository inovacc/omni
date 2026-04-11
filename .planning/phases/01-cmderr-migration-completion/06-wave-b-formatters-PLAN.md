---
phase: 01-cmderr-migration-completion
plan: 06
type: execute
wave: B
depends_on: [04, 05]
files_modified:
  - internal/cli/cssfmt/cssfmt.go
  - internal/cli/htmlfmt/htmlfmt.go
  - internal/cli/sqlfmt/sqlfmt.go
  - internal/cli/xmlutil/xmlutil.go
  - testing/golden/golden_tests.yaml
  - tools/golden/golden_tests.yaml
  - .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
autonomous: true
requirements: [POLISH-01, POLISH-02]
must_haves:
  truths:
    - "cssfmt, htmlfmt, sqlfmt, xmlutil classify parse errors as ErrInvalidInput and I/O errors via Pattern 1"
    - "Each command has at least one error-path golden snapshot in both registries"
  artifacts:
    - path: "internal/cli/cssfmt/cssfmt.go"
      provides: "cmderr classification"
      contains: "cmderr.Wrap"
    - path: "internal/cli/htmlfmt/htmlfmt.go"
      provides: "cmderr classification"
      contains: "cmderr.Wrap"
    - path: "internal/cli/sqlfmt/sqlfmt.go"
      provides: "cmderr classification"
      contains: "cmderr.Wrap"
    - path: "internal/cli/xmlutil/xmlutil.go"
      provides: "cmderr classification"
      contains: "cmderr.Wrap"
  key_links:
    - from: "<fmt>.go CLI wrapper"
      to: "pkg/<fmt>/ raw error"
      via: "errors.Is passthrough to ErrInvalidInput for parse errors"
      pattern: "cmderr\\.ErrInvalidInput"
---

# Plan 06 — Wave B: Format validators (cssfmt, htmlfmt, sqlfmt, xmlutil)

## Goal

Migrate the format/validate commands that delegate to `pkg/cssfmt`, `pkg/htmlfmt`, `pkg/sqlfmt`, and `pkg/xmlfmt` (or the xmlutil wrapper). These are pure Pattern 1 + Pattern 2 work.

## Wave

Wave B.

## Requirements covered

POLISH-01, POLISH-02.

## Depends on

Wave A plans (04, 05) must be merged — Wave B inherits the green coverage gate.

## Parallelizable with

Plans 07, 08 (other Wave B plans).

## Commands touched

- `internal/cli/cssfmt/`
- `internal/cli/htmlfmt/`
- `internal/cli/sqlfmt/`
- `internal/cli/xmlutil/`

## Context

@.planning/phases/01-cmderr-migration-completion/01-RESEARCH.md
@.planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md
@internal/cli/head/head.go
@internal/cli/find/find.go

RESEARCH.md Patterns 1 and 2 apply verbatim.

## Tasks

### Task 1: Migrate cssfmt + htmlfmt

**Files:** `internal/cli/cssfmt/cssfmt.go`, `internal/cli/cssfmt/cssfmt_test.go`, `internal/cli/htmlfmt/htmlfmt.go`, `internal/cli/htmlfmt/htmlfmt_test.go`

**Action:**
- `Format` / `Minify` / `Validate` subcommands each wrap `os.ReadFile` errors via Pattern 1.
- Parse errors returned from `pkg/cssfmt.Validate` / `pkg/htmlfmt.Validate` → `cmderr.Wrap(cmderr.ErrInvalidInput, "cssfmt: parse: %s")`.
- Stdout write errors → `ErrIO`.
- Do NOT touch `pkg/cssfmt/` or `pkg/htmlfmt/`.

Extend `_test.go` files with `errors.Is` assertions on each wrapped branch.

**Verify:**
```
<automated>go test -race ./internal/cli/cssfmt/... ./internal/cli/htmlfmt/...</automated>
```

**Done:** Error paths classified; tests pass.

### Task 2: Migrate sqlfmt + xmlutil

**Files:** `internal/cli/sqlfmt/sqlfmt.go`, `internal/cli/sqlfmt/sqlfmt_test.go`, `internal/cli/xmlutil/xmlutil.go`, `internal/cli/xmlutil/xmlutil_test.go`

**Action:** Identical pattern as Task 1 but against `pkg/sqlfmt` and xml parser. `xmlutil validate` mismatches map to `ErrInvalidInput`. `xmlutil tojson` / `fromjson` parse failures → `ErrInvalidInput`.

**Verify:**
```
<automated>go test -race ./internal/cli/sqlfmt/... ./internal/cli/xmlutil/...</automated>
```

**Done:** Error paths classified; tests pass.

### Task 3: Golden error snapshots (one per command)

**Files:** `testing/golden/golden_tests.yaml`, `tools/golden/golden_tests.yaml`

**Action:** Add to both registries:

```yaml
- name: cssfmt_invalid_input
  args: ["cssfmt", "validate", "-"]
  stdin: "this is { not valid css"
  exit_code: 2
- name: htmlfmt_file_not_found
  args: ["htmlfmt", "format", "nonexistent_file_xyz"]
  exit_code: 1
- name: sqlfmt_invalid_input
  args: ["sqlfmt", "validate", "-"]
  stdin: "SELEKT * FORM"
  exit_code: 2
- name: xmlutil_invalid_xml
  args: ["xmlutil", "validate", "-"]
  stdin: "<a><b></a>"
  exit_code: 2
```

Run `task test:golden:update -- --filter 'cssfmt_|htmlfmt_|sqlfmt_|xmlutil_'` to generate snapshots.

**Verify:**
```
<automated>task test:golden -- --filter 'cssfmt_|htmlfmt_|sqlfmt_|xmlutil_'</automated>
```

**Done:** 4 new snapshots green in both registries.

### Task 4: Log exit-code changes

**Files:** `.planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md`

**Action:** Append rows for observable changes.

**Verify:**
```
<automated>grep -E 'cssfmt|htmlfmt|sqlfmt|xmlutil' .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md</automated>
```

**Done:** Rows appended.

## Golden test additions

- `cssfmt_invalid_input`
- `htmlfmt_file_not_found`
- `sqlfmt_invalid_input`
- `xmlutil_invalid_xml`

## Verification

```bash
go test -race ./internal/cli/cssfmt/... ./internal/cli/htmlfmt/... ./internal/cli/sqlfmt/... ./internal/cli/xmlutil/...
task test:golden -- --filter 'cssfmt_|htmlfmt_|sqlfmt_|xmlutil_'
task lint:cmderr-coverage
```

## Out of scope

- `pkg/*fmt/` changes (CONTEXT Decision 2)
- Other Wave B commands (covered by Plans 07, 08)
