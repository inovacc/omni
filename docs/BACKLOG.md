# Backlog

Prioritized items for future development phases.

---

## High Priority (P0)

### Core Infrastructure
- [ ] Define unified `Command` interface contract
- [ ] Implement consistent error model with exit codes
- [ ] Add `--json` flag to all commands
- [ ] Unified output formatter (text/json/table)

### File Operations (Phase 2)
- [ ] `cp` / `copy` - Copy files with `-r` recursive support
- [ ] `mv` / `move` - Move/rename files
- [ ] `rm` / `remove` - Safe remove with **mandatory** `--dry-run` for destructive ops
- [ ] `mkdir` - Create directories with `-p` parents

### Text Processing (Phase 3)
- [ ] `grep` - Pattern search with `-i`, `-v`, `-n` flags
- [ ] `head` / `tail` - First/last N lines
- [ ] `sort` - Sort lines with `-r`, `-n`, `-u` flags
- [ ] `uniq` - Remove duplicates with `-c` count
- [ ] `wc` - Word/line/byte count

### Pipeline Engine
- [ ] Internal pipeline without shell pipes
- [ ] Stage-based processing (`[]string -> []string`)
- [ ] Memory-efficient streaming for large files

---

## Medium Priority (P1)

### File Operations
- [ ] `stat` - File information with JSON output
- [ ] `touch` - Create/update file timestamps
- [ ] `rmdir` - Remove empty directories

### Text Processing
- [ ] `cut` - Field extraction with `-d` delimiter, `-f` fields
- [ ] `tr` - Character translation
- [ ] `nl` - Number lines

### System Information
- [ ] `df` - Disk free space (cross-platform with build tags)
- [ ] `du` - Directory size
- [ ] `time` - Measure command duration

### Advanced Utilities
- [ ] `xargs` - Parallel execution of internal commands
- [ ] `watch` - Periodic command execution with `-n` interval

### Ecosystem
- [ ] Taskfile linter - Validate portable commands
- [ ] `.env` loader - Parse and load environment files
- [ ] `hash` / `sha256sum` - File checksums
- [ ] `archive` - Create/extract tar/zip files

---

## Low Priority (P2)

### File Operations
- [ ] `ln` - Create symbolic/hard links
- [ ] `readlink` - Resolve symlinks
- [ ] `chmod` - Change file permissions
- [ ] `chown` - Change file ownership (Unix only)

### Text Processing
- [ ] `tac` - Reverse line order
- [ ] `column` - Columnate output
- [ ] `sed` - Basic stream editing (subset)
- [ ] `paste` - Merge lines from files

### System Information
- [ ] `uptime` - System uptime (Unix only)
- [ ] `free` - Memory usage (Linux only)
- [ ] `kill` - Send signals to processes
- [ ] `id` - User/group IDs

### Ecosystem
- [ ] `diff` - Text and JSON comparison
- [ ] Benchmarks vs GNU tools
- [ ] Filter DSL (`--where` conditions)

---

## Low Priority / Nice to Have (P3)

### Complex Commands
- [ ] `awk` - Pattern scanning (subset)
- [ ] `join` - Join sorted files
- [ ] `fold` - Wrap lines
- [ ] `ps` - Process listing
- [ ] `gops` - Go process inspection

### Advanced Features
- [ ] Plugin system
- [ ] WASM build target
- [ ] Embedded mode for other tools
- [ ] `less` / `more` - TUI pagers (consider bubbletea)
- [ ] `nohup` - Background execution (limited in Go)

---

## Technical Debt

- [ ] Refactor pkg/cli to proper pkg/* structure
- [ ] Add context.Context to all long-running operations
- [ ] Improve error messages with actionable suggestions
- [ ] Standardize flag naming across commands
- [ ] Add request/response logging for debugging

---

## Testing

### Unit Tests
- [ ] Table-driven tests for all pkg/* functions
- [ ] Edge cases: empty input, large files, special characters
- [ ] Platform-specific tests with build tags
- [ ] Target: 80% code coverage

### Integration Tests
- [ ] Compare output with GNU tools
- [ ] Test CLI flag combinations
- [ ] Test JSON output parsing
- [ ] E2E flow tests

### Benchmarks
- [ ] `sort` vs GNU sort
- [ ] `grep` vs GNU grep
- [ ] File operations vs native tools
- [ ] Memory usage profiling

### Golden Tests
- [ ] Generate expected output files
- [ ] Automated comparison in CI

---

## Documentation

- [ ] Full command reference (man-page style)
- [ ] Usage examples for each command
- [ ] Library API documentation (godoc)
- [ ] Migration guide from shell scripts
- [ ] Taskfile integration examples
- [ ] CI/CD integration guide

---

## CI/CD

- [ ] GitHub Actions workflow
- [ ] Multi-platform builds (Linux, macOS, Windows)
- [ ] Multi-arch builds (amd64, arm64)
- [ ] Automated releases with goreleaser
- [ ] Code coverage reporting
- [ ] Security scanning (govulncheck, gitleaks)

---

## Cross-Platform Notes

| Command | Linux | macOS | Windows | Implementation Notes |
|---------|:-----:|:-----:|:-------:|---------------------|
| `chmod` | ✅ | ✅ | ⚠️ | Limited permission model |
| `chown` | ✅ | ✅ | ❌ | Not applicable |
| `df` | ✅ | ✅ | ⚠️ | Different syscalls, needs build tags |
| `free` | ✅ | ⚠️ | ⚠️ | /proc on Linux, sysctl on macOS |
| `ps` | ✅ | ✅ | ⚠️ | Different APIs per platform |
| `ln -s` | ✅ | ✅ | ⚠️ | Requires admin privileges |
| `kill` | ✅ | ✅ | ⚠️ | Signal handling differs |

---

## Version Milestones

### v0.1.0 - MVP
- Core commands (ls, pwd, cat, date, dirname, basename, realpath)
- JSON output mode
- Basic documentation

### v0.2.0 - File Operations
- cp, mv, rm, mkdir
- Safe rm with dry-run
- stat, touch

### v0.3.0 - Text Processing
- grep, head, tail, sort, uniq, wc
- Pipeline engine
- cut, tr

### v0.4.0 - System Info
- df, du (cross-platform)
- env, whoami, uname
- time, uptime

### v1.0.0 - Production Ready
- All P0/P1 commands complete
- Full documentation
- Taskfile linter
- 80%+ test coverage
