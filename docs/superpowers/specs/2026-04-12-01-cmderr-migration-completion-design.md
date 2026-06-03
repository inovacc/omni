# Phase 01 — cmderr Migration Completion

**Status:** Completed
**Date:** 2026-04-12
**Requirements:** POLISH-01, POLISH-02, POLISH-03
**Plans:** 18 sub-plans across 6 waves (all complete)

---

## Design / Approach / Components

Migrate every `internal/cli/*` command to return typed `cmderr` sentinels instead of raw errors, establishing the first stable exit-code contract for v1.0.

**Wave structure (strict sequential between waves; parallel within):**

| Wave | Plans | Goal |
|------|-------|------|
| 0 — Infrastructure | 01 (cmderr test baseline), 02 (Taskfile coverage gate), 03 (migration audit) | Establish test harness and authoritative work list |
| A — CI-critical | 04 (core: sort/env/date), 05 (kill + Windows parity) | Highest blast-radius commands first |
| B — Data/format/encoding | 06 (formatters), 07 (structs), 08 (idgen) | JSON, YAML, hash, encoding commands |
| C — System/proc/info | 09 (proc), 10 (disk/mem), 11 (host) | ps, kill, df, free, uptime, uname |
| D — Tail (everything else) | 12 (cloud wrappers), 13 (scaffold), 14 (dev tools), 15 (misc/tail) | Remaining alphabetical commands |
| Z — Cleanup/enforcement | 16 (risky matrices), 17 (docs), 18 (CI enforce) | Exit-code matrix, doc update, CI gate wired in |

**Key components produced:**
- `cmderr_test.go` with ≥ 90% coverage of the cmderr package.
- `EXIT-CODE-CHANGES.md` — per-command ledger of changed exit codes for release notes.
- `MIGRATION-LEDGER.md` — authoritative work list computed from live git state.
- `task lint:cmderr-coverage` in `Taskfile.yml`, wired into `.github/workflows/test.yml`.
- Golden-master error snapshots: minimum 1 per command; 3–5 for `find`, `sed`, `awk`, `dd`, `grep`, `curl`, `tar`, `jq`, `diff`.
- All 84+ previously-unclassified commands now return typed sentinels.

---

## Rationale & Decisions

| Decision | Rationale |
|----------|-----------|
| Migration order = risk-weighted waves (not alphabetical) | Primary audience is CI/CD pipelines; Wave A wins compound across most-used commands |
| `pkg/*` wrapping depth = CLI boundary only | `pkg/*` is library-surface importable by external projects; wrapping lives exclusively in `internal/cli/*` |
| Golden tests = one error snapshot per command minimum | Regression safety for exit-code contract |
| Coverage gate = custom `task lint:cmderr-coverage` | Parses coverage profile; fails if cmderr package coverage < 90% |
| Platform errors = single `ErrPermission` sentinel | `os.ErrPermission` is platform-neutral; Unix EPERM + Windows ERROR_ACCESS_DENIED both match via `errors.Is` |
| No backward-compat audit | v1.0 is the first stable exit-code contract; changed codes logged in EXIT-CODE-CHANGES.md |
| No-classifiable-error commands (`yes`, `printf`, `seq`, `pwd`, `whoami`) | Return `ErrIO` only for write failures, nil otherwise |

---

## Constraints & Assumptions

- No `pkg/*` or `cmd/*` modifications beyond already-working wiring — CLI boundary only.
- No raising of omni-wide coverage (that's Phase 2 — POLISH-04/05).
- No consolidation of the two golden registries (post-v1.0 cleanup).
- Scaffold sub-packages are the CLI boundary (not `cmd/scaffold.go`).
- `remote.go` exec violation deferred — pre-existing design-principle violation, out of scope.

---

## Testing & Acceptance

- `task lint:cmderr-coverage` passes (cmderr package ≥ 90%).
- `.github/workflows/test.yml` `cmderr-coverage` job required and green.
- Every command in `internal/cli/` returns typed cmderr sentinels (verified by `git grep`).
- Golden-master error snapshots present in both `testing/golden/golden_tests.yaml` and `tools/golden/golden_tests.yaml`.
- `EXIT-CODE-CHANGES.md` documents every changed exit code.

**Phase summary self-check: PASSED (18/18 plans complete)**

---

## Review Notes

Migration approach was risk-weighted and mechanical; no scope creep occurred. Deferred items (exit-codes doc, `cmderr.Is<Class>()` helpers, structured logging, custom golangci-lint rule) captured in `deferred-items.md` for Phase 3.
