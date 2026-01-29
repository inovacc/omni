# omni - Implementation Roadmap

## Overview

**omni** is a cross-platform, Go-native replacement for common shell utilities. It provides deterministic, testable implementations designed for Taskfile, CI/CD, and enterprise environments.

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
omni/
├── cmd/                    # Cobra CLI commands (98 commands)
│   ├── root.go
│   ├── ls.go
│   ├── pwd.go
│   └── ...
├── internal/
│   ├── cli/               # Library implementations (79 packages)
│   │   ├── ls/
│   │   ├── pwd/
│   │   └── ...
│   ├── flags/             # Feature flags system
│   ├── logger/            # KSUID-based logging
│   └── twig/              # Tree visualization module
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
import "github.com/inovacc/omni/internal/cli/fs"

fs.Copy("src", "dst", &fs.CopyOptions{Recursive: true})
fs.Move("src", "dst", &fs.MoveOptions{})
fs.Remove("path", &fs.RmOptions{DryRun: true})
fs.Mkdir("path", 0755)
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
// Only executes internal omni commands - NO external exec
type WorkerFunc func(arg string) error

type XargsOptions struct {
    Parallel int  // -P N
}

func Run(args []string, fn WorkerFunc, opt XargsOptions) error
```

Usage:
```bash
omni find . .go | omni xargs -P 4 stat
```

### Watch Design

```go
type WatchOptions struct {
    Interval  time.Duration
    OnChange  bool          // Use fsnotify instead of polling
    Command   string        // Internal omni command to run
}

func Watch(path string, opt WatchOptions, fn func() error) error
```

---

## Phase 6 – Ecosystem & Tooling

Integration, documentation, and developer experience.

### Features

| Feature | Description | Priority | Status |
|---------|-------------|----------|--------|
| Taskfile linter | Validate Taskfile.yml uses portable commands | P1 | |
| .env loader | Parse and load .env files | P1 | ✅ |
| Config handling | Read JSON/YAML configs | P2 | |
| Hash/checksum | `sha256sum`, `md5sum`, `sha512sum` | P1 | ✅ |
| Archive | `tar`/`zip`/`unzip` using `archive/*` | P1 | ✅ |
| Base encoding | `base64`, `base32`, `base58` | P1 | ✅ |
| UUID | UUID v4 generation | P1 | ✅ |
| JSON/YAML | `jq`, `yq` processors | P1 | ✅ |
| Encryption | `encrypt`, `decrypt` (AES-256-GCM) | P1 | ✅ |
| Random | Random numbers, strings, passwords | P1 | ✅ |
| Diff | Text and JSON diff | P2 | ✅ |
| Documentation | Full command reference + examples | P0 | |
| Benchmarks | Compare vs GNU tools | P2 | |
| Test coverage check | List packages with/without tests | P1 | |
| Lua runner | Execute Lua scripts natively | P2 | |
| Python runner | Execute Python scripts natively | P2 | |

### New Features (Added January 2026)

| Feature | Description | Priority | Status |
|---------|-------------|----------|--------|
| ID generators | ksuid, ulid, uuid7, nanoid, Snowflake ID | P0 | |
| JSON beautify | JSON formatter and minifier | P0 | |
| Enhanced jq | Replace jq with gojq for full compatibility | P1 | |
| Template render | Render Go/JSON/YAML templates | P1 | |
| Unified output | text/json/table output formatter | P0 | |
| --json flag | JSON output for all commands | P0 | ✅ Mostly done |

### ID Generation Commands

```bash
omni ksuid              # Generate KSUID
omni ulid               # Generate ULID
omni uuid -v 4          # UUID v4 (existing)
omni uuid -v 7          # UUID v7 (time-ordered)
omni nanoid             # Generate NanoID
omni nanoid -l 10       # Custom length
omni snowflake          # Generate Snowflake ID
```

### JSON Beautify/Minify

```bash
omni json fmt file.json           # Beautify JSON
omni json minify file.json        # Minify JSON
omni json validate file.json      # Validate JSON
echo '{"a":1}' | omni json fmt    # From stdin
```

### Template Rendering

```bash
omni template render template.tmpl data.json
omni template render template.tmpl data.yaml
omni template render -e 'NAME=World' 'Hello {{.NAME}}'
```

### Test Coverage Check

```bash
# List all packages and their test status
omni testcheck ./pkg/cli/
```

Output:
```
basename: HAS TEST
cat: HAS TEST
chmod: HAS TEST
chown: NO TEST
...
```

Implementation using `os.ReadDir()` to scan directories for `*_test.go` files.

### Script Runners

```bash
# Execute Lua scripts
omni lua script.lua

# Execute Python scripts
omni python script.py
```

Uses embedded interpreters (gopher-lua for Lua, potentially starlark for Python subset) to maintain the "no exec" principle.

### Taskfile Linter

```bash
omni lint Taskfile.yml
```

Checks:
- [ ] No shell-specific commands (rm, grep, etc.)
- [ ] All commands are omni-compatible
- [ ] No `exec` or external process calls

### Archive Commands

```bash
omni archive create out.zip ./dir
omni archive list out.zip
omni archive extract out.zip
```

### Hash Commands

```bash
omni hash file.bin
omni hash dir/ --recursive
omni hash --verify checksums.txt
```

---

## Phase 7 – Developer Productivity & Infrastructure

Code generation, data conversion, and internal improvements.

### Data Conversion Commands

| Command | Description | Priority |
|---------|-------------|----------|
| `json2struct` | Convert JSON to Go struct definition | P0 |
| `struct2json` | Generate JSON from Go struct (with tags) | P1 |
| `yaml2struct` | Convert YAML to Go struct definition | P0 |
| `struct2yaml` | Generate YAML from Go struct | P1 |

```bash
omni json2struct api_response.json -p Response
omni yaml2struct config.yaml -p Config
echo '{"name":"test","count":1}' | omni json2struct
```

### Formatters

| Command | Description | Priority |
|---------|-------------|----------|
| `cron` | Parse and format cron expressions | P1 |
| `cron next` | Show next N execution times | P1 |
| `cron explain` | Human-readable cron explanation | P1 |
| `datefmt` | Extended date formatting and parsing | P2 |

```bash
omni cron next "0 9 * * 1-5" -n 5    # Next 5 weekday 9am runs
omni cron explain "*/15 * * * *"     # "Every 15 minutes"
omni datefmt -i "2024-01-15" -f "Jan 2, 2006"
```

### Code Statistics (tokei-like)

| Command | Description | Priority |
|---------|-------------|----------|
| `loc` | Count lines of code by language | P1 |
| `loc --json` | JSON output for CI integration | P1 |

```bash
omni loc .                    # Count LOC in current directory
omni loc --json ./src         # JSON output
omni loc --exclude vendor     # Exclude directories
```

Output:
```
Language        Files    Lines     Code  Comments    Blanks
Go                 98    15234    12456      1234      1544
Markdown           12      890      890         0         0
YAML                5      234      200        20        14
─────────────────────────────────────────────────────────
Total             115    16358    13546      1254      1558
```

### Infrastructure Improvements

| Feature | Description | Priority | Effort |
|---------|-------------|----------|--------|
| Unified input package | `internal/cli/input` - file/stdin handling | P0 | 2-3 hrs |
| Logger JSON export | Export logs to file/stdout in JSON | P1 | 1-2 hrs |
| Missing tests | twig/builder, twig/parser tests | P1 | 1-2 hrs |
| Command interface | Define unified Command interface contract | P2 | 2-3 hrs |

### Unified Input Package

```go
// internal/cli/input/input.go
package input

// Reader handles file or stdin input uniformly
func Reader(args []string) (io.Reader, func(), error)

// Lines reads all lines from files or stdin
func Lines(args []string) ([]string, error)

// ForEach processes each file/stdin with callback
func ForEach(args []string, fn func(r io.Reader, name string) error) error
```

Benefits:
- Reduces ~400 lines of duplicated code across 27+ packages
- Consistent `-` for stdin handling
- Uniform error messages

### Logger JSON Export

```go
// Current
logger.Info("message", "key", value)

// With JSON export
logger.SetFormat(logger.FormatJSON)
logger.SetOutput(file)
```

```bash
omni --log-format=json --log-file=omni.log ls
```

---

## Library API Summary

All commands are available as importable Go packages:

```go
import (
    "github.com/inovacc/omni/internal/cli/fs"       // File system operations
    "github.com/inovacc/omni/internal/cli/text"     // Text processing
    "github.com/inovacc/omni/internal/cli/df"       // Disk usage
    "github.com/inovacc/omni/internal/cli/hash"     // Hashing utilities
    "github.com/inovacc/omni/internal/cli/archive"  // Archive operations
)

// Examples
fs.Copy("src", "dst", &fs.CopyOptions{Recursive: true})
text.Sort(lines, &text.SortOptions{Reverse: true})
usage, _ := df.RunDf(os.Stdout, []string{"/"}, &df.Options{})
sum, _ := hash.SHA256File("file.bin")
```

---

## Output Modes

All commands support multiple output formats:

### Text (Default)
```bash
omni ls
```

### JSON
```bash
omni ls --json
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

### Current Coverage Status (January 2026)

| Category | Packages | Coverage | Status |
|----------|----------|----------|--------|
| **Overall** | 86/88 packages | 97.7% | ✅ Excellent |
| **CLI Packages** | 79/79 packages | 100% | ✅ Complete |
| **Infrastructure** | `flags`, `logger`, `twig/*` | 85%+ | ✅ Good |
| **No Coverage (0%)** | 2 packages (`twig/builder`, `twig/parser`) | 0% | ⚠️ Integration Only |

### Test Statistics
- **Total Test Cases:** 700+
- **CLI Packages with Tests:** 79/79 (100%)
- **Internal Packages with Tests:** 86/88 (97.7%)

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
omni ls testdata/ > testdata/ls.golden

# Compare in tests
func TestLsGolden(t *testing.T) {
    // Compare actual vs golden file
}
```

---

## Testing Milestone - Phase 7

### Milestone 7.1: Core Test Coverage (Target: 95%)

| Package | Current | Target | Priority | Notes |
|---------|---------|--------|----------|-------|
| `basename` | 100% | 100% | ✅ | Done |
| `date` | 100% | 100% | ✅ | Done |
| `dirname` | 100% | 100% | ✅ | Done |
| `cat` | 84% | 95% | P0 | Add binary, encoding tests |
| `grep` | 86% | 95% | P0 | Add context, regex edge cases |
| `head` | 88% | 95% | P0 | Add bytes mode edge cases |
| `ls` | 85% | 95% | P0 | Add symlink, permission tests |
| `pwd` | 83% | 95% | P1 | Add chdir tests |
| `realpath` | 91% | 95% | P1 | Add symlink chain tests |
| `wc` | 94% | 95% | P1 | Add unicode edge cases |

### Milestone 7.2: Utility Test Coverage (Target: 80%)

| Package | Current | Target | Priority | Notes |
|---------|---------|--------|----------|-------|
| `base` | 78% | 85% | P1 | Add base58 edge cases |
| `crypt` | 66% | 80% | P1 | Add key derivation tests |
| `env` | 52% | 80% | P1 | Add filter, unset tests |
| `jq` | 55% | 80% | P1 | Add complex queries |
| `kill` | 58% | 80% | P2 | Platform-specific signals |
| `random` | 76% | 85% | P1 | Add distribution tests |
| `tail` | 69% | 85% | P1 | Add follow mode tests |
| `uname` | 60% | 80% | P2 | Platform-specific fields |
| `uuid` | 88% | 95% | P1 | Add version validation |
| `yq` | 76% | 85% | P1 | Add YAML edge cases |

### Milestone 7.3: Uncovered Packages (Target: 60%)

Priority P0 - Essential commands:
| Package | Priority | Estimated Tests |
|---------|----------|-----------------|
| `copy` | P0 | 20+ tests |
| `rm` | P0 | 15+ tests (safety critical) |
| `mkdir` | P0 | 10+ tests |
| `find` | P0 | 25+ tests |
| `sed` | P0 | 20+ tests |

Priority P1 - Common utilities:
| Package | Priority | Estimated Tests |
|---------|----------|-----------------|
| `chmod` | P1 | 15+ tests |
| `stat` | P1 | 15+ tests |
| `cut` | P1 | 15+ tests |
| `tr` | P1 | 15+ tests |
| `nl` | P1 | 10+ tests |
| `seq` | P1 | 10+ tests |

Priority P2 - Specialized:
| Package | Priority | Estimated Tests | Status |
|---------|----------|-----------------|--------|
| `archive` | P2 | 14 tests | ✅ Done |
| `gzip` | P2 | 12 tests | ✅ Done |
| `bzip2` | P2 | 10 tests | ✅ Done |
| `xz` | P2 | 14 tests | ✅ Done |
| `lint` | P1 | 17 tests | ✅ Done |
| `df` | P2 | 10+ tests | 🔄 Pending |
| `du` | P2 | 15+ tests | 🔄 Pending |
| `ps` | P2 | 10+ tests | 🔄 Pending |

### Testing Completion Criteria

- [ ] All P0 packages have ≥80% coverage
- [ ] All P1 packages have ≥60% coverage
- [ ] All tests pass on Linux, macOS, and Windows
- [ ] No flaky tests (consistent results across 10 runs)
- [ ] Edge cases documented in test names
- [ ] Error paths tested for all error-returning functions

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

### v0.1.0 - MVP ✅
- [x] Phase 1 commands complete
- [x] JSON output mode (partial)
- [x] Basic documentation
- [x] CI/CD pipeline

### v0.2.0 - File Operations ✅
- [x] Phase 2 commands
- [x] Safe rm with dry-run
- [x] Improved error messages

### v0.3.0 - Text Processing ✅
- [x] Phase 3 commands
- [x] grep with regex support
- [x] sort, uniq, cut, etc.

### v0.4.0 - System Info ✅
- [x] Phase 4 commands
- [x] Cross-platform df/du
- [x] Process utilities (ps, kill)

### v0.5.0 - Advanced Utilities ✅
- [x] Phase 5 commands
- [x] xargs, watch, yes
- [x] Archive operations

### v0.6.0 - Ecosystem (Current)
- [x] --json flag for most commands
- [ ] ID generators (ksuid, ulid, uuid7, nanoid, Snowflake)
- [ ] JSON beautify/minify
- [ ] Enhanced jq with gojq
- [ ] Template rendering

### v0.7.0 - Developer Productivity
- [ ] json2struct / yaml2struct
- [ ] cron formatter
- [ ] Code statistics (loc)
- [ ] Unified input package

### v1.0.0 - Production Ready
- [ ] All phases complete
- [ ] Full documentation
- [ ] Taskfile linter
- [ ] Comprehensive tests
- [ ] 95%+ test coverage

---

## Quick Start

```bash
# Build
go build -o omni ./cmd/omni

# Or run directly
go run . ls
go run . pwd --json
go run . date --format "2006-01-02"

# Use as library
import "github.com/inovacc/omni/pkg/fs"
files, _ := fs.Ls(".", fs.LsOptions{})
```

---

## References

- Go standard library: https://pkg.go.dev/std
- Cobra CLI: https://github.com/spf13/cobra
- twig (tree replacement): https://github.com/inovacc/twig
- fsnotify: https://github.com/fsnotify/fsnotify

## Related Documents

- [BACKLOG.md](BACKLOG.md) - Future work and technical debt
- [COMMANDS.md](COMMANDS.md) - Full command reference
- [REUSABILITY.md](REUSABILITY.md) - Code consolidation analysis

---

## Phase 8 – Data Conversion & Validation Tools

Comprehensive data conversion, validation, and developer utilities inspired by CodeBeautify and specialized tools.

### Format Converters

| Command | Description | Priority | Status |
|---------|-------------|----------|--------|
| `json toyaml` | Convert JSON to YAML | P0 | ✅ Done |
| `json fromyaml` | Convert YAML to JSON | P0 | ✅ Done |
| `json fromtoml` | Convert TOML to JSON | P0 | ✅ Done |
| `json toxml` | Convert JSON to XML | P1 | |
| `json fromxml` | Convert XML to JSON | P1 | |
| `json tocsv` | Convert JSON array to CSV | P1 | |
| `json fromcsv` | Convert CSV to JSON | P1 | |
| `yaml toml` | Convert YAML to TOML | P2 | |
| `yaml xml` | Convert YAML to XML | P2 | |
| `xml yaml` | Convert XML to YAML | P2 | |
| `csv sql` | Generate SQL INSERT from CSV | P2 | |

### Formatters & Beautifiers

| Command | Description | Priority | Status |
|---------|-------------|----------|--------|
| `json fmt` | Beautify/format JSON | P0 | ✅ Done |
| `json minify` | Minify JSON | P0 | ✅ Done |
| `json validate` | Validate JSON syntax | P0 | ✅ Done |
| `json stats` | Show JSON statistics | P0 | ✅ Done |
| `json keys` | List all JSON keys | P0 | ✅ Done |
| `xml fmt` | Beautify XML | P1 | |
| `xml minify` | Minify XML | P1 | |
| `xml validate` | Validate XML | P1 | |
| `sql fmt` | Format SQL queries | P2 | |
| `html fmt` | Format HTML | P2 | |
| `css fmt` | Format CSS | P2 | |
| `css minify` | Minify CSS | P2 | |

### Validators

| Command | Description | Priority |
|---------|-------------|----------|
| `json validate` | Validate JSON | ✅ Done |
| `yaml validate` | Validate YAML | P1 |
| `toml validate` | Validate TOML | P1 |
| `xml validate` | Validate XML | P1 |
| `email validate` | Validate email format | P2 |
| `url validate` | Validate URL format | P2 |
| `ip validate` | Validate IP address | P2 |

### Encode/Decode Utilities

| Command | Description | Priority | Status |
|---------|-------------|----------|--------|
| `base64` | Base64 encode/decode | P0 | ✅ Done |
| `base32` | Base32 encode/decode | P0 | ✅ Done |
| `base58` | Base58 encode/decode | P0 | ✅ Done |
| `url encode/decode` | URL encoding | P1 | |
| `html encode/decode` | HTML entity encoding | P1 | |
| `hex encode/decode` | Hex encoding | P1 | |
| `jwt decode` | Decode JWT tokens | P1 | |

### String Utilities

| Command | Description | Priority |
|---------|-------------|----------|
| `case upper` | Convert to UPPERCASE | P1 |
| `case lower` | Convert to lowercase | P1 |
| `case camel` | Convert to camelCase | P1 |
| `case pascal` | Convert to PascalCase | P1 |
| `case snake` | Convert to snake_case | P1 |
| `case kebab` | Convert to kebab-case | P1 |
| `text reverse` | Reverse text | P2 |
| `text dedup` | Remove duplicate lines | P2 |
| `text trim` | Remove empty lines/spaces | P2 |

### Hash & Cryptography (Extended)

| Command | Description | Priority | Status |
|---------|-------------|----------|--------|
| `hash md5` | MD5 hash | P0 | ✅ Done |
| `hash sha1` | SHA1 hash | P0 | ✅ Done |
| `hash sha256` | SHA256 hash | P0 | ✅ Done |
| `hash sha512` | SHA512 hash | P0 | ✅ Done |
| `hash blake2b` | BLAKE2b hash | P1 | |
| `hash blake3` | BLAKE3 hash | P1 | |
| `hmac` | HMAC generation | P2 | |
| `encrypt aes` | AES encryption | P0 | ✅ Done |
| `decrypt aes` | AES decryption | P0 | ✅ Done |

### Random Generators (Extended)

| Command | Description | Priority | Status |
|---------|-------------|----------|--------|
| `random number` | Random numbers | P0 | ✅ Done |
| `random string` | Random strings | P0 | ✅ Done |
| `random uuid` | UUID v4/v7 | P0 | ✅ Done |
| `random ulid` | ULID | P0 | ✅ Done |
| `random ksuid` | KSUID | P0 | ✅ Done |
| `random nanoid` | NanoID | P0 | ✅ Done |
| `random snowflake` | Snowflake ID | P0 | ✅ Done |
| `random password` | Password generation | P1 | |
| `random color` | Random hex color | P2 | |
| `random date` | Random date | P2 | |
| `random name` | Random names | P3 | |

### Brazilian Document Tools (brdoc integration)

| Command | Description | Priority |
|---------|-------------|----------|
| `brdoc cpf generate` | Generate valid CPF | P1 |
| `brdoc cpf validate` | Validate CPF | P1 |
| `brdoc cpf format` | Format CPF (XXX.XXX.XXX-XX) | P1 |
| `brdoc cnpj generate` | Generate valid CNPJ (alphanumeric) | P1 |
| `brdoc cnpj validate` | Validate CNPJ | P1 |
| `brdoc cnpj format` | Format CNPJ | P1 |

Reference: https://github.com/inovacc/brdoc

### Go Struct Tag Tools (tagfixer integration)

| Command | Description | Priority |
|---------|-------------|----------|
| `tagfixer` | Fix struct tags in Go files | P1 |
| `tagfixer analyze` | Analyze tag usage patterns | P1 |
| `tagfixer --case camel` | Convert to camelCase | P1 |
| `tagfixer --case snake` | Convert to snake_case | P1 |
| `tagfixer --case kebab` | Convert to kebab-case | P1 |
| `tagfixer --tags json,yaml` | Fix multiple tag types | P1 |

Reference: https://github.com/inovacc/tagfixer

### Container Signing (cosign-inspired)

| Command | Description | Priority |
|---------|-------------|----------|
| `sign` | Sign files/artifacts | P2 |
| `sign verify` | Verify signatures | P2 |
| `sign keypair` | Generate signing keypairs | P2 |
| `sign blob` | Sign arbitrary blobs | P2 |

Reference: https://github.com/sigstore/cosign

### IP & Network Tools

| Command | Description | Priority |
|---------|-------------|----------|
| `ip info` | Get IP info (local/public) | P2 |
| `ip tohex` | IP to hex conversion | P2 |
| `ip tobin` | IP to binary | P2 |
| `dns lookup` | DNS lookup | P2 |
| `port check` | Check if port is open | P2 |

### Number Converters

| Command | Description | Priority |
|---------|-------------|----------|
| `num tobin` | Decimal to binary | P2 |
| `num tohex` | Decimal to hex | P2 |
| `num tooct` | Decimal to octal | P2 |
| `num frombin` | Binary to decimal | P2 |
| `num fromhex` | Hex to decimal | P2 |
| `num words` | Number to words | P3 |

### Color Converters

| Command | Description | Priority |
|---------|-------------|----------|
| `color hex2rgb` | HEX to RGB | P3 |
| `color rgb2hex` | RGB to HEX | P3 |
| `color hex2hsl` | HEX to HSL | P3 |
| `color random` | Random color | P3 |

### Diff & Compare Tools

| Command | Description | Priority | Status |
|---------|-------------|----------|--------|
| `diff` | Text file diff | P1 | ✅ Done |
| `diff json` | JSON diff | P1 | ✅ Done |
| `diff yaml` | YAML diff | P2 | |
| `cmp` | Binary file compare | P1 | ✅ Done |

### Misc Utilities

| Command | Description | Priority |
|---------|-------------|----------|
| `lorem` | Lorem ipsum generator | P3 |
| `qr generate` | Generate QR codes | P3 |
| `qr decode` | Decode QR codes | P3 |
| `barcode` | Generate barcodes | P3 |

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
