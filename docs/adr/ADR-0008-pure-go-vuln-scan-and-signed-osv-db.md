# ADR-0008: Pure-Go vulnerability scanning, signed OSV DB & deferred reachability

**Status:** Accepted
**Date:** 2026-06-03
**Decision:** `omni scan` is a pure-Go, no-exec, offline-first vulnerability matcher. It matches an SBOM (`pkg/sbom/format.Document`) against a `pkg/sign`-signed `osv-db.zip` (flat upstream OSV-JSON + an omni-synthesized `manifest.json`, both covered by one detached `.minisig`) using `golang.org/x/mod/semver` interval-walk matching, and gates CI via `--fail-on <severity>` → `cmderr.ErrConflict`. **Reachability source scanning (`omni scan source`) is DEFERRED from v1.0** and returns `cmderr.ErrUnsupported`. Severity uses a hand-rolled **CVSS v3.1** base-score band; v4.0 is deferred. `pkg/scan` imports `pkg/sbom/format.Document` only.

## Context

Phase 06 ships `omni scan`. The Phase-05 spike (`docs/superpowers/research/phase-06/RESEARCH.md`) validated the default scanner design but **contradicted the plan's reachability approach** on two of omni's non-negotiable invariants:

1. **No-exec.** Reachability-aware Go-source scanning means `golang.org/x/vuln`. Its `scan.Command` runs in-process, but the analysis loads packages via `golang.org/x/tools/go/packages`, whose default driver **execs `go list`** (the only alternative, `GOPACKAGESDRIVER`, is also a subprocess). omni's "no external processes" rule is foundational; introducing a `go list` exec — even opt-in — is a new footgun pre-v1.0.
2. **Lean `go.mod` (ADR-0007).** A `//go:build omni_govulncheck` tag gates *linking*, not the *module requirement graph*. MVS + `go mod tidy` add `golang.org/x/vuln` and transitively `golang.org/x/tools v0.44.0` (+ `x/mod`, `x/sync`, `x/telemetry`) to the **main** `go.mod`/`go.sum` regardless of the tag — the exact Phase-04 sigstore mistake ADR-0007 codifies and `contrib/sigstore-verify` was built to prevent.

The default SBOM matcher has neither problem: it is pure stdlib + `golang.org/x/mod/semver` (already in `go.sum`), deterministic, and golden-testable.

## Analysis

| Decision | Choice | Rationale | Rejected alternative |
|----------|--------|-----------|----------------------|
| **Reachability (`scan source`)** | **Deferred from v1.0** — returns `cmderr.ErrUnsupported` (exit 6) with a backlog pointer. When it ships it will be a **self-contained `contrib/govulncheck-scan` module** (mirroring `contrib/sigstore-verify`), under a *sanctioned* `go list` exec exception, never the main module. | Satisfies BOTH non-negotiable invariants for v1.0 at zero dependency cost; the SBOM matcher still delivers the core value. | (a) `golang.org/x/vuln` behind `//go:build omni_govulncheck` in the **main** module — rejected: violates no-exec AND pollutes the default `go.mod` via MVS (build tags don't confine the require graph). |
| **DB distribution** | A single `pkg/sign`-signed `osv-db.zip`: upstream OSV-JSON entries (`entries/<ID>.json`, **byte-passthrough**, never re-marshalled) + an omni-synthesized `manifest.json` placed **inside** the zip so one detached `.minisig` covers both. Reject bbolt. | Diffable, language-agnostic, `archive/zip`-loadable with no new deps; byte-passthrough preserves forward-compatible OSV fields; manifest-in-zip lets a single signature anchor freshness + integrity. Matches osv.dev's `Go/all.zip` flat-JSON convention. | bbolt / a binary DB engine — rejected: opaque, non-diffable, adds a dependency, harder to sign-and-verify deterministically. |
| **Trust & freshness** | `pkg/sign.Verify` over the zip bytes on load; tampered/unsigned → `cmderr.ErrConflict` (fail-closed). `--max-db-age` reads the **signed `manifest.generated`** (omni's build/fetch time), NOT upstream's per-entry `modified`. Stale DB fails LOUDLY (`ErrConflict`), never silently degrades. | Reuses the Phase-04 primitive; the signed manifest is the only tamper-evident, offline freshness anchor. | Comparing upstream `modified` to wall-clock — rejected: `modified` is not a build/fetch timestamp and the OSV spec warns against it. |
| **Matching** | Pure-Go over `format.Document` purls. Normalize by prepending `v` to BOTH OSV SEMVER event versions and SBOM module versions before `semver.Compare`. `SEMVER` ranges: interval walk (`introduced` opens — empty/`"0"` = genesis; `fixed` closes with `>=` *unaffected at fixed*; `last_affected` closes with `>` *still affected at it*). OR'd with exact `versions[]` membership. `ECOSYSTEM`/`GIT` ranges → exact `versions[]` membership only. | The official OSV event semantics; pure, deterministic, golden-pinnable; mirrors osv-scanner's matcher minus its network/exec. | Interpreting arbitrary `ECOSYSTEM` ordering — rejected: ecosystem versions are uninterpreted strings. |
| **Severity** | Hand-rolled **CVSS v3.1** base-score band parser (closed-form; mind the v3.1 Roundup and PR-vs-Scope). Prefer a producer-supplied numeric score when present. For CVSS v4.0 records: use the numeric score if present, else `unknown`. No external CVSS lib. | v3.1 base score IS a tiny closed-form formula; avoids MVS bloat. `unknown` is always reported, treated as below `low` for gating. | Hand-rolling CVSS v4.0 — rejected for v1.0: v4.0 has no closed-form equation (needs FIRST's ~270-entry MacroVector table + EQ interpolation). |
| **Boundary** | `pkg/scan` imports `pkg/sbom/format.Document` ONLY (never `pkg/sbom/model`). | Architectural invariant; `format.Document` is the stable v1.0 boundary (ADR-0007). | Importing `model` — rejected: couples to a churn-prone internal type. |

## Consequences

- **All invariants hold for v1.0.** The default `omni scan` adds zero new modules to `go.mod` beyond `golang.org/x/mod` (already present); no exec; deterministic, golden-testable output.
- **Reachability is a known, documented gap**, tracked in `docs/BACKLOG.md`. `omni scan source <dir>` returns `ErrUnsupported` (exit 6) cleanly — never a silent no-op. The future home is a `contrib/govulncheck-scan` module under a sanctioned `go list` exec exception, decided in a follow-up ADR.
- **Fail-closed everywhere.** Unsigned/tampered DB, stale DB past `--max-db-age`, and unknown formats all return classified errors, never a silent pass.
- **Severity is honest.** v3.1 scored; v4.0 numeric-or-`unknown`; `unknown` always reported. Reachability's "eliminates false positives" claim is downgraded to "reduces" (it does not ship in v1.0 regardless).
- **Forward-compatible DB.** Byte-passthrough + tolerating unknown OSV fields (validating only `id`, `modified`, `schema_version`) means future OSV minor versions (1.7.6/1.8.0) do not break `scan`.
