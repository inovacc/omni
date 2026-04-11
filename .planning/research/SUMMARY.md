# Project Research Summary

**Project:** omni — v1.0 supply-chain capability track
**Domain:** Go-native supply-chain tooling (SBOM, signing, vulnerability scanning, SLSA attestation) inside an existing single-binary hexagonal CLI
**Researched:** 2026-04-11
**Confidence:** HIGH on stack, architecture, and pitfalls; MEDIUM on feature categorization (audience-dependent)

## Executive Summary

omni's v1.0 supply-chain track adds four capabilities — `omni sbom`, `omni sign`/`verify`, `omni scan`, `omni attest` — that together let a single pure-Go binary replace the syft + cosign + grype + slsa-github-generator toolchain for a Go-first CI/CD pipeline. Research across all four areas converges on a narrow, opinionated stack: **minisign (aead.dev/minisign) as the default signing path**, cyclonedx-go + spdx/tools-golang for SBOM serialization, govulncheck's public `scan` API for vulnerability matching, and in-toto-golang + go-securesystemslib for DSSE-wrapped SLSA v1.0 provenance. Sigstore bundle verification (sigstore-go) and polyglot OSV scanning (osv-scanner v2) are explicitly gated behind build tags to keep the default binary lean. Syft, grype, and cosign are rejected as libraries by all three research tracks — all pull dependency graphs that would bloat omni by 40–60 MB and, in cosign's case, upstream actively steers integrators to sigstore-go instead.

The four capabilities form a strict DAG — **sign → sbom → scan → attest** — not a set of parallelizable phases. Signing is the cryptographic primitive every other capability eventually reuses; scan must consume SBOMs as serialized documents (never via struct sharing, mirroring grype's boundary with syft); attest composes all three. Architecture research insists the existing hexagonal layout (`cmd/` → `internal/cli/` → `pkg/`) maps cleanly onto syft's and grype's own module splits, with cataloger-per-ecosystem subdirectories registered via `init()` following the proven `pkg/video/extractor/` pattern.

The dominant risk is "security theater" — shipping tools that *look* like supply-chain security but fail open on missing signatures, miss CVEs due to wrong purls, or claim SLSA levels the build process does not actually achieve. Mitigations are mechanical but non-negotiable: fail-closed defaults with a negative-test matrix on every verify path, purl round-trip validation against syft output, OSV regression suite as golden tests, an ADR pinning omni's honest SLSA level (almost certainly L2, not L3) before `omni attest` ships, and DSSE Pre-Authentication Encoding tested against in-toto's reference vectors. Equally important is **ordering discipline**: the `cmderr` polish-track migration must finish before any supply-chain command is touched, so exit-code semantics on security-critical commands stay consistent.

## Key Findings

### Recommended Stack

Seven net-new direct dependencies, all pure-Go, all stdlib-adjacent. Two more behind build tags. No CGO paths, no containerd, no OCI registry clients, no Rekor/Fulcio in the default binary. Full detail in [STACK.md](./STACK.md).

**Core technologies:**
- `aead.dev/minisign` — default signing/verification — pure Go, stdlib-only, zero network deps, interop-compatible with the jedisct1 CLI
- `github.com/CycloneDX/cyclonedx-go` + `github.com/spdx/tools-golang` — SBOM serialization — upstream spec owners, pure Go, zero transitive weight
- `golang.org/x/vuln` (scan subpackage) — vulnerability matching — Go team-maintained public API, stable since v1.0, scans source and binaries
- `github.com/in-toto/in-toto-golang` + `github.com/secure-systems-lab/go-securesystemslib` — SLSA v1.0 provenance + DSSE envelopes — reference implementations, pure Go
- `debug/buildinfo` (stdlib) + `golang.org/x/mod` — module graph extraction — ground truth for binary SBOMs

**Build-tag gated (opt-in, not in default binary):**
- `github.com/sigstore/sigstore-go` (tag: `omni_sigstore`)
- `github.com/google/osv-scanner/v2` (tag: `omni_osv`)

**Explicitly rejected:** syft (as library), grype (as library), cosign (as library), slsa-github-generator, cyclonedx-gomod, jedisct1/go-minisign, anything pulling `mattn/go-sqlite3` or containerd.

### Expected Features

Full detail in [FEATURES.md](./FEATURES.md). Split is opinionated around PROJECT.md's "me + my CI/CD pipelines" audience.

**Must have (table stakes):**
- SBOM from Go module + binary via `debug/buildinfo`, SPDX 2.3 + CycloneDX 1.5 JSON, deterministic ordering, purl identifiers
- Ed25519 blob sign/verify/keygen, detached `.sig` files, passphrase-protected keys, env-var + file key loading
- OSV-backed scanning, `--fail-on <severity>` gating, JSON + text output, SBOM as input, offline mode with cached DB
- in-toto statement + SLSA v1.0 provenance predicate, DSSE envelope, reads GitHub Actions env vars

**Should have (differentiators):**
- One binary, zero deps (the core thesis)
- Pure-Go `pkg/` library API across all four capabilities (unique vs syft/cosign/grype which are binary-first)
- Chainable via `omni pipe` / `omni pipeline` — `omni sbom . | omni scan - | omni attest --predicate -` in one process tree
- SBOM from `debug/buildinfo` of a built binary (verify-what-shipped)
- Reachability analysis via `golang.org/x/vuln` for dramatic false-positive reduction
- Merkle-tree signing of directory trees via `pkg/twig`

**Defer (v2+):** OCI registry push/pull, Fulcio keyless signing flow, Rekor transparency log, container image scanning, full cosign CLI parity, SLSA v0.2 legacy schema, non-Go SBOM ecosystems, policy-as-code engine, daemon mode.

### Architecture Approach

Four `pkg/` packages wired through the existing three-layer split. Each new command registers with `internal/cli/command/` to enable pipe chaining from day one. Full detail in [ARCHITECTURE.md](./ARCHITECTURE.md).

**Major components:**
1. **`pkg/sign/`** — cryptographic primitive: `keys`, `minisign` (pure-Go default), `sigstore` (opt-in), `verify` (dispatch). Zero dependencies on the other three.
2. **`pkg/sbom/`** — `source` (reuses `pkg/twig/scanner`), `cataloger/{gomod,npm,ociimage}/` (`init()`-registered via a `cataloger/all/` barrel mirroring `pkg/video/extractor/all/`), `model` (neutral document type), `format/{spdx,cyclonedx}`.
3. **`pkg/scan/`** — `model`, `db` (signature-verified offline DB loader), `osv`, `matcher`. **Consumes SBOMs via the serialized `format.Document`, never via `pkg/sbom/model` struct imports** — non-negotiable boundary.
4. **`pkg/attest/`** — `intoto`, `slsa`, `dsse` (PAE with spec test vectors pinned), `bundle` (optional). Depends on sign + sbom; does not depend on scan (scan results belong in a VEX predicate, not SLSA provenance).

### Critical Pitfalls

Top five from [PITFALLS.md](./PITFALLS.md). Any of these shipping in v1.0 would be worse than not shipping the feature at all.

1. **Missing or wrong PURLs in SBOMs** — silent security failure. **Prevention:** use module path (not import path) from `debug.BuildInfo.Deps`; normalize pseudo-versions exactly as Go does; round-trip test every SBOM against syft + govulncheck before merge.
2. **Signature verification falling open** — missing sig passes, unknown algorithms "skipped". **Prevention:** fail-closed default, explicit algorithm allowlist, negative-test matrix covering every failure mode (missing, wrong key, wrong issuer, expired, revoked, bad algo, tampered payload).
3. **DSSE Pre-Authentication Encoding** — the #1 attest footgun. PAE format `DSSEv1 <len(type)> <type> <len(payload)> <payload>` trips every hand-rolled implementation. **Prevention:** implement PAE once in `pkg/attest/dsse/pae.go` and pin in-toto's reference test vectors as unit tests.
4. **SLSA level overclaim** — provenance generated in the same container as the build is not tamper-evident; hardcoded `builder.id` is self-asserted. **Prevention:** write an ADR stating which SLSA level omni's release process *actually* achieves (likely L2) before `omni attest` ships; validate predicates against official SLSA JSON schema in CI.
5. **Signing key handling** — keys in git history, slog attributes, `/proc/*/cmdline` via flag values. **Prevention:** never accept key material as a flag value (only file paths or env-var-pointing-to-file); typed `secret.Key` wrapper with redacting `String()`; `check-leaks.sh` as pre-commit gate.

## Implications for Roadmap

Eight phases in strict DAG order. Do **not** compress into fewer phases. Do **not** parallelize supply-chain phases — `cmderr` must finish first and cross-package API commitments between `pkg/sbom` and `pkg/scan` would happen prematurely.

### Phases 1–3: Polish track

**Rationale:** Pitfall 13 (cmderr + supply-chain in parallel) and Pitfall 14 (golden tests baking in timestamps) force this ordering. `cmderr` migration must complete before any supply-chain command is written. Pitfalls 7, 8 (semver + `pkg/` surface) force an API audit while things are still movable.

**Delivers:** cmderr fully adopted (160/160), 60–80% coverage, golden-master harness with timestamp normalization (required for SBOM/attestation goldens later), `pkg/` API audit with unstable surface moved to `internal/` or annotated `// Experimental:`.

**Avoids:** Pitfalls 7, 8, 10, 13, 14.

### Phase 4: `pkg/sign/` — signing primitive

**Rationale:** Lowest-risk warmup (pure `crypto/ed25519` + well-documented minisign format); unblocks Phase 7 (attest) which cannot ship without DSSE signing; sign is the primitive every other capability reuses.

**Delivers:** `omni sign keygen`, `omni sign`, `omni verify`. Minisign-compatible signatures, passphrase-protected Ed25519 keys, `cmderr.ErrConflict` on signature mismatch.

**Uses:** `aead.dev/minisign`, stdlib `crypto/ed25519`, existing `pkg/cryptutil`.

**Avoids:** Pitfalls 2 (fail-closed verify), 3 (key handling), 17 (unmaintained deps).

**Gate:** ADR on key handling policy (storage, rotation, dev-vs-release separation) AND cosign-compat scope must be merged before any code lands.

### Phase 5: `pkg/sbom/` — SBOM generation

**Rationale:** Unblocks scan (serialized document) and attest (subject digests). Medium risk — SPDX 2.3 has footguns, but the `cataloger/<ecosystem>/` pattern is proven in `pkg/video/extractor/`.

**Delivers:** `omni sbom <target> --format spdx|cyclonedx`, Go module + Go binary sources, deterministic ordering, purl identifiers, `cataloger/gomod/` only.

**Uses:** `cyclonedx-go`, `spdx/tools-golang`, `debug/buildinfo`, `golang.org/x/mod`, existing `pkg/twig/scanner`, existing `pkg/hashutil`.

**Avoids:** Pitfalls 1 (purl correctness — round-trip test vs syft mandatory), 12 (SBOM completeness — use `debug.BuildInfo` as ground truth).

**Gate:** ADR on SBOM round-trip test matrix before implementation.

### Phase 6: `pkg/scan/` — vulnerability scanning

**Rationale:** Requires Phase 5's document format and Phase 4's sign (for DB signature verification). Highest integration risk in the track — DB distribution is the hardest part, and OSV ecosystem has moved fast in 2025–2026.

**Delivers:** `omni scan <sbom>`, `omni scan db update`, `--fail-on <severity>`, offline mode with cached DB, `--max-db-age` gate.

**Uses:** `golang.org/x/vuln` (govulncheck public API), OSV schema archives signed via `pkg/sign/verify`, existing `go.etcd.io/bbolt` or flat JSON (spike needed).

**Implements:** `pkg/scan/{model,db,osv,matcher}` — consumes `pkg/sbom/format.Document`, translates at the boundary. Never imports `pkg/sbom/model` structs.

**Avoids:** Pitfalls 4 (OSV regression suite as golden tests), 11 (DB staleness — `--max-db-age` gate, loud warnings, no silent offline fallback).

**NEEDS DEEPER RESEARCH** during phase planning: DB distribution format, OSV-scanner v2 API surface, grype v6 schema direction.

### Phase 7: `pkg/attest/` — SLSA attestation

**Rationale:** Must be last — composes sign + sbom. Low crypto risk, medium spec-compliance risk (PAE and SLSA v1 predicate schema must be byte-exact).

**Delivers:** `omni attest --predicate-type slsa-provenance --predicate file.json --artifact X`, `omni attest verify`, GitHub Actions env var reader, DSSE envelope, optional Sigstore bundle.

**Uses:** `in-toto-golang`, `in-toto/attestation` predicates, `go-securesystemslib`, reuses `pkg/sign/keys`.

**Implements:** `pkg/attest/{intoto,slsa,dsse,bundle}`. Golden tests pinned against in-toto reference vectors.

**Avoids:** Pitfall 5 (SLSA overclaim), DSSE PAE footgun.

**Gate:** ADR pinning honest SLSA level (almost certainly L2 via Actions + OIDC, not L3) before any predicate builder code ships.

### Phase 8: Release — v1.0 cut

**Delivers:** `pkg/` API semver freeze with 30-day deprecation window, signed linux/darwin/windows × amd64/arm64 binaries, reproducible builds (`-trimpath -buildvcs` + dual-build cmp in CI), release notes with explicit "What's NOT included" section linking PITFALLS.md.

**Avoids:** Pitfalls 6 (non-reproducible builds), 15 (announcement overclaim), 16 (Windows parity gaps).

### Phase Ordering Rationale

- **Polish before new code** — PROJECT.md + Pitfall 13 both demand it.
- **sign → sbom → scan → attest DAG** — forced by dependency structure, validated against syft/grype/cosign/sigstore-go module graphs.
- **No parallelism across supply-chain phases** — single-maintainer reality + preventing premature API commitments.
- **ADR gates at Phase 4, 5, 7 entries** — scope-lock decisions, not documentation tasks.

### Research Flags

**Needs deeper research (`/gsd-research-phase`):**
- **Phase 6 (scan):** OSV DB distribution format is the single biggest open question. Training data on osv-scanner v2, grype v6, and `golang.org/x/vuln` public API is stale.
- **Phase 7 (attest):** SLSA v1.0 vs draft build-provenance schema + current GitHub Actions OIDC claim structure.
- **Phase 8 (release):** Reproducibility drift in recent Go toolchain (`-buildvcs`, `toolchain` directive).

**Standard patterns (skip research-phase):**
- **Phase 4 (sign):** Minisign is well-specified, stdlib crypto is frozen.
- **Phase 5 (sbom):** cyclonedx-go + spdx/tools-golang stable; `debug/buildinfo` is stdlib. Main risk is a test-matrix problem, not a research problem.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Pure-Go verified for every library; syft/cosign/grype rejections validated against dep graphs and upstream migration direction |
| Features | MEDIUM | Table-stakes vs differentiator split is audience-dependent — opinionated around "me + CI/CD pipelines" |
| Architecture (boundaries) | HIGH | DAG order and "scan consumes document not structs" validated against grype behavior; cataloger pattern mirrors existing `pkg/video/extractor/` |
| Architecture (package split) | MEDIUM | Exact subdirectory layout is opinion; first implementation may reshape |
| Pitfalls | HIGH | SLSA/SBOM/sigstore pitfalls are spec-derived with reference test vectors |

**Overall confidence:** HIGH on direction and ordering; MEDIUM on feature categorization and package-split details.

### Gaps to Address

- **DB distribution format for `omni scan`** — flat JSON vs bbolt vs modernc SQLite. Leaning flat per-ecosystem JSON (grype v6 direction). Phase 6 entry research.
- **Honest SLSA level for omni's own release** — likely L2 via Actions + OIDC. ADR required before `omni attest` ships and before v1.0 announcement.
- **Cosign-compat scope** — recommendation: minisign-only in v1.0, Sigstore bundle-format verification as the opt-in interop point (not Rekor, not Fulcio, not OCI). Lock at Phase 4 entry.
- **`pkg/` API audit scope** — how much current `pkg/*` moves to `internal/` in Phase 3.
- **Golden-test timestamp normalization** — must be added to harness in Phase 2, not improvised in Phase 5.
- **Windows parity for signing/SBOM** — path separators in purls, CRLF in signature files, binary-mode I/O. Add to Phase 3 parity burn-down; revalidate in Phase 4/5.

## Sources

### Primary (HIGH confidence)
- `.planning/PROJECT.md` (root of truth for audience + out-of-scope)
- SLSA v1.0 specification + attestation model
- Sigstore bundle format spec + sigstore-go upstream "use this, not cosign" guidance
- OSV schema, purl spec, CycloneDX 1.5/1.6, SPDX 2.3/3.0 specifications
- Go `debug/buildinfo`, `golang.org/x/mod/semver`, `golang.org/x/vuln`
- in-toto attestation framework + DSSE v1.0 spec with reference test vectors
- anchore/syft and anchore/grype module graphs (rejection validated); grype architecture docs (scan-consumes-document boundary validated)

### Secondary (MEDIUM confidence)
- osv-scanner v2 API surface (stable but dep-heavy)
- Grype v6 schema direction
- Reproducible-builds.org Go documentation

### Tertiary (LOW confidence)
- Feature table-stakes vs differentiator split — audience-opinionated
- Exact `pkg/` sub-package layout — first-implementation opinion

### Research Files (in `.planning/research/`)
- STACK.md, FEATURES.md, ARCHITECTURE.md, PITFALLS.md

---

### Confidence: HIGH
### Gaps: DB distribution format (Phase 6 spike), honest SLSA level (Phase 7 ADR), cosign-compat scope (Phase 4 ADR), `pkg/` audit scope (Phase 3), golden timestamp normalization (Phase 2), Windows parity (Phase 3 + revalidate).
