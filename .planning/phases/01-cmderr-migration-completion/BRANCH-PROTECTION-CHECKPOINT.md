# Branch Protection Checkpoint — Plan 18 (Wave Z)

**Status:** Automated work complete. One manual admin action required.

## What was automated

- `.github/workflows/test.yml` — `cmderr-coverage-gate` job hardened:
  - Added `needs: quality-check` (gate only runs after tests pass)
  - Added Phase 1 golden smoke step
  - Runs on `ubuntu-latest` only (cross-platform parser concern addressed)
- `.github/branch-protection.md` — created with required check list and CLI alternative

## Manual step required

GitHub branch protection requires admin credentials and cannot be automated
without a personal admin token scoped to `repo`.

### Option A — GitHub UI (recommended)

1. Visit https://github.com/inovacc/omni/settings/branches
2. Under "Branch protection rules" click **Edit** next to `main` (or create rule if absent)
3. Check "Require status checks to pass before merging"
4. Search for and add these checks:
   - `quality-check`
   - `cmderr-coverage-gate`
5. Check "Require branches to be up to date before merging"
6. Click **Save changes**

### Option B — GitHub CLI (requires admin token)

```bash
gh api -X PUT /repos/inovacc/omni/branches/main/protection \
  --field required_status_checks[strict]=true \
  --field 'required_status_checks[contexts][]=quality-check' \
  --field 'required_status_checks[contexts][]=cmderr-coverage-gate' \
  --field enforce_admins=false \
  --field required_pull_request_reviews=null \
  --field restrictions=null
```

## Smoke test verification

After enabling, confirm the gate works:

1. Create a test branch: `git checkout -b test/cmderr-gate`
2. Delete one test function from `internal/cli/cmderr/cmderr_test.go`
3. Push and open a PR
4. Confirm `cmderr-coverage-gate` check fails and the PR is blocked
5. Close/revert the test PR without merging

## Resume signal

Once the UI change is saved and smoke test confirms the gate blocks merge,
Phase 1 is fully complete. No further agent action is needed — just close
out by typing `enforced`.
