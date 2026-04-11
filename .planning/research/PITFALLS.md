# Domain Pitfalls

**Domain:** Supply-chain tooling (SBOM, signing, scanning, attestation) + Go CLI v1.0 release
**Researched:** 2026-04-11
**Confidence:** HIGH for well-known SLSA/SBOM/sigstore pitfalls (spec-derived); MEDIUM for Go module/semver pitfalls (ecosystem-derived)

---

## Critical Pitfalls

These cause either (a) false sense of security, (b) broken signatures, (c) wrong vuln matches, or (d) spec violations. Any of these shipping in v1.0 would be worse than not shipping the feature at all.

### Pitfall 1: SBOM with missing or wrong PURLs

**What goes wrong:** SBOM generator emits components without a valid `purl` (Package URL), or emits purls with wrong type/namespace/version. Downstream scanners can't match components to CVEs, so `omni scan` against the SBOM returns "0 vulnerabilities" on an actually-vulnerable artifact.

**Why it happens:** Go modules don't map cleanly to purl — `pkg:golang/github.com/foo/bar@v1.2.3` requires the module path, not the package import path. Replace directives, `+incompatible` suffixes, pseudo-versions (`v0.0.0-YYYYMMDDHHMMSS-abcdef123456`), and vendored modules all break naive implementations.

**Consequences:** Silent security failure. SBOM looks correct, scanner finds nothing, user ships vulnerable code thinking they're clean. This is THE canonical SBOM foot-gun.

**Warning signs:**
- Scanner returns suspiciously low/zero findings on a known-vulnerable binary
- purls missing `pkg:` prefix or containing file paths
- `+incompatible`, `replace` directives, or pseudo-versions not round-tripping
- CycloneDX validator warns on missing `bom-ref` or `purl` fields

**Prevention:**
- Use module path (from `debug.BuildInfo.Deps`), never import path, for purl namespace
- Normalize pseudo-versions exactly as Go does — `vX.Y.Z-0.YYYYMMDDHHMMSS-abcdef123456`
- Preserve `replace` directives as separate components with original + replacement both listed
- Cross-validate generated SBOMs against syft output on the same binary in CI
- Round-trip test: generate SBOM → scan with `osv-scanner` → compare to `govulncheck` output on source

**Phase:** Supply-chain (phase 4, `omni sbom`). MUST have round-trip tests before merge.

---

### Pitfall 2: Signature verification that silently falls open

**What goes wrong:** `omni verify` returns exit 0 when it should return non-zero. Common causes: unknown algorithm treated as "skip", missing signature treated as "valid", key/cert chain verification that only checks cryptographic validity but not trust root, time-of-check bypass via clock skew.

**Why it happens:** Developers write `if sig != nil && sigValid(sig)` — missing signature passes. Or they verify the signature cryptographically but forget to verify the *identity* bound to the cert (Fulcio OIDC issuer/subject). Or they trust `NotBefore`/`NotAfter` from the cert without independent timestamp (RFC 3161) verification.

**Consequences:** Attacker-supplied unsigned or wrong-signer artifact passes verification. This is **exactly** what supply chain tooling is supposed to prevent, so shipping this bug is worse than shipping nothing.

**Warning signs:**
- Test suite doesn't include "empty signature", "wrong key", "expired cert", "wrong issuer" negative cases
- `verify` has any code path that returns nil without having actually checked a signature
- No explicit "algorithm not allowed" list (crypto-agility done wrong)
- Clock checks use `time.Now()` without bounding to signing-time via RFC 3161 TSA

**Prevention:**
- **Fail-closed default:** no signature = error. Unknown algorithm = error. Missing cert = error.
- Separate the "signature is mathematically valid" check from the "signer is authorized" check — both must pass
- Negative test matrix: every failure mode (missing, wrong key, wrong issuer, expired, revoked, bad algo, tampered payload) MUST have an explicit test asserting non-zero exit
- Use RFC 3161 timestamps for long-term verification, not cert validity windows
- Allowlist algorithms explicitly (ed25519, ecdsa-p256-sha256, rsa-pss-sha256); reject all others
- If doing Sigstore/Fulcio: bind and verify the OIDC issuer + subject, not just cert-chain validity

**Phase:** Supply-chain (phase 4, `omni sign`/`omni verify`). This is security-critical; flag for security review before merge.

---

### Pitfall 3: Signing key handling — committed, logged, or leaked via process table

**What goes wrong:** Private keys end up in git history, in logs (slog captures them on error), in `/proc/*/cmdline` because they were passed as a flag, or in panic stack traces. Or the release signing key is reused for local testing and gets exfiltrated from a dev machine.

**Why it happens:** Convenience. `--key /path/to/priv.pem` feels natural. `slog.Info("signing", "key", key)` happens in a 3am debug session and never gets removed. Test fixtures contain real-looking keys that get committed. No separation between "dev test key" and "release signing key".

**Consequences:** Silent long-term compromise. If the v1.0 release key leaks, every released binary signature becomes worthless and can't be distinguished from attacker forgeries.

**Warning signs:**
- `gitleaks`/`check-leaks.sh` flags PEM blocks or base64 key material
- slog attributes ever include `key`, `secret`, `signature` as values (not just names)
- Keys passed as CLI flags instead of file paths or env vars + file descriptors
- No documented key-rotation procedure

**Prevention:**
- **Never** accept key material as a flag value — only file paths or env var names pointing to files
- slog: use `slog.LogValuer` to redact key types, or use a typed `secret.Key` wrapper with custom `String()` returning `<redacted>`
- Pre-commit hook: run `check-leaks.sh` on every commit, block on match
- Test fixtures: generate ephemeral keys in `TestMain`; never commit key files even in `testdata/`
- Release signing key lives only in GitHub Actions OIDC → Sigstore (keyless) or in a hardware token. Never on a dev machine.
- Document rotation: if leaked, how to rotate and re-sign within 24 hours

**Phase:** Supply-chain (phase 4) and Release (phase 6). Key handling policy MUST be written before any signing code is merged.

---

### Pitfall 4: Vulnerability scanner matching false negatives (and false positives)

**What goes wrong:** `omni scan` matches CVEs by package name + version, but misses vulns because:
- Go stdlib vulns (GO-YYYY-NNNN) aren't in package databases at all — require `govulncheck` symbol-level matching
- Version range semantics wrong: OSV uses `introduced`/`fixed` events, not simple ranges
- Ecosystem mismatch: treating `pkg:golang/x` as `pkg:generic/x` against npm DB
- Module replaces not followed — vuln exists in original module, replacement is reported
- `+incompatible` versions not normalized

**Why it happens:** Naive implementations do `if vuln.Package == dep.Name && vuln.Version == dep.Version`. Real OSV matching requires a proper semver range evaluator per ecosystem, symbol-level reachability for Go, and aliasing between CVE/GHSA/OSV IDs.

**Consequences:** Silent false negatives (security theater) or noisy false positives (users disable scanning). Both destroy trust in the tool.

**Warning signs:**
- Scanner results disagree with `govulncheck` or `osv-scanner` on the same input
- No test fixtures covering version-range edge cases (pre-releases, `+incompatible`, `0.0.0`)
- CVE/GHSA/OSV alias resolution not implemented
- Stdlib vulns never reported

**Prevention:**
- Use OSV schema verbatim for DB matching; don't invent a version comparator
- For Go specifically: integrate `golang.org/x/vuln` for symbol-level matching, or clearly document "package-level only, use govulncheck for reachability"
- Test against the OSV-scanner regression suite as a golden-master set
- Always report the matching source (OSV ID + DB version) so users can audit
- Honor `withdrawn` field — withdrawn advisories must NOT match
- Range evaluation: use ecosystem-specific semver (Go's `golang.org/x/mod/semver`, not generic semver)

**Phase:** Supply-chain (phase 5, `omni scan`). Must include OSV regression suite as golden tests.

---

### Pitfall 5: SLSA provenance that doesn't actually prove anything

**What goes wrong:** `omni attest` generates SLSA v1.0 provenance predicates, but:
- `builder.id` is self-asserted (attacker can claim to be GitHub Actions)
- `buildType` URI doesn't resolve or isn't a stable identifier
- `externalParameters` omits source commit / ref / workflow — provenance can't be linked to source
- `resolvedDependencies` missing or incomplete — can't verify hermeticity
- Predicate signed by the same key as the artifact, collapsing two trust roots into one
- Generated inside the build step instead of an isolated builder — not tamper-evident

**Why it happens:** SLSA specs are easy to misread as "emit this JSON blob". The spec's trust model requires an *isolated builder* that the artifact build step cannot modify, and provenance validity depends on who signed it, not what's in it.

**Consequences:** SLSA level claimed is higher than actually achieved. Consumers trust L3 provenance that's really L1. Compliance-motivated users pass audits on false pretenses.

**Warning signs:**
- Provenance is generated in the same process/container as the build
- `builder.id` is a string constant, not derived from a verifiable source (e.g., GitHub OIDC token `aud`+`iss`+`sub`)
- No test asserting `externalParameters` contains the exact source commit
- Single key signs both artifact and attestation

**Prevention:**
- Document achievable SLSA level HONESTLY — if omni's builder model is L2, say L2, not L3
- For v1.0: target SLSA v1.0 Build L2 via GitHub Actions reusable workflow + OIDC → Sigstore (keyless). Don't claim L3 without an isolated builder.
- Validate generated predicates against official SLSA JSON schema in CI
- `builder.id` derived from OIDC claims, not hardcoded
- `resolvedDependencies` populated from `debug.BuildInfo` + go.sum hashes
- Write an ADR stating which SLSA level omni's release process achieves and why

**Phase:** Supply-chain (phase 5, `omni attest`) + Release (phase 6, apply to omni itself).

---

### Pitfall 6: Non-reproducible builds invalidating the entire chain

**What goes wrong:** `omni build` produces different binaries on different machines/runs due to embedded timestamps, absolute paths, VCS state, build cache contents, or non-deterministic map iteration order in codegen. SBOMs don't match, signatures don't verify cross-machine, reproducible-build verifiers flag the release.

**Why it happens:** Forgetting `-trimpath`, forgetting `-buildvcs=false` (or forgetting to commit first), embedding `time.Now()` at build time, using `go generate` with non-deterministic output, CGO introducing toolchain drift.

**Consequences:** Users can't independently verify omni's own release. "Supply chain tooling that isn't itself verifiable" is the worst look.

**Warning signs:**
- `go build` + `go build` on same source → different `sha256sum`
- `-trimpath` not in Taskfile release target
- Binary contains `/home/$user/...` or `C:\Users\...` paths (check with `strings`)
- CI build and local build disagree on hash

**Prevention:**
- Release build flags: `-trimpath -ldflags="-s -w -buildid=" -buildvcs=true` with clean working tree
- Pin Go toolchain exact version via `go.mod`'s `go 1.XX.Y` and `toolchain` directive
- CI job: build twice, diff binaries, fail on mismatch (`diffoscope` or `cmp`)
- No `time.Now()` or random values baked into codegen; use `SOURCE_DATE_EPOCH` if needed
- Zero CGO — already a project constraint, keep enforcing
- Document the exact reproducer command in release notes

**Phase:** Release (phase 6). Add reproducibility check to CI before cutting v1.0.

---

## Moderate Pitfalls

### Pitfall 7: Go semver and `v2+` module path trap

**What goes wrong:** Project ships v1.0, later needs v2.0 with breaking changes, but import path is still `github.com/inovacc/omni` without `/v2` suffix. Go modules require the major version suffix in the import path for v2+, so either (a) every consumer breaks or (b) the project can never actually ship v2.

**Prevention:**
- Before tagging v1.0, decide: is the `pkg/` API going to need breaking changes within 2 years? If yes, plan the `/v2` migration path now.
- Document in ROADMAP.md that post-v1.0 breaking changes will ship as `github.com/inovacc/omni/v2` via new directory layout
- Don't tag v1.0.0 until `pkg/` API is genuinely reviewed — `v0.x` is free; `v1.x` is a promise
- For internal APIs not meant to be stable: move to `internal/` now. Anything in `pkg/` is a public commitment.

**Phase:** Release (phase 6), but API audit should happen during Polish (phase 1–3) while things are still movable.

---

### Pitfall 8: `pkg/` surface too large at v1.0

**What goes wrong:** Every currently-exported function in `pkg/` becomes a semver commitment the moment v1.0 ships. Most `pkg/` code was written as internal helpers and is over-exposed. Every future refactor triggers a 30-day deprecation cycle for APIs nobody uses.

**Prevention:**
- Audit `pkg/*` before v1.0: anything that shouldn't be stable → move to `internal/` or rename with `Internal` prefix
- Add `// Deprecated:` or `// Experimental:` comments to anything unsure — gives legal cover to break later
- For each `pkg/` package, write a one-line "what's the stability promise?" in a package doc comment
- Consider splitting public API into `pkg/<name>/api` sub-packages so internal helpers stay unexported

**Phase:** Polish (phase 3 — API audit) before release freeze in phase 6.

---

### Pitfall 9: Cosign/Sigstore compatibility half-done

**What goes wrong:** v1.0 claims "cosign-compatible" but only implements a subset — missing OCI signature discovery, missing bundle format support, missing Rekor transparency log upload. Users try `cosign verify omni-signed-blob` and it fails.

**Prevention:**
- Be explicit in docs about what subset is supported: "omni produces/verifies Sigstore bundles (v0.3), does not push to OCI registries, does not query Rekor"
- Implement the **bundle format** (protobuf spec) as the interop point — it's the canonical serialization and cosign/policy-controller/gitsign all consume it
- Integration test: sign with omni → verify with cosign CLI → sign with cosign → verify with omni. Both directions must pass.
- If full cosign parity is too big for v1.0, ship "minisign-style pure-Go signing" as the ONLY supported mode and explicitly don't claim cosign compatibility

**Phase:** Supply-chain (phase 4). Scope decision must happen at phase entry, not mid-implementation.

---

### Pitfall 10: CLI flag stability vs. output format stability

**What goes wrong:** v1.0 commits to `--json` output shape. Users pipe it to `jq` in CI. Later, omni needs to add a field → safe. Later, omni needs to rename a field → breaks every consumer's pipeline. Worse: omni never committed to a shape explicitly, so every "minor version" feels breaking to someone.

**Prevention:**
- Version the JSON schema: every `--json` output includes `"schemaVersion": "1"` at the top level
- Document explicitly: "additive changes (new fields) = minor. Renaming or removing fields = major + schemaVersion bump."
- Golden master tests pin the exact JSON shape — any diff is a deliberate decision
- Consider `--json-schema` flag that prints the JSON Schema for the current version

**Phase:** Polish (phase 2) — apply to existing commands before v1.0 freeze.

---

### Pitfall 11: Vuln DB staleness + offline mode semantics

**What goes wrong:** `omni scan` caches the OSV DB locally. Cache goes stale (days old), scanner reports clean, real vulns exist. Or: user runs `omni scan` offline in an air-gapped CI, gets "no vulns" because DB was never downloaded, but exit code is 0 and they don't notice.

**Prevention:**
- Always print the DB version/age in output: "Matched against OSV DB snapshot 2026-04-10T12:00:00Z (1 day old)"
- `--max-db-age` flag with sane default (e.g., 7 days), fail with non-zero exit if exceeded
- Offline mode must be explicit (`--offline`) and warn loudly
- Never default to offline; never silently fall back to offline on network failure — fail instead

**Phase:** Supply-chain (phase 5, `omni scan`).

---

### Pitfall 12: SBOM component completeness — missing transitive deps and binaries

**What goes wrong:** SBOM includes direct deps from `go.mod` but misses:
- Transitive deps pulled only by `// indirect` paths
- Build-time tools (protoc plugins, code generators) that influenced output
- Embedded files via `//go:embed` (licensing implications)
- CGO-linked C libraries (omni is pure Go so mostly N/A, but document the assumption)
- The Go toolchain itself as a component

**Prevention:**
- Use `debug.BuildInfo` from the compiled binary, not `go.mod` parsing — it's the ground truth of what's actually in the binary
- Include Go toolchain version as a component (`pkg:golang/std@1.XX.Y`)
- Validate SBOM completeness against `go version -m <binary>` output
- Document what omni's SBOM does and doesn't include in help text

**Phase:** Supply-chain (phase 4).

---

## Minor Pitfalls

### Pitfall 13: Taking on `cmderr` migration and supply chain in parallel

**What goes wrong:** Half-migrated `cmderr` state means supply-chain commands have to decide which error model to use. Some use `cmderr`, some don't, exit codes become inconsistent specifically on the security-critical commands.

**Prevention:** Finish `cmderr` migration in Polish phase BEFORE starting supply-chain commands. Non-negotiable ordering.

**Phase:** Polish (phase 1) must complete before Supply-chain (phase 4) starts.

---

### Pitfall 14: Golden tests for supply-chain output that hardcode timestamps

**What goes wrong:** SBOM/attestation output includes timestamps. Golden masters snapshot them. Tests pass once then fail forever or get regenerated blindly.

**Prevention:** Normalize timestamps in golden test harness (replace ISO-8601 with `<TIMESTAMP>` before diffing); or inject a fixed clock via dependency injection; or document which fields are volatile and exclude from diff.

**Phase:** Polish (phase 2, golden test coverage) + Supply-chain (phase 4, SBOM tests).

---

### Pitfall 15: Release announcement overclaiming

**What goes wrong:** v1.0 announcement says "full supply chain security" when what shipped is "SBOM + signing + basic scanning, no SLSA L3, no Rekor, no OCI registry support". Users feel misled; security researchers write blog posts about the gap.

**Prevention:**
- Release notes include an **explicit "What's NOT included"** section
- SLSA level claimed must have a citation link to the self-assessment
- Link to this PITFALLS.md in release notes as "known limitations we chose"

**Phase:** Release (phase 6).

---

### Pitfall 16: Windows-specific supply chain gaps

**What goes wrong:** Signing/verification tested only on Linux. Windows path handling (backslashes in purls, case-insensitive file paths in SBOMs) produces different output across platforms. Signature files with CRLF line endings break verification.

**Prevention:**
- Every supply-chain command has a CI job for linux/darwin/windows
- SBOM paths normalized to forward-slash before serialization
- Signature I/O explicitly binary mode — never touched by text-mode newline conversion
- Golden master tests run on all three platforms

**Phase:** Polish (phase 3, platform parity) + Supply-chain (phase 4–5).

---

### Pitfall 17: Dependency on an unmaintained or security-sensitive library

**What goes wrong:** Pulling in a signing/SBOM library that's itself unmaintained, vendoring more third-party code, or adding a dependency that has its own CVE track record. Given omni's "pure Go stdlib-first" stance, every supply-chain dep is a big deal.

**Prevention:**
- Prefer `golang.org/x/crypto` and stdlib over third-party signing libs
- For SPDX/CycloneDX: write the serializer by hand (spec is not that large) rather than importing `spdx/tools-golang` if it adds transitive weight
- Document every new dep's maintenance status, last release, and contingency plan in an ADR
- Run `govulncheck` on omni's own `go.mod` as a release gate

**Phase:** Supply-chain (phase 4–5) — dependency decisions early, before coding.

---

## Phase-Specific Warnings

| Phase | Likely Pitfall | Mitigation |
|-------|---------------|------------|
| Phase 1 — Polish (cmderr finish) | Parallel work with supply-chain (Pitfall 13) | Serialize: finish cmderr before touching phase 4 |
| Phase 2 — Polish (coverage + goldens) | Golden tests baking in timestamps (Pitfall 14) | Normalize volatile fields in test harness now |
| Phase 3 — Polish (API audit) | `pkg/` surface too large (Pitfall 8); semver trap (Pitfall 7) | Move internals to `internal/`, write stability notes per package |
| Phase 4 — `omni sbom` + `omni sign`/`verify` | purl correctness (1); fail-open verify (2); key handling (3); SBOM completeness (12); cosign scope creep (9) | Fail-closed defaults; round-trip tests vs syft/cosign; key policy ADR before coding |
| Phase 5 — `omni scan` + `omni attest` | Vuln DB matching (4); SLSA overclaim (5); DB staleness (11) | OSV regression suite; honest SLSA level ADR; DB-age gates |
| Phase 6 — Release | Non-reproducible builds (6); announcement overclaim (15); Windows gaps (16); own supply chain (self-hosting gap) | Dual-build reproducibility check; explicit "not included" in release notes; all-platform CI; dogfood omni on its own release |

---

## Meta-Pitfall: The "Security Theater" Risk

The single biggest risk across this milestone is shipping tools that *look* like supply chain security but have enough gaps that users gain false confidence. For every command in phases 4–5, the acceptance criterion MUST include:

1. **What does this tool NOT protect against?** — explicit in `--help` and docs
2. **What's the failure mode?** — fail-closed, loud, non-zero exit
3. **Is there a negative test for every failure case?** — test matrix, not just happy path
4. **Does the tool tell the truth about its own limitations in output?** — e.g., "scanned 45/47 deps, 2 skipped due to missing purl"

If any of these is unclear at phase planning time, that phase needs deeper research before implementation.

---

## Sources

- SLSA v1.0 specification (slsa.dev/spec/v1.0) — HIGH confidence, spec-derived
- Sigstore bundle format spec + cosign documentation — HIGH confidence
- OSV schema documentation (ossf.github.io/osv-schema) — HIGH confidence
- Package URL (purl) spec (github.com/package-url/purl-spec) — HIGH confidence
- Go module reference and `golang.org/x/mod/semver` — HIGH confidence
- CycloneDX + SPDX specifications — HIGH confidence
- Reproducible-builds.org Go documentation — MEDIUM confidence
- `golang.org/x/vuln` / govulncheck design — HIGH confidence
- Historical post-mortems: SolarWinds, XZ-utils backdoor, npm typosquat incidents — general lessons, MEDIUM confidence
- omni's own `.planning/PROJECT.md` + `CONCERNS.md` for context-specific pitfalls — HIGH confidence
