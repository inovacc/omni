# Deferred Items — Phase 02

## Pre-existing exec test failures

**Discovered during:** Plan 02-10, Step 2
**Package:** `internal/cli/exec`
**Tests:** TestDetectNpm_Missing, TestDetectKubectl_Present, TestDetectGo_PrivateWithNetrc
**Root cause:** Environment-dependent — tests check for absence of NPM_TOKEN env var, presence of kubeconfig, presence of .netrc. Fail on this machine because env does not match test assumptions.
**Status:** Pre-existing before Plan 02-10. Out of scope for Phase 2 coverage work.
**Suggested fix:** Add `t.Skip()` when required env/files are absent, or mock the file system lookups.
