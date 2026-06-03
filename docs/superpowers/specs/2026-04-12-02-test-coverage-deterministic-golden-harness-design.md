# Phase 02 — Test Coverage + Deterministic Golden Harness

**Status:** Completed (02-VERIFICATION.md present but untracked in git)
**Date:** 2026-04-12
**Requirements:** POLISH-04, POLISH-05, POLISH-06, POLISH-07, POLISH-11, POLISH-12, POLISH-13, POLISH-14
**Plans:** 10 plans across 3 waves (all complete; verification score 10/10)

---

## Design / Approach / Components

Raise test coverage to v1.0 targets AND land golden-master normalization hooks (timestamps, random IDs) that supply-chain commands (Phases 4–7) will require.

**Wave structure:**

| Wave | Plans | Goal |
|------|-------|------|
| 1 — Infrastructure (parallel) | 02-01 (covgate tool), 02-02 (golden normalize hooks), 02-03 (helplint tool) | Build the measurement and enforcement infrastructure before touching tests |
| 2 — Coverage depth (after Wave 1) | 02-04 (pkg/ API baselines), 02-05 (youtube+downloader), 02-06 (jsonutil+twig), 02-07 (internal/cli/ baseline), 02-08 (happy-path goldens), 02-09 (cmdtree regen + hook) | Fill coverage gaps per risk priority |
| 3 — Cleanup (after Wave 2) | 02-10 (final gate + straggler cleanup) | Confirm all gates pass; fix remaining stragglers |

**Key components produced:**
- `tools/covgate/` — dual-rule coverage gate tool: weighted avg + per-package floor, separately for `pkg/` and `internal/cli/`.
- `tools/helplint/` — lints that every Cobra command has non-empty `Short:`, `Long:`, and at least one `omni <cmd>...` example.
- Golden `normalize:` hook — YAML field with `{pattern, replacement}` pairs (Go regexp); supported by both `testing/golden_engine.py` and `tools/golden/src/golden/`; backwards-compatible.
- `task lint:coverage` in `Taskfile.yml` wired into `.github/workflows/test.yml`.
- `task lint:help` for help docstring enforcement.
- Pre-commit hook via `task hooks:install` for cmdtree regeneration.
- Happy-path golden entries for every top-level command + key subcommands with normalize hooks applied.

---

## Rationale & Decisions

| Decision | Rationale |
|----------|-----------|
| Coverage gates dual-mode (avg + floor) | A high average can hide zero-covered packages; the floor prevents that |
| `pkg/` target: weighted avg ≥75%, floor ≥40% | Balances realistic achievability with meaningful quality bar |
| `internal/cli/` target: weighted avg ≥60%, floor ≥30% | Lower bar because CLI wrappers are harder to unit-test in isolation |
| Vendored buf subtrees excluded from both counts | They inflate counts but are not omni-owned code |
| Golden normalize hook = YAML field (not a separate file) | Keeps normalization co-located with the test entry; backwards-compatible with existing entries |
| No backreferences in normalize replacement (v1) | Simpler engine, can add in v2 if needed |
| Help docstring = loose template, example required | Avoids over-prescribing format while ensuring discoverability |
| Wave 1 infrastructure first | Coverage tools and golden hooks must exist before running coverage targets |

---

## Constraints & Assumptions

- Vendored buf packages excluded from coverage measurement throughout.
- Two golden-master engines are not consolidated — that remains post-v1.0 cleanup.
- Both engines must support the `normalize:` hook for forward compatibility with supply-chain golden tests.
- Test style: black-box (`package foo_test`) for API-contract tests; white-box allowed elsewhere.
- `pkg/video/extractor/youtube` (4% → ≥40%) and `pkg/video/downloader` (32.9% → ≥40%) are risk-prioritized in Wave 2.

---

## Testing & Acceptance

Verification report: **10/10 criteria passed**, verified 2026-04-12.

**Observable truths required:**
1. `task lint:coverage` exits 0 (all gates pass).
2. `pkg/` weighted average ≥75%, no package below 40% floor.
3. `internal/cli/` weighted average ≥60%, no package below 30% floor.
4. Both golden engines support `normalize:` hook; existing entries without it remain unchanged.
5. `tools/helplint/` reports zero violations.
6. Every top-level command has a happy-path golden entry with normalization applied.
7. cmdtree regenerated; pre-commit hook installed via `task hooks:install`.
8. CI job `coverage-gate` is required and green.

---

## Review Notes

All 10 plans complete. Anti-patterns found and addressed during verification: none flagged as blocking. `02-VERIFICATION.md` present in phase directory but untracked (not yet committed to git) — this is the anomaly flagged in the migration report.
