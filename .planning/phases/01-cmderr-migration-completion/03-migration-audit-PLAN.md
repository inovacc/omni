---
phase: 01-cmderr-migration-completion
plan: 03
type: execute
wave: 0
depends_on: []
files_modified:
  - .planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md
autonomous: true
requirements: [POLISH-01]
must_haves:
  truths:
    - "Authoritative remaining-command list exists and supersedes any stale lists in CLAUDE.md or RESEARCH.md"
    - "Every command in internal/cli/ is classified as: already-migrated / needs-migration / verify-topup / non-command-helper"
    - "Wave A/B/C/D plans (04–19) reference the ledger, not the snapshot in RESEARCH.md"
  artifacts:
    - path: ".planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md"
      provides: "Authoritative work list, regenerated from live git state"
      contains: "Already-Migrated, Needs-Migration, Verify-Topup, Excluded"
  key_links:
    - from: "MIGRATION-LEDGER.md"
      to: "Wave A/B/C/D plans"
      via: "plan frontmatter references this file"
      pattern: "MIGRATION-LEDGER"
---

# Plan 03 — Migration Audit Ledger (Wave 0)

## Goal

Compute the authoritative remaining-command list from live git state per RESEARCH.md Assumption A1, resolve the `env`/`pipe`/`buf` drift (Research Open Question 2), and produce `MIGRATION-LEDGER.md` that Wave A-D plans execute against. RESEARCH.md's list is a 2026-04-11 snapshot — this plan refreshes it at execution time.

## Wave

Wave 0. Runs in parallel with Plans 01 and 02.

## Requirements covered

POLISH-01 (scoping only; no commands migrated in this plan).

## Depends on

Nothing.

## Parallelizable with

Plans 01, 02.

## Commands touched

None (this plan produces the list of commands the later waves touch).

## Context

@.planning/phases/01-cmderr-migration-completion/01-CONTEXT.md
@.planning/phases/01-cmderr-migration-completion/01-RESEARCH.md
@CLAUDE.md

## Tasks

### Task 1: Recompute the migration ledger

**Files:** `.planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md`

**Action:** Run the authoritative grep and produce the ledger.

```bash
# Step 1 — list all internal/cli subdirs
omni ls internal/cli/ > /tmp/cli-dirs.txt

# Step 2 — classify each dir
for d in $(omni ls internal/cli/); do
  if omni grep -rq 'cmderr\.' "internal/cli/$d/" 2>/dev/null; then
    echo "MIGRATED $d"
  else
    echo "TODO $d"
  fi
done > /tmp/migration-classification.txt

# Step 3 — hand-triage ambiguous dirs per CONTEXT Decision 1 exclusions
# (cmderr, command, input, safepath, timeutil, scaffolding-sublibs ARE NOT user-facing)
```

Produce `MIGRATION-LEDGER.md` with four sections:

```markdown
# Phase 1 Migration Ledger
Generated: <date>
Source: live git state (not RESEARCH.md snapshot)

## Already Migrated (no action)
<list of dirs where `grep -rq cmderr.` returned true AND coverage is adequate>

## Needs Migration — Wave Assignment
### Wave A (CI-critical)
<dirs assigned to Wave A with brief note>
### Wave B (data/format/encoding)
...
### Wave C (system/proc/info)
...
### Wave D (tail)
...

## Verify + Top-up (already partially migrated)
<dirs that have SOME cmderr calls but may have unclassified error returns>
- env — per Research Open Q2
- pipe — per Research Open Q2
- buf — per Research Open Q2
<any others discovered>

## Excluded (non-command helpers)
- cmderr, command, input, safepath, timeutil
- scaffolding sub-libs (cobra/handler/repository/testgen) — note: the `scaffold` subcommand IS user-facing
```

Use the ledger to reconcile wave sizes. If Wave A actually needs >6 plans (not 2), flag it and propose a re-split BEFORE moving on to execute Wave A plans.

**Verify:**
```
<automated>test -f .planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md && grep -c "^###" .planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md</automated>
```
Expect ≥4 (Wave A/B/C/D subsections).

**Done:** Ledger exists; every dir under `internal/cli/` is classified into exactly one of the four sections; Wave sizing is confirmed or re-split is proposed.

## Golden test additions

None.

## Verification

```bash
# Count classification coverage
omni wc -l .planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md
# Confirm no dir is unclassified
for d in $(omni ls internal/cli/); do
  omni grep -q "$d" .planning/phases/01-cmderr-migration-completion/MIGRATION-LEDGER.md || echo "MISSING: $d"
done
```

## Out of scope

- Actually migrating commands (Wave A+)
- Modifying CLAUDE.md's baseline list (done in Wave Z)
