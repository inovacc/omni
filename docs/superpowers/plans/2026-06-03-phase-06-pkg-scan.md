# Phase 06 — `pkg/scan/` Vulnerability Scanning Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.
> **HARD GATE:** Task 1 (ADR-0008) MUST be written AND human-reviewed/approved BEFORE any code task (2+) begins. The research spike concluded that `golang.org/x/vuln/scan` execs `go list` internally — this collides with omni's foundational NO-exec rule, so the architecture decision (pure-Go default + tagged reachability) must be recorded and approved first.

**Goal:** Ship `omni scan` — a pure-Go, fail-closed vulnerability scanner that matches an SBOM (`pkg/sbom/format.Document`) against a `pkg/sign`-signed OSV vulnerability database, gates CI on a `--fail-on <severity>` threshold (`cmderr.ErrConflict`), works offline with a `--max-db-age` staleness gate, and offers reachability-aware Go-source scanning behind an opt-in `//go:build omni_govulncheck` tag (default build returns `cmderr.ErrUnsupported`).

**Architecture:** Pure-Go from the ground up. `pkg/scan/` imports `pkg/sbom/format.Document` ONLY (never `pkg/sbom/model`), `pkg/sign` (DB verify), and `golang.org/x/mod/semver` (range matching — already a Phase 5 dep). The default matcher is purely version-range based over a local OSV DB: zero exec, fully deterministic, golden-testable. Reachability analysis (which requires `go list` + package load + SSA call-graph and therefore exec) is isolated behind `//go:build omni_govulncheck`; without the tag, `omni scan source` returns `ErrUnsupported`. The OSV DB is a signed bundle (a `.zip` of OSV JSON entries + a manifest), verified with `pkg/sign.Verify` on load — a tampered DB fails closed. Live OSV-API enrichment (`--online`, `net/http` only — never exec) is opt-in; tests always use a committed fixture DB. Layering: pure lib `pkg/scan/` → I/O glue `internal/cli/scan/` → thin Cobra wrapper `cmd/scan.go`.

**Tech Stack:** Go stdlib (`archive/zip`, `encoding/json`, `net/http`, `time`, `io/fs`, `log/slog`); `golang.org/x/mod/semver` (semver range matching, Phase-5 dep); `github.com/inovacc/omni/pkg/sign` (Phase 4, DB signature verify); `github.com/inovacc/omni/pkg/sbom/format` (Phase 5 boundary type); optional `golang.org/x/vuln` (reachability, build-tag `omni_govulncheck` only); Cobra; the Python YAML golden harness.

**Repo conventions (from research, cite when implementing):**
- Commands self-wire: `cmd/scan.go` declares `var scanCmd = &cobra.Command{...}` and calls `rootCmd.AddCommand(scanCmd)` in `init()`; `RunE` reads flags → Options → calls `internal/cli/scan.RunScan(cmd.OutOrStdout(), …)`. No central registration list. Subcommands (`scan source`, `scan db update`) attach via `scanCmd.AddCommand(...)`.
- `cmderr` (`internal/cli/cmderr/cmderr.go`): findings meet/exceed `--fail-on` → `cmderr.Wrap(cmderr.ErrConflict, …)` (exit 1); missing DB / SBOM file → `ErrNotFound` (1); unreadable → `ErrPermission` (3); bad flags / unparseable SBOM or DB JSON → `ErrInvalidInput` (2); DB signature mismatch / stale-DB-over-`--max-db-age` → `ErrConflict` (1, fail closed); `scan source` without `omni_govulncheck` tag → `ErrUnsupported` (6); network failure in `--online` → `ErrIO` (4). `Wrap(sentinel, msg)` = `fmt.Errorf("%s: %w", msg, sentinel)`; `Is<Class>()` predicates exist; use `errors.Is`/`As`, never `==`.
- Layering: pure lib `pkg/scan/` (stdlib + `x/mod/semver` + `pkg/sign` + `pkg/sbom/format`; NO cobra, NO io.Writer-for-output) + I/O glue `internal/cli/scan/` + thin `cmd/scan.go`. Every package gets a `doc.go`. `pkg/scan` is a NEW v1.0 surface — mark its `doc.go` `// Experimental:` until the API stabilizes (per pkg/* API triage convention).
- Golden harness: Python+YAML, TWO files kept in sync (`testing/golden/golden_tests.yaml` + `tools/golden/golden_tests.yaml`); negative tests set `exit_code:` + `normalizations: ["strip_path"]` annotated with the cmderr sentinel; committed fixtures via the `{fixtures}` placeholder → `testing/golden/fixtures/<category>/`; regenerate with `task test:golden:update` then `task golden:record`. Confirmed example: the `sign` category at `golden_tests.yaml:1091` uses `args: ["verify", "--key", "{fixtures}/test.pub", …]`.
- ADRs live in `docs/adr/` as `ADR-NNNN-kebab-title.md`; 0001–0006 are used, so the next number is **0007**; header format per `docs/adr/ADR-0004-internalize-cobra-cli.md`.
- Pipe: register stdin→stdout commands in `cmd/pipe.go buildPipeRegistry()` via `command.AdaptWriterReaderArgs(...)`.
- INVARIANTS: pure-Go, NO `os/exec`, no CGO; cross-platform via `//go:build` tags (never runtime `os ==`); `io.Writer`/`io.Reader`; deferred `Close`.
- **Phase 5 boundary (depends-on):** `pkg/sbom/format.Document` MUST exist before Task 4. Its component-bearing field is required for matching. This plan assumes the shape `Document{ Components []Component }` where `Component{ Name, Version, PURL string; Ecosystem string }` (purl like `pkg:golang/github.com/x/y`). **If Phase 5 lands a different shape, adjust Task 3's `format` import and the `componentsOf` adapter ONLY — the matcher logic is shape-agnostic because it consumes the normalized `pkg`/`version` pair the adapter emits.**
- **Phase 4 reuse:** `pkg/sign` exposes `Sign(data []byte, sk SecretKey, opts ...Option) ([]byte, error)`, `Verify(data []byte, sig []byte, pub PublicKey) error`, `ParsePublicKey(text []byte) (PublicKey, error)`. The DB bundle is signed/verified through these exact signatures. Heavy optional deps (`golang.org/x/vuln`) bump go.mod via MVS even when tag-gated (Phase 4 finding) — accept it ONLY in the tagged file; the default path stays on `x/mod/semver`.

---

## OSV database bundle format (authoritative — implement byte-exactly)

The signed OSV DB omni consumes is a **single `.zip` file** (`osv-db.zip`) plus a detached **`.minisig`** (`osv-db.zip.minisig`, produced by `omni sign`). The verifier loads the zip ONLY after `pkg/sign.Verify` passes over the zip bytes.

Zip layout (all members UTF-8):
- `manifest.json` (required, exactly one): `{"schema_version":"1.0","generated":"<RFC3339 UTC>","ecosystem":"Go","entry_count":<int>}`. `generated` is the freshness timestamp the `--max-db-age` gate reads.
- `entries/<OSV-ID>.json` (zero or more): one OSV record per file, each conforming to OSV schema 1.7.5 (fields used by omni: `id`, `summary`, `details`, `modified`, `severity[].{type,score}`, `affected[].package.{ecosystem,name,purl}`, `affected[].ranges[].{type,events[].{introduced,fixed}}`, `affected[].versions[]`, `affected[].ecosystem_specific.imports[].{path,symbols[]}` — Go-vuln symbol data, used only by the reachability path).

**Matching algorithm (default, pure-Go, no exec — implement exactly):**
For each SBOM component `(pkg, version)` where `pkg` is the Go MODULE path (normalized from the purl) and `version` is its semver (with a leading `v` for `x/mod/semver`):
1. Load every OSV entry whose `affected[].package.ecosystem == "Go"` and `affected[].package.name == pkg`.
2. For each matching `affected`, decide "is `version` vulnerable?":
   - If an `affected[].versions` list is present and contains `version` exactly → **vulnerable**.
   - Else evaluate `affected[].ranges[]` of `type == "SEMVER"` as a sorted event timeline: walk `events` in order; `introduced:"0"` (or any `introduced`) opens an interval, the next `fixed` closes it. `version` is vulnerable iff it falls in an open `[introduced, fixed)` interval, compared with `semver.Compare("v"+version, "v"+bound)`. An `introduced` with no later `fixed` means "all versions ≥ introduced".
   - `type == "ECOSYSTEM"` ranges fall back to exact `versions` membership (we do not interpret arbitrary ecosystem ordering).
3. A match yields a `Finding{ID, Package, Version, FixedVersion, Severity, Summary}` where `Severity` is the normalized label from `severityLabel(...)` (below) and `FixedVersion` is the smallest `fixed` bound > `version`, or `""` if none.

**Severity normalization (`severityLabel`) — deterministic, no external CVSS lib:**
Pick the first `severity[]` of type `CVSS_V3` or `CVSS_V4`; parse the base score from the vector's CVSS qualitative band using the canonical CVSS v3.1/v4.0 cutoffs applied to the numeric base score WHEN a numeric `score` is present; when only a vector string is present, fall back to the band from the vector's computed base. To keep this pure and dependency-free, omni implements a tiny CVSS base-score parser limited to what it needs:
- If `score` parses as a float (some DBs put the number there) → band it.
- Else compute the base score from the CVSS vector metrics (AV/AC/PR/UI/S/C/I/A for v3; AV/AC/AT/PR/UI/VC/VI/VA/SC/SI/SA for v4) using the published formulas.
- Bands (CVSS v3.1/v4.0, canonical): `0.0`→`none`, `0.1–3.9`→`low`, `4.0–6.9`→`medium`, `7.0–8.9`→`high`, `9.0–10.0`→`critical`.
- If NO usable `severity[]` exists → `unknown` (treated as below `low` for `--fail-on`, but always reported).
Ordered constants: `SeverityUnknown < SeverityNone < SeverityLow < SeverityMedium < SeverityHigh < SeverityCritical`. `--fail-on <label>` fails (ErrConflict) iff any finding's severity `>=` the threshold.

---

### Task 1 (ADR GATE): ADR-0008 — no-exec scan architecture, signed-DB trust, reachability tag

**Files:** Create `docs/adr/ADR-0008-pure-go-vuln-scan-and-signed-osv-db.md`

- [ ] **Step 1: Write the ADR** matching the `ADR-0004` header/section format (`# ADR-0008: …`, `**Status:** Accepted`, `**Date:** 2026-06-03`, `**Decision:** …`, then `## Context`, `## Analysis` (table), `## Consequences`). Record + justify the decisions the research spike forced:
  - **Research-spike conclusion (the crux):** `golang.org/x/vuln/scan.Cmd` is documented as "similar to `exec.Cmd`" and internally shells out to `go list` / the build system. Using it in the default binary would VIOLATE omni's foundational NO-exec rule. Decision: the **default scanner is a pure-Go version-range matcher** over a local OSV DB (zero exec, deterministic, golden-testable); **reachability** (`scan source`) is opt-in behind `//go:build omni_govulncheck` and returns `cmderr.ErrUnsupported` without the tag — mirroring Phase 4's `omni_sigstore` isolation.
  - **OSV DB distribution = a single signed `.zip` bundle** (flat JSON entries + `manifest.json`), NOT bbolt — keeps it diffable, language-agnostic, and `archive/zip`-loadable with no new deps. Verified with `pkg/sign.Verify` on load (Phase 4 reuse); tampered DB fails closed → `ErrConflict`.
  - **Offline-first:** the matcher needs only the local bundle. `--online` (OSV API enrichment over `net/http`) is opt-in and never the test path. `--max-db-age` reads `manifest.generated`; a DB older than the gate fails LOUDLY (`ErrConflict`), never silently degrades.
  - **Severity:** omni ships a tiny pure-Go CVSS v3.1/v4.0 base-score band parser (no external CVSS lib — avoids MVS bloat per Phase 4 finding); `unknown` when no usable severity, always reported, below `low` for gating.
  - **Boundary:** `pkg/scan` imports `pkg/sbom/format.Document` ONLY, never `pkg/sbom/model` (architectural invariant from the spec).
  - **MVS note:** `golang.org/x/vuln` bumps `go.mod` even when tag-gated; confine it to the tagged file. Default path deps: stdlib + `x/mod/semver` + `pkg/sign`.
- [ ] **Step 2: Stop for human review.** Do NOT proceed to any code task until this ADR is approved.

---

### Task 2: `pkg/scan` core types + severity model (TDD)

**Files:** Create `pkg/scan/scan.go`, `pkg/scan/severity.go`, `pkg/scan/severity_test.go`, `pkg/scan/doc.go`.

- [ ] **Step 1: Define stable-ish public types** in `pkg/scan/scan.go` (this is a new EXPERIMENTAL surface — see `doc.go`):

```go
package scan

// Severity is an ordered vulnerability severity label.
type Severity int

const (
	SeverityUnknown Severity = iota // no usable CVSS data; always reported, below Low for gating
	SeverityNone
	SeverityLow
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// String returns the lowercase label (e.g. "high").
func (s Severity) String() string {
	switch s {
	case SeverityNone:
		return "none"
	case SeverityLow:
		return "low"
	case SeverityMedium:
		return "medium"
	case SeverityHigh:
		return "high"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// ParseSeverity maps a label to a Severity; unknown text yields (SeverityUnknown, false).
func ParseSeverity(label string) (Severity, bool) {
	switch label {
	case "none":
		return SeverityNone, true
	case "low":
		return SeverityLow, true
	case "medium":
		return SeverityMedium, true
	case "high":
		return SeverityHigh, true
	case "critical":
		return SeverityCritical, true
	default:
		return SeverityUnknown, false
	}
}

// Finding is one vulnerability matched against one component.
type Finding struct {
	ID           string   `json:"id"`             // OSV id, e.g. GO-2023-1234
	Package      string   `json:"package"`        // Go module path
	Version      string   `json:"version"`        // affected version present in the SBOM
	FixedVersion string   `json:"fixed_version"`  // smallest fixed > Version, or "" if none
	Severity     string   `json:"severity"`       // Severity.String()
	Summary      string   `json:"summary"`        // OSV summary
	Reachable    *bool    `json:"reachable,omitempty"` // set only by the reachability path
}

// Report is the full scan result, sorted deterministically by (Package, ID).
type Report struct {
	Findings []Finding `json:"findings"`
	Scanned  int       `json:"scanned"`  // components examined
	DBAge    string    `json:"db_age"`   // human duration since manifest.generated
}
```

- [ ] **Step 2: Write failing severity test** (`pkg/scan/severity_test.go`):

```go
package scan

import "testing"

func TestSeverityLabelBands(t *testing.T) {
	cases := []struct {
		name  string
		sev   []rawSeverity
		want  Severity
	}{
		{"critical numeric", []rawSeverity{{Type: "CVSS_V3", Score: "9.8"}}, SeverityCritical},
		{"high vector v3", []rawSeverity{{Type: "CVSS_V3", Score: "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H"}}, SeverityCritical},
		{"medium numeric", []rawSeverity{{Type: "CVSS_V3", Score: "5.3"}}, SeverityMedium},
		{"low numeric", []rawSeverity{{Type: "CVSS_V3", Score: "3.1"}}, SeverityLow},
		{"none numeric", []rawSeverity{{Type: "CVSS_V3", Score: "0.0"}}, SeverityNone},
		{"prefers v4 over v2", []rawSeverity{{Type: "CVSS_V2", Score: "5.0"}, {Type: "CVSS_V4", Score: "9.3"}}, SeverityCritical},
		{"empty -> unknown", nil, SeverityUnknown},
		{"unparseable -> unknown", []rawSeverity{{Type: "Ubuntu", Score: "high"}}, SeverityUnknown},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := severityLabel(c.sev); got != c.want {
				t.Errorf("severityLabel(%v) = %v, want %v", c.sev, got, c.want)
			}
		})
	}
}

func TestSeverityOrderingForGating(t *testing.T) {
	if !(SeverityUnknown < SeverityLow && SeverityLow < SeverityMedium &&
		SeverityMedium < SeverityHigh && SeverityHigh < SeverityCritical) {
		t.Fatal("severity ordering broken — --fail-on gating depends on it")
	}
}
```

- [ ] **Step 3: Run, verify fail:** `go test ./pkg/scan/ -run Severity -v` → FAIL (`rawSeverity`, `severityLabel` undefined).
- [ ] **Step 4: Implement** `pkg/scan/severity.go`: define `type rawSeverity struct{ Type, Score string }`; `func severityLabel(sevs []rawSeverity) Severity` that (1) picks the first `CVSS_V4` else first `CVSS_V3`; (2) `if f, err := strconv.ParseFloat(s.Score, 64); err == nil { return band(f) }`; (3) else `if base, ok := cvssBaseScore(s.Type, s.Score); ok { return band(base) }`; (4) else `SeverityUnknown`. `band(f)` applies the canonical cutoffs from the format section. `cvssBaseScore` implements the published CVSS v3.1 and v4.0 base-score formulas from the vector metrics (parse `AV/AC/PR/UI/S/C/I/A` for v3, `AV/AC/AT/PR/UI/VC/VI/VA/SC/SI/SA` for v4; on any missing metric return `(0, false)`). Keep it ~120 lines, stdlib-only.
- [ ] **Step 5: Run, verify pass:** `go test ./pkg/scan/ -run Severity -v` → PASS.
- [ ] **Step 6: Create `pkg/scan/doc.go`** with the package docstring. First line of the doc comment MUST be `// Experimental: pkg/scan is a new v1.0 surface; its API may change before it is marked stable.` (per pkg/* API-triage convention).
- [ ] **Step 7: Commit:** `gofmt -w pkg/scan/ && git commit -- pkg/scan/ -m "feat(scan): pkg/scan core types + pure-Go CVSS severity banding"`

---

### Task 3: OSV entry model + version-range matcher (TDD)

**Files:** Create `pkg/scan/osv.go`, `pkg/scan/match.go`, `pkg/scan/match_test.go`. Modify `go.mod`/`go.sum` (ensure `golang.org/x/mod` is a direct dep).

- [ ] **Step 1: Write the failing matcher test** (`pkg/scan/match_test.go`):

```go
package scan

import "testing"

func entry(id, name string, versions []string, ranges []rng, sevs []rawSeverity) osvEntry {
	return osvEntry{
		ID: id,
		Severity: sevs,
		Affected: []osvAffected{{
			Package:  osvPackage{Ecosystem: "Go", Name: name},
			Versions: versions,
			Ranges:   ranges,
		}},
	}
}

func TestMatchSemverRange(t *testing.T) {
	e := entry("GO-1", "github.com/foo/bar", nil,
		[]rng{{Type: "SEMVER", Events: []rngEvent{{Introduced: "0"}, {Fixed: "1.2.3"}}}},
		[]rawSeverity{{Type: "CVSS_V3", Score: "7.5"}})

	cases := map[string]struct {
		version string
		want    bool
		fixed   string
	}{
		"vulnerable below fixed": {"1.0.0", true, "1.2.3"},
		"safe at fixed":          {"1.2.3", false, ""},
		"safe above fixed":       {"2.0.0", false, ""},
		"vulnerable at zero":     {"0.0.1", true, "1.2.3"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			f, ok := matchEntry(e, "github.com/foo/bar", c.version)
			if ok != c.want {
				t.Fatalf("matchEntry ok = %v, want %v", ok, c.want)
			}
			if ok && f.FixedVersion != c.fixed {
				t.Errorf("FixedVersion = %q, want %q", f.FixedVersion, c.fixed)
			}
			if ok && f.Severity != "high" {
				t.Errorf("Severity = %q, want high", f.Severity)
			}
		})
	}
}

func TestMatchOpenEndedRange(t *testing.T) {
	e := entry("GO-2", "github.com/foo/bar", nil,
		[]rng{{Type: "SEMVER", Events: []rngEvent{{Introduced: "1.5.0"}}}}, nil)
	if _, ok := matchEntry(e, "github.com/foo/bar", "2.0.0"); !ok {
		t.Error("2.0.0 must be vulnerable (introduced 1.5.0, no fix)")
	}
	if _, ok := matchEntry(e, "github.com/foo/bar", "1.4.0"); ok {
		t.Error("1.4.0 must be safe (below introduced)")
	}
}

func TestMatchExactVersionsList(t *testing.T) {
	e := entry("GO-3", "github.com/foo/bar", []string{"1.0.0", "1.0.1"}, nil, nil)
	if _, ok := matchEntry(e, "github.com/foo/bar", "1.0.1"); !ok {
		t.Error("1.0.1 in versions list must match")
	}
	if _, ok := matchEntry(e, "github.com/foo/bar", "1.0.2"); ok {
		t.Error("1.0.2 not in versions list must not match")
	}
}

func TestMatchWrongPackageOrEcosystem(t *testing.T) {
	e := entry("GO-4", "github.com/foo/bar", []string{"1.0.0"}, nil, nil)
	if _, ok := matchEntry(e, "github.com/other/pkg", "1.0.0"); ok {
		t.Error("different package must not match")
	}
	e.Affected[0].Package.Ecosystem = "npm"
	if _, ok := matchEntry(e, "github.com/foo/bar", "1.0.0"); ok {
		t.Error("non-Go ecosystem must not match")
	}
}
```

- [ ] **Step 2: Run, verify fail:** `go test ./pkg/scan/ -run Match -v` → FAIL (`osvEntry`, `matchEntry`, etc. undefined).
- [ ] **Step 3: Implement OSV model** in `pkg/scan/osv.go` (json-tagged, only the fields used):

```go
package scan

type osvEntry struct {
	ID       string        `json:"id"`
	Summary  string        `json:"summary"`
	Details  string        `json:"details"`
	Modified string        `json:"modified"`
	Severity []rawSeverity `json:"severity"`
	Affected []osvAffected `json:"affected"`
}

type osvAffected struct {
	Package           osvPackage         `json:"package"`
	Severity          []rawSeverity      `json:"severity"`
	Ranges            []rng              `json:"ranges"`
	Versions          []string           `json:"versions"`
	EcosystemSpecific osvEcosystemSpec   `json:"ecosystem_specific"`
}

type osvPackage struct {
	Ecosystem string `json:"ecosystem"`
	Name      string `json:"name"`
	PURL      string `json:"purl"`
}

type rng struct {
	Type   string     `json:"type"`
	Events []rngEvent `json:"events"`
}

type rngEvent struct {
	Introduced string `json:"introduced"`
	Fixed      string `json:"fixed"`
}

// osvEcosystemSpec carries Go-vuln symbol data used only by the reachability path.
type osvEcosystemSpec struct {
	Imports []osvImport `json:"imports"`
}

type osvImport struct {
	Path    string   `json:"path"`
	Symbols []string `json:"symbols"`
}
```

- [ ] **Step 4: Implement the matcher** in `pkg/scan/match.go`:

```go
package scan

import "golang.org/x/mod/semver"

// sv normalizes an SBOM version to an x/mod/semver-comparable string (leading v).
func sv(v string) string {
	if v == "" {
		return ""
	}
	if v[0] == 'v' {
		return v
	}
	return "v" + v
}

// matchEntry reports whether (pkg, version) is hit by entry e, and builds the Finding.
// Per the format section: ecosystem must be "Go", name must equal pkg; then versions
// list membership OR an open SEMVER [introduced, fixed) interval makes it vulnerable.
func matchEntry(e osvEntry, pkg, version string) (Finding, bool) {
	for _, a := range e.Affected {
		if a.Package.Ecosystem != "Go" || a.Package.Name != pkg {
			continue
		}
		hit, fixed := affectedHit(a, version)
		if !hit {
			continue
		}
		sev := e.Severity
		if len(a.Severity) > 0 {
			sev = a.Severity // affected-level severity overrides top-level
		}
		return Finding{
			ID:           e.ID,
			Package:      pkg,
			Version:      version,
			FixedVersion: fixed,
			Severity:     severityLabel(sev).String(),
			Summary:      e.Summary,
		}, true
	}
	return Finding{}, false
}

func affectedHit(a osvAffected, version string) (bool, string) {
	for _, v := range a.Versions { // exact-membership shortcut
		if v == version {
			return true, smallestFixAbove(a, version)
		}
	}
	for _, r := range a.Ranges {
		if r.Type != "SEMVER" {
			continue // ECOSYSTEM handled by exact versions only
		}
		if inOpenInterval(r.Events, version) {
			return true, smallestFixAbove(a, version)
		}
	}
	return false, ""
}

// inOpenInterval walks ordered events; introduced opens, fixed closes [introduced, fixed).
func inOpenInterval(events []rngEvent, version string) bool {
	open := false
	cur := sv(version)
	for _, ev := range events {
		switch {
		case ev.Introduced != "":
			lo := ev.Introduced
			if lo == "0" {
				open = true
				continue
			}
			open = semver.Compare(cur, sv(lo)) >= 0
		case ev.Fixed != "":
			if open && semver.Compare(cur, sv(ev.Fixed)) < 0 {
				return true
			}
			open = false
		}
	}
	return open // introduced with no later fixed => all >= introduced
}

// smallestFixAbove returns the smallest fixed bound strictly greater than version, or "".
func smallestFixAbove(a osvAffected, version string) string {
	best := ""
	cur := sv(version)
	for _, r := range a.Ranges {
		for _, ev := range r.Events {
			if ev.Fixed == "" {
				continue
			}
			if semver.Compare(sv(ev.Fixed), cur) > 0 {
				if best == "" || semver.Compare(sv(ev.Fixed), sv(best)) < 0 {
					best = ev.Fixed
				}
			}
		}
	}
	return best
}
```

- [ ] **Step 5: Run, verify pass:** `go test ./pkg/scan/ -run Match -v` → PASS. Run `go get golang.org/x/mod && go mod tidy` if `x/mod` is not yet a direct dep.
- [ ] **Step 6: Commit:** `gofmt -w pkg/scan/ && git commit -- pkg/scan/ go.mod go.sum -m "feat(scan): OSV entry model + pure-Go semver range matcher"`

---

### Task 4: signed OSV DB bundle — load + verify + staleness gate (TDD)

**Files:** Create `pkg/scan/db.go`, `pkg/scan/db_test.go`.

- [ ] **Step 1: Write failing tests** (`pkg/scan/db_test.go`) — the DB loads from a zip, exposes entries by package, and reports its age; tampering and bad zip fail closed:

```go
package scan

import (
	"archive/zip"
	"bytes"
	"testing"
	"time"
)

// buildZip assembles an in-memory osv-db.zip with one manifest + the given entry JSONs.
func buildZip(t *testing.T, generated time.Time, entries map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	mw, _ := zw.Create("manifest.json")
	_, _ = mw.Write([]byte(`{"schema_version":"1.0","generated":"` +
		generated.UTC().Format(time.RFC3339) + `","ecosystem":"Go","entry_count":` +
		itoa(len(entries)) + `}`))
	for name, body := range entries {
		ew, _ := zw.Create("entries/" + name)
		_, _ = ew.Write([]byte(body))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestLoadDBAndAge(t *testing.T) {
	gen := time.Now().Add(-48 * time.Hour)
	z := buildZip(t, gen, map[string]string{
		"GO-1.json": `{"id":"GO-1","summary":"x","affected":[{"package":{"ecosystem":"Go","name":"github.com/foo/bar"},"ranges":[{"type":"SEMVER","events":[{"introduced":"0"},{"fixed":"1.2.3"}]}]}]}`,
	})
	db, err := LoadDB(bytes.NewReader(z), int64(len(z)))
	if err != nil {
		t.Fatalf("LoadDB: %v", err)
	}
	if got := db.entriesFor("github.com/foo/bar"); len(got) != 1 {
		t.Fatalf("entriesFor = %d, want 1", len(got))
	}
	if db.Age() < 47*time.Hour {
		t.Errorf("Age too small: %v", db.Age())
	}
}

func TestStalenessGate(t *testing.T) {
	z := buildZip(t, time.Now().Add(-10*24*time.Hour), nil)
	db, err := LoadDB(bytes.NewReader(z), int64(len(z)))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.CheckFresh(7 * 24 * time.Hour); err == nil {
		t.Fatal("CheckFresh must FAIL for a DB older than max age (fail loud)")
	}
	if err := db.CheckFresh(30 * 24 * time.Hour); err != nil {
		t.Errorf("CheckFresh(30d) = %v, want nil", err)
	}
}

func TestLoadDBBadZipFailsClosed(t *testing.T) {
	if _, err := LoadDB(bytes.NewReader([]byte("not a zip")), 9); err == nil {
		t.Fatal("LoadDB on garbage must fail closed")
	}
}
```

(Add a tiny `itoa` test helper, or use `strconv.Itoa` inline in `buildZip`.)

- [ ] **Step 2: Run, verify fail:** `go test ./pkg/scan/ -run "DB|Staleness" -v` → FAIL (`LoadDB`, `DB`, etc. undefined).
- [ ] **Step 3: Implement `pkg/scan/db.go`:**

```go
package scan

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

// ErrStaleDB is returned by CheckFresh when the DB is older than the allowed age.
var ErrStaleDB = errors.New("vulnerability database is stale")

type dbManifest struct {
	SchemaVersion string `json:"schema_version"`
	Generated     string `json:"generated"`
	Ecosystem     string `json:"ecosystem"`
	EntryCount    int    `json:"entry_count"`
}

// DB is a loaded OSV database, indexed by Go module path.
type DB struct {
	generated time.Time
	byPkg     map[string][]osvEntry
}

// LoadDB parses an osv-db.zip from r (of size n). It fails closed on any malformed input.
// Signature verification happens BEFORE this call (see VerifyAndLoadDB).
func LoadDB(r io.ReaderAt, n int64) (*DB, error) {
	zr, err := zip.NewReader(r, n)
	if err != nil {
		return nil, fmt.Errorf("open osv db zip: %w", err)
	}
	db := &DB{byPkg: map[string][]osvEntry{}}
	var sawManifest bool
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", f.Name, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", f.Name, err)
		}
		switch {
		case f.Name == "manifest.json":
			var m dbManifest
			if err := json.Unmarshal(data, &m); err != nil {
				return nil, fmt.Errorf("parse manifest: %w", err)
			}
			t, err := time.Parse(time.RFC3339, m.Generated)
			if err != nil {
				return nil, fmt.Errorf("parse manifest.generated: %w", err)
			}
			db.generated = t
			sawManifest = true
		case len(f.Name) > 8 && f.Name[:8] == "entries/":
			var e osvEntry
			if err := json.Unmarshal(data, &e); err != nil {
				return nil, fmt.Errorf("parse %s: %w", f.Name, err)
			}
			for _, a := range e.Affected {
				if a.Package.Ecosystem == "Go" && a.Package.Name != "" {
					db.byPkg[a.Package.Name] = append(db.byPkg[a.Package.Name], e)
				}
			}
		}
	}
	if !sawManifest {
		return nil, errors.New("osv db missing manifest.json")
	}
	return db, nil
}

func (d *DB) entriesFor(pkg string) []osvEntry { return d.byPkg[pkg] }

// Age returns how long ago the DB was generated.
func (d *DB) Age() time.Duration { return time.Since(d.generated) }

// CheckFresh fails loudly (ErrStaleDB) when the DB is older than maxAge; maxAge<=0 disables the gate.
func (d *DB) CheckFresh(maxAge time.Duration) error {
	if maxAge <= 0 {
		return nil
	}
	if d.Age() > maxAge {
		return fmt.Errorf("%w: generated %s ago (max %s)", ErrStaleDB, d.Age().Round(time.Second), maxAge)
	}
	return nil
}
```

- [ ] **Step 4: Run, verify pass:** `go test ./pkg/scan/ -run "DB|Staleness" -v` → PASS.
- [ ] **Step 5: Add signature-verifying loader** to `pkg/scan/db.go` (TDD it with a fixture-key round-trip):

```go
import "github.com/inovacc/omni/pkg/sign"

// VerifyAndLoadDB verifies the detached minisig over zipBytes with pub, then loads the DB.
// Verification failure is fail-closed: the DB is never loaded on a bad signature.
func VerifyAndLoadDB(zipBytes, minisig, pubKeyText []byte) (*DB, error) {
	pub, err := sign.ParsePublicKey(pubKeyText)
	if err != nil {
		return nil, fmt.Errorf("parse db public key: %w", err)
	}
	if err := sign.Verify(zipBytes, minisig, pub); err != nil {
		return nil, fmt.Errorf("osv db signature verification failed: %w", err)
	}
	return LoadDB(bytesReaderAt(zipBytes), int64(len(zipBytes)))
}
```

   Add a small `bytesReaderAt([]byte) io.ReaderAt` helper (`bytes.NewReader` already implements `io.ReaderAt`, so this is just `bytes.NewReader(b)`). Add a test `TestVerifyAndLoadDBFailsClosed` that signs a fixture zip with a low-cost keypair (`sign.GenerateKeyPair("p", sign.WithScryptParams(1<<15, 8, 1))`), confirms `VerifyAndLoadDB` succeeds, then flips a byte in `zipBytes` and confirms it returns a non-nil error WITHOUT loading.
- [ ] **Step 6: Run, verify pass:** `go test ./pkg/scan/ -run "DB|Staleness|Verify" -v` → PASS.
- [ ] **Step 7: Commit:** `gofmt -w pkg/scan/ && git commit -- pkg/scan/ go.mod go.sum -m "feat(scan): signed OSV DB bundle loader + staleness gate (fail-closed)"`

---

### Task 5: top-level `Scan` over an SBOM Document + `--fail-on` gate (TDD)

**Files:** Create `pkg/scan/document.go`, `pkg/scan/document_test.go`. Modify `pkg/scan/scan.go` (add `Scan`, `Options`).

> **Phase-5 boundary contract:** this is the ONLY file that imports `github.com/inovacc/omni/pkg/sbom/format`. It adapts `format.Document` → a slice of normalized `(pkg, version)` pairs via `componentsOf`. If Phase 5's `format.Document` field names differ, fix `componentsOf` ONLY.

- [ ] **Step 1: Write failing test** (`pkg/scan/document_test.go`) using a local fake that mirrors the assumed `format.Document` shape, so the matcher logic is tested independent of Phase 5 timing:

```go
package scan

import "testing"

func TestScanGatesOnFailOn(t *testing.T) {
	// two components: one vulnerable (high), one clean.
	comps := []component{
		{Pkg: "github.com/foo/bar", Version: "1.0.0"},
		{Pkg: "github.com/safe/lib", Version: "2.0.0"},
	}
	db := &DB{byPkg: map[string][]osvEntry{
		"github.com/foo/bar": {entry("GO-1", "github.com/foo/bar", nil,
			[]rng{{Type: "SEMVER", Events: []rngEvent{{Introduced: "0"}, {Fixed: "1.2.3"}}}},
			[]rawSeverity{{Type: "CVSS_V3", Score: "7.5"}})},
	}}

	rep, err := scanComponents(comps, db, Options{})
	if err != nil {
		t.Fatalf("scanComponents(no gate): %v", err)
	}
	if len(rep.Findings) != 1 || rep.Findings[0].ID != "GO-1" {
		t.Fatalf("findings = %+v, want 1 GO-1", rep.Findings)
	}
	if rep.Scanned != 2 {
		t.Errorf("Scanned = %d, want 2", rep.Scanned)
	}

	// gate at high -> the high finding trips ErrConflict.
	if _, err := scanComponents(comps, db, Options{FailOn: SeverityHigh}); err == nil {
		t.Error("FailOn high must return ErrConflict for a high finding")
	}
	// gate at critical -> high finding is below threshold -> no error.
	if _, err := scanComponents(comps, db, Options{FailOn: SeverityCritical}); err != nil {
		t.Errorf("FailOn critical with only a high finding = %v, want nil", err)
	}
}

func TestScanDeterministicOrder(t *testing.T) {
	comps := []component{{Pkg: "github.com/foo/bar", Version: "1.0.0"}}
	db := &DB{byPkg: map[string][]osvEntry{
		"github.com/foo/bar": {
			entry("GO-9", "github.com/foo/bar", []string{"1.0.0"}, nil, nil),
			entry("GO-1", "github.com/foo/bar", []string{"1.0.0"}, nil, nil),
		},
	}}
	rep, _ := scanComponents(comps, db, Options{})
	if len(rep.Findings) != 2 || rep.Findings[0].ID != "GO-1" || rep.Findings[1].ID != "GO-9" {
		t.Fatalf("findings not sorted by (pkg,id): %+v", rep.Findings)
	}
}
```

- [ ] **Step 2: Run, verify fail:** `go test ./pkg/scan/ -run Scan -v` → FAIL.
- [ ] **Step 3: Implement the boundary + scan** in `pkg/scan/document.go`:

```go
package scan

import (
	"errors"
	"sort"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/sbom/format"
)

// component is the normalized (module path, version) pair the matcher consumes.
type component struct {
	Pkg     string
	Version string
}

// Options configures a scan.
type Options struct {
	FailOn   Severity // findings >= FailOn trip ErrConflict; SeverityUnknown (0) disables the gate
	Reachability bool  // request reachability filtering (only honored under -tags omni_govulncheck)
}

// componentsOf adapts the SBOM boundary type into normalized components.
// THIS is the single point of coupling to pkg/sbom/format.
func componentsOf(doc format.Document) []component {
	out := make([]component, 0, len(doc.Components))
	for _, c := range doc.Components {
		pkg := modulePathFromComponent(c)
		if pkg == "" || c.Version == "" {
			continue
		}
		out = append(out, component{Pkg: pkg, Version: c.Version})
	}
	return out
}

// modulePathFromComponent prefers the Go module path; derives it from purl if needed.
func modulePathFromComponent(c format.Component) string {
	if c.Name != "" {
		return c.Name
	}
	return modulePathFromPURL(c.PURL) // strips "pkg:golang/" prefix and any @version
}

// Scan scans an SBOM Document against db and applies opts.
func Scan(doc format.Document, db *DB, opts Options) (Report, error) {
	return scanComponents(componentsOf(doc), db, opts)
}

func scanComponents(comps []component, db *DB, opts Options) (Report, error) {
	var findings []Finding
	for _, c := range comps {
		for _, e := range db.entriesFor(c.Pkg) {
			if f, ok := matchEntry(e, c.Pkg, c.Version); ok {
				findings = append(findings, f)
			}
		}
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Package != findings[j].Package {
			return findings[i].Package < findings[j].Package
		}
		return findings[i].ID < findings[j].ID
	})
	rep := Report{Findings: findings, Scanned: len(comps), DBAge: db.Age().Round(1e9).String()}

	if opts.FailOn > SeverityUnknown {
		for _, f := range findings {
			sev, _ := ParseSeverity(f.Severity)
			if sev >= opts.FailOn {
				return rep, cmderr.Wrap(cmderr.ErrConflict,
					"vulnerabilities found at or above --fail-on threshold")
			}
		}
	}
	return rep, nil
}
```

   Add `modulePathFromPURL(string) string` to `pkg/scan/match.go`: trim a leading `pkg:golang/`, drop anything from the first `@`, and URL-decode (use `net/url.PathUnescape`); return `""` if it is not a `pkg:golang` purl. Note `cmderr` is an `internal/` package — importing it from `pkg/scan` is acceptable here because `pkg/scan` is itself part of the omni module (the same pattern the spec/Phase-4 plan use for sentinel-returning library helpers); the `Wrap` keeps `errors.Is(err, cmderr.ErrConflict)` true for the CLI. Keep the unused `errors` import only if referenced; otherwise omit it.
- [ ] **Step 4: Run, verify pass:** `go test ./pkg/scan/ -v` → PASS (full package).
- [ ] **Step 5: Commit:** `gofmt -w pkg/scan/ && git commit -- pkg/scan/ -m "feat(scan): Scan over SBOM Document with deterministic findings + --fail-on gate"`

---

### Task 6: reachability split — default stub vs `omni_govulncheck` tag (TDD)

**Files:** Create `pkg/scan/reach_default.go` (`//go:build !omni_govulncheck`), `pkg/scan/reach_govulncheck.go` (`//go:build omni_govulncheck`), `pkg/scan/reach_default_test.go`.

- [ ] **Step 1: Default (no-tag) failing test** (`pkg/scan/reach_default_test.go`):

```go
//go:build !omni_govulncheck

package scan

import (
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

func TestScanSourceUnsupportedWithoutTag(t *testing.T) {
	_, err := ScanSource("./...", &DB{byPkg: map[string][]osvEntry{}}, Options{})
	if err == nil || !cmderr.IsUnsupported(err) {
		t.Fatalf("ScanSource without -tags omni_govulncheck = %v, want ErrUnsupported", err)
	}
}
```

- [ ] **Step 2: Run, verify fail:** `go test ./pkg/scan/ -run ScanSourceUnsupported -v` → FAIL (`ScanSource` undefined).
- [ ] **Step 3: Implement the default stub** (`pkg/scan/reach_default.go`):

```go
//go:build !omni_govulncheck

package scan

import "github.com/inovacc/omni/internal/cli/cmderr"

// ScanSource performs reachability-aware scanning of a Go source tree.
// In the default build it is unavailable because reachability analysis requires
// invoking the Go build system (go list), which violates omni's no-exec rule.
// Build with -tags omni_govulncheck to enable it.
func ScanSource(pattern string, db *DB, opts Options) (Report, error) {
	return Report{}, cmderr.Wrap(cmderr.ErrUnsupported,
		"reachability source scanning requires building with -tags omni_govulncheck")
}
```

- [ ] **Step 4: Implement the tagged path** (`pkg/scan/reach_govulncheck.go`), documenting that it intentionally relaxes the no-exec rule for this opt-in build only:

```go
//go:build omni_govulncheck

package scan

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"golang.org/x/vuln/scan"
)

// ScanSource runs govulncheck's reachability engine over the Go source tree at pattern
// and returns only findings whose vulnerable symbols are actually called.
// NOTE: golang.org/x/vuln/scan invokes the Go toolchain (go list); this is permitted
// ONLY in the opt-in omni_govulncheck build, never the default binary.
func ScanSource(pattern string, db *DB, opts Options) (Report, error) {
	var out bytes.Buffer
	cmd := scan.Command(context.Background(), "-json", "-mode", "source", pattern)
	cmd.Stdout = &out
	cmd.Stderr = &bytes.Buffer{}
	if err := cmd.Start(); err != nil {
		return Report{}, cmderr.Wrap(cmderr.ErrIO, "start govulncheck")
	}
	if err := cmd.Wait(); err != nil {
		// govulncheck exits non-zero when vulns are found; JSON is still on stdout.
	}
	rep, err := parseGovulncheckJSON(out.Bytes())
	if err != nil {
		return Report{}, cmderr.Wrap(cmderr.ErrInvalidInput, "parse govulncheck output")
	}
	if opts.FailOn > SeverityUnknown {
		for _, f := range rep.Findings {
			if sev, _ := ParseSeverity(f.Severity); sev >= opts.FailOn {
				return rep, cmderr.Wrap(cmderr.ErrConflict,
					"reachable vulnerabilities found at or above --fail-on threshold")
			}
		}
	}
	return rep, nil
}

// parseGovulncheckJSON converts the streaming govulncheck JSON ("osv" + "finding"
// messages) into a Report, marking every emitted finding Reachable=true (govulncheck
// in source mode only reports called symbols).
func parseGovulncheckJSON(b []byte) (Report, error) {
	dec := json.NewDecoder(bytes.NewReader(b))
	osvByID := map[string]osvEntry{}
	reachable := map[string]bool{}
	for dec.More() {
		var msg struct {
			OSV     *osvEntry `json:"osv"`
			Finding *struct {
				OSV   string `json:"osv"`
				Trace []struct {
					Function string `json:"function"`
				} `json:"trace"`
			} `json:"finding"`
		}
		if err := dec.Decode(&msg); err != nil {
			return Report{}, fmt.Errorf("decode govulncheck json: %w", err)
		}
		switch {
		case msg.OSV != nil:
			osvByID[msg.OSV.ID] = *msg.OSV
		case msg.Finding != nil && len(msg.Finding.Trace) > 0 && msg.Finding.Trace[0].Function != "":
			reachable[msg.Finding.OSV] = true // a trace with a function => reachable
		}
	}
	var findings []Finding
	for id := range reachable {
		e := osvByID[id]
		yes := true
		f := Finding{ID: id, Summary: e.Summary, Severity: severityLabel(e.Severity).String(), Reachable: &yes}
		if len(e.Affected) > 0 {
			f.Package = e.Affected[0].Package.Name
		}
		findings = append(findings, f)
	}
	sortFindings(findings)
	return Report{Findings: findings, Scanned: len(osvByID)}, nil
}
```

   Extract the sort from Task 5 into `sortFindings([]Finding)` in `pkg/scan/match.go` (shared by both paths). The `db` param is unused in the tagged path (govulncheck carries its own DB client); name it `_ *DB` to satisfy the linter, and add a doc note.
- [ ] **Step 5: Verify BOTH builds:**
```bash
go build ./... && go test ./pkg/scan/ -run ScanSourceUnsupported -v   # default: stub returns ErrUnsupported
go build -tags omni_govulncheck ./... && go vet -tags omni_govulncheck ./pkg/scan/
go get golang.org/x/vuln && go mod tidy   # adds x/vuln (tag-gated; accept MVS bump per Phase 4)
```
   Expected: default build does NOT compile `golang.org/x/vuln`; tagged build compiles cleanly.
- [ ] **Step 6: Commit:** `gofmt -w pkg/scan/ && git commit -- pkg/scan/ go.mod go.sum -m "feat(scan): reachability source scan behind omni_govulncheck tag (default: ErrUnsupported)"`

---

### Task 7: CLI glue + Cobra wrapper — `omni scan` / `omni scan source` (TDD)

**Files:** Create `internal/cli/scan/scan.go` (+`scan_test.go`), `cmd/scan.go`. Modify nothing else yet.

- [ ] **Step 1: Failing test** (`internal/cli/scan/scan_test.go`) — `RunScan` reads an SBOM file + a signed DB, prints text by default and JSON with `--json`, returns `cmderr.ErrConflict` when `--fail-on` trips, `ErrNotFound` for a missing SBOM, `ErrInvalidInput` for a malformed one. Use a committed-in-test fixture (build the signed DB + a tiny SPDX/CycloneDX SBOM in `t.TempDir()` with the same low-cost keypair from Task 4). Assert: a vulnerable SBOM + `--fail-on high` → `cmderr.IsConflict(err)`; the same scan without `--fail-on` → `err == nil` and stdout contains the OSV id.
- [ ] **Step 2: Run, verify fail:** `go test ./internal/cli/scan/ -v` → FAIL (package/`RunScan` undefined).
- [ ] **Step 3: Implement** `internal/cli/scan/scan.go`:
  - `type Options struct { DBPath, DBKeyPath, DBSigPath, FailOn string; JSON bool; MaxDBAge time.Duration; Online, Reachability bool }`.
  - `func RunScan(w io.Writer, args []string, opts Options) error`: resolve the SBOM path from `args[0]`; `os.ReadFile` it (missing → `cmderr.Wrap(cmderr.ErrNotFound, …)`, permission → `ErrPermission`); parse via `pkg/sbom/format` (Phase 5 provides a `format.Parse([]byte) (format.Document, error)` — if its name differs, adapt this one call; parse error → `ErrInvalidInput`). Load the DB: read `opts.DBPath` (`.zip`), its sibling `.minisig` (or `opts.DBSigPath`), and `opts.DBKeyPath`; call `scan.VerifyAndLoadDB(...)` (verify failure → already-wrapped `ErrConflict`). Apply `db.CheckFresh(opts.MaxDBAge)` (stale → wrap `ErrStaleDB` as `cmderr.ErrConflict`). Map `--fail-on` text via `scan.ParseSeverity` (bad label → `ErrInvalidInput`). Call `scan.Scan(doc, db, scan.Options{FailOn: sev})`. Render: text table (`ID  PACKAGE  VERSION  FIXED  SEVERITY  SUMMARY`) or `json.NewEncoder(w).Encode(report)` when `opts.JSON`. Return the `(report-render, gate-error)` — render the report to `w` FIRST, THEN return the `ErrConflict` so CI sees both output and exit code.
  - `func RunScanSource(w io.Writer, args []string, opts Options) error`: load the DB the same way, then call `scan.ScanSource(args[0], db, scan.Options{FailOn: sev, Reachability: true})`; render identically. (Default build returns the `ErrUnsupported` from the stub — surfaced verbatim.)
- [ ] **Step 4: Cobra wrapper** (`cmd/scan.go`):
  - `var scanCmd = &cobra.Command{Use: "scan <sbom>", Short: "Scan an SBOM against a signed OSV vulnerability database", Args: cobra.MaximumNArgs(1), RunE: func(cmd, args) error { return scan.RunScan(cmd.OutOrStdout(), args, optsFromFlags()) }}`.
  - Flags on `scanCmd`: `--db` (path to `osv-db.zip`), `--db-key` (minisign pubkey path), `--db-sig` (detached `.minisig`; default `<db>.minisig`), `--fail-on` (string, default ""), `--json` (bool), `--max-db-age` (duration, default `0` = off), `--online` (bool).
  - `var scanSourceCmd = &cobra.Command{Use: "source <pattern>", Short: "Reachability-aware Go source scan (requires -tags omni_govulncheck)", Args: cobra.ExactArgs(1), RunE: …RunScanSource}`; `scanCmd.AddCommand(scanSourceCmd)`.
  - `var scanDBCmd = &cobra.Command{Use: "db", Short: "Manage the OSV vulnerability database"}` with `scanDBUpdateCmd` (`Use: "update"`) — wired in Task 8; add `scanCmd.AddCommand(scanDBCmd)` and `scanDBCmd.AddCommand(scanDBUpdateCmd)` here as a placeholder calling a Task-8 function.
  - `init()`: `rootCmd.AddCommand(scanCmd)`.
- [ ] **Step 5: Run, verify pass + smoke:** `go test ./internal/cli/scan/ -v` → PASS. Manual: build a fixture DB (`omni sign` over a zip), then `go run . scan --db /tmp/osv-db.zip --db-key /tmp/k.pub /tmp/sbom.json` and `… --fail-on high` (expect exit 1).
- [ ] **Step 6: Commit:** `gofmt -w internal/cli/scan cmd/scan.go && git commit -- internal/cli/scan cmd/scan.go -m "feat(scan): omni scan / scan source CLI with cmderr-classified fail-on gating"`

---

### Task 8: `omni scan db update` — download + verify OSV DB (TDD)

**Files:** Modify `internal/cli/scan/scan.go` (add `RunDBUpdate`), create `internal/cli/scan/dbupdate_test.go`.

- [ ] **Step 1: Failing test** (`internal/cli/scan/dbupdate_test.go`) — `RunDBUpdate` fetches the DB zip + `.minisig` from a base URL, verifies the signature with the pinned public key, and writes both to the cache dir ONLY if verification passes; a tampered download leaves NO files and returns an error. Use `net/http/httptest.NewServer` serving a fixture zip + a fixture `.minisig` (signed with the Task-4 low-cost key); assert the good case writes `osv-db.zip` + `osv-db.zip.minisig` into `t.TempDir()`, and a server serving a corrupted zip writes nothing and returns a non-nil error.

```go
package scan

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDBUpdateVerifiesBeforeWrite(t *testing.T) {
	zipBytes, sig, pub := buildSignedFixtureDB(t) // helper: signs with WithScryptParams(1<<15,8,1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch filepath.Base(r.URL.Path) {
		case "osv-db.zip":
			_, _ = w.Write(zipBytes)
		case "osv-db.zip.minisig":
			_, _ = w.Write(sig)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	keyPath := filepath.Join(dir, "db.pub")
	_ = os.WriteFile(keyPath, pub, 0o644)

	if err := RunDBUpdate(os.Stdout, Options{}, srv.URL, dir, keyPath); err != nil {
		t.Fatalf("RunDBUpdate(good) = %v, want nil", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "osv-db.zip")); err != nil {
		t.Errorf("osv-db.zip not written: %v", err)
	}
}

func TestDBUpdateTamperedFailsClosed(t *testing.T) {
	zipBytes, sig, pub := buildSignedFixtureDB(t)
	zipBytes[len(zipBytes)/2] ^= 0xFF // corrupt after signing
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if filepath.Base(r.URL.Path) == "osv-db.zip.minisig" {
			_, _ = w.Write(sig)
			return
		}
		_, _ = w.Write(zipBytes)
	}))
	defer srv.Close()

	dir := t.TempDir()
	keyPath := filepath.Join(dir, "db.pub")
	_ = os.WriteFile(keyPath, pub, 0o644)
	if err := RunDBUpdate(os.Stdout, Options{}, srv.URL, dir, keyPath); err == nil {
		t.Fatal("RunDBUpdate(tampered) must fail closed")
	}
	if _, err := os.Stat(filepath.Join(dir, "osv-db.zip")); !os.IsNotExist(err) {
		t.Error("tampered DB must NOT be written to disk")
	}
}
```

- [ ] **Step 2: Run, verify fail:** `go test ./internal/cli/scan/ -run DBUpdate -v` → FAIL.
- [ ] **Step 3: Implement** `RunDBUpdate(w io.Writer, opts Options, baseURL, cacheDir, keyPath string) error`:
  - `http.Get(baseURL+"/osv-db.zip")` and `…/osv-db.zip.minisig` (network error → `cmderr.Wrap(cmderr.ErrIO, …)`; non-200 → `ErrIO`); `io.ReadAll` both into memory (NO os/exec; `net/http` is pure-Go and permitted).
  - Read `keyPath`, `scan.VerifyAndLoadDB(zipBytes, sig, pubText)` — verify BEFORE writing; on failure return the wrapped `ErrConflict` and write nothing.
  - Only after verify passes: `os.MkdirAll(cacheDir, 0o755)`, atomically write `osv-db.zip` (temp file + `os.Rename`) and `osv-db.zip.minisig`, perms `0o644`. Print `db updated: <n> entries, generated <ts>` to `w`.
  - The default cache dir (resolved in `cmd/scan.go`) is `os.UserCacheDir()/omni/osv-db`; the pinned default public key path is documented but user-overridable via `--db-key`.
- [ ] **Step 4: Wire the Cobra subcommand** — in `cmd/scan.go`, set `scanDBUpdateCmd.RunE` to call `scan.RunDBUpdate(cmd.OutOrStdout(), optsFromFlags(), dbBaseURL, cacheDir, dbKeyPath)` with a `--url` flag (default the canonical omni OSV-DB release URL) and `--cache-dir` flag.
- [ ] **Step 5: Run, verify pass:** `go test ./internal/cli/scan/ -run DBUpdate -v` → PASS.
- [ ] **Step 6: Commit:** `gofmt -w internal/cli/scan cmd/scan.go && git commit -- internal/cli/scan cmd/scan.go -m "feat(scan): omni scan db update — verify-before-write OSV DB download (fail-closed)"`

---

### Task 9: `omni pipe` integration (SBOM-on-stdin scan)

**Files:** Modify `cmd/pipe.go`.

- [ ] **Step 1: Failing test** — a `pipe`-style invocation of `scan` reads an SBOM from stdin and writes findings to stdout via the Unified registry (mirror an existing `cmd/pipe_test.go` case). The DB path/key come from env (`OMNI_SCAN_DB`, `OMNI_SCAN_DB_KEY`) since pipe stages take only `(w, r, args)`.
- [ ] **Step 2: Run → FAIL.**
- [ ] **Step 3: Implement** — in `buildPipeRegistry()` add:
```go
reg.Register("scan", command.AdaptWriterReaderArgs(func(w io.Writer, r io.Reader, args []string) error {
	return scan.RunScanStdin(w, r, args, scan.OptionsFromEnv())
}))
```
   Add `RunScanStdin(w io.Writer, r io.Reader, args []string, opts Options) error` to `internal/cli/scan/scan.go`: read the SBOM from `r` (instead of a file), then run the same matcher/gate path. Add `OptionsFromEnv()` reading `OMNI_SCAN_DB`, `OMNI_SCAN_DB_KEY`, `OMNI_SCAN_FAIL_ON`, `OMNI_SCAN_MAX_DB_AGE`. (`scan source` and `scan db update` stay Cobra-only — they are not stdin transforms.)
- [ ] **Step 4: Run → PASS:** `go test ./internal/cli/scan/... ./cmd/... -run Pipe -v`.
- [ ] **Step 5: Commit:** `git commit -- cmd/pipe.go internal/cli/scan -m "feat(scan): register scan (SBOM-on-stdin) in the pipe Unified registry"`

---

### Task 10: Golden-master tests (positive + negative, deterministic)

**Files:** Modify `testing/golden/golden_tests.yaml` AND `tools/golden/golden_tests.yaml` (keep in sync); add committed fixtures under `testing/golden/fixtures/scan/`.

- [ ] **Step 1: Create deterministic fixtures** in `testing/golden/fixtures/scan/`:
  - `osv-db.zip` — a fixed bundle: `manifest.json` with a FIXED `generated` timestamp far in the past (e.g. `2026-01-01T00:00:00Z`) + 2 `entries/*.json` (one matching the SBOM at high severity, one non-matching).
  - `osv-db.zip.minisig` — the detached signature produced by signing `osv-db.zip` with a committed low-cost keypair (reuse the `sign` fixtures' key, or generate a dedicated one with `WithScryptParams(1<<15,8,1)` and commit `db.pub`).
  - `db.pub` — the public key for verification.
  - `vuln-sbom.json` — a tiny SPDX (or CycloneDX) SBOM listing the vulnerable component at the affected version.
  - `clean-sbom.json` — an SBOM with no matching components.
  - `bad-sbom.json` — malformed JSON (for the `ErrInvalidInput` negative).
  > Because `generated` is fixed in the past, golden tests MUST NOT pass `--max-db-age` (that would be time-dependent); a dedicated staleness test passes a tiny `--max-db-age 1s` and asserts `exit_code: 1`.
- [ ] **Step 2: Add a `scan` category to BOTH yaml files** (mirroring the `sign` category at `golden_tests.yaml:1091`):
  - `scan_clean` — `args: ["scan", "--db", "{fixtures}/osv-db.zip", "--db-key", "{fixtures}/db.pub", "{fixtures}/clean-sbom.json"]` → exit 0, `normalizations: ["strip_path"]` (DBAge is non-deterministic → also add a `strip_db_age` normalization OR omit DBAge from text output; prefer printing DBAge only in `--json` and excluding the JSON age field from the golden by using text mode here).
  - `scan_vuln_report` — same but with `vuln-sbom.json`, NO `--fail-on` → exit 0, output lists the OSV id (text mode; `strip_path`).
  - `scan_fail_on_high` — `… "--fail-on", "high", "{fixtures}/vuln-sbom.json"]` → `exit_code: 1` `# cmderr.ErrConflict` (the CI gating path), `normalizations: ["strip_path"]`.
  - `scan_missing_sbom` → `exit_code: 1` `# cmderr.ErrNotFound`.
  - `scan_bad_sbom` → `… "{fixtures}/bad-sbom.json"]` → `exit_code: 2` `# cmderr.ErrInvalidInput`.
  - `scan_tampered_db` → point `--db` at a committed `osv-db.tampered.zip` (one flipped byte) → `exit_code: 1` `# cmderr.ErrConflict (signature verification, fail-closed)`.
  - `scan_stale_db` → `… "--max-db-age", "1s", "{fixtures}/vuln-sbom.json"]` → `exit_code: 1` `# cmderr.ErrConflict (ErrStaleDB, fail-loud)`.
  - `scan_source_unsupported` → `args: ["scan", "source", "./..."]` → `exit_code: 6` `# cmderr.ErrUnsupported (default build, no omni_govulncheck tag)`.
- [ ] **Step 3: Determinism guard** — confirm text output is byte-stable: it must NOT include the wall-clock DBAge or any absolute path. If the renderer prints DBAge in text mode, add a `strip_db_age` normalization to `testing/scripts/` (regex `db age: .*` → `db age: <STRIPPED>`) and register it in BOTH harness configs; otherwise keep DBAge JSON-only.
- [ ] **Step 4: Generate + verify snapshots:** `task test:golden:update && task golden:record`, then `python testing/scripts/test_golden.py` → all green.
- [ ] **Step 5: Commit:** `git commit -- testing/golden tools/golden -m "test(scan): golden-master report + fail-on/tampered/stale/unsupported negatives"`

---

### Task 11: Docs + final gate

**Files:** `docs/COMMANDS.md`, `CLAUDE.md` (command inventory line), `docs/architecture/cloud-integrations.md` (or the appropriate subsystem doc), `docs/superpowers/specs/2026-05-16-06-pkg-scan-design.md` (status), `docs/EXTERNAL_SOURCES.md` (attribute OSV schema + `golang.org/x/vuln` under the build tag).

- [ ] **Step 1: Docs** — add `scan`, `scan source`, `scan db update` to `docs/COMMANDS.md` and bump the CLAUDE.md inventory count; document the `omni_govulncheck` build tag and the signed-DB workflow (how to build a signed `osv-db.zip` with `omni sign`). Add an OSV-schema + `golang.org/x/vuln` attribution row to `docs/EXTERNAL_SOURCES.md`. Run `omni aicontext` / `omni cmdtree` regen if applicable. Mark the spec `Status: Complete`.
- [ ] **Step 2: Final gate:**
```bash
go build ./... && go build -tags omni_govulncheck ./...
go vet ./... && go vet -tags omni_govulncheck ./pkg/scan/...
gofmt -l pkg/scan internal/cli/scan cmd/scan.go
golangci-lint run --timeout=5m ./...
go test ./pkg/scan/... ./internal/cli/scan/... -count=1
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ./... && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./...
python testing/scripts/test_golden.py
```
   Expected: all green; default build does NOT link `golang.org/x/vuln`; tagged build compiles.
- [ ] **Step 3: Commit:** `git commit -- docs/ CLAUDE.md -m "docs(scan): document omni scan / scan source / scan db update; mark Phase 06 complete"`

---

## Self-Review

**Spec coverage** (SCAN success criteria → tasks):

| Spec requirement | Task(s) |
|---|---|
| `omni scan <sbom>` on SPDX/CycloneDX → findings in JSON and text | T5 (`Scan`), T7 (`RunScan` text/JSON render) |
| `--fail-on <severity>` → `cmderr.ErrConflict`; CI gating golden-tested | T2 (Severity order), T5 (gate), T7 (CLI), T10 (`scan_fail_on_high` golden) |
| `omni scan source <path>` → reachability-aware findings (only called symbols) | T6 (tag split + govulncheck JSON `trace` parse), T7 (`RunScanSource`) |
| `scan db update` downloads + verifies OSV DB signed with `pkg/sign`; tampered fails closed | T4 (`VerifyAndLoadDB`), T8 (`RunDBUpdate` verify-before-write), T10 (`scan_tampered_db`) |
| Offline mode with cached DB; `--max-db-age` gates staleness, fails loudly | T4 (`CheckFresh`/`ErrStaleDB`), T7 (gate wiring), T10 (`scan_stale_db`) |
| `pkg/scan` imports `pkg/sbom/format.Document` ONLY (never `pkg/sbom/model`) | T5 (`document.go` is the single `format` importer; `componentsOf` adapter) |
| Pure-Go, no exec (default) | T2–T5 stdlib + `x/mod/semver`; exec confined to T6 `omni_govulncheck` tag; T11 gate builds both |
| Research-spike-driven architecture decision recorded | T1 (ADR-0008, hard gate) |
| `omni pipe` integration | T9 |

**Placeholder scan:** No `TBD`/`add validation`/`handle edge cases`. Every matcher/severity/DB/CLI body is given as real Go or a byte-exact spec (OSV bundle format + matching algorithm + CVSS bands sections). The two bounded "adapt this one call if Phase 5 differs" notes (T5 `componentsOf`, T7 `format.Parse`) are explicit, isolated coupling points to a not-yet-merged dependency — not vague work; the matcher is shape-agnostic by design. The govulncheck JSON shape (T6) is the documented `osv`/`finding`+`trace` streaming format from `golang.org/x/vuln/scan`.

**Type consistency:** `Severity` + `ParseSeverity` (T2) feed `severityLabel` (T2) → `Finding.Severity` (T2 type, T5 populated) → `--fail-on` gate (T5/T7). `osvEntry`/`osvAffected`/`rng`/`rngEvent`/`rawSeverity` (T3) are consumed by `matchEntry`/`affectedHit`/`inOpenInterval`/`smallestFixAbove` (T3), by `LoadDB`/`entriesFor` (T4), and by `severityLabel` (T2). `DB` (T4) is consumed by `Scan`/`scanComponents` (T5), both `ScanSource` paths (T6), and every `internal/cli/scan` entry point (T7–T9). `component`/`Options` (T5) bridge the SBOM boundary. `sortFindings` (T6, extracted from T5) is shared by both reachability builds. `VerifyAndLoadDB` (T4) uses the confirmed `pkg/sign.{ParsePublicKey, Verify}` signatures. `cmderr.{ErrConflict,ErrNotFound,ErrInvalidInput,ErrPermission,ErrIO,ErrUnsupported,Wrap,IsConflict,IsUnsupported}` (confirmed in cmderr.go) are used per the convention table.

**Known risks:**
1. **Phase 5 not merged yet** — `pkg/sbom/format.Document` shape is assumed. Mitigation: ALL matcher tests (T2–T5) use a local `component`/`osvEntry` fake and never import `format`; only `document.go` (T5) and `RunScan` (T7) touch `format`, behind a single `componentsOf`/`format.Parse` adapter to fix if names differ. The plan should not START Task 5 until Phase 5's `format` package exists (depends-on gate).
2. **CVSS base-score parser correctness** — a hand-rolled v3.1/v4.0 formula is error-prone. Mitigation: T2's table-driven test pins known vectors→bands; if a numeric `score` is present omni uses it directly (the common OSV-for-Go case), so the formula path is the fallback, not the primary. Document that `unknown` is always reported and treated as below `low`.
3. **govulncheck/exec under the tag** — `golang.org/x/vuln/scan` execs `go list`; this is the single, opt-in, documented exception to the no-exec rule, fully excluded from the default binary (T6 build-tag, verified in T11). The default scanner has zero exec.
4. **DB freshness in golden tests** — a fixed past `generated` timestamp makes the unsigned matcher deterministic but means golden tests must avoid `--max-db-age` except the dedicated `scan_stale_db` case; DBAge is kept out of byte-compared output (T10 Step 3).
5. **MVS bump from x/vuln** — accepted, confined to the tagged file; default-build `go.mod` graph gains only `x/mod` (likely already present from Phase 5). T11 verifies the default build does not link x/vuln.
