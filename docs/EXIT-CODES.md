# omni Exit Codes

omni classifies every error through `internal/cli/cmderr` sentinels; `cmderr.ExitCodeFor(err)` (called in `cmd/root.go`) maps them to process exit codes. This page is the human-readable reference for that mapping.

| Exit code | Meaning | Sentinel(s) | `Is*` helper |
|-----------|---------|-------------|--------------|
| 0 | Success | — (`nil`) | — |
| 1 | Not found / conflict / unclassified | `ErrNotFound`, `ErrConflict`, any unclassified error | `IsNotFound`, `IsConflict` |
| 2 | Invalid input / usage | `ErrInvalidInput` | `IsInvalidInput` |
| 3 | Permission denied | `ErrPermission` | `IsPermission` |
| 4 | I/O error | `ErrIO` | `IsIO` |
| 5 | Timeout | `ErrTimeout` | `IsTimeout` |
| 6 | Unsupported operation | `ErrUnsupported` | `IsUnsupported` |

Notes:
- A `SilentError`/`ExitError` carries its own explicit code (see `cmderr.WithExitCode`).
- A recovered panic exits with the dedicated panic code set in `cmd/root.go` (`panicExitCode`).
- Any error not matching a sentinel falls through to exit code **1**.

Source of truth: `internal/cli/cmderr/cmderr.go` (`ExitCodeFor`). Keep this table in sync when sentinels change.
