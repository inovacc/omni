---
gsd_state_version: 1.0
milestone: v1.0
milestone_name: Release Cut
status: completed
stopped_at: Completed 02-08-PLAN.md
last_updated: "2026-04-12T21:54:59.344Z"
progress:
  total_phases: 8
  completed_phases: 0
  total_plans: 18
  completed_plans: 14
  percent: 78
---

# omni — Project State

## Project Reference

**Project:** omni — cross-platform, Go-native shell utility replacement (160+ commands)
**Core Value:** One static binary replaces every shell utility a Go-based CI/CD pipeline needs — deterministically, on every OS, with no external processes spawned.
**Current Focus:** v1.0 milestone — Polish → Supply chain capabilities → Release
**Audience:** me + my CI/CD pipelines (broader adoption = bonus, not driver)

## Current Position

**Milestone:** v1.0
**Phase:** Phase 2 — Test Coverage + Deterministic Golden Harness (in progress)
**Plan:** 10 plans across 3 waves (Wave 1: infra tools, Wave 2: tests + goldens, Wave 3: gate enforcement)
**Status:** 3/10 plans complete (Phase 1 complete: 18/18)
**Progress:** [████████░░] 78%

## Performance Metrics

- **Requirements mapped:** 58/58 (100%)
- **Phases defined:** 8
- **ADR gates pending:** 3 (Phase 4 entry, Phase 5 entry, Phase 7 entry)
- **Research spikes pending:** 1 (Phase 6 — OSV DB distribution)
- **Granularity:** standard

## Accumulated Context

### Key Decisions

| Decision | Phase | Rationale |
|----------|-------|-----------|
| Polish track must complete before any supply-chain phase | 1–3 before 4+ | Pitfall 13 — parallel work creates inconsistent exit-code semantics on security commands |
| Golden-master timestamp normalization lands in Phase 2, not later | 2 | Pitfall 14 — supply-chain goldens depend on it |
| Strict DAG: sign → sbom → scan → attest | 4→5→6→7 | Cryptographic primitive reuse, serialized-document boundary between sbom and scan |
| ADR gates at Phase 4, 5, 7 entries before any code | 4, 5, 7 | Scope-lock decisions (key handling, cosign-compat, round-trip matrix, honest SLSA level) |
| Minisign-only default, Sigstore verification behind build tag | 4 | Keeps default binary lean; Fulcio/Rekor/OCI rejected as v1.0 scope |
| `pkg/scan/` imports `pkg/sbom/format.Document` only, never `pkg/sbom/model` | 5/6 | Non-negotiable architectural boundary — enables third-party SBOM input |
| Honest SLSA level (likely L2) pinned by ADR before attest ships | 7 | Pitfall 5 — no SLSA overclaim |
| Phase 01-cmderr-migration-completion P10 | 10 | 4 tasks | 8 files |
| Phase 01-cmderr-migration-completion P11 | 10 | 4 tasks | 5 files |
| scaffold sub-packages are the CLI boundary (not cmd/scaffold.go) | 13 | No top-level Run* dispatcher in scaffolding.go; sub-package Run* functions own error paths |
| remote.go exec violation deferred | 13 | Pre-existing design-principle violation; out of scope for cmderr migration |
| Phase 01-cmderr-migration-completion P16 | 25 | 2 tasks | 2 files |
| Phase 01-cmderr-migration-completion P17 | 15 | 3 tasks | 4 files |
| Phase 02 P03 | 25m | 5 tasks | 65 files |
| Phase 02 P04 | 10m | 4 tasks | 7 files |
| Phase 02 P05 | 15m | 3 tasks | 6 files |
| Phase 02 P06 | 5m | 3 tasks | 2 files |
| Phase 02 P07 | 10m | 1 task | 6 files |
| Phase 02 P08 | 10m | 1 tasks | 2 files |

### Open Todos

- [ ] Run `/gsd-plan-phase 1` to decompose Phase 1 into executable plans
- [ ] Before Phase 4 entry: write ADR on key handling + cosign-compat scope
- [ ] Before Phase 5 entry: write ADR on SBOM round-trip test matrix
- [ ] Before Phase 6 entry: run `/gsd-research-phase 6` for DB distribution format spike
- [ ] Before Phase 7 entry: write ADR pinning honest SLSA level

### Blockers

None — ready to plan Phase 1.

### Research Flags

- **Phase 6 (scan):** NEEDS DEEPER RESEARCH — OSV DB distribution format, osv-scanner v2 API, grype v6 schema, `golang.org/x/vuln` public surface. Training-data stale.
- **Phase 7 (attest):** Recommended secondary research on SLSA v1.0 vs draft build-provenance schema and current GitHub Actions OIDC claim structure.
- **Phase 8 (release):** Recommended secondary research on Go toolchain reproducibility drift (`-buildvcs`, `toolchain` directive).

## Session Continuity

**Last session:** 2026-04-12T21:54:59.336Z
**Stopped at:** Completed 02-08-PLAN.md
**Next action:** `/gsd-execute-phase 2` — continue Wave 2: remaining test coverage plans

### Files of Record

- `.planning/PROJECT.md` — core value, constraints, out-of-scope
- `.planning/REQUIREMENTS.md` — 58 v1 requirements with traceability table
- `.planning/ROADMAP.md` — 8 phases, success criteria, ordering constraints
- `.planning/research/SUMMARY.md` — research synthesis
- `.planning/research/ARCHITECTURE.md` — build-order DAG
- `.planning/research/PITFALLS.md` — 17 pitfalls with phase mapping
- `.planning/research/STACK.md` — dependency recommendations
- `.planning/research/FEATURES.md` — feature categorization
- `.planning/config.json` — granularity=standard, parallelization=true, mode=normal

---

*Last updated: 2026-04-11 at initialization*
