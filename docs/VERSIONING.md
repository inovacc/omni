# Versioning & API Stability

> Last reconciled: 2026-06-04 (verified `frozen ∩ experimental = ∅`, `freeze:check` in sync).

## Policy

Semantic versioning applies to the **frozen, non-`// Experimental:` `pkg/*` public
surface** — the library API external consumers may depend on. The authoritative
frozen set is enumerated by `task freeze:check` against `docs/API-FREEZE.md`
(ADR-0010); breaking changes to it follow the CLAUDE.md 30-day deprecation protocol.

Packages whose `doc.go` (or a source file) carries `// Experimental:` are **exempt**
from the freeze and from semver guarantees until promoted. The CLI surface itself
(`omni <cmd>` flags/output) is governed separately by the golden-master suite, not
by this freeze.

**Promoting an Experimental package to stable:** remove its `// Experimental:`
marker → it then appears in `tools/freeze` output → update `docs/API-FREEZE.md`
(`task freeze:check`) in the same commit → from then on it is frozen.

## Current state (2026-06-04)

| Set | Count | Source of truth |
|-----|-------|-----------------|
| Frozen `pkg/*` packages | 27 (598 symbols) | `docs/API-FREEZE.md` / `task freeze:check` |
| `// Experimental:` packages | 22 | `grep -rl '// Experimental:' pkg/` |
| Overlap (must be ∅) | 0 | verified disjoint 2026-06-04 |

> ADR-0010's Context cites "17 Experimental packages" — that was true at phase
> entry. The count is now **22** (the gops process-tools port and additional
> `pkg/video` subpackages were marked Experimental afterwards). The freeze set is
> unaffected: all 22 remain outside the frozen surface.

### The 22 Experimental packages (keep-vs-promote rationale)

**Supply-chain (Phases 05–07, recent — keep Experimental until formats settle):**
`pkg/attest`, `pkg/scan`, `pkg/sbom/model`, `pkg/sbom/collect`, `pkg/sbom/purl`.
Their on-disk formats and option structs may still change. The deliberate
exception is `pkg/sbom/format` — the cross-package stable boundary `pkg/scan`
imports — which **is frozen** (ADR-0007).

**Process tools (recent gops port — keep Experimental):**
`pkg/gopsagent`, `pkg/procmetrics`, `pkg/procutil`, `pkg/obfuscate`. New surface,
runtime-coupled; allow it to settle before freezing.

**Video (inherently unstable — keep Experimental):**
`pkg/video` + `pkg/video/{cache,downloader,extractor,extractor/all,extractor/generic,extractor/youtube,format,jsinterp,m3u8,nethttp,types,utils}`.
YouTube player-JS changes routinely break signature decryption (see BUG-0001 and
`docs/BACKLOG.md`); freezing this API would promise stability the upstream cannot
honor. Not part of the supply-chain core value.

No package is currently a promotion candidate for the honest supply-chain release —
the frozen 27 already cover the stable library surface.

## Release tag number — decided: `v1.6.0` (2026-06-04)

The repo already carries `v1.0.0`–`v1.5.0` tags from the earlier command-coverage
milestones, so the honest supply-chain release cannot re-publish `v1.0.0`. **Decision
(maintainer, 2026-06-04): the next tag is `v1.6.0`.**

Rationale: the supply-chain work is purely *additive* (new `sign`/`sbom`/`scan`/
`attest`/`reprocheck` commands + new packages; zero breaking changes to the frozen
27-package surface). Semver: additive ⇒ minor bump. (`v2.0.0` was considered as a
"reset" signal but rejected — nothing in the frozen surface breaks.)

Release notes live in `docs/RELEASE-NOTES-v1.6.0.md`. The tag itself is **not yet
cut** — that remains a deliberate operator action requiring GitHub Actions to be
re-enabled (ADR-0010).
