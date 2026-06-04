# ADR-0009: Honest SLSA level & builder.id policy

**Status:** Accepted
**Date:** 2026-06-03
**Decision:** `omni attest` emits DSSE-wrapped in-toto Statement v1 / SLSA Provenance v1 attestations claiming **only SLSA Build L2** — the honest level omni's GitHub-Actions+OIDC release process actually achieves. The claimed level is conveyed solely by an **ADR-pinned `builder.id` allowlist** (no numeric `slsaLevel` field is emitted); the generator refuses any `builder.id` outside the allowlist (`cmderr.ErrInvalidInput`). There is no flag, code path, or option that can claim L3. The attestation format and signing are pure-Go (stdlib + `pkg/sign`); no in-toto/sigstore SDK enters `go.mod`.

## Context

Phase 07 ships `omni attest` / `omni attest verify`. SLSA provenance is a trust claim; over-claiming the build level (Pitfall 5) is a supply-chain integrity failure worse than claiming nothing. omni's releases run on GitHub Actions with OIDC-identified runners that generate and sign provenance — this is **Build L2** (a hosted build platform generates and signs provenance). It is **not L3**: omni does not run on an isolated/hermetic build service with non-forgeable provenance and strong tenant isolation. Local (developer-machine) generation is weaker still — effectively L1 (unverified).

A second integrity requirement (ADR-0007 lesson) frames the implementation: the official in-toto-golang / sigstore SDKs drag the Rekor / go-tuf / protobuf-specs tree into `go.mod` via MVS even when only referenced behind a tag. The format must therefore be a pure-Go reimplementation.

This ADR is the **hard gate** for Phase 07 — it pins the honesty boundary before any generator code lands. The spec makes it non-negotiable: "no provenance output may claim a higher level than the ADR."

## Analysis

| Decision | Choice | Rationale | Rejected alternative |
|----------|--------|-----------|----------------------|
| **Claimed level** | **SLSA Build L2** — and only L2. | GitHub Actions + OIDC generates+signs provenance on a hosted platform = L2 exactly. | Claim L3 — rejected: omni has no isolated/hermetic builder with non-forgeable provenance; L3 would be a false trust claim (Pitfall 5). Claim a numeric `slsaLevel` field — rejected: SLSA v1.0 implies level via `builder.id`, not a self-asserted number. |
| **How level is conveyed** | An **ADR-pinned `builder.id` allowlist** in `pkg/attest`. Release path: `https://github.com/inovacc/omni/.github/workflows/release.yml@refs/heads/main`. Local/non-CI fallback: `https://github.com/inovacc/omni/attest/local@v1` (a LOWER, unverified tier). No `slsaLevel` field is emitted. | The verifiable identity of the builder IS the level claim under SLSA v1.0. An allowlist makes over-claim a code-level impossibility, not a convention. | A free-form `--builder-id` accepting any value — rejected: lets a caller forge a release-tier id from a local run. |
| **`buildType`** | `https://slsa-framework.github.io/github-actions-buildtypes/workflow/v1` when `--from-env` detects GitHub Actions; otherwise `https://github.com/inovacc/omni/attest/local-buildtype/v1`. | The buildType must match the actual build mechanism; the local type signals the weaker tier. | A single buildType regardless of environment — rejected: misrepresents local runs as CI. |
| **Enforcement** | The generator rejects any `builder.id` not in the allowlist set with `cmderr.ErrInvalidInput`; `--from-env` selects the release id ONLY when the GitHub-Actions env is present, else the local id. | Honesty is enforced in code, not left to operator discipline. | Trust the operator to pass the right id — rejected: discipline is not a control. |
| **Format & deps** | Pure-Go in-toto Statement v1 + SLSA Provenance v1 + DSSE PAE on `encoding/json`/`base64`/`crypto/sha256`; signed via `pkg/sign` (Ed25519, reused from Phase 04). | Zero new third-party deps in the default build (ADR-0007 MVS rule); reuses the existing key infrastructure (`omni sign keygen`). | in-toto-golang / sigstore SDK — rejected: pulls Rekor/go-tuf/protobuf-specs into `go.mod` via MVS. |
| **Schema validation** | A CI-only `task` target validates every emitted predicate against the official SLSA Provenance v1 JSON schema before release; the binary itself does not embed a validator. | Catches drift from the spec without bloating the shipped binary (mirrors the SBOM `syft` oracle, ADR-0007). | Embed a JSON-schema validator in the binary — rejected: heavy dep for a CI-only check. |

## Consequences

- **No over-claim is possible.** The only way to assert the release tier is an ADR-listed `builder.id`, and the generator refuses anything else. Local attestations carry the local builder.id/buildType (weaker tier), never the release one.
- **Honest by omission.** No `slsaLevel`/`completeness` field is emitted; the level is read from `builder.id` per SLSA v1.0. The honesty contract is "emit only an ADR-listed builder.id."
- **Default build stays lean & pure-Go.** `pkg/attest` adds zero third-party deps; signing reuses `pkg/sign`. No exec.
- **Fail-closed verify.** `omni attest verify` returns a classified `cmderr` error on every failure mode (bad signature, wrong key, tampered payload, malformed envelope, unknown builder.id), never a silent pass.
- **Dogfoodable in Phase 08.** The release pipeline calls `omni attest --from-env` so the v1.0 release's provenance is produced by omni itself at the honest L2 tier.
- **Changing the claimed level is a deliberate, reviewable edit** to this ADR + the `pkg/attest` allowlist — never a runtime flag.
- **Interop limitation (v1.0): omni-signed attestations are self-consistent, NOT cosign/sigstore-interoperable.** The DSSE `signatures[].sig` carries omni's minisign signature blob (Ed25519 with an internal Blake2b-512 prehash), not a bare Ed25519 signature over the PAE. `omni attest verify` re-derives the same PAE and verifies via `pkg/sign`, so omni's envelopes round-trip correctly; but a generic cosign/sigstore verifier (which expects a raw Ed25519 signature) will not validate them. This is an accepted v1.0 trade-off (reuse the audited `pkg/sign` primitive, zero new deps). A future phase may add a raw-Ed25519 DSSE signer behind a flag for cross-tool interop.
