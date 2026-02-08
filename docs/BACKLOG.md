# Backlog

Prioritized items for future development.

> Last updated: February 2026

---

## Completed

All core, file, text, system, archive, encoding, hashing, data, formatting, search, flow,
security, pager, comparison, and tooling commands are implemented (148+ commands).

Completed integrations: Kubernetes, Terraform, Vault, AWS (EC2/S3/IAM/STS/SSM).
Completed hacks: Git (12 shortcuts), Kubectl (17 shortcuts).
Completed engines: `pipe` (Cobra dispatch), `pipeline` (streaming io.Pipe stages).

See CLAUDE.md for the full command inventory.

---

## High Priority (P0)

### Core Infrastructure
- [ ] Unified `Command` interface contract
- [ ] Consistent error model with exit codes
- [ ] Add `--json` flag to remaining commands that lack it
- [ ] Unified output formatter (text/json/table)

---

## Medium Priority (P1)

### Data Formatting
- [ ] `yaml fmt` - YAML formatter/beautifier
  - Consistent indentation, key sorting, quote normalization
- [ ] `yaml k8s` - Kubernetes YAML formatter
  - Standard key ordering (apiVersion, kind, metadata, spec, status)
  - Multi-document support, remove empty fields

### GitHub Hacks
- [ ] `gh-pr-checkout` - Checkout PR by number
- [ ] `gh-pr-diff` - Show PR diff locally
- [ ] `gh-pr-approve` - Quick approve PR
- [ ] `gh-issue-mine` - List issues assigned to me
- [ ] `gh-repo-clone-org` - Clone all repos from org
- [ ] `gh-actions-rerun` - Rerun failed workflow

### Cloud & DevOps Integrations

#### Consul Integration
- [ ] `consul members` - List cluster members
- [ ] `consul kv` - Key-value store operations
- [ ] `consul services` - Service catalog operations

#### Nomad Integration
- [ ] `nomad job` - Job management
- [ ] `nomad node` - Node operations
- [ ] `nomad alloc` - Allocation operations

#### Packer Integration
- [ ] `packer build` - Build images
- [ ] `packer validate` - Validate templates
- [ ] `packer fmt` - Format templates

---

## Low Priority (P2)

### Advanced Features
- [ ] Plugin system
- [ ] WASM build target
- [ ] Embedded mode for other tools
- [ ] Filter DSL (`--where` conditions)

### AI Code Generation
- [ ] `ai generate` - AI assistant for code generation
  - Configurable system prompts, multiple providers
  - Context-aware code generation
  - Code review, explain, refactor modes

---

## Technical Debt

- [ ] Standardize flag naming across commands
- [ ] Add context.Context to all long-running operations
- [ ] Improve error messages with actionable suggestions
- [ ] Split large packages (archive.go ~500 lines)

---

## Testing

### Current Status (February 2026)
- **Total Test Cases:** ~700+ tests across all packages
- **Packages with Tests:** 86/88 (97.7%)
- **CLI Packages with Tests:** 79/79 (100%)
- **Packages without Tests:** 2 (twig/builder, twig/parser)

### Remaining
- [ ] Tests for twig/builder and twig/parser
- [ ] Platform-specific tests (Windows edge cases, symlinks, permissions)
- [ ] Large file (>1GB) handling tests
- [ ] Benchmarks vs GNU tools (sort, grep, file operations)
- [ ] Golden tests with expected output files
- [ ] CI coverage threshold enforcement (80%)

---

## Documentation

- [ ] Full command reference (man-page style)
- [ ] Library API documentation (godoc)
- [ ] Migration guide from shell scripts
- [ ] Taskfile integration examples

---

## CI/CD

### Done
- [x] GitHub Actions test workflow (lint, fmt, vulncheck, test -race)
- [x] Release workflow

### Remaining
- [ ] Multi-platform builds (Linux, macOS, Windows)
- [ ] Multi-arch builds (amd64, arm64)
- [ ] Automated releases with goreleaser
- [ ] Code coverage reporting + badges
- [ ] Coverage threshold enforcement

---

## Cross-Platform Notes

| Command | Linux | macOS | Windows | Notes |
|---------|:-----:|:-----:|:-------:|-------|
| `chmod` | ✅ | ✅ | ⚠️ | Limited permission model |
| `chown` | ✅ | ✅ | ❌ | Not applicable |
| `df` | ✅ | ✅ | ⚠️ | Different syscalls |
| `free` | ✅ | ⚠️ | ⚠️ | /proc vs sysctl |
| `ps` | ✅ | ✅ | ⚠️ | Different APIs |
| `ln -s` | ✅ | ✅ | ⚠️ | Requires admin |
| `kill` | ✅ | ✅ | ⚠️ | Signal handling differs |

---

## Version Milestones

### v0.1.0 - MVP ✅
Core commands, JSON output, basic docs

### v0.2.0 - File Operations ✅
cp, mv, rm, mkdir, stat, touch, ln, readlink, chmod, chown

### v0.3.0 - Text Processing ✅
grep, head, tail, sort, uniq, wc, cut, tr, sed, awk, rg, pipeline

### v0.4.0 - System Info ✅
df, du, env, whoami, uname, time, uptime, free, ps, kill, id, lsof

### v0.5.0 - Test Coverage ✅
97.7% package coverage (86/88), 700+ test cases

### v0.6.0 - Cloud & DevOps ✅
Kubernetes, Terraform, Vault, AWS, Git hacks, Kubectl hacks

### v0.7.0 - Engines & Media ✅
pipe (Cobra dispatch), pipeline (streaming), video (pure Go youtube-dl), buf (protobuf)

### v1.0.0 - Production Ready
- All P0 infrastructure items complete
- Full documentation
- 80%+ overall test coverage
- CI coverage enforcement
