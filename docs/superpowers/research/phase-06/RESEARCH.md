# Phase 06 — `omni scan` Research Synthesis

**Date:** 2026-06-03
**Scope:** Vulnerability scanning of `pkg/sbom/format.Document` purls against an offline, signed OSV database, with optional Go reachability. Synthesis of four research reports covering OSV distribution, `golang.org/x/vuln`, the OSV schema + CVSS, and prior art (osv-scanner v2, grype v6).

---

## Executive summary

The Phase-06 plan is **largely confirmed** with two precise corrections and one decisive fork.

- **DB format — CONFIRMED.** A single, `pkg/sign`-detached-minisig-signed `osv-db.zip` of flat OSV-JSON (one `<ID>.json` per record) plus an omni-synthesized `manifest.json` **inside** the signed zip is exactly the industry-idiomatic offline layout. It is pure-stdlib (`archive/zip` + `encoding/json`), byte-deterministic, diffable, and golden-testable. **Reject bbolt.** Upstream osv.dev *is* flat-JSON-in-zip, and osv-scanner's own offline store is `{ecosystem}/all.zip` with no embedded DB engine ([osv.dev data sources](https://google.github.io/osv.dev/data/), [osv-scanner offline mode](https://google.github.io/osv-scanner/usage/offline-mode/)).
- **Matching — CONFIRMED.** The official OSV `IsVulnerable` algorithm (sorted-event interval walk OR'd with exact `versions[]` membership) using `golang.org/x/mod/semver` is pure-Go, no-exec, deterministic, and golden-testable ([OSV evaluation](https://ossf.github.io/osv-schema/#evaluation)). The single load-bearing correctness detail is the **leading-`v` normalization**.
- **Severity — PARTIALLY CONFIRMED.** CVSS v3.1 base score is a tiny closed-form formula and is fine to hand-roll. CVSS v4.0 is **NOT** "tiny" — it needs FIRST's ~270-entry MacroVector lookup table + EQ interpolation. Ship v3.1 hand-rolled; defer/avoid hand-rolling v4.0 ([CVSS v4.0 spec](https://www.first.org/cvss/v4-0/specification-document)).
- **Reachability — THE CENTRAL FORK.** Source-mode Go reachability is **impossible** under omni's no-exec invariant (it requires `go/packages`, which execs `go list`) **and** drags `golang.org/x/vuln` + `golang.org/x/tools v0.44.0` into the **main** `go.mod` via MVS — the exact Phase-04 sigstore mistake ADR-0007 codifies. **Plan option (a) is contradicted on both hard constraints.** Recommendation below: **ship (c) for v1.0** (default scan source → `cmderr.ErrUnsupported`, backlog reachability), with **(b)** (a self-contained `contrib/govulncheck-scan` module) as the eventual home if/when reachability ships.

---

## OSV DB distribution & local bundle format

**Upstream distribution (osv.dev).** The Go ecosystem dump is at `gs://osv-vulnerabilities/Go/all.zip`, fetchable over plain HTTPS at `https://osv-vulnerabilities.storage.googleapis.com/Go/all.zip` (ecosystem name is the literal `Go`). Inside, the layout is **flat**: one OSV-JSON file per vulnerability named `<ID>.json` (e.g. `GO-2024-2965.json`) — **no aggregation, no manifest.json** in the upstream zip. Individual records also live at `gs://osv-vulnerabilities/Go/<ID>.json` ([osv.dev data sources](https://google.github.io/osv.dev/data/), [unzip listing confirmation](https://yossarian.net/til/post/osv-dev-provides-data-dumps/)).

**Alternate authoritative source (vuln.go.dev).** The Go team's `vuln.go.dev` serves the **same OSV schema** in a **different shape**: an HTTP index API (`/index/db.json` carrying a single `modified` RFC3339-`Z` timestamp, `/index/modules.json`, `/index/vulns.json`) plus per-ID `/ID/GO-YYYY-NNNN.json`, and a whole-DB bundle at `vuln.go.dev/vulndb.zip`. HTTP clients fetch `.json.gz`; `file://` clients fetch `.json` ([Go vuln DB format](https://go.dev/doc/security/vuln/database)). vuln.go.dev is Go-team-curated and carries **symbol-level** data in `ecosystem_specific.imports` that the reachability feature would need; osv.dev's `Go/all.zip` is a mirror of `github.com/golang/vulndb` and is the better choice for a uniform single-format feed.

**Recommended local store.** A single `osv-db.zip` whose **flat layout mirrors upstream** (one `<ID>.json` per entry) **plus an omni-synthesized `manifest.json` placed INSIDE the zip**, with the whole zip covered by **one detached `pkg/sign` minisig** (`.minisig`). This matches osv-scanner's `{cache}/osv-scanner/{ecosystem}/all.zip` pattern (pure zip-of-JSON, no DB engine) ([osv-scanner offline mode](https://google.github.io/osv-scanner/usage/offline-mode/)). grype v6 uses a SQLite-blob store (~65 MB, CGO-free `modernc.org/sqlite`) — heavier, OSV-inspired, and an optional scale-up, **not** a requirement ([Anchore grype-db v5→v6](https://anchore.com/blog/grype-db-schema-evolution-from-v5-to-v6-smaller-faster-better/)).

**Why zip, not bbolt.** The upstream format already *is* flat JSON in a zip; `archive/zip` is stdlib and streamable; entries are independently diffable/golden-testable; and a zip with **deterministic entry order + zeroed timestamps + fixed compression** is byte-reproducible. bbolt adds a non-stdlib dep, an opaque non-diffable blob, and non-deterministic page layout — all anti-goals.

**Determinism + signing.** Mirror `pkg/sbom/format.Document`'s established pattern: sort entries by ID, zero/fix archive timestamps, use Store or fixed-level Deflate, and **byte-passthrough** of upstream entry JSON (do **not** re-marshal — the OSV schema is forward-compatible and re-marshal risks dropping unknown fields). The `manifest.json` (omni build/fetch time, source URL, schema_version, entry count, content hash) goes inside the signed zip so the single `.minisig` covers everything. Upstream provides **no** cryptographic signing of dumps (integrity is only GCS TLS + per-record `modified`), so omni's own `pkg/sign` minisig layer is sound and additive — and `pkg/sign.Verify` is fail-closed (any failure → `ErrVerification`) ([OSV schema forward-compat](https://ossf.github.io/osv-schema/)).

**Freshness.** osv.dev targets ≤15 min staleness 99.5% of the time and regenerates dumps continuously; offline copies do **not** auto-update. `--max-db-age` must gate on `manifest.generated` (omni's fetch/build time, signed and tamper-evident) — **not** upstream's internal `modified`, which the spec warns "should not be compared to wall clock time." Incremental refresh is possible via `modified_id.csv` (reverse-chronological, stream-and-stop) ([osv.dev data sources](https://google.github.io/osv.dev/data/)).

**Size.** Go corpus is low-thousands of entries (GO-IDs reach GO-2025-3xxx), each JSON a few KB; `Go/all.zip` is a few MB — small enough to ship and golden-test fully. Exact byte size is unpublished; **measure at first fetch** ([osv.dev Go list](https://osv.dev/list?ecosystem=Go)).

**The fetch step must stay OUT of the default runtime.** It needs network and would tempt exec. Build + sign `osv-db.zip` in CI or a self-contained `contrib/` module; default `omni scan` runs purely offline against the local signed zip.

---

## `golang.org/x/vuln`: API, exec behavior, dependency weight

**`scan.Command` does NOT itself fork a subprocess.** It runs govulncheck in-process in a goroutine. `scan/scan.go` (v1.3.0) imports only `context`, `errors`, `io`, `os`, and `golang.org/x/vuln/internal/scan` — **no `os/exec`**. `Start()` launches `go func(){ c.err = c.scan() }()`; `Wait()` reads off the done channel. The godoc "similar to exec.Cmd" describes the **API shape**, not a real child process ([x/vuln scan.go v1.3.0](https://go.googlesource.com/vuln/+/v1.3.0/scan/scan.go)).

**The no-exec violation is ONE LAYER DEEPER.** In-process source/reachability analysis loads packages via `golang.org/x/tools/go/packages`, whose default driver shells out to `go list` (the Go toolchain). There is no in-process, no-exec mode for source reachability. The only documented bypass — `GOPACKAGESDRIVER` — is *also* a subprocess ([go/packages](https://pkg.go.dev/golang.org/x/tools/go/packages), [golang/go#62114](https://github.com/golang/go/issues/62114)). Binary mode (`-mode=binary`) avoids `go list` but yields **symbol-presence, not a call graph** (no call stacks, possible false positives, falls back to all-deps on stripped binaries) ([govulncheck cmd](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)).

**Dependency weight.** x/vuln v1.3.0 directly requires `golang.org/x/tools v0.44.0` (the `go/packages` + `go/ssa` + `go/analysis` universe), plus `x/mod`, `x/sync`, `x/telemetry`, `x/sys`. `x/tools` is the dominant transitive weight ([x/vuln go.mod](https://go.googlesource.com/vuln/+/v1.3.0/go.mod)).

**MVS bumps the MAIN go.mod regardless of build tags.** A `//go:build omni_govulncheck` import still adds `require golang.org/x/vuln` (and transitively `x/tools v0.44.0`) to the main `go.mod`/`go.sum` — build tags gate **linking**, never the module requirement graph that `go mod tidy`/MVS computes ([Go Modules MVS](https://go.dev/ref/mod#minimal-version-selection)). This is precisely the Phase-04 sigstore lesson ADR-0007 codifies, and why `contrib/sigstore-verify` is a **separate module** (`module github.com/inovacc/omni/contrib/sigstore-verify`).

**JSON stream + internal structs.** govulncheck `-json` is a newline-delimited stream of `Message` objects (exactly one of `config`/`progress`/`SBOM`/`osv`/`finding`). `Finding` carries the OSV `id`, `fixed_version`, and a `Trace []*Frame` ordered from the vulnerable symbol up to the entry point (`Frame` = Module/Version/Package/Function/Receiver/Position). But these live in `golang.org/x/vuln/internal/govulncheck` — **not importable**; a parser must redeclare them ([internal/govulncheck](https://pkg.go.dev/golang.org/x/vuln/internal/govulncheck)).

---

## OSV schema & matching algorithm

**Schema version.** Current is **1.7.5 (2026-01-21)** — matches the plan. Required fields are only `id`, `modified`, and (for >1.0.0) `schema_version`; everything else is optional. The schema is explicitly **forward-compatible**: "a client that knows how to read 1.2.0 can process 1.3.0 by ignoring any unexpected fields." omni's reader must validate only the 3 required fields and **tolerate unknown fields**, or a future 1.7.6/1.8.0 entry breaks scan ([OSV schema](https://ossf.github.io/osv-schema/)).

**Matching algorithm (verbatim from the spec).** A version is affected if `IncludedInVersions(v) || IncludedInRanges(v)` (the two are **OR'd, never AND'd**). The range walk iterates sorted events: `introduced && v>=introduced → vulnerable=true`; `fixed && v>=fixed → vulnerable=false`; `last_affected && v>last_affected → vulnerable=false`. **Correctness traps:**
- `fixed` uses `>=` (the fixed version is **NOT** vulnerable) but `last_affected` uses strict `>` (the last_affected version **IS** still vulnerable) — do not conflate.
- Treat empty/`0` `introduced` as "from genesis."
- Route **ECOSYSTEM** and **GIT** ranges to exact `versions[]` membership (the spec says GIT ranges still require an enumerated `versions[]`; osv.dev auto-populates `versions[]` from ECOSYSTEM ranges at ingestion). If an ECOSYSTEM range has **no** `versions[]`, treat as non-match (or surface as indeterminate) rather than guess ([OSV evaluation](https://ossf.github.io/osv-schema/#evaluation), [range types](https://ossf.github.io/osv-schema/#affectedrangestype-field)).

**THE load-bearing normalization.** OSV SEMVER ranges store versions **WITHOUT** a leading `v` (`1.2.3`), but `golang.org/x/mod/semver` **REQUIRES** a leading `v` (`v1.2.3`). The matcher MUST prepend `v` to **both** the OSV event versions and the SBOM module version before `semver.Compare`. Go-specific gotchas also handled correctly by `x/mod/semver`: `+incompatible` (build metadata ignored), pseudo-versions (`v0.0.0-2021...-abcdef` sort correctly), and the Go DB's conservative `introduced: 0` ([x/mod/semver](https://pkg.go.dev/golang.org/x/mod/semver), [Go DB version notes](https://go.dev/doc/security/vuln/database)).

**Go `ecosystem_specific`.** Always an object with a single key `imports`, an array of `{path, symbols[], goos[], goarch[]}` — this is the symbol data the reachability feature would consume ([Go vuln DB](https://go.dev/doc/security/vuln/database)).

**No heavy/exec deps needed.** Matcher = `encoding/json` + `golang.org/x/mod/semver` (pure-Go, already promoted into the graph for the SBOM purl policy per ADR-0007). No ADR-0007 / MVS contamination.

---

## CVSS severity bands

**Severity is a VECTOR STRING, never a number.** OSV `severity[].score` is always the compressed CVSS vector (e.g. `CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H`), with `severity[].type` ∈ {CVSS_V2, CVSS_V3 [≥3.0,<4.0], CVSS_V4 [≥4.0,<5.0], Ubuntu, …}. Severity can appear top-level **and** per-`affected`. So the "parser" must first **parse the vector** then **compute the score** — it is a scorer, not just a parser ([OSV severity](https://ossf.github.io/osv-schema/#severity-field)).

**The qualitative band table is identical across v3.1 and v4.0** and matches the plan: None 0.0 / Low 0.1–3.9 / Medium 4.0–6.9 / High 7.0–8.9 / Critical 9.0–10.0 ([CVSS v3.1 §5 Table 14](https://www.first.org/cvss/v3.1/specification-document), CVSS v4.0 Table 22). So band mapping is trivial once a score exists.

**CVSS v3.1 base score — tiny, hand-roll it.** Closed-form arithmetic (§7.1): `ISS = 1 - (1-C)(1-I)(1-A)`; `Impact = 6.42·ISS` (Scope Unchanged) or `7.52·(ISS-0.029) - 3.25·(ISS-0.02)^15` (Changed); `Exploitability = 8.22·AV·AC·PR·UI`; `BaseScore = Roundup(min(Impact+Exploitability,10))` unchanged or `Roundup(min(1.08·(...),10))` changed; 0 if Impact≤0. **Only traps:** exact metric weights, PR's weight depending on Scope, and the **v3.1-specific Roundup** (ceil-to-1-decimal via integer math, NOT naive `math.Ceil`) ([CVSS v3.1 §7.1](https://www.first.org/cvss/v3.1/specification-document)).

**CVSS v4.0 base score — NOT tiny, do not hand-roll for v1.0.** v4.0 has **no closed-form equation**; it requires FIRST's ~270-entry MacroVector lookup table (`cvss_lookup.js`) plus EQ1–EQ6 equivalence-class interpolation. Disproportionate for a CI gate. **Prefer any producer-supplied numeric score for v4.0 records and defer a full v4.0 scorer** ([CVSS v4.0 spec](https://www.first.org/cvss/v4-0/specification-document)).

**Strategy:** prefer numeric score when present (e.g. producer `database_specific`); else compute v3.1; map via the band table; gate `--fail-on` → `cmderr.ErrConflict` (exit 1). Golden-test the v3.1 scorer against FIRST reference vectors at the band boundaries (3.9/4.0/6.9/7.0/8.9/9.0).

---

## Reachability feasibility & the no-exec/MVS fork

This is the central decision. **Source-based Go reachability is the govulncheck library** (`x/vuln` + `x/tools/go/{packages,ssa,callgraph}`, embedded in gopls), which osv-scanner invokes behind `--experimental-call-analysis` ([Go vuln editor integration](https://go.dev/doc/security/vuln/editor)). Adopting it means importing the exact ADR-0007-forbidden stack **and** execing `go list` — **both** hard constraints violated simultaneously. SSA needs fully type-checked packages → `go/packages` → `go list` subprocess (the only bypass, `GOPACKAGESDRIVER`, is also exec). Hand-rolling `go/parser`+`go/types` without `go/packages` cannot resolve the module graph deterministically. **There is no pure-Go, no-exec path to source reachability** ([go/packages](https://pkg.go.dev/golang.org/x/tools/go/packages), [golang/go#62114](https://github.com/golang/go/issues/62114)).

Binary mode (`-mode=binary`) reads the compiled binary's symbol table — no source, no `go list`, no SSA — but gives symbol-presence (not call stacks) and is what omni effectively already does from the SBOM ([govulncheck cmd](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)).

### Constraint-by-constraint scoring of the three options

| Constraint (NON-NEGOTIABLE) | (a) x/vuln in main go.mod behind `//go:build omni_govulncheck` | (b) separate `contrib/govulncheck-scan` module | (c) drop from v1.0; source → `ErrUnsupported`; backlog |
|---|---|---|---|
| **No-exec invariant** | VIOLATED — source mode execs `go list` | Tolerable only behind an **explicit sanctioned `go list` exception** (like buf-generate/git-hacks per MEMORY.md); module is exec-allowed | SATISFIED — no exec at all |
| **Lean go.mod / ADR-0007 MVS rule** | VIOLATED — MVS pulls `x/vuln` + `x/tools v0.44.0` into main `go.mod` regardless of tag | SATISFIED — heavy deps confined to the contrib module, never the root graph | SATISFIED — zero new deps |
| **Pure-Go / deterministic / golden-testable** | reachability output is non-deterministic across toolchain versions | same caveat, but isolated | SATISFIED — core is fully deterministic |
| **v1.0 timeline (polish→supply-chain→release)** | low effort but wrong | second module + sanctioned-exception policy work | lowest effort, ships now |

**Verdict:** Option (a) is **contradicted** — it fails the two hardest constraints. Option (c) is the only choice that keeps **every** invariant intact for v1.0. Option (b) is the correct **future** home for reachability when it ships — never a main-module build tag, and only behind a documented sanctioned-`go list`-exec exception added to the no-exec policy.

---

## Plan validation (confirmed / contradicted / unclear)

| # | Plan claim | Verdict | Note |
|---|---|---|---|
| 1 | OSV DB = single signed `.zip` (flat JSON + `manifest.json`), NOT bbolt | **confirmed** | Upstream IS flat-JSON-in-zip; osv-scanner's offline store is `{ecosystem}/all.zip`. Refinement: upstream has NO `manifest.json` — omni must **synthesize** it and place it **inside** the signed zip so the one `.minisig` covers it. Byte-passthrough entry JSON, do not re-marshal. |
| 2 | `manifest.generated` drives `--max-db-age` | **confirmed** | Gate on omni's build/fetch time (signed, tamper-evident), NOT upstream `modified` (which "should not be compared to wall clock"). Embed upstream `modified` for provenance only. |
| 3 | Entries conform to OSV schema 1.7.5 | **confirmed** | 1.7.5 is current (2026-01-21). Validate only the 3 required fields (`id`, `modified`, `schema_version`) and **tolerate unknown fields** — schema is forward-compatible. |
| 4 | `x/vuln/scan` execs `go list` → violates no-exec → reachability opt-in | **confirmed (mechanism corrected)** | Conclusion right; mechanism imprecise. `scan.Command` runs **in-process**; the exec is one layer deeper — `go/packages` execs `go list`. Net effect identical. Reword the plan accordingly. |
| 5 | MVS bump from x/vuln accepted, "confined to the tagged file"; default build does not link x/vuln | **contradicted** | "Confined to the tagged file" is false for `go.mod`. MVS adds `require x/vuln` + `x/tools v0.44.0` to the MAIN `go.mod` regardless of tag — the Phase-04 sigstore mistake. Switch to option (b) or (c). |
| 6 | Tiny pure-Go CVSS v3.1/v4.0 base-score parser, no external CVSS lib | **unclear (split)** | v3.1 CONFIRMED tiny (closed-form). v4.0 CONTRADICTED as "tiny" — needs ~270-entry MacroVector table + EQ interpolation. Ship v3.1; defer v4.0 / prefer producer-supplied score. It is a scorer, not just a parser. |
| 7 | Match via SEMVER event interval walk + exact `versions` fallback | **confirmed** | Matches official `IsVulnerable` verbatim. Honor `fixed`(`>=`, unaffected) vs `last_affected`(`>`, still affected); add leading-`v` normalization. |
| 8 | ECOSYSTEM ranges fall back to exact `versions` membership | **confirmed** | Spec-sanctioned (uninterpreted strings). Same for GIT. If ECOSYSTEM range has no `versions[]`, non-match/indeterminate — do not guess. |
| 9 | Reachability eliminates false positives, opt-in only | **confirmed (soften wording)** | Soften "eliminates" → "substantially reduces" (reflection/dynamic calls invisible; binary mode still reports present-but-unreachable). Opt-in is correct — but MUST be a separate module, never a main-module build tag. |
| 10 | Default scanner is pure-Go version-range matching, zero exec | **confirmed** | Fully feasible, well-precedented (osv-scanner's matching minus its network/exec). `encoding/json` + `x/mod/semver`, golden-testable, `ErrConflict` (exit 1) gate. |

---

## Open questions

1. **Exact `Go/all.zip` byte size + entry count** are unpublished — measure at first fetch before committing embed-vs-ship.
2. **Authoritative feed:** osv.dev `Go/all.zip` (uniform single-format) vs `vuln.go.dev/vulndb.zip` (Go-team-curated, carries `ecosystem_specific.imports` symbol data needed for reachability). If reachability ever ships, confirm the chosen feed carries symbol data.
3. **Sanctioned-exec policy:** Does omni's no-exec invariant admit a sanctioned `go list` exception (as it does for buf-generate/git-hacks/terraform per MEMORY.md)? This gates whether reachability is ever buildable. If yes, it lives in a contrib module; if categorically off-limits, reachability is permanently out.
4. **Severity numeric source:** Does the signed DB carry any producer-supplied numeric/qualitative rating (in `database_specific`) that would let omni skip CVSS computation entirely and dodge the v4.0 scorer problem?
5. **Cross-version severity precedence:** When a record has both CVSS_V3 and CVSS_V4 (or multiple per-`affected`) entries, pick a deterministic, documented rule (recommend: highest computed score). FIRST does not define cross-version precedence.
6. **Struct redeclaration:** If reachability ever ships, its parser must redeclare `Message`/`Finding`/`Frame`/`OSV` (they are in `x/vuln/internal/`). Prefer matching the signed DB against the SBOM directly so `x/vuln` is not a core dependency at all.

---

## Recommendation

**Adopt the plan with these locked decisions:**

1. **DB format — ship as written, refined.** A single `pkg/sign`-detached-minisig-signed `osv-db.zip`: flat `<ID>.json` mirroring osv.dev's `Go/all.zip`, **plus an omni-synthesized `manifest.json` inside the zip** so the one `.minisig` covers both. Determinism exactly like `pkg/sbom/format.Document`: sort entries by ID, zero archive timestamps, fixed compression, **byte-passthrough** of upstream JSON (no re-marshal). **Reject bbolt.** Keep the **fetch step out of the default runtime** (CI job or self-contained `contrib/` module produces + signs the zip; default `omni scan` is purely offline).

2. **Matching — pure-Go, as written.** Official OSV `IsVulnerable` (sorted-event walk OR'd with exact `versions[]`) over `format.Document` purls, using `golang.org/x/mod/semver`. Load-bearing: **prepend `v`** to both OSV and SBOM versions before compare; honor `fixed`(`>=`) vs `last_affected`(`>`); empty/`0` introduced = genesis; route ECOSYSTEM/GIT to `versions[]` membership. Validate only the 3 OSV-required fields; tolerate unknown fields.

3. **Severity — v3.1 only for v1.0.** Hand-roll the closed-form CVSS v3.1 base scorer (mind the v3.1 Roundup and PR-vs-Scope); map via the shared band table; prefer producer-supplied numeric score for v4.0 and **defer a full v4.0 scorer**. Golden-test v3.1 at band boundaries.

4. **Reachability — choose option (c) for v1.0.** `omni scan <source-dir>` returns `cmderr.ErrUnsupported` (exit 6) with a message pointing at the backlog. This keeps **every** non-negotiable intact (no exec, lean `go.mod`, pure-Go, golden-testable) with **zero** new deps. **Do NOT take option (a)** — build tags do not keep `x/vuln`+`x/tools v0.44.0` out of the main `go.mod` (MVS), and source mode execs `go list`; it fails both hard constraints (the Phase-04 sigstore / ADR-0007 lesson). When reachability eventually ships, it belongs in **option (b)**: a self-contained `contrib/govulncheck-scan` module (mirroring `contrib/sigstore-verify`), kept entirely out of the root graph, and **only** behind an explicit, documented sanctioned-`go list`-exec exception. Soften the plan's "eliminates false positives" to "substantially reduces."

5. **Plumbing — reuse existing fail-closed surface.** DB signature verification through `pkg/sign.Verify` (its `ErrVerification` maps to `cmderr.ErrConflict`); `--fail-on <severity>` → `cmderr.ErrConflict` (exit 1) for the CI gate; `--max-db-age` is a pure time comparison against the signed `manifest.generated`.

**Net:** the deterministic, pure-Go CORE (signed-OSV-DB + SBOM module/version matching + v3.1 severity gate, offline-first) ships in v1.0 satisfying all three hard constraints; reachability is explicitly backlogged to a future contrib module.
