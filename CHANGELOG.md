# Changelog

All notable changes to **omni** are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versions aim to follow
[Semantic Versioning](https://semver.org/) for the **frozen** `pkg/*` surface
(see `docs/API-FREEZE.md` — `// Experimental:` packages are exempt until triage).

## [Unreleased]

The supply-chain milestone (a self-verifying, dependency-light v1.0). All pure-Go,
no-CGO, no new external processes in the binary.

### Added
- **`omni sign` / `omni verify` / `omni sign keygen`** — minisign-compatible Ed25519
  signing with fail-closed verification; redacting `secret.Key` wrapper
  (ADR-0005, ADR-0006).
- **`omni sbom`** — byte-deterministic SPDX 2.3 / CycloneDX 1.5 SBOMs for a Go module
  dir or binary; pure-stdlib emitter; optional schema validation behind the
  `omni_sbomvalidate` build tag (ADR-0007). Stable `pkg/sbom/format.Document`
  boundary + read side (`format.Parse`, `Document.Components`, `purl.Parse`).
- **`omni scan`** — pure-Go OSV vulnerability matcher over a `pkg/sign`-signed
  `osv-db.zip`; `--fail-on <severity>` CI gate, `--max-db-age` staleness gate,
  CVSS v3.1 severity bands; `omni scan db update` (verify-before-write) (ADR-0008).
  Reachability (`omni scan source`) deferred to a future `contrib/` module.
- **`omni attest` / `omni attest verify`** — in-toto Statement v1 / SLSA Provenance v1
  in a DSSE (PAE) envelope, signed via `pkg/sign`; honest SLSA Build **L2** via a
  code-enforced `builder.id` allowlist (no overclaim); CI `task attest:validate-schema`
  gate (ADR-0009).
- **`omni reprocheck`** — deterministic sha256 digest-pair reproducibility gate;
  build-info-derived `omni --version` (ADR-0010).
- **Release machinery** — reproducible GoReleaser builds (`-trimpath`/`-buildvcs`/
  pinned timestamp) that dogfood `omni sign`/`sbom`/`attest`; self-contained release
  workflow with a dual-build reproducibility gate + cross-OS dogfood; `task freeze:check`
  + `docs/API-FREEZE.md` API-freeze gate; `docs/RELEASE-NOTES-v1.0.md`.
- Hardening sweep (51 findings): tar/zip-slip containment, both-platform command-
  injection fixes, decompression-bomb caps, fail-closed crypto.

### Changed
- CI `quality` jobs are self-contained (the previously-broken external reusable
  workflow was removed).
- `golangci-lint` clean; cross-platform (Linux/macOS/Windows × amd64/arm64) builds verified.

### Notes
- Golden-master fixtures are `-text`-locked (`.gitattributes`) so hashed/signed
  inputs are byte-stable across platforms.
- Cutting a tagged release publishes signed + SBOM'd + SLSA-L2-attested archives;
  see `docs/RELEASE-NOTES-v1.0.md` for the honest "what's NOT protected" scope.

## [v1.5.0] – earlier
Command-coverage milestones (160+ Unix-utility replacements across Core/File/Text/
System/Process/Archive/Hash/Encoding/Data/Network/Cloud-DevOps). See `git log` and
`docs/ROADMAP.md` for the per-phase history prior to the supply-chain track.

[Unreleased]: https://github.com/inovacc/omni/compare/v1.5.0...HEAD
[v1.5.0]: https://github.com/inovacc/omni/releases/tag/v1.5.0
