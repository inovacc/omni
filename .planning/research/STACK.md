# Technology Stack — Supply Chain Capability Track

**Project:** omni (v1.0 supply-chain track)
**Researched:** 2026-04-11
**Scope:** Pure-Go libraries for `omni sbom`, `omni sign`/`verify`, `omni scan`, `omni attest`
**Overall confidence:** HIGH on format libs & govulncheck; MEDIUM on osv-scanner & sigstore-go (heavy deps); HIGH on minisign fallback

---

## Recommended Stack

### `omni sbom` — SBOM Generation

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| `github.com/CycloneDX/cyclonedx-go` | v0.9.x (Jan 2026) | CycloneDX 1.6 serializer/deserializer (JSON + XML) | Official OWASP library, pure Go, zero deps beyond stdlib, tiny surface. Produce & consume. |
| `github.com/spdx/tools-golang` | v0.5.x (Jan 2026) | SPDX 2.2 / 2.3 / 3.0 serializer/deserializer | Official SPDX reference Go library, Apache-2.0, pure Go. |
| `golang.org/x/mod/modfile` + `golang.org/x/mod/sumdb` | latest (matches Go 1.25) | Parse `go.mod` / `go.sum`, resolve module graph | Already stdlib-adjacent; zero risk. Same code Go toolchain uses. |
| `debug/buildinfo` (stdlib) | Go 1.25 | Extract embedded module info from Go binaries | Native. Matches `go version -m`. Perfect for binary SBOMs. |

**Decision:** Build `pkg/sbom/` as a thin generator on top of `golang.org/x/mod` + `debug/buildinfo`, emitting via cyclonedx-go and spdx/tools-golang. **Do NOT import `github.com/anchore/syft` as a library** — see anti-stack.

**Confidence:** HIGH. Both format libraries are stable, pure-Go, and maintained by the upstream spec owners.

---

### `omni sign` / `omni verify` — Artifact Signing

**Two-tier strategy** (matches PROJECT.md "minisign-style acceptable as minimum"):

#### Tier 1 — Default: Minisign (pure Go, zero deps)

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| `aead.dev/minisign` | v0.3.x | Ed25519 minisign signing + verification | Pure Go, stdlib-only (`crypto/ed25519`, `golang.org/x/crypto/blake2b`). Signs, verifies, and generates keys. Interop-compatible with `jedisct1/minisign` CLI. |

**Why default:** Zero OCI/registry dependencies, zero Rekor/Fulcio network calls, works offline in CI, tiny dependency footprint. Key pairs = two files. This is the "CI/CD pipelines first" answer.

#### Tier 2 — Opt-in: Sigstore bundle verification

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| `github.com/sigstore/sigstore-go` | v0.7.x | Verify Sigstore protobuf-bundle signatures (keyless, Fulcio certs, Rekor entries) | Officially the successor Go API. Upstream guidance (sigstore docs): "use sigstore-go for verification, not cosign". Minimal dep tree vs `sigstore/cosign`. |
| `github.com/sigstore/protobuf-specs` | pinned by sigstore-go | Bundle format types | Transitive, expected. |

**Why opt-in only:** sigstore-go still pulls a non-trivial transitive graph (TUF client, x509 chains, TSA). Gate behind a build tag (`//go:build omni_sigstore`) so the default binary stays lean. Signing (keyless OIDC flow) is explicitly excluded from v1.0 per PROJECT.md — sigstore-go does verification only in the default path.

**Confidence:**
- Minisign: HIGH (tiny, verified, stdlib-based)
- sigstore-go: MEDIUM (stable per upstream, but transitive dep cost warrants build-tag gating)

---

### `omni scan` — Vulnerability Scanning

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| `golang.org/x/vuln/scan` | v1.1.x+ (tracks Go release) | **Primary** — govulncheck programmatic API for Go code/binaries | Stable public API since v1.0. Go team-maintained. Uses Go vuln DB (OSV-formatted). Pure Go. Can scan source (`./...`) AND compiled binaries (`-mode=binary`). |
| `github.com/google/osv-scanner/v2/pkg/osvscanner` | v2.x (Mar 2026) | **Secondary** — multi-ecosystem lockfile scanning (npm, pypi, cargo, etc.) | Covers non-Go lockfiles if omni users want multi-lang scans. Pure Go, Google-maintained. |
| OSV database offline bundle | `osv.dev` zipball | Air-gapped scanning option | Downloadable, cacheable in `~/.cache/omni/osv/`. |

**Decision:** `omni scan` defaults to `govulncheck` semantics for Go projects (the primary use case), with `--ecosystem multi` flag opting into osv-scanner v2 for polyglot scans. Gate osv-scanner v2 behind a build tag as well (`//go:build omni_osv`) because of its dep weight.

**Confidence:**
- govulncheck API: HIGH (stable, official)
- osv-scanner v2: MEDIUM (stable API but dep-heavy; build-tag gate mitigates)

---

### `omni attest` — SLSA Provenance

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| `github.com/in-toto/in-toto-golang` | v0.9.x | Generate + verify in-toto statements and SLSA v1.0 provenance predicates | Reference implementation. Pure Go. Types for SLSA 0.1 / 0.2 / 1.0. |
| `github.com/in-toto/attestation` (Go predicates) | latest | Canonical predicate type definitions (provenance v1, VSA, SCAI) | Authoritative predicate schemas, small package. |
| `github.com/secure-systems-lab/go-securesystemslib` | v0.9.x | DSSE envelope signing/verification | Standard envelope format for in-toto; pure Go. Pairs with minisign or sigstore-go for the signing step. |

**Decision:** `pkg/attest/` builds SLSA v1.0 provenance from omni's own build context (env, VCS info, materials) and wraps it in a DSSE envelope signed via Tier 1 minisign by default. **Do NOT pull in `slsa-framework/slsa-github-generator`** — it's a GitHub Action generator, not a library; CGO-adjacent and GitHub-coupled.

**Confidence:** HIGH on in-toto-golang and go-securesystemslib; both are reference implementations.

---

## Alternatives Considered (Anti-Stack — DO NOT USE)

| Rejected | Category | Why Not |
|----------|----------|---------|
| `github.com/anchore/syft` (as library) | SBOM | Massive dep graph (stereoscope, OCI registry clients, rekor clients), imports ~400 transitive modules. Would bloat `omni` binary by 40–60 MB. Syft's cataloger architecture is valuable but the *library entry points* are not designed for lean embedding. Reimplement their Go-module cataloger pattern instead. |
| `github.com/anchore/grype` (as library) | Scan | Same bloat profile as syft; depends on syft. govulncheck + osv-scanner v2 cover the same ground. |
| `github.com/sigstore/cosign` (as library) | Sign | Upstream Sigstore docs explicitly say: **"Cosign is not recommended for integration"** — pulls huge dep tree, lacks protobuf bundle format. Use sigstore-go. |
| `github.com/jedisct1/go-minisign` | Sign | Verify-only. aead.dev/minisign is a superset (sign + verify + keygen) and equally pure-Go. |
| `github.com/slsa-framework/slsa-github-generator` | Attest | GitHub-Action-shaped, not a clean library. Coupled to GH workflow tokens. |
| `github.com/CycloneDX/cyclonedx-gomod` | SBOM | It's a CLI, not intended as an embedded library. Copy its approach, import only cyclonedx-go. |
| Any library that transitively imports `github.com/mattn/go-sqlite3` | All | CGO. omni already uses `modernc.org/sqlite` (pure-Go). Hard block. |
| Any library importing `github.com/containerd/containerd` | SBOM/Sign | CGO on some platforms, huge surface. Container-image SBOM is out of scope for v1.0 anyway. |

---

## Duplication Check vs Existing omni Deps

Already-in-tree deps that cover part of this work (from `.planning/codebase/STACK.md`):

| Existing dep | Already covers |
|---|---|
| `golang.org/x/crypto` v0.48.0 | blake2b for minisign; ed25519 via stdlib |
| `modernc.org/sqlite` v1.44.3 | Pure-Go SQLite for local OSV cache (no new DB dep needed) |
| `go.etcd.io/bbolt` v1.4.3 | Alternative KV store for vuln cache if SQLite overkill |
| `gopkg.in/yaml.v3` v3.0.1 | SPDX YAML if we ever need it (rare) |
| `github.com/BurntSushi/toml` v1.6.0 | Config for sign/attest key paths |

**Net new direct dependencies required:** 7
1. `github.com/CycloneDX/cyclonedx-go`
2. `github.com/spdx/tools-golang`
3. `aead.dev/minisign`
4. `golang.org/x/vuln` (scan subpackage)
5. `github.com/in-toto/in-toto-golang`
6. `github.com/in-toto/attestation`
7. `github.com/secure-systems-lab/go-securesystemslib`

**Build-tag-gated (not in default binary):**
- `github.com/sigstore/sigstore-go` (tag: `omni_sigstore`)
- `github.com/google/osv-scanner/v2` (tag: `omni_osv`)

This keeps the default `omni` binary lean while making heavyweight verification paths available for users who opt in.

---

## Pure-Go / CGO Verification

| Library | Pure Go? | Notes |
|---------|----------|-------|
| cyclonedx-go | YES | stdlib only |
| spdx/tools-golang | YES | stdlib only |
| aead.dev/minisign | YES | uses `crypto/ed25519` + `x/crypto/blake2b` |
| golang.org/x/vuln | YES | Go team; builds with CGO_ENABLED=0 |
| osv-scanner/v2 | YES | confirmed in upstream release notes; builds static |
| in-toto-golang | YES | stdlib + x/crypto |
| go-securesystemslib | YES | stdlib crypto |
| sigstore-go | YES | stdlib + x/crypto; no CGO (but large dep tree) |

All recommended libraries build with `CGO_ENABLED=0`. Confirmed HIGH confidence for the primary stack; MEDIUM only because final version pinning happens at implementation time.

---

## Installation (draft)

```bash
# Core supply-chain stack
go get github.com/CycloneDX/cyclonedx-go@latest
go get github.com/spdx/tools-golang@latest
go get aead.dev/minisign@latest
go get golang.org/x/vuln@latest
go get github.com/in-toto/in-toto-golang@latest
go get github.com/in-toto/attestation@latest
go get github.com/secure-systems-lab/go-securesystemslib@latest

# Build-tag-gated (optional)
go get github.com/sigstore/sigstore-go@latest
go get github.com/google/osv-scanner/v2@latest
```

Build the lean default binary:
```bash
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o omni .
```

Build with optional verification paths:
```bash
CGO_ENABLED=0 go build -tags "omni_sigstore omni_osv" -trimpath -o omni-full .
```

---

## Confidence Assessment

| Area | Confidence | Reason |
|------|------------|--------|
| SBOM libs (cyclonedx-go, spdx/tools-golang) | HIGH | Upstream spec owners, pure Go, Jan 2026 releases verified |
| Minisign (aead.dev/minisign) | HIGH | Tiny, stdlib-based, widely used |
| govulncheck API | HIGH | Go team, stable since v1.0, public `scan` package |
| osv-scanner v2 | MEDIUM | Stable but dep-heavy; build-tag gate recommended |
| sigstore-go | MEDIUM | Upstream-recommended but transitive dep cost; gate behind tag |
| in-toto-golang + securesystemslib | HIGH | Reference implementations, pure Go |
| Syft-as-library rejection | HIGH | Verified dep graph bloat; upstream positions it as a CLI |

---

## Open Questions for Roadmap

1. **Container-image SBOM:** Out of scope for v1.0? (No pure-Go OCI registry client is lean enough.) Recommend: YES, defer.
2. **Keyless signing flow (OIDC → Fulcio → Rekor):** Out of scope? Recommend: YES, verification-only for sigstore path in v1.0.
3. **Vuln DB cache format:** SQLite (modernc) or bbolt? Recommend: bbolt (simpler, already in tree, no schema migrations).
4. **Binary SBOM vs source SBOM:** Both? Recommend: both, sharing the `pkg/sbom/` core; source for `go.mod` trees, binary for `debug/buildinfo`.

---

## Sources

- [cyclonedx-go on GitHub](https://github.com/CycloneDX/cyclonedx-go)
- [spdx/tools-golang on pkg.go.dev](https://pkg.go.dev/github.com/spdx/tools-golang)
- [aead.dev/minisign on pkg.go.dev](https://pkg.go.dev/aead.dev/minisign)
- [sigstore-go repo](https://github.com/sigstore/sigstore-go) — upstream "use this, not cosign" guidance
- [Sigstore Go language client docs](https://docs.sigstore.dev/language_clients/go/)
- [golang.org/x/vuln (govulncheck)](https://pkg.go.dev/golang.org/x/vuln) — public `scan` API since v1.0
- [osv-scanner v2 on pkg.go.dev](https://pkg.go.dev/github.com/google/osv-scanner/v2)
- [in-toto-golang on pkg.go.dev](https://pkg.go.dev/github.com/in-toto/in-toto-golang/in_toto)
- [in-toto attestation Go predicates](https://pkg.go.dev/github.com/in-toto/attestation/go/predicates/provenance/v1)
- [anchore/syft repo](https://github.com/anchore/syft) — library entry points reviewed, dep graph confirmed heavy

*Last updated: 2026-04-11*
