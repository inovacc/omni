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
├── .github/workflows/  # CI/CD workflows
├── cmd/                # Cobra CLI commands (thin wrappers)
├── internal/cli/       # Library implementations (all logic here)
│   ├── <command>/      # Each command in its own package
│   │   ├── <command>.go
│   │   ├── <command>_test.go
│   │   ├── <command>_unix.go    # Unix-specific (optional)
│   │   └── <command>_windows.go # Windows-specific (optional)
├── tests/              # Integration tests
├── docs/               # Documentation
├── Taskfile.yml        # Task automation
└── main.go             # Entry point
```

### Code Patterns

#### Command Implementation Pattern

1. **Options struct** in `internal/cli/<command>/`:
```go
// internal/cli/ls/ls.go
type Options struct {
    All       bool   // -a
    Long      bool   // -l
    Recursive bool   // -R
}
```

2. **Run function** with io.Writer:
```go
func Run(w io.Writer, args []string, opts Options) error {
    // Implementation
    _, _ = fmt.Fprintln(w, output)
    return nil
}
```

3. **Cobra wrapper** in `cmd/`:
```go
// cmd/ls.go
var lsCmd = &cobra.Command{
    Use:   "ls [OPTION]... [FILE]...",
    Short: "List directory contents",
    RunE: func(cmd *cobra.Command, args []string) error {
        opts := ls.Options{}
        opts.All, _ = cmd.Flags().GetBool("all")
        return ls.Run(cmd.OutOrStdout(), args, opts)
    },
}
```

#### Platform-Specific Code

Use build tags for platform-specific implementations:

```
internal/cli/df/
├── df.go           # Interface + shared logic
├── df_unix.go      # //go:build unix
└── df_windows.go   # //go:build windows

internal/cli/kill/
├── kill.go         # Shared logic
├── kill_unix.go    # Unix signals (30 signals)
└── kill_windows.go # Windows signals (INT, KILL, TERM only)
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

### Implemented (120+ commands)

| Category | Commands |
|----------|----------|
| **Core** | ls, pwd, cat, date, dirname, basename, realpath, tree, arch, sleep, seq, printf, for |
| **File** | cp, mv, rm, mkdir, rmdir, touch, stat, ln, readlink, chmod, chown, find, dd, file, which |
| **Text** | grep, egrep, fgrep, head, tail, sort, uniq, wc, cut, tr, nl, paste, tac, column, fold, join, sed, awk, shuf, split, rev, comm, cmp, strings |
| **System** | env, whoami, id, uname, uptime, free, df, du, ps, kill, time |
| **Flow** | xargs, watch, yes, pipe |
| **Archive** | tar, zip, unzip |
| **Compression** | gzip, gunzip, zcat, bzip2, bunzip2, bzcat, xz, unxz, xzcat |
| **Hash** | hash, sha256sum, sha512sum, md5sum |
| **Encoding** | base64, base32, base58, url encode/decode, html encode/decode, hex encode/decode, xxd |
| **Data** | jq, yq, dotenv, json (tostruct, tocsv, fromcsv, toxml, fromxml), yaml tostruct, yaml validate, toml validate, xml (validate, tojson, fromjson) |
| **Formatting** | sql fmt/minify/validate, html fmt/minify/validate, css fmt/minify/validate |
| **Protobuf** | buf lint, buf format, buf compile, buf breaking, buf generate, buf mod init/update, buf ls-files |
| **Code Gen** | generate handler, generate repository, generate test |
| **Security** | encrypt, decrypt, uuid, random, jwt decode |
| **Pagers** | less, more |
| **Comparison** | diff |
| **Tooling** | lint, cmdtree, loc, cron |
| **Network** | curl |

### Backlog

| Command | Notes |
|---------|-------|
| `pipeline` | Internal streaming engine |

### Pipe Command (Variable Substitution)

The `pipe` command chains omni commands together with variable substitution support:

```go
type Options struct {
    JSON      bool   // --json: output pipeline result as JSON
    Separator string // --sep: command separator (default "|")
    Verbose   bool   // --verbose: show intermediate steps
    VarName   string // --var: variable name for output substitution (default "OUT")
}
```

**Variable Patterns:**
- `$OUT` or `${OUT}` - Single value substitution (uses last non-empty line)
- `[$OUT...]` - Iteration: execute command for each line of output

**Examples:**
```bash
# Generate UUID and create folder
omni pipe '{uuid -v 7}' '{mkdir $OUT}'

# Custom variable name
omni pipe --var UUID '{uuid -v 7}' '{mkdir $UUID}'

# Iteration: create folder for each UUID
omni pipe '{uuid -v 7 -n 10}' '{mkdir [$OUT...]}'

# Chain with verbose output
omni pipe -v '{cat file.txt}' '{grep pattern}' '{sort}'
```

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
| `encoding/json` | jq, JSON output, json tocsv/fromcsv |
| `encoding/csv` | csv/json conversions |
| `encoding/xml` | xml operations, json toxml/fromxml |
| `archive/tar`, `archive/zip` | tar, zip, unzip |
| `compress/gzip` | tar -z, gzip, gunzip, zcat |
| `compress/bzip2` | bzip2, bunzip2, bzcat (decompress only) |
| `crypto/*` | hash, encrypt/decrypt |
| `syscall` | df, free, uptime, ps |
| `go/parser`, `go/ast` | generate test |
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

### Run Tests

```bash
# Run all tests with race detection and coverage
go test -race -cover ./...

# Run specific package tests
go test -race -v ./internal/cli/pipe/...

# Run specific test
go test -race ./internal/cli/... -run TestSubstituteVariables -v

# Generate coverage report
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run integration tests
task test:integration
```

### Test Coverage

Current coverage: ~26% (internal/cli)

### Test Files

| File | Tests |
|------|-------|
| `internal/cli/hash/hash_test.go` | SHA256, SHA512, MD5 hashing |
| `internal/cli/base/base_test.go` | Base64, Base32, Base58 encoding |
| `internal/cli/uuid/uuid_test.go` | UUID generation |
| `internal/cli/random/random_test.go` | Random string, hex, password, int |
| `internal/cli/fs/fs_test.go` | mkdir, cp, touch, stat, ln, readlink |
| `internal/cli/text/text_test.go` | tr, cut, nl, uniq, paste, tac, fold, column, join |
| `internal/cli/kill/kill_test.go` | Signal listing, PID validation |
| `internal/cli/jq/jq_test.go` | JSON querying |
| `internal/cli/yq/yq_test.go` | YAML querying |
| `internal/cli/diff/diff_test.go` | File comparison |
| `internal/cli/crypt/crypt_test.go` | Encrypt/decrypt |
| `internal/cli/path/path_test.go` | dirname, basename, realpath |
| `internal/cli/pipe/pipe_test.go` | Command parsing, variable substitution |
| `internal/cli/xxd/xxd_test.go` | Hex dump, reverse, plain, include, bits modes |

### Test Pattern

```go
func TestRun(t *testing.T) {
    var buf bytes.Buffer
    err := ls.Run(&buf, []string{"."}, ls.Options{})
    if err != nil {
        t.Fatal(err)
    }
    // Assert on buf.String()
}
```

### Linter Directives

Use `//nolint` for intentional patterns:

```go
//nolint:dupword // intentional duplicate content for testing uniq
content := "a\na\na\nb\nb\n"
```

---

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
5. Update docs/ROADMAP.md
6. Update CLAUDE.md command categories

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
