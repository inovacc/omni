# omni — Project Overview Design

**Status:** In-progress (Phase 2 of 8 complete; Phase 2 verification pending commit)
**Milestone:** v1.0 — Polish → Supply chain capabilities → Release
**Requirements mapped:** 58/58
**Phases defined:** 8

---

## Vision

omni is a cross-platform, Go-native shell utility replacement providing deterministic, testable implementations of 160+ Unix commands for use in Taskfile, CI/CD, and enterprise environments. It ships as a single static binary plus a growing set of reusable `pkg/` libraries. Primary user: the author's CI/CD pipelines. Broader open-source adoption is a welcome bonus, not the driver.

**Core value:** One static binary, zero runtime deps, every command behaves identically on Linux, macOS, and Windows.

---

## Requirements Summary

58 v1 requirements across six tracks:

| Track | Requirements | Phases |
|-------|-------------|--------|
| Polish — cmderr migration | POLISH-01 to POLISH-03 | 1 |
| Polish — test coverage & golden harness | POLISH-04 to POLISH-14 | 2 |
| Polish — tech debt, docs, API audit | POLISH-15 to POLISH-17 | 3 |
| Supply chain — signing | SIGN-01 to SIGN-09 | 4 |
| Supply chain — SBOM | SBOM-01 to SBOM-10 | 5 |
| Supply chain — scanning | SCAN-01 to SCAN-09 | 6 |
| Supply chain — attestation | ATTEST-01 to ATTEST-09 | 7 |
| Release | REL-01 to REL-08 | 8 |

v2 requirements are deferred. Out-of-scope items (UI, OCI registry, language runtimes) are explicitly rejected with reasoning in REQUIREMENTS.md.

---

## Architecture Overview

**Layout:** Hexagonal/Clean — `cmd/` (Cobra thin wrappers) → `internal/cli/` (I/O, flags, stdin) → `pkg/` (pure-Go algorithms, zero Cobra, zero I/O).

**Key principles:**
- No `exec` — all capabilities must be pure Go, no shelling out to system tools.
- `pkg/*` exports raw errors; `internal/cli/*` wraps them in `cmderr` sentinels.
- `cmderr` is the v1.0 exit-code contract — stable from v1.0 forward, breaking-change protocol applies.
- Two golden-master test engines coexist (`testing/golden/` and `tools/golden/`) sharing `golden_tests.yaml`; consolidation is post-v1.0.
- Cross-platform parity: Linux, macOS, Windows. Platform-specific files use build tags (`_windows.go`, `_linux.go`).
- Strict supply-chain DAG: sign (4) → sbom (5) → scan (6) → attest (7) — no parallel supply-chain work.

**Technology stack:** Go 1.25, Cobra v1.10.2, Bubbletea v1.3.10 (TUI pagers), Lipgloss, GoReleaser v2, Task v3.

---

## Phase Index

| # | Slug | Spec | Status |
|---|------|------|--------|
| 1 | cmderr-migration-completion | [01-cmderr-migration-completion-design.md](./2026-04-12-01-cmderr-migration-completion-design.md) | Completed |
| 2 | test-coverage-deterministic-golden-harness | [02-test-coverage-golden-harness-design.md](./2026-04-12-02-test-coverage-deterministic-golden-harness-design.md) | Completed (verification untracked) |
| 3 | concerns-burndown-windows-parity-docs | [03-concerns-burndown-design.md](./2026-05-16-03-concerns-burndown-design.md) | Planned |
| 4 | pkg-sign-signing-primitive | [04-pkg-sign-design.md](./2026-05-16-04-pkg-sign-design.md) | Planned |
| 5 | pkg-sbom-generation | [05-pkg-sbom-design.md](./2026-05-16-05-pkg-sbom-design.md) | Planned |
| 6 | pkg-scan-vulnerability-scanning | [06-pkg-scan-design.md](./2026-05-16-06-pkg-scan-design.md) | Planned |
| 7 | pkg-attest-slsa-attestation | [07-pkg-attest-design.md](./2026-05-16-07-pkg-attest-design.md) | Planned |
| 8 | v1-release-cut | [08-v1-release-cut-design.md](./2026-05-16-08-v1-release-cut-design.md) | Planned |

---

## Key Architectural Decisions

| Decision | Rationale |
|----------|-----------|
| Polish track (1–3) must complete before supply-chain (4+) | Pitfall 13 — parallel work creates inconsistent exit-code semantics on security commands |
| Golden-master timestamp normalization in Phase 2, not later | Pitfall 14 — supply-chain goldens depend on it |
| Strict DAG: sign → sbom → scan → attest | Cryptographic primitive reuse; serialized-document boundary between sbom and scan |
| ADR gates at Phase 4, 5, 7 entries before any code | Scope-lock decisions (key handling, cosign-compat, round-trip matrix, honest SLSA level) |
| Minisign-only default; Sigstore verification behind build tag | Keeps default binary lean; Fulcio/Rekor/OCI rejected as v1.0 scope |
| `pkg/scan/` imports `pkg/sbom/format.Document` only | Non-negotiable architectural boundary — enables third-party SBOM input |
| Honest SLSA level (likely L2) pinned by ADR before attest ships | Pitfall 5 — no SLSA overclaim |

---

## Constraints

- **Timeline:** 3–4 months to v1.0.
- **No exec:** All new capabilities must be pure Go — non-negotiable.
- **Cross-platform:** Linux, macOS, Windows — all must pass CI.
- **Breaking changes:** 30-day deprecation window, log warnings on deprecated paths, cleanup commit separate from feature commits.
- **Security:** No committed secrets, parameterized queries, bcrypt ≥ 10, distroless base for containers.
- **Audience:** Author + CI/CD pipelines first. General open-source fitness is not a v1.0 input.
- **License:** BSD 3-Clause (already in place).
