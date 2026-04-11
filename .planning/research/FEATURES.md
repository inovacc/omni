# Feature Landscape

**Domain:** Go-native supply-chain tooling (SBOM, signing, vuln scanning, SLSA attestation) for a single-binary CLI used in CI/CD
**Researched:** 2026-04-11
**Audience bias:** "me + my CI/CD pipelines" — not an OCI registry operator, not a platform team, not a GUI user

---

## Scope Note

Four capability areas, each mapped against the reference tool in that space:

| Area | Reference tool | omni command |
|------|----------------|--------------|
| SBOM generation | Anchore `syft` | `omni sbom` |
| Artifact signing | Sigstore `cosign` (or `minisign`) | `omni sign` / `omni verify` |
| Vulnerability scan | Anchore `grype` / Google `osv-scanner` | `omni scan` |
| Provenance / attestation | `slsa-github-generator`, `in-toto`, `cosign attest` | `omni attest` |

All four must remain pure-Go, no-exec, one-shot, and usable as `pkg/` libraries. That constraint alone rules out a huge chunk of what `cosign` and `syft` do.

---

## Table Stakes

Features a CI pipeline user will walk away if missing. Each one has to land for the capability to be "useful" rather than "cute."

### SBOM (`omni sbom`)

| Feature | Why expected | Complexity | CI relevance |
|---------|--------------|------------|--------------|
| Generate SBOM from a Go module directory (`go.mod` + `go.sum`) | Primary input for a Go-native tool; `syft` has this | Low | HIGH — every Go build step |
| SPDX 2.3 JSON output | De facto standard; required by many compliance regimes | Medium | HIGH |
| CycloneDX 1.5 JSON output | Second de facto standard; OWASP ecosystem | Medium | HIGH |
| Generate SBOM from a built binary (parse `debug/buildinfo`) | Stdlib support exists; unique Go-specific capability | Low | HIGH — verify what shipped matches what built |
| Component list with purl (`pkg:golang/...`) identifiers | Required for downstream correlation with vuln scanners | Low | HIGH |
| License field extraction (best-effort from module metadata) | Compliance teams ask for it; often the #1 reason SBOMs are generated | Medium | MEDIUM |
| Deterministic output (stable ordering, no timestamps by default) | Required for reproducible builds + SBOM diffing | Low | HIGH — otherwise CI diffs are noise |
| Write to file or stdout | Both patterns show up in CI (artifact upload vs pipe) | Low | HIGH |

### Signing (`omni sign` / `omni verify`)

| Feature | Why expected | Complexity | CI relevance |
|---------|--------------|------------|--------------|
| Sign a file (blob) with a local private key | Minimum viable signing path | Low | HIGH |
| Verify a signature against a public key | Mirror of sign; useless without it | Low | HIGH |
| Ed25519 as default algorithm | Pure-Go, fast, small, stdlib support | Low | HIGH |
| Detached signature files (`.sig`) | Standard pattern; matches `cosign sign-blob` output shape | Low | HIGH |
| Key generation (`omni sign keygen`) | Users need a way to get a key pair | Low | HIGH |
| Passphrase-protected private keys (scrypt or argon2id KDF) | Bare private keys on disk are a security smell | Medium | HIGH |
| Read key material from env var or file | CI secret injection pattern — mandatory | Low | HIGH |
| Machine-readable verify exit codes (0 = valid, nonzero = invalid/missing) | Scriptability in pipelines | Low | HIGH |

### Vulnerability scan (`omni scan`)

| Feature | Why expected | Complexity | CI relevance |
|---------|--------------|------------|--------------|
| Scan a Go module or binary against OSV | OSV is the canonical, pure-data, vendor-neutral source | Medium | HIGH |
| Accept an SBOM as input (SPDX or CycloneDX) | Decouples scanning from language-specific discovery | Medium | HIGH |
| Report format: JSON (machine) + text table (human) | Both needed for CI + local use | Low | HIGH |
| Severity filter (`--fail-on high`) | The #1 CI gating pattern | Low | HIGH |
| Exit code reflects findings threshold | Required for pipeline break semantics | Low | HIGH |
| Offline mode with a pre-fetched OSV database | CI runners often cannot egress to OSV | Medium | HIGH |
| Suppress/ignore list by CVE ID or purl | Every real-world user has accepted risks to suppress | Medium | HIGH |
| `govulncheck`-style reachability (callsite analysis) | Go-specific, reduces false positives dramatically; `govulncheck` already does this pure-Go | High | HIGH — differentiator-adjacent but expected by Go users |

### Attestation (`omni attest`)

| Feature | Why expected | Complexity | CI relevance |
|---------|--------------|------------|--------------|
| Generate an in-toto statement wrapping an SBOM or provenance predicate | in-toto is the SLSA data model | Medium | HIGH |
| SLSA v1.0 provenance predicate generation | Current spec level; v0.2 is legacy | Medium | HIGH |
| Sign the attestation (reuses `omni sign`) | Attestation without signature = plain JSON | Low | HIGH |
| Verify an attestation: signature + predicate shape + subject hash match | The whole point of attestations is to verify them later | Medium | HIGH |
| Read build metadata from GitHub Actions env vars | Where CI runs today; `GITHUB_*` env is the de facto source | Low | HIGH |
| DSSE envelope format | Sigstore/in-toto standard; tools downstream expect it | Medium | HIGH |

---

## Differentiators

Where omni can credibly beat "just shell out to syft + cosign + grype" for the target user. Each one maps to an omni-specific advantage (single binary, pure Go, no-exec, stdlib-first, library-first).

### Cross-capability differentiators (apply to all four)

| Feature | Value proposition | Complexity | CI relevance |
|---------|-------------------|------------|--------------|
| **One binary, zero dependencies** | Replace 4 tool installs + 4 version pins with one `omni` pin | Baseline | HIGH — the core thesis |
| **Pure-Go library API in `pkg/`** | Import `pkg/sbom`, `pkg/sign`, `pkg/scan`, `pkg/attest` from other Go projects without shelling out to binaries | Medium | HIGH — unique vs syft/cosign/grype which are binary-first |
| **Deterministic output across all commands** | Stable ordering, no wall-clock in artifacts → reproducible SBOMs and attestations | Medium | HIGH |
| **Unified `cmderr` exit codes** | `scan` returns `ErrConflict` on findings, `verify` returns `ErrConflict` on bad sig, both classifiable the same way in CI | Low | HIGH |
| **Works offline by default** (with a fetched DB/keyring) | CI runners with no egress are first-class | Medium | HIGH |
| **Chainable via `omni pipe` / `omni pipeline`** | `omni sbom ./ \| omni scan --from-sbom - \| omni attest --predicate -` in a single process tree | Medium | HIGH — unique to omni's existing pipe model |

### SBOM differentiators

| Feature | Value proposition | Complexity | CI relevance |
|---------|-------------------|------------|--------------|
| SBOM from `debug/buildinfo` of a built binary | Verify what you shipped, not what you thought you shipped. syft does this but less precisely for Go | Low | HIGH |
| SBOM diff (`omni sbom diff old.json new.json`) | Answers "what dependencies changed this PR" directly | Medium | HIGH — reuses `pkg/twig/comparer` patterns |
| VEX-compatible annotations on components | Ties directly into `omni scan` suppression | Medium | MEDIUM |
| Vendored-dep awareness (flag vendored packages separately) | Real Go projects have vendored buf/protobuf etc.; `syft` handles this poorly | Medium | MEDIUM |

### Signing differentiators

| Feature | Value proposition | Complexity | CI relevance |
|---------|-------------------|------------|--------------|
| Minisign-compatible signature format | Wide ecosystem compat without dragging in Sigstore transparency log | Medium | HIGH |
| `cosign`-compatible signature format for **blobs only** (not OCI) | Users who already consume `cosign verify-blob` can swap in `omni verify` | High | MEDIUM |
| Sign an entire directory tree (Merkle root via `pkg/twig`) | Unique to omni — leverages existing tree hashing | Medium | MEDIUM |
| Timestamped signatures via RFC 3161 TSA (optional, pure-Go client) | Non-repudiation without Rekor/transparency-log complexity | High | LOW-MEDIUM |

### Scan differentiators

| Feature | Value proposition | Complexity | CI relevance |
|---------|-------------------|------------|--------------|
| Built-in OSV database fetcher + cache in XDG dirs | No need for separate `grype db update` | Medium | HIGH |
| Reachability analysis via `golang.org/x/vuln` (govulncheck lib) | Pure-Go, dramatically cuts false positives on Go code | High | HIGH |
| Scan directly from `debug/buildinfo` of a binary | Scan the artifact itself, not just the repo | Low | HIGH |
| Streaming NDJSON output | Matches existing `omni rg --json-stream` pattern | Low | MEDIUM |
| `--fail-on` + `--ignore-until <date>` | Temporal suppressions are a real-world need; grype has them, but awkwardly | Medium | HIGH |

### Attest differentiators

| Feature | Value proposition | Complexity | CI relevance |
|---------|-------------------|------------|--------------|
| GitHub Actions builder detection with zero config | `omni attest` in an Actions step "just works" | Low | HIGH |
| Attest + sign + verify round-trip in a single command | `omni attest --sign --verify` for local dev confidence | Medium | MEDIUM |
| Store attestations as sidecar files (not OCI refs) | Filesystem-native; no registry required | Low | HIGH |
| Verify an attestation bundle without network access | No Rekor lookup, no Fulcio CA fetch by default | Medium | HIGH |

---

## Anti-Features

Deliberately out of scope. Each one either violates a design principle (no-exec, pure-Go, one-shot, single-binary) or serves a user that isn't the target audience.

| Anti-feature | Why avoid | What to do instead |
|--------------|-----------|--------------------|
| **OCI registry push/pull for signatures** (`cosign sign <image>`) | Requires a registry client, auth chains, OCI media-type handling; not the CI use case for a Go CLI | Sign blobs + binaries only. Users who need OCI signing keep using `cosign` for that step |
| **Fulcio / keyless signing with OIDC** | Requires live network to Sigstore PKI; violates offline-first; huge surface area | Local keys + (optional) RFC 3161 TSA if non-repudiation is needed |
| **Rekor transparency log submission & lookup** | Network-dependent, centralized, complex client; not needed for "CI pipeline verifies its own artifacts" | Detached signatures + DSSE envelopes are sufficient |
| **Container image scanning** (extracting layers, scanning OS packages) | Would require OCI image parsing, apk/dpkg/rpm format handling — massive scope | `omni scan` scans Go modules/binaries/SBOMs. Container OS scanning stays with `grype` |
| **GUI dashboard / web UI / report server** | Long-running process; violates one-shot design | JSON output → user pipes into their existing dashboard |
| **Hosted vulnerability DB service** | Not a product surface; not a service business | Fetch OSV data locally, cache in XDG dirs |
| **Plugin system for custom extractors** | Explicit out-of-scope in PROJECT.md | Build extractors into the binary or expose `pkg/` for library users |
| **Daemon mode / watch mode for continuous scanning** | Long-running; out-of-scope in PROJECT.md | One-shot scans called by cron or CI on a schedule |
| **Policy-as-code engine** (OPA / Rego / cue) | Huge dep surface; CI users already have OPA if they want it | Simple built-in filters (`--fail-on`, `--ignore`, severity thresholds) |
| **Full cosign CLI parity** | PROJECT.md explicitly marks this as possibly too large for v1.0 | Blob signing only; OCI stays out |
| **Automatic PR comments / GitHub check annotations** | Tool-vs-integration boundary; Actions users compose this themselves | Emit clean JSON; let a separate Action post it |
| **SLSA v0.2 provenance** (legacy schema) | v1.0 is the current spec; supporting both doubles test surface | v1.0 only; document the choice |
| **SBOM for non-Go ecosystems** (npm, pip, cargo, maven) | Massive scope for a tool whose audience is Go CI/CD | Go only. Users with polyglot repos keep `syft` for the other languages |
| **Signed CI runner attestations via TPM / TEE** | Hardware-attestation rabbit hole | Software-only provenance from env vars |

---

## Feature Dependencies

```
omni sign (keygen, sign, verify)
    ├── used by → omni attest (attestation signing)
    ├── used by → omni sbom (optional: sign SBOM output)
    └── used by → CI release step (sign release binaries)

omni sbom
    ├── used by → omni scan (--from-sbom input)
    └── used by → omni attest (SBOM-as-predicate in in-toto statement)

omni scan
    ├── depends on → OSV database fetcher/cache
    └── optional input → omni sbom output

omni attest
    ├── depends on → omni sign (DSSE envelope signing)
    ├── depends on → omni sbom (SBOM predicate) OR raw provenance predicate
    └── depends on → CI env var reader (GitHub Actions metadata)
```

**Build order implication:** `sign` must land first (keygen + sign + verify as a unit). `sbom` and `scan` can proceed in parallel after `sign` exists as a dependency. `attest` is last because it composes all three.

---

## MVP Recommendation

The minimum viable supply-chain slice that delivers end-to-end value in a CI pipeline:

1. **`omni sign keygen` + `omni sign` + `omni verify`** — Ed25519, passphrase-protected keys, detached `.sig` files, env-var key loading
2. **`omni sbom`** — Go module + Go binary input, SPDX 2.3 JSON + CycloneDX 1.5 JSON output, deterministic ordering, purl identifiers
3. **`omni scan`** — OSV database, accepts SBOM input, `--fail-on <severity>` gating, JSON + text output, offline-capable
4. **`omni attest`** — in-toto + SLSA v1.0 provenance, DSSE envelope, signs via `omni sign`, reads GitHub Actions env vars

**Defer (but keep architecturally possible):**
- Reachability analysis via `golang.org/x/vuln` — valuable but high-complexity; ship a basic OSV match first, add reachability in a follow-up
- Merkle-tree signing of directory trees — slick demo of omni's integration story, but not on the critical path
- RFC 3161 TSA client — non-repudiation is a nice-to-have, not a CI-blocker
- `cosign`-compatible blob signature format — compat layer is polish, not MVP
- SBOM diff — useful but a separate feature from core generation

**Explicitly not in MVP (see Anti-Features):** OCI signing, Fulcio, Rekor, container scanning, policy engines, non-Go SBOM ecosystems.

---

## CI Relevance Summary

Every table-stakes and differentiator feature was rated against "does this matter for `me + my CI/CD pipeline`?":

- **HIGH**: Must work for the CI use case. Missing = pipeline broken or user walks away. Most table-stakes features and ~half of differentiators.
- **MEDIUM**: Nice for CI, mainly valuable for local dev or polish. VEX annotations, signature format compat, streaming NDJSON.
- **LOW-MEDIUM**: Exists because it's technically possible in pure Go, not because CI needs it today. RFC 3161 TSA is the main example.

Anti-features were selected specifically on CI-irrelevance + design-principle violations, not on "too hard."

---

## Sources

- Anchore syft capability reference (training data, HIGH confidence for documented features, LOW for version-specific flags)
- Sigstore cosign documentation (training data, HIGH for blob-signing model, LOW for current keyless flow specifics)
- OSV schema and osv-scanner (training data, HIGH for schema, MEDIUM for scanner behavior)
- SLSA v1.0 provenance specification (training data, MEDIUM)
- in-toto attestation framework (training data, MEDIUM)
- Go `golang.org/x/vuln` / govulncheck (training data, HIGH — stdlib-adjacent)
- Go `debug/buildinfo` stdlib package (HIGH — directly verified via Go stdlib knowledge)
- PROJECT.md (root of truth for audience + out-of-scope)

**Confidence note:** Feature categorization into table-stakes vs differentiator vs anti-feature is opinionated and driven by PROJECT.md's "me + CI/CD pipelines" audience statement. A different audience (platform team, OSS-first tool) would produce a different split. Flagging for review during milestone planning.
