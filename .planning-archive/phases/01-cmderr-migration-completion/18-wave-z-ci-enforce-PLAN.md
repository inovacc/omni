---
phase: 01-cmderr-migration-completion
plan: 18
type: execute
wave: Z
depends_on: [16, 17]
files_modified:
  - .github/workflows/test.yml
  - .github/branch-protection.md
  - CLAUDE.md
autonomous: false
requirements: [POLISH-03]
must_haves:
  truths:
    - "task lint:cmderr-coverage is a required CI check blocking merge to main"
    - "A PR that drops cmderr coverage below 90% fails and cannot merge"
    - "The full phase success criteria from ROADMAP Phase 1 are verifiable from CI alone"
  artifacts:
    - path: ".github/workflows/test.yml"
      provides: "coverage gate wired as required job with needs: test"
      contains: "lint:cmderr-coverage"
    - path: ".github/branch-protection.md"
      provides: "Documentation of required checks for main branch protection"
      contains: "cmderr coverage gate"
  key_links:
    - from: "branch-protection.md"
      to: ".github/workflows/test.yml"
      via: "list of required job names"
      pattern: "cmderr coverage gate"
---

# Plan 18 — Wave Z: CI Enforcement of the Coverage Gate

## Goal

Promote `task lint:cmderr-coverage` from "step that runs" (Plan 02) to a **required** check on `main` branch protection, so a PR that drops cmderr coverage below 90% cannot merge. Also run the phase-wide smoke test: every Phase 1 success criterion is verifiable from CI.

This is the Phase 1 exit gate. Includes a human checkpoint because GitHub branch protection requires an admin action.

## Wave

Wave Z (final plan of Phase 1).

## Requirements covered

POLISH-03 (enforcement tightening).

## Depends on

Plans 16 (risky matrices green), 17 (docs updated).

## Parallelizable with

Nothing — this is the last plan.

## Commands touched

None.

## Context

@.planning/phases/01-cmderr-migration-completion/01-RESEARCH.md
@.github/workflows/test.yml
@CLAUDE.md

## Tasks

### Task 1: Harden the workflow step

**Files:** `.github/workflows/test.yml`

**Action:**
1. Ensure the `cmderr coverage gate` step added in Plan 02 has a stable job name (e.g. `cmderr-coverage-gate`) so branch-protection rules can reference it.
2. Add `needs: test` so the gate only runs after the main test job passes (avoids noise if tests already fail).
3. Confirm the step runs only on `ubuntu-latest` (Linux runner, per Research Open Q3 — awk-based parser is not cross-platform).
4. Add a phase-1 smoke test step that runs all Phase 1 golden filters in one shot:
   ```yaml
   - name: Phase 1 golden smoke
     run: task test:golden -- --filter '_file_not_found|_invalid_|_unsupported|_conflict'
   ```

**Verify:**
```
<automated>grep -q 'cmderr-coverage-gate\\|lint:cmderr-coverage' .github/workflows/test.yml</automated>
```

**Done:** Workflow has a stable job name referencing the gate; CI run succeeds green.

### Task 2: Document required checks (checkpoint prep)

**Files:** `.github/branch-protection.md` (create if not present)

**Action:** Create or update `.github/branch-protection.md` listing the required status checks for `main`:

```markdown
# Main branch protection — required checks

This project requires the following GitHub Actions checks to pass before
merging to `main`:

- `test` — full unit + race test suite
- `cmderr-coverage-gate` — internal/cli/cmderr coverage ≥ 90%
- `golden-smoke` — Phase 1 error-path golden subset

To set: Settings → Branches → Branch protection rules → main → Require status
checks to pass → select the above jobs.
```

**Verify:**
```
<automated>test -f .github/branch-protection.md && grep -q "cmderr-coverage-gate" .github/branch-protection.md</automated>
```

**Done:** Documentation file exists.

### Task 3: Human checkpoint — enable branch protection in GitHub UI

**Type:** `checkpoint:human-action`

**What's built:** The workflow job and docs are in place; the required-check selection is the single step Claude cannot perform (it requires admin credentials on the GitHub repo settings page).

**How to verify:**
1. Visit `https://github.com/inovacc/omni/settings/branches`.
2. Under "Branch protection rules" → `main` → Edit.
3. In "Require status checks to pass before merging", add:
   - `test`
   - `cmderr-coverage-gate` (or whatever name Plan 18 Task 1 chose)
   - `golden-smoke`
4. Save.
5. Open a trivial PR that lowers cmderr coverage (delete one test) and confirm the gate blocks merge.
6. Close/revert the test PR.

**Resume signal:** Type `enforced` once the UI change is saved and the smoke-test PR confirms the gate blocks merge.

**Why human-required:** GitHub branch-protection edits use the GitHub API under the hood, and `gh api` can technically script this — but it requires admin token scopes that vary by user setup. If the user prefers the CLI route, they can run `gh api -X PATCH /repos/inovacc/omni/branches/main/protection -F required_status_checks[contexts][]=cmderr-coverage-gate ...` themselves. Claude will offer the command as an alternative during the checkpoint.

### Task 4: Phase 1 closure smoke test

**Files:** none — verification only.

**Action:** After branch protection is live, run the phase-wide closure check locally:

```bash
# Every command classified
for d in $(omni ls internal/cli/); do
  if ! omni grep -rq 'cmderr' "internal/cli/$d/" 2>/dev/null; then
    # filter CONTEXT-excluded helpers
    case "$d" in cmderr|command|input|safepath|timeutil) ;; *) echo "UNMIGRATED: $d" ;; esac
  fi
done

# Coverage gate green
task lint:cmderr-coverage

# All goldens green
task test:golden

# No exit-code changes unreported
omni wc -l .planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md
```

**Verify:**
```
<automated>task lint:cmderr-coverage && task test:golden</automated>
```

Expect: both pass; closure check produces no `UNMIGRATED:` output.

**Done:** Phase 1 success criteria from ROADMAP §"Phase 1 Success Criteria" are all verifiable from CI and locally.

## Golden test additions

None.

## Verification

```bash
task lint:cmderr-coverage         # required CI check
task test:golden                  # all goldens
grep -c UNMIGRATED <(for d in $(omni ls internal/cli/); do ! omni grep -rq cmderr "internal/cli/$d/" 2>/dev/null && echo "UNMIGRATED: $d"; done)   # must be 0 (after helper exclusion)
```

## Out of scope

- Phase 2 coverage targets (POLISH-04/05)
- POLISH-09 full Windows parity implementation (Phase 3)
- Cross-command exit-code matrix (Phase 2+)
