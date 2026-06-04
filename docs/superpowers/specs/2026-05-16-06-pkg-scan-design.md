# Phase 06 — pkg/scan/ — Vulnerability Scanning

**Status:** Complete (2026-06-03) — executed via `docs/superpowers/plans/2026-06-03-phase-06-pkg-scan.md`. ADR-0008 written & decided; research spike → `docs/superpowers/research/phase-06/RESEARCH.md`. Delivered: `pkg/scan` (pure-Go OSV matcher, CVSS v3.1 severity, signed-DB load/verify, `--max-db-age` staleness), `pkg/sbom/format.Parse`+`Document.Components()`+`purl.Parse` (read side of the boundary), `omni scan` / `scan db update` CLI, pipe integration, golden masters (8 scan + reader round-trips). **Reachability dropped from v1.0 (SCAN-02 deferred):** `omni scan source` → `ErrUnsupported`; `golang.org/x/vuln` violates both no-exec and lean-`go.mod` (MVS) — future `contrib/govulncheck-scan` module.
**Date:** 2026-05-16 (synthesized from ROADMAP — no phase directory yet)
**Requirements:** SCAN-01 through SCAN-09 (SCAN-02 reachability deferred per ADR-0008)
**Depends on:** Phase 5
**Research spike:** Done — `docs/superpowers/research/phase-06/RESEARCH.md` (OSV DB distribution, `golang.org/x/vuln` API/exec/weight, OSV schema + CVSS bands, reachability feasibility)
**Plans:** `docs/superpowers/plans/2026-06-03-phase-06-pkg-scan.md`
**ADR Gate:** `docs/adr/ADR-0008-pure-go-vuln-scan-and-signed-osv-db.md` (Accepted)

---

## Design / Approach / Components

Ship `omni scan` to scan an SBOM or Go source tree against a signed OSV vulnerability database and gate CI on severity thresholds, with reachability analysis for Go source to eliminate false positives.

**Expected components:**
- `pkg/scan/` — imports `pkg/sbom/format.Document` only; core scanning logic.
- `internal/cli/scan/` — CLI wrapper for `omni scan`.
- SBOM scan: `omni scan <sbom>` on SPDX or CycloneDX document → findings in JSON and text.
- Source scan: `omni scan source <path>` on Go source directory → reachability-aware findings via `golang.org/x/vuln`.
- Severity gating: `omni scan --fail-on <severity>` → `cmderr.ErrConflict` when any finding meets/exceeds threshold.
- DB management: `omni scan db update` downloads latest OSV database, signed with `pkg/sign/`, verified on load.
- Offline mode with `--max-db-age` staleness gate (stale DB fails loudly, never silently degrades).

**Research spike required before planning:**
- OSV DB distribution format (flat JSON vs bbolt).
- OSV-scanner v2 API surface.
- grype v6 schema direction.
- `golang.org/x/vuln` public scan API shape.

---

## Rationale & Decisions

| Decision | Rationale |
|----------|-----------|
| `pkg/scan/` imports `pkg/sbom/format.Document` only | Enables third-party SBOM input; strict architectural boundary |
| Reachability analysis via `golang.org/x/vuln` | Eliminates false positives on unreachable symbols |
| DB signed with `pkg/sign/` | Tampered DBs fail closed; reuses Phase 4 primitive |
| `--max-db-age` staleness gate | Stale DB fails loudly — never silent degradation |

---

## Constraints & Assumptions

- Pure Go only — no exec.
- Must not import `pkg/sbom/model` (only `pkg/sbom/format.Document`).
- Research spike must complete before plan generation.
- No ADR gate formally required, but research spike output drives key decisions.
- Plan decomposition TBD after research spike.

---

## Testing & Acceptance

Success criteria (from ROADMAP):
1. `omni scan <sbom>` on SPDX/CycloneDX document → findings in JSON and text; `--fail-on <severity>` returns `cmderr.ErrConflict` — CI gating path is golden-tested.
2. `omni scan source <path>` on Go source → reachability-aware findings; only vulnerabilities where the vulnerable symbol is actually called are reported.
3. `omni scan db update` downloads and verifies latest OSV DB signed with `pkg/sign/`; tampered DBs fail closed.
4. Offline mode works with cached DB; `--max-db-age` gates staleness — stale DB fails loudly, never silently.
