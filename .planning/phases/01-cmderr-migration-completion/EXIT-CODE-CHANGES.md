# Exit-code changes introduced by Phase 1

Each wave appends rows documenting commands whose exit code changed from
un-classified (exit 1) to a classified cmderr sentinel. Used as input to
v1.0 release notes per CONTEXT.md Decision 6.

| Wave | Command | Before | After (sentinel → code) | Notes |
|------|---------|--------|-------------------------|-------|
| A | kill (no args) | 1 | ErrInvalidInput → 2 | Usage error now classified |
| A | kill (bad pid, e.g. `kill not-a-pid`) | 1 | ErrInvalidInput → 2 | Non-numeric pid |
| A | kill (unknown signal via `-s`) | 1 | ErrInvalidInput → 2 | e.g. `kill -s BOGUSX 1` |
| A | kill (nonexistent pid, e.g. `kill 999999999`) | 1 | ErrNotFound → 1 | Same exit; sentinel now set for downstream callers |
| A | kill (permission denied on `process.Signal`) | 1 | ErrPermission → 3 | Was unclassified; EPERM/os.ErrPermission now mapped |
| A | kill (windows, unsupported POSIX signal e.g. `-USR1`) | 1 | ErrUnsupported → 6 | Locked message: `kill: signal SIG%s not supported on windows (INT/KILL/TERM only)`; pins POLISH-09 format |
| A | sort (file not found) | 1 | ErrNotFound → 1 | No-op: already wrapped in `internal/cli/text/text.go` (RunSort). Ledger entry for `internal/cli/sort/` referred to a library-only helper with no CLI wiring. |
| A | sort (permission denied) | 1 | ErrPermission → 3 | Already wrapped in `text.RunSort`; exit code shifts from 1 to 3. |
| A | sort (`-c` disorder) | 1 | ErrConflict → 1 | Already wrapped in `text.RunSort`; exit code unchanged. |
| A | env (write failure, e.g. broken pipe) | 0 (silently muted) | ErrIO → 4 | Previously `_, _ = fmt.Fprint(...)` silently dropped write errors; now classified. |
| A | date (write failure, e.g. broken pipe) | 0 (silently muted) | ErrIO → 4 | Previously `_, _ = fmt.Fprintln(...)` silently dropped write errors; now classified. |
