# Known Issues & Limitations

> Last updated: February 2026

---

## Platform Limitations

### Windows

| Issue | Severity | Workaround |
|-------|----------|------------|
| `chmod` has limited permission model | Low | Windows ACLs don't map 1:1 to Unix permissions |
| `chown` not supported | Low | Not applicable on Windows |
| `ln -s` requires elevated privileges | Medium | Run as Administrator or enable Developer Mode |
| `kill` supports only INT, KILL, TERM signals | Low | Windows does not have POSIX signals |
| `free` uses different APIs (WMI) | Low | Output format matches Linux |
| `df` uses different syscalls | Low | Output format matches Linux |

### macOS

| Issue | Severity | Workaround |
|-------|----------|------------|
| `free` uses sysctl instead of /proc | Low | Transparent to user |

---

## Functional Limitations

### Text Processing
- `sed` supports basic substitution patterns but not the full GNU sed feature set (multi-line, hold space, branching)
- `awk` covers common patterns but does not implement the complete AWK language specification

### Video Download
- YouTube signature decryption depends on goja JS runtime; changes to YouTube's player JS may require updates
- HLS downloads do not support SAMPLE-AES encryption (only AES-128-CBC)
- No FFmpeg integration for format merging (video+audio must be in same container)

### Search (rg)
- Binary file detection uses heuristic (null byte check), may misclassify some files
- `.gitignore` pattern support covers common patterns but edge cases with nested negation may differ from ripgrep

### Database
- SQLite is pure Go (modernc.org/sqlite), slightly slower than CGO-based drivers for large datasets
- BBolt is limited to single-writer transactions

---

## Test Coverage Gaps

| Area | Issue | Priority |
|------|-------|----------|
| pkg/video/jsinterp | No unit tests | P2 |
| pkg/twig/builder, pkg/twig/parser | No tests | P2 |
| Overall coverage 30.5% (51.6% omni-owned) | Heavily skewed by vendored buf packages | P1 |
| cmderr adoption ~19/155+ commands | ~130 commands still return raw fmt.Errorf without exit code classification | P1 |

### Recently Resolved

| Area | Resolution |
|------|------------|
| pkg/video/downloader | Added progress, fragment, selector tests (Feb 2026) |
| pkg/video/nethttp | Added cookies, SAPISID hash tests (Feb 2026) |
| pkg/video/extractor | Added helpers, ParseM3U8Formats tests (Feb 2026) |
| pkg/video/options | Added applyOptions, With* option tests (Feb 2026) |

---

## Build & CI

| Issue | Severity | Notes |
|-------|----------|-------|
| `go test ./...` can be slow due to buf package compilation | Low | Use `-short` flag or test specific packages |
| No automated release pipeline | Medium | Manual `go build` for releases |
| No coverage threshold enforcement in CI | Medium | Target: 80% for omni-owned packages |
