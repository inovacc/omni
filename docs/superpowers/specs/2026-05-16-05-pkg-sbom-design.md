# Phase 05 — pkg/sbom/ — SBOM Generation

**Status:** Planned
**Date:** 2026-05-16 (synthesized from ROADMAP — no phase directory yet)
**Requirements:** SBOM-01 through SBOM-10
**Depends on:** Phase 4
**ADR Gate:** Required before any code lands (1 ADR)
**Plans:** TBD

---

## Design / Approach / Components

Ship `omni sbom` producing byte-deterministic SPDX 2.3 and CycloneDX 1.5 documents for Go modules and built Go binaries, with correct purls that downstream scanners can actually match.

**Expected components:**
- `pkg/sbom/` — core types and document builders.
- `pkg/sbom/format.Document` — the stable boundary type; `pkg/scan/` (Phase 6) imports only this, never `pkg/sbom/model`.
- `internal/cli/sbom/` — CLI wrapper for `omni sbom`.
- Go module SBOM: derived from `go.mod` + `go.sum`.
- Binary SBOM: derived from `debug/buildinfo` — "what actually shipped," Go toolchain listed as component.
- Purl normalization: module path (not import path), pseudo-version and `+incompatible` handling via `golang.org/x/mod/semver`.
- Byte-deterministic output: same input → identical bytes → enables golden-master pinning.
- Format support: `--format spdx` (SPDX 2.3 JSON), `--format cyclonedx` (CycloneDX 1.5 JSON).

**ADR gate (must exist before code):**
- SBOM round-trip test matrix ADR — which oracles (`syft convert`, `govulncheck`) validate omni's output in CI, what "purl correctness" means, how pseudo-versions/replace-directives are normalized.

---

## Rationale & Decisions

| Decision | Rationale |
|----------|-----------|
| `pkg/sbom/format.Document` boundary | Non-negotiable architectural boundary — enables third-party SBOM input into `pkg/scan/` |
| Byte-deterministic output | Enables reproducible-build extension to SBOMs and golden-master pinning |
| Binary SBOM from `debug/buildinfo` | Describes "what actually shipped" vs "what was declared" |
| Purl via `golang.org/x/mod/semver` | Correct handling of pseudo-versions and +incompatible suffixes; purl round-trip tests catch drift |

---

## Constraints & Assumptions

- Pure Go only — no exec.
- `pkg/sbom/` must be usable as standalone library after Phase 3 API triage.
- ADR gate must exist before implementation begins.
- `pkg/scan/` (Phase 6) architectural constraint: imports `pkg/sbom/format.Document` only, never `pkg/sbom/model`.
- Plan decomposition TBD at phase planning time.

---

## Testing & Acceptance

Success criteria (from ROADMAP):
1. `omni sbom <dir> --format spdx` and `--format cyclonedx` produce valid SPDX 2.3 and CycloneDX 1.5 JSON, parsed cleanly by `syft convert` in CI.
2. `omni sbom <binary>` produces an SBOM from `debug/buildinfo` describing what actually shipped; Go toolchain listed as component.
3. Every component has a correctly normalized purl; purl round-trip tests catch any drift.
4. SBOM output is byte-deterministic — two runs on the same input produce identical bytes.
