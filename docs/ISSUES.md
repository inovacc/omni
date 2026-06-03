# Known Issues & Limitations

> Last updated: 2026-06-03

This page tracks **open defects** only. Feature-completeness gaps, test-coverage
gaps, and CI hardening live in `docs/BACKLOG.md`. Intentional, permanent
platform/design tradeoffs are listed at the bottom under
[Platform & design notes (not defects)](#platform--design-notes-not-defects).

---

## Open Issues

_None._ No open defects are currently tracked. (Backlog/feature work →
`docs/BACKLOG.md`; design tradeoffs → below.)

---

## Platform & design notes (not defects)

These are intentional, permanent tradeoffs of omni's pure-Go, no-exec,
cross-platform design — not bugs and not on any fix track.

### Windows

| Note | Detail |
|------|--------|
| `chmod` has a limited permission model | Windows ACLs do not map 1:1 to Unix permissions |
| `chown` not supported | Not applicable on Windows |
| `ln -s` requires elevated privileges | Run as Administrator or enable Developer Mode |
| `kill` supports only INT, KILL, TERM signals | Windows has no POSIX signal set |
| `free` / `df` use different APIs (WMI/different syscalls) | Output format still matches Linux |

### macOS

| Note | Detail |
|------|--------|
| `free` uses sysctl instead of `/proc` | Transparent to the user; `/proc` does not exist on macOS |

### Database

| Note | Detail |
|------|--------|
| SQLite is pure Go (`modernc.org/sqlite`) | Slightly slower than CGO-based drivers for large datasets — accepted for the no-CGO invariant |
| BBolt single-writer | Limited to single-writer transactions by design |

### Build & test

| Note | Detail |
|------|--------|
| `go test ./...` can be slow | Driven by buf package compilation; use `-short` or test specific packages |
