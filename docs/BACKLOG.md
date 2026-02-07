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

### Data Formatting
- [ ] `yaml fmt` - YAML formatter/beautifier
  - Consistent indentation (2 or 4 spaces)
  - Key sorting (alphabetical or custom)
  - Remove trailing whitespace
  - Normalize quotes (single/double)
  - JSON output mode (`--json`)
- [ ] `yaml k8s` - Kubernetes YAML formatter
  - Standard key ordering: apiVersion, kind, metadata, spec, status
  - Metadata ordering: name, namespace, labels, annotations
  - Remove empty fields and null values
  - Multi-document support (---)
  - Validate against common K8s conventions

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
- [ ] `lsof` - List open files (cross-platform)

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

## Cloud & DevOps Integrations (P1)

### Kubernetes Integration ✅ DONE
- [x] `kubectl` / `k` - Full kubectl via k8s.io/kubectl
- Source: `B:\shared\personal\repos\kubernetes\kubectl\pkg\cmd`
- All kubectl commands available: get, describe, logs, exec, apply, delete, rollout, scale, etc.

### Terraform Integration ✅ DONE
- Source: Subprocess wrapper (terraform binary required)
- Alias: `omni terraform` / `omni tf`
- [x] `terraform init` - Initialize working directory
- [x] `terraform plan` - Show execution plan
- [x] `terraform apply` - Apply changes
- [x] `terraform destroy` - Destroy infrastructure
- [x] `terraform state` - State management (list, show, mv, rm, pull, push)
- [x] `terraform workspace` - Workspace management (list, new, select, delete, show)
- [x] `terraform output` - Show outputs
- [x] `terraform validate` - Validate configuration
- [x] `terraform fmt` - Format configuration
- [x] Additional: import, taint, untaint, refresh, graph, console, providers, get, test, show, version

### Vault Integration ✅ DONE
- Library: `github.com/hashicorp/vault/api`
- [x] `vault status` - Server status
- [x] `vault login` - Authenticate (token, userpass, approle)
- [x] `vault read` - Read secrets
- [x] `vault write` - Write secrets
- [x] `vault list` - List secrets
- [x] `vault delete` - Delete secrets
- [x] `vault token` - Token operations (lookup, renew, revoke)
- [x] `vault kv` - KV v2 operations (get, put, delete, list, destroy, undelete, metadata)

### Consul Integration
- Source: `B:\shared\personal\repos\hashicorp\consul`
- [ ] `consul members` - List cluster members
- [ ] `consul kv` - Key-value store operations
- [ ] `consul services` - Service catalog operations

### Nomad Integration
- Source: `B:\shared\personal\repos\hashicorp\nomad`
- [ ] `nomad job` - Job management
- [ ] `nomad node` - Node operations
- [ ] `nomad alloc` - Allocation operations

### Packer Integration
- Source: `B:\shared\personal\repos\hashicorp\packer`
- [ ] `packer build` - Build images
- [ ] `packer validate` - Validate templates
- [ ] `packer fmt` - Format templates

---

## Hacks & Shortcuts (P1)

### Git Hacks ✅ DONE
- Commands: `omni git <subcommand>` or standalone aliases
- [x] `git quick-commit` / `gqc` - Stage all + commit with message
- [x] `git branch-clean` / `gbc` - Delete merged branches (with --dry-run)
- [x] `git undo` - Undo last commit (soft reset)
- [x] `git amend` - Amend without editing message
- [x] `git stash-staged` - Stash only staged changes
- [x] `git log-graph` / `lg` - Pretty log with graph
- [x] `git diff-words` - Word-level diff
- [x] `git blame-line` - Blame specific line range
- [x] Additional: `git status`, `git push`, `git pull-rebase`, `git fetch-all`

### GitHub Hacks
- [ ] `gh-pr-checkout` - Checkout PR by number
- [ ] `gh-pr-diff` - Show PR diff locally
- [ ] `gh-pr-approve` - Quick approve PR
- [ ] `gh-issue-mine` - List issues assigned to me
- [ ] `gh-repo-clone-org` - Clone all repos from org
- [ ] `gh-actions-rerun` - Rerun failed workflow

### Kubectl Hacks ✅ DONE
- Standalone commands with short names
- [x] `kga` - Get all resources in namespace (pods, svc, deploy, etc.)
- [x] `klf <pod>` - Follow logs with timestamp
- [x] `keb <pod>` - Exec into pod with bash (falls back to sh)
- [x] `kpf <target> <local:remote>` - Quick port forward
- [x] `kdp <selector>` - Delete pods by selector
- [x] `krr <deployment>` - Restart deployment
- [x] `kge` - Get events sorted by time
- [x] `ktp` - Top pods by resource usage
- [x] `ktn` - Top nodes by resource usage
- [x] `kcs [context]` - Context switcher (list or switch)
- [x] `kns [namespace]` - Namespace switcher (list or switch)
- [x] Additional: `kwp` (watch pods), `kscale`, `kdebug`, `kdrain`, `krun`, `kconfig`

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

### AI Code Generation
- [ ] `ai generate` - AI assistant for code generation with system prompts
  - Configurable system prompts for different code styles/patterns
  - Support for multiple AI providers (OpenAI, Anthropic, local models)
  - Template-based prompt management
  - Context-aware code generation (reads project structure)
  - Examples:
    ```bash
    omni ai generate handler --name UserHandler
    omni ai generate test ./pkg/user.go
    omni ai generate struct --from json api_response.json
    omni ai --prompt "Add error handling to this function" file.go
    ```
  - Features:
    - [ ] System prompt configuration (`~/.omni/ai-prompts/`)
    - [ ] Project-specific prompts (`.omni/prompts/`)
    - [ ] Code review mode (`omni ai review file.go`)
    - [ ] Explain code mode (`omni ai explain file.go:10-50`)
    - [ ] Refactor suggestions (`omni ai refactor --style clean file.go`)

---

## Technical Debt

- [ ] Refactor duplicated file reading patterns (see [REUSABILITY.md](REUSABILITY.md))
- [ ] Create `internal/cli/input` package for shared input handling
- [ ] Create `internal/cli/output` package for consistent JSON output
- [ ] Create `internal/cli/pipeline` package for text processing chains
- [ ] Add context.Context to all long-running operations
- [ ] Improve error messages with actionable suggestions
- [ ] Standardize flag naming across commands
- [ ] Add request/response logging for debugging
- [ ] Split large packages (archive.go ~500 lines)

---

## Testing

### Current Status (January 2026)
- **Total Test Cases:** ~700+ tests across all packages
- **Packages with Tests:** 86/88 (97.7%)
- **CLI Packages with Tests:** 79/79 (100%)
- **Packages without Tests:** 2 (twig/builder, twig/parser)
- **Recently Added Tests:**
  - Compression: archive (14), bzip2 (10), gzip (12), xz (14)
  - Tooling: lint (17), testcheck (8), echo (9)
  - Internal/twig: twig (25), models (17), formatter (14), scanner (15)
- **Average Coverage (tested packages):** ~90%

### Unit Tests - Completed ✅
- [x] Table-driven tests for core functions
- [x] Edge cases: empty input, large files, special characters
- [x] Unicode content handling tests
- [x] Binary file handling tests
- [x] Consistency tests (multiple calls = same result)
- [x] Output format tests (newlines, whitespace)

### Unit Tests - In Progress 🔄
- [ ] Platform-specific tests with build tags (Windows edge cases)
- [ ] Symlink handling tests across platforms
- [ ] Permission-related tests
- [ ] Large file (>1GB) handling tests

### Unit Tests - Pending ❌
- [ ] Tests for 2 uncovered packages (twig/builder, twig/parser)
- [ ] Error path coverage for edge cases
- [ ] Timeout/context cancellation tests
- [ ] Concurrent access tests

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

### Test Infrastructure
- [ ] Add coverage reporting to CI
- [ ] Add test result badges to README
- [ ] Set up coverage threshold enforcement (80%)
- [ ] Add mutation testing

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

### v0.1.0 - MVP ✅
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

### v0.5.0 - Test Coverage Milestone ✅ (Achieved)
**Goal:** Achieve 90%+ coverage for tested packages, 60%+ overall
**Result:** 97.7% package coverage (86/88 packages have tests)

| Milestone | Target | Status |
|-----------|--------|--------|
| 7.1 Core Coverage | 95% for core packages | ✅ Achieved |
| 7.2 Utility Coverage | 80% for utility packages | ✅ Achieved |
| 7.3 Uncovered Packages | 60% for remaining packages | ✅ Exceeded (97.7%) |
| 7.4 CLI Coverage | 100% for CLI packages | ✅ Achieved (79/79) |

**Completed:**
- [x] Expanded tests for 86 packages (~700+ test cases)
- [x] Fixed platform-specific test failures
- [x] Added edge case tests (unicode, binary, large files)
- [x] Added consistency and output format tests
- [x] Added compression package tests (archive, bzip2, gzip, xz)
- [x] Added lint package tests (17 tests)
- [x] Added testcheck command with tests
- [x] Added echo command with tests
- [x] Added twig module tests (twig, models, formatter, scanner, expander)

**Remaining:**
- [ ] Add tests for twig/builder (integration tests)
- [ ] Add tests for twig/parser (integration tests)
- [ ] Add platform-specific test variants
- [ ] Set up CI coverage enforcement

### v1.0.0 - Production Ready
- All P0/P1 commands complete
- Full documentation
- Taskfile linter
- 90%+ test coverage for core packages
- 80%+ overall test coverage
