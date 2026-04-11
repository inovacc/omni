# omni

## What This Is

omni is a cross-platform, Go-native shell utility replacement providing deterministic, testable implementations of 160+ Unix commands for use in Taskfile, CI/CD, and enterprise environments. Today it ships as a single binary plus a growing set of reusable `pkg/` libraries. The primary user is me and my CI/CD pipelines — broader open-source adoption is a welcome bonus, not the driver.

## Core Value

**One static binary replaces every shell utility a Go-based CI/CD pipeline needs — deterministically, on every OS, with no external processes spawned.** If everything else fails, this must remain true.

## Requirements

### Validated

<!-- Shipped and confirmed valuable. Inferred from codebase map. -->

- ✓ 160+ cross-platform command implementations (cat, grep, rg, find, ls, tar, jq, yq, etc.) — existing
- ✓ Hexagonal architecture: `cmd/` (Cobra) → `internal/cli/` (I/O glue) → `pkg/` (reusable libraries) — existing
- ✓ Unified `Command` interface + `Registry` in `internal/cli/command/` — existing
- ✓ `cmderr` sentinel error model with typed exit codes (NotFound, Conflict, InvalidInput, Permission, IO, Timeout, Unsupported) — existing, partially adopted (~84/160 commands)
- ✓ Pure-Go stdlib-first implementation — no external process spawns, no CGO where avoidable — existing
- ✓ Cross-platform build (Linux, macOS, Windows) with build-tag platform splits — existing
- ✓ Cloud/DevOps integrations: kubectl (`k`), terraform (`tf`), aws, git hacks — existing
- ✓ Pure-Go video download engine (`pkg/video/`) with YouTube + HLS + generic extractors — existing
- ✓ Streaming text pipeline engine (`pkg/pipeline/`) with 20 built-in stages — existing
- ✓ Ripgrep-compatible search with gitignore support (`pkg/search/rg/`) — existing
- ✓ Tree scanner with parallel walking, NDJSON streaming, and JSON snapshot comparison (`pkg/twig/`) — existing
- ✓ Golden master test harness (lightweight + full-featured dual systems) — existing
- ✓ CI via GitHub Actions: golangci-lint, gofmt, govulncheck, race-enabled tests — existing
- ✓ MCP server scaffolding templates aligned with go-sdk v1.4.0 — existing

### Active

<!-- v1.0 milestone. Polish → Supply chain capability → Release, in that order. -->

**Polish track (Phase 1–3):**

- [ ] Finish `cmderr` migration for remaining ~76 commands — every command returns classified sentinels with correct exit codes
- [ ] Raise test coverage from ~25.8% → 60–80% across omni-owned `pkg/` and `internal/cli/` packages
- [ ] Burn down `.planning/codebase/CONCERNS.md` tech-debt backlog (security, fragile areas, platform parity gaps)
- [ ] Documentation completeness — every command documented with usage examples + golden master test cases
- [ ] Every command covered by at least one golden master snapshot (both `testing/golden/` and `tools/golden/` registries)

**Supply chain capability track (Phase 4–5):**

- [ ] `omni sbom` — SPDX/CycloneDX SBOM generation for Go modules and containers (syft-style, pure Go)
- [ ] `omni sign` / `omni verify` — artifact signing and signature verification (cosign-compatible where feasible, or minisign-style for a fully pure-Go path)
- [ ] `omni scan` — vulnerability scanning against OSV/GHSA databases (grype-style)
- [ ] `omni attest` — SLSA provenance generation for CI builds
- [ ] All supply-chain commands usable as `pkg/` libraries from external Go projects

**Release track (Phase 6):**

- [ ] Freeze `pkg/` public API under semver with a 30-day deprecation window (per CLAUDE.md breaking-change protocol)
- [ ] Signed GitHub release binaries for linux/darwin/windows × amd64/arm64
- [ ] Reproducible builds with `-trimpath` + `-buildvcs` metadata
- [ ] Release notes generated from conventional commits
- [ ] v1.0 announcement with honest scope statement

### Out of Scope

<!-- Explicit boundaries for the v1.0 milestone. -->

- **Interactive shell / REPL / TUI shell** — omni is a utility binary, not a shell. Stays out until post-1.0.
- **Homebrew / scoop / winget / apt / deb / rpm distribution** — signed GitHub release binaries are the v1.0 channel; package-manager distribution is a post-1.0 concern.
- **Docker / distroless image** — nice-to-have but not on the v1.0 critical path. Deferred.
- **Plugin / dynamic-load system** — adds surface area and security risk that doesn't serve the CI/CD use case.
- **Daemon / HTTP-API mode** — omni is one-shot by design; long-running mode is a different product.
- **Rebrand / rename / package split** — omni stays as `github.com/inovacc/omni` through 1.0.
- **Broadening public adoption as a goal** — users beyond me + my CI/CD pipelines are welcome but are not a design driver or prioritization input.
- **Full `cosign` OCI/registry compatibility** — may be too large for v1.0; a minisign-style pure-Go signing path is acceptable as the minimum. Revisit during phase planning.
- **Cross-breaking `pkg/` refactors after the API freeze** — anything breaking must follow the 30-day deprecation protocol, no exceptions.

## Context

**Codebase state:** Mature Go CLI, ~160 commands, hexagonal layout, broad `pkg/` surface. Some vendored `buf` packages skew total coverage numbers; omni-owned packages average ~78%. See `.planning/codebase/` for the full map.

**Error model transition in progress:** `cmderr` was introduced recently and adopted in batches (84 commands so far across 7 batches). Migration is well-understood and mechanical — the remaining work is mostly patience and test coverage.

**Two golden-master systems coexist:** A lightweight engine in `testing/golden/` and a full-featured system in `tools/golden/`. Both share the same `golden_tests.yaml` registry. They are not being consolidated during v1.0 — that's post-1.0 cleanup.

**Platform parity is uneven:** Some commands (df, free, kill, ps, etc.) have separate `_unix.go` / `_windows.go` implementations. Parity gaps are a known `CONCERNS.md` item.

**Vendored dependencies:** The `buf` tooling pulls in vendored protobuf compiler packages, which inflate the total line count and skew coverage. These are not omni-owned code.

**Reference documents (authoritative):**
- Project instructions: `CLAUDE.md` (root)
- Codebase map: `.planning/codebase/STACK.md`, `ARCHITECTURE.md`, `STRUCTURE.md`, `CONVENTIONS.md`, `TESTING.md`, `INTEGRATIONS.md`, `CONCERNS.md`
- Roadmap (pre-GSD): `docs/ROADMAP.md`, `docs/BACKLOG.md`, `docs/ISSUES.md`

## Constraints

- **Tech stack**: Go (stdlib-first, Cobra CLI, pure-Go deps only) — deterministic, portable, no CGO where avoidable. No exec of external processes is a foundational rule.
- **Timeline**: 3–4 months to v1.0 — polish → supply chain → release, in that order. No aggressive parallelism across tracks.
- **Cross-platform**: Linux, macOS, Windows must all work. Platform-specific code uses build tags, never silent runtime branches.
- **Breaking changes**: Follow CLAUDE.md breaking-change protocol — 30-day deprecation window, log warnings on deprecated paths, cleanup commit separate from feature commits.
- **Security**: No committed secrets (`~/.claude/scripts/check-leaks.sh`), parameterized queries, bcrypt ≥ 10, distroless base for any containers. Pre-v1.0 cannot introduce new security footguns.
- **Licensing**: BSD 3-Clause, already in place.
- **Audience**: Design decisions serve me + CI/CD pipelines first. "Would a general open-source user want X?" is never a v1.0 prioritization input.
- **No new external processes**: The "no exec" design principle is non-negotiable — new capabilities must be pure Go.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Polish before new features before release | Shipping v1.0 on an unfinished `cmderr` migration and ~25% coverage would be dishonest — foundation first | — Pending |
| Supply chain & signing as the only new capability track for v1.0 | CI/CD relevance + pure-Go fit + 2026 industry importance; other capabilities (secrets, artifact storage, dev server) are deferred | — Pending |
| Semver with 30-day deprecation window for `pkg/` API | Matches CLAUDE.md breaking-change protocol; strict "no breaks ever" would block needed refactors | — Pending |
| GitHub releases + signed binaries only at v1.0 | Minimal viable distribution path; package managers add release complexity without serving the primary CI/CD use case | — Pending |
| Keep both golden-master systems for v1.0 | Consolidating them is post-1.0 cleanup; forcing it now would churn every command's test case | — Pending |
| Audience = me + my CI/CD pipelines | Broader adoption as a design driver creates scope pressure that would blow the 3–4 month timeline | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-04-11 after initialization*
