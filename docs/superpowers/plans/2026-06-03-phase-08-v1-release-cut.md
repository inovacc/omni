# Phase 08 — v1.0 Release Cut Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.
> **HARD GATE — SATISFIED (2026-06-03):** ADR-0010 is written and accepted at `docs/adr/ADR-0010-v1-release-policy.md` (API freeze of non-Experimental `pkg/*`, reproducibility-or-fail CI gate, dogfooded `omni sign/sbom/attest` assets, honest SLSA-L2 + no-overclaim announce). Pipeline/CI tasks (2+) proceed. NOTE: this phase delivers + commits the release MACHINERY; cutting the live `v1.0.0` tag / GitHub release (which requires GitHub Actions enabled) remains a deliberate operator trigger, not part of the merge.

**Goal:** Cut omni v1.0 as **six signed, reproducible binaries** (`linux/{amd64,arm64}`, `darwin/{amd64,arm64}`, `windows/{amd64,arm64}`) published as a single GitHub release whose assets are produced by omni itself: each archive signed with `omni sign`, an SBOM per archive via `omni sbom`, and a SLSA v1.0 provenance attestation via `omni attest` at the ADR-pinned honest level (Build L2). A CI dual-build job fails the release on any reproducibility drift. The `pkg/*` surface is frozen at phase entry and every later breaking change follows the 30-day CLAUDE.md deprecation protocol. The release notes carry an explicit "What's NOT protected against" section and an audience-scope statement.

**Architecture:** This is an **orchestration / CI plan**, not a `pkg/` library. The deliverables are configuration and glue: `.goreleaser.yaml` (extended for `-buildvcs`, six-target post-build signing, SBOM emission, and checksum), `.github/workflows/release.yml` (extended with a reproducibility dual-build gate, an attest step, and asset upload), new `Taskfile.yml` targets (`release:reproduce`, `release:sign-all`, `release:sbom-all`, `release:attest`, `freeze:check`), a small pure-Go reproducibility-diff helper command `omni reprocheck` (so the dual-build comparison is itself a dogfooded omni command and runs on every OS), a version-stamping wiring in `cmd/` so `omni --version` is build-info-derived (feeds SBOM/attest), and the release-notes generator/template. Every artifact-producing step shells out to **the omni binary built in this same run** (`omni sign`/`sbom`/`attest`), not to cosign/syft/slsa-generator — that is the dogfooding requirement. No new Go dependencies enter the default build path. GoReleaser is a build-orchestration tool invoked by CI (it is allowed to exec — it is not part of the omni binary), but the artifacts it post-processes are signed/described by omni's own pure-Go commands.

**Tech Stack:** GoReleaser v2 (already present), GitHub Actions (`goreleaser/goreleaser-action@v6`, `actions/setup-go@v5`, `actions/upload-artifact@v4`, `softprops/action-gh-release` or GoReleaser's own `release:` block), Go stdlib for `omni reprocheck` (`crypto/sha256`, `debug/buildinfo`, `os`, `io`); the Phase 04–07 omni commands `omni sign`/`verify`/`sbom`/`scan`/`attest`; `runtime/debug.ReadBuildInfo` + `-buildvcs=true` for version stamping; the Python YAML golden harness (for `reprocheck` only — release artifacts themselves are verified by real-artifact assertions in CI, not golden snapshots, because they embed per-run VCS metadata).

**Repo conventions (from research, cite when implementing):**
- Commands self-wire: `cmd/<name>.go` declares `var xCmd = &cobra.Command{...}` and calls `rootCmd.AddCommand(xCmd)` in `init()`; `RunE` reads flags → Options → calls `internal/cli/<name>.RunX(cmd.OutOrStdout(), …)`. No central registration list. (Confirmed against `cmd/sign.go`.) The ONLY new command here is `omni reprocheck` (Task 5).
- `cmderr` (`internal/cli/cmderr/cmderr.go`): `reprocheck` drift (digests differ) → `cmderr.Wrap(cmderr.ErrConflict, …)` (exit 1, fail-the-release); missing input file → `ErrNotFound` (1); unreadable → `ErrPermission` (3); bad flags / not-a-file → `ErrInvalidInput` (2); read failure → `ErrIO` (4). Use `cmderr.Wrap(sentinel, "msg")` and `errors.Is`/`As`, never `==`. `Is<Class>()` predicates exist.
- Confirmed Phase 04 CLI surface (REUSE — do not reimplement): `omni sign --key <secret.key> --sig <out.minisig> <artifact>` produces a **detached** signature (default `<artifact>.minisig`); passphrase via `OMNI_SIGN_PASSPHRASE`, NEVER a flag. `omni sign keygen --pub <out.pub> --key <out.key>`. `omni verify --key <pub> --sig <sig> <artifact>` (fail-closed → exit 1 on mismatch).
- Confirmed Phase 05 CLI surface: `omni sbom <PATH> --format spdx|cyclonedx --from auto|module|binary --source-date <RFC3339> --out <FILE> [--sign --key <secret.key>]`. Binary SBOM is auto-detected when `PATH` is a regular file. Output is byte-deterministic; pass `--source-date` for an explicit fixed timestamp.
- Confirmed Phase 07 CLI surface: `omni attest --key <secret.key> --artifact <FILE> --predicate-type slsa-provenance [--builder-id <ADR-allowed-URI>] [--from-env] --out <FILE>`; `omni attest verify --key <pub> --artifact <FILE> <ENVELOPE>`. `--from-env` populates the SLSA predicate from `GITHUB_*` and emits the **release** builder.id (L2); omitting it emits the **local** builder.id. The generator REFUSES any non-allowlisted builder.id (`ErrInvalidInput`). No numeric SLSA level field is emitted — level is implied by `builder.id` (honesty contract).
- ADRs live in `docs/adr/` as `ADR-NNNN-kebab-title.md`. **0001–0006 already exist on disk** (`ADR-0005-pure-go-minisign-signing.md`, `ADR-0006-key-handling-and-secret-redaction.md` shipped with Phase 04). Phases 05, 06, 07 each consume the next free number in sequence as they land → **0007 (sbom), 0008 (scan), 0009 (attest)**. Therefore **Phase 08's ADR is `ADR-0010`**. Header format per `docs/adr/ADR-0004-internalize-cobra-cli.md` (`# ADR-NNNN: Title`, `**Status:** Accepted`, `**Date:** 2026-06-03`, `**Decision:** …`, then `## Context`, `## Analysis` (table), `## Consequences`). (If, at execution time, 0007–0009 are NOT yet on disk because Phases 5–7 have not merged, still claim 0010 for this phase and add a one-line note in the ADR reserving 0007–0009 for sbom/scan/attest, so the numbering never reshuffles.)
- INVARIANTS: pure-Go, NO `os/exec`, no CGO **inside the omni binary** (`omni reprocheck` is pure stdlib). Cross-platform via `//go:build` tags, never runtime `os ==`. `io.Writer`/`io.Reader`; deferred `Close`. (GoReleaser and CI shell steps run OUTSIDE the binary and are exempt — they are the release machinery, not omni functionality.)
- Existing repo state (confirmed by research): `.goreleaser.yaml` is v2, already builds the 6 GOOS×GOARCH matrix with `CGO_ENABLED=0`, `-trimpath`, `ldflags: -s -w`, tar.gz archives (zip on Windows), and a conventional-commit changelog. It does NOT yet add `-buildvcs`, sign archives, emit SBOMs, or upload provenance. `.github/workflows/release.yml` triggers on `v*` tags, runs the reusable check workflow, then `goreleaser release --clean`. There is **no `rootVersion()` accessor and no `cmd.Version`** wired today — Task 4 adds it.

---

## Authoritative release-asset layout (implement exactly)

For tag `vX.Y.Z`, the GitHub release MUST contain, for EACH of the six targets `<os>_<arch>` (archive base name `omni_<TitleOs>_<archlabel>` where `archlabel` is `x86_64` for amd64 and `arm64` for arm64, matching the existing `name_template`):

```
omni_<TitleOs>_<archlabel>.tar.gz            (zip for windows)   ← GoReleaser archive
omni_<TitleOs>_<archlabel>.tar.gz.minisig    ← omni sign (detached, release key)
omni_<TitleOs>_<archlabel>.spdx.json         ← omni sbom --from binary --format spdx (over the RAW binary inside)
omni_<TitleOs>_<archlabel>.intoto.jsonl      ← omni attest --from-env (subject = the .tar.gz)
checksums.txt                                ← GoReleaser sha256 of all archives (one file, all targets)
checksums.txt.minisig                        ← omni sign over checksums.txt
omni.pub                                     ← the release PUBLIC key (so consumers can verify offline)
```

Conventions that make this deterministic and honest:
- **Signing subject = the archive bytes** (`.tar.gz`/`.zip`), via `omni sign --key <release.key> --sig <archive>.minisig <archive>`. The release **public** key (`omni.pub`) is published as an asset; the secret key never leaves CI secrets.
- **SBOM subject = the raw compiled binary** extracted from (or staged before) the archive, via `omni sbom <binary> --from binary --format spdx --source-date "$SOURCE_DATE_EPOCH_RFC3339" --out <target>.spdx.json`. (Binary SBOMs carry the Go toolchain + module set via `debug/buildinfo` — Phase 05.) `--source-date` is set to the tag's committer date in RFC-3339 (deterministic, not wall clock).
- **Attestation subject = the archive** (the thing users download), via `omni attest --key <release.key> --artifact <archive> --predicate-type slsa-provenance --from-env --out <target>.intoto.jsonl`. `--from-env` makes omni emit `BuilderIDRelease` (L2) and fill `externalParameters` from `GITHUB_WORKFLOW/REPOSITORY/REF/SHA` and metadata from `GITHUB_RUN_ID` (Phase 07).
- **Reproducibility scope:** the omni binaries are built TWICE (two independent CI runners / two clean checkouts) with identical `GOOS/GOARCH/CGO_ENABLED=0 -trimpath -buildvcs=true -ldflags="-s -w -X …"` and a pinned `SOURCE_DATE_EPOCH`. `omni reprocheck` compares the sha256 of each pair; ANY mismatch fails the job (exit 1) BEFORE the release publishes. `-buildvcs=true` embeds the (deterministic, tag-pinned) VCS stamp; the two builds share the same commit so the stamp is identical.

---

### Task 1 (ADR GATE): ADR-0010 — API freeze, reproducibility-or-fail, dogfooded supply chain, honest announce

**Files:** Create `docs/adr/ADR-0010-v1-release-policy.md`

- [ ] **Step 1: Write the ADR** matching the `ADR-0004` header/section format (`# ADR-0010: v1.0 Release Policy`, `**Status:** Accepted`, `**Date:** 2026-06-03`, `**Decision:** …`, then `## Context`, `## Analysis` (table), `## Consequences`). Record + justify these decisions (the spec already made them — this ADR pins them as the gate):
  - **`pkg/*` API freeze at phase entry.** At the start of Phase 08, the public surface of every NON-`// Experimental:` package under `pkg/` is FROZEN. From that point, every breaking change to a frozen package follows the CLAUDE.md 30-day deprecation protocol (add new alongside old; `// Deprecated: … Will be removed after YYYY-MM-DD.`; log a slog warning on the deprecated path; track in `docs/BACKLOG.md` with a `DEPRECATION` tag; remove in a separate cleanup commit after the date). Experimental packages (those carrying `// Experimental:` in `doc.go`, e.g. `pkg/sbom/{model,collect,purl}`, `pkg/scan`) are NOT frozen and may still change. List, in the ADR, the exact set of frozen packages (run `freeze:check`, Task 2, to enumerate them).
  - **Reproducibility is a release gate, not a nice-to-have (Pitfall 6).** v1.0 binaries are built with `CGO_ENABLED=0 -trimpath -buildvcs=true` and a pinned `SOURCE_DATE_EPOCH`; a CI dual-build job recompiles each target on a second clean checkout and `omni reprocheck` fails the release on ANY sha256 drift. Rationale: a non-reproducible binary makes its SBOM and signature unverifiable cross-machine — the whole chain collapses.
  - **Dogfood the supply chain (the spec's core bet).** Release assets are produced by the omni binary built in the same run — `omni sign` (archives + checksums), `omni sbom` (per-binary SPDX), `omni attest` (per-archive SLSA provenance) — NOT by cosign/syft/slsa-github-generator. Rationale: validates the primitives on the most important artifact; if omni cannot sign/describe its own release, it is not ready to claim it does this for others.
  - **Honest SLSA level = the ADR-0009 builder.id (Build L2), no overclaim (Pitfall 5/15).** The release path uses `omni attest --from-env`, which emits ONLY the allowlisted release `builder.id`; there is no flag to claim L3. The announcement and release notes state the achieved level plainly.
  - **No-overclaim announcement (Pitfall 15) + Windows parity (Pitfall 16).** Release notes MUST include a "What's NOT protected against" section (linking `.planning-archive/research/PITFALLS.md`) and an explicit audience-scope line ("Built for me + my CI/CD pipelines; broader use is welcome but not the design driver"). CI MUST run the sign/verify/sbom/attest dogfood path on Windows AND Linux AND macOS runners so cross-platform signature/SBOM determinism (CRLF, backslash-in-purl, case-folding) is proven, not assumed.
  - **No new features in this phase.** Only the `omni reprocheck` helper command (a release-tooling primitive) and configuration/CI/docs land. Any feature request is deferred to post-v1.0.
- [ ] **Step 2: Stop for human review.** Do NOT proceed to any pipeline/CI/code task until ADR-0010 is approved. This is the hard gate.

---

### Task 2: `pkg/*` API-freeze enumerator — `task freeze:check` (TDD via a script + golden list)

**Files:** Create `tools/freeze/freeze.go` (`//go:build ignore` standalone helper) and `docs/API-FREEZE.md`; modify `Taskfile.yml` (add `freeze:check`). No omni-binary code changes.

The freeze gate is a deterministic enumeration of every exported identifier in every non-`// Experimental:` `pkg/` package, written to `docs/API-FREEZE.md`. CI re-runs it and fails if the committed list and the regenerated list differ (i.e. someone added/removed/renamed a frozen export without going through the deprecation protocol).

- [ ] **Step 1: Write the failing check.** Add to `Taskfile.yml`:

```yaml
  freeze:check:
    desc: Verify the frozen pkg/* public API matches docs/API-FREEZE.md
    cmds:
      - go run tools/freeze/freeze.go > .freeze.actual
      - cmd: diff -u docs/API-FREEZE.md .freeze.actual
        platforms: [ linux, darwin ]
      - cmd: fc docs\API-FREEZE.md .freeze.actual
        platforms: [ windows ]
```

  Run `task freeze:check` now → FAIL (`tools/freeze/freeze.go` and `docs/API-FREEZE.md` do not exist).

- [ ] **Step 2: Implement `tools/freeze/freeze.go`** — a pure-stdlib `go/packages`-free walker using `go/parser` + `go/ast` over `pkg/...` (no external dep; `go/packages` would pull `golang.org/x/tools`). For each directory under `pkg/`, parse the package, skip it if any file's package doc comment contains `Experimental:`, else collect exported top-level identifiers (funcs, types, consts, vars, plus exported methods on exported types and exported struct fields). Emit a stable, sorted, line-per-symbol report:

```go
//go:build ignore

// Command freeze enumerates the frozen (non-Experimental) public API of pkg/*.
// Output is deterministic (sorted) so CI can diff it against docs/API-FREEZE.md.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	root := "pkg"
	var lines []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return err
		}
		fset := token.NewFileSet()
		pkgs, perr := parser.ParseDir(fset, path, func(fi os.FileInfo) bool {
			return strings.HasSuffix(fi.Name(), ".go") && !strings.HasSuffix(fi.Name(), "_test.go")
		}, parser.ParseComments)
		if perr != nil || len(pkgs) == 0 {
			return nil
		}
		for name, pkg := range pkgs {
			if experimental(pkg) {
				continue
			}
			rel := filepath.ToSlash(path)
			for _, sym := range exportedSymbols(pkg) {
				lines = append(lines, rel+" "+name+"."+sym)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	sort.Strings(lines)
	fmt.Println("# API Freeze — frozen pkg/* public surface (regenerate: task freeze:check)")
	for _, l := range lines {
		fmt.Println(l)
	}
}

// experimental reports whether any file's package doc marks the package Experimental.
func experimental(pkg *ast.Package) bool {
	for _, f := range pkg.Files {
		if f.Doc != nil && strings.Contains(f.Doc.Text(), "Experimental:") {
			return true
		}
	}
	return false
}

// exportedSymbols returns sorted exported top-level identifiers (+ exported
// methods and struct fields) declared in pkg.
func exportedSymbols(pkg *ast.Package) []string {
	seen := map[string]struct{}{}
	for _, f := range pkg.Files {
		for _, decl := range f.Decls {
			switch dd := decl.(type) {
			case *ast.FuncDecl:
				if !dd.Name.IsExported() {
					continue
				}
				if dd.Recv != nil { // method
					recv := recvType(dd.Recv)
					if ast.IsExported(recv) {
						seen[recv+"."+dd.Name.Name+"()"] = struct{}{}
					}
					continue
				}
				seen[dd.Name.Name+"()"] = struct{}{}
			case *ast.GenDecl:
				for _, spec := range dd.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if s.Name.IsExported() {
							seen[s.Name.Name] = struct{}{}
							if st, ok := s.Type.(*ast.StructType); ok {
								for _, fld := range st.Fields.List {
									for _, n := range fld.Names {
										if n.IsExported() {
											seen[s.Name.Name+"#"+n.Name] = struct{}{}
										}
									}
								}
							}
						}
					case *ast.ValueSpec:
						for _, n := range s.Names {
							if n.IsExported() {
								seen[n.Name] = struct{}{}
							}
						}
					}
				}
			}
		}
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func recvType(fl *ast.FieldList) string {
	if len(fl.List) == 0 {
		return ""
	}
	switch t := fl.List[0].Type.(type) {
	case *ast.StarExpr:
		if id, ok := t.X.(*ast.Ident); ok {
			return id.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return ""
}
```

- [ ] **Step 3: Generate the committed baseline:** `go run tools/freeze/freeze.go > docs/API-FREEZE.md`. Inspect it — confirm it lists `pkg/sign`, `pkg/secret`, `pkg/sbom/format`, `pkg/attest` (the stable surfaces) and does NOT list `pkg/sbom/model`, `pkg/sbom/collect`, `pkg/sbom/purl`, `pkg/scan` (Experimental). Paste this frozen-package set into ADR-0010's freeze list (Task 1).
- [ ] **Step 4: Run, verify pass:** `task freeze:check` → exit 0 (no diff). Then add a sacrificial exported `func ZZZTemp()` to any frozen package, re-run → FAIL (diff non-empty), and remove it → PASS. This proves the gate trips.
- [ ] **Step 5: Commit:** `gofmt -w tools/freeze/ && git commit -- tools/freeze docs/API-FREEZE.md Taskfile.yml -m "build(release): pkg/* API-freeze enumerator + freeze:check gate"`

---

### Task 3: `omni reprocheck` library — deterministic digest-pair comparison (TDD)

**Files:** Create `internal/cli/reprocheck/reprocheck.go`, `internal/cli/reprocheck/reprocheck_test.go`. (Pure stdlib; no `pkg/` library needed — this is release tooling glue, like other `internal/cli` commands. No `doc.go` requirement for `internal/`, but add a package comment.)

`reprocheck` takes two parallel lists of files (build A and build B), computes sha256 of each pair, and returns `cmderr.ErrConflict` if any pair differs. It is the dogfooded reproducibility gate, runnable on every OS.

- [ ] **Step 1: Write failing tests** (`internal/cli/reprocheck/reprocheck_test.go`):

```go
package reprocheck_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/internal/cli/reprocheck"
)

func write(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestReproIdenticalPasses(t *testing.T) {
	d := t.TempDir()
	a := write(t, d, "a-omni", []byte("BINARYBYTES"))
	b := write(t, d, "b-omni", []byte("BINARYBYTES"))
	var w bytes.Buffer
	if err := reprocheck.Run(&w, reprocheck.Options{A: []string{a}, B: []string{b}}); err != nil {
		t.Fatalf("Run(identical) = %v, want nil", err)
	}
	if !strings.Contains(w.String(), "reproducible") {
		t.Errorf("output missing OK line: %q", w.String())
	}
}

func TestReproDriftFailsConflict(t *testing.T) {
	d := t.TempDir()
	a := write(t, d, "a-omni", []byte("BINARYBYTES"))
	b := write(t, d, "b-omni", []byte("DIFFERENT!!"))
	err := reprocheck.Run(&bytes.Buffer{}, reprocheck.Options{A: []string{a}, B: []string{b}})
	if !cmderr.IsConflict(err) {
		t.Fatalf("Run(drift) = %v, want cmderr.ErrConflict", err)
	}
}

func TestReproMismatchedListLen(t *testing.T) {
	err := reprocheck.Run(&bytes.Buffer{}, reprocheck.Options{A: []string{"x"}, B: nil})
	if !cmderr.IsInvalidInput(err) {
		t.Fatalf("Run(len mismatch) = %v, want ErrInvalidInput", err)
	}
}

func TestReproMissingFile(t *testing.T) {
	d := t.TempDir()
	a := write(t, d, "a", []byte("x"))
	err := reprocheck.Run(&bytes.Buffer{}, reprocheck.Options{A: []string{a}, B: []string{filepath.Join(d, "nope")}})
	if !cmderr.IsNotFound(err) {
		t.Fatalf("Run(missing) = %v, want ErrNotFound", err)
	}
}
```

- [ ] **Step 2: Run, verify fail:** `go test ./internal/cli/reprocheck/ -v` → FAIL (package/`Run` undefined).
- [ ] **Step 3: Implement** `internal/cli/reprocheck/reprocheck.go`:

```go
// Package reprocheck compares two parallel sets of built artifacts and fails
// closed if any sha256 digest pair differs — the dogfooded reproducible-build
// gate for the v1.0 release. Pure stdlib; runs on every OS.
package reprocheck

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// Options configures a reproducibility check. A and B are equal-length, index-
// aligned lists of file paths from two independent builds.
type Options struct {
	A []string
	B []string
}

// Run compares each (A[i], B[i]) pair by sha256. It writes a per-pair report to
// w and returns cmderr.ErrConflict on the first drift, after reporting all pairs.
func Run(w io.Writer, opts Options) error {
	if len(opts.A) != len(opts.B) || len(opts.A) == 0 {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "reprocheck requires equal, non-empty A/B file lists")
	}
	drift := false
	for i := range opts.A {
		ha, err := digest(opts.A[i])
		if err != nil {
			return err
		}
		hb, err := digest(opts.B[i])
		if err != nil {
			return err
		}
		if ha == hb {
			_, _ = fmt.Fprintf(w, "reproducible  %s  %s\n", ha[:16], opts.A[i])
		} else {
			drift = true
			_, _ = fmt.Fprintf(w, "DRIFT         %s != %s  %s\n", ha[:16], hb[:16], opts.A[i])
		}
	}
	if drift {
		return cmderr.Wrap(cmderr.ErrConflict, "reproducibility drift detected (binaries differ between builds)")
	}
	return nil
}

func digest(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", cmderr.Wrap(cmderr.ErrNotFound, "open "+path)
		}
		if os.IsPermission(err) {
			return "", cmderr.Wrap(cmderr.ErrPermission, "open "+path)
		}
		return "", cmderr.Wrap(cmderr.ErrIO, "open "+path)
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", cmderr.Wrap(cmderr.ErrIO, "read "+path)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
```

- [ ] **Step 4: Run, verify pass:** `go test ./internal/cli/reprocheck/ -v` → PASS (all four cases).
- [ ] **Step 5: Commit:** `gofmt -w internal/cli/reprocheck/ && git commit -- internal/cli/reprocheck -m "feat(reprocheck): pure-Go reproducible-build digest-pair gate (fail-closed on drift)"`

---

### Task 4: `omni reprocheck` Cobra wrapper + build-info version stamping (TDD + smoke)

**Files:** Create `cmd/reprocheck.go`; modify `cmd/root.go` (wire `rootCmd.Version` from build info); create `cmd/version_info.go` (a tiny `Version()` accessor reused by sbom/attest/reprocheck).

omni currently has no `cmd.Version` / `rootVersion()` (confirmed). Phase 05's `cmd/sbom.go` already calls `rootVersion()`; this task creates that accessor so the SBOM/attest tool-version fields are populated honestly from the embedded `-buildvcs`/`-ldflags` data.

- [ ] **Step 1: Write failing test** (`cmd/version_info_test.go`):

```go
package cmd

import "testing"

func TestRootVersionNonEmpty(t *testing.T) {
	if rootVersion() == "" {
		t.Fatal("rootVersion() must never be empty (fallback to (devel) or build info)")
	}
}
```

- [ ] **Step 2: Run, verify fail:** `go test ./cmd/ -run TestRootVersion -v` → FAIL (`rootVersion` undefined).
- [ ] **Step 3: Implement** `cmd/version_info.go`:

```go
package cmd

import "runtime/debug"

// version is overridable at build time via -ldflags "-X github.com/inovacc/omni/cmd.version=vX.Y.Z".
// When unset, rootVersion falls back to the VCS stamp embedded by -buildvcs=true.
var version = ""

// rootVersion returns the omni version string for --version, SBOM tool fields,
// and attestation builder.version. It prefers the ldflags value, then the
// buildvcs main-module version, then "(devel)".
func rootVersion() string {
	if version != "" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return "(devel)"
}
```

  In `cmd/root.go` `init()` (or wherever `rootCmd` is declared), set `rootCmd.Version = rootVersion()` so `omni --version` works.

- [ ] **Step 4: Implement** `cmd/reprocheck.go` (thin wrapper; flags `--a` and `--b` are repeatable string slices):

```go
package cmd

import (
	"github.com/inovacc/omni/internal/cli/reprocheck"
	"github.com/spf13/cobra"
)

var reprocheckCmd = &cobra.Command{
	Use:   "reprocheck --a FILE [--a FILE...] --b FILE [--b FILE...]",
	Short: "Fail if any A/B build artifact pair differs (reproducible-build gate)",
	Long: `Compare two index-aligned sets of built artifacts by sha256 and exit
non-zero (cmderr.ErrConflict) on any drift. Used by the v1.0 release dual-build
job to dogfood reproducibility across all six targets.

Example:
  omni reprocheck --a buildA/omni-linux-amd64 --b buildB/omni-linux-amd64`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		a, _ := cmd.Flags().GetStringArray("a")
		b, _ := cmd.Flags().GetStringArray("b")
		return reprocheck.Run(cmd.OutOrStdout(), reprocheck.Options{A: a, B: b})
	},
}

func init() {
	reprocheckCmd.Flags().StringArray("a", nil, "artifact path from build A (repeatable)")
	reprocheckCmd.Flags().StringArray("b", nil, "artifact path from build B (repeatable)")
	rootCmd.AddCommand(reprocheckCmd)
}
```

- [ ] **Step 5: Run, verify pass + smoke:** `go test ./cmd/ -run "TestRootVersion" -v` → PASS; `go build ./... && go run . --version` prints a version; build two identical binaries and `go run . reprocheck --a A/omni --b B/omni` exits 0; corrupt one and confirm exit 1.
- [ ] **Step 6: Commit:** `gofmt -w cmd/ && git commit -- cmd/reprocheck.go cmd/version_info.go cmd/root.go -m "feat(reprocheck): omni reprocheck CLI + build-info version stamping (rootVersion)"`

---

### Task 5: GoReleaser pipeline — `-buildvcs`, deterministic ldflags, checksums, signing & SBOM hooks

**Files:** Modify `.goreleaser.yaml`.

Extend the existing v2 config (do NOT rewrite the working matrix) to: stamp version via ldflags, add `-buildvcs=true`, pin `mod_timestamp` for reproducibility, emit a `checksums.txt`, and sign every archive + the checksum file with `omni sign`, and emit a per-binary SBOM via `omni sbom`. GoReleaser's `signs:` and `sboms:` blocks let us point at an arbitrary command — we point them at the omni binary built earlier in the same run (path `./dist/omni_<...>/omni` is unstable across targets, so we stage a single host-built omni at `./omni-host` in a `before` hook and call that).

- [ ] **Step 1: Stage a host omni + extend builds.** Add to the `before.hooks` and `builds` sections:

```yaml
before:
  hooks:
    - go mod tidy
    # Build a host-arch omni used to sign/sbom/attest the release artifacts (dogfooding).
    - go build -trimpath -buildvcs=true -ldflags "-s -w -X github.com/inovacc/omni/cmd.version={{ .Version }}" -o omni-host .

builds:
  - id: omni
    binary: omni
    env:
      - CGO_ENABLED=0
    mod_timestamp: '{{ .CommitTimestamp }}'   # reproducibility: pin embedded mod time
    flags:
      - -trimpath
      - -buildvcs=true
    ldflags:
      - -s -w -X github.com/inovacc/omni/cmd.version={{ .Version }}
    goos: [linux, windows, darwin]
    goarch: [amd64, arm64]
```

  (`.CommitTimestamp` and `.Version` are GoReleaser template fields; `mod_timestamp` removes the per-run file-mtime nondeterminism.)

- [ ] **Step 2: Add a deterministic checksums block:**

```yaml
checksum:
  name_template: 'checksums.txt'
  algorithm: sha256
```

- [ ] **Step 3: Add the signing block (dogfood `omni sign`)** — sign every archive AND the checksums file with the release key; publish the public key:

```yaml
signs:
  - id: omni-sign
    cmd: ./omni-host
    args: ["sign", "--key", "{{ .Env.OMNI_RELEASE_KEY_PATH }}", "--sig", "${signature}", "${artifact}"]
    signature: "${artifact}.minisig"
    artifacts: all          # archives + checksums.txt
    output: true
    # OMNI_SIGN_PASSPHRASE is exported in the workflow from a GitHub secret.
```

  And add the public key as an extra release file (Step 5).

- [ ] **Step 4: Add the SBOM block (dogfood `omni sbom`)** — one SPDX SBOM per built binary, over the RAW binary (not the archive), with a pinned source date:

```yaml
sboms:
  - id: omni-sbom
    cmd: ./omni-host
    args:
      - "sbom"
      - "$artifact"
      - "--from"
      - "binary"
      - "--format"
      - "spdx"
      - "--source-date"
      - "{{ .CommitDate }}"   # RFC-3339 committer date — deterministic
      - "--out"
      - "$document"
    documents:
      - "${artifact}.spdx.json"
    artifacts: binary
```

- [ ] **Step 5: Publish the release public key + ensure provenance assets upload.** Add an `extra_files` to the `release` block (create the block if absent) so consumers can verify offline; the `.intoto.jsonl` files are added in Task 6 (uploaded via the workflow):

```yaml
release:
  extra_files:
    - glob: ./omni.pub        # release PUBLIC key, written by the workflow before release
    - glob: ./dist/*.intoto.jsonl
```

- [ ] **Step 6: Validate the config offline:** `goreleaser check` (validates `.goreleaser.yaml` schema; does not build). Then a local dry run that exercises build+sbom+sign WITHOUT publishing: generate a throwaway local key (`go run . sign keygen --pub omni.pub --key /tmp/rel.key` with `OMNI_SIGN_PASSPHRASE=test`), export `OMNI_RELEASE_KEY_PATH=/tmp/rel.key OMNI_SIGN_PASSPHRASE=test`, and run `goreleaser release --snapshot --clean --skip=publish`. Confirm `dist/` contains, per target, the `.tar.gz`/`.zip`, a `.minisig`, and a `.spdx.json`, plus `checksums.txt` + `checksums.txt.minisig`. Verify one: `go run . verify --key omni.pub --sig dist/omni_Linux_x86_64.tar.gz.minisig dist/omni_Linux_x86_64.tar.gz` → exit 0.
- [ ] **Step 7: Commit:** `git commit -- .goreleaser.yaml -m "build(release): buildvcs+pinned-timestamp reproducible builds; dogfood omni sign/sbom; checksums + pubkey assets"`

---

### Task 6: Release workflow — dual-build reproducibility gate + cross-OS dogfood + attest

**Files:** Modify `.github/workflows/release.yml`.

Add three things to the tag-triggered release: (a) a `reproduce` matrix job that builds all six targets on TWO independent runners and uploads them, then a `repro-gate` job that downloads both and runs `omni reprocheck` (fail-the-release on drift); (b) a cross-OS `dogfood` matrix (ubuntu + windows + macos) that builds omni and round-trips sign/verify + sbom + attest/attest-verify so Windows/macОS parity is proven (Pitfall 16); (c) an `attest` step in the release job that, after GoReleaser produces archives, runs `omni attest --from-env` per archive and lets GoReleaser upload the `.intoto.jsonl` (already globbed in Task 5). The release secret key is materialized from `secrets.OMNI_RELEASE_KEY` / `secrets.OMNI_RELEASE_PASSPHRASE`.

- [ ] **Step 1: Add the reproducibility dual-build jobs** before `release` (and make `release` need them):

```yaml
  reproduce:
    needs: test
    strategy:
      matrix:
        build: [A, B]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }
      - uses: actions/setup-go@v5
        with: { go-version-file: 'go.mod' }
      - name: Build all six targets (deterministic)
        run: |
          export SOURCE_DATE_EPOCH="$(git log -1 --format=%ct)"
          mkdir -p out
          for os in linux darwin windows; do
            for arch in amd64 arm64; do
              ext=""; [ "$os" = "windows" ] && ext=".exe"
              GOOS=$os GOARCH=$arch CGO_ENABLED=0 \
                go build -trimpath -buildvcs=true \
                -ldflags "-s -w -X github.com/inovacc/omni/cmd.version=${GITHUB_REF_NAME}" \
                -o "out/omni-${os}-${arch}${ext}" .
            done
          done
      - uses: actions/upload-artifact@v4
        with:
          name: build-${{ matrix.build }}
          path: out/

  repro-gate:
    needs: reproduce
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version-file: 'go.mod' }
      - uses: actions/download-artifact@v4
        with: { name: build-A, path: A }
      - uses: actions/download-artifact@v4
        with: { name: build-B, path: B }
      - name: reprocheck (fail release on any drift)
        run: |
          A=$(ls A | sed 's#^#--a A/#'); B=$(ls B | sed 's#^#--b B/#')
          go run . reprocheck $A $B
```

- [ ] **Step 2: Add the cross-OS dogfood job** (proves Pitfall 16 — sign/verify/sbom/attest work identically on Windows + macOS):

```yaml
  dogfood:
    needs: test
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
    runs-on: ${{ matrix.os }}
    env:
      OMNI_SIGN_PASSPHRASE: dogfood-pass
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version-file: 'go.mod' }
      - name: round-trip sign/verify + sbom + attest (cross-platform)
        shell: bash
        run: |
          go build -o omni-host .
          ./omni-host sign keygen --pub k.pub --key k.key
          echo "artifact" > art.bin
          ./omni-host sign --key k.key --sig art.bin.minisig art.bin
          ./omni-host verify --key k.pub --sig art.bin.minisig art.bin
          ./omni-host sbom ./omni-host --from binary --format spdx --source-date 1970-01-01T00:00:00Z --out omni.spdx.json
          ./omni-host attest --key k.key --artifact art.bin --predicate-type slsa-provenance --out art.intoto.jsonl
          ./omni-host attest verify --key k.pub --artifact art.bin art.intoto.jsonl
```

- [ ] **Step 3: Gate `release` on the new jobs + materialize the release key + attest archives.** Change `release.needs` to `[test, repro-gate, dogfood]` and add key-materialization + post-GoReleaser attest steps:

```yaml
  release:
    needs: [test, repro-gate, dogfood]
    runs-on: ubuntu-latest
    permissions:
      contents: write
      id-token: write          # OIDC identity feeds attest --from-env externalParameters
    env:
      OMNI_SIGN_PASSPHRASE: ${{ secrets.OMNI_RELEASE_PASSPHRASE }}
      OMNI_RELEASE_KEY_PATH: ${{ github.workspace }}/release.key
    steps:
      - uses: actions/checkout@v4
        with: { fetch-depth: 0 }
      - uses: actions/setup-go@v5
        with: { go-version-file: 'go.mod' }
      - name: Materialize release signing key
        run: |
          printf '%s' "${{ secrets.OMNI_RELEASE_KEY }}" > release.key
          printf '%s' "${{ secrets.OMNI_RELEASE_PUB }}" > omni.pub
          chmod 600 release.key
      - name: GoReleaser (build + sign + sbom)
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: '~> v2'
          args: release --clean --skip=publish     # defer publish until attest is added
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      - name: Attest each archive (dogfood omni attest --from-env, L2)
        run: |
          go build -o omni-host .
          for a in dist/*.tar.gz dist/*.zip; do
            [ -e "$a" ] || continue
            ./omni-host attest --key release.key --artifact "$a" \
              --predicate-type slsa-provenance --from-env --out "${a}.intoto.jsonl"
          done
      - name: Publish release with all assets
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

  > NOTE for the implementer: GoReleaser cannot be invoked twice with `--clean` (the second wipes `dist/`). Resolve at execution time by EITHER (a) using a single GoReleaser run with a `signs`/`sboms`/post-hook that also calls `omni attest` (preferred — add an `after.hooks` entry calling `./omni-host attest … --out "{{ .ArtifactPath }}.intoto.jsonl"` per archive), OR (b) running GoReleaser once with `--skip=publish`, attesting, then publishing via `softprops/action-gh-release@v2` uploading `dist/*`. Pick (a) if GoReleaser's `after.hooks` exposes per-archive paths; else (b). Whichever is chosen, the asset set MUST match the "release-asset layout" section. Document the choice in the workflow comment.

- [ ] **Step 4: Validate the workflow** with `actionlint` (if available) or `go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/release.yml` is NOT allowed (no new dep / no exec-in-binary concern — this is CI-only tooling, acceptable). Minimum: YAML-lint the file and confirm job dependency graph (`test → reproduce → repro-gate`, `test → dogfood`, `release needs [test, repro-gate, dogfood]`). Push to a throwaway branch with a pre-release tag (`v0.0.0-rc1`) on a fork or with a manual `workflow_dispatch` guard to smoke the jobs WITHOUT publishing to the real repo; confirm `repro-gate` and `dogfood` pass and `release --skip=publish` produces the full asset set in the job log.
- [ ] **Step 5: Commit:** `git commit -- .github/workflows/release.yml -m "ci(release): dual-build reproducibility gate, cross-OS sign/sbom/attest dogfood, SLSA L2 provenance upload"`

---

### Task 7: Taskfile release targets — local dogfood of the full asset pipeline

**Files:** Modify `Taskfile.yml` (add `release:reproduce`, `release:sign-all`, `release:sbom-all`, `release:attest`, `release:verify-all`, and an umbrella `release:dryrun`).

These let a maintainer reproduce the CI pipeline locally on any OS (Pitfall 16 parity) and are the manual smoke gate for Task 8. They reuse the omni binary built by `task build`.

- [ ] **Step 1: Add the targets** (cross-platform via `platforms:`; the bash bodies run on linux/darwin, with Windows variants where the loop differs — keep Windows variants minimal by delegating to a `go run tools/...` helper if a shell loop is awkward):

```yaml
  release:reproduce:
    desc: Build all six targets twice and reprocheck (local reproducibility gate)
    cmds:
      - cmd: |
          export SOURCE_DATE_EPOCH="$(git log -1 --format=%ct)"
          for pass in A B; do
            mkdir -p "repro/$pass"
            for os in linux darwin windows; do for arch in amd64 arm64; do
              ext=""; [ "$os" = windows ] && ext=".exe"
              GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -trimpath -buildvcs=true \
                -ldflags "-s -w" -o "repro/$pass/omni-$os-$arch$ext" .
            done; done
          done
          A=$(ls repro/A | sed 's#^#--a repro/A/#'); B=$(ls repro/B | sed 's#^#--b repro/B/#')
          go run . reprocheck $A $B
        platforms: [ linux, darwin ]

  release:sbom-all:
    desc: Emit an SPDX SBOM per built binary (deterministic source-date)
    deps: [ build:all ]
    cmds:
      - cmd: |
          d="$(git log -1 --format=%cI)"
          for b in {{.BUILD_DIR}}/omni-*; do
            go run . sbom "$b" --from binary --format spdx --source-date "$d" --out "$b.spdx.json"
          done
        platforms: [ linux, darwin ]

  release:sign-all:
    desc: Sign every built binary with $OMNI_RELEASE_KEY_PATH (OMNI_SIGN_PASSPHRASE in env)
    deps: [ build:all ]
    cmds:
      - cmd: |
          for b in {{.BUILD_DIR}}/omni-*; do
            case "$b" in *.spdx.json|*.minisig|*.intoto.jsonl) continue;; esac
            go run . sign --key "$OMNI_RELEASE_KEY_PATH" --sig "$b.minisig" "$b"
          done
        platforms: [ linux, darwin ]

  release:attest:
    desc: Generate a local (L1) SLSA attestation per built binary
    deps: [ build:all ]
    cmds:
      - cmd: |
          for b in {{.BUILD_DIR}}/omni-*; do
            case "$b" in *.spdx.json|*.minisig|*.intoto.jsonl) continue;; esac
            go run . attest --key "$OMNI_RELEASE_KEY_PATH" --artifact "$b" \
              --predicate-type slsa-provenance --out "$b.intoto.jsonl"
          done
        platforms: [ linux, darwin ]

  release:verify-all:
    desc: Verify every signature + attestation against the public key
    cmds:
      - cmd: |
          for b in {{.BUILD_DIR}}/omni-*; do
            case "$b" in *.spdx.json|*.minisig|*.intoto.jsonl) continue;; esac
            go run . verify --key "$OMNI_RELEASE_PUB_PATH" --sig "$b.minisig" "$b"
            go run . attest verify --key "$OMNI_RELEASE_PUB_PATH" --artifact "$b" "$b.intoto.jsonl"
          done
        platforms: [ linux, darwin ]

  release:dryrun:
    desc: Full local dogfood — reproduce, sbom, sign, attest, verify, goreleaser snapshot
    cmds:
      - task: release:reproduce
      - task: release:sbom-all
      - task: release:sign-all
      - task: release:attest
      - task: release:verify-all
      - cmd: goreleaser release --snapshot --clean --skip=publish
        platforms: [ linux, darwin ]
```

- [ ] **Step 2: Smoke it locally.** Generate a throwaway key: `OMNI_SIGN_PASSPHRASE=test go run . sign keygen --pub /tmp/omni.pub --key /tmp/omni.key`. Export `OMNI_RELEASE_KEY_PATH=/tmp/omni.key OMNI_RELEASE_PUB_PATH=/tmp/omni.pub OMNI_SIGN_PASSPHRASE=test`. Run `task release:reproduce` (expect "reproducible" lines, exit 0), then `task release:sbom-all release:sign-all release:attest release:verify-all` → all exit 0; confirm a `.spdx.json`, `.minisig`, `.intoto.jsonl` exist per binary.
- [ ] **Step 3: Commit:** `git commit -- Taskfile.yml -m "build(release): local reproduce/sbom/sign/attest/verify dogfood targets"`

---

### Task 8: Release notes + announcement — "What's NOT protected against" + audience scope (Pitfall 15)

**Files:** Create `docs/RELEASE-NOTES-v1.0.md` (the human-authored body) and `.goreleaser.yaml` `release.header`/`release.footer` (point GoReleaser at the honesty sections); create `docs/superpowers/specs/` status update.

GoReleaser generates the changelog from conventional commits (already configured). We supply a fixed header (audience scope) and footer ("What's NOT protected against") so every release carries the honesty statement. No overclaim of general-purpose fitness or SLSA L3.

- [ ] **Step 1: Write `docs/RELEASE-NOTES-v1.0.md`** with these REQUIRED sections (verbatim intent, fill specifics):
  - **Audience & scope:** "omni is built for me and my CI/CD pipelines. Broader open-source adoption is welcome but is not the design driver. Design decisions optimize for deterministic, dependency-light CI use, not for being a general-purpose security suite."
  - **What v1.0 protects:** signed binaries (minisign-compatible Ed25519 via `omni sign`), per-binary SPDX SBOMs (`omni sbom`), SLSA **Build L2** provenance (`omni attest`, builder.id = the release workflow), reproducible builds verified by a dual-build CI gate, OSV vulnerability scanning against a signed DB (`omni scan`).
  - **What's NOT protected against** (link `.planning-archive/research/PITFALLS.md`): NOT SLSA L3 (no hermetic/isolated build service, no non-falsifiable provenance — Pitfall 5); NO Rekor transparency-log upload, NO Fulcio/OIDC certificate issuance, NO OCI registry push (sigstore is verification-only behind `-tags omni_sigstore` — Pitfall 9); reachability scanning (`omni scan source`) is opt-in behind `-tags omni_govulncheck` and absent from release binaries; the OSV DB is only as fresh as its last `omni scan db update` (Pitfall 11); SBOM transitive-completeness is bounded by `debug/buildinfo` (Pitfall 12). State plainly: "This is not a turnkey enterprise supply-chain platform."
  - **Verify-it-yourself** block: the exact commands a consumer runs — `omni verify --key omni.pub --sig <archive>.minisig <archive>`, `omni attest verify --key omni.pub --artifact <archive> <archive>.intoto.jsonl`, and how to inspect the `.spdx.json`.
- [ ] **Step 2: Wire GoReleaser to emit the honesty sections** — add to `.goreleaser.yaml`:

```yaml
release:
  header: |
    omni {{ .Tag }} — built for me + my CI/CD pipelines (see "Audience & scope" below).
  footer: |
    ## What's NOT protected against
    SLSA Build L2 only (NOT L3). No Rekor / Fulcio / OCI. Reachability scan and
    sigstore bundle verify are opt-in build tags, absent from these binaries.
    Full scope and pitfalls: docs/RELEASE-NOTES-v1.0.md and PITFALLS.md.
    ## Verify these artifacts
    omni verify --key omni.pub --sig omni_<OS>_<arch>.tar.gz.minisig omni_<OS>_<arch>.tar.gz
    omni attest verify --key omni.pub --artifact omni_<OS>_<arch>.tar.gz omni_<OS>_<arch>.tar.gz.intoto.jsonl
```

- [ ] **Step 3: Verify** the notes render: `goreleaser release --snapshot --clean --skip=publish` and inspect `dist/CHANGELOG.md` / the rendered release body in the GoReleaser output — confirm the header (audience) and footer (NOT-protected + verify) appear. Confirm no string in the notes claims "L3", "full supply chain", or "Rekor/transparency log" as a delivered feature (grep the rendered notes for those terms → none in a positive-claim context).
- [ ] **Step 4: Commit:** `git commit -- docs/RELEASE-NOTES-v1.0.md .goreleaser.yaml -m "docs(release): v1.0 release notes — audience scope + What's NOT protected (no overclaim)"`

---

### Task 9: Golden-master test for `omni reprocheck` (deterministic, fixtures)

**Files:** Modify `testing/golden/golden_tests.yaml` AND `tools/golden/golden_tests.yaml` (keep in sync); add fixtures under `testing/golden/fixtures/reprocheck/` (two identical files `same-a`/`same-b` and a differing pair `diff-a`/`diff-b`).

Release artifacts themselves embed per-run VCS metadata and CANNOT be golden-pinned (Pitfall 14) — so the ONLY golden coverage here is `omni reprocheck`, whose inputs are committed fixtures with fixed bytes.

- [ ] **Step 1: Create fixtures** under `testing/golden/fixtures/reprocheck/`: `same-a` and `same-b` (identical bytes, e.g. `repro-fixture-bytes-v1\n`), `diff-a` (`AAAA\n`), `diff-b` (`BBBB\n`). These are tiny committed files (NOT real binaries) — `reprocheck` only hashes bytes, so the content is arbitrary as long as it is fixed.
- [ ] **Step 2: Add a `reprocheck` category to BOTH yaml files** (mirror the `sign` category shape):
  - `reprocheck_match` — `args: ["reprocheck", "--a", "{fixtures}/same-a", "--b", "{fixtures}/same-b"]` → exit 0; `normalizations: ["strip_path"]` (the `reproducible <hash16> <path>` line contains a path; the hash16 is deterministic for fixed bytes so it is NOT stripped — assert it matches).
  - `reprocheck_drift` — `args: ["reprocheck", "--a", "{fixtures}/diff-a", "--b", "{fixtures}/diff-b"]` → `exit_code: 1` `# cmderr.ErrConflict`; `normalizations: ["strip_path"]`.
  - `reprocheck_bad_args` — `args: ["reprocheck", "--a", "{fixtures}/same-a"]` (no `--b`) → `exit_code: 2` `# cmderr.ErrInvalidInput`.
  - `reprocheck_missing` — `args: ["reprocheck", "--a", "{fixtures}/same-a", "--b", "{fixtures}/does-not-exist"]` → `exit_code: 1` `# cmderr.ErrNotFound`; `normalizations: ["strip_path"]`.
- [ ] **Step 3: Generate + verify snapshots:** `task test:golden:update && task golden:record`, then `python testing/scripts/test_golden.py` → all green.
- [ ] **Step 4: Commit:** `git commit -- testing/golden tools/golden -m "test(reprocheck): golden-master match + drift/bad-args/missing negatives"`

---

### Task 10: Docs + final gate

**Files:** `docs/COMMANDS.md`, `CLAUDE.md` (command inventory line), `docs/ROADMAP.md` (mark Phase 08), `docs/superpowers/specs/2026-05-16-08-v1-release-cut-design.md` (status), `docs/architecture/patterns.md` (note the release-tooling layer if appropriate). The `docs/EXTERNAL_SOURCES.md` needs a GoReleaser attribution row only if not already present.

- [ ] **Step 1: Docs** — add `reprocheck` (with `--a`/`--b`) to `docs/COMMANDS.md` and bump the CLAUDE.md inventory count; document the v1.0 release process (the `task release:dryrun` local dogfood + the CI dual-build gate + the asset layout) in a short "Release" section of `docs/ARCHITECTURE.md` or a new `docs/RELEASE.md` referenced from CLAUDE.md's topical index. Reference ADR-0010. Run `omni aicontext` / `omni cmdtree` regen if applicable. Mark the spec `Status: Complete` and the ROADMAP Phase 08 row done.
- [ ] **Step 2: Final gate (run all; every line must pass):**

```bash
# binary + cross-compile sanity (all six targets compile with the release flags)
for os in linux darwin windows; do for arch in amd64 arm64; do \
  GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -trimpath -buildvcs=true -ldflags "-s -w" -o /dev/null . ; \
done; done
go vet ./... && gofmt -l cmd/reprocheck.go cmd/version_info.go internal/cli/reprocheck tools/freeze
golangci-lint run --timeout=5m ./...
go test ./internal/cli/reprocheck/... ./cmd/... -count=1
task freeze:check
python testing/scripts/test_golden.py
goreleaser check                         # .goreleaser.yaml schema valid
# full local dogfood (throwaway key in env):
OMNI_SIGN_PASSPHRASE=test go run . sign keygen --pub /tmp/omni.pub --key /tmp/omni.key
OMNI_RELEASE_KEY_PATH=/tmp/omni.key OMNI_RELEASE_PUB_PATH=/tmp/omni.pub OMNI_SIGN_PASSPHRASE=test task release:dryrun
```

  Expected: all six targets compile; `freeze:check` exits 0 (frozen surface unchanged); golden suite green; `goreleaser check` valid; `release:dryrun` produces signed + SBOM'd + attested artifacts and `release:verify-all` (inside dryrun) passes every signature and attestation. Confirm the rendered release notes contain the audience-scope and "What's NOT protected against" sections.
- [ ] **Step 3: Commit:** `git commit -- docs/ CLAUDE.md -m "docs(release): document omni reprocheck + v1.0 release process; mark Phase 08 complete"`
- [ ] **Step 4 (the actual cut — human-gated):** Once the above is green AND ADR-0010 is approved AND the `OMNI_RELEASE_KEY`/`OMNI_RELEASE_PUB`/`OMNI_RELEASE_PASSPHRASE` GitHub secrets are set, tag `vX.Y.Z` and push. The workflow runs `test → reproduce → repro-gate`, `test → dogfood`, then `release`. Verify on the published release: six archives, six `.minisig`, six `.spdx.json`, six `.intoto.jsonl`, `checksums.txt`(+`.minisig`), `omni.pub`. Download one archive and run the consumer verify commands from the release footer to confirm an outside party can verify it.

---

## Self-Review

**Spec coverage** (REL success criteria → tasks):

| Success criterion (spec) | Task(s) |
|---|---|
| 1. `pkg/` API freeze tagged; 30-day deprecation protocol applies to future breaking changes | **T1** (ADR-0010 records the freeze + protocol) + **T2** (`freeze:check` enumerates + CI-enforces the frozen surface) |
| 2. Signed binaries published for all 6 targets as a GitHub release; each signature by `omni sign` | **T5** (`.goreleaser.yaml signs:` → `omni sign`, all targets + checksums) + **T6** (workflow materializes the release key, publishes) + **T7** (local `release:sign-all`) |
| 3. Reproducible builds (`-trimpath` + `-buildvcs=true`); CI dual-build job fails on drift | **T3** (`reprocheck` lib, fail-closed) + **T4** (`omni reprocheck` CLI + version stamping) + **T5** (`mod_timestamp`/`-buildvcs` in goreleaser) + **T6** (`reproduce`+`repro-gate` jobs) + **T7** (`release:reproduce`) + **T9** (golden) |
| 4. Release assets include SBOM from `omni sbom` + SLSA v1.0 provenance from `omni attest` at the honest level | **T5** (`sboms:` → `omni sbom --from binary`) + **T6** (`attest --from-env` → L2 builder.id per archive) + **T1** (ADR pins honest L2, no overclaim flag) |
| 5. Release notes include "What's NOT protected against"; announcement states audience scope, no overclaim | **T8** (`docs/RELEASE-NOTES-v1.0.md` + goreleaser `header`/`footer`) + **T6 dogfood job** + **T1** (ADR no-overclaim + Windows-parity decisions) |
| Pitfall 6 (reproducibility), 15 (overclaim), 16 (Windows gaps) | **T3/T4/T6** (6), **T8/T1** (15), **T6 dogfood matrix across ubuntu/windows/macos + T7** (16) |

**Placeholder scan:** No "TBD"/"add validation"/"handle edge cases"/"similar to Task N". Every command interface used is the EXACT confirmed surface from the Phase 04–07 plans (flags quoted verbatim: `omni sign --key --sig`, `omni sbom <path> --from binary --format spdx --source-date --out`, `omni attest --key --artifact --predicate-type slsa-provenance --from-env --out`, `omni attest verify --key --artifact <env>`). The only NEW code (`internal/cli/reprocheck`, `cmd/reprocheck.go`, `cmd/version_info.go`, `tools/freeze/freeze.go`) is given in full, compilable Go with tests written first. The GoReleaser/workflow/Taskfile blocks are concrete YAML, not prose. The single deliberately-deferred decision — GoReleaser single-run-with-after-hook vs. run-twice-then-publish (T6 Step 3) — is bounded to two named, fully-specified alternatives with a selection rule and a documented constraint (`--clean` wipes `dist/`), not an open "figure it out".

**Type consistency:** `reprocheck.Options{A, B []string}` (T3) is consumed by `cmd/reprocheck.go` via `--a`/`--b` `StringArray` flags (T4) and by the `repro-gate` job + `release:reproduce` shell loops (T6/T7) which pass index-aligned `--a`/`--b` pairs. `rootVersion()` (T4) is the accessor Phase 05's `cmd/sbom.go` already calls and is reused for `-ldflags -X …cmd.version` (T5) and `rootCmd.Version` (T4). `cmderr.{ErrConflict(1), ErrInvalidInput(2), ErrPermission(3), ErrNotFound(1), ErrIO(4), Wrap, IsConflict, IsInvalidInput, IsNotFound}` (confirmed in `cmderr.go`) are used per the convention table and asserted in T3 tests + T9 golden exit codes. The release-asset names in the layout section match the existing `.goreleaser.yaml` `name_template` (`omni_<TitleOs>_x86_64`/`arm64`), so the workflow/footer paths line up. `tools/freeze/freeze.go`'s `Experimental:` detection matches the `// Experimental:` doc.go convention Phases 05/06 use (so `pkg/sbom/{model,collect,purl}` and `pkg/scan` are correctly excluded; `pkg/sign`, `pkg/secret`, `pkg/sbom/format`, `pkg/attest` correctly frozen).

**Known risks:**
1. **GoReleaser cannot run twice with `--clean`** (the second wipes `dist/`). Mitigated in T6 Step 3 with two fully-specified alternatives and a hard NOTE; the preferred path is a single GoReleaser run plus an `after.hooks`/`signs`/`sboms` chain that also attests, keeping all asset production in one invocation.
2. **`omni sbom` subject choice (binary vs archive).** SBOM is over the RAW binary (richest `debug/buildinfo` data); signature + attestation are over the ARCHIVE (the thing users download). This split is intentional and stated in the asset-layout section; if a consumer expects the SBOM to cover the archive, document that the SBOM describes the binary inside it.
3. **`mod_timestamp` / `SOURCE_DATE_EPOCH` parity between GoReleaser and the dual-build job.** GoReleaser uses `mod_timestamp: {{ .CommitTimestamp }}`; the `reproduce` job exports `SOURCE_DATE_EPOCH=$(git log -1 --format=%ct)`. Both derive from the same commit, so they agree — but the dual-build job intentionally compares its OWN two builds to each other (not against GoReleaser's), which is the cleaner reproducibility claim. If a future goal is "GoReleaser output reproduces the raw `go build` output," that requires matching archive packaging too and is explicitly OUT of scope for v1.0 (binaries-reproduce-binaries is the gate).
4. **Release secret key management.** The release key lives only in GitHub secrets (`OMNI_RELEASE_KEY`), materialized to a `0600` file at job time and never logged (passphrase via `OMNI_SIGN_PASSPHRASE` from a secret, per Phase 04 policy — never a flag). Risk: a leaked secret. Mitigation is operational (rotate key, ADR-0006 dev-vs-release key separation already mandates a distinct release key ID).
5. **ADR numbering depends on Phases 5–7 landing first.** This plan claims `ADR-0010` and reserves 0007–0009 for sbom/scan/attest. If those phases land out of order or skip a number, adjust the ADR filename at execution time (T1) — but 0010 is safe because it is past the highest reserved number.
6. **Windows shell loops in Taskfile.** The `release:*` loop bodies are `platforms: [linux, darwin]`-gated; Windows maintainers use the CI `dogfood` job (which uses `shell: bash` on `windows-latest`) or WSL. The cross-platform PROOF (Pitfall 16) comes from the CI `dogfood` matrix, not the local Taskfile, so Windows parity is genuinely tested.
