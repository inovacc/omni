# Phase 05 — `pkg/sbom/` SBOM Generation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.
> **HARD GATE:** Task 1 (ADR-0007) MUST be written AND human-reviewed/approved BEFORE any code task (2+) begins. The spec (`docs/superpowers/specs/2026-05-16-05-pkg-sbom-design.md`) declares an ADR gate ("Required before any code lands (1 ADR)") — non-negotiable.

**Goal:** Ship `omni sbom` producing **byte-deterministic** SPDX 2.3 JSON and CycloneDX 1.5 JSON for (a) a Go module directory (`go.mod` + `go.sum`) and (b) a built Go binary (via `debug/buildinfo`). Every component carries a correctly normalized Go purl; the Go toolchain is a listed component for binary SBOMs; identical input → identical bytes (golden-master pinnable). Optionally sign the emitted SBOM by reusing the Phase 04 `pkg/sign` primitive (`--sign`).

**Architecture:** Pure-stdlib JSON emission. `pkg/sbom/model` holds an internal source-format-agnostic representation; `pkg/sbom/format` holds the **stable boundary type** `format.Document` (the only type `pkg/scan/` Phase 6 may import) plus two pure emitters that marshal hand-written minimal SPDX/CycloneDX structs via `encoding/json`. We do **NOT** depend on `github.com/spdx/tools-golang` or `github.com/CycloneDX/cyclonedx-go` in the default build path — their encoders pull heavy transitive trees and do not guarantee deterministic field order / inject random `serialNumber`s (Phase 04 MVS finding: even build-tag-gated heavy deps bump `go.mod` via MVS). Those two libraries are imported **only** behind `//go:build omni_sbomvalidate` as an in-process round-trip *validator* used by tests/CI, never by the emitter. Layering: pure libs `pkg/sbom/{model,format,collect,purl}/` → I/O glue `internal/cli/sbom/` → thin Cobra wrapper `cmd/sbom.go`. Signing reuses `pkg/sign` (no re-impl).

**Tech Stack:** Go stdlib (`encoding/json`, `debug/buildinfo`, `runtime/debug`, `sort`, `crypto/sha256`, `bufio`); `golang.org/x/mod/{semver,module}` (**already in `go.sum` at v0.36.0** — promote to direct `require`, no new transitive cost); `pkg/sign` (Phase 04, for `--sign`); Cobra; the Python YAML golden harness. Optional `github.com/spdx/tools-golang` v0.5.7 + `github.com/CycloneDX/cyclonedx-go` v0.11.0 behind `//go:build omni_sbomvalidate` (validator only).

**Repo conventions (from research, cite when implementing):**
- Commands self-wire: `cmd/<name>.go` declares `var xCmd = &cobra.Command{...}` and calls `rootCmd.AddCommand(xCmd)` in `init()`; `RunE` reads flags → Options → calls `internal/cli/<name>.RunX(cmd.OutOrStdout(), ...)`. No central registration list. (Confirmed against `cmd/sign.go`.)
- `cmderr` (`internal/cli/cmderr/cmderr.go`): missing path → `ErrNotFound` (exit 1); unreadable/bad-perms → `ErrPermission` (3); bad flags / malformed `go.mod` / not-a-Go-binary / unknown `--format` → `ErrInvalidInput` (2); write failure → `ErrIO` (4); `--validate`/`--sign` feature unavailable (missing build tag / no key) → `ErrUnsupported` (6) or `ErrInvalidInput` (2) as noted per task; determinism self-check failure → `ErrConflict` (1). Use `cmderr.Wrap(sentinel, "msg")` and `errors.Is`/`As`, never `==`. `Is<Class>()` predicates exist.
- Layering: pure lib `pkg/<X>/` (stdlib + minimal pure-Go deps, NO cobra, NO `io.Writer`-for-output) + I/O glue `internal/cli/<X>/` + thin `cmd/<X>.go`. `doc.go` package docstring required per package.
- pkg API triage: `pkg/sbom/format` is the **stable** boundary (no `// Experimental:` marker). `pkg/sbom/model`, `pkg/sbom/collect`, `pkg/sbom/purl` carry `// Experimental:` in their `doc.go` until Phase 3 triage promotes them.
- Golden harness is Python+YAML, TWO files kept in sync: `testing/golden/golden_tests.yaml` and `tools/golden/golden_tests.yaml`; committed fixtures referenced via the `{fixtures}` placeholder → `testing/golden/fixtures/<category>/`; negatives set `exit_code:` + `normalizations: [strip_path]` annotated with the cmderr sentinel; regenerate with `task test:golden:update` then `task golden:record`. (Confirmed: the `sign` category at line ~1093 uses exactly this shape; fixtures use a `gen_fixtures.go` `//go:build ignore` generator.)
- Pipe: register stdin→stdout commands in `cmd/pipe.go buildPipeRegistry()` via `command.AdaptWriterReaderArgs(...)`. (sbom reads a path argument, not stdin — see Task 9 for the constrained registration.)
- ADRs live in `docs/adr/` as `ADR-NNNN-kebab-title.md`; 0001–0006 exist, so the next number is **0007**; header format per `docs/adr/ADR-0004-internalize-cobra-cli.md` (`# ADR-NNNN: …`, `**Status:**`, `**Date:**`, `**Decision:**`, then `## Context`, `## Analysis`, `## Consequences`).
- INVARIANTS: pure-Go, NO `os/exec`, no CGO; cross-platform via `//go:build` tags (never runtime `os ==`); `io.Writer`/`io.Reader`; deferred `Close`.

---

## Authoritative output format (implement byte-exactly)

Both emitters use a single deterministic JSON writer (Task 4): a `*json.Encoder` over the destination with `SetIndent("", "  ")` and `SetEscapeHTML(false)`, fed a value whose every slice is pre-sorted and whose every map is avoided (structs only — Go marshals struct fields in declaration order, giving stable key order). A trailing newline is written. No `serialNumber`, no wall-clock timestamp unless a fixed value is supplied via `--source-date` (RFC-3339). Default timestamp when `--source-date` is omitted: the literal `1970-01-01T00:00:00Z` (epoch) so output is deterministic without a flag.

### purl (Go type) — `pkg/sbom/purl`
`pkg:golang/<escaped-lowercased-module-path>@<version>`
- The module **path** (not import path) is the namespace+name; lowercase it, then apply `module.EscapePath` is **not** used (purl uses percent-encoding, not Go's `!`-escaping). Instead: lowercase the path, split on the LAST `/` into namespace (everything before) and name (after); each segment is percent-encoded per RFC-3986 *except* the `/` separators inside the namespace are preserved. Concretely (sufficient for Go module paths, which only contain `[a-z0-9.\-_~/]` after lowercasing plus an occasional uppercase that we already lowercased): emit `pkg:golang/` + lowercased-path + `@` + version. (Go module paths never contain characters requiring percent-encoding once lowercased; assert this in tests.)
- **version**: use `semver.Canonical(v)` when `semver.IsValid(v)`, EXCEPT preserve a trailing `+incompatible` suffix (strip before canonicalizing, re-append after). Pseudo-versions (`v0.0.0-20240101000000-abcdef123456`) are already canonical-valid and pass through unchanged. An empty version (replace-with-local-dir, or main module with no tag) yields a purl WITHOUT the `@<version>` suffix.
- **replace directives**: when a module is replaced, the purl and version describe the **effective** (replacement) module — its path and version — because that is "what actually shipped." Record the original path in a model field for the SPDX `comment`/CycloneDX property, but the purl points at the replacement.

### SPDX 2.3 JSON (minimal, deterministic) — emitted struct shape
Top level (struct field/JSON order is the on-disk order):
```
spdxVersion: "SPDX-2.3"
dataLicense:  "CC0-1.0"
SPDXID:       "SPDXRef-DOCUMENT"
name:         <document name>
documentNamespace: "https://spdx.org/spdxdocs/omni/" + <name> + "-" + <contentHash>
creationInfo: { created: <ts>, creators: ["Tool: omni-sbom-<omniVersion>"] }
packages:     [ {name, SPDXID:"SPDXRef-Package-"+<slug>, versionInfo, downloadLocation:"NOASSERTION", filesAnalyzed:false, licenseConcluded:"NOASSERTION", licenseDeclared:"NOASSERTION", copyrightText:"NOASSERTION", externalRefs:[{referenceCategory:"PACKAGE-MANAGER", referenceType:"purl", referenceLocator:<purl>}] }, ... sorted by SPDXID ]
relationships: [ {spdxElementId:"SPDXRef-DOCUMENT", relatedSpdxElement:"SPDXRef-Package-"+<rootSlug>, relationshipType:"DESCRIBES"}, then for each dep {spdxElementId:"SPDXRef-Package-"+<rootSlug>, relatedSpdxElement:"SPDXRef-Package-"+<depSlug>, relationshipType:"DEPENDS_ON"} ... sorted ]
```
- `<slug>` = the module path with every non-`[A-Za-z0-9.]` rune replaced by `-` (SPDX IDs allow only letters/digits/`.`/`-`), de-duplicated against collisions by appending `-<n>`.
- `<contentHash>` = first 16 hex chars of `sha256` over the sorted list of `purl` strings — deterministic, no UUID.

### CycloneDX 1.5 JSON (minimal, deterministic) — emitted struct shape
```
$schema:    "http://cyclonedx.org/schema/bom-1.5.schema.json"
bomFormat:  "CycloneDX"
specVersion:"1.5"
version:    1
metadata:   { timestamp:<ts>, tools:{ components:[{type:"application", name:"omni", version:<omniVersion>}] }, component:{ "bom-ref":<rootPurlOrSlug>, type:"application", name:<rootName>, version:<rootVersion>, purl:<rootPurl> } }
components: [ {"bom-ref":<purl>, type:"library", name:<modulePath>, version:<version>, purl:<purl>}, ... sorted by bom-ref ]
dependencies:[ {ref:<rootRef>, dependsOn:[<depPurl>...sorted]}, then one {ref:<depPurl>} per dep with no dependsOn ... sorted by ref ]
```
- **No `serialNumber`** (it is optional and would be random) — omit the field entirely for determinism.

### Binary SBOM specifics
For `--from binary` (or auto-detected when the path is a regular file that `buildinfo.ReadFile` parses), the root component is `bi.Main` (path+version, version often empty → `(devel)` from buildinfo passes through as-is), deps are `bi.Deps` (each a `*debug.Module`, following `.Replace`), AND the Go toolchain is emitted as an extra component: name `"go"` (CycloneDX `type:"application"`, purl `pkg:golang/std@<goversion>` where `<goversion>` is `bi.GoVersion` with the leading `go` kept, e.g. `go1.25.0`; SPDX package name `"go"`, versionInfo `bi.GoVersion`). The toolchain participates in `DEPENDS_ON`/`dependsOn` from the root.

---

### Task 1 (ADR GATE): ADR-0007 — SBOM round-trip oracle & purl-correctness policy

**Files:** Create `docs/adr/ADR-0007-sbom-determinism-and-purl-policy.md`

- [ ] **Step 1: Write the ADR** matching the `ADR-0004` header/section format (`# ADR-0007: …`, `**Status:** Accepted`, `**Date:** 2026-06-03`, `**Decision:** …`, then `## Context`, `## Analysis` (table), `## Consequences`). Record + justify these decisions (taken from the spec — this ADR is the gate that locks them):
  - **Emitter is pure-stdlib `encoding/json`** over hand-written minimal SPDX-2.3 / CycloneDX-1.5 structs; the upstream libraries (`spdx/tools-golang`, `CycloneDX/cyclonedx-go`) are **NOT** in the default build. Rationale: byte-determinism (struct field order + pre-sorted slices), zero random `serialNumber`/timestamps, and avoidance of the heavy transitive trees those encoders pull in (Phase 04 MVS finding — even tag-gated deps bump `go.mod`).
  - **Round-trip oracle for CI** = the upstream libraries behind `//go:build omni_sbomvalidate` decode-then-revalidate omni's output (proves schema-validity), PLUS the external `syft convert` oracle in CI (`syft convert omni.spdx.json -o cyclonedx-json` must succeed) per the success criteria. The validator build tag keeps those deps out of `omni` releases.
  - **"purl correctness"** = `pkg:golang/<lowercased-module-path>@<canonical-version>`; pseudo-versions pass through; `+incompatible` preserved; replace-directives resolve to the **effective** module; empty version → no `@` suffix. A purl round-trip test (build purl → re-parse path+version) is the drift guard.
  - **Determinism contract**: identical input bytes → identical output bytes. No UUID, no wall clock. `--source-date` (RFC-3339) overrides the default epoch timestamp. `documentNamespace`/content-hash derive from a sha256 over sorted purls, not randomness.
  - **`pkg/sbom/format.Document` is the only cross-package boundary** — `pkg/scan/` (Phase 6) imports it and never `pkg/sbom/model`.
- [ ] **Step 2: Stop for human review.** Do NOT proceed to code until this ADR is approved.

---

### Task 2: `pkg/sbom/purl` — Go purl construction (TDD)

**Files:** Create `pkg/sbom/purl/purl.go`, `pkg/sbom/purl/purl_test.go`, `pkg/sbom/purl/doc.go`; modify `go.mod` (promote `golang.org/x/mod` to a direct `require`).

- [ ] **Step 1: Write failing tests** (`pkg/sbom/purl/purl_test.go`):

```go
package purl_test

import (
	"testing"

	"github.com/inovacc/omni/pkg/sbom/purl"
)

func TestForModule(t *testing.T) {
	cases := []struct {
		name, path, version, want string
	}{
		{"tagged", "github.com/spf13/cobra", "v1.10.2", "pkg:golang/github.com/spf13/cobra@v1.10.2"},
		{"shorthand canonicalized", "golang.org/x/mod", "v0.36", "pkg:golang/golang.org/x/mod@v0.36.0"},
		{"uppercase lowered", "github.com/BurntSushi/toml", "v1.6.0", "pkg:golang/github.com/burntsushi/toml@v1.6.0"},
		{"pseudo passthrough", "github.com/dop251/goja", "v0.0.0-20260106131823-651366fbe6e3", "pkg:golang/github.com/dop251/goja@v0.0.0-20260106131823-651366fbe6e3"},
		{"incompatible preserved", "github.com/foo/bar", "v2.0.0+incompatible", "pkg:golang/github.com/foo/bar@v2.0.0+incompatible"},
		{"empty version no suffix", "example.com/local", "", "pkg:golang/example.com/local"},
		{"std toolchain", "std", "go1.25.0", "pkg:golang/std@go1.25.0"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := purl.ForModule(c.path, c.version); got != c.want {
				t.Errorf("ForModule(%q,%q) = %q, want %q", c.path, c.version, got, c.want)
			}
		})
	}
}

func TestForModuleRejectsEncodingNeeded(t *testing.T) {
	// Go module paths never need percent-encoding once lowercased; guard the assumption.
	got := purl.ForModule("github.com/a_b/c.d-e", "v1.0.0")
	if got != "pkg:golang/github.com/a_b/c.d-e@v1.0.0" {
		t.Errorf("got %q", got)
	}
}
```

- [ ] **Step 2: Run, verify fail:** `go test ./pkg/sbom/purl/ -v` → FAIL (package/`ForModule` undefined).
- [ ] **Step 3: Implement** `pkg/sbom/purl/purl.go`:

```go
package purl

import (
	"strings"

	"golang.org/x/mod/semver"
)

// ForModule returns the package-url for a Go module given its path and version.
// Format: pkg:golang/<lowercased-module-path>[@<canonical-version>].
// Pseudo-versions pass through; a "+incompatible" suffix is preserved; an empty
// version yields a purl with no "@" suffix. Special path "std" + a "goX.Y.Z"
// version represents the Go toolchain.
func ForModule(modulePath, version string) string {
	p := "pkg:golang/" + strings.ToLower(modulePath)
	v := normalizeVersion(version)
	if v == "" {
		return p
	}
	return p + "@" + v
}

// normalizeVersion canonicalizes a semver while preserving "+incompatible";
// non-semver values (e.g. "go1.25.0", "(devel)") pass through unchanged.
func normalizeVersion(version string) string {
	if version == "" {
		return ""
	}
	base, incompat := version, ""
	if strings.HasSuffix(version, "+incompatible") {
		base = strings.TrimSuffix(version, "+incompatible")
		incompat = "+incompatible"
	}
	if semver.IsValid(base) {
		return semver.Canonical(base) + incompat
	}
	return version
}
```

Create `pkg/sbom/purl/doc.go` (package docstring + `// Experimental:` marker). Run `go get golang.org/x/mod@v0.36.0 && go mod tidy` (it is already in `go.sum`; this only promotes it to a direct `require` — verify `go.sum` is unchanged afterward).

- [ ] **Step 4: Run, verify pass:** `go test ./pkg/sbom/purl/ -v` → PASS.
- [ ] **Step 5: Commit:** `gofmt -w pkg/sbom/purl/ && git commit -- pkg/sbom/purl go.mod go.sum -m "feat(sbom): Go purl construction with canonical-version + incompatible handling"`

---

### Task 3: `pkg/sbom/model` — source-agnostic SBOM model (TDD)

**Files:** Create `pkg/sbom/model/model.go`, `pkg/sbom/model/model_test.go`, `pkg/sbom/model/doc.go`.

- [ ] **Step 1: Write failing tests** (`pkg/sbom/model/model_test.go`):

```go
package model_test

import (
	"testing"

	"github.com/inovacc/omni/pkg/sbom/model"
)

func TestSortNormalizesOrder(t *testing.T) {
	m := &model.SBOM{
		Root: model.Component{Path: "example.com/app", Version: "v1.0.0"},
		Components: []model.Component{
			{Path: "github.com/z/z", Version: "v1.0.0"},
			{Path: "github.com/a/a", Version: "v2.0.0"},
			{Path: "github.com/a/a", Version: "v1.0.0"}, // duplicate path, lower version
		},
	}
	m.Normalize()
	if len(m.Components) != 3 {
		t.Fatalf("len = %d, want 3", len(m.Components))
	}
	// Sorted by (Path, Version): a@v1, a@v2, z@v1.
	want := []string{"github.com/a/a@v1.0.0", "github.com/a/a@v2.0.0", "github.com/z/z@v1.0.0"}
	for i, c := range m.Components {
		if got := c.Path + "@" + c.Version; got != want[i] {
			t.Errorf("Components[%d] = %q, want %q", i, got, want[i])
		}
	}
}

func TestSlugSanitizes(t *testing.T) {
	if got := model.Slug("github.com/spf13/cobra"); got != "github.com-spf13-cobra" {
		t.Errorf("Slug = %q", got)
	}
}
```

- [ ] **Step 2: Run, verify fail:** `go test ./pkg/sbom/model/ -v` → FAIL.
- [ ] **Step 3: Implement** `pkg/sbom/model/model.go`:

```go
package model

import (
	"sort"
	"strings"
)

// Kind describes whether a Component is the root subject or a dependency.
type Kind int

const (
	// KindRoot is the main module / binary the SBOM describes.
	KindRoot Kind = iota
	// KindLibrary is a dependency module.
	KindLibrary
	// KindToolchain is the Go toolchain (binary SBOMs only).
	KindToolchain
)

// Component is one node in the SBOM: a Go module, the root, or the toolchain.
type Component struct {
	Path           string // module path (effective path after replace)
	Version        string // module version ("" if unknown)
	Kind           Kind
	OriginalPath   string // pre-replace module path, "" if not replaced
	OriginalVersion string // pre-replace version, "" if not replaced
}

// SBOM is the source-format-agnostic representation produced by collectors and
// consumed by the SPDX/CycloneDX emitters.
type SBOM struct {
	Name       string      // document/subject name
	Root       Component   // the described subject
	Components []Component // dependencies (+ toolchain), excludes Root
}

// Normalize sorts Components deterministically by (Path, Version) and removes
// exact (Path,Version) duplicates so emitted output is stable.
func (s *SBOM) Normalize() {
	sort.Slice(s.Components, func(i, j int) bool {
		if s.Components[i].Path != s.Components[j].Path {
			return s.Components[i].Path < s.Components[j].Path
		}
		return s.Components[i].Version < s.Components[j].Version
	})
	out := s.Components[:0]
	var prevP, prevV string
	first := true
	for _, c := range s.Components {
		if !first && c.Path == prevP && c.Version == prevV {
			continue
		}
		out = append(out, c)
		prevP, prevV, first = c.Path, c.Version, false
	}
	s.Components = out
}

// Slug converts a module path into an SPDX-ID-safe token: every rune that is
// not a letter, digit, or '.' becomes '-'.
func Slug(path string) string {
	var b strings.Builder
	for _, r := range path {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	return b.String()
}
```

Create `pkg/sbom/model/doc.go` (docstring + `// Experimental:`).

- [ ] **Step 4: Run, verify pass:** `go test ./pkg/sbom/model/ -v` → PASS.
- [ ] **Step 5: Commit:** `gofmt -w pkg/sbom/model/ && git commit -- pkg/sbom/model -m "feat(sbom): source-agnostic SBOM model with deterministic Normalize + Slug"`

---

### Task 4: `pkg/sbom/format` — stable boundary + deterministic emitters (TDD)

**Files:** Create `pkg/sbom/format/format.go` (boundary `Document` + `From`), `pkg/sbom/format/spdx.go`, `pkg/sbom/format/cyclonedx.go`, `pkg/sbom/format/format_test.go`, `pkg/sbom/format/doc.go`.

- [ ] **Step 1: Write failing tests** (`pkg/sbom/format/format_test.go`) — assert byte-determinism + key fields:

```go
package format_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/inovacc/omni/pkg/sbom/format"
	"github.com/inovacc/omni/pkg/sbom/model"
)

func sample() *model.SBOM {
	return &model.SBOM{
		Name: "omni",
		Root: model.Component{Path: "github.com/inovacc/omni", Version: "v1.0.0", Kind: model.KindRoot},
		Components: []model.Component{
			{Path: "github.com/spf13/cobra", Version: "v1.10.2", Kind: model.KindLibrary},
			{Path: "golang.org/x/mod", Version: "v0.36.0", Kind: model.KindLibrary},
		},
	}
}

func TestEmitDeterministic(t *testing.T) {
	doc := format.From(sample(), format.Options{OmniVersion: "v0.1.0", SourceDate: "1970-01-01T00:00:00Z"})
	for _, f := range []format.Kind{format.SPDX, format.CycloneDX} {
		var a, b bytes.Buffer
		if err := doc.Encode(&a, f); err != nil {
			t.Fatalf("encode %v: %v", f, err)
		}
		if err := doc.Encode(&b, f); err != nil {
			t.Fatalf("encode2 %v: %v", f, err)
		}
		if !bytes.Equal(a.Bytes(), b.Bytes()) {
			t.Errorf("%v: two encodes differ (non-deterministic)", f)
		}
		if !bytes.HasSuffix(a.Bytes(), []byte("\n")) {
			t.Errorf("%v: missing trailing newline", f)
		}
	}
}

func TestSPDXShape(t *testing.T) {
	doc := format.From(sample(), format.Options{OmniVersion: "v0.1.0"})
	var buf bytes.Buffer
	if err := doc.Encode(&buf, format.SPDX); err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["spdxVersion"] != "SPDX-2.3" || m["dataLicense"] != "CC0-1.0" {
		t.Errorf("bad header: %v / %v", m["spdxVersion"], m["dataLicense"])
	}
	if !strings.Contains(buf.String(), "pkg:golang/github.com/spf13/cobra@v1.10.2") {
		t.Error("missing cobra purl")
	}
}

func TestCycloneDXShape(t *testing.T) {
	doc := format.From(sample(), format.Options{OmniVersion: "v0.1.0"})
	var buf bytes.Buffer
	if err := doc.Encode(&buf, format.CycloneDX); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.Contains(s, `"bomFormat": "CycloneDX"`) || !strings.Contains(s, `"specVersion": "1.5"`) {
		t.Errorf("bad header: %s", s[:120])
	}
	if strings.Contains(s, "serialNumber") {
		t.Error("serialNumber must be omitted for determinism")
	}
}
```

- [ ] **Step 2: Run, verify fail:** `go test ./pkg/sbom/format/ -v` → FAIL.
- [ ] **Step 3a: Implement the boundary** `pkg/sbom/format/format.go`:

```go
package format

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"sort"

	"github.com/inovacc/omni/pkg/sbom/model"
	"github.com/inovacc/omni/pkg/sbom/purl"
)

// Kind selects the output document format.
type Kind int

const (
	// SPDX selects SPDX 2.3 JSON.
	SPDX Kind = iota
	// CycloneDX selects CycloneDX 1.5 JSON.
	CycloneDX
)

const defaultDate = "1970-01-01T00:00:00Z"

// Options tunes document generation. SourceDate (RFC-3339) is the fixed
// creation timestamp; empty means the epoch default (keeps output deterministic
// with no flag). OmniVersion labels the generating tool.
type Options struct {
	OmniVersion string
	SourceDate  string
}

// entry is the internal, fully-resolved view of one component.
type entry struct {
	c    model.Component
	purl string
	slug string
}

// Document is the STABLE cross-package boundary type. pkg/scan/ (Phase 6)
// depends only on this type, never on pkg/sbom/model. It carries a resolved,
// sorted, format-agnostic snapshot ready for deterministic emission.
type Document struct {
	name        string
	created     string
	omniVersion string
	root        entry
	entries     []entry // sorted by purl/slug; excludes root
	contentHash string  // 16 hex chars over sorted purls
}

// From resolves a model.SBOM into a Document (purls + slugs computed, slices
// sorted, content hash derived). The input is normalized defensively.
func From(s *model.SBOM, opt Options) *Document {
	s.Normalize()
	created := opt.SourceDate
	if created == "" {
		created = defaultDate
	}
	d := &Document{
		name:        s.Name,
		created:     created,
		omniVersion: opt.OmniVersion,
		root:        toEntry(s.Root),
	}
	slugs := map[string]int{}
	dedupeSlug(&d.root, slugs)
	purls := make([]string, 0, len(s.Components))
	for _, c := range s.Components {
		e := toEntry(c)
		dedupeSlug(&e, slugs)
		d.entries = append(d.entries, e)
		purls = append(purls, e.purl)
	}
	sort.Slice(d.entries, func(i, j int) bool { return d.entries[i].purl < d.entries[j].purl })
	sort.Strings(purls)
	h := sha256.New()
	for _, p := range purls {
		_, _ = h.Write([]byte(p))
		_, _ = h.Write([]byte{'\n'})
	}
	d.contentHash = hex.EncodeToString(h.Sum(nil))[:16]
	return d
}

func toEntry(c model.Component) entry {
	return entry{c: c, purl: purl.ForModule(c.Path, c.Version), slug: model.Slug(c.Path)}
}

// dedupeSlug ensures slugs are unique by appending -<n> on collision.
func dedupeSlug(e *entry, seen map[string]int) {
	base := e.slug
	if n, ok := seen[base]; ok {
		seen[base] = n + 1
		e.slug = base + "-" + itoa(n+1)
	} else {
		seen[base] = 0
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// Encode writes the document in the requested format to w (deterministic bytes,
// trailing newline).
func (d *Document) Encode(w io.Writer, k Kind) error {
	switch k {
	case SPDX:
		return d.encodeSPDX(w)
	case CycloneDX:
		return d.encodeCycloneDX(w)
	default:
		return errUnknownFormat
	}
}
```

Add to `format.go` a sentinel: `var errUnknownFormat = errors.New("unknown sbom format")` (import `errors`).

- [ ] **Step 3b: Implement SPDX emitter** `pkg/sbom/format/spdx.go` — declare minimal structs in the exact JSON-field order from the format section, build the value (root package first, deps after, `DESCRIBES` then `DEPENDS_ON` relationships sorted by `relatedSpdxElement`), then:

```go
func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v) // Encode appends a trailing newline.
}
```

Struct skeleton (fields in on-disk order; `omitempty` only where the format section allows):

```go
type spdxDoc struct {
	SPDXVersion       string             `json:"spdxVersion"`
	DataLicense       string             `json:"dataLicense"`
	SPDXID            string             `json:"SPDXID"`
	Name              string             `json:"name"`
	DocumentNamespace string             `json:"documentNamespace"`
	CreationInfo      spdxCreation       `json:"creationInfo"`
	Packages          []spdxPackage      `json:"packages"`
	Relationships     []spdxRelationship `json:"relationships"`
}
type spdxCreation struct {
	Created  string   `json:"created"`
	Creators []string `json:"creators"`
}
type spdxExtRef struct {
	ReferenceCategory string `json:"referenceCategory"`
	ReferenceType     string `json:"referenceType"`
	ReferenceLocator  string `json:"referenceLocator"`
}
type spdxPackage struct {
	Name             string       `json:"name"`
	SPDXID           string       `json:"SPDXID"`
	VersionInfo      string       `json:"versionInfo,omitempty"`
	DownloadLocation string       `json:"downloadLocation"`
	FilesAnalyzed    bool         `json:"filesAnalyzed"`
	LicenseConcluded string       `json:"licenseConcluded"`
	LicenseDeclared  string       `json:"licenseDeclared"`
	CopyrightText    string       `json:"copyrightText"`
	ExternalRefs     []spdxExtRef `json:"externalRefs,omitempty"`
}
type spdxRelationship struct {
	SPDXElementID      string `json:"spdxElementId"`
	RelatedSPDXElement string `json:"relatedSpdxElement"`
	RelationshipType   string `json:"relationshipType"`
}
```

`encodeSPDX` builds `spdxDoc{SPDXVersion:"SPDX-2.3", DataLicense:"CC0-1.0", SPDXID:"SPDXRef-DOCUMENT", Name:d.name, DocumentNamespace:"https://spdx.org/spdxdocs/omni/"+d.name+"-"+d.contentHash, CreationInfo:{Created:d.created, Creators:[]string{"Tool: omni-sbom-"+d.omniVersion}}, ...}`. Each package's `ExternalRefs` is `[]spdxExtRef{{"PACKAGE-MANAGER","purl",e.purl}}` (omit if purl empty). Package SPDXID = `"SPDXRef-Package-"+e.slug`. Root package emitted first, then deps in `d.entries` order. Relationships: one `DESCRIBES` (DOCUMENT→rootSlug), then one `DEPENDS_ON` (rootSlug→depSlug) per entry, already sorted because `d.entries` is sorted by purl.

- [ ] **Step 3c: Implement CycloneDX emitter** `pkg/sbom/format/cyclonedx.go` — minimal structs in on-disk order, reuse `writeJSON`:

```go
type cdxBOM struct {
	Schema       string          `json:"$schema"`
	BOMFormat    string          `json:"bomFormat"`
	SpecVersion  string          `json:"specVersion"`
	Version      int             `json:"version"`
	Metadata     cdxMetadata     `json:"metadata"`
	Components   []cdxComponent  `json:"components"`
	Dependencies []cdxDependency `json:"dependencies"`
}
type cdxMetadata struct {
	Timestamp string        `json:"timestamp"`
	Tools     cdxTools      `json:"tools"`
	Component cdxComponent  `json:"component"`
}
type cdxTools struct {
	Components []cdxComponent `json:"components"`
}
type cdxComponent struct {
	BOMRef  string `json:"bom-ref,omitempty"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	PURL    string `json:"purl,omitempty"`
}
type cdxDependency struct {
	Ref       string   `json:"ref"`
	DependsOn []string `json:"dependsOn,omitempty"`
}
```

`encodeCycloneDX` sets `Schema:"http://cyclonedx.org/schema/bom-1.5.schema.json"`, `BOMFormat:"CycloneDX"`, `SpecVersion:"1.5"`, `Version:1`. Tool component = `{Type:"application", Name:"omni", Version:d.omniVersion}` (no purl/bom-ref). Metadata.Component = root entry as `{BOMRef:rootRef, Type:"application", Name:root.c.Path, Version:root.c.Version, PURL:root.purl}` where `rootRef = root.purl` (or `root.slug` if purl empty). Components = `d.entries` mapped to `{BOMRef:e.purl, Type:"library", Name:e.c.Path, Version:e.c.Version, PURL:e.purl}` for `KindLibrary`, `Type:"application"` for `KindToolchain`. Dependencies: first `{Ref:rootRef, DependsOn:[sorted e.purl ...]}`, then one `{Ref:e.purl}` per entry. **No `serialNumber` field exists on the struct** → guaranteed omitted.

Create `pkg/sbom/format/doc.go` — package docstring; **no `// Experimental:` marker** (this is the stable boundary). State clearly: "Document is the stable v1.0 boundary type; importers outside pkg/sbom must depend only on this package."

- [ ] **Step 4: Run, verify pass:** `go test ./pkg/sbom/format/ -v` → PASS (determinism + both shapes).
- [ ] **Step 5: Commit:** `gofmt -w pkg/sbom/format/ && git commit -- pkg/sbom/format -m "feat(sbom): stable format.Document boundary + deterministic SPDX-2.3/CycloneDX-1.5 emitters"`

---

### Task 5: `pkg/sbom/collect` — module-dir collector from `go.mod`/`go.sum` (TDD)

**Files:** Create `pkg/sbom/collect/module.go`, `pkg/sbom/collect/module_test.go`, `pkg/sbom/collect/doc.go`.

We parse `go.mod` ourselves with a tiny line scanner (no new dep — `golang.org/x/mod/modfile` would pull `golang.org/x/mod` packages we are NOT promoting beyond `semver`; a line scanner is sufficient for `module`, `go`, and `require` blocks and stays stdlib). `go.sum` is consulted only to confirm a module is actually used (presence) — versions come from `go.mod` `require` lines.

- [ ] **Step 1: Write failing tests** (`pkg/sbom/collect/module_test.go`) — build a temp module dir:

```go
package collect_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/inovacc/omni/pkg/sbom/collect"
	"github.com/inovacc/omni/pkg/sbom/model"
)

const goMod = `module github.com/example/app

go 1.25.0

require (
	github.com/spf13/cobra v1.10.2
	golang.org/x/mod v0.36.0 // indirect
)

require github.com/single/dep v1.2.3
`

func writeMod(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestModuleDir(t *testing.T) {
	dir := t.TempDir()
	writeMod(t, dir, goMod)
	sb, err := collect.ModuleDir(dir)
	if err != nil {
		t.Fatalf("ModuleDir: %v", err)
	}
	if sb.Root.Path != "github.com/example/app" || sb.Root.Kind != model.KindRoot {
		t.Errorf("root = %+v", sb.Root)
	}
	sb.Normalize()
	got := map[string]string{}
	for _, c := range sb.Components {
		got[c.Path] = c.Version
	}
	for path, ver := range map[string]string{
		"github.com/spf13/cobra": "v1.10.2",
		"golang.org/x/mod":       "v0.36.0",
		"github.com/single/dep":  "v1.2.3",
	} {
		if got[path] != ver {
			t.Errorf("dep %s = %q, want %q", path, got[path], ver)
		}
	}
}

func TestModuleDirMissing(t *testing.T) {
	if _, err := collect.ModuleDir(t.TempDir()); err == nil {
		t.Error("expected error for missing go.mod")
	}
}
```

- [ ] **Step 2: Run, verify fail:** `go test ./pkg/sbom/collect/ -run ModuleDir -v` → FAIL.
- [ ] **Step 3: Implement** `pkg/sbom/collect/module.go`. `ModuleDir(dir string) (*model.SBOM, error)`: open `filepath.Join(dir, "go.mod")` (missing → `fmt.Errorf("go.mod: %w", os.ErrNotExist)` so the CLI maps to `ErrNotFound`); scan line-by-line with `bufio.Scanner`:
  - `module <path>` → `sb.Root = model.Component{Path:path, Kind:KindRoot}` and `sb.Name = lastPathElement(path)`.
  - Track a `inRequireBlock` bool toggled by a line equal to `require (` / `)`. Inside the block OR on a `require <path> <ver>` single line, parse `fields := strings.Fields(line)`; skip a leading `require` token; if `len(fields) >= 2` and `fields[0]` looks like a module path (contains `.` or `/`) take `fields[0]`,`fields[1]` as path,version; ignore a trailing `// indirect` comment (strip at `//`). Append `model.Component{Path, Version, Kind:KindLibrary}`.
  - Ignore `go`, `toolchain`, `replace`, `exclude`, `retract` directive lines for v1.0 (note: replace handling for module-dir SBOMs is binary-only in this phase; document the limitation in `doc.go`).
  - Return `&sb` (do NOT call Normalize here — leave to the emitter/caller; tests call it explicitly).

Define `lastPathElement(p string) string { if i := strings.LastIndexByte(p, '/'); i >= 0 { return p[i+1:] }; return p }`. Create `pkg/sbom/collect/doc.go` (docstring + `// Experimental:`).

- [ ] **Step 4: Run, verify pass:** `go test ./pkg/sbom/collect/ -run ModuleDir -v` → PASS.
- [ ] **Step 5: Commit:** `gofmt -w pkg/sbom/collect/ && git commit -- pkg/sbom/collect -m "feat(sbom): module-dir collector parsing go.mod require blocks"`

---

### Task 6: `pkg/sbom/collect` — binary collector via `debug/buildinfo` (TDD)

**Files:** Modify `pkg/sbom/collect/`: add `pkg/sbom/collect/binary.go`, add tests to `pkg/sbom/collect/module_test.go` (or a new `binary_test.go`).

- [ ] **Step 1: Write failing test** (`pkg/sbom/collect/binary_test.go`) — drive `Binary` off an in-memory `*debug.BuildInfo` to avoid building a real binary in unit tests; `Binary` takes the parsed info so it is pure and testable, and `BinaryFile(path)` is the thin `buildinfo.ReadFile` wrapper:

```go
package collect_test

import (
	"runtime/debug"
	"testing"

	"github.com/inovacc/omni/pkg/sbom/collect"
	"github.com/inovacc/omni/pkg/sbom/model"
)

func TestBinaryFromBuildInfo(t *testing.T) {
	bi := &debug.BuildInfo{
		GoVersion: "go1.25.0",
		Path:      "github.com/inovacc/omni",
		Main:      debug.Module{Path: "github.com/inovacc/omni", Version: "v1.0.0"},
		Deps: []*debug.Module{
			{Path: "github.com/spf13/cobra", Version: "v1.10.2"},
			{
				Path:    "github.com/old/mod",
				Version: "v1.0.0",
				Replace: &debug.Module{Path: "github.com/new/mod", Version: "v2.0.0"},
			},
		},
	}
	sb := collect.Binary(bi)
	if sb.Root.Path != "github.com/inovacc/omni" || sb.Root.Version != "v1.0.0" {
		t.Errorf("root = %+v", sb.Root)
	}
	var sawToolchain, sawReplaced bool
	for _, c := range sb.Components {
		if c.Kind == model.KindToolchain && c.Path == "std" && c.Version == "go1.25.0" {
			sawToolchain = true
		}
		if c.Path == "github.com/new/mod" && c.Version == "v2.0.0" && c.OriginalPath == "github.com/old/mod" {
			sawReplaced = true
		}
	}
	if !sawToolchain {
		t.Error("toolchain component (std@go1.25.0) missing")
	}
	if !sawReplaced {
		t.Error("replace directive not resolved to effective module")
	}
}
```

- [ ] **Step 2: Run, verify fail:** `go test ./pkg/sbom/collect/ -run Binary -v` → FAIL.
- [ ] **Step 3: Implement** `pkg/sbom/collect/binary.go`:

```go
package collect

import (
	"debug/buildinfo"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/inovacc/omni/pkg/sbom/model"
)

// BinaryFile reads build information embedded in a Go binary at path and builds
// an SBOM describing what actually shipped. A non-Go binary (or unreadable file)
// returns an error the CLI maps to ErrInvalidInput / ErrNotFound.
func BinaryFile(path string) (*model.SBOM, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("read binary %q: %w", path, err)
	}
	bi, err := buildinfo.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("parse build info from %q: %w", path, err)
	}
	return Binary(bi), nil
}

// Binary converts already-parsed build info into an SBOM. The Go toolchain is
// added as a KindToolchain component; replaced modules resolve to the effective
// (replacement) module while preserving the original path/version.
func Binary(bi *debug.BuildInfo) *model.SBOM {
	sb := &model.SBOM{
		Name: lastPathElement(bi.Main.Path),
		Root: model.Component{Path: bi.Main.Path, Version: bi.Main.Version, Kind: model.KindRoot},
	}
	sb.Components = append(sb.Components, model.Component{
		Path: "std", Version: bi.GoVersion, Kind: model.KindToolchain,
	})
	for _, d := range bi.Deps {
		c := model.Component{Path: d.Path, Version: d.Version, Kind: model.KindLibrary}
		if d.Replace != nil {
			c.OriginalPath, c.OriginalVersion = d.Path, d.Version
			c.Path, c.Version = d.Replace.Path, d.Replace.Version
		}
		sb.Components = append(sb.Components, c)
	}
	return sb
}
```

- [ ] **Step 4: Run, verify pass:** `go test ./pkg/sbom/collect/ -run Binary -v` → PASS.
- [ ] **Step 5: Commit:** `gofmt -w pkg/sbom/collect/ && git commit -- pkg/sbom/collect -m "feat(sbom): binary collector via debug/buildinfo (toolchain + replace resolution)"`

---

### Task 7: CLI glue `internal/cli/sbom` + Cobra wrapper `omni sbom` (TDD)

**Files:** Create `internal/cli/sbom/sbom.go`, `internal/cli/sbom/sbom_test.go`, `cmd/sbom.go`.

`RunSBOM(w io.Writer, args []string, opts SBOMOptions) error`. `SBOMOptions{ Format string /* "spdx"|"cyclonedx" */, From string /* "auto"|"module"|"binary" */, SourceDate string, OmniVersion string, Sign bool, KeyPath string }`. Auto-detect: if the path is a directory → module; if a regular file → binary; mismatch with explicit `--from` → `ErrInvalidInput`.

- [ ] **Step 1: Write failing test** (`internal/cli/sbom/sbom_test.go`):

```go
package sbom_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
	clisbom "github.com/inovacc/omni/internal/cli/sbom"
)

func TestRunSBOMModuleSPDX(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "go.mod"),
		[]byte("module github.com/example/app\n\ngo 1.25.0\n\nrequire github.com/spf13/cobra v1.10.2\n"), 0o644)
	var buf bytes.Buffer
	err := clisbom.RunSBOM(&buf, []string{dir}, clisbom.SBOMOptions{Format: "spdx", OmniVersion: "v0.1.0"})
	if err != nil {
		t.Fatalf("RunSBOM: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["spdxVersion"] != "SPDX-2.3" {
		t.Errorf("spdxVersion = %v", m["spdxVersion"])
	}
}

func TestRunSBOMBadFormat(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n\ngo 1.25.0\n"), 0o644)
	err := clisbom.RunSBOM(&bytes.Buffer{}, []string{dir}, clisbom.SBOMOptions{Format: "xml"})
	if !cmderr.IsInvalidInput(err) {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

func TestRunSBOMMissingPath(t *testing.T) {
	err := clisbom.RunSBOM(&bytes.Buffer{}, []string{filepath.Join(t.TempDir(), "nope")}, clisbom.SBOMOptions{Format: "spdx"})
	if !cmderr.IsNotFound(err) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
```

- [ ] **Step 2: Run, verify fail:** `go test ./internal/cli/sbom/ -v` → FAIL.
- [ ] **Step 3: Implement** `internal/cli/sbom/sbom.go`:
  - Validate `Format`: map `"spdx"`→`format.SPDX`, `"cyclonedx"`/`"cdx"`→`format.CycloneDX`, default `"spdx"` when empty; anything else → `cmderr.Wrap(cmderr.ErrInvalidInput, "unknown --format (want spdx|cyclonedx)")`.
  - Resolve path arg (require exactly one; else `ErrInvalidInput`). `os.Stat`: `os.IsNotExist` → `cmderr.Wrap(cmderr.ErrNotFound, ...)`; permission error → `cmderr.Wrap(cmderr.ErrPermission, ...)`.
  - Choose collector: dir → `collect.ModuleDir`; file → `collect.BinaryFile`; honor explicit `From` and conflict-check. Translate collector errors: `errors.Is(err, os.ErrNotExist)` → `ErrNotFound`; otherwise (parse failure / not a Go binary / malformed go.mod) → `cmderr.Wrap(cmderr.ErrInvalidInput, ...)`.
  - `doc := format.From(sb, format.Options{OmniVersion: opts.OmniVersion, SourceDate: opts.SourceDate})`.
  - If `opts.Sign`: encode to an in-memory buffer first, then write the buffer to `w`; reuse `pkg/sign`: require `opts.KeyPath != ""` (else `cmderr.Wrap(cmderr.ErrInvalidInput, "--sign requires --key")`); read passphrase from `OMNI_SIGN_PASSPHRASE` (same convention as Phase 04 — NEVER a flag); produce `<output>.minisig` next to the path implied by `--out`, or document that with stdout `--sign` writes the signature to `<stdoutCapturedToFile>` is unsupported → if no `--out`, return `cmderr.Wrap(cmderr.ErrInvalidInput, "--sign requires --out")`. (Keep signing file-based; reuse `sign.RunSign` or `pkg/sign.Sign` over the emitted bytes.)
  - Else: `if err := doc.Encode(w, k); err != nil { return cmderr.Wrap(cmderr.ErrIO, "write sbom") }`.
- [ ] **Step 4: Cobra wrapper** `cmd/sbom.go`:

```go
package cmd

import (
	"github.com/inovacc/omni/internal/cli/sbom"
	"github.com/spf13/cobra"
)

var sbomCmd = &cobra.Command{
	Use:   "sbom [OPTION]... PATH",
	Short: "Generate an SBOM (SPDX 2.3 or CycloneDX 1.5) for a Go module dir or binary",
	Long: `Generate a byte-deterministic Software Bill of Materials for a Go module
directory (go.mod) or a built Go binary (debug/buildinfo). Every component
carries a normalized Go purl. Output is identical bytes for identical input,
enabling reproducible-build and golden-master pinning.

      --format spdx|cyclonedx   output format (default: spdx)
      --from   auto|module|binary  source kind (default: auto-detect from PATH)
      --source-date RFC3339     fixed creation timestamp (default: epoch)
      --out FILE                write to FILE instead of stdout
      --sign                    sign --out with a minisign key (requires --key, --out)
  -k, --key FILE                secret key path for --sign (passphrase via OMNI_SIGN_PASSPHRASE)

Examples:
  omni sbom . --format spdx
  omni sbom ./bin/omni --format cyclonedx
  omni sbom . --format spdx --out omni.spdx.json --sign --key release.key`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := sbom.SBOMOptions{OmniVersion: rootVersion()} // rootVersion(): reuse the existing version accessor in cmd/
		opts.Format, _ = cmd.Flags().GetString("format")
		opts.From, _ = cmd.Flags().GetString("from")
		opts.SourceDate, _ = cmd.Flags().GetString("source-date")
		opts.OutPath, _ = cmd.Flags().GetString("out")
		opts.Sign, _ = cmd.Flags().GetBool("sign")
		opts.KeyPath, _ = cmd.Flags().GetString("key")
		return sbom.RunSBOM(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(sbomCmd)
	sbomCmd.Flags().String("format", "spdx", "output format: spdx|cyclonedx")
	sbomCmd.Flags().String("from", "auto", "source kind: auto|module|binary")
	sbomCmd.Flags().String("source-date", "", "fixed RFC-3339 creation timestamp (default: epoch)")
	sbomCmd.Flags().String("out", "", "write to FILE instead of stdout")
	sbomCmd.Flags().Bool("sign", false, "sign --out with a minisign key")
	sbomCmd.Flags().StringP("key", "k", "", "secret key path for --sign")
}
```

  (If `cmd/` has no `rootVersion()` accessor, add `OmniVersion` defaulting to the package-level version var used by `cmd/root.go`/`version.go`; grep `version` in `cmd/` during implementation and reuse it. Add `OutPath string` to `SBOMOptions`; when set, open the file with `os.Create` and write there instead of `w`, mapping create/write failure → `ErrIO`/`ErrPermission`.)
- [ ] **Step 5: Run, verify pass + manual smoke:** `go test ./internal/cli/sbom/ -v`; then `go run . sbom . --format spdx | head -5` and `go run . sbom . --format cyclonedx | head -5`; build self and `go run . sbom <built-binary> --format cyclonedx`.
- [ ] **Step 6: Commit:** `gofmt -w internal/cli/sbom cmd/sbom.go && git commit -- internal/cli/sbom cmd/sbom.go -m "feat(sbom): omni sbom CLI (spdx|cyclonedx, module|binary, --out, --sign via pkg/sign)"`

---

### Task 8: Optional schema validator behind `//go:build omni_sbomvalidate`

**Files:** Create `internal/cli/sbom/validate_on.go` (`//go:build omni_sbomvalidate`), `internal/cli/sbom/validate_off.go` (`//go:build !omni_sbomvalidate`); add a `--validate` bool flag in `cmd/sbom.go`.

- [ ] **Step 1: Default (no-tag) failing test** (`internal/cli/sbom/sbom_test.go`) — `RunSBOM` with `Validate:true` and no tag returns `cmderr.IsUnsupported(err)`:

```go
func TestRunSBOMValidateNoTag(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n\ngo 1.25.0\n"), 0o644)
	err := clisbom.RunSBOM(&bytes.Buffer{}, []string{dir}, clisbom.SBOMOptions{Format: "spdx", Validate: true, OmniVersion: "v0.1.0"})
	if !cmderr.IsUnsupported(err) {
		t.Errorf("err = %v, want ErrUnsupported", err)
	}
}
```

- [ ] **Step 2: Implement the no-tag stub** `validate_off.go`:

```go
//go:build !omni_sbomvalidate

package sbom

import "github.com/inovacc/omni/internal/cli/cmderr"

func validateDocument(_ []byte, _ string) error {
	return cmderr.Wrap(cmderr.ErrUnsupported, "sbom schema validation requires building with -tags omni_sbomvalidate")
}
```

In `RunSBOM`, after producing the bytes (encode to a buffer when `Validate` is set), call `validateDocument(buf.Bytes(), format)` and return its error before writing output.

- [ ] **Step 3: Implement the tagged path** `validate_on.go` (`//go:build omni_sbomvalidate`): decode the emitted JSON with the upstream libraries to prove schema-validity — SPDX via `spdxjson "github.com/spdx/tools-golang/json"` (`spdxjson.Read(bytes.NewReader(b))` returns `*v2_3.Document`, error on malformed); CycloneDX via `cdx "github.com/CycloneDX/cyclonedx-go"` (`cdx.NewBOMDecoder(bytes.NewReader(b), cdx.BOMFileFormatJSON).Decode(new(cdx.BOM))`). Any decode error → `cmderr.Wrap(cmderr.ErrConflict, "emitted SBOM failed schema validation: "+err.Error())`. Add the deps ONLY here: `go get github.com/spdx/tools-golang@v0.5.7 github.com/CycloneDX/cyclonedx-go@v0.11.0 && go mod tidy`.
- [ ] **Step 4: Add `--validate` flag** in `cmd/sbom.go` (`sbomCmd.Flags().Bool("validate", false, "validate emitted document against the upstream schema (requires -tags omni_sbomvalidate)")`) and wire `opts.Validate, _ = cmd.Flags().GetBool("validate")`.
- [ ] **Step 5: Verify BOTH builds:** `go build ./...` (default — validator deps NOT compiled) and `go build -tags omni_sbomvalidate ./...` (compiles validator). Run `go test ./internal/cli/sbom/...` (default → unsupported path passes) and `go test -tags omni_sbomvalidate ./internal/cli/sbom/...` (tagged → a positive test that `Validate:true` succeeds for a real module dir).
- [ ] **Step 6: Commit:** `git commit -- internal/cli/sbom cmd/sbom.go go.mod go.sum -m "feat(sbom): optional upstream-schema validator behind omni_sbomvalidate (default: ErrUnsupported)"`

---

### Task 9: `omni pipe` registration

**Files:** Modify `cmd/pipe.go`.

`sbom` takes a PATH argument and ignores stdin; it still benefits from registry dispatch. Register it via `AdaptWriterReaderArgs` with a reader-ignoring closure so it participates in the unified registry like `sign`.

- [ ] **Step 1: Failing test** — mirror an existing pipe test (`cmd/`): a registry lookup of `"sbom"` resolves and, given `args=[<tempModuleDir>]`, writes SPDX JSON to the writer (reader unused).
- [ ] **Step 2: Implement** — add the import `"github.com/inovacc/omni/internal/cli/sbom"` and in `buildPipeRegistry()`:

```go
reg.Register("sbom", command.AdaptWriterReaderArgs(
	func(w io.Writer, _ io.Reader, args []string) error {
		return sbom.RunSBOM(w, args, sbom.SBOMOptions{Format: "spdx", From: "auto", OmniVersion: rootVersion()})
	},
))
```

- [ ] **Step 3: Run, verify pass:** `go test ./cmd/... -run Pipe -count=1`.
- [ ] **Step 4: Commit:** `git commit -- cmd/pipe.go -m "feat(sbom): register sbom in the pipe unified registry"`

---

### Task 10: Golden-master tests (deterministic, fixtures)

**Files:** Modify `testing/golden/golden_tests.yaml` AND `tools/golden/golden_tests.yaml` (keep in sync); add fixtures under `testing/golden/fixtures/sbom/` via a `gen_fixtures.go` (`//go:build ignore`) generator mirroring the `sign` fixtures pattern.

Output embeds a wall-clock-free timestamp only when `--source-date` is omitted (epoch default), so SBOM stdout is already deterministic; we still pin `--source-date 1970-01-01T00:00:00Z` explicitly in the golden args for clarity. The module dir fixture must be a FIXED `go.mod` (committed) so the dependency set never drifts.

- [ ] **Step 1: Create fixtures.** `testing/golden/fixtures/sbom/go.mod.fixture` (a committed, frozen go.mod with 2–3 pinned requires — NOT the repo's live go.mod), plus `testing/golden/fixtures/sbom/gen_fixtures.go` (`//go:build ignore`) that copies `go.mod.fixture` → a temp dir as `go.mod` is unnecessary; instead the golden test points `args` at a committed dir `testing/golden/fixtures/sbom/mod/` containing exactly `go.mod` (frozen). Create `testing/golden/fixtures/sbom/mod/go.mod`:

```
module github.com/example/golden-app

go 1.25.0

require (
	github.com/spf13/cobra v1.10.2
	golang.org/x/mod v0.36.0
)
```

  (`gen_fixtures.go` documents the freeze and, if ever needed, re-materializes the dir; it is run by hand, never in CI.)
- [ ] **Step 2: Add an `sbom` category to BOTH yaml files:**

```yaml
  - name: sbom
    tests:
      # SPDX module SBOM is deterministic (source-date pinned).
      - name: spdx_module
        args: ["sbom", "{fixtures}/mod", "--format", "spdx", "--source-date", "1970-01-01T00:00:00Z"]
        exit_code: 0
        normalizations: ["strip_path"]

      # CycloneDX module SBOM is deterministic.
      - name: cyclonedx_module
        args: ["sbom", "{fixtures}/mod", "--format", "cyclonedx", "--source-date", "1970-01-01T00:00:00Z"]
        exit_code: 0
        normalizations: ["strip_path"]

      # Negative: unknown format -> ErrInvalidInput (exit 2).
      - name: bad_format
        args: ["sbom", "{fixtures}/mod", "--format", "xml"]
        exit_code: 2
        normalizations: ["strip_path"]

      # Negative: missing path -> ErrNotFound (exit 1).
      - name: missing_path
        args: ["sbom", "{fixtures}/does-not-exist", "--format", "spdx"]
        exit_code: 1
        normalizations: ["strip_path"]

      # Negative: --validate without the build tag -> ErrUnsupported (exit 6).
      - name: validate_no_tag
        args: ["sbom", "{fixtures}/mod", "--format", "spdx", "--validate"]
        exit_code: 6
        normalizations: ["strip_path"]
```

- [ ] **Step 3: Generate + verify snapshots:** `task test:golden:update && task golden:record`, then `python testing/scripts/test_golden.py` → all green. Confirm the recorded SPDX/CycloneDX snapshots contain `pkg:golang/github.com/spf13/cobra@v1.10.2` and have NO `serialNumber`.
- [ ] **Step 4: Commit:** `git commit -- testing/golden tools/golden -m "test(sbom): golden-master deterministic spdx/cyclonedx + fail-closed negatives"`

---

### Task 11: Docs + final gate

**Files:** `docs/COMMANDS.md`, `CLAUDE.md` (command-inventory count line), `docs/superpowers/specs/2026-05-16-05-pkg-sbom-design.md` (status), `docs/EXIT-CODES.md` (no change expected — sbom reuses existing sentinels; verify).

- [ ] **Step 1: Docs** — add `sbom` to `docs/COMMANDS.md` (with `--format`, `--from`, `--source-date`, `--out`, `--sign`, `--validate`) and bump the CLAUDE.md inventory count; note the `omni_sbomvalidate` build tag and the `pkg/sbom/format.Document` stable boundary (mention Phase 6 `pkg/scan` will import it). Run `omni aicontext` / `omni cmdtree` regen if applicable. Mark the spec `Status: Complete`.
- [ ] **Step 2: Final gate:**
```bash
go build ./... && go build -tags omni_sbomvalidate ./...
go vet ./... && gofmt -l pkg/sbom internal/cli/sbom cmd/sbom.go
golangci-lint run --timeout=5m ./...
go test ./pkg/sbom/... ./internal/cli/sbom/... -count=1
go test -tags omni_sbomvalidate ./internal/cli/sbom/... -count=1
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ./... && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./...
python testing/scripts/test_golden.py
# determinism cross-check (must print identical hashes):
go run . sbom . --format spdx --source-date 1970-01-01T00:00:00Z | sha256sum
go run . sbom . --format spdx --source-date 1970-01-01T00:00:00Z | sha256sum
# external oracle (if syft is available in CI image):
go run . sbom . --format spdx --source-date 1970-01-01T00:00:00Z --out /tmp/omni.spdx.json && syft convert /tmp/omni.spdx.json -o cyclonedx-json >/dev/null
```
Expected: all green; the two `sha256sum` lines match (byte-determinism); `syft convert` exits 0 (schema-valid).
- [ ] **Step 3: Commit:** `git commit -- docs/ CLAUDE.md -m "docs(sbom): document omni sbom; mark Phase 05 complete"`

---

## Self-Review

**Spec coverage** (SBOM-01…10 / success criteria → tasks):
- SPDX 2.3 + CycloneDX 1.5 valid JSON, `syft convert`-parseable → **T4** (emitters) + **T7** (CLI) + **T8** (in-process validator) + **T11** (syft oracle in final gate). (Success criterion 1.)
- Binary SBOM from `debug/buildinfo`, Go toolchain listed as a component → **T6** (`std@go<ver>` toolchain component) + **T7** (binary path routing). (Success criterion 2.)
- Every component a correctly normalized purl; purl round-trip drift guard → **T2** (`purl.ForModule`, canonical/`+incompatible`/pseudo/empty cases) + **T1** (ADR locks the rule). (Success criterion 3.)
- Byte-deterministic output (two runs identical bytes) → **T4** (`TestEmitDeterministic`, struct-order + pre-sorted slices + no `serialNumber`/UUID, epoch-default ts) + **T3** (`Normalize`) + **T10** (golden pinning) + **T11** (sha256 cross-check). (Success criterion 4.)
- `pkg/sbom/format.Document` stable boundary; `pkg/scan` imports it not `model` → **T4** (boundary type + no `// Experimental:` on `format/doc.go`; `model`/`collect`/`purl` marked Experimental) + **T1** (ADR records the constraint).
- Pure-Go, no exec, no heavy default deps; heavy libs tag-gated → **T4** (stdlib `encoding/json` emitter) + **T8** (`omni_sbomvalidate`) + **T1** (ADR rationale). `golang.org/x/mod` already in `go.sum` → no new transitive cost (**T2** promotes to direct require only).
- Sign the SBOM by reusing Phase 04 → **T7** (`--sign` via `pkg/sign`, passphrase via `OMNI_SIGN_PASSPHRASE`, never a flag).
- ADR gate (1 ADR before code) → **T1** (hard gate).

**Placeholder scan:** No "TBD"/"add validation"/"handle edge cases". Every type is defined before use: `model.Component`/`SBOM`/`Slug`/`Kind` (T3) precede their consumers `format.entry`/`Document` (T4) and `collect` (T5/T6); `format.Kind`/`Options`/`Document.Encode` (T4) precede `internal/cli/sbom` (T7); `purl.ForModule` (T2) precedes `format.toEntry` (T4). Emitter struct shapes are given byte-exactly (field names, JSON tags, on-disk order, fixed string constants) in the format section and Task 4. The validator API (T8) is the concrete `spdxjson.Read` / `cdx.NewBOMDecoder(...).Decode` chain confirmed from the indexed library docs. Collector parsing rules (T5) are spelled out line-by-line; binary rules (T6) use the confirmed `debug.BuildInfo`/`Module`/`Replace` fields.

**Type consistency:** `model.Component.Kind` (T3) drives `cdxComponent.Type` (`KindToolchain`→`application`, else `library`) and SPDX package emission in T4. `format.From` consumes `*model.SBOM` and `purl.ForModule` (T2). `collect.ModuleDir`/`BinaryFile` (T5/T6) return `*model.SBOM` consumed by `format.From` in T7. `SBOMOptions.Format` strings map to `format.Kind` in T7. `validateDocument([]byte,string) error` has matching signatures in both `validate_on.go`/`validate_off.go` (T8). `rootVersion()`/version var reused (not invented) in T7/T9 — implementer greps `cmd/` to bind the existing accessor. cmderr classes match `docs/EXIT-CODES.md`: ErrNotFound(1), ErrInvalidInput(2), ErrPermission(3), ErrIO(4), ErrUnsupported(6), ErrConflict(1).

**Known risks:** (1) **purl percent-encoding assumption** — Go module paths are asserted to need no encoding once lowercased; `TestForModuleRejectsEncodingNeeded` (T2) guards it, but an exotic path with a `%`/space would slip through. Mitigation: if a future module path violates this, add RFC-3986 segment encoding in `purl.ForModule` (isolated, one function). (2) **`go.mod` line-scanner vs `modfile`** — the hand scanner ignores `replace` for module-dir SBOMs (binary SBOMs DO honor replace via buildinfo). Documented in `collect/doc.go`; acceptable for v1.0 since "what shipped" (binary path) is the authoritative replace-aware source. (3) **upstream-lib schema drift** — pinned at spdx v0.5.7 / cyclonedx v0.11.0 (confirmed in research); the validator is tag-gated so a future bump never touches default builds, but CI must run BOTH the tagged validator and the external `syft convert` oracle (T11) to catch real-world scanner mismatches. (4) **`syft` availability** — the external oracle in T11 is best-effort (guarded "if available"); the in-process `omni_sbomvalidate` validator (T8) is the always-on gate.
