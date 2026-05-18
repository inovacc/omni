# Architecture Research — Supply-Chain Capability Track

**Domain:** Supply-chain tooling (SBOM, signing, scanning, attestation) inside a hexagonal Go CLI
**Researched:** 2026-04-11
**Confidence:** HIGH for component boundaries and build order (validated against syft/grype/cosign/sigstore-go/slsa-github-generator); MEDIUM for exact package-split recommendations (single maintainer, omni-specific constraints).

---

## Executive Findings

1. **The four capabilities form a DAG, not a ring.** SBOM is the root data type; scanning consumes SBOMs; attestation wraps artifacts + optionally SBOMs; signing is the cryptographic primitive every other capability eventually reuses. Build order is driven by this DAG.
2. **The industry reference tools (syft, grype, cosign) already decompose cleanly along the same lines omni's hexagonal layout wants.** Syft's `syft/sbom`, `syft/pkg`, `syft/source`, `syft/cataloger` split maps directly onto `pkg/sbom/{model,source,cataloger}`. Grype's `grype/pkg` + `grype/vulnerability` + `grype/matcher` split maps onto `pkg/scan/{model,db,matcher}`. Cosign is mid-refactor onto `sigstore-go`, which is the leaner pure-library path omni should shadow.
3. **`pkg/sign/` is the natural shared kernel.** Both `attest` (which signs in-toto statements) and any future container/artifact verifier need it. Build signing first as a primitive, then layer attest on top.
4. **Scan must not depend on sbom's internal types.** Both tools should speak a neutral SBOM document format (SPDX JSON or CycloneDX JSON) over the wire, even when chained in-process. Grype already does this — it accepts an SBOM file or a Syft-generated in-memory SBOM through a documented interface, never via direct struct sharing. This is the single biggest boundary decision and it is non-negotiable for long-term maintainability.

---

## Component Boundaries

### System Overview

```
                          ┌───────────────────────────────────────────────┐
                          │  cmd/ (Cobra entry points — thin wrappers)    │
                          │  sbom.go   sign.go   verify.go                │
                          │  scan.go   attest.go                          │
                          └──────┬────────┬────────┬────────┬─────────────┘
                                 │        │        │        │
                          ┌──────┴────────┴────────┴────────┴─────────────┐
                          │  internal/cli/ (I/O glue, cmderr, flags)      │
                          │  sbom/   sign/   scan/   attest/              │
                          └──────┬────────┬────────┬────────┬─────────────┘
                                 │        │        │        │
┌────────────────────────────────┴────────┴────────┴────────┴─────────────┐
│                          pkg/ (pure-Go reusable libraries)              │
│                                                                         │
│   ┌────────────────┐    ┌────────────────┐    ┌────────────────────┐    │
│   │   pkg/sbom/    │    │   pkg/sign/    │    │    pkg/scan/       │    │
│   │                │    │                │    │                    │    │
│   │  source/       │    │  keys/         │    │  model/            │    │
│   │  cataloger/    │    │  sigstore/     │◄───┤  db/               │    │
│   │    gomod/      │    │  minisign/     │    │  matcher/          │    │
│   │    npm/        │    │  verify/       │    │  osv/              │    │
│   │    oci/        │    │                │    │  (consumes SBOM    │    │
│   │  model/        │    │                │    │   document)        │    │
│   │  format/       │    │                │    │                    │    │
│   │    spdx/       │    │                │    │                    │    │
│   │    cyclonedx/  │    │                │    │                    │    │
│   └───────┬────────┘    └───────┬────────┘    └─────────┬──────────┘    │
│           │                     │                       │               │
│           │ produces            │ primitive used by     │ consumes      │
│           │ SBOM doc            ▼                       │ SBOM doc      │
│           │             ┌───────────────┐               │               │
│           └────────────►│  pkg/attest/  │◄──────────────┘               │
│                         │               │                               │
│                         │  intoto/      │   in-toto Statement builder   │
│                         │  slsa/        │   SLSA v1 provenance predicate│
│                         │  dsse/        │   DSSE envelope wrap + sign   │
│                         │  bundle/      │   Sigstore Bundle assembly    │
│                         └───────────────┘                               │
└─────────────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Owns | Imports from omni | Mirrors (upstream) |
|-----------|------|-------------------|--------------------|
| `pkg/sbom/model` | Neutral document type (Document, Package, Relationship, Source) + SPDX/CycloneDX marshal | nothing | `syft/sbom`, `syft/pkg` |
| `pkg/sbom/source` | Abstraction over scan targets: dir, file tree, go.mod, OCI layout | `pkg/twig/scanner` for parallel walk, `pkg/hashutil` | `syft/source` |
| `pkg/sbom/cataloger/*` | One sub-package per ecosystem (gomod, npm, pypi, ociimage). Each is a pluggable `Cataloger` | `pkg/sbom/model`, `pkg/sbom/source` | `syft/cataloger/*` |
| `pkg/sbom/format` | SPDX 2.3 + CycloneDX 1.5 encoders/decoders | `pkg/sbom/model` only | Spec-driven |
| `pkg/sign/keys` | Ed25519/ECDSA key gen, PEM encode/decode, passphrase-encrypted private keys | `pkg/cryptutil` | cosign `pkg/cosign` key handling |
| `pkg/sign/minisign` | Pure-Go minisign-compatible sign/verify — the v1.0 "definitely pure-Go" path | `pkg/sign/keys` | minisign spec |
| `pkg/sign/sigstore` | sigstore-go-style keyless/bundle verification (optional v1.0) | `pkg/sign/keys` | `sigstore/sigstore-go` |
| `pkg/sign/verify` | Unified verification entry point — dispatches by signature format | all above | — |
| `pkg/scan/model` | Neutral vulnerability types (Vulnerability, Match, Advisory, Severity) | `pkg/sbom/model` **via document decode only** (no struct import) | `grype/vulnerability` |
| `pkg/scan/db` | Local DB loader (download, checksum, open). SQLite or flat JSON. | `pkg/sign/verify` (to verify DB signature) | `grype/db` |
| `pkg/scan/osv` | OSV.dev schema + offline OSV archive reader | `pkg/scan/model` | `osv-scanner` |
| `pkg/scan/matcher` | Per-ecosystem matchers (gomod, npm, pypi) keyed off SBOM package types | `pkg/scan/model`, `pkg/scan/db` | `grype/matcher` |
| `pkg/attest/intoto` | in-toto Statement v1 builder (subject digests + predicate envelope) | `pkg/hashutil` | `in-toto/attestation-go` |
| `pkg/attest/slsa` | SLSA v1 provenance predicate struct + builder | `pkg/attest/intoto` | `slsa-framework/slsa-github-generator` |
| `pkg/attest/dsse` | DSSE envelope encode/decode (pre-authentication encoding, signatures array) | `pkg/sign/keys` | `secure-systems-lab/go-securesystemslib` |
| `pkg/attest/bundle` | Sigstore Bundle assembly (statement + signature + cert chain) | `pkg/attest/dsse`, `pkg/sign/sigstore` | `sigstore-go/bundle` |

### Non-Negotiable Boundary Rules

1. **Scan MUST NOT import `pkg/sbom/model` for its struct types.** It may import the format decoder to parse a `Document`, then translate into its own `pkg/scan/model.Package`. This keeps vulnerability matching independently evolvable and allows `omni scan` to accept SBOMs produced by syft, trivy, or any other tool — not just omni's.
2. **Sign MUST NOT depend on SBOM, scan, or attest.** It is a primitive. Everything else consumes it.
3. **Attest MAY depend on sign and sbom (for subject digests) but MUST NOT depend on scan.** Scan results are not part of a provenance attestation; they are a separate VEX/VulnAssess predicate type that is out of scope for v1.0.
4. **No cataloger may import another cataloger.** Each is self-contained and registered via `init()` (same pattern as `pkg/video/extractor/`).

---

## Data Flow

### Flow 1 — `omni sbom <target>`

```
target (dir|file|oci)
  → pkg/sbom/source.Resolve()        (walk tree, build file index)
  → pkg/sbom/cataloger.Run(ctx, src) (fan-out to registered catalogers)
  → []pkg/sbom/model.Package
  → pkg/sbom/model.Document (add relationships, source descriptor)
  → pkg/sbom/format.Encode(doc, SPDX|CycloneDX)
  → io.Writer (stdout or --output file)
```

### Flow 2 — `omni scan <sbom-file>` (or `omni sbom ... | omni scan -`)

```
io.Reader (SBOM JSON)
  → pkg/sbom/format.Decode()          (parse SPDX or CycloneDX)
  → pkg/scan/model.FromDocument()     (translate to scan types — boundary wall)
  → pkg/scan/db.Open()                (verified local DB)
  → pkg/scan/matcher.Match(packages, db)
  → []pkg/scan/model.Match
  → output helper (text|json|sarif)
```

**Key property:** the chained pipeline `omni sbom . | omni scan -` has the same data path as `omni sbom .` followed by `omni scan file.json`. Composable because the boundary is a serialized document, not a Go struct.

### Flow 3 — `omni sign <artifact>` / `omni verify <artifact> <sig>`

```
sign:
  artifact → pkg/hashutil.HashFile (SHA-256 digest)
          → pkg/sign/keys.Load(priv)
          → pkg/sign/minisign.Sign(digest, key)
          → .sig file

verify:
  artifact + sig → pkg/sign/verify.Detect(sig) (minisign? sigstore bundle?)
                → pkg/sign/minisign.Verify OR pkg/sign/sigstore.Verify
                → bool + identity metadata
```

### Flow 4 — `omni attest --artifact X --predicate slsa-provenance.json`

```
artifact(s)                → pkg/hashutil (subject digests)
predicate JSON or builder  → pkg/attest/slsa.Build() or raw JSON
                           → pkg/attest/intoto.NewStatement(subjects, predicate)
                           → pkg/attest/dsse.Wrap(statement)
                           → pkg/sign/keys.Sign (Pre-Auth Encoding of payload)
                           → pkg/attest/bundle.Assemble (optional sigstore bundle)
                           → .intoto.jsonl or .sigstore.json
```

**Critical:** `pkg/attest/dsse` signs the PAE (pre-authentication encoding) of the payload, not the raw JSON. This is the detail that both home-rolled implementations and LLM-generated ones get wrong. Test against in-toto's reference vectors.

---

## Build Order & Rationale

The dependency DAG forces this order. Each phase is independently shippable.

### Phase A — `pkg/sign/` first (signing primitive)

**Why first:** attestation cannot ship without DSSE signing; no other capability unblocks it.
**Scope:** `keys` + `minisign` + `verify` dispatch. Sigstore verification is a later bolt-on.
**Deliverables:** `omni sign`, `omni verify`, `omni keygen`. All three commands land together.
**Risk:** LOW — pure-Go `crypto/ed25519` + `crypto/ecdsa`; minisign format is well-documented and tiny.
**Golden-test hook:** deterministic sign with fixed key → stable signature output (use `SOURCE_DATE_EPOCH` pattern).

### Phase B — `pkg/sbom/` (SBOM generation)

**Why second:** independent of sign for now; unblocks scan and attest-subject-digesting. Start here in parallel with Phase A if a second worker is available, but single-maintainer reality says sequential.
**Scope:** `source` + `model` + `format/spdx` + `format/cyclonedx` + `cataloger/gomod` + `cataloger/ociimage` (tarball layout, not registry pulls — respects the "no exec, no network-by-default" rule).
**Deliverables:** `omni sbom <target> --format spdx|cyclonedx`.
**Risk:** MEDIUM — SPDX 2.3 JSON is verbose and the spec has footguns (relationships, license expression parsing). Start with `gomod` as the only cataloger; add `npm`, `pypi` incrementally.
**Extensibility:** cataloger registration via `init()` mirrors `pkg/video/extractor/` — proven pattern in the codebase.

### Phase C — `pkg/scan/` (vulnerability scanning)

**Why third:** requires SBOM document format to exist, requires sign (to verify DB signatures on download), benefits from both being battle-tested.
**Scope:** `model` + `osv` offline reader + `db` (download-and-verify a pre-built OSV archive) + `matcher/gomod`.
**Deliverables:** `omni scan <sbom>`, `omni scan db update`.
**Risk:** MEDIUM-HIGH — DB distribution is the hardest part. v1.0 should ship with the DB hosted as a signed GitHub release asset on a schedule, verified via `pkg/sign/verify`. Do NOT try to embed live OSV.dev API calls.
**Decision needed in phase planning:** sqlite vs flat JSON archive for the DB. Flat JSON archive is simpler, fits the "no CGO" rule, and grype has been moving toward flat formats (v6 schema). Recommendation: flat per-ecosystem JSON archives, checksum-indexed.

### Phase D — `pkg/attest/` (SLSA attestation)

**Why last:** requires sign (for DSSE signatures) and sbom (to compute subject digests for source dirs).
**Scope:** `intoto` + `dsse` + `slsa` + (optional) `bundle`.
**Deliverables:** `omni attest --predicate-type slsa-provenance --predicate file.json --artifact X`, `omni attest verify`.
**Risk:** LOW on crypto (delegates to sign), MEDIUM on spec compliance (in-toto PAE + SLSA v1 predicate schema must be byte-exact).
**Test strategy:** golden-master against the in-toto reference test vectors; do not invent expected outputs.

### Why not parallelize?

Single-maintainer, 3–4 month timeline, and Phase C (scan) is where all the integration risk lives. Sequential ordering lets Phase A harden before Phase D loads it with DSSE requirements, and lets Phase B's document format stabilize before Phase C tries to consume it. Parallelism would force premature API commitments between pkg/sbom and pkg/scan.

---

## Integration with Existing omni Patterns

| Existing pattern | How the new capabilities adopt it |
|------------------|-----------------------------------|
| `cmd/` → `internal/cli/<cmd>/` → `pkg/<domain>/` three-layer split | Enforced: each of sbom/sign/scan/attest is a full vertical. No shortcuts. |
| `cmderr` sentinels + exit codes | `ErrInvalidInput` (malformed SBOM, bad flags), `ErrNotFound` (key file, DB), `ErrIO` (download failure), `ErrConflict` (signature mismatch → conflict is semantically right per existing grep/sort usage), `ErrUnsupported` (unknown predicate type). No new sentinels needed. |
| `Command` interface + Registry | Each of `sbom`, `sign`, `verify`, `scan`, `attest` registers with `internal/cli/command/`. Enables `omni pipe '{sbom .}' '{scan -}'` from day one. |
| `input.Open(args, r)` abstraction | `sbom` treats `.` / paths / `-` uniformly; `scan` treats `-` as "read SBOM from stdin" naturally. No custom stdin plumbing. |
| `output.New(w, format)` helper | All five commands expose `--json` / `--table` via the existing helper. `scan` additionally supports `--format sarif` as a scan-specific extension. |
| Functional options in `pkg/` (`WithIndent`, `WithMaxFiles`) | Every new `pkg/` uses the same pattern: `sbom.Generate(src, sbom.WithCataloger(gomod.New()), sbom.WithFormat(sbom.SPDX))`. |
| Golden-master tests (`testing/golden/` + `tools/golden/`) | Mandatory. Every new command gets snapshots in both registries. Fixed inputs (test fixtures checked into `internal/cli/<cmd>/testdata/`), deterministic outputs (pinned timestamps, sorted keys). |
| Build-tag platform splits | Not expected for v1.0 — all four capabilities are pure-stdlib + pure-Go crypto. No platform forks anticipated. Flag if signing key storage needs OS keychain integration (defer to post-1.0). |
| `pkg/twig/scanner` for tree walks | `pkg/sbom/source` reuses it directly. One less wheel to reinvent, and it already handles MaxFiles/parallelism/progress. |
| `pkg/hashutil` | Used by `sbom` (file hashes in SPDX), `sign` (digest input), `attest` (subject digests). Central dependency — worth auditing for SHA-512 support before Phase A. |

### What does NOT fit the existing patterns

- **Long-running DB downloads (`omni scan db update`)** break the "one-shot command" ergonomic. Acceptable — precedent exists in `omni video` which downloads large assets. Surface progress via the same pattern `pkg/video/downloader/progress.go` uses.
- **Network access for DB updates** is the first omni capability to *require* network by default. Treat this explicitly: default to `--offline` mode (use embedded or previously-downloaded DB), make network fetch an explicit flag. This keeps the CI/CD user story clean.

---

## Anti-Patterns to Avoid

### Anti-Pattern 1 — Direct struct sharing across capabilities
**What:** `pkg/scan` imports `pkg/sbom/model.Package` and pattern-matches on concrete types.
**Why bad:** Couples vulnerability matching to omni's internal SBOM shape. Blocks accepting third-party SBOMs. Makes either package impossible to evolve without a breaking change.
**Instead:** Scan accepts a decoded `sbom/format.Document` (public neutral type) and translates to `scan/model.Package` at the boundary. Slightly more code, massively more flexibility.

### Anti-Pattern 2 — One monolithic cataloger
**What:** `pkg/sbom/cataloger.go` with a 2000-line switch over ecosystem types.
**Why bad:** Syft tried early versions of this and abandoned it. Every new ecosystem becomes a merge-conflict magnet; testing requires fixtures for every ecosystem even when you touch one.
**Instead:** `pkg/sbom/cataloger/gomod/`, `pkg/sbom/cataloger/npm/`, etc. Each registers itself via `init()` in a blank-imported `pkg/sbom/cataloger/all/` barrel (same pattern omni already uses for `pkg/video/extractor/all/`).

### Anti-Pattern 3 — Rolling your own DSSE encoding
**What:** Hand-crafting the PAE (Pre-Authentication Encoding) inline in `pkg/attest`.
**Why bad:** The PAE format is `DSSEv1 <len(type)> <type> <len(payload)> <payload>` and every implementation that invents it ships a bug. The spec has test vectors specifically because this is a footgun.
**Instead:** Implement PAE once in `pkg/attest/dsse/pae.go` with the spec test vectors as unit tests. Reference: in-toto DSSE spec v1.0.

### Anti-Pattern 4 — Embedding cosign as a library
**What:** `go get github.com/sigstore/cosign` and calling `cosign.SignBlob`.
**Why bad:** Cosign's dependency tree is enormous (containerd, OCI registries, Fulcio clients, TUF). It will double omni's binary size and force CGO in some transitive paths. Cosign maintainers themselves are refactoring onto `sigstore-go` for exactly this reason.
**Instead:** Depend on `sigstore-go` if you need keyless flows, minisign-only if you don't. v1.0's "minimum acceptable" path from `.planning/PROJECT.md` is already minisign-style pure-Go.

### Anti-Pattern 5 — Scan results inside provenance attestations
**What:** Bundling `omni scan` output as a SLSA provenance predicate.
**Why bad:** Wrong predicate type. Scan results belong in a VEX or vulnerability-assessment predicate, not in SLSA provenance. Mixing them ships malformed attestations.
**Instead:** Out of scope for v1.0. If needed post-1.0, add `pkg/attest/vex/` as a separate predicate builder.

---

## Scalability Considerations

| Concern | Small repo (omni itself) | Medium (1K deps) | Large (10K+ files, container images) |
|---------|--------------------------|------------------|----------------------------------------|
| SBOM generation time | <1s | 1–5s | 10–60s — parallel walk via `pkg/twig/scanner.WithParallel()` |
| Scan matcher time | <100ms | 500ms–2s | 5–15s — pre-index DB by ecosystem, short-circuit on name mismatch |
| DB size on disk | N/A | ~50MB compressed | Same (DB doesn't scale with target) |
| Signing time | <10ms | <10ms | <10ms (signs digest, not content) |
| Attestation size | ~2KB | ~2KB | Grows only with subject count, not target size |

No architectural concerns at these scales. Performance tuning is per-phase after the boundary shape is right.

---

## Roadmap Implications (for the phase planner)

- **Four phases maps cleanly onto four capabilities in DAG order.** Do not try to compress into fewer phases — each capability is non-trivial (2–4 weeks each for a single maintainer) and each has a shippable user-visible surface.
- **Phase A (sign) is the lowest-risk warmup.** Good first phase after the polish track completes — lets the maintainer get comfortable with the new-capability workflow before Phase B's spec-compliance grind.
- **Phase C (scan) likely needs a mid-phase research spike** on DB distribution format. Flag this for deeper research when Phase C is scheduled. The OSV ecosystem has moved fast in 2025–2026 and training-data knowledge is stale.
- **Phase D (attest) should reuse Phase A heavily.** If Phase A ships `pkg/sign/keys` and `pkg/sign/verify` with a clean API, Phase D is mostly spec-translation work, not crypto work.
- **Golden-master coverage cost** for supply-chain commands is higher than average because outputs include timestamps, hashes, and signatures. Budget time in each phase for fixture scrubbing + deterministic-output flags (accept `--source-date-epoch`, sort keys, pin random sources).

---

## Confidence Assessment

| Claim | Confidence | Basis |
|-------|------------|-------|
| DAG order sign → sbom → scan → attest | HIGH | Dependency structure is forced by crypto-primitive requirements; verified against syft/grype/cosign module graphs |
| Scan must speak document format not structs | HIGH | Validated behavior of grype (accepts arbitrary SBOMs); industry best practice |
| Minisign-style pure-Go is sufficient for v1.0 sign | HIGH | Already stated in `.planning/PROJECT.md` "Out of Scope" as the acceptable minimum |
| Cosign dependency tree is too heavy | HIGH | Confirmed by sigstore-go's stated purpose of minimizing deps, and cosign's own migration direction |
| Cataloger-per-ecosystem subdir + init()-register | HIGH | Direct analogue of existing `pkg/video/extractor/` pattern, which is working well in omni |
| Flat JSON DB archives over sqlite for scan | MEDIUM | Grype v6 schema direction + no-CGO rule; needs a dedicated spike in Phase C planning |
| DSSE PAE implementation is a footgun | HIGH | Spec-documented test vectors exist specifically because implementations get it wrong |
| Scan results do not belong in SLSA provenance | HIGH | Spec-defined predicate types; SLSA provenance schema does not contain vuln fields |
| No platform-specific splits needed in any of the four | MEDIUM | No obvious need identified, but OS keychain integration for key storage is a classic late-stage surprise |

---

## Sources

- [anchore/syft — GitHub](https://github.com/anchore/syft)
- [syft package godoc](https://pkg.go.dev/github.com/anchore/syft/syft)
- [syft/sbom package godoc](https://pkg.go.dev/github.com/anchore/syft/syft/sbom)
- [How Syft Scans Software to Generate SBOMs — Anchore blog](https://anchore.com/blog/how-syft-scans-software-to-generate-sboms/)
- [Grype architecture — Anchore OSS docs](https://oss.anchore.com/docs/architecture/grype/)
- [anchore/grype — GitHub](https://github.com/anchore/grype)
- [Deep Dive: Where Does Grype Data Come From? — Chainguard](https://dev.to/chainguard/deep-dive-where-does-grype-data-come-from-n9e)
- [sigstore/cosign — GitHub](https://github.com/sigstore/cosign)
- [cosign/pkg/cosign godoc](https://pkg.go.dev/github.com/sigstore/cosign/pkg/cosign)
- [sigstore/sigstore-go — GitHub](https://github.com/sigstore/sigstore-go)
- [SLSA Provenance spec v1.0](https://slsa.dev/spec/v1.0/distributing-provenance)
- [SLSA Build Provenance draft](https://slsa.dev/spec/draft/build-provenance)
- [slsa-github-generator/internal/builders/generic godoc](https://pkg.go.dev/github.com/slsa-framework/slsa-github-generator/internal/builders/generic)
- [SLSA Software attestations model](https://slsa.dev/attestation-model)
- [SLSA + in-toto blog post](https://slsa.dev/blog/2023/05/in-toto-and-slsa)

---

*Architecture research: 2026-04-11 — omni supply-chain capability track*
