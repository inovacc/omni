---
phase: 01-cmderr-migration-completion
plan: 08
type: execute
wave: B
depends_on: [04, 05]
files_modified:
  - internal/cli/ksuid/ksuid.go
  - internal/cli/ulid/ulid.go
  - internal/cli/nanoid/nanoid.go
  - internal/cli/snowflake/snowflake.go
  - testing/golden/golden_tests.yaml
  - tools/golden/golden_tests.yaml
  - .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
autonomous: true
requirements: [POLISH-01, POLISH-02]
must_haves:
  truths:
    - "ksuid/ulid/nanoid/snowflake classify invalid parse-or-count flags as ErrInvalidInput"
    - "Write failures classify as ErrIO (Pattern 5)"
    - "Each has ≥1 error-path golden snapshot"
  artifacts:
    - path: "internal/cli/ksuid/ksuid.go"
      provides: "cmderr classification"
      contains: "cmderr.Wrap"
    - path: "internal/cli/ulid/ulid.go"
      provides: "cmderr classification"
      contains: "cmderr.Wrap"
    - path: "internal/cli/nanoid/nanoid.go"
      provides: "cmderr classification"
      contains: "cmderr.Wrap"
    - path: "internal/cli/snowflake/snowflake.go"
      provides: "cmderr classification"
      contains: "cmderr.Wrap"
  key_links:
    - from: "idgen CLI wrappers"
      to: "pkg/idgen"
      via: "Pattern 2 wrap on invalid flag"
      pattern: "cmderr\\."
---

# Plan 08 — Wave B: ID Generators (ksuid, ulid, nanoid, snowflake)

## Goal

Migrate the 4 ID-generator CLI wrappers. These are small Pattern 2 + Pattern 5 commands.

## Wave

Wave B.

## Requirements covered

POLISH-01, POLISH-02.

## Depends on

Plans 04, 05.

## Parallelizable with

Plans 06, 07.

## Commands touched

- `internal/cli/ksuid/`
- `internal/cli/ulid/`
- `internal/cli/nanoid/`
- `internal/cli/snowflake/`

## Context

@.planning/phases/01-cmderr-migration-completion/01-RESEARCH.md
@.planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md

RESEARCH.md Patterns 2 and 5.

## Tasks

### Task 1: Migrate ksuid + ulid

**Files:** `internal/cli/ksuid/ksuid.go`, `internal/cli/ksuid/ksuid_test.go`, `internal/cli/ulid/ulid.go`, `internal/cli/ulid/ulid_test.go`

**Action:**
- Invalid `-n <count>` (negative / non-numeric if not Cobra-gated) → `ErrInvalidInput`.
- `parse <id>` invalid ID → `ErrInvalidInput`.
- Write-loop errors → `ErrIO` once after the loop per Pitfall 7 (hot-loop classification).
- Do NOT touch `pkg/idgen/`.

**Verify:**
```
<automated>go test -race ./internal/cli/ksuid/... ./internal/cli/ulid/...</automated>
```

**Done:** Classified; tests pass.

### Task 2: Migrate nanoid + snowflake

**Files:** `internal/cli/nanoid/nanoid.go`, `internal/cli/nanoid/nanoid_test.go`, `internal/cli/snowflake/snowflake.go`, `internal/cli/snowflake/snowflake_test.go`

**Action:** Same patterns as Task 1. Nanoid supports a `-size` flag — validate parse; Snowflake has a `-node` flag — validate range.

**Verify:**
```
<automated>go test -race ./internal/cli/nanoid/... ./internal/cli/snowflake/...</automated>
```

**Done:** Classified; tests pass.

### Task 3: Golden error snapshots

**Files:** `testing/golden/golden_tests.yaml`, `tools/golden/golden_tests.yaml`

**Action:**

```yaml
- name: ksuid_invalid_parse
  args: ["ksuid", "parse", "not-a-ksuid"]
  exit_code: 2
- name: ulid_invalid_parse
  args: ["ulid", "parse", "not-a-ulid"]
  exit_code: 2
- name: nanoid_invalid_size
  args: ["nanoid", "-size", "-5"]
  exit_code: 2
- name: snowflake_invalid_node
  args: ["snowflake", "-node", "-1"]
  exit_code: 2
```

**Verify:**
```
<automated>task test:golden -- --filter 'ksuid_|ulid_|nanoid_|snowflake_'</automated>
```

**Done:** Snapshots green.

### Task 4: Log exit-code changes

**Files:** `.planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md`

## Golden test additions

- `ksuid_invalid_parse`, `ulid_invalid_parse`, `nanoid_invalid_size`, `snowflake_invalid_node`

## Verification

```bash
go test -race ./internal/cli/ksuid/... ./internal/cli/ulid/... ./internal/cli/nanoid/... ./internal/cli/snowflake/...
task test:golden -- --filter 'ksuid_|ulid_|nanoid_|snowflake_'
task lint:cmderr-coverage
```

## Out of scope

- `pkg/idgen/` changes
- uuid/random/jwt (already migrated per CLAUDE.md — confirmed in Plan 03 ledger)
