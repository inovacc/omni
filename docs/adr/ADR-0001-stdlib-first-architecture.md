# ADR-0001: Standard Library First Architecture

## Status

Accepted

## Date

2025 (project inception)

## Context

omni aims to replace common shell utilities (ls, grep, cat, sort, etc.) with Go-native implementations for use in Taskfile, CI/CD, and enterprise environments. The key question was how to implement 100+ commands: should we shell out to native tools, use heavy third-party libraries, or build from Go's standard library?

### Options Considered

1. **Exec-based wrappers** - Shell out to native tools (GNU coreutils, ripgrep, etc.)
2. **Third-party library heavy** - Use existing Go implementations for each command
3. **Standard library first** - Build from Go stdlib, add dependencies only when necessary

## Decision

We chose **standard library first** with the following rules:

1. **Never exec external processes** - All functionality is pure Go
2. **Prefer Go stdlib** - Use `os`, `io`, `path/filepath`, `regexp`, `encoding/*`, `crypto/*`, `archive/*`, `compress/*` before considering third-party packages
3. **Add dependencies only for capabilities stdlib cannot provide** - e.g., Cobra for CLI framework, bubbletea for TUI, goja for JavaScript runtime
4. **Library-first design** - Core logic in `pkg/` as importable packages, CLI wrappers in `internal/cli/`, thin Cobra commands in `cmd/`

## Consequences

### Positive

- **Cross-platform by default** - Go stdlib works identically on Linux, macOS, Windows
- **Single binary** - No runtime dependencies, no PATH issues, no version conflicts
- **Deterministic** - Same input always produces same output, unlike shell tools that vary by OS/version
- **Testable** - io.Writer pattern enables unit testing without filesystem or process spawning
- **Importable** - `pkg/` packages usable by external Go projects
- **Fast builds** - Minimal dependency tree (outside buf packages)

### Negative

- **Feature parity gaps** - Some commands (sed, awk) don't cover the full GNU feature set
- **Performance** - Pure Go implementations may be slower than optimized C tools for some operations
- **Maintenance burden** - We maintain the implementations rather than wrapping battle-tested tools
- **Large binary** - Includes all commands even if user only needs a few

### Dependencies Added (justified)

| Dependency | Justification |
|------------|--------------|
| `spf13/cobra` | CLI framework with subcommands, flags, completions - no stdlib equivalent |
| `charmbracelet/bubbletea` | TUI framework for less/more pagers - no stdlib equivalent |
| `golang.org/x/crypto` | PBKDF2 key derivation for encryption - extends stdlib crypto |
| `gopkg.in/yaml.v3` | YAML parsing for yq and lint - no stdlib YAML |
| `dop251/goja` | JavaScript runtime for YouTube signature decryption - unique requirement |
| `k8s.io/kubectl` | Kubernetes CLI - too complex to reimplement |
| `modernc.org/sqlite` | Pure Go SQLite - no CGO required |
| `go.etcd.io/bbolt` | BoltDB key-value store - specialized storage engine |

## Notes

This decision has proven effective across 120+ commands. The main cost is incomplete feature parity with GNU tools (particularly sed and awk), which is acceptable for the target use case of Taskfile and CI/CD environments where common patterns suffice.
