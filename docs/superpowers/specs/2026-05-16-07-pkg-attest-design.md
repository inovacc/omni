# Phase 07 — pkg/attest/ — SLSA Attestation

**Status:** Complete (2026-06-03) — executed via `docs/superpowers/plans/2026-06-03-phase-07-pkg-attest.md`. ADR-0009 written & accepted (honest SLSA Build **L2**, builder.id allowlist enforced in code, no numeric level field). Delivered: `pkg/attest` (pure-Go in-toto Statement v1 / SLSA Provenance v1 / DSSE PAE on stdlib — zero new deps), envelope Sign + fail-closed Verify (reuse `pkg/sign` Ed25519), `omni attest` / `attest verify` CLI + pipe `attest-verify`, 5 golden masters, and a CI `task attest:validate-schema` gate (SLSA v1.0 schema + builder.id allowlist). **Interop caveat:** omni's DSSE sig is a minisign blob — verifiable by `omni attest verify`, not by generic cosign (ADR-0009).
**Date:** 2026-05-16 (synthesized from ROADMAP — no phase directory yet)
**Requirements:** ATTEST-01 through ATTEST-09
**Depends on:** Phase 6
**ADR Gate:** `docs/adr/ADR-0009-honest-slsa-level-and-builder-id.md` (Accepted)
**Plans:** `docs/superpowers/plans/2026-06-03-phase-07-pkg-attest.md`

---

## Design / Approach / Components

Ship `omni attest` to generate and verify DSSE-wrapped, in-toto-formatted SLSA v1.0 provenance attestations at the honest SLSA level omni's own release process achieves.

**Expected components:**
- `pkg/attest/` — core attestation types and envelope builder.
- `internal/cli/attest/` — CLI wrapper for `omni attest`.
- `omni attest --predicate-type slsa-provenance --predicate <file> --artifact <path>` → in-toto Statement with SLSA v1.0 provenance predicate in DSSE envelope (PAE format).
- `omni attest verify <envelope>` → fail-closed verification with `cmderr` classification on every error mode.
- `--from-env` flag: auto-populates provenance fields from `GITHUB_RUN_ID`, `GITHUB_WORKFLOW`, `GITHUB_SHA`; `builder.id` derived from OIDC claims.
- SLSA predicate validated against official SLSA v1.0 JSON schema in CI before every release.
- Claimed SLSA level identical to ADR-pinned level (likely L2 via GitHub Actions + OIDC).

**ADR gate (must exist before code):**
- ADR pinning the honest SLSA level omni's release process achieves (almost certainly L2, not L3). No provenance output may claim a higher level than the ADR.

---

## Rationale & Decisions

| Decision | Rationale |
|----------|-----------|
| Honest SLSA level pinned by ADR before code ships | Pitfall 5 — no SLSA overclaim |
| DSSE envelope with PAE format | Validated against in-toto reference test vectors in unit tests |
| `--from-env` for GitHub Actions | Enables dogfooding in the release pipeline without hardcoding |
| Fail-closed verify | Every error mode returns non-zero with `cmderr` classification |

---

## Constraints & Assumptions

- Pure Go only — no exec.
- ADR gate must exist before any predicate builder code ships.
- Claimed level must not exceed the ADR-pinned level.
- Plan decomposition TBD at phase planning time.
- Pitfalls addressed: 5 (SLSA overclaim).

---

## Testing & Acceptance

Success criteria (from ROADMAP):
1. `omni attest --predicate-type slsa-provenance --predicate <file> --artifact <path>` produces in-toto Statement with SLSA v1.0 provenance predicate in DSSE envelope; PAE format validated against in-toto reference test vectors.
2. `omni attest verify <envelope>` with a pubkey → fail-closed; every error mode returns non-zero with `cmderr` classification.
3. `--from-env` in GitHub Actions auto-populates provenance fields from env vars; `builder.id` derived from OIDC claims, not hardcoded.
4. SLSA predicates validated against official SLSA v1.0 JSON schema in CI before every release; claimed level matches ADR-pinned level.
