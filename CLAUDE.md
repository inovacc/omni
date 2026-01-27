# omni - Project Memory

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

1. **No exec** - Never spawn external processes
2. **Stdlib first** - Prefer Go standard library
3. **Cross-platform** - Linux, macOS, Windows support
4. **Library-first** - All commands usable as Go packages
5. **Safe defaults** - Destructive operations require explicit flags
6. **Testable** - io.Writer pattern for all output

---

## Architecture

### Directory Structure

```
omni/
├── cmd/           # Cobra CLI commands (thin wrappers)
├── pkg/cli/       # Library implementations (all logic here)
├── tests/         # Integration tests
├── docs/          # Documentation
└── main.go        # Entry point
```

### Code Patterns

#### Command Implementation Pattern

1. **Options struct** in `pkg/cli/`:
```go
type LsOptions struct {
    All       bool   // -a
    Long      bool   // -l
    Recursive bool   // -R
}
```

2. **Run function** with io.Writer:
```go
func RunLs(w io.Writer, args []string, opts LsOptions) error {
    // Implementation
    _, _ = fmt.Fprintln(w, output)
    return nil
}
```

3. **Cobra wrapper** in `cmd/`:
```go
var lsCmd = &cobra.Command{
    Use:   "ls [OPTION]... [FILE]...",
    Short: "List directory contents",
    RunE: func(cmd *cobra.Command, args []string) error {
        opts := cli.LsOptions{}
        opts.All, _ = cmd.Flags().GetBool("all")
        return cli.RunLs(os.Stdout, args, opts)
    },
}
```

#### Platform-Specific Code

Use build tags for platform-specific implementations:

```
pkg/cli/
├── df.go           # Interface + shared logic
├── df_unix.go      # //go:build unix
└── df_windows.go   # //go:build windows
```

#### Error Handling

- Always wrap errors with context: `fmt.Errorf("command: %w", err)`
- Write errors to stderr: `fmt.Fprintf(os.Stderr, "error: %v\n", err)`
- Return errors, don't panic

#### Output Patterns

- Mute unused return values: `_, _ = fmt.Fprintln(w, ...)`
- Use defers with anonymous functions:
```go
defer func() {
    _ = file.Close()
}()
```

---

## Command Categories

### Implemented (70+ commands)

| Category | Commands |
|----------|----------|
| **Core** | ls, pwd, cat, date, dirname, basename, realpath |
| **File** | cp, mv, rm, mkdir, rmdir, touch, stat, ln, readlink, chmod, chown |
| **Text** | grep, egrep, fgrep, head, tail, sort, uniq, wc, cut, tr, nl, paste, tac, column, fold, join, sed, awk |
| **System** | env, whoami, id, uname, uptime, free, df, du, ps, kill, time |
| **Flow** | xargs, watch, yes, nohup |
| **Archive** | tar, zip, unzip |
| **Hash** | hash, sha256sum, sha512sum, md5sum |
| **Encoding** | base64, base32, base58 |
| **Data** | jq, yq, dotenv |
| **Security** | encrypt, decrypt, uuid, random |
| **Pagers** | less, more |
| **Comparison** | diff |
| **Tooling** | lint |

### Backlog

| Command | Notes |
|---------|-------|
| `pipeline` | Internal streaming engine |

---

## Dependencies

### Direct Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `golang.org/x/crypto` | PBKDF2 for encryption |
| `github.com/charmbracelet/bubbletea` | TUI framework for pagers |
| `github.com/charmbracelet/lipgloss` | Terminal styling |
| `gopkg.in/yaml.v3` | YAML parsing for yq, lint |

### Standard Library Usage

| Package | Commands |
|---------|----------|
| `os`, `io`, `io/fs` | All file operations |
| `path/filepath` | Path manipulation |
| `regexp` | grep, sed pattern matching |
| `encoding/json` | jq, JSON output |
| `archive/tar`, `archive/zip` | tar, zip, unzip |
| `compress/gzip` | tar -z |
| `crypto/*` | hash, encrypt/decrypt |
| `syscall` | df, free, uptime, ps |

---

## Testing

### Run Tests

```bash
go test -race -cover ./...
```

### Test Pattern

```go
func TestLs(t *testing.T) {
    var buf bytes.Buffer
    err := cli.RunLs(&buf, []string{"."}, cli.LsOptions{})
    if err != nil {
        t.Fatal(err)
    }
    // Assert on buf.String()
}
```

---

## Build & Release

### Build

```bash
go build -ldflags="-s -w" -o omni .
```

### Cross-Compile

```bash
GOOS=linux GOARCH=amd64 go build -o omni-linux .
GOOS=darwin GOARCH=amd64 go build -o omni-darwin .
GOOS=windows GOARCH=amd64 go build -o omni.exe .
```

---

## Common Tasks

### Add New Command

1. Create `pkg/cli/newcmd.go` with Options struct and RunNewCmd function
2. Create `cmd/newcmd.go` with Cobra wrapper
3. Add to rootCmd in init()
4. Update docs/COMMANDS.md
5. Update docs/ROADMAP.md

### Add Platform-Specific Code

1. Create interface in main file
2. Create `_unix.go` and `_windows.go` variants
3. Use `//go:build unix` and `//go:build windows` tags

---

## Style Guide

- Follow project CLAUDE.md conventions (defer patterns, error muting)
- Use short receiver names (1-2 letters)
- Prefer table-driven tests
- Keep cmd/ files thin - all logic in pkg/cli/
- Always accept io.Writer for testability

---

## Links

- **Repo**: https://github.com/inovacc/omni
- **Taskfile**: https://taskfile.dev/
- **Cobra**: https://github.com/spf13/cobra
