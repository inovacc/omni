# Deferred Items — Phase 1

## Out-of-scope discoveries logged during plan execution

### remote.go exec violation (discovered: Plan 13)

**File:** `internal/cli/repo/remote.go`
**Issue:** Uses `os/exec` to run `gh repo clone` and `git clone` — violates project design principle "No exec — Never spawn external processes"
**Discovered during:** Task 2 (repo migration), Plan 13
**Status:** Pre-existing, out of scope for cmderr migration
**Suggested action:** Implement pure-Go git clone (e.g., via `go-git`) to replace exec calls
