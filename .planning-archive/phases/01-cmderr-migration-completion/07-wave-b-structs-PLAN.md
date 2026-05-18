---
phase: 01-cmderr-migration-completion
plan: 07
type: execute
wave: B
depends_on: [04, 05]
files_modified:
  - internal/cli/yamlutil/yamlutil.go
  - internal/cli/json2struct/json2struct.go
  - internal/cli/yaml2struct/yaml2struct.go
  - internal/cli/csvutil/csvutil.go
  - testing/golden/golden_tests.yaml
  - tools/golden/golden_tests.yaml
  - .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
autonomous: true
requirements: [POLISH-01, POLISH-02]
must_haves:
  truths:
    - "yamlutil, json2struct, yaml2struct, csvutil classify parse errors as ErrInvalidInput"
    - "Each has ≥1 error-path golden snapshot in both registries"
  artifacts:
    - path: "internal/cli/yamlutil/yamlutil.go"
      provides: "cmderr classification"
      contains: "cmderr.Wrap"
    - path: "internal/cli/json2struct/json2struct.go"
      provides: "cmderr classification"
      contains: "cmderr.Wrap"
    - path: "internal/cli/yaml2struct/yaml2struct.go"
      provides: "cmderr classification"
      contains: "cmderr.Wrap"
    - path: "internal/cli/csvutil/csvutil.go"
      provides: "cmderr classification"
      contains: "cmderr.Wrap"
  key_links:
    - from: "*2struct.go wrappers"
      to: "cmderr.ErrInvalidInput"
      via: "parse error classification"
      pattern: "cmderr\\.ErrInvalidInput"
---

# Plan 07 — Wave B: Struct/CSV/YAML utilities

## Goal

Migrate `yamlutil`, `json2struct`, `yaml2struct`, `csvutil` — parser-heavy commands that classify parse failures as `ErrInvalidInput` and I/O failures via Pattern 1.

## Wave

Wave B.

## Requirements covered

POLISH-01, POLISH-02.

## Depends on

Plans 04, 05 (Wave A done).

## Parallelizable with

Plans 06, 08.

## Commands touched

- `internal/cli/yamlutil/`
- `internal/cli/json2struct/`
- `internal/cli/yaml2struct/`
- `internal/cli/csvutil/`

## Context

@.planning/phases/01-cmderr-migration-completion/01-RESEARCH.md
@.planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md
@internal/cli/head/head.go
@internal/cli/find/find.go

RESEARCH.md Patterns 1 and 2.

## Tasks

### Task 1: Migrate yamlutil + csvutil

**Files:** `internal/cli/yamlutil/yamlutil.go`, `internal/cli/yamlutil/yamlutil_test.go`, `internal/cli/csvutil/csvutil.go`, `internal/cli/csvutil/csvutil_test.go`

**Action:**
- File I/O errors → Pattern 1.
- `yaml.Unmarshal` errors → `cmderr.ErrInvalidInput` with message `yamlutil: %s: %s`.
- CSV parse errors from `encoding/csv` → `cmderr.ErrInvalidInput`.
- Stdout write → `ErrIO`.

**Verify:**
```
<automated>go test -race ./internal/cli/yamlutil/... ./internal/cli/csvutil/...</automated>
```

**Done:** Error paths classified; tests pass.

### Task 2: Migrate json2struct + yaml2struct

**Files:** `internal/cli/json2struct/json2struct.go`, `internal/cli/json2struct/json2struct_test.go`, `internal/cli/yaml2struct/yaml2struct.go`, `internal/cli/yaml2struct/yaml2struct_test.go`

**Action:**
- Bad JSON/YAML input → `ErrInvalidInput`.
- Template execution errors → `ErrInvalidInput` (user-supplied template).
- File I/O → Pattern 1.

**Verify:**
```
<automated>go test -race ./internal/cli/json2struct/... ./internal/cli/yaml2struct/...</automated>
```

**Done:** Error paths classified.

### Task 3: Golden error snapshots

**Files:** `testing/golden/golden_tests.yaml`, `tools/golden/golden_tests.yaml`

**Action:** Add to both:

```yaml
- name: yamlutil_invalid_input
  args: ["yamlutil", "validate", "-"]
  stdin: "key: : bad"
  exit_code: 2
- name: csvutil_malformed
  args: ["csvutil", "tojson", "-"]
  stdin: "a,b,c\n\"unterminated"
  exit_code: 2
- name: json2struct_invalid_json
  args: ["json2struct", "-"]
  stdin: "{not json"
  exit_code: 2
- name: yaml2struct_invalid_yaml
  args: ["yaml2struct", "-"]
  stdin: ":\n  - bad: : yaml"
  exit_code: 2
```

**Verify:**
```
<automated>task test:golden -- --filter 'yamlutil_|csvutil_|json2struct_|yaml2struct_'</automated>
```

**Done:** Snapshots green in both registries.

### Task 4: Log exit-code changes

**Files:** `.planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md`

## Golden test additions

- `yamlutil_invalid_input`
- `csvutil_malformed`
- `json2struct_invalid_json`
- `yaml2struct_invalid_yaml`

## Verification

```bash
go test -race ./internal/cli/yamlutil/... ./internal/cli/csvutil/... ./internal/cli/json2struct/... ./internal/cli/yaml2struct/...
task test:golden -- --filter 'yamlutil_|csvutil_|json2struct_|yaml2struct_'
task lint:cmderr-coverage
```

## Out of scope

- Any `pkg/*` changes
- Remaining Wave B commands (Plan 08)
