# Phase 03 — CONCERNS Burn-down (post-hardening remainder) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the remaining Phase 03 concerns-burndown deliverables that the 2026-06-03 security/robustness hardening sweep did NOT already cover — i.e. the non-security polish needed before the v1.0 `pkg/*` API freeze.

**Architecture:** Small, mostly-independent tasks. Two are code+TDD (`cmderr.Is<Class>()` helpers; `id` Windows library helpers), the rest are docs/godoc bookkeeping (EXIT-CODES.md, ISSUES triage, `// Experimental:` API markers, stale-TODO reconciliation). No behavior changes to existing commands; the six Windows-parity commands are already correct.

**Tech Stack:** Go, `errors`, `internal/cli/cmderr`, build tags (`//go:build`), Markdown docs.

**Scope note (why this is small):** The hardening sweep (branch `harden/audit-fixes`, see `docs/quality/HARDENING.md`) closed every Critical/High security concern (archive path-traversal/symlink/hardlink, panic/DoS guards) and injection-hardened `forloop`/`task`. The Windows-parity audit found all six target commands already have real `*_windows.go` implementations. The archived `.planning-archive/codebase/CONCERNS.md` "76 commands without cmderr" item is stale (Phase 1 reached 100% adoption). `crypto-02` (PBKDF2 envelope) and the darwin `ioreg` machine-id item are intentionally DEFERRED and tracked in `docs/BACKLOG.md` — NOT in this plan.

**Repo hygiene:** Work on a branch off `harden/audit-fixes` (or `main` after that merges). Use explicit pathspecs in every commit (`git commit -- <files>`), never `git add -A` (the tree carries unrelated untracked junk: `c/`, `bin/`, coverage logs). No AI attribution. Conventional commits.

---

### Task 1: `cmderr.Is<Class>()` convenience helpers (POLISH — deferred from Phase 1)

**Files:**
- Modify: `internal/cli/cmderr/cmderr.go`
- Test: `internal/cli/cmderr/cmderr_test.go`

- [ ] **Step 1: Confirm the sentinel set and that `errors` is imported**

Read `internal/cli/cmderr/cmderr.go`. Confirm the 7 sentinels exist: `ErrNotFound`, `ErrInvalidInput`, `ErrPermission`, `ErrIO`, `ErrConflict`, `ErrTimeout`, `ErrUnsupported`, and that the file imports `"errors"` (it uses `errors.Is`/`As` already — if not, add it in Step 3).

- [ ] **Step 2: Write the failing test**

Append to `internal/cli/cmderr/cmderr_test.go`:

```go
func TestIsClassHelpers(t *testing.T) {
	cases := []struct {
		name     string
		fn       func(error) bool
		sentinel error
	}{
		{"IsNotFound", IsNotFound, ErrNotFound},
		{"IsInvalidInput", IsInvalidInput, ErrInvalidInput},
		{"IsPermission", IsPermission, ErrPermission},
		{"IsIO", IsIO, ErrIO},
		{"IsConflict", IsConflict, ErrConflict},
		{"IsTimeout", IsTimeout, ErrTimeout},
		{"IsUnsupported", IsUnsupported, ErrUnsupported},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.fn(tc.sentinel) {
				t.Errorf("%s(sentinel) = false, want true", tc.name)
			}
			if !tc.fn(Wrap(tc.sentinel, "context")) {
				t.Errorf("%s(wrapped) = false, want true", tc.name)
			}
			if tc.fn(errors.New("unrelated")) {
				t.Errorf("%s(unrelated) = true, want false", tc.name)
			}
			if tc.fn(nil) {
				t.Errorf("%s(nil) = true, want false", tc.name)
			}
		})
	}
}
```

(If `cmderr_test.go` does not already import `"errors"`, add it.)

- [ ] **Step 3: Run the test, verify it fails**

Run: `go test ./internal/cli/cmderr/ -run TestIsClassHelpers -v`
Expected: FAIL — `undefined: IsNotFound` (etc.).

- [ ] **Step 4: Implement the helpers**

Append to `internal/cli/cmderr/cmderr.go` (ensure `"errors"` is imported):

```go
// IsNotFound reports whether err is, or wraps, ErrNotFound.
func IsNotFound(err error) bool { return errors.Is(err, ErrNotFound) }

// IsInvalidInput reports whether err is, or wraps, ErrInvalidInput.
func IsInvalidInput(err error) bool { return errors.Is(err, ErrInvalidInput) }

// IsPermission reports whether err is, or wraps, ErrPermission.
func IsPermission(err error) bool { return errors.Is(err, ErrPermission) }

// IsIO reports whether err is, or wraps, ErrIO.
func IsIO(err error) bool { return errors.Is(err, ErrIO) }

// IsConflict reports whether err is, or wraps, ErrConflict.
func IsConflict(err error) bool { return errors.Is(err, ErrConflict) }

// IsTimeout reports whether err is, or wraps, ErrTimeout.
func IsTimeout(err error) bool { return errors.Is(err, ErrTimeout) }

// IsUnsupported reports whether err is, or wraps, ErrUnsupported.
func IsUnsupported(err error) bool { return errors.Is(err, ErrUnsupported) }
```

- [ ] **Step 5: Run the test, verify it passes**

Run: `go test ./internal/cli/cmderr/ -run TestIsClassHelpers -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
gofmt -w internal/cli/cmderr/cmderr.go internal/cli/cmderr/cmderr_test.go
git commit -- internal/cli/cmderr/cmderr.go internal/cli/cmderr/cmderr_test.go -m "feat(cmderr): add Is<Class>() convenience helpers"
```

---

### Task 2: `docs/EXIT-CODES.md` reference (POLISH — deferred from Phase 1)

**Files:**
- Create: `docs/EXIT-CODES.md`

- [ ] **Step 1: Confirm the authoritative mapping**

Read `internal/cli/cmderr/cmderr.go` `ExitCodeFor()`. Confirm the mapping (source of truth):
`nil`→0; `SilentError`→its Code; `ExitError`→its Code; `ErrNotFound`→1; `ErrConflict`→1; `ErrInvalidInput`→2; `ErrPermission`→3; `ErrIO`→4; `ErrTimeout`→5; `ErrUnsupported`→6; any other error→1.

- [ ] **Step 2: Create the doc**

Create `docs/EXIT-CODES.md`:

```markdown
# omni Exit Codes

omni classifies every error through `internal/cli/cmderr` sentinels; `cmderr.ExitCodeFor(err)` (called in `cmd/root.go`) maps them to process exit codes. This page is the human-readable reference for that mapping.

| Exit code | Meaning | Sentinel(s) | `Is*` helper |
|-----------|---------|-------------|--------------|
| 0 | Success | — (`nil`) | — |
| 1 | Not found / conflict / unclassified | `ErrNotFound`, `ErrConflict`, any unclassified error | `IsNotFound`, `IsConflict` |
| 2 | Invalid input / usage | `ErrInvalidInput` | `IsInvalidInput` |
| 3 | Permission denied | `ErrPermission` | `IsPermission` |
| 4 | I/O error | `ErrIO` | `IsIO` |
| 5 | Timeout | `ErrTimeout` | `IsTimeout` |
| 6 | Unsupported operation | `ErrUnsupported` | `IsUnsupported` |

Notes:
- A `SilentError`/`ExitError` carries its own explicit code (see `cmderr.WithExitCode`).
- A recovered panic exits with the dedicated panic code set in `cmd/root.go` (`panicExitCode`).
- Any error not matching a sentinel falls through to exit code **1**.

Source of truth: `internal/cli/cmderr/cmderr.go` (`ExitCodeFor`). Keep this table in sync when sentinels change.
```

- [ ] **Step 3: Verify it matches the code**

Run: `go doc ./internal/cli/cmderr` (or re-read `ExitCodeFor`) and visually confirm every sentinel/code pair in the table matches. Run `go build ./...` (no code changed; sanity only).
Expected: table matches `ExitCodeFor`.

- [ ] **Step 4: Commit**

```bash
git commit -- docs/EXIT-CODES.md -m "docs: add EXIT-CODES.md reference for cmderr exit-code mapping"
```

---

### Task 3: `id` Windows library helpers — fix or `ErrUnsupported` (POLISH — Windows parity)

**Context:** The `id` *command* (`RunID`) works on Windows (prints `os/user` string IDs). The latent gap is the library helpers `GetUID()/GetGID()/GetGroups()` in `internal/cli/id/id.go`, which `strconv.Atoi` the value `os/user` returns — on Windows that is a SID string (e.g. `S-1-5-21-...`), so `Atoi` errors. Per the spec, return `cmderr.ErrUnsupported` with a clear message on Windows rather than a confusing parse error.

**Files:**
- Read: `internal/cli/id/id.go` (record exact signatures of `GetUID`/`GetGID`/`GetGroups`)
- Create: `internal/cli/id/id_unix.go`, `internal/cli/id/id_windows.go`
- Modify: `internal/cli/id/id.go` (move the numeric helpers out)
- Test: `internal/cli/id/id_test.go`

- [ ] **Step 1: Read and record the exact signatures**

Read `internal/cli/id/id.go`. Record the EXACT signatures and bodies of `GetUID`, `GetGID`, `GetGroups` (return types, e.g. `func GetUID() (int, error)` / `func GetGroups() ([]int, error)`). Note which file-level imports they use (`os/user`, `strconv`). Do not edit yet.

- [ ] **Step 2: Write the failing test (platform-aware)**

Add to `internal/cli/id/id_test.go` (adapt the helper names/return types to the actual signatures recorded in Step 1):

```go
func TestNumericHelpers_PlatformContract(t *testing.T) {
	_, err := GetUID()
	if runtime.GOOS == "windows" {
		if !cmderr.IsUnsupported(err) {
			t.Errorf("GetUID on windows: want ErrUnsupported, got %v", err)
		}
	} else {
		if err != nil {
			t.Errorf("GetUID on %s: unexpected error %v", runtime.GOOS, err)
		}
	}
}
```

(Add imports `"runtime"` and `"github.com/inovacc/omni/internal/cli/cmderr"` if not present. Mirror the assertion for `GetGID`/`GetGroups` if you want fuller coverage — same shape.)

- [ ] **Step 3: Run, verify it fails on Windows**

Run: `go test ./internal/cli/id/ -run TestNumericHelpers_PlatformContract -v`
Expected: FAIL on Windows — `GetUID` currently returns a `strconv.Atoi` parse error, not `ErrUnsupported`.

- [ ] **Step 4: Split the helpers by platform**

Move the existing numeric bodies (recorded in Step 1) from `id.go` into a new `internal/cli/id/id_unix.go`:

```go
//go:build !windows

package id

// <paste the existing GetUID/GetGID/GetGroups bodies verbatim from id.go here,
// keeping their exact signatures and the strconv/os/user logic>
```

Create `internal/cli/id/id_windows.go`:

```go
//go:build windows

package id

import "github.com/inovacc/omni/internal/cli/cmderr"

// On Windows, os/user returns SID strings (e.g. "S-1-5-21-..."), which have no
// numeric uid/gid equivalent. The `id` command itself prints the string IDs and
// works; these numeric library helpers are unsupported here.

// GetUID is unsupported on Windows (SIDs are not numeric).
func GetUID() (int, error) {
	return 0, cmderr.Wrap(cmderr.ErrUnsupported, "id: numeric UID unavailable on Windows (SID-based)")
}

// GetGID is unsupported on Windows (SIDs are not numeric).
func GetGID() (int, error) {
	return 0, cmderr.Wrap(cmderr.ErrUnsupported, "id: numeric GID unavailable on Windows (SID-based)")
}

// GetGroups is unsupported on Windows (SIDs are not numeric).
func GetGroups() ([]int, error) {
	return nil, cmderr.Wrap(cmderr.ErrUnsupported, "id: numeric group IDs unavailable on Windows (SID-based)")
}
```

(MATCH the exact signatures from Step 1. If a helper returns different types, adapt the Windows stub's zero-value return accordingly. Remove the moved functions from `id.go`; if that leaves `strconv` unused in `id.go`, drop the import there — `gofmt`/`go build` will tell you.)

- [ ] **Step 5: Run the test, verify it passes; full package**

Run: `go test ./internal/cli/id/ -count=1 -v`
Expected: PASS (the platform contract test + all existing `id` tests).

- [ ] **Step 6: Cross-compile both platforms**

Run: `go build ./internal/cli/id/...` then `CGO_ENABLED=0 GOOS=linux go build ./internal/cli/id/...` and `CGO_ENABLED=0 GOOS=windows go build ./internal/cli/id/...`
Expected: all succeed.

- [ ] **Step 7: Commit**

```bash
gofmt -w internal/cli/id/
git commit -- internal/cli/id/id.go internal/cli/id/id_unix.go internal/cli/id/id_windows.go internal/cli/id/id_test.go -m "fix(id): return ErrUnsupported for numeric UID/GID helpers on Windows"
```

---

### Task 4: `pkg/*` API triage — `// Experimental:` markers + `pkg/private` relocation (POLISH-07/16/17)

**Context:** Before the v1.0 API freeze, every `pkg/*` package must be bucketed. No `// Experimental:` or `// Deprecated:` markers exist today. Buckets (from the scoping audit):
- **Stable (freeze, no marker needed):** `hashutil`, `cryptutil`, `encoding`, `idgen`, `jsonutil`, `textutil`, `userdirs`, `cobra/helper/output`.
- **Experimental (add marker):** `procutil`, `gopsagent`, `procmetrics`, `obfuscate`, and the whole `pkg/video/` tree (11 subpackages).
- **Mixed (leave for now, no marker):** `figlet`, `sqlfmt`, `cssfmt`, `htmlfmt`, `search`, `pipeline`, `twig`.
- **Not API:** `pkg/private/` — zero Go files (vendored buf/googleapis testdata cache).

**Files:**
- Create/modify `doc.go` in each experimental package
- Modify: `.gitignore`

- [ ] **Step 1: Add the Experimental marker to each experimental package**

For each of these packages, ensure a `doc.go` exists carrying a package doc comment with an `// Experimental:` line. If a `doc.go` already exists, prepend the marker to the existing package comment; otherwise create it.

Packages: `pkg/procutil`, `pkg/gopsagent`, `pkg/procmetrics`, `pkg/obfuscate`, and each `pkg/video/*` subpackage that has Go files (`pkg/video`, `pkg/video/downloader`, `pkg/video/extractor` (+ its subpkgs), `pkg/video/format`, `pkg/video/utils`, `pkg/video/nethttp`, `pkg/video/m3u8`, `pkg/video/jsinterp`, `pkg/video/cache`, `pkg/video/types`).

Template (set `<pkgname>` to the actual package clause and tailor the one-line reason):

```go
// Package <pkgname> ...<keep or write a one-line summary>...
//
// Experimental: this package's API may change before a stable release and is
// not covered by the v1.0 compatibility guarantee.
package <pkgname>
```

Special reason for the `pkg/video/*` tree: append "It tracks third-party site internals (YouTube/innertube/HLS) and will change as those change." For `pkg/video/types` and `pkg/video/m3u8`, keep the marker but note in the comment they are candidates for promotion to stable later.

Use the smallest correct change: if a package already has its doc comment on a non-`doc.go` file, you may add `doc.go` solely for the marker (Go merges package comments — but only ONE file should carry the package doc comment; if one already exists, edit THAT file instead of creating a duplicate, to avoid a "duplicate package comment" vet warning).

- [ ] **Step 2: Verify markers compile and are present**

Run:
```bash
go build ./... && go vet ./pkg/...
grep -rl "Experimental:" pkg/procutil pkg/gopsagent pkg/procmetrics pkg/obfuscate pkg/video
```
Expected: build+vet clean; grep lists a file for every experimental package (no "duplicate package comment" vet error).

- [ ] **Step 3: Handle `pkg/private` (not a package)**

`pkg/private/` contains zero Go files (vendored buf/googleapis testdata). It must not look like a public package. Add it to `.gitignore` so it is never committed under `pkg/`:

```gitignore
# vendored buf/googleapis testdata cache — not a public package
/pkg/private/
```

(If `pkg/private/` is already tracked in git — check `git ls-files pkg/private | head` — instead `git rm -r --cached pkg/private` and relocate it to `testdata/` or `third_party/`; record which you did. Per the scoping audit it is currently untracked, so the `.gitignore` line is sufficient.)

- [ ] **Step 4: Commit**

```bash
gofmt -w pkg/
git commit -- .gitignore $(git diff --name-only -- 'pkg/**/doc.go') -m "docs(pkg): mark experimental pkg/* APIs; exclude pkg/private vendored testdata"
```

(If `doc.go` files were newly created they are untracked — add them explicitly by path instead of relying on `git diff`.)

---

### Task 5: Triage `docs/ISSUES.md` to empty (POLISH — spec criterion #3)

**Context:** `docs/ISSUES.md` has 12 entries; NONE are true defects. `docs/BUGS.md` is an empty template. Reach an ISSUES.md with no open defects by relocating each entry.

**Files:**
- Modify: `docs/ISSUES.md`, `docs/BACKLOG.md`

- [ ] **Step 1: Move genuine backlog items to `docs/BACKLOG.md`**

Append these to the appropriate `docs/BACKLOG.md` priority section (they are feature/coverage/CI gaps, not bugs):
- `sed`/`awk` do not implement the full GNU/AWK feature set (feature-completeness).
- Video download limits (YouTube sig fragility, no SAMPLE-AES HLS, no FFmpeg merge — note: FFmpeg merge conflicts with the no-exec rule).
- `rg` fidelity gaps (heuristic binary detection, `.gitignore` nested-negation edge cases).
- `pkg/video/extractor/generic` has no unit tests (test-coverage).
- `pkg/video/extractor/youtube` minimal tests (~4%) (test-coverage).
- No coverage-threshold enforcement in CI (target 80%) (CI-hardening).

- [ ] **Step 2: Convert design-tradeoff notes into a non-tracker section**

The platform/design notes are intentional, permanent tradeoffs, not issues. Move them out of the "open issues" list into a clearly-labelled `## Platform & design notes (not defects)` section at the BOTTOM of `docs/ISSUES.md` (or into `docs/ARCHITECTURE.md` if a "Known limitations" section fits better):
- Windows platform limitations (chmod ACLs, chown unsupported, `ln -s` elevation, kill signal subset, free/df differing APIs).
- macOS `free` uses sysctl instead of `/proc`.
- SQLite pure-Go slower than CGO; BBolt single-writer.
- `go test ./...` slow due to buf package compilation (has `-short`/package-scoped workaround).

- [ ] **Step 3: Delete the stale/resolved entries**

- Remove the struck-through "cmderr adoption (76 commands w/o classification)" line and its large "Recently Resolved" changelog block — already RESOLVED Apr 2026 (100% adoption). If history is wanted, it already lives in the Phase 1 archive under `.planning-archive/`.

- [ ] **Step 4: Verify-then-close the release-pipeline note**

- "No automated release pipeline (manual go build)" is stale: confirm `.goreleaser.yaml` and `.github/workflows/release.yml` exist (`ls .goreleaser.yaml .github/workflows/release.yml`). If present, delete the entry (resolved). If somehow absent, move it to `docs/BACKLOG.md` instead.

- [ ] **Step 5: Verify ISSUES.md has no open defects**

Read `docs/ISSUES.md`. Confirm the "open issues" list is empty (only the template/header and the "Platform & design notes" section remain).

- [ ] **Step 6: Commit**

```bash
git commit -- docs/ISSUES.md docs/BACKLOG.md -m "docs: triage ISSUES.md to empty — relocate backlog/design-notes, drop resolved"
```

---

### Task 6: Reconcile stale `TODO:no-exec-violation` comments with the accepted decision (hygiene)

**Context:** `internal/cli/terraform/terraform.go:5-6` (and the `git/hacks` / `gh/hacks` files) still carry `// TODO:no-exec-violation ... Tracked for Plan 17` comments plus `//nolint:gosec`. The no-exec boundary decision (2026-06-03) ACCEPTED these as sanctioned external-tool wrappers (`docs/architecture/patterns.md` § "No-exec invariant: scope & sanctioned exceptions"). The stale TODOs now contradict the documented decision.

**Files:**
- Modify: `internal/cli/terraform/terraform.go`; the `git/hacks` and `gh/hacks` source files (locate via grep).

- [ ] **Step 1: Find the stale comments**

Run: `grep -rn "TODO:no-exec-violation\|no-exec-violation\|no-exec violation" internal/`
Record each file:line.

- [ ] **Step 2: Replace each stale TODO with a sanctioned-exception note**

For each hit, replace the `TODO:no-exec-violation ...` text with a comment of the form (keep any `//nolint:gosec` directive if the linter still requires it):

```go
// Sanctioned exec exception: this command's purpose is to orchestrate an
// external tool. Permitted under the no-exec invariant — see
// docs/architecture/patterns.md § "No-exec invariant: scope & sanctioned exceptions".
```

Do NOT change any executable code — comments only. (If a `//nolint:gosec` is no longer needed because the call is provably argv-only, you may drop it, but only if `golangci-lint` still passes.)

- [ ] **Step 3: Verify no stale TODOs remain; build/lint**

Run:
```bash
grep -rn "TODO:no-exec-violation" internal/ || echo "none remaining"
go build ./... && go vet ./... && golangci-lint run --timeout=5m ./...
```
Expected: no stale TODOs; build/vet/lint clean.

- [ ] **Step 4: Commit**

```bash
git commit -- internal/cli/terraform/terraform.go <git-hacks-and-gh-hacks-files> -m "chore: reconcile no-exec TODOs with accepted sanctioned-wrapper decision"
```

---

### Task 7: Final gate

**Files:** none (verification).

- [ ] **Step 1: Full build/vet/cross-compile/lint**

Run:
```bash
go build ./... && go vet ./...
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ./...
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=5m ./...
```
Expected: all exit 0.

- [ ] **Step 2: Targeted tests for changed packages**

Run: `go test ./internal/cli/cmderr/... ./internal/cli/id/... -count=1`
Expected: PASS. (The `internal/cli/exec` package has a known pre-existing test-hermeticity failure — see `docs/BACKLOG.md`; it is unrelated to this plan and must NOT be "fixed" by weakening assertions here.)

- [ ] **Step 3: Update phase status**

In `docs/superpowers/specs/2026-05-16-03-concerns-burndown-design.md`, change `**Status:** Planned` to `**Status:** Complete (2026-06-XX)` with a one-line pointer to this plan and `docs/quality/HARDENING.md` (which covered the security half).

- [ ] **Step 4: Commit**

```bash
git commit -- docs/superpowers/specs/2026-05-16-03-concerns-burndown-design.md -m "docs(phase-03): mark concerns-burndown complete"
```

---

## Self-Review

**Spec coverage** (concerns-burndown spec workstreams → tasks):
- "Resolve/defer every Critical and High concern" → **closed by the hardening sweep** (see Scope note + `docs/quality/HARDENING.md`); `crypto-02` + darwin `ioreg` explicitly DEFERRED in `docs/BACKLOG.md`. No task needed here.
- "Windows parity for ps/df/free/kill/uptime/id (fix or ErrUnsupported)" → **Task 3** (only the `id` numeric helpers remained; the six commands already pass). Verified by the parity audit.
- "Triage every exported pkg/* symbol; Experimental godoc; audit pkg/video" → **Task 4**.
- "Triage docs/ISSUES.md to empty" → **Task 5**.
- "Generate docs/EXIT-CODES.md" → **Task 2**.
- "Add cmderr.Is<Class>() helpers" → **Task 1**.
- "Adopt breaking-change protocol for pkg/* changes" → covered by the Experimental markers (Task 4) + existing CLAUDE.md protocol; no code task.
- Stale `TODO:no-exec-violation` hygiene (surfaced during scoping) → **Task 6**.

**Placeholder scan:** Tasks 1, 2, 6 contain complete code/content. Tasks 3 and 4 contain explicit "read the exact signatures / package clause first, then apply this exact stub" instructions with full stub code — bounded read-then-fill, not vague placeholders (mirrors the project's existing scaffold-service plan convention). No "TBD/handle edge cases/add validation" left.

**Type consistency:** `Is<Class>` names in Task 1 match the 7 sentinels and are reused in Task 2's EXIT-CODES table and Task 3's test (`cmderr.IsUnsupported`). `cmderr.Wrap(cmderr.ErrUnsupported, ...)` in Task 3 matches the existing `Wrap` signature confirmed in Task 1/Step 1. Experimental-marker packages in Task 4 match the scoping buckets exactly.

**Known risk:** Task 3 depends on the exact `GetUID/GetGID/GetGroups` signatures, which Step 1 mandates reading before writing the split (the stubs adapt to the real return types). Task 4's "one package-comment file only" guard prevents a duplicate-package-comment vet error.
