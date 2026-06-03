# omni - Project Memory

**Detail-lookup convention:** This CLAUDE.md is intentionally lean. For details on any topic — phase histories, schema, migrations, deprecations, subsystem internals — search the `docs/` tree before assuming or asking. The **Topical docs** index below points to the right file.

## Topical docs

| Topic | File |
|-------|------|
| Full command inventory (170+) | `docs/COMMANDS.md` |
| Architecture patterns (cmderr, Command interface, platform split, output) | `docs/architecture/patterns.md` |
| Cloud & DevOps integrations (kubectl/tf/aws/git/kubectl hacks) | `docs/architecture/cloud-integrations.md` |
| Testing reference (unit/golden/integration/coverage) | `docs/architecture/testing.md` |
| High-level architecture diagrams | `docs/ARCHITECTURE.md` |
| Roadmap / backlog / features / issues / bugs | `docs/{ROADMAP,BACKLOG,FEATURES,ISSUES,BUGS}.md` |
| External source attribution | `docs/EXTERNAL_SOURCES.md` |
| Third-party licenses for ported code | `THIRD_PARTY_LICENSES/` |


## Project Overview

**omni** is a cross-platform, Go-native shell utility replacement. It provides deterministic, testable implementations of common Unix commands for use in Taskfile, CI/CD, and enterprise environments.

### Branding

| Item | Value |
|------|-------|
| **Name** | omni |
| **Package** | `github.com/inovacc/omni` |
| **Binary** | `omni` |
| **Tagline** | "Shell utilities, rewritten in Go" |

### Design Principles

1. **No exec** - Never spawn external processes (scope + sanctioned wrapper exceptions: see `docs/architecture/patterns.md` § "No-exec invariant: scope & sanctioned exceptions")
2. **Stdlib first** - Prefer Go standard library
3. **Cross-platform** - Linux, macOS, Windows support
4. **Library-first** - All commands usable as Go packages
5. **Safe defaults** - Destructive operations require explicit flags
6. **Testable** - io.Writer pattern for all output

---

## Command Categories

The full command inventory — 170+ commands across Core/File/Text/System/Process/Flow/Archive/Hash/Encoding/Data/Formatting/Protobuf/Code Gen/Security/Pagers/Comparison/Tooling/Network/Video/Cloud-DevOps/Git Hacks/Checks/Kubectl Hacks — lives in **`docs/COMMANDS.md`** (also browsable as `omni cmdtree`).

Run `omni cmdtree` for the live tree, or `omni <verb> --help` for any verb's flags.

## Cloud & DevOps Integrations

Kubernetes (`omni kubectl` / `omni k`), Terraform (`omni tf`), AWS SDK, Git hacks (12), Kubectl hacks (17) — full subcommand tables and source-replace directives are in **`docs/architecture/cloud-integrations.md`**.

## Dependencies

### Direct Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `golang.org/x/crypto` | PBKDF2 for encryption |
| `github.com/charmbracelet/bubbletea` | TUI framework for pagers |
| `github.com/charmbracelet/lipgloss` | Terminal styling |
| `gopkg.in/yaml.v3` | YAML parsing for yq, lint |
| `github.com/dop251/goja` | Pure Go JS runtime (YouTube signature decryption) |
| `github.com/spf13/afero` | Filesystem abstraction for testable scaffolding |
| `github.com/bufbuild/protocompile` | Pure Go protobuf compiler (AST parser for buf format/lint) |

### Standard Library Usage

| Package | Commands |
|---------|----------|
| `os`, `io`, `io/fs` | All file operations |
| `path/filepath` | Path manipulation |
| `regexp` | grep, sed, rg pattern matching |
| `sync`, `sync/atomic` | rg parallel walking, tree parallel scanning |
| `encoding/json` | jq, JSON output, json tocsv/fromcsv |
| `encoding/csv` | csv/json conversions |
| `encoding/xml` | xml operations, json toxml/fromxml |
| `archive/tar`, `archive/zip` | tar, zip, unzip |
| `compress/gzip` | tar -z, gzip, gunzip, zcat |
| `compress/bzip2` | bzip2, bunzip2, bzcat (decompress only) |
| `crypto/*` | hash, encrypt/decrypt |
| `syscall` | df, free, uptime, ps |
| `go/parser`, `go/ast` | scaffold test |
| `text/template` | code generators |

---

## CI/CD

### GitHub Actions

The project uses GitHub Actions for continuous integration:

- **Workflow**: `.github/workflows/test.yml`
- **Triggers**: Push/PR to non-main branches
- **Checks**:
  - `golangci-lint` - Code quality
  - `gofmt` - Code formatting
  - `govulncheck` - Security vulnerabilities
  - `go test -race` - Unit tests with race detection

### Running CI Locally

```bash
task lint      # Run golangci-lint
task test      # Run tests with coverage
task build     # Build binary
```

---

## Testing

Test commands, coverage targets, golden master tests, Docker test environments, and per-package test catalogs live in **`docs/architecture/testing.md`**.

Quick reference: `go test -race -cover ./...` for unit; `task test:golden` for golden masters; `task test:integration` for the Python black-box suite.

## Build

### Using Taskfile

```bash
# Build for current platform
task build

# Build for all platforms (Linux, Windows)
task build:all
```

### Manual Build

```bash
go build -ldflags="-s -w" -o omni .
```

### Cross-Compile

```bash
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o omni-linux .
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o omni-darwin .
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o omni.exe .
```

---

## Common Tasks

### Add New Command

1. Create `internal/cli/newcmd/newcmd.go` with Options struct and Run function
2. Create `internal/cli/newcmd/newcmd_test.go` with tests
3. Create `cmd/newcmd.go` with Cobra wrapper
4. Add to rootCmd in init()
5. Add golden test cases to `testing/golden/golden_tests.yaml` and `tools/golden/golden_tests.yaml`
6. Run `task test:golden:update` and `task golden:record` to generate snapshots
7. Update docs/ROADMAP.md
8. Update CLAUDE.md command categories

### Add Platform-Specific Code

1. Create interface in main file
2. Create `_unix.go` and `_windows.go` variants
3. Use `//go:build unix` and `//go:build windows` tags

---

## Style Guide

- Follow project CLAUDE.md conventions (defer patterns, error muting)
- Use short receiver names (1-2 letters)
- Prefer table-driven tests
- Keep cmd/ files thin - all logic in internal/cli/
- Always accept io.Writer for testability
- Use `cmd.OutOrStdout()` in Cobra commands for proper output capture
- Use `cmd.InOrStdin()` for stdin in commands that need piped input

---

## Links

- **Repo**: https://github.com/inovacc/omni
- **Taskfile**: https://taskfile.dev/
- **Cobra**: https://github.com/spf13/cobra

<!-- GSD:project-start source:PROJECT.md -->
## Project

**omni**

omni is a cross-platform, Go-native shell utility replacement providing deterministic, testable implementations of 160+ Unix commands for use in Taskfile, CI/CD, and enterprise environments. Today it ships as a single binary plus a growing set of reusable `pkg/` libraries. The primary user is me and my CI/CD pipelines — broader open-source adoption is a welcome bonus, not the driver.

**Core Value:** **One static binary replaces every shell utility a Go-based CI/CD pipeline needs — deterministically, on every OS, with no external processes spawned.** If everything else fails, this must remain true.

### Constraints

- **Tech stack**: Go (stdlib-first, Cobra CLI, pure-Go deps only) — deterministic, portable, no CGO where avoidable. No exec of external processes is a foundational rule.
- **Timeline**: 3–4 months to v1.0 — polish → supply chain → release, in that order. No aggressive parallelism across tracks.
- **Cross-platform**: Linux, macOS, Windows must all work. Platform-specific code uses build tags, never silent runtime branches.
- **Breaking changes**: Follow CLAUDE.md breaking-change protocol — 30-day deprecation window, log warnings on deprecated paths, cleanup commit separate from feature commits.
- **Security**: No committed secrets (`~/.claude/scripts/check-leaks.sh`), parameterized queries, bcrypt ≥ 10, distroless base for any containers. Pre-v1.0 cannot introduce new security footguns.
- **Licensing**: BSD 3-Clause, already in place.
- **Audience**: Design decisions serve me + CI/CD pipelines first. "Would a general open-source user want X?" is never a v1.0 prioritization input.
- **No new external processes**: The "no exec" design principle is non-negotiable — new capabilities must be pure Go.
<!-- GSD:project-end -->

<!-- GSD:stack-start source:codebase/STACK.md -->
## Technology Stack

## Languages
- Go 1.25.0 - All core logic, CLI commands, libraries, and utilities
## Runtime
- Go runtime (no external runtime dependencies)
- Cross-platform: Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64, arm64)
- Go Modules (`go.mod`, `go.sum`)
- Lockfile: `go.sum` present
## Frameworks
- Cobra v1.10.2 - Command-line interface framework (located in `cmd/`)
- Go testing package (`testing` stdlib)
- Table-driven test patterns throughout
- Charmbracelet Bubbletea v1.3.10 - Terminal UI framework (used by pagers)
- Charmbracelet Lipgloss v1.1.0 - Terminal styling and ANSI formatting
- GoReleaser v2 - Cross-platform binary building and releases (config: `.goreleaser.yaml`)
- Task v3 - Task automation (config: `Taskfile.yml`)
## Key Dependencies
- `github.com/spf13/cobra` v1.10.2 - CLI command framework (all commands)
- `github.com/charmbracelet/bubbletea` v1.3.10 - TUI pagers (less, more)
- `github.com/dop251/goja` v0.0.0-20260106131823 - Pure Go JavaScript runtime (YouTube signature decryption in video package)
- `gopkg.in/yaml.v3` v3.0.1 - YAML parsing (yq, dotenv, config parsing)
- `github.com/aws/aws-sdk-go-v2` v1.41.1 - AWS SDK (with services: EC2, IAM, S3, SSM, STS)
- `github.com/hashicorp/vault/api` v1.22.0 - HashiCorp Vault client (secret management)
- `k8s.io/kubectl` v0.35.0 - Kubernetes kubectl client (with `k8s.io/cli-runtime`)
- `go.etcd.io/bbolt` v1.4.3 - Pure Go key-value store (embedded database)
- `modernc.org/sqlite` v1.44.3 - Pure Go SQLite driver (channel DB for video downloader, SQLite shell)
- `golang.org/x/crypto` v0.48.0 - Stdlib crypto extensions (PBKDF2 for encryption)
- `github.com/btcsuite/btcd/btcutil` v1.1.6 - Bitcoin utilities (Base58 encoding)
- `github.com/bufbuild/protocompile` v0.14.1 - Pure Go protobuf compiler (buf format/lint)
- `google.golang.org/protobuf` v1.36.11 - Protocol Buffers runtime
- `github.com/segmentio/ksuid` v1.0.4 - Sortable unique IDs (KSUID generation)
- `github.com/spf13/afero` v1.15.0 - Filesystem abstraction (testable file operations in scaffolding)
- Standard `text/template` - Code generation templates
- `github.com/shirou/gopsutil/v3` v3.24.5 - OS/process utilities (ps, df, du, free, uptime)
- `github.com/google/gops` v0.3.29 - Go process analysis
- `github.com/BurntSushi/toml` v1.6.0 - TOML parsing
- `github.com/fatih/color` v1.18.0 - Colored terminal output
- `github.com/xlab/treeprint` v1.2.0 - Tree printing (legacy, replaced by `pkg/twig`)
- `os`, `io`, `io/fs` - File operations
- `path/filepath` - Path manipulation
- `regexp` - Pattern matching (grep, sed, rg)
- `sync`, `sync/atomic` - Concurrency (parallel scanning in rg, tree)
- `encoding/json` - JSON operations (jq, output)
- `encoding/xml` - XML parsing/output
- `encoding/csv` - CSV operations
- `archive/tar`, `archive/zip` - Archive operations
- `compress/gzip`, `compress/bzip2` - Compression
- `crypto/*` - Hashing, encryption
- `syscall` - System calls (for df, ps, etc.)
- `go/parser`, `go/ast` - Go code analysis (scaffold testgen)
## Configuration
- `VAULT_ADDR` - Vault server address (default: `https://127.0.0.1:8200`)
- `VAULT_TOKEN` - Vault authentication token
- `VAULT_NAMESPACE` - Vault namespace
- `AWS_PROFILE` - AWS profile selection
- `AWS_REGION` - AWS region
- `OMNI_CLOUD_PROFILE` - Custom omni cloud profile (alternative to `--profile omni:name`)
- `CGO_ENABLED` - Enables/disables CGO (set to 0 for cross-platform builds)
- `.golangci.yml` - Linting configuration (golangci-lint)
- `.goreleaser.yaml` - Release binary building
- `Taskfile.yml` - Task automation for build, test, golden tests
- `docker/docker-compose.test.yml` - Containerized test environments
- `.github/workflows/test.yml` - GitHub Actions test pipeline
- `.github/workflows/release.yml` - Release pipeline (GoReleaser)
## Platform Requirements
- Go 1.25.0 or later
- Task 3.x for local automation
- Docker (optional, for containerized tests)
- Python 3.8+ (for black-box tests)
- Target: Linux, macOS, Windows
- Architecture: x86_64 (amd64), ARM64 (arm64)
- No external runtime dependencies (static binaries via CGO_ENABLED=0)
- Supports AWS SDK (for aws commands)
- Supports kubectl (for k8s commands)
- Supports HashiCorp Vault (optional, for vault commands)
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

## Naming Patterns
- Command packages: `internal/cli/<command>/` (lowercase, single word or hyphenated)
- Library packages: `pkg/<domain>/` (e.g., `pkg/video/`, `pkg/twig/`)
- Test files: `<name>_test.go` (always suffix test, never prefix)
- Platform-specific: `<name>_unix.go`, `<name>_windows.go` with `//go:build` tags
- Options structs: `<Command>Options` (PascalCase, e.g., `HeadOptions`, `HashOptions`)
- Result structs: `<Command>Result` or `<Commands>Result` for plurals (e.g., `HashResult`, `HashesResult`)
- Exported functions: `Run<Command>` or `Run` (e.g., `RunHead()`, `RunHash()`)
- Unexported helpers: lowercase with underscores (e.g., `verifyChecksums()`, `parseCommands()`)
- All exported functions: PascalCase starting with action verb (e.g., `RunHash`, `WithIndent`, `FormatBytes`)
- CLI wrappers in `cmd/`: Cobra command variables named `<command>Cmd` (e.g., `headCmd`, `lsCmd`)
- Package init functions: Standard `init()` (for flag preprocessing, command registration)
- Receiver names: 1-2 letters (e.g., `func (f *Formatter) Format()`)
- Options: CamelCase (e.g., `tmpDir`, `testFile`, `showHeaders`)
- Boolean flags: prefix with action or state (e.g., `wantError`, `isJSON`, `showHeaders`)
- Package-level: all lowercase (e.g., `defaultHeadLines`, `omniBanner`)
- Constants: ALL_CAPS (e.g., `DefaultHeadLines`, `ErrNotFound`)
- Structs: PascalCase (e.g., `HashOptions`, `ColorScheme`, `StreamMessage`)
- Interfaces: PascalCase (e.g., `Command`, `Extractor`, `Formatter`)
- Sentinel errors: PascalCase prefixed with `Err` (e.g., `ErrNotFound`, `ErrInvalidInput`)
## Code Style
- Tool: `gofmt` (Go standard)
- Linter: `golangci-lint` with `govet` enabled
- Configuration: `.golangci.yml` with minimal rules (excludes generated code, relaxed comment checks)
- No custom formatter config detected
- Enabled: `govet` only (strict mode)
- Excluded: generated code in `pkg/buf`
- Run: `golangci-lint run --fix ./... --timeout=5m`
## Import Organization
## Error Handling
| Sentinel | Use For |
|----------|---------|
| `cmderr.ErrNotFound` | File/resource not found (exit code 1) |
| `cmderr.ErrInvalidInput` | Bad flags, missing operands, parse errors (exit code 2) |
| `cmderr.ErrPermission` | Permission denied (exit code 3) |
| `cmderr.ErrIO` | I/O errors (exit code 4) |
| `cmderr.ErrTimeout` | Timeouts (exit code 5) |
| `cmderr.ErrUnsupported` | Unsupported operations (exit code 6) |
| `cmderr.ErrConflict` | Verification failures, sort disorder (exit code 1) |
## Logging
- Used only for debug/execution tracking, not in command logic
- Commands accept `io.Writer` for stdout (testable)
- Errors written to stderr explicitly
## Comments
- Package-level docstrings (required for all packages)
- Exported function/type documentation (required by golangci-lint)
- Non-obvious logic or workarounds
- Complex algorithms or edge cases
## Function Design
- Max 3-4 parameters (use Options struct for many config values)
- Order: context/writer first, then inputs, then options
- Example: `func RunHead(w io.Writer, r io.Reader, args []string, opts HeadOptions) error`
- Always return `error` as last value
- For result data, use struct (e.g., `HashResult`) with JSON tags
- Multiple returns only for (value, error) pairs
## Module Design
- Options structs
- Result structs
- Run functions
- High-level public APIs
- Example: `pkg/video/types/types.go` exports `VideoInfo`, `Format`, etc.
## Output Patterns
## Platform-Specific Code
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

Details — directory layout, command implementation pattern, args preprocessor pattern, pkg library pattern, unified `Command` interface, platform-specific layout, error-handling (cmderr) sentinel table, output patterns — live in **`docs/architecture/patterns.md`**.

TL;DR: thin Cobra wrappers in `cmd/` → I/O glue in `internal/cli/<command>/` → pure-Go libraries in `pkg/`. Errors are classified via `internal/cli/cmderr` sentinels mapped to exit codes by `cmd/root.go`. Platform-specific code uses build tags (single-file or split layout — see [Platform-Specific Code](#platform-specific-code) below).

## Pattern Overview
- Thin CLI command wrappers in `cmd/` that delegate to reusable packages
- Core logic abstracted into `pkg/` libraries importable by external projects
- I/O glue layer in `internal/cli/` that bridges Cobra and pkg logic
- Unified Command interface for registry-based dispatch
- Error classification via sentinel values with mapped exit codes
- All commands accept `io.Writer` and `io.Reader` for testability
## Layers
- Purpose: Cobra CLI command definitions and flag parsing
- Location: `cmd/*.go` (160+ command files)
- Contains: Command definitions, flag configuration, args preprocessing
- Depends on: `internal/cli/<command>` packages for business logic
- Used by: Cobra CLI framework (main entry point)
- Pattern: Commands are thin wrappers that parse flags and delegate to `internal/cli/<cmd>/Run()` functions
- Purpose: Handle input/output operations, orchestrate stdin/stdout, manage file I/O
- Location: `internal/cli/<command>/` (160+ subdirectories, one per command)
- Contains: Options structs, Run functions with `(io.Writer, io.Reader, []string)` signatures
- Depends on: `pkg/` libraries for core logic, `internal/cli/cmderr` for error handling
- Used by: `cmd/` wrappers and pipe command
- Pattern: Each command has a Run function that coordinates I/O, calls pkg logic, and formats output
- Purpose: Reusable, testable command implementations
- Location: `pkg/<domain>/` (18+ subdirectories)
- Contains: Core algorithms (hashing, encryption, encoding, formatting), data structures, processing engines
- Depends on: Go stdlib only (no Cobra, no io.Writer output)
- Used by: `internal/cli/` packages and external Go projects
- Pattern: Pure functions with options using functional option pattern (`WithFlag(value) Option`)
- Purpose: Cross-cutting concerns and utilities
- Location: `internal/cli/cmderr/`, `internal/cli/command/`, `internal/cli/input/`, `pkg/cobra/`
- Contains: Error handling, command registry, input source management, output formatting
## Data Flow
- Commands return errors instead of calling `os.Exit()`
- Errors are classified via sentinel values in `internal/cli/cmderr/`
- `cmderr.ExitCodeFor(err)` in root command maps errors to exit codes:
## Key Abstractions
- Purpose: Unified interface for command dispatch in pipe/pipeline
- Location: `internal/cli/command/command.go`
- Pattern: `Command interface { Run(ctx, w, io.Writer, r io.Reader, args []string) error }`
- Adapters: `AdaptWriterArgs()`, `AdaptWriterReaderArgs()`, `AdaptFull()` for legacy Run signatures
- Registry: Thread-safe map of command name → Command implementation
- Purpose: Flexible, type-safe command configuration
- Example: `head.HeadOptions{Lines: 10, Quiet: false}`
- pkg pattern: `func (opt Option) Apply(*Options)` for functional options
- Cobra pattern: Parse flags → populate Options struct → pass to Run function
- Purpose: Unified handling of stdin, files, and multiple input sources
- Location: `internal/cli/input/input.go`
- Pattern: `input.Open(args, r)` returns `[]Source` where Source has Name and Reader
- Handles: `-` (stdin), multiple files, missing files with appropriate error classification
- Purpose: Flexible output (text/JSON/table) without duplicating logic
- Location: `pkg/cobra/helper/output/`
- Pattern: `output.New(w, format)` → methods like `Render()`, `IsJSON()`
- Usage: Commands check `f.IsJSON()` to decide between text and JSON output
- Purpose: Add context while preserving sentinel error type
- Location: `internal/cli/cmderr/cmderr.go`
- Pattern: `cmderr.Wrap(sentinel, "context message")` returns `fmt.Errorf("context: %w", sentinel)`
- Effect: `errors.Is(err, ErrNotFound)` still matches wrapped errors
## Entry Points
- Location: `main.go`
- Triggers: `omni` binary execution
- Responsibilities: Delegates to `cmd.Execute()`
- Location: `cmd/root.go`
- Triggers: All subcommands route through here
- Responsibilities:
- Location: `cmd/<cmd>.go init()` functions
- Purpose: Rewrite `os.Args` before Cobra parsing for GNU compatibility
- Examples:
## Cross-Cutting Concerns
- Framework: `log/slog` (structured JSON)
- Location: `internal/logger/`
- Pattern: Initialized in `cmd/root.go` PersistentPreRun
- Output: Only when `OMNI_LOGGING` env var set
- Pattern: Input validation in `internal/cli/<cmd>/` before calling pkg logic
- Example: `head` validates Lines/Bytes are non-negative
- Error type: `cmderr.ErrInvalidInput` for bad arguments
- Cloud commands (AWS, kubectl, vault): Use environment-based auth
- Location: `internal/cli/aws/`, `internal/cli/kubectl/`, `internal/cli/vault/`
- Pattern: Load config from `~/.aws/credentials`, `~/.kube/config`, `~/.vault-token`
- Pattern: Use `pkg/twig/scanner` for directory walking with parallelism and progress
- Pattern: Use `internal/cli/input` for unified input source handling
- Pattern: Always defer file closes with `defer func() { _ = f.Close() }()`
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->
## Project Skills

No project skills found. Add skills to any of: `.claude/skills/`, `.agents/skills/`, `.cursor/skills/`, or `.github/skills/` with a `SKILL.md` index file.
<!-- GSD:skills-end -->

<!-- superpowers:workflow-start -->
## Superpowers Workflow

Before starting any feature, fix, or phase work, follow this sequence:

1. **Brainstorm** — use `/superpowers:brainstorm` to explore intent, requirements, and design before touching code.
2. **Spec** — write or update a spec in `docs/superpowers/specs/` that captures the what and why.
3. **Plan** — use `/superpowers:writing-plans` to break the spec into an executable plan with clear tasks.
4. **Execute with TDD** — implement each task test-first; use `/superpowers:executing-plans` to drive execution.
5. **Verify** — use `/superpowers:verification-before-completion` to confirm all acceptance criteria pass before marking work done.

Use these entry points:
- `/superpowers:brainstorm` before any creative or design work
- `/superpowers:writing-plans` when you have a spec or requirements for a multi-step task
- `/superpowers:executing-plans` to drive plan execution
- `/superpowers:verification-before-completion` before declaring a phase or task complete
<!-- superpowers:workflow-end -->

<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->

# context-mode — MANDATORY routing rules

You have context-mode MCP tools available. These rules are NOT optional — they protect your context window from flooding. A single unrouted command can dump 56 KB into context and waste the entire session.

## BLOCKED commands — do NOT attempt these

### curl / wget — BLOCKED
Any Bash command containing `curl` or `wget` is intercepted and replaced with an error message. Do NOT retry.
Instead use:
- `ctx_fetch_and_index(url, source)` to fetch and index web pages
- `ctx_execute(language: "javascript", code: "const r = await fetch(...)")` to run HTTP calls in sandbox

### Inline HTTP — BLOCKED
Any Bash command containing `fetch('http`, `requests.get(`, `requests.post(`, `http.get(`, or `http.request(` is intercepted and replaced with an error message. Do NOT retry with Bash.
Instead use:
- `ctx_execute(language, code)` to run HTTP calls in sandbox — only stdout enters context

### WebFetch — BLOCKED
WebFetch calls are denied entirely. The URL is extracted and you are told to use `ctx_fetch_and_index` instead.
Instead use:
- `ctx_fetch_and_index(url, source)` then `ctx_search(queries)` to query the indexed content

## REDIRECTED tools — use sandbox equivalents

### Bash (>20 lines output)
Bash is ONLY for: `git`, `mkdir`, `rm`, `mv`, `cd`, `ls`, `npm install`, `pip install`, and other short-output commands.
For everything else, use:
- `ctx_batch_execute(commands, queries)` — run multiple commands + search in ONE call
- `ctx_execute(language: "shell", code: "...")` — run in sandbox, only stdout enters context

### Read (for analysis)
If you are reading a file to **Edit** it → Read is correct (Edit needs content in context).
If you are reading to **analyze, explore, or summarize** → use `ctx_execute_file(path, language, code)` instead. Only your printed summary enters context. The raw file content stays in the sandbox.

### Grep (large results)
Grep results can flood context. Use `ctx_execute(language: "shell", code: "grep ...")` to run searches in sandbox. Only your printed summary enters context.

## Tool selection hierarchy

1. **GATHER**: `ctx_batch_execute(commands, queries)` — Primary tool. Runs all commands, auto-indexes output, returns search results. ONE call replaces 30+ individual calls.
2. **FOLLOW-UP**: `ctx_search(queries: ["q1", "q2", ...])` — Query indexed content. Pass ALL questions as array in ONE call.
3. **PROCESSING**: `ctx_execute(language, code)` | `ctx_execute_file(path, language, code)` — Sandbox execution. Only stdout enters context.
4. **WEB**: `ctx_fetch_and_index(url, source)` then `ctx_search(queries)` — Fetch, chunk, index, query. Raw HTML never enters context.
5. **INDEX**: `ctx_index(content, source)` — Store content in FTS5 knowledge base for later search.

## Subagent routing

When spawning subagents (Agent/Task tool), the routing block is automatically injected into their prompt. Bash-type subagents are upgraded to general-purpose so they have access to MCP tools. You do NOT need to manually instruct subagents about context-mode.

## Output constraints

- Keep responses under 500 words.
- Write artifacts (code, configs, PRDs) to FILES — never return them as inline text. Return only: file path + 1-line description.
- When indexing content, use descriptive source labels so others can `ctx_search(source: "label")` later.

## ctx commands

| Command | Action |
|---------|--------|
| `ctx stats` | Call the `ctx_stats` MCP tool and display the full output verbatim |
| `ctx doctor` | Call the `ctx_doctor` MCP tool, run the returned shell command, display as checklist |
| `ctx upgrade` | Call the `ctx_upgrade` MCP tool, run the returned shell command, display as checklist |
