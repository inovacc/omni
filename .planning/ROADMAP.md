# omni v1.0 Roadmap

**Milestone:** v1.0 — Polish → Supply chain capabilities → Release
**Granularity:** standard (8 phases)
**Timeline:** 3–4 months
**Coverage:** 58/58 v1 requirements mapped
**Ordering:** Strict DAG — Polish (1→2→3) → sign (4) → sbom (5) → scan (6) → attest (7) → release (8). No parallelization across supply-chain phases.

---

## Phases

- [ ] **Phase 1: cmderr Migration Completion** — Finish cmderr adoption across all remaining commands so exit-code semantics are consistent before any security-critical work begins.
- [ ] **Phase 2: Test Coverage + Deterministic Golden Harness** — Raise coverage to target and land timestamp/random-ID normalization hooks that supply-chain goldens will depend on.
- [ ] **Phase 3: CONCERNS Burn-down, Windows Parity, pkg/ API Audit, Docs** — Close tech-debt backlog, resolve platform parity gaps, triage pkg/ public surface, and complete command documentation.
- [ ] **Phase 4: pkg/sign/ — Signing Primitive** — Ship `omni sign`/`verify`/`keygen` with minisign-compatible, fail-closed Ed25519 signing. (ADR gate required)
- [ ] **Phase 5: pkg/sbom/ — SBOM Generation** — Ship `omni sbom` producing deterministic SPDX 2.3 and CycloneDX 1.5 documents for Go modules and binaries. (ADR gate required)
- [ ] **Phase 6: pkg/scan/ — Vulnerability Scanning** — Ship `omni scan` with OSV-backed matching, reachability analysis, and signed DB distribution. (NEEDS DEEPER RESEARCH)
- [ ] **Phase 7: pkg/attest/ — SLSA Attestation** — Ship `omni attest` with DSSE-wrapped SLSA v1.0 provenance at the honest level. (ADR gate required)
- [ ] **Phase 8: v1.0 Release Cut** — API freeze, reproducible signed binaries for 6 targets, release dogfoods sign/sbom/attest, honest announcement.

---

## Phase Details

### Phase 1: cmderr Migration Completion
**Goal**: Every command in omni returns classified `cmderr` sentinels with correct exit codes, eliminating inconsistency before any supply-chain command is written.
**Depends on**: Nothing (first phase)
**Requirements**: POLISH-01, POLISH-02, POLISH-03
**Success Criteria** (what must be TRUE):
  1. Every command in `cmd/` + `internal/cli/` returns a classified `cmderr` sentinel — a user running any command and triggering a failure sees a predictable exit code via `cmderr.ExitCodeFor()`.
  2. Every migrated command has at least one golden-master test case exercising an error path, so exit-code regressions are caught immediately by CI.
  3. `internal/cli/cmderr` is covered at ≥90% and the threshold is enforced in CI — a PR that drops coverage fails.
  4. No command silently returns raw `os.ErrNotExist`/`os.ErrPermission` — all I/O errors are wrapped at the CLI boundary.
**Plans**: TBD
**Pitfalls addressed**: 13 (parallel cmderr + supply-chain)

### Phase 2: Test Coverage + Deterministic Golden Harness
**Goal**: Raise test coverage to v1.0 targets and land the golden-master normalization hooks (timestamps, random IDs) that supply-chain commands will require in Phases 4–7.
**Depends on**: Phase 1
**Requirements**: POLISH-04, POLISH-05, POLISH-06, POLISH-07, POLISH-11, POLISH-12, POLISH-13, POLISH-14
**Success Criteria** (what must be TRUE):
  1. Omni-owned `pkg/*` packages average ≥75% coverage and `internal/cli/*` packages average ≥60%, verified by a per-PR CI report that surfaces uncovered files.
  2. Every `pkg/` package has at least one baseline test establishing its public API contract — no package ships to v1.0 with zero tests.
  3. The golden-master harness normalizes timestamps and random identifiers via documented hooks — a developer writing a golden for SBOM output in Phase 5 can rely on deterministic diffs.
  4. Every command is registered in both golden-master registries with at least one happy-path snapshot, and every command's `--help` surfaces a usage docstring with a concrete example.
  5. `omni cmdtree` output is regenerated and committed so `aicontext` documentation reflects current command coverage.
**Plans**: TBD
**Pitfalls addressed**: 10, 14

### Phase 3: CONCERNS Burn-down, Windows Parity, pkg/ API Audit, Docs
**Goal**: Close the tech-debt backlog, resolve or explicitly document Windows parity gaps, and triage the `pkg/*` public surface before the API freeze so experimental code can still move.
**Depends on**: Phase 2
**Requirements**: POLISH-08, POLISH-09, POLISH-10, POLISH-15, POLISH-16, POLISH-17
**Success Criteria** (what must be TRUE):
  1. Every Critical and High item in `.planning/codebase/CONCERNS.md` is resolved or explicitly deferred with written rationale in `docs/BACKLOG.md` — no silent "we'll figure it out" items remain.
  2. Windows parity gaps in `ps`, `df`, `free`, `kill`, `uptime`, `id` are resolved or return `cmderr.ErrUnsupported` with a clear message — a Windows user never sees a silent runtime branch or a cryptic syscall error.
  3. `docs/ISSUES.md` is triaged to empty — every bug either fixed, backlogged, or closed with rationale.
  4. Every currently-exported `pkg/*` symbol is triaged into stable/experimental/internal buckets; experimental symbols carry `// Experimental:` godoc, internal ones are moved or unexported.
  5. `pkg/video/` public API is audited and either stabilized or moved behind a smaller frozen surface, and the CLAUDE.md breaking-change protocol is adopted for every future `pkg/*` public-API change.
**Plans**: TBD
**Pitfalls addressed**: 7, 8, 16

### Phase 4: pkg/sign/ — Signing Primitive
**Goal**: Users can generate keys, sign arbitrary files, and verify detached signatures using a pure-Go, fail-closed, minisign-compatible signing path — the cryptographic primitive every later supply-chain phase reuses.
**Depends on**: Phase 3
**ADR Gate (required before any code lands)**:
  - ADR on key handling policy — storage, rotation, dev-vs-release key separation, never-as-flag rule, `secret.Key` wrapper design.
  - ADR on cosign-compat scope — v1.0 is minisign-only by default with Sigstore bundle *verification* behind the `omni_sigstore` build tag; no Rekor, no Fulcio, no OCI.
**Requirements**: SIGN-01, SIGN-02, SIGN-03, SIGN-04, SIGN-05, SIGN-06, SIGN-07, SIGN-08, SIGN-09
**Success Criteria** (what must be TRUE):
  1. A user can run `omni sign keygen` and receive a passphrase-protected Ed25519 keypair written with `0600`/`0644` permissions, compatible with the minisign file format.
  2. A user can sign a file with `omni sign <file>` and verify it with `omni verify <file> --pubkey <key>` — verification fails closed on every failure mode (missing sig, wrong key, tampered payload, bad algorithm, expired key, unknown key ID) with `cmderr.ErrConflict`, each case locked by a golden-master negative test.
  3. Private keys are never accepted as CLI flag values (file path or env-var-pointing-to-file only) and never appear in slog output, error messages, or panic traces — a `secret.Key` wrapper with redacting `String()`/`GoString()`/`LogValue()` prevents leakage.
  4. `pkg/sign/` is usable as a standalone Go library with stable public types (`Signer`, `Verifier`, `KeyPair`, `Signature`) and `omni sign`/`omni verify` both register with `internal/cli/command/` Registry and work inside `omni pipe`/`omni pipeline`.
  5. A user building with `-tags omni_sigstore` can run `omni verify --bundle <path>` against a Sigstore bundle; without the tag the command returns `cmderr.ErrUnsupported` cleanly.
**Plans**: TBD
**Pitfalls addressed**: 2 (fail-open verify), 3 (key handling), 9 (cosign scope), 17 (unmaintained deps)

### Phase 5: pkg/sbom/ — SBOM Generation
**Goal**: Users can generate byte-deterministic SPDX 2.3 and CycloneDX 1.5 documents for a Go module directory or a built Go binary, with correct purls that downstream scanners can actually match.
**Depends on**: Phase 4
**ADR Gate (required before any code lands)**:
  - ADR on SBOM round-trip test matrix — exactly which oracles (`syft convert`, `govulncheck`) validate omni's output in CI, what "purl correctness" means, and how pseudo-versions/replace-directives are normalized.
**Requirements**: SBOM-01, SBOM-02, SBOM-03, SBOM-04, SBOM-05, SBOM-06, SBOM-07, SBOM-08, SBOM-09, SBOM-10
**Success Criteria** (what must be TRUE):
  1. A user can run `omni sbom <dir> --format spdx` and `omni sbom <dir> --format cyclonedx` on a Go module and receive valid SPDX 2.3 and CycloneDX 1.5 JSON, parsed cleanly by `syft convert` in CI.
  2. A user can run `omni sbom <binary>` on a built Go binary and receive an SBOM derived from `debug/buildinfo` that describes "what actually shipped," not "what was declared" — the Go toolchain is listed as a component.
  3. Every component has a correctly normalized purl using the Go module path (not import path), with pseudo-version and `+incompatible` handling matching `golang.org/x/mod/semver` — purl round-trip tests catch any drift.
  4. SBOM output is byte-deterministic — running the command twice on the same input produces identical bytes, enabling golden-master pinning and reproducible-build extension to SBOMs.
  5. `pkg/sbom/` ships with `cataloger/gomod/` and `cataloger/gobinary/` registered via `init()` under `cataloger/all/`, exposes the stable public type `format.Document` as the *only* boundary for `pkg/scan/`, and `omni sbom` registers with the `internal/cli/command/` Registry and works inside pipelines.
**Plans**: TBD
**Pitfalls addressed**: 1 (purl correctness), 12 (SBOM completeness), 14 (golden timestamps)

### Phase 6: pkg/scan/ — Vulnerability Scanning
**Goal**: Users can scan an SBOM or a Go source tree against a signed OSV vulnerability database and gate CI on severity thresholds, with reachability analysis for Go source to kill false positives.
**Depends on**: Phase 5
**Research status**: **NEEDS DEEPER RESEARCH** at phase-planning time per SUMMARY.md — DB distribution format (flat JSON vs bbolt), OSV-scanner v2 API surface, grype v6 schema direction, `golang.org/x/vuln` public scan API shape. Schedule `/gsd-research-phase 6` before plan generation.
**Requirements**: SCAN-01, SCAN-02, SCAN-03, SCAN-04, SCAN-05, SCAN-06, SCAN-07, SCAN-08, SCAN-09
**Success Criteria** (what must be TRUE):
  1. A user can run `omni scan <sbom>` on an SPDX or CycloneDX document and receive findings in JSON and text, and `omni scan --fail-on <severity>` returns `cmderr.ErrConflict` when any finding meets or exceeds the threshold — the CI gating path is golden-tested.
  2. A user can run `omni scan source <path>` on a Go source directory and receive reachability-aware findings via `golang.org/x/vuln` — only vulnerabilities where the vulnerable symbol is actually called are reported.
  3. A user can run `omni scan db update` to download the latest OSV database to a local cache; the archive is signed with `pkg/sign/` and verified on load — tampered DBs fail closed.
  4. Offline mode works with the cached DB, and `--max-db-age` gates staleness — a stale DB fails loudly, never degrades silently. An OSV regression suite of known-CVE fixtures runs as golden-master tests.
  5. `pkg/scan/` consumes SBOMs *only* via the serialized `format.Document` type (zero imports from `pkg/sbom/model`) — this hard boundary is enforced by a compile-time import check, and `omni scan` registers with the `internal/cli/command/` Registry and works inside pipelines.
**Plans**: TBD
**Pitfalls addressed**: 4 (vuln matcher accuracy), 11 (DB staleness), 17 (dep bloat)

### Phase 7: pkg/attest/ — SLSA Attestation
**Goal**: Users can generate and verify DSSE-wrapped, in-toto-formatted SLSA v1.0 provenance attestations at the honest SLSA level omni's own release process actually achieves.
**Depends on**: Phase 6
**ADR Gate (required before any code lands)**:
  - ADR pinning the honest SLSA level omni's release process achieves — almost certainly L2 via GitHub Actions + OIDC, not L3. No provenance output may claim a higher level than the ADR. Required before any predicate builder code ships.
**Requirements**: ATTEST-01, ATTEST-02, ATTEST-03, ATTEST-04, ATTEST-05, ATTEST-06, ATTEST-07, ATTEST-08, ATTEST-09
**Success Criteria** (what must be TRUE):
  1. A user can run `omni attest --predicate-type slsa-provenance --predicate <file> --artifact <path>` and receive an in-toto Statement with a SLSA v1.0 provenance predicate wrapped in a DSSE envelope — the PAE format is validated against in-toto reference test vectors in unit tests.
  2. A user can run `omni attest verify <envelope>` with a pubkey and get a fail-closed verification — every error mode returns non-zero with `cmderr` classification.
  3. A user running in GitHub Actions can use `--from-env` to auto-populate provenance fields from `GITHUB_RUN_ID`, `GITHUB_WORKFLOW`, `GITHUB_SHA`, etc. — `builder.id` is derived from OIDC claims, not hardcoded.
  4. SLSA predicates are validated against the official SLSA v1.0 JSON schema in a CI step that runs before every release; the claimed level is identical to the ADR.
  5. DSSE Pre-Authentication Encoding is implemented in exactly one place (`pkg/attest/dsse/pae.go`) and is never inlined; `pkg/attest/` depends on `pkg/sign/` and consumes `pkg/sbom/format.Document` for subject digests, does not depend on `pkg/scan/`, and `omni attest` registers with the `internal/cli/command/` Registry.
**Plans**: TBD
**Pitfalls addressed**: 5 (SLSA overclaim), DSSE PAE footgun

### Phase 8: v1.0 Release Cut
**Goal**: Ship v1.0 as signed, reproducible binaries for six target platforms with SBOM and SLSA provenance generated by omni itself — dogfood the supply chain before announcing it.
**Depends on**: Phase 7
**Requirements**: REL-01, REL-02, REL-03, REL-04, REL-05, REL-06, REL-07, REL-08
**Success Criteria** (what must be TRUE):
  1. The `pkg/` API freeze is tagged at phase entry and the 30-day CLAUDE.md deprecation protocol applies to every breaking change from that point forward.
  2. Signed binaries are published for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`, `windows/arm64` as a GitHub release — each signature is produced by `omni sign` (dogfooding).
  3. Binaries are built reproducibly with `-trimpath` + `-buildvcs=true`, and a CI dual-build comparison job fails on any reproducibility drift.
  4. The release asset set includes an SBOM generated by `omni sbom` and a SLSA v1.0 provenance attestation generated by `omni attest` at the honest level from ATTEST-07.
  5. Release notes (generated from conventional commits) include an explicit "What's NOT protected against" section linking `.planning/research/PITFALLS.md`, and the v1.0 announcement explicitly states the target audience ("me + my CI/CD pipelines") without overclaiming general-purpose fitness.
**Plans**: TBD
**Pitfalls addressed**: 6 (reproducibility), 15 (announcement overclaim), 16 (Windows gaps)

---

## Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. cmderr Migration Completion | 0/? | Not started | - |
| 2. Test Coverage + Deterministic Golden Harness | 0/? | Not started | - |
| 3. CONCERNS Burn-down, Windows Parity, pkg/ API Audit, Docs | 0/? | Not started | - |
| 4. pkg/sign/ — Signing Primitive | 0/? | Not started | - |
| 5. pkg/sbom/ — SBOM Generation | 0/? | Not started | - |
| 6. pkg/scan/ — Vulnerability Scanning | 0/? | Not started | - |
| 7. pkg/attest/ — SLSA Attestation | 0/? | Not started | - |
| 8. v1.0 Release Cut | 0/? | Not started | - |

---

## Ordering Constraints (non-negotiable)

1. Phase 1 (cmderr) must finish before Phase 4 begins — prevents inconsistent exit-code semantics on security commands (Pitfall 13).
2. Golden-master timestamp/random-ID normalization must land in Phase 2, not improvised in Phase 5+ (Pitfall 14).
3. No parallelization across supply-chain phases (4→5→6→7) — premature API commitments between sbom and scan would harden the boundary before it's tested.
4. ADR gates at Phase 4, 5, 7 entries must merge before any implementation code lands in those phases.
5. Phase 6 requires a `/gsd-research-phase 6` spike before plan generation — DB distribution format is the open question.
6. Parallelization *within* a phase across independent plans is permitted.

---

*Last updated: 2026-04-11 at initialization*
