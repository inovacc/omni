---
phase: 01-cmderr-migration-completion
plan: 15
type: execute
wave: D
depends_on: [09, 10, 11]
files_modified:
  - internal/cli/banner/banner.go
  - internal/cli/brdoc/brdoc.go
  - internal/cli/cron/cron.go
  - internal/cli/forloop/forloop.go
  - internal/cli/echo/echo.go
  - internal/cli/arch/arch.go
  - internal/cli/aicontext/aicontext.go
  - internal/cli/tagfixer/tagfixer.go
  - internal/cli/timecmd/timecmd.go
  - internal/cli/note/note.go
  - internal/cli/pager/pager.go
  - internal/cli/video/video.go
  - testing/golden/golden_tests.yaml
  - tools/golden/golden_tests.yaml
  - .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
autonomous: true
requirements: [POLISH-01, POLISH-02]
must_haves:
  truths:
    - "Every remaining tail command in internal/cli/ is classified via cmderr"
    - "Pattern 5 commands (echo, arch) return ErrIO only for write failures"
    - "video has ≥1 error-path golden covering a common failure mode (bad URL)"
    - "Phase 1 work list is exhausted — no unmigrated commands remain"
  artifacts:
    - path: "internal/cli/banner/banner.go"
      provides: "cmderr classification"
      contains: "cmderr."
    - path: "internal/cli/video/video.go"
      provides: "cmderr classification for video download errors"
      contains: "cmderr."
  key_links:
    - from: "tail command CLI wrappers"
      to: "cmderr sentinels"
      via: "Patterns 1/2/5 classification"
      pattern: "cmderr\\."
---

# Plan 15 — Wave D: Misc tail (banner, brdoc, cron, forloop, echo, arch, aicontext, tagfixer, timecmd, note, pager, video)

## Goal

Migrate all remaining tail commands. This plan closes the Wave D work list — after it merges, every command under `internal/cli/` is cmderr-classified.

## Wave

Wave D.

## Requirements covered

POLISH-01, POLISH-02.

## Depends on

Plans 09, 10, 11.

## Parallelizable with

Plans 12, 13, 14.

## Commands touched

- `internal/cli/banner/`, `brdoc/`, `cron/`, `forloop/`, `echo/`, `arch/`, `aicontext/`, `tagfixer/`, `timecmd/`, `note/`, `pager/`, `video/`

If MIGRATION-LEDGER.md (Plan 03) finds additional unmigrated commands not listed here, ADD them to this plan's file list and task action — this plan is the Wave D catch-all.

## Context

@.planning/phases/01-cmderr-migration-completion/01-RESEARCH.md
@.planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md
@internal/cli/head/head.go
@internal/cli/find/find.go

## Tasks

### Task 1: Migrate Pattern 5 commands (echo, arch) + simple Pattern 2 commands (banner, brdoc, forloop, cron, aicontext, tagfixer, timecmd, note, pager)

**Files:** as listed in frontmatter (minus video/).

**Action:** Mechanical batch migration:

- **echo, arch** — Pattern 5: write errors → `ErrIO`, otherwise nil. No input validation (Cobra handles flags).
- **banner** — Pattern 5 + `ErrInvalidInput` for unknown font.
- **brdoc** — Validator: Pattern 2; invalid doc format → `ErrInvalidInput`; mismatch → `ErrConflict`.
- **cron** — Cron expression parse error → `ErrInvalidInput`.
- **forloop** — Invalid range / step → `ErrInvalidInput`; write errors → `ErrIO`.
- **aicontext** — File I/O → Pattern 1; template errors → `ErrInvalidInput`.
- **tagfixer** — File I/O → Pattern 1; AST parse error → `ErrInvalidInput`.
- **timecmd** — Time parse error → `ErrInvalidInput`.
- **note** — File I/O → Pattern 1; write-fail → `ErrIO`.
- **pager** — Terminal setup errors → `ErrIO`; file not found → `ErrNotFound`.

**Verify:**
```
<automated>for d in banner brdoc cron forloop echo arch aicontext tagfixer timecmd note pager; do go test -race ./internal/cli/$d/... || exit 1; done</automated>
```

**Done:** Each command's error paths classified; tests pass.

### Task 2: Migrate video (command group with subcommands)

**Files:** `internal/cli/video/video.go`, `internal/cli/video/download.go`, `internal/cli/video/info.go`, `internal/cli/video/formats.go`, `internal/cli/video/channel.go`, `internal/cli/video/*_test.go`

**Action:**
- Invalid URL → `ErrInvalidInput`.
- Unresolvable host / network failure → `ErrIO`.
- HTTP 404 from extractor → `ErrNotFound`.
- Extractor not found for URL → `ErrUnsupported`.
- Channel DB open failure → `ErrIO`.
- `pkg/video/*` stays raw-error (CONTEXT Decision 2). Wrap only at `internal/cli/video/*.go` boundaries.

**Verify:**
```
<automated>go test -race ./internal/cli/video/...</automated>
```

**Done:** All subcommand error paths classified.

### Task 3: Golden error snapshots

**Files:** `testing/golden/golden_tests.yaml`, `tools/golden/golden_tests.yaml`

Minimum per-command snapshots (12 entries):

```yaml
- name: banner_unknown_font
  args: ["banner", "--font", "definitely-not-a-font", "hi"]
  exit_code: 2
- name: brdoc_invalid_cpf
  args: ["brdoc", "cpf", "000.000.000-00"]
  exit_code: 1
- name: cron_invalid_expression
  args: ["cron", "parse", "not a cron"]
  exit_code: 2
- name: forloop_invalid_range
  args: ["for", "--start", "10", "--end", "1", "--step", "1"]
  exit_code: 2
- name: echo_basic
  args: ["echo", "hello"]
  exit_code: 0
- name: arch_basic
  args: ["arch"]
  exit_code: 0
- name: aicontext_nonexistent
  args: ["aicontext", "nonexistent_dir_xyz"]
  exit_code: 1
- name: tagfixer_nonexistent
  args: ["tagfixer", "nonexistent_file_xyz.go"]
  exit_code: 1
- name: timecmd_invalid_format
  args: ["timecmd", "parse", "not-a-time"]
  exit_code: 2
- name: note_nonexistent
  args: ["note", "read", "nonexistent-note-xyz"]
  exit_code: 1
- name: pager_file_not_found
  args: ["pager", "nonexistent_file_xyz"]
  exit_code: 1
- name: video_bad_url
  args: ["video", "info", "not-a-url"]
  exit_code: 2
```

Commands without a natural error path get a happy-path snapshot (echo, arch) per CONTEXT Decision 7 — Pattern 5 commands have nothing to classify except write failures, which can't be golden-tested easily.

**Verify:**
```
<automated>task test:golden -- --filter 'banner_|brdoc_|cron_|forloop_|echo_|arch_|aicontext_|tagfixer_|timecmd_|note_|pager_|video_'</automated>
```

**Done:** Snapshots green in both registries.

### Task 4: Log exit-code changes + final Wave D sweep

**Files:** `.planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md`

Also run the authoritative closing check:

```bash
for d in $(omni ls internal/cli/); do
  if ! omni grep -rq 'cmderr' "internal/cli/$d/" 2>/dev/null; then
    echo "STILL UNMIGRATED: $d"
  fi
done
```

Any output must be either a non-command helper from CONTEXT Decision 2 exclusions or a new plan (reopen MIGRATION-LEDGER.md).

## Golden test additions

12 snapshots listed above (some happy-path where Pattern 5 precludes error).

## Verification

```bash
go test -race ./internal/cli/banner/... ./internal/cli/brdoc/... ./internal/cli/cron/... ./internal/cli/forloop/... ./internal/cli/echo/... ./internal/cli/arch/... ./internal/cli/aicontext/... ./internal/cli/tagfixer/... ./internal/cli/timecmd/... ./internal/cli/note/... ./internal/cli/pager/... ./internal/cli/video/...
task test:golden
task lint:cmderr-coverage
```

## Out of scope

- Any `pkg/*` changes
- Commands NOT in the Wave D catch-all — if Plan 03's ledger lists extras, add them here OR spawn an addendum plan (Phase 1 must not leave anything unmigrated)
