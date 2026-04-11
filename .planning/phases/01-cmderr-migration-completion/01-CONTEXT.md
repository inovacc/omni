# Phase 1 Context — cmderr Migration Completion

**Phase:** 1
**Goal:** Every command in omni returns classified `cmderr` sentinels with correct exit codes, eliminating inconsistency before any supply-chain command is written.
**Requirements:** POLISH-01, POLISH-02, POLISH-03
**Baseline:** 84 commands (272 cmderr call sites) already migrated; ~76 remaining.

## Decisions

### Migration Order — Risk-weighted

Migrate in waves, ordered by CI-impact rather than alphabet:

1. **Wave A (CI-critical, user-facing, highest blast radius):** `cp`, `mv`, `rm`, `mkdir`, `chmod`, `touch`, `stat`, `ln`, `cat`, `grep`, `find`, `ls`, `wc`, `tr`, `sort`, `uniq`, `head`, `tail`, `cut`, `diff`, `sed`, `awk`, `xargs`, `env`, `which`, `file`, `tar`, `zip`, `unzip`, `gzip`, `gunzip`, `curl`, `date`, `printf`
2. **Wave B (data/format/encoding):** `jq`, `yq`, `json*`, `yaml*`, `toml*`, `xml*`, `sql`, `html*`, `css*`, `base*`, `hex*`, `urlenc`, `htmlenc`, `hash`, `sha*`, `md5*`, `crc*`, `xxd`, `uuid`, `random`, `jwt`
3. **Wave C (system/proc/info):** `ps`, `kill`, `df`, `du`, `free`, `uptime`, `uname`, `whoami`, `id`, `dotenv`, `watch`, `yes`, `time`, `sleep`, `seq`
4. **Wave D (everything else, tail):** remaining commands in alphabetical order — `banner`, `brdoc`, `cron`, `forloop`, `input`, `lint`, `loc`, `note`, `pager`, `safepath`, `tagfixer`, `tree`, etc.

**Rationale:** user's primary audience is their CI/CD pipelines. Wins in Wave A compound across the most-used commands; later waves are mechanical polish.

**Note:** 84-command "already migrated" list from `CLAUDE.md` contains some Wave A targets (`cat`, `grep`, `find`, `sort`, `head`, `tail`, etc.) — the actual remaining set must be computed at plan time by diffing `git grep 'cmderr\.'` against `ls internal/cli/*/`. The wave ordering above is the intent, not a literal work list.

### pkg/ wrapping depth — CLI boundary only

`pkg/*` packages continue to return raw errors (including `os.ErrNotExist`, `os.ErrPermission`, typed parser errors, etc.). The `internal/cli/*/<cmd>.go` wrapper is the sole place where errors get classified into `cmderr` sentinels via `errors.Is` + `cmderr.Wrap`.

**Why:** `pkg/*` is importable by external Go projects (per PROJECT.md and CLAUDE.md "library-first" principle). Those consumers must not be forced to import `internal/cli/cmderr` semantics.

**Implication for planning:** every migration wave touches ONLY `internal/cli/<cmd>/<cmd>.go` files (plus their tests). `pkg/*` files are off-limits during this phase.

### Golden test breadth — one error snapshot per command (minimum)

POLISH-02 literal requirement: "every migrated command has at least one golden-master test case exercising an error path". That's the floor.

**Per-command baseline:** one error snapshot. Pick the most likely error path (usually "file not found" for file commands, "invalid flag" for flag-heavy commands, "malformed input" for parsers).

**Exceptions — add a richer error matrix** for these rich-surface commands: `find`, `sed`, `awk`, `dd`, `grep`, `curl`, `tar`, `jq`, `diff`. Each gets 3–5 error snapshots covering distinct classes.

**Registry:** snapshots land in both `testing/golden/golden_tests.yaml` and `tools/golden/golden_tests.yaml` per CLAUDE.md convention.

### Coverage gate — custom Taskfile target

Add a new `task lint:cmderr-coverage` target to `Taskfile.yml` that:

1. Runs `go test -coverprofile=$TEMP/cmderr-cov.out ./internal/cli/cmderr/...`
2. Parses the profile with `go tool cover` and extracts the cmderr package coverage number
3. Fails with exit code 1 if coverage < 90%
4. Is called from `.github/workflows/test.yml` as a required job

**Why:** no new external CI services; everything stays in-repo; matches CLAUDE.md "no exec external" aesthetic (even though this runs in CI, not omni). Tool written in Go, lives in `tools/cmderr-cov/` or similar.

### Platform-specific errors — single sentinel, platform-detail in message

Windows ACL denials (`syscall.ERROR_ACCESS_DENIED`) and Unix EPERM both map to `cmderr.ErrPermission`. The wrapped message includes platform-specific detail via `fmt.Errorf`, e.g.:

- Unix: `cmderr: permission denied: open /etc/shadow: permission denied`
- Windows: `cmderr: permission denied: open C:\Users\admin\NTUSER.DAT: Access is denied.`

**Classification logic** in each command:
```go
if errors.Is(err, os.ErrPermission) {
    return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("%s: %s", cmdName, err))
}
```

`os.ErrPermission` is the Go stdlib's platform-neutral sentinel and already wraps the syscall-level errors correctly — no manual GOOS branching needed.

### Breaking-change audit — none, document in release notes

omni has never cut a stable release — exit codes are not yet a public contract. Wave migrations are intentional behavior changes, not regressions.

**Mitigation:**
- v1.0 release notes will explicitly state "v1.0 is the first stable exit-code contract"
- Any command whose exit code changes during Phase 1 is logged in a `.planning/phases/01-.../EXIT-CODE-CHANGES.md` tracking file for the release notes to pull from later
- Each migration commit references POLISH-01 and names the command(s) affected, so git history is the audit trail

### No-classifiable-error commands — ErrIO for write failures, nil otherwise

Commands like `yes`, `printf`, `seq`, `pwd`, `whoami`, `arch`, `uname` have no meaningful input validation to reach an invalid-input error (Cobra already gates bad flags). Their only runtime failure mode is stdout write errors (broken pipe, disk full on redirect).

**Convention:**
```go
if _, err := fmt.Fprintln(w, output); err != nil {
    return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("yes: write: %s", err))
}
return nil
```

No special marker comments, no belt-and-suspenders ErrInvalidInput for args Cobra already validates.

## Specifics the researcher / planner should know

- **The existing 84-command baseline is proven.** Find one representative from each wave (e.g., `find`, `jq`, `ps`, `banner`) and lift its error-wrapping pattern verbatim — no reinvention.
- **`internal/cli/cmderr/cmderr.go`** is the canonical source for sentinels. If a migration reveals a missing sentinel, extending `cmderr` is in-scope for this phase but must be discussed before landing.
- **Golden master registries live in two places** (`testing/golden/` + `tools/golden/`). Both must stay in sync — do not consolidate during Phase 1 (that's post-1.0 cleanup per PROJECT.md).
- **Don't touch `pkg/*`** in this phase except to read. Error-wrapping stays at `internal/cli/*` boundaries.
- **Don't touch `cmd/*`** beyond adjusting exit-code tests if they exist. Cobra wiring is already in place.
- **Coverage target is scoped to `internal/cli/cmderr` only** — raising omni-wide coverage is a *different* phase (Phase 2).
- **Don't chase commands that are `test.go`-only wrappers, timeutil helpers, or safepath libraries** — those aren't user-facing commands.

## Deferred ideas (not scope creep, noted for backlog)

- **Exit-code documentation page** — a generated `docs/EXIT-CODES.md` listing every sentinel and its meaning. Deferred to Phase 3 (docs completeness).
- **`cmderr.Is<Class>()` convenience helpers** for external packages to check sentinel classes without importing the package — deferred; decide during Phase 3 API audit.
- **Structured logging integration** — emitting cmderr classifications as slog attributes automatically — deferred to post-1.0.
- **Error-type lint rule** — a `golangci-lint` custom linter that flags raw `os.ErrX` returns in `internal/cli/*` — nice-to-have, revisit in Phase 2 if mechanical regressions keep appearing.
- **Cross-command exit-code golden matrix** — a single `exit-codes.yaml` asserting every command's exit code for every classified error — Phase 2 or later.

## Open questions for the researcher

- **Is there a canonical list of "what's already migrated"?** `git grep cmderr\.` gives 84 files; that's the authoritative set. Researcher should compute the diff against `ls internal/cli/*/` to produce the exact work list.
- **Are there commands in `internal/cli/` that are intentionally excluded from the CLI surface** (internal helpers, test fixtures)? Researcher should flag those so they don't pollute the migration count.
- **What's the current golden-master test naming convention for error paths?** (so new snapshots follow it)
- **Does `Taskfile.yml` already have a coverage-related target we should extend vs add new?**

## Ready for

`/gsd-plan-phase 1` — researcher investigates the exact remaining-command list, planner decomposes into wave-based plans (Wave A / B / C / D) with 3–5 plans per wave.
