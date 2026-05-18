---
phase: 01-cmderr-migration-completion
plan: 02
type: execute
wave: 0
depends_on: [01]
files_modified:
  - Taskfile.yml
  - .github/workflows/test.yml
autonomous: true
requirements: [POLISH-03]
must_haves:
  truths:
    - "`task lint:cmderr-coverage` runs cmderr tests and fails hard if coverage <90%"
    - "CI workflow invokes the gate on every PR"
    - "Gate passes today (post Plan 01) with the new cmderr_test.go"
  artifacts:
    - path: "Taskfile.yml"
      provides: "lint:cmderr-coverage target"
      contains: "lint:cmderr-coverage"
    - path: ".github/workflows/test.yml"
      provides: "CI step invoking the gate"
      contains: "lint:cmderr-coverage"
  key_links:
    - from: ".github/workflows/test.yml"
      to: "Taskfile.yml lint:cmderr-coverage"
      via: "task CLI invocation in workflow step"
      pattern: "task lint:cmderr-coverage"
---

# Plan 02 — Taskfile Coverage Gate + CI Wiring (Wave 0)

## Goal

Add `task lint:cmderr-coverage` target (per RESEARCH.md §"Coverage Gate — Taskfile Target") and wire it into `.github/workflows/test.yml` as a required CI step.

## Wave

Wave 0. Depends on Plan 01 so first run succeeds (POLISH-03 failure-to-launch guard, Pitfall 5).

## Requirements covered

POLISH-03.

## Depends on

Plan 01 (cmderr_test.go must exist so the gate doesn't fail at 0.0%).

## Parallelizable with

Plan 03 (84-command audit) — independent files.

## Commands touched

None.

## Context

@.planning/phases/01-cmderr-migration-completion/01-RESEARCH.md
@Taskfile.yml
@.github/workflows/test.yml

## Tasks

### Task 1: Add `lint:cmderr-coverage` target to Taskfile.yml

**Files:** `Taskfile.yml`

**Action:** Append a new target under the existing `lint:` section (near line 337 per RESEARCH.md). Use the "simpler alternative" from RESEARCH.md §"Drop-in Target":

```yaml
  lint:cmderr-coverage:
    desc: Gate internal/cli/cmderr coverage at >=90%
    cmds:
      - go test -coverprofile=cmderr-cov.out ./internal/cli/cmderr/...
      - |
        pct=$(go tool cover -func=cmderr-cov.out | awk '/^total:/ {gsub("%",""); print $3}')
        awk -v p="$pct" -v t=90 'BEGIN { if (p+0 < t+0) { printf "FAIL: cmderr coverage %s%% < %d%%\n", p, t; exit 1 } else { printf "OK: cmderr coverage %s%% >= %d%%\n", p, t } }'
```

Scope the gate to Linux runner only (per Research Open Question 3); if the existing `lint:` target uses `platforms:`, mirror it.

**Verify:**
```
<automated>task lint:cmderr-coverage</automated>
```
Must print `OK: cmderr coverage` and exit 0.

**Done:** `task --list | grep cmderr-coverage` shows the new target; invocation passes locally.

### Task 2: Wire the gate into `.github/workflows/test.yml`

**Files:** `.github/workflows/test.yml`

**Action:** Add a new step after the existing `task test` step (or its equivalent):

```yaml
      - name: cmderr coverage gate
        run: task lint:cmderr-coverage
```

If the workflow uses a matrix, scope this step to `matrix.os == 'ubuntu-latest'` with an `if:` condition — do NOT run the awk-based parser on Windows runners.

**Verify:**
```
<automated>grep -q 'lint:cmderr-coverage' .github/workflows/test.yml</automated>
```

**Done:** Workflow references the new task name; `gh workflow view test.yml` (after push) shows the step.

## Golden test additions

None.

## Verification

```bash
task lint:cmderr-coverage            # Must pass locally
grep -c 'lint:cmderr-coverage' Taskfile.yml .github/workflows/test.yml
```

## Out of scope

- Raising the threshold above 90% (phase-wide target is exactly 90%)
- Gating other packages (Phase 2 scope)
- Windows coverage parsing (deferred — Linux runner only)
