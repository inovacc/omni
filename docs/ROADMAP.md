# goshell - Implementation Roadmap

## Overview

**goshell** is a cross-platform, Go-native replacement for common shell utilities. It provides deterministic, testable implementations designed for Taskfile, CI/CD, and enterprise environments.

### Design Principles

| Principle | Description |
|-----------|-------------|
| No exec | Never spawn external processes |
| Stdlib-first | Prefer Go standard library |
| Cross-platform | Linux, macOS, Windows |
| Library-first | All commands usable as Go packages |
| JSON output | Structured output for automation |
| Safe defaults | No dangerous operations without explicit flags |

---

## Phase 1 – Core (MVP) ✅ DONE

Foundation commands using Go standard library.

### Commands

| Command | Go Implementation | Flags | Status |
|---------|-------------------|-------|--------|
| `ls` | `os.ReadDir()` | `-l`, `-a`, `-1`, `--json` | ✅ Done |
| `pwd` | `os.Getwd()` | `--json` | ✅ Done |
| `cat` | `os.ReadFile()`, `bufio.Scanner` | `-n` (line numbers) | ✅ Done |
| `date` | `time.Now().Format()` | `--format`, `--utc`, `--json` | ✅ Done |
| `dirname` | `filepath.Dir()` | — | ✅ Done |
| `basename` | `filepath.Base()` | `-s` (suffix) | ✅ Done |
| `realpath` | `filepath.Abs()` + `filepath.EvalSymlinks()` | — | ✅ Done |

### Architecture

```
goshell/
├── cmd/                    # Cobra CLI commands
│   ├── root.go
│   ├── ls.go
│   ├── pwd.go
│   └── ...
├── pkg/cli/               # Library implementations
│   ├── ls.go
│   ├── pwd.go
│   └── ...
└── main.go
```

### Features

- [x] Cobra CLI framework
- [x] JSON output mode (partial)
- [x] Library-first architecture
- [x] Basic error handling

---

## Phase 2 – File Operations

File manipulation commands with safe defaults.

### Commands

| Command | Go Implementation | Flags | Priority |
|---------|-------------------|-------|----------|
| `cp` / `copy` | `io.Copy()` | `-r`, `-f`, `--dry-run` | P0 |
| `mv` / `move` | `os.Rename()` | `-f`, `--dry-run` | P0 |
| `rm` / `remove` | `os.Remove()`, `os.RemoveAll()` | `-r`, `-f`, `--dry-run` (**required**) | P0 |
| `rmdir` | `os.Remove()` | — | P1 |
| `mkdir` | `os.MkdirAll()` | `-p` | P0 |
| `ln` | `os.Symlink()`, `os.Link()` | `-s` (symlink) | P2 ✅ |
| `readlink` | `os.Readlink()` | `-f` | P2 ✅ |
| `stat` | `os.Stat()` | `--json` | P1 |
| `touch` | `os.OpenFile()` + close | `-a`, `-m` | P1 |
| `chmod` | `os.Chmod()` | — | P2 ✅ |
| `chown` | `os.Chown()` | `-R` | P2 ✅ |

### Safe rm Design

```go
type RMOptions struct {
    Recursive bool   // -r flag
    Force     bool   // -f flag
    DryRun    bool   // --dry-run (REQUIRED for destructive operations)
}

// rm without --dry-run or explicit confirmation = error
```

### Library API

```go
import "github.com/inovacc/goshell/pkg/fs"

fs.Copy("src", "dst", fs.CopyOptions{Recursive: true})
fs.Move("src", "dst")
fs.Remove("path", fs.RMOptions{DryRun: true})
fs.Mkdir("path", 0755)
fs.Stat("path")
```

---

## Phase 3 – Text Processing

Pattern matching and text manipulation.

### Commands

| Command | Go Implementation | Flags | Priority |
|---------|-------------------|-------|----------|
| `grep` | `strings.Contains()`, `regexp.MatchString()` | `-i`, `-v`, `-n`, `-c`, `-E` | P0 ✅ |
| `egrep` | `regexp.MatchString()` | (alias for `grep -E`) | P1 ✅ |
| `fgrep` | `strings.Contains()` | (alias for `grep -F`) | P1 ✅ |
| `head` | `bufio.Scanner` (N lines) | `-n`, `-c` | P0 ✅ |
| `tail` | Read from end | `-n`, `-c`, `-f` | P0 ✅ |
| `tac` | Reverse line order | — | P2 ✅ |
| `sort` | `sort.Strings()`, `sort.Slice()` | `-r`, `-n`, `-u` | P0 ✅ |
| `uniq` | `map[string]struct{}` | `-c`, `-d`, `-u` | P0 ✅ |
| `wc` | Count lines/words/bytes | `-l`, `-w`, `-c` | P0 ✅ |
| `nl` | Line numbering | `-b`, `-n` | P2 ✅ |
| `cut` | Field extraction | `-b`, `-c`, `-d`, `-f`, `-s`, `--complement` | P1 ✅ |
| `paste` | Merge lines | `-d` | P2 ✅ |
| `join` | Join sorted files | — | P3 ✅ |
| `fold` | Wrap lines | `-w` | P3 ✅ |
| `column` | Columnate lists | `-t`, `-s` | P2 ✅ |
| `tr` | Character translation | `-c`, `-d`, `-s`, `-t` | P1 ✅ |
| `sed` | Stream editor (basic) | `-e`, `-i` | P3 ✅ |
| `awk` | Pattern scanning (subset) | — | P3 ✅ |

### Grep Implementation

```go
type GrepOptions struct {
    IgnoreCase  bool   // -i
    InvertMatch bool   // -v
    LineNumbers bool   // -n
    Count       bool   // -c
    ExtendedRE  bool   // -E (use regexp)
}

func GrepWithOptions(lines []string, pattern string, opt GrepOptions) []string
```

### Pipeline Engine (Internal)

```go
// Replace: cat file | grep foo | sort | uniq
type Stage func([]string) ([]string, error)

out, err := pipeline.Run(
    readFile("file.txt"),
    grep("foo"),
    sortAsc(),
    uniq(),
)
```

---

## Phase 4 – System & Information

System inspection and process information.

### Commands

| Command | Go Implementation | Flags | Priority | Platform Notes |
|---------|-------------------|-------|----------|----------------|
| `env` | `os.Environ()`, `os.Getenv()` | `-0`, `-u`, `-i` | P0 ✅ | All |
| `whoami` | `os/user.Current()` | — | P0 ✅ | All |
| `id` | `os/user` package | `-u`, `-g`, `-G`, `-n`, `-r` | P1 ✅ | All |
| `uname` | `runtime.GOOS`, `runtime.GOARCH` | `-a`, `-s`, `-n`, `-r`, `-v`, `-m`, `-p`, `-i`, `-o` | P0 ✅ | All |
| `uptime` | `syscall` (platform-specific) | `-p`, `-s` | P2 ✅ | Linux/macOS/Windows |
| `time` | `time.Now()`, measure duration | — | P1 ✅ | All |
| `df` | `syscall.Statfs()` | `-H`, `-i`, `-B`, `--total`, `-t`, `-x`, `-l`, `-P` | P1 ✅ | Build tags |
| `du` | `filepath.Walk()` + `info.Size()` | `-a`, `-b`, `-c`, `-H`, `-s`, `-d`, `-x`, `-0`, `-B` | P1 ✅ | All |
| `free` | `/proc/meminfo` or `syscall` | `-b`, `-k`, `-m`, `-g`, `-H`, `-w`, `-t` | P2 ✅ | Linux/macOS/Windows |
| `ps` | `/proc` or Win32 API | `-a`, `-f`, `-l`, `-u`, `-p` | P3 ✅ | Linux/Windows |
| `gops` | `github.com/google/gops` | — | P3 | External dep |
| `top` | (Not planned - too complex) | — | — | — |
| `kill` | `os.Process.Signal()` | `-s`, `-l`, `-v` | P2 ✅ | All |

### Platform-Specific Files

```
pkg/sys/
├── df.go           # Interface
├── df_unix.go      # Linux, macOS, BSD
├── df_windows.go   # Windows implementation
└── df_test.go
```

### Build Tags Example

```go
//go:build unix
// +build unix

package sys

import "syscall"

func DF(path string) (*DiskUsage, error) {
    var stat syscall.Statfs_t
    if err := syscall.Statfs(path, &stat); err != nil {
        return nil, err
    }
    return &DiskUsage{
        Total: stat.Blocks * uint64(stat.Bsize),
        Free:  stat.Bavail * uint64(stat.Bsize),
    }, nil
}
```

---

## Phase 5 – Advanced Utilities & Flow

Concurrency, streaming, and flow control.

### Commands

| Command | Go Implementation | Flags | Priority |
|---------|-------------------|-------|----------|
| `xargs` | `goroutines` + `channels` | `-0`, `-d`, `-n`, `-P`, `-r`, `-t`, `-I` | P1 ✅ |
| `yes` | Infinite loop + context cancel | — | P2 ✅ |
| `nohup` | Signal handling + output redirect | — | P3 ✅ |
| `watch` | `time.Ticker` + file monitoring | `-n`, `-d`, `-t`, `-b`, `-e`, `-p`, `-c` | P1 ✅ |
| `less` | (TUI - consider `bubbletea`) | — | P3 |
| `more` | Simple pager | — | P3 |
| `pipeline` | Internal streaming engine | — | P0 |

### xargs Design (Safe)

```go
// Only executes internal goshell commands - NO external exec
type WorkerFunc func(arg string) error

type XargsOptions struct {
    Parallel int  // -P N
}

func Run(args []string, fn WorkerFunc, opt XargsOptions) error
```

Usage:
```bash
goshell find . .go | goshell xargs -P 4 stat
```

### Watch Design

```go
type WatchOptions struct {
    Interval  time.Duration
    OnChange  bool          // Use fsnotify instead of polling
    Command   string        // Internal goshell command to run
}

func Watch(path string, opt WatchOptions, fn func() error) error
```

---

## Phase 6 – Ecosystem & Tooling

Integration, documentation, and developer experience.

### Features

| Feature | Description | Priority |
|---------|-------------|----------|
| Taskfile linter | Validate Taskfile.yml uses portable commands | P1 |
| .env loader | Parse and load .env files | P1 ✅ |
| Config handling | Read JSON/YAML configs | P2 |
| Hash/checksum | `sha256sum`, `md5sum`, `sha512sum` | P1 ✅ |
| Archive | `tar`/`zip`/`unzip` using `archive/*` | P1 ✅ |
| Base encoding | `base64`, `base32`, `base58` | P1 ✅ |
| UUID | UUID v4 generation | P1 ✅ |
| JSON/YAML | `jq`, `yq` processors | P1 ✅ |
| Encryption | `encrypt`, `decrypt` (AES-256-GCM) | P1 ✅ |
| Random | Random numbers, strings, passwords | P1 ✅ |
| Diff | Text and JSON diff | P2 |
| Documentation | Full command reference + examples | P0 |
| Benchmarks | Compare vs GNU tools | P2 |

### Taskfile Linter

```bash
goshell lint Taskfile.yml
```

Checks:
- [ ] No shell-specific commands (rm, grep, etc.)
- [ ] All commands are goshell-compatible
- [ ] No `exec` or external process calls

### Archive Commands

```bash
goshell archive create out.zip ./dir
goshell archive list out.zip
goshell archive extract out.zip
```

### Hash Commands

```bash
goshell hash file.bin
goshell hash dir/ --recursive
goshell hash --verify checksums.txt
```

---

## Library API Summary

All commands are available as importable Go packages:

```go
import (
    "github.com/inovacc/goshell/pkg/fs"      // File system operations
    "github.com/inovacc/goshell/pkg/text"    // Text processing
    "github.com/inovacc/goshell/pkg/sys"     // System information
    "github.com/inovacc/goshell/pkg/hash"    // Hashing utilities
    "github.com/inovacc/goshell/pkg/archive" // Archive operations
)

// Examples
files, _ := fs.Ls(".", fs.LsOptions{All: true})
lines, _ := text.Grep(content, "pattern", text.GrepOptions{IgnoreCase: true})
usage, _ := sys.DF("/")
sum, _ := hash.SHA256File("file.bin")
```

---

## Output Modes

All commands support multiple output formats:

### Text (Default)
```bash
goshell ls
```

### JSON
```bash
goshell ls --json
```
```json
[
  {"name": "main.go", "size": 2312, "isDir": false, "mode": "rw-r--r--"}
]
```

### Implementation

```go
type OutputFormat int

const (
    FormatText OutputFormat = iota
    FormatJSON
)

func printOutput(cmd *cobra.Command, data any, format OutputFormat) error {
    if format == FormatJSON {
        return json.NewEncoder(cmd.OutOrStdout()).Encode(data)
    }
    // Text output...
}
```

---

## Testing Strategy

### Unit Tests
- Table-driven tests for all functions
- Edge cases: empty input, large files, special characters
- Platform-specific tests with build tags

### Integration Tests
- Compare output with GNU tools (where applicable)
- Test CLI flags and combinations
- Test JSON output parsing

### Benchmarks
```go
func BenchmarkSortGo(b *testing.B) {
    lines := generateLines(10_000)
    for i := 0; i < b.N; i++ {
        text.Sort(lines)
    }
}
```

### Golden Tests
```bash
# Generate expected output
goshell ls testdata/ > testdata/ls.golden

# Compare in tests
func TestLsGolden(t *testing.T) {
    // Compare actual vs golden file
}
```

---

## Priority Matrix

| Priority | Description | Examples |
|----------|-------------|----------|
| **P0** | Core functionality, MVP | `ls`, `cat`, `grep`, `cp`, `rm` |
| **P1** | Important utilities | `stat`, `head`, `tail`, `sort`, `uniq` |
| **P2** | Nice to have | `chmod`, `chown`, `sed`, `tr` |
| **P3** | Future/Complex | `awk`, `ps`, `top`, `nohup` |

---

## Cross-Platform Considerations

| Command | Linux | macOS | Windows | Notes |
|---------|-------|-------|---------|-------|
| `ls` | ✅ | ✅ | ✅ | |
| `chmod` | ✅ | ✅ | ⚠️ | Limited on Windows |
| `chown` | ✅ | ✅ | ❌ | Not applicable |
| `df` | ✅ | ✅ | ⚠️ | Different syscalls |
| `free` | ✅ | ⚠️ | ⚠️ | Platform-specific |
| `ps` | ✅ | ✅ | ⚠️ | Different APIs |
| `ln -s` | ✅ | ✅ | ⚠️ | Requires admin on Windows |

---

## Release Plan

### v0.1.0 - MVP
- [ ] Phase 1 commands complete
- [ ] JSON output mode
- [ ] Basic documentation
- [ ] CI/CD pipeline

### v0.2.0 - File Operations
- [ ] Phase 2 commands
- [ ] Safe rm with dry-run
- [ ] Improved error messages

### v0.3.0 - Text Processing
- [ ] Phase 3 commands
- [ ] Pipeline engine
- [ ] grep with regex support

### v0.4.0 - System Info
- [ ] Phase 4 commands
- [ ] Cross-platform df/du
- [ ] Process utilities

### v1.0.0 - Production Ready
- [ ] All phases complete
- [ ] Full documentation
- [ ] Taskfile linter
- [ ] Comprehensive tests

---

## Quick Start

```bash
# Build
go build -o goshell ./cmd/goshell

# Or run directly
go run . ls
go run . pwd --json
go run . date --format "2006-01-02"

# Use as library
import "github.com/inovacc/goshell/pkg/fs"
files, _ := fs.Ls(".", fs.LsOptions{})
```

---

## References

- Go standard library: https://pkg.go.dev/std
- Cobra CLI: https://github.com/spf13/cobra
- twig (tree replacement): https://github.com/inovacc/twig
- fsnotify: https://github.com/fsnotify/fsnotify

---

## Go stdlib Equivalents Reference

| Linux Command | Go stdlib |
|---------------|-----------|
| `cd` | `os.Chdir()` |
| `pwd` | `os.Getwd()` |
| `ls` | `os.ReadDir()` |
| `cat` | `os.ReadFile()`, `bufio.Scanner` |
| `stat` | `os.Stat()` |
| `realpath` | `filepath.Abs()` + `filepath.EvalSymlinks()` |
| `dirname` | `filepath.Dir()` |
| `basename` | `filepath.Base()` |
| `tree` | `filepath.WalkDir()` |
| `find` | `filepath.WalkDir()` + filter |
| `head` | `bufio.Scanner` (N lines) |
| `tail` | Read from end / `bufio.Scanner` |
| `wc` | Count bytes/lines/words |
| `sort` | `sort.Strings()`, `sort.Slice()` |
| `uniq` | `map[string]struct{}` |
| `grep` | `strings.Contains()`, `regexp` |
| `date` | `time.Now().Format()` |
| `uname` | `runtime.GOOS`, `runtime.GOARCH` |
| `whoami` | `os/user.Current()` |
| `env` | `os.Environ()`, `os.Getenv()` |
| `which` | `exec.LookPath()` (no exec) |
| `cp` | `io.Copy()` |
| `mv` | `os.Rename()` |
| `rm` | `os.Remove()`, `os.RemoveAll()` |
| `mkdir` | `os.MkdirAll()` |
| `touch` | `os.OpenFile()` + close |
| `chmod` | `os.Chmod()` |
| `ln` | `os.Symlink()`, `os.Link()` |
| `df` | `syscall.Statfs()` |
| `du` | `filepath.Walk()` + `info.Size()` |
| `sha256sum` | `crypto/sha256` |
| `tar` | `archive/tar` |
| `zip` | `archive/zip` |
| `gzip` | `compress/gzip` |
