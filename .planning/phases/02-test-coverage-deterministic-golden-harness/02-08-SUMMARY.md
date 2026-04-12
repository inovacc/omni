---
phase: "02"
plan: "08"
subsystem: golden-harness
tags: [golden, normalize, idgen, date, non-deterministic]
dependency_graph:
  requires: ["02-02"]
  provides: ["normalize-hooks-idgen", "normalize-hooks-date"]
  affects: ["testing/golden/snapshots/"]
tech_stack:
  added: []
  patterns: ["normalize: regex replacement hooks in golden YAML"]
key_files:
  modified:
    - testing/golden/golden_tests.yaml
    - tools/golden/golden_tests.yaml
  created:
    - testing/golden/snapshots/idgen/uuid_v4.json
    - testing/golden/snapshots/idgen/uuid_v4.stdout
    - testing/golden/snapshots/idgen/uuid_v7.json
    - testing/golden/snapshots/idgen/uuid_v7.stdout
    - testing/golden/snapshots/idgen/ulid_basic.json
    - testing/golden/snapshots/idgen/ulid_basic.stdout
    - testing/golden/snapshots/idgen/ksuid_basic.json
    - testing/golden/snapshots/idgen/ksuid_basic.stdout
    - testing/golden/snapshots/idgen/nanoid_basic.json
    - testing/golden/snapshots/idgen/nanoid_basic.stdout
    - testing/golden/snapshots/idgen/snowflake_basic.json
    - testing/golden/snapshots/idgen/snowflake_basic.stdout
    - testing/golden/snapshots/date_happy/date_utc.json
    - testing/golden/snapshots/date_happy/date_utc.stdout
    - testing/golden/snapshots/date_happy/date_format_iso.json
    - testing/golden/snapshots/date_happy/date_format_iso.stdout
decisions:
  - "Used normalize: inline regex rules (not named normalizations) because idgen/date output is structurally predictable but value-unpredictable"
  - "snowflake pattern matches 15-20 digit integers to cover node-0 vs node-1 variance"
  - "ulid pattern [0-9A-Z]{26} is precise; ksuid [0-9A-Za-z]{27} covers base62 alphabet"
  - "Pre-existing failures (ps_invalid_sort_key, find_invalid_type_flag, find_invalid_maxdepth) are out of scope for this plan"
metrics:
  duration: "~10 minutes"
  completed: "2026-04-12"
  tasks: 1
  files: 2
---

# Phase 2 Plan 08: Happy-Path Golden Entries with normalize: Hooks Summary

**One-liner:** Added `normalize:` regex hooks for idgen (uuid, ulid, ksuid, nanoid, snowflake) and date commands, turning non-deterministic outputs into stable golden snapshots.

## What Was Done

Added two new golden test categories to both `testing/golden/golden_tests.yaml` and `tools/golden/golden_tests.yaml`:

### `idgen` category (6 entries)
- `uuid_v4` — `omni uuid`, normalizes UUID v4 pattern to `XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX`
- `uuid_v7` — `omni uuid -v 7`, same UUID pattern normalization
- `ulid_basic` — `omni ulid`, normalizes 26-char uppercase alphanumeric ULID
- `ksuid_basic` — `omni ksuid`, normalizes 27-char base62 KSUID
- `nanoid_basic` — `omni nanoid`, normalizes 21-char URL-safe nanoid
- `snowflake_basic` — `omni snowflake`, normalizes 15-20 digit integer snowflake ID

### `date_happy` category (2 entries)
- `date_utc` — `omni date -u`, normalizes both Unix-style and ISO timestamp formats
- `date_format_iso` — `omni date -u +%Y-%m-%dT%H:%M:%SZ`, normalizes ISO 8601 output

### Snapshot regeneration
- Ran `python testing/scripts/test_golden.py --update` with `OMNI_BIN=./bin/omni`
- Result: 195 updated, 0 errors
- Verify run: 192 pass, 3 fail (all pre-existing, unrelated to this plan)

## Deviations from Plan

None — plan executed exactly as written. The `normalize:` field name was confirmed from the golden engine source (`testing/golden_engine.py` line 31) before implementing.

## Pre-existing Failures (Out of Scope)

These 3 tests were already failing before this plan and are logged here for tracking:

| Test | Category | Issue |
|------|----------|-------|
| `ps_invalid_sort_key` | cmderr_wave_z | stdout differs from snapshot |
| `find_invalid_type_flag` | cmderr_wave_z_risky | stdout differs from snapshot |
| `find_invalid_maxdepth` | cmderr_wave_z_risky | stdout differs from snapshot |

These are pre-existing regressions, not introduced by this plan.

## Self-Check: PASSED

- `testing/golden/golden_tests.yaml` modified: confirmed
- `tools/golden/golden_tests.yaml` modified: confirmed
- Snapshots in `testing/golden/snapshots/idgen/` and `testing/golden/snapshots/date_happy/`: created by `--update` run
- Commit `b3dad1fe` exists: confirmed
