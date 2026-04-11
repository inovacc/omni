---
phase: 01-cmderr-migration-completion
plan: 17
type: execute
wave: Z
depends_on: [12, 13, 14, 15]
files_modified:
  - CLAUDE.md
  - .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
  - docs/BACKLOG.md
autonomous: true
requirements: [POLISH-01]
must_haves:
  truths:
    - "CLAUDE.md's 'Error Handling (cmderr)' section reflects 100% adoption (the '~76 remaining' note is removed)"
    - "EXIT-CODE-CHANGES.md is finalized and ready for v1.0 release notes"
    - "Any deferred backlog items (exec spawn violations, docs/EXIT-CODES.md, etc.) are recorded in docs/BACKLOG.md"
  artifacts:
    - path: "CLAUDE.md"
      provides: "Updated cmderr adoption statement"
      contains: "100% adoption"
    - path: "docs/BACKLOG.md"
      provides: "Backlog rows for deferred items discovered during Phase 1"
      contains: "DEPRECATION"
  key_links:
    - from: "CLAUDE.md Error Handling section"
      to: "EXIT-CODE-CHANGES.md"
      via: "link in release-notes context"
      pattern: "EXIT-CODE-CHANGES"
---

# Plan 17 — Wave Z: Documentation update + backlog sync

## Goal

Update `CLAUDE.md` to reflect 100% cmderr adoption and record any deferred findings (exec-spawn violations, dropped sentinel proposals, etc.) in `docs/BACKLOG.md`. Finalize `EXIT-CODE-CHANGES.md` for v1.0 release notes consumption.

## Wave

Wave Z.

## Requirements covered

POLISH-01 (closure — documenting the completed state).

## Depends on

Plans 12, 13, 14, 15. Parallelizable with Plan 16.

## Parallelizable with

Plan 16.

## Commands touched

None.

## Context

@CLAUDE.md
@.planning/phases/01-cmderr-migration-completion/01-CONTEXT.md
@.planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md

## Tasks

### Task 1: Update CLAUDE.md "Error Handling (cmderr)" section

**Files:** `CLAUDE.md`

**Action:** Locate the section titled `#### Error Handling (cmderr)` in the project CLAUDE.md. Update:

1. Replace the line `**Commands adopted (84):** cat, curl, ... exist` with a statement of 100% adoption: `**Commands adopted (ALL):** every command in internal/cli/ returns classified cmderr sentinels.`
2. Remove the line `**Commands NOT yet adopted: ~76 remaining — adopt in future batches following the same pattern.**`.
3. Add a line: `**Exit-code contract:** v1.0 is the first stable exit-code contract. Changes from this point forward follow the CLAUDE.md breaking-change protocol.`
4. Add a link: `See .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md for the Phase 1 transition ledger.`

Do NOT touch other sections of CLAUDE.md.

**Verify:**
```
<automated>grep -c "Commands NOT yet adopted" CLAUDE.md</automated>
```
Must return 0.

**Done:** CLAUDE.md accurately reflects 100% migration; breaking-change protocol language present.

### Task 2: Finalize EXIT-CODE-CHANGES.md

**Files:** `.planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md`

**Action:**
1. Review every row appended by Waves A-D. Normalize format.
2. Add a closing summary header: `## Summary: N commands changed exit codes during Phase 1`.
3. Add a final section `## Release-notes template` with a paragraph suitable for dropping into v1.0 release notes verbatim.
4. This file becomes read-only after Plan 17 merges.

**Verify:**
```
<automated>grep -c "Release-notes template" .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md</automated>
```
Must return 1.

**Done:** File is publication-ready.

### Task 3: Add backlog rows for deferrals

**Files:** `docs/BACKLOG.md`

**Action:** Append any deferred items surfaced during Phase 1:

- `docs/EXIT-CODES.md` generation (deferred to Phase 3 per CONTEXT Deferred Ideas)
- `cmderr.Is<Class>()` convenience helpers (Phase 3)
- `golangci-lint` rule for raw `os.ErrX` returns (Phase 2)
- Cross-command exit-code golden matrix (Phase 2+)
- Any exec-spawn violations found during Plan 12 / Plan 14 audit (pre-existing CLAUDE.md violations, each a separate backlog row)
- Any missing sentinels discovered (Research §"Gaps Flagged" says none, but confirm)

Use the existing `docs/BACKLOG.md` format. Include dates per CLAUDE.md Breaking Changes & Deprecation section where applicable.

**Verify:**
```
<automated>grep -q "exit code" docs/BACKLOG.md && grep -q "Phase 1" docs/BACKLOG.md</automated>
```

**Done:** Backlog reflects all deferrals; no "we'll figure it out" items remain.

## Golden test additions

None.

## Verification

```bash
grep -c "Commands NOT yet adopted" CLAUDE.md      # expect 0
grep -c "Release-notes template" .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md   # expect 1
omni ls docs/BACKLOG.md        # exists
```

## Out of scope

- Writing the actual v1.0 release notes (Phase 8)
- Generating `docs/EXIT-CODES.md` (Phase 3)
- Fixing backlog items (each is its own future work)
