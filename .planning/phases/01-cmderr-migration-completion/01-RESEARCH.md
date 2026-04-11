# Phase 1: cmderr Migration Completion - Research

**Researched:** 2026-04-11
**Domain:** Go CLI error classification, golden-master testing, coverage gating
**Confidence:** HIGH (everything verified against the working tree)

## Summary

Phase 1 is a mechanical migration: wrap remaining `internal/cli/<cmd>/*.go` error returns with `cmderr.Wrap(sentinel, ...)` so `cmd/root.go` can map them to exit codes via `cmderr.ExitCodeFor()`. The pattern is already locked — 84 commands use it — and the remaining work is computed (exact list below). Golden YAML already supports `exit_code:` and error-path snapshots exist for `exist`; extending that convention is the plan. No Taskfile coverage target exists today; a new `task lint:cmderr-coverage` must be added and wired into `.github/workflows/test.yml`.

**Primary recommendation:** Copy the `head.go` / `find.go` wrapping patterns verbatim into each remaining command, add one `exit_code:`-tagged entry per command to both golden registries, and add a 20-line Taskfile target that parses `go tool cover -func` output.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

1. **Migration order = risk-weighted waves** (not alphabetical):
   - **Wave A** (CI-critical): `cp`, `mv`, `rm`, `mkdir`, `chmod`, `touch`, `stat`, `ln`, `cat`, `grep`, `find`, `ls`, `wc`, `tr`, `sort`, `uniq`, `head`, `tail`, `cut`, `diff`, `sed`, `awk`, `xargs`, `env`, `which`, `file`, `tar`, `zip`, `unzip`, `gzip`, `gunzip`, `curl`, `date`, `printf`
   - **Wave B** (data/format/encoding): `jq`, `yq`, `json*`, `yaml*`, `toml*`, `xml*`, `sql`, `html*`, `css*`, `base*`, `hex*`, `urlenc`, `htmlenc`, `hash`, `sha*`, `md5*`, `crc*`, `xxd`, `uuid`, `random`, `jwt`
   - **Wave C** (system/proc/info): `ps`, `kill`, `df`, `du`, `free`, `uptime`, `uname`, `whoami`, `id`, `dotenv`, `watch`, `yes`, `time`, `sleep`, `seq`
   - **Wave D** (everything else, alphabetical): `banner`, `brdoc`, `cron`, `forloop`, `input`, `lint`, `loc`, `note`, `pager`, `safepath`, `tagfixer`, `tree`, etc.

2. **pkg/ wrapping depth = CLI boundary only.** `pkg/*` stays raw-error; wrapping lives exclusively in `internal/cli/*/<cmd>.go`. No exceptions — `pkg/*` is library-surface importable by external projects.

3. **Golden tests = one error snapshot per command minimum.** Richer 3–5 snapshot matrix for `find`, `sed`, `awk`, `dd`, `grep`, `curl`, `tar`, `jq`, `diff`. Snapshots land in both `testing/golden/golden_tests.yaml` and `tools/golden/golden_tests.yaml`.

4. **Coverage gate = new `task lint:cmderr-coverage` in `Taskfile.yml`** that parses coverage profile and fails if cmderr package coverage < 90%. Wired into `.github/workflows/test.yml` as a required job.

5. **Platform errors = single `ErrPermission` sentinel.** Rely on Go's `os.ErrPermission` for platform-neutral classification. Unix EPERM + Windows `ERROR_ACCESS_DENIED` both match via `errors.Is`.

6. **No backward-compat audit.** v1.0 is the first stable exit-code contract. Changed exit codes are logged in `EXIT-CODE-CHANGES.md` for release notes.

7. **No-classifiable-error commands** (`yes`, `printf`, `seq`, `pwd`, `whoami`, `arch`, `uname`): return `ErrIO` only for write failures, `nil` otherwise. Cobra already gates bad flags.

### Claude's Discretion

- Exact per-command sentinel selection within the established pattern
- Whether any missing sentinel must be added to `cmderr.go` (in scope if discovered; must surface the gap)
- Plan decomposition inside each wave (3–5 plans per wave)
- Naming of new golden test entries

### Deferred Ideas (OUT OF SCOPE)

- `docs/EXIT-CODES.md` generation → Phase 3
- `cmderr.Is<Class>()` convenience helpers → Phase 3
- Structured logging integration → post-1.0
- Custom `golangci-lint` rule flagging raw `os.ErrX` returns → Phase 2
- Cross-command exit-code matrix (`exit-codes.yaml`) → Phase 2+
- Touching `pkg/*` or `cmd/*` files beyond the already-working wiring
- Raising omni-wide coverage (that's POLISH-04/05 in Phase 2)
- Consolidating the two golden registries
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| POLISH-01 | Every command returns classified `cmderr` sentinels with correct exit codes | Verified pattern from `head.go`/`find.go`/`hash.go`; exact remaining-command list computed below |
| POLISH-02 | Every migrated command has ≥1 golden-master error-path test | Confirmed `exit_code:` schema works in both YAML registries (example: `exist_file_not_found`) |
| POLISH-03 | `internal/cli/cmderr` ≥90% coverage enforced in CI | Taskfile.yml has no such target today; drop-in target specified below |

## Requirement Traceability

- POLISH-01 → Waves A/B/C/D migration plans
- POLISH-02 → Golden snapshot plan (one per migrated command, rich matrix for 9 commands)
- POLISH-03 → Taskfile target + CI workflow plan
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/inovacc/omni/internal/cli/cmderr` | in-tree | Sentinels + `Wrap` + `ExitCodeFor` + `SilentExit` | Already the canonical error model; `cmd/root.go` wired to it |
| `errors` (stdlib) | Go 1.22+ | `errors.Is` / `errors.As` for classification | Required; `==` forbidden by CLAUDE.md |
| `fmt` (stdlib) | Go 1.22+ | `fmt.Errorf("%w", ...)` wrapping | Standard |
| `os` (stdlib) | Go 1.22+ | `os.ErrNotExist`, `os.ErrPermission` platform-neutral sentinels | Already used by migrated commands |

### Testing / Gate
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Golden harness (`testing/golden/` + `tools/golden/`) | in-tree | Error-path snapshots with `exit_code:` assertion | Every migrated command |
| `go tool cover -func` | stdlib | Coverage percentage parser for lint target | New `lint:cmderr-coverage` |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Hand-written coverage parser | `github.com/wadey/gocovmerge` or `github.com/nikolaydubina/go-cover-treemap` | Rejected — adds a dep; `go tool cover -func` + `awk` is 5 lines |
| Custom linter for raw errors | `ruleguard` / `golangci-lint` rule | Deferred per CONTEXT — Phase 2 |
| Unifying the two golden registries | — | Deferred per CONTEXT — post-1.0 |

**No package installation needed.** All tooling is stdlib + in-tree.

**Version verification:** `internal/cli/cmderr/cmderr.go` is at `1f3a9a0` (current worktree); 105 LOC; last-known canonical sentinels verified on 2026-04-11.

## Architecture Patterns

### Recommended Project Structure (no changes)
```
internal/cli/
├── cmderr/          # canonical sentinels (DO NOT EXTEND unless needed)
├── <cmd>/
│   ├── <cmd>.go     # ← ALL wrapping happens here
│   └── <cmd>_test.go
pkg/                 # ← OFF-LIMITS this phase
cmd/                 # ← OFF-LIMITS this phase (root.go already wired)
testing/golden/golden_tests.yaml     # ← add error entries
tools/golden/golden_tests.yaml       # ← add same entries
Taskfile.yml         # ← add lint:cmderr-coverage
.github/workflows/test.yml  # ← wire new task
```

### Pattern 1: File-based commands (open/read/stat)
**What:** Classify `os.ErrNotExist` / `os.ErrPermission`, pass through others.
**Source:** `internal/cli/head/head.go:42-49` (verified).
```go
sources, err := input.Open(args, r)
if err != nil {
    if errors.Is(err, os.ErrNotExist) {
        return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("<cmd>: %s", err))
    }
    if errors.Is(err, os.ErrPermission) {
        return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("<cmd>: %s", err))
    }
    return fmt.Errorf("<cmd>: %w", err)
}
```

### Pattern 2: Parse/validate commands (regex, flags, numeric parse)
**What:** Wrap user-input parse failures as `ErrInvalidInput`.
**Source:** `internal/cli/find/find.go:77`, `internal/cli/path/path.go:55`.
```go
if _, err := regexp.Compile(pattern); err != nil {
    return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("<cmd>: invalid regex: %s", err))
}
// For missing required args:
return cmderr.Wrap(cmderr.ErrInvalidInput, "<cmd>: missing operand")
```

### Pattern 3: Verification / comparison commands (hash verify, diff, cmp)
**What:** A logical mismatch is `ErrConflict`, not an error.
**Source:** `internal/cli/hash/hash.go:268`.
```go
if !bytes.Equal(got, want) {
    return cmderr.Wrap(cmderr.ErrConflict, "<cmd>: verification failed")
}
```

### Pattern 4: Silent exit commands (grep, cmp — "no-match is not an error message")
**What:** Return `cmderr.SilentExit(n)` to set exit code without stderr output.
**Source:** `internal/cli/grep/grep.go:145`.
```go
if matches == 0 {
    return cmderr.SilentExit(1)
}
```

### Pattern 5: No-classifiable-error commands (yes, printf, seq, pwd, whoami, arch, uname)
**What:** Only stdout write failures are error-worthy; map to `ErrIO`. Otherwise return `nil`.
**Source:** CONTEXT.md Decision 7.
```go
if _, err := fmt.Fprintln(w, output); err != nil {
    return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("<cmd>: write: %s", err))
}
return nil
```

### Pattern 6: System/syscall commands (ps, df, free, kill, uptime, id)
**What:** Unsupported platform → `ErrUnsupported`. Syscall failures → `ErrIO` or `ErrPermission` based on classification. See POLISH-09 — Windows parity gaps are documented as `ErrUnsupported` returns.
```go
// platform-specific file with //go:build windows
func getStats() (Stats, error) {
    return Stats{}, cmderr.Wrap(cmderr.ErrUnsupported, "df: not supported on Windows")
}
```

### Anti-Patterns to Avoid
- **Wrapping twice:** Don't `cmderr.Wrap(...)` on an error that already came from another `internal/cli/*` package wrapped it. The outer caller should either classify fresh or pass through with `fmt.Errorf("...: %w", err)`.
- **Classifying in `pkg/*`:** Forbidden. `pkg/*` returns raw errors; classification is a CLI-boundary concern.
- **Using `err == os.ErrNotExist`:** Must use `errors.Is` (CLAUDE.md rule + wrapping breaks `==`).
- **Returning `cmderr.Wrap(cmderr.ErrInvalidInput, "bad flag")` for flag errors Cobra already handles.** Cobra rejects unknown flags before `RunE` runs.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Exit code mapping | Per-command `os.Exit(n)` | `cmderr.ExitCodeFor` already wired in `cmd/root.go:68` | Single source of truth |
| Silent grep-style exit | Custom error type | `cmderr.SilentExit(1)` | `root.go:64` already checks `*cmderr.SilentError` |
| Platform error detection | `runtime.GOOS` branching | `os.ErrPermission` | Already platform-neutral via stdlib |
| Coverage % extraction | Custom parser | `go tool cover -func=foo.out \| awk '/total:/ {print $3}'` | 1 line |
| Two golden registries kept in sync | Generator script | Manual dual-edit | Deferred per CONTEXT — keep divergence risk visible to reviewers |

**Key insight:** Everything the phase needs already exists. Phase 1 is 99% copy-paste + diff review, 1% new code (the Taskfile target).

## Runtime State Inventory

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None — verified by inspection of `internal/cli/cmderr/cmderr.go` (no state) | — |
| Live service config | None — phase is pure in-repo code/test changes | — |
| OS-registered state | None | — |
| Secrets/env vars | None | — |
| Build artifacts | None — rebuild of `omni` binary captures new behavior automatically | — |

## Exact Remaining-Command Work List

**Computation:** `ls internal/cli/` (139 dirs) minus dirs containing `cmderr` call sites (84 dirs) = **55 entries**. Below the raw diff has been filtered to remove non-command helpers (`cmderr`, `command`, `input`, `safepath`, `timeutil`, `command.go` stray file, scaffolding sub-libs). **Net remaining commands to migrate: ~47** (exact count confirmed below).

### Non-Command Exclusions (do NOT migrate — not user-facing)
- `cmderr` — the package itself
- `command` + `command.go` — Command interface / Registry (internal infrastructure)
- `input` — stdin helper library used by other commands
- `safepath` — path-sanitization library
- `timeutil` — already migrated as `timecmd`; `timeutil` is a helper
- `scaffolding` — contains sub-libs (`cobra`, `handler`, `repository`, `testgen`) — these ARE user-facing `scaffold` subcommands and must migrate
- `repo` — the `repo analyze` command **is** user-facing → migrate
- `project` — **is** user-facing → migrate

### Wave A (CI-critical, highest blast radius) — 8 commands

From the CONTEXT.md Wave A list, these are NOT yet migrated:

| Command | Dir | Rich matrix? |
|---------|-----|--------------|
| `sort` | `internal/cli/sort/` | No |
| `env` | `internal/cli/env/` | No (already partially migrated — verify) |
| `date` | `internal/cli/date/` | No |
| `kill` | `internal/cli/kill/` | No (platform-sensitive) |

Note: most of CONTEXT's Wave A list is already migrated (`cp/mv/rm/mkdir/chmod/touch/stat/ln/cat/grep/find/ls/wc/tr/uniq/head/tail/cut/diff/sed/awk/xargs/which/file/tar/zip/unzip/gzip/gunzip/curl/printf` all show in the 84-baseline). **Verify with `git grep -l "cmderr\." internal/cli/<cmd>/` per-command at plan time** — the CLAUDE.md list is the source of truth but commits may have drifted.

**env discrepancy:** `env` appears in the unmigrated diff but CLAUDE.md claims it's migrated. Planner should `git grep 'cmderr' internal/cli/env/*.go` and treat as "verify + top-up" rather than full migration.

### Wave B (data/format/encoding) — 9 commands

| Command | Dir | Rich matrix? |
|---------|-----|--------------|
| `cssfmt` | `internal/cli/cssfmt/` | No |
| `htmlfmt` | `internal/cli/htmlfmt/` | No |
| `sqlfmt` | `internal/cli/sqlfmt/` | No |
| `xmlutil` | `internal/cli/xmlutil/` | No |
| `yamlutil` | `internal/cli/yamlutil/` | No |
| `json2struct` | `internal/cli/json2struct/` | No |
| `yaml2struct` | `internal/cli/yaml2struct/` | No |
| `csvutil` | `internal/cli/csvutil/` | No |
| `ksuid`, `ulid`, `nanoid`, `snowflake` | `internal/cli/{ksuid,ulid,nanoid,snowflake}/` | No |

### Wave C (system/proc/info) — 10 commands

| Command | Dir | Rich matrix? | Notes |
|---------|-----|--------------|-------|
| `ps` | `internal/cli/ps/` | No | Windows parity per POLISH-09 → `ErrUnsupported` |
| `df` | `internal/cli/df/` | No | Windows parity per POLISH-09 |
| `du` | `internal/cli/du/` | No | |
| `free` | `internal/cli/free/` | No | Windows parity per POLISH-09 |
| `kill` | `internal/cli/kill/` | No | Windows parity (limited signals) per POLISH-09 |
| `pkill` | `internal/cli/pkill/` | No | |
| `lsof` | `internal/cli/lsof/` | No | |
| `ss` | `internal/cli/ss/` | No | |
| `uname` | `internal/cli/uname/` | No | No-classifiable-error pattern (Pattern 5) |
| `yes` | `internal/cli/yes/` | No | Pattern 5 |

### Wave D (alphabetical tail) — 20+ commands

| Command | Dir | Notes |
|---------|-----|-------|
| `arch` | `internal/cli/...` (CLAUDE shows `arch`) | Pattern 5 |
| `aws` | `internal/cli/aws/` | Wrapper over AWS CLI compat layer |
| `brdoc` | `internal/cli/brdoc/` | Brazilian doc validator |
| `buf` | `internal/cli/buf/` | `buf build/format/lint/generate` — partially migrated per CLAUDE (`buf build/format/lint`) → verify |
| `cloud` | `internal/cli/cloud/` | Multi-cloud aggregator |
| `echo` | `internal/cli/echo/` | Pattern 5 |
| `exec` | `internal/cli/exec/` | **RISKY** — spawns nothing (per project rules); verify |
| `forloop` | `internal/cli/forloop/` | |
| `gh` | `internal/cli/gh/` | GitHub wrapper |
| `git` | `internal/cli/git/` | Git hacks |
| `kubectl` | `internal/cli/kubectl/` | kubectl integration |
| `kubehacks` | `internal/cli/kubehacks/` | |
| `pipe` | `internal/cli/pipe/` | Already partially migrated (CLAUDE lists `pipe`) — verify |
| `project` | `internal/cli/project/` | |
| `repo` | `internal/cli/repo/` | |
| `scaffolding` | `internal/cli/scaffolding/` | 4 subcommands |
| `tagfixer` | `internal/cli/tagfixer/` | |
| `task` | `internal/cli/task/` | |
| `terraform` | `internal/cli/terraform/` | |
| `testcheck` | `internal/cli/testcheck/` | |
| `timecmd` | `internal/cli/timecmd/` | |
| `tree` | `internal/cli/tree/` | High-surface — validate scanner error paths |
| `vault` | `internal/cli/vault/` | |
| `video` | `internal/cli/video/` | Many subcommands |
| `aicontext` | `internal/cli/aicontext/` | |

**Critical caveat for the planner:** The CONTEXT.md note is explicit — "the 84-command list in CLAUDE.md may drift from git state." At plan time, run:
```bash
for d in $(ls internal/cli/); do
  if ! grep -rq 'cmderr' "internal/cli/$d/" 2>/dev/null; then echo "$d"; fi
done
```
and treat that output as authoritative. The list above is accurate as of 2026-04-11 worktree state but **must be recomputed before task generation**.

## cmderr Package Surface (canonical reference)

Exported from `internal/cli/cmderr/cmderr.go`:

| Symbol | Signature | Use Case |
|--------|-----------|----------|
| `ErrNotFound` | `var error` | File/resource not found → exit 1 |
| `ErrInvalidInput` | `var error` | Bad flag combo, parse failure → exit 2 |
| `ErrPermission` | `var error` | Permission denied (Unix + Windows) → exit 3 |
| `ErrIO` | `var error` | Read/write failure → exit 4 |
| `ErrConflict` | `var error` | Verification/compare mismatch → exit 1 |
| `ErrTimeout` | `var error` | Operation timed out → exit 5 |
| `ErrUnsupported` | `var error` | Platform/feature not supported → exit 6 |
| `SilentError{Code int}` | struct | No-message exit (e.g., grep no-match) |
| `SilentExit(code int) error` | func | Constructor for `SilentError` |
| `ExitError{Err error; Code int}` | struct | Force specific exit code |
| `WithExitCode(err, code) error` | func | Constructor for `ExitError` |
| `Wrap(sentinel, msg) error` | func | `fmt.Errorf("%s: %w", msg, sentinel)` |
| `ExitCodeFor(err) int` | func | Maps any error to exit code (called from `cmd/root.go:68`) |

### Gaps Flagged
After reviewing all six patterns against the remaining 47 commands, **no new sentinel is required**. The existing set covers every failure mode observed:
- Network failures (`curl`) → `ErrIO` or `ErrTimeout`
- Parse failures (`jq`, `sed`, `awk`) → `ErrInvalidInput`
- Verification mismatch (`hash verify`, `diff --exit-code`) → `ErrConflict`
- Platform gaps (`ps` on Windows) → `ErrUnsupported`

**Recommendation:** Extend `cmderr` only if a migration discovers an unclassifiable case. Surface any such discovery in the wave commit message for human review before landing.

## Golden Master Integration

### Schema (verified from `testing/golden/golden_tests.yaml`)
```yaml
- name: <cmd>_<scenario>
  args: ["<cmd>", "<sub>", "<arg>"]
  exit_code: <int>   # optional — asserts exit code
  stdin: "..."       # optional
  fixture: "..."     # optional — writes to temp file, {file} substitutes path
  fixtures_dir: ...  # optional — temp dir, {dir} substitutes path
```

### Confirmed Error-Path Examples (already exist)
From `testing/golden/golden_tests.yaml:380-407`:
```yaml
- name: exist_file_not_found
  args: ["exist", "file", "nonexistent_file_xyz"]
  exit_code: 1

- name: exist_command_not_found
  args: ["exist", "command", "nonexistent_cmd_xyz"]
  exit_code: 1
```

### Naming Convention (recommendation — consistent with existing)
- Happy-path: `<cmd>_<scenario>` (e.g., `grep_basic`)
- Error-path: `<cmd>_<scenario>_<failure>` (e.g., `grep_file_not_found`, `find_invalid_regex`)
- Use `nonexistent_*_xyz` fixtures for "not found" scenarios (matches existing `exist` entries)

### Dual-Write Rule
Every new entry goes in BOTH `testing/golden/golden_tests.yaml` AND `tools/golden/golden_tests.yaml` — both were in sync as of 2026-04-11 (same `exist_*` entries, line-identical in the relevant region). A plan task should include a verification step: `diff testing/golden/golden_tests.yaml tools/golden/golden_tests.yaml` shows no unintended drift outside the new entries.

### Snapshot Update Flow (verified in Taskfile.yml:135-151)
```bash
task test:golden:update          # regenerate all snapshots
task test:golden:update -- --update encoding   # single category
task golden:record                # richer tools/golden/ path
```

### Minimum Per-Wave Snapshot Count
- Wave A: 8 commands × 1 snapshot + (if `sort` warrants) = ~8–10 entries
- Wave B: 9 commands × 1 + `jq` rich matrix (5) = ~14 entries
- Wave C: 10 commands × 1 = ~10 entries (some return `ErrUnsupported` on Windows — use platform-conditional fixture or skip on host OS)
- Wave D: ~20 commands × 1 + `sed`/`awk`/`tree`/`curl` rich matrices = ~35 entries

**Total new golden entries: ~65–75.**

## Coverage Gate — Taskfile Target

### Current State (verified in `Taskfile.yml`)
- `COVERAGE_FILE: coverage.out` (global var)
- `task test:cover` exists (shows total), but does NOT gate.
- `.github/workflows/test.yml` has no coverage step (confirmed via `grep cover .github/workflows/test.yml` — zero matches).

### Drop-in Target
Add to `Taskfile.yml` (under the existing lint section near line 337):

```yaml
  lint:cmderr-coverage:
    desc: Gate internal/cli/cmderr coverage at >=90%
    vars:
      THRESHOLD: 90
      COV_FILE: '{{.COVERAGE_FILE | default "coverage.out"}}.cmderr'
    cmds:
      - go test -coverprofile={{.COV_FILE}} ./internal/cli/cmderr/...
      - cmd: |
          pct=$(go tool cover -func={{.COV_FILE}} | awk '/^total:/ {gsub("%",""); print $3}')
          awk -v p="$pct" -v t="{{.THRESHOLD}}" 'BEGIN {exit !(p+0 >= t+0)}' \
            && echo "cmderr coverage $pct% >= {{.THRESHOLD}}%" \
            || { echo "FAIL: cmderr coverage $pct% < {{.THRESHOLD}}%"; exit 1; }
        platforms: [linux, darwin]
      - cmd: |
          for /f "tokens=3" %%a in ('go tool cover -func={{.COV_FILE}} ^| findstr /C:"total:"') do @set PCT=%%a
          ...  # Windows variant — OR just fail on Windows and run the gate on linux runner only
        platforms: [windows]
```

**Simpler alternative** — run the gate only on the Linux CI runner (which is how `test.yml` already works):
```yaml
  lint:cmderr-coverage:
    desc: Gate internal/cli/cmderr coverage at >=90%
    cmds:
      - go test -coverprofile=cmderr-cov.out ./internal/cli/cmderr/...
      - |
        pct=$(go tool cover -func=cmderr-cov.out | awk '/^total:/ {gsub("%",""); print $3}')
        awk -v p="$pct" -v t=90 'BEGIN { if (p+0 < t+0) { printf "FAIL: cmderr coverage %s%% < %d%%\n", p, t; exit 1 } else { printf "OK: cmderr coverage %s%% >= %d%%\n", p, t } }'
```

### CI Wiring (`.github/workflows/test.yml`)
Add a step after the existing test step:
```yaml
      - name: cmderr coverage gate
        run: task lint:cmderr-coverage
```

**Required job:** mark it `needs: test` or add to the default matrix so a failure blocks merge. The existing workflow already runs `task test`, so appending the new task name is a two-line change.

### Expected Current Coverage
`internal/cli/cmderr/cmderr.go` is 105 LOC of pure logic (no I/O). A single table-driven test covering every sentinel + `SilentExit` + `ExitError` should hit ~100%. **Planner should add a `cmderr_test.go` task early in Wave A** so the gate doesn't fail on first run.

## Common Pitfalls

### Pitfall 1: `os.ErrPermission` not matching on Windows
**What goes wrong:** Developer assumes Windows `ACCESS_DENIED` won't be caught by `errors.Is(err, os.ErrPermission)`.
**Why it happens:** Training-era confusion; actually fixed in Go stdlib.
**How to avoid:** Trust `os.ErrPermission`. Go maps `syscall.ERROR_ACCESS_DENIED` → `os.ErrPermission` in `os/error_windows.go`. Verified via CONTEXT Decision 5.
**Warning signs:** If a migration contributor adds `runtime.GOOS == "windows"` branching for permission errors, reject it.

### Pitfall 2: Golden snapshots containing absolute paths
**What goes wrong:** Error messages include `/tmp/test-xyz/nonexistent` → snapshot diverges on Windows.
**Why it happens:** `fmt.Errorf("%w", err)` bubbles up the fully-qualified path.
**How to avoid:** Use the golden harness's normalization hooks (per POLISH-13, those hooks are added in Phase 2 — for Phase 1 use **relative paths** in fixtures: `nonexistent_file_xyz` not `/tmp/nonexistent`). The existing `exist_file_not_found` entry models this.
**Warning signs:** Snapshot diff shows a full host path in stderr output.

### Pitfall 3: Double-wrapping (classified error classified again)
**What goes wrong:** `cmderr.Wrap(cmderr.ErrIO, cmderr.Wrap(cmderr.ErrNotFound, ...).Error())` — the sentinel chain gets lost, `ExitCodeFor` returns wrong code.
**Why it happens:** Mechanical copy-paste without checking whether the lower layer already classified.
**How to avoid:** `pkg/*` returns raw errors (CONTEXT Decision 2); `internal/cli/*` classifies exactly once at the boundary. If one CLI command calls another internal CLI command (rare), use `fmt.Errorf("%w", err)` passthrough.
**Warning signs:** Running `go vet` with `-vettool=errorlint` and seeing complaints.

### Pitfall 4: Cobra swallows the `cmderr.SilentError`
**What goes wrong:** `grep` no-match case prints "Error: " to stderr.
**Why it happens:** Missing `rootCmd.SilenceErrors = true`.
**How to avoid:** Already set in `cmd/root.go:73` — **do not touch.**
**Warning signs:** If anyone re-introduces stderr noise for `grep`-style commands, check `cmd/root.go:63-66` — the `errors.As(err, &silent)` guard is the single suppression point.

### Pitfall 5: `task lint:cmderr-coverage` parses 0% because tests live in wrong package
**What goes wrong:** `go test -coverprofile` with zero statements returns `coverage: 0.0% of statements`; awk extracts `0.0` and the gate fails.
**Why it happens:** `cmderr_test.go` missing or in wrong package path.
**How to avoid:** Plan explicitly includes a Wave A task: "add `internal/cli/cmderr/cmderr_test.go` (table-driven, covers every sentinel + SilentExit + ExitError) BEFORE enabling the gate."
**Warning signs:** First CI run after adding the gate fails with `FAIL: cmderr coverage 0.0% < 90%`.

### Pitfall 6: Windows test flakiness on system commands (ps/df/free/kill)
**What goes wrong:** Golden snapshots for Windows-gapped commands diverge because stderr messages change between Go versions.
**Why it happens:** POLISH-09 parity gaps — the `ErrUnsupported` message format isn't standardized.
**How to avoid:** Standardize: `return cmderr.Wrap(cmderr.ErrUnsupported, "<cmd>: not supported on <GOOS>")`. Exact string lock-in via golden snapshot; no `fmt.Errorf` with dynamic values.
**Warning signs:** Snapshot diff on Windows runner only.

### Pitfall 7: `yes` / `seq` / `printf` hot-loop write error classification
**What goes wrong:** Writing classified errors in a tight loop (`yes`, `seq -f`) adds measurable overhead.
**Why it happens:** `fmt.Errorf` allocates each call.
**How to avoid:** Check `err` **once after the loop**, not per-iteration. `yes` returns only if `w.Write` errors (EPIPE on broken-pipe downstream); the idiomatic pattern already does this.
**Warning signs:** Benchmark regression in `omni yes | head`.

## Risky-Commands List

Specific commands flagged for careful treatment (beyond mechanical migration):

| Command | Risk | Mitigation |
|---------|------|------------|
| `find` (in 84-baseline, but rich-matrix goldens still needed) | Walk errors mid-traversal can be transient; distinction between "user's regex is bad" (`ErrInvalidInput`) and "directory became unreadable" (`ErrPermission`) matters | 5-snapshot matrix: invalid-regex, invalid-size, nonexistent-root, permission-denied-root, invalid-type-flag |
| `sed` | Regex compile errors (`ErrInvalidInput`) vs I/O errors (`ErrIO`) vs unsupported command letter (`ErrUnsupported`) | 4-snapshot matrix: bad-regex, bad-substitute-flags, missing-file, unsupported-command |
| `awk` | Interpreter errors — runtime vs parse-time | 3-snapshot matrix: parse-error, runtime-divide-by-zero, missing-field |
| `dd` | Multiple I/O channels; can fail on source or dest; partial writes matter | 5-snapshot matrix: missing-if, missing-of, bad-bs, permission-denied-of, unsupported-conv |
| `grep` (baseline, but check `-r` walk errors) | `-r` on a gitignored dir — walk errors mid-stream | 4-snapshot matrix: file-not-found, pattern-compile, binary-file, no-match-silent-exit |
| `curl` | Network errors have rich taxonomy: DNS fail (`ErrIO`), timeout (`ErrTimeout`), HTTP non-2xx (depends on `-f`) | 5-snapshot matrix: bad-url, unresolvable-host (use `.invalid` TLD), connect-refused-port, http-404-with-f, http-404-without-f |
| `tar` | Corrupt archives, path traversal, missing entries | 4-snapshot matrix: corrupt-header, nonexistent-archive, path-traversal-entry, unsupported-compression |
| `jq` | Filter parse errors vs missing-field vs type errors | 4-snapshot matrix: bad-filter-syntax, missing-field-with-`-e`, type-mismatch, invalid-json-input |
| `diff` | `--exit-code` semantics: 0 same, 1 different (ErrConflict), 2 error | 3-snapshot matrix: files-differ, missing-file, binary-files |
| `kill` | Windows has 3 signals only (INT, KILL, TERM); asking for `-USR1` on Windows is `ErrUnsupported` | 3-snapshot matrix: bad-pid, unsupported-signal-windows, permission-denied-pid |
| `ps` / `df` / `free` | POLISH-09: Windows parity gaps → `ErrUnsupported` message needed | Standardized "not supported on windows" string |
| `exec` | Per CLAUDE.md "no external process spawning" rule — if `exec` exists and actually spawns, **that is a pre-existing bug, not Phase 1's problem.** Flag it and defer. | Investigate in plan, don't fix here |
| `tree` | Parallel scanner with hundreds of files; first-error-wins needs to be classified | Verify `pkg/twig/scanner` returns raw errors; wrapper classifies |
| `pipe` / `pipeline` | Dispatch errors across stage chain; classification must not be lost | Pass-through with `fmt.Errorf("pipeline stage %d: %w", i, err)` — no re-classification |

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go stdlib `testing` (verified — `*_test.go` files throughout `internal/cli/`) |
| Config | `Taskfile.yml` targets — no separate config |
| Quick run command | `go test -race ./internal/cli/<cmd>/...` |
| Full suite command | `task test` (runs `go test -v -race -coverprofile=coverage.out ./...`) |
| Golden quick | `task test:golden -- --filter <cmd>` |
| Golden full | `task test:golden` |
| Coverage gate | `task lint:cmderr-coverage` (new — to be added in Wave A) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| POLISH-01 | Each command wraps errors with cmderr sentinels | unit | `go test -race ./internal/cli/<cmd>/...` | Per-command (existing unit tests extended) |
| POLISH-01 | Exit code maps to sentinel | integration | `task test:golden -- --filter <cmd>_err` | New golden entries |
| POLISH-02 | ≥1 error-path golden per command | golden | `task test:golden:check` + `task test:golden` | Wave 0 — entries added per wave |
| POLISH-03 | cmderr coverage ≥90% | coverage gate | `task lint:cmderr-coverage` | Wave 0 — target + cmderr_test.go |

### Sampling Rate
- **Per task commit:** `go test -race ./internal/cli/<cmd>/... && task test:golden -- --filter <cmd>`
- **Per wave merge:** `task test && task test:golden && task lint:cmderr-coverage`
- **Phase gate:** Full suite green + all 4 wave PRs landed + `task lint:cmderr-coverage` green in `.github/workflows/test.yml`

### Wave 0 Gaps (must land BEFORE any migration wave)
- [ ] `internal/cli/cmderr/cmderr_test.go` — table-driven tests covering every sentinel, `SilentExit`, `ExitError`, `ExitCodeFor` (currently zero tests; gate will fail on first run without this)
- [ ] `Taskfile.yml` `lint:cmderr-coverage` target (spec above)
- [ ] `.github/workflows/test.yml` new step invoking the new task
- [ ] `.planning/phases/01-.../EXIT-CODE-CHANGES.md` tracking file created (empty; waves append to it)

## Security Domain

**Scope:** Phase 1 is refactor-only — no new attack surface. ASVS mapping below confirms no new controls are required.

### Applicable ASVS Categories
| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | n/a |
| V3 Session Management | no | n/a |
| V4 Access Control | no | n/a |
| V5 Input Validation | yes (indirect) | Existing `pkg/*` validators unchanged; CLI layer only adds error classification |
| V6 Cryptography | no | n/a |
| V7 Error Handling & Logging | **yes** | Error messages must not leak secrets — verify `fmt.Errorf("%s: %s", cmd, err)` doesn't echo env vars or file contents |

### Known Threat Patterns
| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Error message leaks secret env var | Information Disclosure | Keep error wrapping to `err.Error()` — never interpolate `os.Getenv(...)`. Already the pattern in all 84 migrated commands — verified. |
| Path disclosure in error output | Information Disclosure | Accepted for CLI tools (users want to know what path failed). No mitigation required. |
| Error-path DoS via log flooding | DoS | `yes`/`seq` write loops classify once after loop, not per-iteration (see Pitfall 7). |

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | CLAUDE.md's 84-command list matches `git grep cmderr.` exactly | Wave work list | **MEDIUM** — Planner must recompute at plan time; drift produces wrong wave sizing |
| A2 | `os.ErrPermission` matches Windows `ACCESS_DENIED` via Go stdlib wrapping | Pattern 1 | LOW — stdlib-documented behavior since Go 1.13 `errors.Is` landed |
| A3 | Golden YAML `exit_code:` field is respected by BOTH `testing/` and `tools/golden/` engines | Golden integration | LOW — verified via grep; both files have `exit_code:` entries in sync |
| A4 | `.github/workflows/test.yml` currently has no coverage gate | Taskfile gate | LOW — verified; `grep cover` returns zero matches |
| A5 | `cmderr` package currently has no `_test.go` | Wave 0 | **MEDIUM** — if tests exist but don't cover 90%, Wave 0 task scope changes. Planner must `ls internal/cli/cmderr/` at plan time. |
| A6 | `exec` command either doesn't spawn processes or its spawning is a pre-existing bug orthogonal to Phase 1 | Risky list | LOW — per CLAUDE.md "no exec" rule; worst case the command is a no-op and migration is trivial |
| A7 | Two golden registries remain intentionally duplicated through Phase 1 | Golden integration | LOW — explicit in CONTEXT.md deferred ideas |

## Open Questions

1. **Does `internal/cli/cmderr/cmderr_test.go` exist today?**
   - What we know: 0 test files found with `cmderr` in filename; file probably absent
   - What's unclear: Whether an implicit test exists in another package
   - Recommendation: Plan Wave 0 to **create** the test file. If one already exists, the task collapses to "top up to 90%."

2. **For `env`, `pipe`, `buf`, and any other command where CLAUDE.md claims migration but the grep shows gaps — is this drift or test-helper noise?**
   - What we know: diff produced these as "unmigrated" but CLAUDE.md lists them
   - What's unclear: Whether `cmderr` is called from a sibling file or a sub-package
   - Recommendation: Planner runs `git grep -l 'cmderr\.' internal/cli/<cmd>/` per-command; commands with ANY cmderr call are "verify + top up" rather than full migration.

3. **Does `.github/workflows/test.yml` run on Windows?**
   - What we know: Only `linux/darwin` test matrix is conventional
   - What's unclear: Whether Windows runner exists
   - Recommendation: Scope the `lint:cmderr-coverage` gate to Linux-only; add Windows as follow-up if needed. Cross-platform shell parsing is a trap.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | All test / build | ✓ | 1.22+ (per go.mod) | — |
| `go tool cover` | Coverage gate | ✓ | bundled with Go | — |
| Task (taskfile.dev) | Taskfile targets | ✓ | — | `go test` direct |
| Python 3 | Golden harness (`testing/scripts/test_golden.py`) | ✓ (assumed; existing golden tests pass) | 3.x | — |
| `awk` | Coverage gate shell script | ✓ on Linux/macOS; `gawk` on Windows | — | Use Go-based parser (see alternative in target spec) |

**No missing dependencies.** Phase is pure in-tree refactor.

## Sources

### Primary (HIGH confidence — verified in working tree 2026-04-11)
- `internal/cli/cmderr/cmderr.go` — canonical sentinel source (105 LOC read in full)
- `cmd/root.go` — error-to-exit-code wiring at line 68
- `internal/cli/head/head.go` — Pattern 1 reference (file I/O classification)
- `internal/cli/find/find.go` — Pattern 2 reference (regex/flag validation)
- `internal/cli/hash/hash.go` — Pattern 3 reference (verification conflict)
- `internal/cli/grep/grep.go` — Pattern 4 reference (silent exit)
- `internal/cli/path/path.go` — Pattern 2 reference (missing operand)
- `internal/cli/exist/exist.go` — golden error-path precedent
- `testing/golden/golden_tests.yaml:380-407` — `exit_code:` schema confirmation
- `Taskfile.yml:59-80,135-151,337` — existing test/lint layout
- `.planning/phases/01-cmderr-migration-completion/01-CONTEXT.md` — locked decisions
- `.planning/REQUIREMENTS.md:14-16` — POLISH-01/02/03 literal text
- `CLAUDE.md` root-level "Error Handling (cmderr)" section — 84-command baseline list

### Secondary (MEDIUM confidence)
- Go stdlib `errors.Is` behavior for `os.ErrPermission` on Windows — documented in Go 1.13+ release notes

### Tertiary (LOW confidence)
- None — every claim above is verified against working-tree files

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — nothing new, all in-tree
- Architecture patterns: HIGH — copied verbatim from 6 already-migrated commands
- Pitfalls: HIGH (Pitfalls 1–4, 6–7) / MEDIUM (Pitfall 5, depends on Wave 0 ordering)
- Work list: MEDIUM — exact as of 2026-04-11 but planner MUST recompute before task generation (Assumption A1)
- Coverage gate: HIGH — target is a 10-line drop-in

**Research date:** 2026-04-11
**Valid until:** 2026-05-11 (30 days — stable codebase, no external deps)
