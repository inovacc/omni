---
phase: "02"
plan: "02"
subsystem: golden-harness
tags: [golden, testing, normalization, regex]
dependency_graph:
  requires: [02-01]
  provides: [normalize-hook]
  affects: [testing/golden_engine.py, tools/golden/src/golden/]
tech_stack:
  added: []
  patterns: [regex-normalize-hook, dataclass-field-extension]
key_files:
  created: []
  modified:
    - testing/golden_engine.py
    - tools/golden/src/golden/normalize.py
    - tools/golden/src/golden/types.py
    - tools/golden/src/golden/runner.py
    - tools/golden/src/golden/discovery.py
    - docs/GOLDEN_MASTER_TESTING.md
decisions:
  - "apply_normalize_rules runs after built-in normalizations, not before, so path/newline cleanup happens first"
  - "normalize field uses list (not list[dict]) in types.py for forward-compatibility with Python 3.9"
metrics:
  duration: "~10 min"
  completed: "2026-04-12"
  tasks_completed: 4
  files_modified: 6
---

# Phase 2 Plan 02: normalize: Regex Hook for Golden Harness Engines Summary

**One-liner:** Added `normalize: [{pattern, replacement}]` YAML field to both golden harness engines so non-deterministic output (UUIDs, timestamps, ports) can be masked before snapshot diff.

## What Was Done

### Engine 1 — `testing/golden_engine.py`

The previous agent had already completed this engine fully:
- `GoldenTestCase` dataclass has `normalize: list[dict] = field(default_factory=list)`
- `GoldenEngine.normalize()` applies `re.sub(rule["pattern"], rule["replacement"], text)` for each rule after named normalizers
- `load_registry()` passes `normalize=test.get("normalize", [])` when constructing test cases

### Engine 2 — `tools/golden/src/golden/`

Four files updated:

1. **`normalize.py`** — Added `apply_normalize_rules(text, rules)` function that iterates `{pattern, replacement}` dicts and applies `re.sub`.

2. **`types.py`** — Added `normalize: list = field(default_factory=list)` to `TestCase` dataclass after `normalizations`.

3. **`runner.py`** — Updated import to include `apply_normalize_rules`; both `stdout` and `stderr` now pass through `apply_normalize_rules(..., test_case.normalize)` after the existing `normalize()` call.

4. **`discovery.py`** — Added `normalize=test.get("normalize", [])` when constructing `TestCase` from YAML.

### Documentation

`docs/GOLDEN_MASTER_TESTING.md` — Added `## normalize: Hook` section covering: syntax, common patterns table (UUID, timestamp, epoch, port), and how-it-works explanation for both engines.

## Verification

- `python testing/golden_engine.py --list` exits 0
- All 6 files committed in `4de798a2`

## Deviations from Plan

None — plan executed exactly as written. Engine 1 was already complete by the previous agent; Engine 2 work completed as specified.

## Known Stubs

None.

## Threat Flags

None — changes are test infrastructure only, no new network endpoints or auth paths.

## Self-Check: PASSED

- `testing/golden_engine.py` — exists and parses (--list exit 0)
- `tools/golden/src/golden/normalize.py` — `apply_normalize_rules` added
- `tools/golden/src/golden/types.py` — `normalize` field added
- `tools/golden/src/golden/runner.py` — import and call wired
- `tools/golden/src/golden/discovery.py` — YAML loading updated
- `docs/GOLDEN_MASTER_TESTING.md` — normalize: Hook section added
- Commit `4de798a2` verified in git log
