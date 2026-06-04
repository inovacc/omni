# ADR-0010: v1.0 Release Policy

**Status:** Accepted
**Date:** 2026-06-03
**Decision:** v1.0 is cut as **six signed, reproducible binaries** (`linux/{amd64,arm64}`, `darwin/{amd64,arm64}`, `windows/{amd64,arm64}`) whose release assets are produced by **omni itself** (`omni sign`/`sbom`/`attest`), not third-party tools. Four honesty commitments are pinned by this gate: (1) the public surface of every non-`// Experimental:` `pkg/*` package is **frozen** at phase entry (CLAUDE.md 30-day deprecation protocol thereafter); (2) **reproducibility is a release gate** — a CI dual-build recompiles each target and `omni reprocheck` fails the release on any sha256 drift; (3) provenance claims **only the ADR-0009 honest level (Build L2)** — no overclaim; (4) the announcement carries an explicit "What's NOT protected against" section and an audience-scope line. No new features land in this phase — only the `omni reprocheck` tooling primitive plus configuration/CI/docs.

## Context

omni already carries `v1.0.0`–`v1.5.0` tags, but those predate the supply-chain work and overclaim stability (17 `pkg/*` packages are still `// Experimental:`, binaries carry no version stamp, releases are unsigned/undescribed). Phase 08 makes a v1.0 release that is *honest and verifiable*: signed by omni, described by an omni SBOM, attested by omni at the level it actually achieves, and reproducible so those artifacts mean something on another machine.

This is an orchestration phase: the deliverables are `.goreleaser.yaml`, `.github/workflows/release.yml`, `Taskfile.yml` targets, the pure-Go `omni reprocheck` helper, version stamping in `cmd/`, release-notes content, and docs. GoReleaser and CI shell steps run **outside** the omni binary (release machinery, exempt from no-exec); the omni binary itself stays pure-Go/no-exec/no-CGO.

This ADR is the **hard gate**: the release machinery may not be wired until these commitments are recorded.

## Analysis

| Decision | Choice | Rationale | Rejected alternative |
|----------|--------|-----------|----------------------|
| **API freeze** | At phase entry, every non-`// Experimental:` `pkg/*` package is frozen; breaking changes thereafter follow the CLAUDE.md 30-day deprecation protocol (add-alongside, `// Deprecated: … after YYYY-MM-DD`, slog warning, BACKLOG `DEPRECATION` tag, separate cleanup commit). The authoritative frozen set is enumerated by `task freeze:check` (Task 2) and snapshotted as a golden list. Experimental (NOT frozen): `pkg/sbom/{model,collect,purl}`, `pkg/scan`, `pkg/attest`, and any other `doc.go` carrying `// Experimental:`. The stable boundary `pkg/sbom/format` IS frozen (ADR-0007). | A library is not v1.0 if its API can change silently; freezing the stable surface while leaving genuinely-unstable packages marked Experimental is honest versioning. | Freeze everything (forces premature stability on churny internals) or freeze nothing (v1.0 means nothing) — both rejected. |
| **Reproducibility** | Release gate, not advisory (Pitfall 6). `CGO_ENABLED=0 -trimpath -buildvcs=true` + pinned `SOURCE_DATE_EPOCH`/`mod_timestamp`; a CI dual-build recompiles each target on a second clean checkout and `omni reprocheck` (pure-Go sha256 digest-pair diff) **fails the release** on any drift. | A non-reproducible binary makes its SBOM and signature unverifiable on another machine — the whole chain collapses. | "Best-effort reproducibility" — rejected: drift would ship silently and break downstream verification. |
| **Dogfooding** | Release assets are produced by the omni binary built in the same run: `omni sign` (archives + `checksums.txt`), `omni sbom` (per-binary SPDX), `omni attest --from-env` (per-archive SLSA L2 provenance). NOT cosign/syft/slsa-github-generator. | Validates the Phase 04–07 primitives on the most important artifact; if omni cannot sign/describe/attest its own release it is not ready to claim it does so for others. | Use established third-party tooling — rejected: defeats the dogfooding bet and re-introduces the heavy deps omni avoided. |
| **Honest level** | Provenance uses `omni attest --from-env`, which emits ONLY the ADR-0009 allowlisted release `builder.id` (Build L2). No flag claims L3. Release notes state the achieved level plainly. | Pitfall 5/15 — no SLSA overclaim; the level is the verifiable builder identity, not a marketing number. | Claim L3 / omit the level — rejected (false claim / opacity). |
| **Honest announce** | Release notes MUST carry a "What's NOT protected against" section (linking `.planning-archive/research/PITFALLS.md`) and an audience-scope line ("Built for me + my CI/CD pipelines; broader use welcome but not the design driver"). CI runs the sign/verify/sbom/attest dogfood on Windows AND Linux AND macOS so cross-platform determinism (CRLF, backslash-in-purl, case-folding — Pitfall 16) is proven, not assumed. | Honest scope + proven cross-platform parity prevent the two classic release overclaims. | A generic "production-ready for everyone" announcement — rejected (Pitfall 15). |
| **Scope** | No new features. Only `omni reprocheck` (a release-tooling primitive) + config/CI/docs. | Keeps the release cut focused and auditable; feature pressure is deferred post-v1.0. | Bundle features into the release — rejected: expands the audit surface at the worst time. |

## Consequences

- **The release is self-verifying.** Every archive ships with an omni signature, an omni SBOM, and an omni SLSA-L2 attestation; the public key is published as a release asset for offline verification.
- **Reproducibility is enforced, not hoped.** A drifting build fails CI before publish; the SBOM/signature/attestation therefore mean the same thing on any machine.
- **Versioning becomes honest.** The frozen set is explicit and golden-checked; Experimental packages keep their marker until a future triage; binaries carry a build-info-derived version (`omni --version`) that feeds SBOM/attest.
- **No overclaim.** Level is L2 by builder.id; the announcement states scope and non-protections plainly.
- **The omni binary stays pure.** `omni reprocheck` is pure stdlib; GoReleaser/CI exec is outside the binary (release machinery).
- **Cutting the tag is a deliberate operator action.** This phase delivers and commits the machinery; publishing the actual tag + GitHub release (which requires GitHub Actions enabled) is a separate, intentional trigger — not an automatic consequence of merging this work.

## Addendum — 2026-06-04 (versioning reconciliation)

- The Experimental count cited in the Context ("17 `pkg/*` packages") is now **22**
  (the gops process-tools port and extra `pkg/video` subpackages were marked
  afterwards). The frozen set is unaffected — `frozen ∩ experimental = ∅` was
  re-verified and `freeze:check` is in sync. See `docs/VERSIONING.md` for the full
  reconciliation and the per-package keep/promote rationale.
- **Tag number — decided `v1.6.0`:** `v1.0.0`–`v1.5.0` already exist (command-coverage
  milestones), so the honest supply-chain release cannot re-publish `v1.0.0`. It needs a
  **new** tag; the maintainer chose **`v1.6.0`** (2026-06-04) — the work is additive, with
  no breaking change to the frozen surface. Release notes: `docs/RELEASE-NOTES-v1.6.0.md`.
  See `docs/VERSIONING.md`. The tag is not yet cut (deliberate operator action).
