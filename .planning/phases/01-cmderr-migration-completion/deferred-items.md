# Deferred Items — Phase 1

## Out-of-scope discoveries logged during plan execution

### exec.go spawn violation (discovered: Plan 14)

**File:** `internal/cli/exec/exec.go`
**Tag:** no-exec-violation
**Issue:** Package `exec` uses `os/exec` to spawn arbitrary external processes via `osexec.Command(command, args...)` — violates project design principle "No exec — Never spawn external processes"
**Discovered during:** Task 2 (exec migration), Plan 14
**Status:** Pre-existing, out of scope for cmderr migration. cmderr sentinels applied to user-facing error paths (empty command, strict mode abort, user abort) without touching the spawn logic itself.
**Suggested action:** Rewrite `exec` to use the omni unified command Registry for dispatching internal commands instead of spawning external processes.

---

### remote.go exec violation (discovered: Plan 13)

**File:** `internal/cli/repo/remote.go`
**Issue:** Uses `os/exec` to run `gh repo clone` and `git clone` — violates project design principle "No exec — Never spawn external processes"
**Discovered during:** Task 2 (repo migration), Plan 13
**Status:** Pre-existing, out of scope for cmderr migration
**Suggested action:** Implement pure-Go git clone (e.g., via `go-git`) to replace exec calls
