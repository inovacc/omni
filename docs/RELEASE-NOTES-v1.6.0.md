# omni v1.0

> **Honesty first.** This release ships real, dogfooded supply-chain tooling at the level omni actually achieves — and says plainly what it does **not** do. No SLSA-L3 claim, no "enterprise supply-chain platform" claim.

## Audience & scope

omni is built **for me and my CI/CD pipelines**. Broader open-source adoption is welcome but is **not** the design driver. Design decisions optimize for deterministic, dependency-light CI use on every OS — not for being a general-purpose security suite. If a choice helped "a general user" but cost determinism or added a heavy dependency, omni chose determinism.

## What v1.0 protects

Every release archive is produced by the omni binary built in the same run (dogfooded), and is independently verifiable offline:

- **Signed binaries** — minisign-compatible Ed25519 signatures via `omni sign` (key handling per ADR-0006; passphrase only via `OMNI_SIGN_PASSPHRASE`, never a flag).
- **Per-archive SBOMs** — byte-deterministic SPDX 2.3 / CycloneDX 1.5 via `omni sbom` (pure-stdlib emitter; ADR-0007).
- **SLSA Build L2 provenance** — in-toto/DSSE attestations via `omni attest`, `builder.id` = the release workflow (ADR-0009). The claimed level is enforced in code by a `builder.id` allowlist; there is no flag to claim L3.
- **Reproducible builds** — `CGO_ENABLED=0 -trimpath -buildvcs=true` with a pinned timestamp; a CI **dual-build gate** recompiles each target and `omni reprocheck` fails the release on any sha256 drift (ADR-0010). Reproducibility is what makes the SBOM and signature mean the same thing on your machine.
- **Vulnerability scanning** — `omni scan` matches an SBOM against a `pkg/sign`-signed OSV database, gates CI via `--fail-on <severity>`, and fails loudly on a stale or tampered DB (ADR-0008).

All of the above is **pure-Go, no-CGO, no external processes** inside the omni binary, on Linux, macOS, and Windows (amd64 + arm64).

## What's NOT protected against

See `.planning-archive/research/PITFALLS.md` for the full analysis. In short:

- **NOT SLSA L3.** omni does not run on a hermetic/isolated build service with non-falsifiable provenance. The honest level is **Build L2** (Pitfall 5). omni refuses to emit any other `builder.id`.
- **No Rekor / Fulcio / OCI.** There is no transparency-log upload, no Fulcio/OIDC certificate issuance, and no OCI registry push. Sigstore support is **verification-only**, opt-in behind `-tags omni_sigstore`, and absent from release binaries (Pitfall 9).
- **Reachability scanning is not in the release binaries.** `omni scan source` (call-graph reachability) is deferred from v1.0 — it requires `golang.org/x/vuln`, which execs `go list` and would bloat `go.mod` via MVS. It returns `unsupported` in these binaries; its future home is a separate `contrib/govulncheck-scan` module.
- **The OSV DB is only as fresh as your last `omni scan db update`.** Offline scanning against a stale DB misses newly-disclosed vulnerabilities; `--max-db-age` makes staleness fail loudly rather than silently (Pitfall 11).
- **SBOM transitive completeness is bounded by `debug/buildinfo`.** Binary SBOMs reflect what the Go toolchain recorded, not a full source-level dependency resolution (Pitfall 12).
- **Cross-tool attestation interop:** omni's DSSE signature is a minisign blob, verifiable by `omni attest verify` but **not** by a generic cosign/sigstore verifier (ADR-0009).

**This is not a turnkey enterprise supply-chain platform.** It is a focused, honest, dependency-light toolkit that does a few things deterministically and verifiably.

## Verify these artifacts yourself

The release publishes the public key `omni.pub`. For any archive:

```bash
# 1. Verify the signature (fail-closed).
omni verify --key omni.pub --sig omni_<OS>_<arch>.tar.gz.minisig omni_<OS>_<arch>.tar.gz

# 2. Verify the SLSA provenance binds to the archive (fail-closed).
omni attest verify --key omni.pub --artifact omni_<OS>_<arch>.tar.gz omni_<OS>_<arch>.tar.gz.intoto.jsonl

# 3. Inspect the SBOM (every component carries a normalized Go purl).
omni cat omni_<OS>_<arch>.spdx.json | omni jq '.packages[].externalRefs'

# 4. (Optional) scan the SBOM against a signed OSV DB.
omni scan db update --url <db-url> --db-key omni.pub
omni scan omni_<OS>_<arch>.spdx.json --db osv-db.zip --db-key omni.pub --fail-on high
```

Each verification command exits non-zero on any failure — a wrong key, a tampered archive, a mismatched digest, or an unsigned/stale input. If a command exits 0, the claim it checks holds.
