# omni v1.0 Requirements

**Milestone:** v1.0 — Polish → Supply chain capabilities → Release
**Timeline:** 3–4 months
**Audience:** me + my CI/CD pipelines (broader adoption = bonus, not a driver)
**Source:** Synthesized from `.planning/PROJECT.md` and `.planning/research/SUMMARY.md` with user scoping decisions.

---

## v1 Requirements

### Polish — cmderr migration

- [x] **POLISH-01**: Every command in `cmd/` + `internal/cli/` returns classified `cmderr` sentinels (ErrNotFound / ErrInvalidInput / ErrPermission / ErrIO / ErrTimeout / ErrUnsupported / ErrConflict) with correct exit codes via `cmderr.ExitCodeFor()` — finish the remaining ~76 commands beyond the 84 already migrated.
- [x] **POLISH-02**: Every migrated command has at least one golden-master test case exercising an error path to lock exit-code behavior.
- [ ] **POLISH-03**: `internal/cli/cmderr` has a coverage threshold of ≥90% enforced in CI.

### Polish — test coverage

- [ ] **POLISH-04**: Omni-owned `pkg/*` packages average ≥75% test coverage (excluding vendored `buf` subtrees, which are tracked separately in `CONCERNS.md`).
- [ ] **POLISH-05**: Omni-owned `internal/cli/*` packages average ≥60% test coverage.
- [ ] **POLISH-06**: Coverage reports are generated per-PR in CI and surface uncovered files clearly.
- [ ] **POLISH-07**: Every `pkg/` package that currently has no `_test.go` files gets at least one baseline test establishing its public API contract.

### Polish — tech debt & CONCERNS.md burn-down

- [ ] **POLISH-08**: Every Critical and High item in `.planning/codebase/CONCERNS.md` is either resolved or explicitly deferred with written rationale in `docs/BACKLOG.md`.
- [ ] **POLISH-09**: Windows parity gaps in `ps`, `df`, `free`, `kill`, `uptime`, `id` are resolved or documented as known limitations with explicit `cmderr.ErrUnsupported` returns on Windows.
- [ ] **POLISH-10**: `docs/ISSUES.md` is triaged to empty (every bug either fixed, converted to a backlog item, or closed as "not a bug" with rationale).

### Polish — docs & golden master completeness

- [x] **POLISH-11**: Every command has a top-level usage docstring surfaced by `omni <cmd> --help` that includes at least one concrete example.
- [ ] **POLISH-12**: Every command is registered in both golden-master registries (`testing/golden/golden_tests.yaml` and `tools/golden/golden_tests.yaml`) with at least one happy-path snapshot.
- [ ] **POLISH-13**: The golden-master harness supports timestamp and random-ID normalization hooks, so supply-chain commands (sbom/sign/attest) can have deterministic goldens in later phases.
- [ ] **POLISH-14**: `omni cmdtree` output is regenerated and committed at the end of the polish track so `aicontext` documentation reflects current command coverage.

### Polish — pkg/ API audit

- [ ] **POLISH-15**: Every currently-exported symbol in `pkg/*` is triaged into `stable` / `experimental` / `internal` buckets — experimental symbols are annotated with `// Experimental:` godoc comments, internal symbols are moved to `internal/` or unexported.
- [ ] **POLISH-16**: `pkg/video/` public API is audited and either stabilized or moved behind a smaller, frozen surface (candidate for `/v2` module or `internal/video` with a thin public facade).
- [ ] **POLISH-17**: `CLAUDE.md` Breaking Changes protocol is adopted for every `pkg/*` public-API change going forward, with deprecation dates tracked in `docs/BACKLOG.md`.

---

### Supply Chain — `pkg/sign/` (signing primitive)

- [ ] **SIGN-01**: `omni sign keygen` generates a passphrase-protected Ed25519 keypair compatible with the minisign file format, writing `<name>.key` and `<name>.pub` to disk with `0600` / `0644` permissions.
- [ ] **SIGN-02**: `omni sign <file>` produces a detached `.sig` signature for any file using `aead.dev/minisign`, supporting both trusted and untrusted comments.
- [ ] **SIGN-03**: `omni verify <file> --pubkey <key>` verifies a detached signature, returning `cmderr.ErrConflict` on any mismatch and failing closed on every error mode (missing sig, wrong key, tampered payload, bad algorithm).
- [ ] **SIGN-04**: Private keys are accepted only via file path or env-var-pointing-to-file, never as flag values (prevents `/proc/*/cmdline` leakage).
- [ ] **SIGN-05**: A typed `secret.Key` wrapper with a redacting `String()` / `GoString()` / `LogValue()` prevents keys from ever appearing in slog output, error messages, or panics.
- [ ] **SIGN-06**: `pkg/sign/` is usable as a standalone Go library with stable public types — `Signer`, `Verifier`, `KeyPair`, `Signature`.
- [ ] **SIGN-07**: Sigstore bundle-format **verification** is available behind the `omni_sigstore` build tag — `omni verify --bundle <path>` works when the binary is built with the tag, returns `cmderr.ErrUnsupported` when not.
- [ ] **SIGN-08**: Negative-test matrix covers every verify failure mode: missing sig, wrong key, wrong algorithm, tampered payload, expired key, unknown key ID. Every case is a golden-master test.
- [ ] **SIGN-09**: `omni sign` / `omni verify` commands register with `internal/cli/command/` Registry and work inside `omni pipe` / `omni pipeline` chains.

### Supply Chain — `pkg/sbom/` (SBOM generation)

- [ ] **SBOM-01**: `omni sbom <target> --format spdx` generates SPDX 2.3 JSON for a Go module at a directory path (parses `go.mod` + `go.sum`).
- [ ] **SBOM-02**: `omni sbom <target> --format cyclonedx` generates CycloneDX 1.5 JSON for the same target.
- [ ] **SBOM-03**: `omni sbom <binary>` generates an SBOM for a built Go binary by reading `debug/buildinfo`, producing "what actually shipped" rather than "what was declared."
- [ ] **SBOM-04**: Every component in the SBOM has a correctly normalized purl using Go module path (not import path), with pseudo-version handling matching `golang.org/x/mod/semver`.
- [ ] **SBOM-05**: SBOM output is byte-deterministic — same input produces the same bytes — so goldens work and reproducible builds extend to SBOMs.
- [ ] **SBOM-06**: Catalogers are registered via `init()` in `pkg/sbom/cataloger/all/` mirroring the `pkg/video/extractor/all/` pattern. v1.0 ships `cataloger/gomod/` and `cataloger/gobinary/`; other ecosystems are stubbed for future additions.
- [ ] **SBOM-07**: A round-trip test matrix validates that `omni sbom` output parses cleanly through `syft convert` and `govulncheck` in CI (not because we use those tools, but because they're the compatibility oracle).
- [ ] **SBOM-08**: `pkg/sbom/` exposes a stable public type `format.Document` that `pkg/scan/` consumes — this is the *only* boundary; `pkg/scan/` must never import `pkg/sbom/model` structs directly.
- [ ] **SBOM-09**: Container-image SBOM generation is explicitly **NOT** in v1.0 (see Out of Scope).
- [ ] **SBOM-10**: `omni sbom` command registers with `internal/cli/command/` Registry and works inside pipelines.

### Supply Chain — `pkg/scan/` (vulnerability scanning)

- [ ] **SCAN-01**: `omni scan <sbom>` consumes an SPDX or CycloneDX document, matches components against an OSV-format vulnerability database, and reports findings in JSON and text formats.
- [ ] **SCAN-02**: `omni scan --fail-on <severity>` gates exit code on severity threshold — returns `cmderr.ErrConflict` if any finding meets or exceeds the threshold, enabling CI gating.
- [ ] **SCAN-03**: `omni scan source <path>` scans a Go source directory directly using `golang.org/x/vuln` (govulncheck's public scan API), enabling **reachability analysis** — reports only vulnerabilities where the vulnerable symbol is actually called, dramatically reducing false positives vs SBOM-only matching.
- [ ] **SCAN-04**: `omni scan db update` downloads the latest OSV vulnerability database to a local cache.
- [ ] **SCAN-05**: Vulnerability DB archives are signed with `pkg/sign/` and verified on load — tampered DBs fail closed.
- [ ] **SCAN-06**: `omni scan` supports an offline mode using the cached DB, with `--max-db-age` gating how stale the DB can be before scan fails loudly (no silent degradation).
- [ ] **SCAN-07**: An OSV regression suite of known-CVE fixtures runs as golden-master tests — if matching logic drifts, goldens fail immediately.
- [ ] **SCAN-08**: `pkg/scan/` consumes SBOMs via the serialized `format.Document` type only — zero imports from `pkg/sbom/model`. This is a hard architectural boundary.
- [ ] **SCAN-09**: `omni scan` command registers with `internal/cli/command/` Registry and works inside pipelines.

### Supply Chain — `pkg/attest/` (SLSA attestation)

- [ ] **ATTEST-01**: `omni attest --predicate-type slsa-provenance --predicate <file> --artifact <path>` produces an in-toto Statement with a SLSA v1.0 provenance predicate for the named artifact.
- [ ] **ATTEST-02**: Attestations are wrapped in a DSSE envelope using the Pre-Authentication Encoding (PAE) format `DSSEv1 <len(type)> <type> <len(payload)> <payload>`, validated against the in-toto reference test vectors as unit tests.
- [ ] **ATTEST-03**: PAE is implemented in exactly one place (`pkg/attest/dsse/pae.go`) and is never inlined.
- [ ] **ATTEST-04**: `omni attest verify <envelope>` verifies a DSSE envelope against a pubkey, failing closed on any error.
- [ ] **ATTEST-05**: A `--from-env` mode reads GitHub Actions environment variables (`GITHUB_RUN_ID`, `GITHUB_WORKFLOW`, `GITHUB_SHA`, etc.) to populate provenance fields automatically.
- [ ] **ATTEST-06**: SLSA predicates are validated against the official SLSA v1.0 JSON schema in CI before every release.
- [ ] **ATTEST-07**: An ADR exists at `docs/adr/` pinning the honest SLSA level omni's *own* release process achieves (most likely L2 via GitHub Actions + OIDC, not L3). No provenance output may claim a higher level than the ADR.
- [ ] **ATTEST-08**: `pkg/attest/` depends on `pkg/sign/` and consumes `pkg/sbom/format.Document` for subject digests — depends on both, does not depend on `pkg/scan/`.
- [ ] **ATTEST-09**: `omni attest` command registers with `internal/cli/command/` Registry and works inside pipelines.

---

### Release — v1.0 cut

- [ ] **REL-01**: A `pkg/` API freeze snapshot is tagged at the start of the release phase — any breaking change from that point triggers the 30-day deprecation protocol from `CLAUDE.md`.
- [ ] **REL-02**: Signed binaries are built for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`, `windows/arm64` and published as a GitHub release.
- [ ] **REL-03**: Binaries are built reproducibly with `-trimpath` + `-buildvcs=true` and a CI job does a dual-build comparison to catch reproducibility drift.
- [ ] **REL-04**: omni signs its own release binaries using `omni sign` (dogfooding — eat our own output before announcing).
- [ ] **REL-05**: omni generates its own release SBOM using `omni sbom` and publishes it as a release asset.
- [ ] **REL-06**: omni generates a SLSA v1.0 provenance attestation for its own release using `omni attest` with the honest level from ATTEST-07.
- [ ] **REL-07**: Release notes are generated from conventional commits and include a mandatory "What's NOT protected against" section linking `.planning/research/PITFALLS.md` — no security theater.
- [ ] **REL-08**: The v1.0 announcement explicitly states the target audience ("me + my CI/CD pipelines") and does not overclaim general-purpose fitness.

---

## v2 Requirements (deferred, not building now)

- Full cosign CLI parity (signing with Fulcio OIDC + Rekor transparency log)
- Container-image SBOM generation (parsing Go binaries inside OCI layers)
- OCI registry push/pull for any artifact type
- Polyglot SBOM catalogers (npm, pypi, crates, maven) — v1.0 is Go-only
- OSV-scanner v2 polyglot mode integration (currently gated behind `omni_osv` build tag)
- Policy-as-code engine for scan gating (OPA/Rego-style)
- Native Linux distro packaging (apt / deb / rpm)
- Homebrew, scoop, winget distribution
- Docker / distroless image publishing
- Daemon / HTTP-API mode
- Interactive shell / REPL / TUI
- Golden-master system consolidation (merge `testing/golden/` + `tools/golden/`)
- Plugin / dynamic-load system

## Out of Scope (explicit, with reasoning)

- **Interactive shell / REPL / TUI shell** — omni is a one-shot utility by design; long-running interactive mode is a different product entirely.
- **Plugin or dynamic-load system** — adds surface area and security risk that doesn't serve the CI/CD use case; everything stays statically linked.
- **Daemon / HTTP-API mode** — same reason; omni is one-shot.
- **Broadening public adoption as a design driver** — welcome to adopt, but not a prioritization input; prevents scope creep.
- **Breaking `pkg/` refactors without 30-day deprecation** — non-negotiable per `CLAUDE.md`.
- **Rebrand / rename / module path change** — stays as `github.com/inovacc/omni` through v1.0.
- **Full SLSA L3 claim for omni's own release** — honest self-assessment over marketing; see ATTEST-07.
- **syft / cosign / grype as library dependencies** — research rejected all three (dep bloat + cosign upstream migration).
- **Rekor transparency log, Fulcio keyless flow, OCI registry clients in the default binary** — too large a dep surface, CI use case doesn't require it; Sigstore bundle *verification* behind a build tag is the interop compromise.
- **Container image SBOM / scan** — no pure-Go OCI registry client is lean enough in 2026; pushed to v2.
- **Polyglot SBOM (npm, pypi, crates, maven)** — Go-first audience; polyglot is v2.
- **Writing new CGO code** — non-negotiable per `CLAUDE.md` design principles.
- **Spawning external processes from any command** — non-negotiable per `CLAUDE.md` design principles.

---

## Traceability

Each requirement maps to exactly one phase. Coverage: **58/58 (100%)**.

| REQ-ID | Phase | Status |
|--------|-------|--------|
| POLISH-01 | Phase 1 | Complete |
| POLISH-02 | Phase 1 | Complete |
| POLISH-03 | Phase 1 | Pending |
| POLISH-04 | Phase 2 | Pending |
| POLISH-05 | Phase 2 | Pending |
| POLISH-06 | Phase 2 | Pending |
| POLISH-07 | Phase 2 | Pending |
| POLISH-11 | Phase 2 | Complete |
| POLISH-12 | Phase 2 | Pending |
| POLISH-13 | Phase 2 | Pending |
| POLISH-14 | Phase 2 | Pending |
| POLISH-08 | Phase 3 | Pending |
| POLISH-09 | Phase 3 | Pending |
| POLISH-10 | Phase 3 | Pending |
| POLISH-15 | Phase 3 | Pending |
| POLISH-16 | Phase 3 | Pending |
| POLISH-17 | Phase 3 | Pending |
| SIGN-01 | Phase 4 | Pending |
| SIGN-02 | Phase 4 | Pending |
| SIGN-03 | Phase 4 | Pending |
| SIGN-04 | Phase 4 | Pending |
| SIGN-05 | Phase 4 | Pending |
| SIGN-06 | Phase 4 | Pending |
| SIGN-07 | Phase 4 | Pending |
| SIGN-08 | Phase 4 | Pending |
| SIGN-09 | Phase 4 | Pending |
| SBOM-01 | Phase 5 | Pending |
| SBOM-02 | Phase 5 | Pending |
| SBOM-03 | Phase 5 | Pending |
| SBOM-04 | Phase 5 | Pending |
| SBOM-05 | Phase 5 | Pending |
| SBOM-06 | Phase 5 | Pending |
| SBOM-07 | Phase 5 | Pending |
| SBOM-08 | Phase 5 | Pending |
| SBOM-09 | Phase 5 | Pending |
| SBOM-10 | Phase 5 | Pending |
| SCAN-01 | Phase 6 | Pending |
| SCAN-02 | Phase 6 | Pending |
| SCAN-03 | Phase 6 | Pending |
| SCAN-04 | Phase 6 | Pending |
| SCAN-05 | Phase 6 | Pending |
| SCAN-06 | Phase 6 | Pending |
| SCAN-07 | Phase 6 | Pending |
| SCAN-08 | Phase 6 | Pending |
| SCAN-09 | Phase 6 | Pending |
| ATTEST-01 | Phase 7 | Pending |
| ATTEST-02 | Phase 7 | Pending |
| ATTEST-03 | Phase 7 | Pending |
| ATTEST-04 | Phase 7 | Pending |
| ATTEST-05 | Phase 7 | Pending |
| ATTEST-06 | Phase 7 | Pending |
| ATTEST-07 | Phase 7 | Pending |
| ATTEST-08 | Phase 7 | Pending |
| ATTEST-09 | Phase 7 | Pending |
| REL-01 | Phase 8 | Pending |
| REL-02 | Phase 8 | Pending |
| REL-03 | Phase 8 | Pending |
| REL-04 | Phase 8 | Pending |
| REL-05 | Phase 8 | Pending |
| REL-06 | Phase 8 | Pending |
| REL-07 | Phase 8 | Pending |
| REL-08 | Phase 8 | Pending |
