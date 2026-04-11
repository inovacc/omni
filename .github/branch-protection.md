# Main branch protection — required checks

This project requires the following GitHub Actions checks to pass before
merging to `main`:

- `quality-check` — full unit + race test suite, lint, and vulnerability scan
- `cmderr-coverage-gate` — internal/cli/cmderr coverage >= 90% + Phase 1 golden smoke

To set: Settings -> Branches -> Branch protection rules -> main -> Require status
checks to pass -> select the above jobs.

## CLI alternative (requires admin token with `repo` scope)

```bash
gh api -X PUT /repos/inovacc/omni/branches/main/protection \
  --field required_status_checks[strict]=true \
  --field 'required_status_checks[contexts][]=quality-check' \
  --field 'required_status_checks[contexts][]=cmderr-coverage-gate' \
  --field enforce_admins=false \
  --field required_pull_request_reviews=null \
  --field restrictions=null
```

## Verification

After enabling, open a PR that lowers cmderr coverage (e.g., delete one test in
`internal/cli/cmderr/`) and confirm the `cmderr-coverage-gate` check fails and
blocks merge. Revert the test PR once confirmed.
