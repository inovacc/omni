# ADR-0005: Pure-Go minisign-compatible signing scheme & cosign-compat scope

**Status:** Accepted
**Date:** 2026-06-03
**Decision:** Implement `omni sign` / `omni verify` / `omni sign keygen` as a pure-Go, fail-closed, minisign-compatible Ed25519 signing primitive in `pkg/sign/`, reimplemented on `crypto/ed25519` + `golang.org/x/crypto/{scrypt,blake2b}` rather than an external minisign library; default to the prehashed `"ED"` signature scheme (legacy `"Ed"` read-only); and scope Sigstore to **verification only**, gated behind a `//go:build omni_sigstore` build tag.

## Context

Phase 04 delivers the cryptographic foundation that every later supply-chain phase (sbom / scan / attest / release) reuses. omni's foundational constraints are non-negotiable: pure Go, no `os/exec`, no CGO, deterministic, cross-platform via build tags. A signing primitive must therefore:

1. Produce and verify detached signatures over arbitrary artifacts.
2. Interoperate with an existing, well-understood on-disk format so signatures and keys are inspectable and usable by third-party tooling.
3. Protect secret key material at rest (passphrase-derived encryption) and never leak it through logs, errors, flags, or panics.
4. Fail closed on verification — any parse error, key-ID mismatch, bad base64, short buffer, failed data signature, or failed global signature must reject; never accept on partial success.

Two ecosystem formats dominate: **minisign** (compact, Ed25519, single-file keys, human-inspectable text) and **Sigstore/cosign** (keyless, transparency-log backed, OCI-oriented, heavy dependency tree). The choice of default scheme, the build-vs-buy decision for the minisign codec, and the scope of Sigstore support drive the rest of the phase.

## Analysis

### Summary

| Field | Value |
|-------|-------|
| Default signing algorithm | Ed25519, prehashed (`"ED"` — `ed25519.Sign(sk, Blake2b512(data))`) |
| Legacy algorithm | `"Ed"` (raw `ed25519.Sign(sk, data)`) — read-only, never emitted by default |
| Key-at-rest KDF | scrypt (libsodium SENSITIVE: `N=1<<20`, `r=8`, `p=1`; `opslimit=33554432`, `memlimit=1073741824`) |
| Checksum | Blake2b-256 over `"Ed" \|\| key_id \|\| ed25519_secret` (wrong-passphrase detection) |
| On-disk format | minisign-compatible `.pub` / `.key` / `.minisig` (byte-exact) |
| Implementation | Reimplemented on stdlib + `golang.org/x/crypto` — NOT an external minisign library |
| Sigstore support | Verification only, behind `//go:build omni_sigstore`; default build returns `ErrUnsupported` |
| Crypto deps added | `golang.org/x/crypto/scrypt`, `golang.org/x/crypto/blake2b` (already a direct dep) |

### Decision 1 — Default scheme: minisign-compatible Ed25519, prehashed (`"ED"`)

minisign is a compact, audited, widely-deployed format built on Ed25519 with single-file keys and human-inspectable Base64 text files. It fits omni's pure-Go, deterministic, no-exec constraints exactly: Ed25519 is in `crypto/ed25519`, scrypt and Blake2b are in `golang.org/x/crypto` (a stdlib-adjacent, already-direct dependency).

The default emitted algorithm is the **prehashed** variant `"ED"` — `ed25519.Sign(sk, Blake2b512(data))`. Prehashing lets the signer stream large artifacts through Blake2b without buffering the whole payload, and matches modern minisign defaults. The legacy raw variant `"Ed"` (`ed25519.Sign(sk, data)`) remains **read-only**: `Verify` dispatches on the signature's algorithm bytes and accepts both, but `Sign` never emits `"Ed"`.

The byte-exact on-disk format (public key, secret key, signature, global signature, and the libsodium SENSITIVE scrypt constants) is specified authoritatively in the Phase 04 implementation plan (`docs/superpowers/plans/2026-06-03-phase-04-pkg-sign.md`, "Minisign on-disk format" section) and is implemented verbatim. In brief:

- **`*.pub`**: `untrusted comment:` line, then `base64("Ed" \|\| key_id[8] \|\| ed25519_pub[32])`.
- **`*.key`**: comment line, then `base64(sig_alg "Ed" \|\| kdf_alg "Sc" \|\| cksum_alg "B2" \|\| kdf_salt[32] \|\| opslimit[8 LE] \|\| memlimit[8 LE] \|\| encrypted_keynum_sk)`, where `keynum_sk = key_id[8] \|\| ed25519_secret[64] \|\| Blake2b-256("Ed" \|\| key_id \|\| ed25519_secret)`, XOR-encrypted with a scrypt-derived one-time stream.
- **`*.minisig`**: comment line; `base64(sig_alg[2] \|\| key_id[8] \|\| signature[64])`; `trusted comment:` line; `base64(global_sig[64])` where `global_sig = ed25519.Sign(sk, signature \|\| trusted_comment_bytes)`.

### Decision 2 — Reimplement on stdlib + x/crypto, not an external minisign library

| Option | Capabilities | Dependency cost | Verdict |
|--------|--------------|-----------------|---------|
| `jedisct1/go-minisign` | Verify only — **no keygen, no sign** | external dep | Insufficient (cannot keygen/sign) |
| `aead/minisign` | keygen + sign + verify | external dep, additional maintenance surface | Rejected (Pitfall 17: avoid unmaintained/extra deps for a foundational primitive) |
| Reimplement on `crypto/ed25519` + `golang.org/x/crypto` | full keygen + sign + verify, full control | only `scrypt` + `blake2b` (x/crypto already direct) | **Chosen** |

`jedisct1/go-minisign` is verify-only and cannot keygen or sign, so it cannot satisfy the phase. `aead/minisign` would work but adds an external dependency to omni's most foundational, security-critical primitive — directly contrary to omni's stdlib-first principle and Pitfall 17 (avoid taking on unmaintained or surplus dependencies for code we must fully own and audit). Reimplementing the codec on `crypto/ed25519` plus `golang.org/x/crypto/{scrypt,blake2b}` keeps full control over keygen + sign + verify, adds no external module beyond the already-direct `x/crypto`, and lets us audit every byte of the format. The format is small and fully specified, so reimplementation cost is bounded.

### Decision 3 — Sigstore = verification only, behind `//go:build omni_sigstore`

Sigstore/cosign interop is valuable for verifying third-party bundles, but the `sigstore-go` dependency tree is heavy (Rekor, go-openapi, Certificate Transparency, TSA, go-tuf). Compiling it into the default binary would inflate omni for every user, the overwhelming majority of whom never touch Sigstore — violating the "one lean static binary" core value.

Therefore:

- Sigstore support is **verification only** and is isolated behind `//go:build omni_sigstore`. The heavy dep tree is compiled **only** when that tag is set.
- The default build (`go build ./...`) ships a no-tag stub: `omni verify --bundle …` returns `cmderr.ErrUnsupported` ("requires building with `-tags omni_sigstore`"). CI builds **both** the tagged and untagged binaries to keep both paths honest.
- **v1.0 explicitly EXCLUDES Rekor upload, Fulcio issuance, and OCI/registry operations.** omni does not issue keyless certificates, upload to a transparency log, or push/pull OCI artifacts. Only local bundle *verification* (with a supplied trusted root) is in scope, and only behind the tag.

### Decision 4 — scrypt SENSITIVE cost (N=2²⁰) is intentional; tests use fixtures

Secret keys are encrypted at rest with the libsodium **SENSITIVE** scrypt profile (`opslimit=33554432`, `memlimit=1073741824` → `N=1<<20`, `r=8`, `p=1`). This is a deliberate at-rest protection choice: ~1 GiB of RAM and multiple seconds per derivation make offline brute-forcing of a stolen `.key` file expensive.

The cost has a direct testing implication: **live default-cost keygen must never run in automated tests** — it would consume ~1 GiB and stall CI. The `pkg/sign` API therefore exposes a functional option `WithScryptParams(n, r, p)` so every test can pass a low cost (e.g. `WithScryptParams(1<<15, 8, 1)`), and the golden harness verifies against committed low-cost fixtures rather than generating keys live (see Phase 04 plan, Task 10). The single test exercising the real SENSITIVE path is `-short`-skippable. The production default remains SENSITIVE.

## Decision

Implement `pkg/sign/` as a pure-Go, fail-closed, minisign-compatible Ed25519 signing primitive:

1. **Default scheme** — minisign-compatible Ed25519, prehashed (`"ED"`); legacy raw `"Ed"` read-only (verify-accepted, never emitted).
2. **Build, don't buy** — reimplement the `.pub`/`.key`/`.minisig` codec on `crypto/ed25519` + `golang.org/x/crypto/{scrypt,blake2b}`; no external minisign library.
3. **Sigstore = verification only, behind `//go:build omni_sigstore`**; v1.0 excludes Rekor upload, Fulcio issuance, and OCI; default build returns `cmderr.ErrUnsupported` for `--bundle`.
4. **scrypt SENSITIVE (N=2²⁰)** for keys at rest; tests use `WithScryptParams(low)` or committed fixtures, never live default-cost keygen.

Key-handling and secret-redaction policy (the `pkg/secret.Key` wrapper, file permissions, dev-vs-release key separation) is recorded separately in **ADR-0006**.

## Consequences

- **Pure-Go, no new heavy deps in the default binary.** Only `golang.org/x/crypto/{scrypt,blake2b}` are pulled in (x/crypto is already a direct dep). The sigstore-go tree is compiled only under `-tags omni_sigstore`.
- **Full ownership and auditability** of the keygen + sign + verify path; the format is implemented byte-exactly and is cross-checkable against the reference `minisign` tool.
- **Interoperability** with the minisign ecosystem: omni-produced `.minisig`/`.pub`/`.key` files are usable by `minisign`, and vice-versa.
- **Fail-closed verification** by construction — key-ID checked before any crypto, both the data signature and the trusted-comment global signature must verify, and every error path rejects; the CLI maps verification failures to `cmderr.ErrConflict` (exit 1).
- **CI must build both** the default and `-tags omni_sigstore` binaries, and test the untagged `ErrUnsupported` path, so neither path silently rots.
- **Test cost is controlled** via `WithScryptParams` and committed fixtures; the production SENSITIVE cost is preserved and exercised by a single `-short`-skippable test.
- **Future supply-chain phases** (sbom / scan / attest / release) build on `pkg/sign` as their signing foundation without re-deciding the scheme.
