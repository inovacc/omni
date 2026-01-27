# goshell

goshell is a cross-platform, safe, Go-native replacement for common shell utilities,
designed for Taskfile, CI/CD, and enterprise environments.

## Goals
- No exec / no external binaries
- 100% Go standard library (or well-justified small deps)
- Portable: Linux, macOS, Windows
- CLI + Library mode

## Why
Traditional shell commands (ls, grep, sort, uniq, date, etc) are:
- Not portable
- Hard to test
- Fragile in CI

goshell replaces them with deterministic Go implementations.

## Status
🚧 Early development

See ROADMAP.md and BACKLOG.md.
# goshell
