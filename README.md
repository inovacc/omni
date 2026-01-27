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

### Implemented Commands (Phase 1)
- `ls`: List directory contents
- `pwd`: Print working directory
- `cat`: Concatenate and print files
- `date`: Print current date and time
- `dirname`: Strip last component from file name
- `basename`: Strip directory and suffix from file names
- `realpath`: Print the resolved path

## Usage
Build the project:
```bash
go build -o goshell main.go
```

Run a command:
```bash
./goshell ls
./goshell pwd
```

### Package Categories
The project is organized into functional packages within `pkg/`, following a library-first architecture:
- `pkg/fs`: Filesystem operations (ls, pwd, cat, chmod, path manipulation).
- `pkg/text`: Text processing (sort, uniq, grep, trimming).
- `pkg/timeutil`: Time and date formatting.
- `pkg/cli`: CLI-specific logic, argument handling, and output formatting.

See ROADMAP.md and BACKLOG.md.
# goshell
