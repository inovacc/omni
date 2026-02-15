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
| `path clean` | `filepath.Clean()` | `--json` | ✅ Done |
| `path abs` | `filepath.Abs()` | `--json` | ✅ Done |

### Architecture

```
omni/
├── cmd/                    # Cobra CLI commands (155+ commands)
│   ├── root.go
│   ├── ls.go
│   ├── pwd.go
│   ├── sqlite.go          # SQLite database management
│   ├── bbolt.go           # BoltDB key-value store
│   ├── generate.go        # Code generation tools
│   └── ...
├── pkg/                    # Reusable Go libraries (importable externally)
│   ├── idgen/              # UUID, ULID, KSUID, Nanoid, Snowflake
│   ├── hashutil/           # MD5, SHA256, SHA512, CRC32, CRC64 hashing
│   ├── jsonutil/           # jq-style JSON query engine
│   ├── encoding/           # Base64, Base32, Base58 encode/decode
│   ├── cryptutil/          # AES-256-GCM encrypt/decrypt
│   ├── sqlfmt/             # SQL format/minify/validate
│   ├── cssfmt/             # CSS format/minify/validate
│   ├── htmlfmt/            # HTML format/minify/validate
│   ├── textutil/           # Sort, Uniq, Trim + diff/
│   ├── search/grep/        # Pattern search with options
│   ├── search/rg/          # Gitignore parsing, file type matching
│   ├── pipeline/           # Streaming text processing engine
│   ├── figlet/             # FIGlet font parser and ASCII art
│   ├── twig/               # Tree scanning, formatting, comparison
│   └── video/              # Video download engine (YouTube, HLS, generic)
├── internal/
│   ├── cli/               # CLI wrappers (I/O, flags, stdin handling)
│   │   ├── ls/
│   │   ├── pwd/
│   │   ├── sqlite/        # SQLite operations
│   │   ├── bbolt/         # BoltDB operations
│   │   ├── generate/      # Code generation with templates
│   │   └── ...
│   ├── flags/             # Feature flags system
│   └── logger/            # KSUID-based logging with query support
├── include/               # Template reference files
│   └── cobra/             # Cobra app templates
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

### Enhanced Search (ripgrep-inspired)

Incorporate features from [ripgrep](https://github.com/BurntSushi/ripgrep) for fast, modern search:

| Feature | Description | Priority | Status |
|---------|-------------|----------|--------|
| `rg` | ripgrep-style search command | P1 | ✅ Done |
| Recursive search | Search directories recursively by default | P1 | ✅ Done |
| Gitignore support | Respect .gitignore files | P1 | ✅ Done |
| File type filtering | `--type`, `-t` for language-specific search | P1 | ✅ Done |
| Glob patterns | `--glob`, `-g` for file matching | P1 | ✅ Done |
| Context lines | `-A`, `-B`, `-C` for context around matches | P1 | ✅ Done |
| JSON output | `--json` for structured output | P1 | ✅ Done |
| Binary detection | Skip binary files automatically | P1 | ✅ Done |
| Replace mode | `--replace`, `-r` for search and replace | P2 | |
| Multiline search | `-U` for multiline patterns | P2 | |
| PCRE2 support | `--pcre2` for advanced regex | P3 | |

#### Ripgrep Gaps (vs Rust implementation)

Based on comparison with ripgrep Rust source code, the following improvements are planned:

| Feature | Description | Priority | Impact | Status |
|---------|-------------|----------|--------|--------|
| Parallel walking | Work-stealing parallel directory traversal | P1 | 2-10x speedup on multi-core | ✅ Done |
| Gitignore negation | Support `!pattern` negation in .gitignore | P1 | Correctness | ✅ Done |
| .ignore files | Respect `.ignore` files (ripgrep-specific) | P1 | Compatibility | ✅ Done |
| Global gitignore | Read `~/.config/git/ignore` | P1 | Completeness | ✅ Done |
| .git/info/exclude | Read `.git/info/exclude` patterns | P1 | Completeness | ✅ Done |
| Directory-only patterns | Support `dir/` patterns in gitignore | P1 | Correctness | ✅ Done |
| Literal optimization | Fast path for literal string patterns | P3 | Performance | ✅ Done |
| Streaming JSON | Stream JSON output instead of batch | P3 | Memory for large results | ✅ Done |
| Color output | ANSI color highlighting for matches | P2 | UX | |
| Match highlighting | Highlight matched text within lines | P2 | UX | |
| More file types | Expand from ~20 to 100+ type definitions | P2 | Compatibility | |
| Type composition | `--type-add 'web:include:html,css,js'` | P2 | Power users | |
| Column numbers | Show column position of matches | P2 | IDE integration | ✅ Done |
| Candidate detection | Two-path search with fast candidate detection | P3 | Performance | |

**Architecture differences:**
- Ripgrep uses modular crates (grep-searcher, grep-matcher, grep-printer, ignore)
- Ripgrep uses trait-based `Matcher` interface supporting multiple regex engines
- Ripgrep uses `Sink` pattern for push-based result streaming
- Ripgrep has three specialized searchers: `ReadByLine`, `SliceByLine`, `MultiLine`
- Ripgrep uses compiled `GlobSet` for O(1) pattern matching vs our O(n)

```go
type RgOptions struct {
    Recursive      bool     // search recursively (default: true)
    IgnoreCase     bool     // -i: case insensitive
    SmartCase      bool     // -S: smart case matching
    WordRegexp     bool     // -w: match whole words
    LineNumber     bool     // -n: show line numbers
    Count          bool     // -c: count matches
    FilesWithMatch bool     // -l: only show file names
    Context        int      // -C: lines of context
    Before         int      // -B: lines before match
    After          int      // -A: lines after match
    Types          []string // -t: file types to include
    Glob           []string // -g: glob patterns
    Hidden         bool     // --hidden: search hidden files
    NoIgnore       bool     // --no-ignore: don't respect gitignore
    Replace        string   // -r: replacement text
    JSON           bool     // --json: JSON output
}
```

**Examples:**
```bash
# Search recursively (default)
omni rg "pattern" ./src

# Search specific file types
omni rg -t go "func main"
omni rg -t js -t ts "import"

# With context
omni rg -C 3 "error" ./logs

# Respect/ignore gitignore
omni rg --no-ignore "TODO"
omni rg --hidden "secret"  # include hidden files

# JSON output for tooling
omni rg --json "pattern" | omni jq '.data.lines'

# Search and replace (dry-run)
omni rg "old_name" -r "new_name" --dry-run

# Glob patterns
omni rg -g "*.go" -g "!*_test.go" "pattern"
```

**Go Implementation Notes:**
- Use `filepath.WalkDir` for recursive traversal
- Parse `.gitignore` with custom parser or `github.com/sabhiram/go-gitignore`
- Maintain ripgrep CLI compatibility where possible
- Consider using `regexp2` for PCRE2-like features

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
| `pipe` | Chain omni commands with variable substitution | `--var`, `--json`, `--sep`, `-v` | P0 ✅ |
| `pipeline` | Internal streaming engine | — | P0 |
| `exec` | Execute arbitrary scripts using OS features | `--shell`, `--timeout`, `--env` | P1 |

### exec Design (External Process Runner)

```bash
# Execute scripts using the OS native shell
omni exec my_script.sh          # Unix: sh, Windows: cmd
omni exec my_script.ps1         # PowerShell (auto-detected on Windows)
omni exec my_script.py          # Delegates to python3/python
omni exec --shell bash script.sh  # Explicit shell override
omni exec --timeout 30s long_task.sh  # With timeout
omni exec --env KEY=VAL script.sh     # Inject env vars
```

> **Note:** This is an intentional exception to the "no exec" principle.
> `exec` exists as an escape hatch for scripts that cannot be expressed
> as pure-Go omni commands. It uses `os/exec` to spawn the appropriate
> interpreter based on file extension or `--shell` flag.

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

### Pipe Design (Variable Substitution)

```go
type Options struct {
    JSON      bool   // --json: output pipeline result as JSON
    Separator string // --sep: command separator (default "|")
    Verbose   bool   // --verbose: show intermediate steps
    VarName   string // --var: variable name for output substitution (default "OUT")
}
```

Variable substitution patterns:
- `$OUT` or `${OUT}` - Single value substitution (uses last line of output)
- `[$OUT...]` - Iteration over each line of output

Usage:
```bash
# Generate UUID and create folder with that name
omni pipe '{uuid -v 7}' '{mkdir $OUT}'

# Custom variable name
omni pipe --var UUID '{uuid -v 7}' '{mkdir $UUID}'

# Generate 10 UUIDs and create a folder for each
omni pipe '{uuid -v 7 -n 10}' '{mkdir [$OUT...]}'

# Chain with processing
omni pipe '{uuid -v 7 -n 5}' '{echo prefix-[$OUT...]-suffix}'
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

### Code Statistics (tokei-like) ✅ DONE

| Command | Description | Status |
|---------|-------------|--------|
| `loc` | Count lines of code by language | ✅ Done |
| `loc --json` | JSON output for CI integration | ✅ Done |
| `loc --exclude` | Exclude directories from counting | ✅ Done |
| `loc --hidden` | Include hidden files | ✅ Done |

**Features:**
- 40+ language definitions with proper syntax awareness
- String literal tracking (URLs in strings not counted as comments)
- Block comment support (including nested for Rust, Haskell, etc.)
- Literate mode for Markdown (extracts embedded code blocks)
- Embedded code displayed as subtotals (`|- Shell`, `|- Go`, etc.)
- Parses embedded code with language syntax for accurate comment detection

```bash
omni loc .                    # Count LOC in current directory
omni loc --json ./src         # JSON output
omni loc --exclude vendor     # Exclude directories
omni loc --hidden             # Include hidden files
```

Output:
```
Language              Files      Lines       Code   Comments     Blanks
-------------------------------------------------------------------------------
 Go                     380      74056      56744       3060      14252
 Markdown                 6       2878          0       2141        737
 |- Shell                          808        710         57         41
 |- Go                             280        200         40         40
 (Total)                         4007        948       2238        821
===============================================================================
 Total                  414      93411      72990       5310      15111
```

### Infrastructure Improvements

| Feature | Description | Priority | Effort |
|---------|-------------|----------|--------|
| Unified input package | `internal/cli/input` - file/stdin handling | P0 | 2-3 hrs |
| Logger JSON export | Export logs to file/stdout in JSON | P1 | 1-2 hrs |
| Missing tests | twig/builder, twig/parser tests | P1 | 1-2 hrs |
| ~~Command interface~~ | ~~Define unified Command interface contract~~ | ✅ Done | — |

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

### Query Logging (Implemented)

Database queries can be logged with timing and result information:

```go
// Log query with result
logger.LogQueryResult(database, query, rowCount, duration, err)

// Log query with result data (for debugging)
logger.LogQueryWithData(database, query, columns, rows, duration, err)

// Use QueryLogger for convenience
ql := logger.NewQueryLogger(l, "/path/to/db.sqlite")
ql.Log(query, rowCount, duration, err)
```

```bash
# Enable query logging
eval "$(omni logger --path /tmp/omni-logs)"

# Run query with logging
omni sqlite query mydb.sqlite "SELECT * FROM users"

# Include result data in logs (use with caution)
omni sqlite query mydb.sqlite "SELECT * FROM users" --log-data
```

Log entry example:
```json
{
  "msg": "query_result",
  "database": "mydb.sqlite",
  "query": "SELECT * FROM users",
  "status": "success",
  "rows": 10,
  "duration_ms": 25,
  "timestamp": "2026-01-29T12:00:00Z"
}
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
- **Query Logging Tests:** Comprehensive coverage for LogQuery, LogQueryResult, LogQueryWithData, QueryLogger

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
- ripgrep (search inspiration): https://github.com/BurntSushi/ripgrep

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
| `xxd` | Hex dump and reverse (like Unix xxd) | P0 | ✅ Done |
| `url encode/decode` | URL encoding | P1 | |
| `html encode/decode` | HTML entity encoding | P1 | |
| `hex encode/decode` | Hex encoding | P1 | |
| `jwt decode` | Decode JWT tokens | P1 | |

### String Utilities

| Command | Description | Priority | Status |
|---------|-------------|----------|--------|
| `case upper` | Convert to UPPERCASE | P1 | ✅ Done |
| `case lower` | Convert to lowercase | P1 | ✅ Done |
| `case title` | Convert to Title Case | P1 | ✅ Done |
| `case sentence` | Convert to Sentence case | P1 | ✅ Done |
| `case camel` | Convert to camelCase | P1 | ✅ Done |
| `case pascal` | Convert to PascalCase | P1 | ✅ Done |
| `case snake` | Convert to snake_case | P1 | ✅ Done |
| `case kebab` | Convert to kebab-case | P1 | ✅ Done |
| `case constant` | Convert to CONSTANT_CASE | P1 | ✅ Done |
| `case dot` | Convert to dot.case | P1 | ✅ Done |
| `case path` | Convert to path/case | P1 | ✅ Done |
| `case swap` | Swap case (Hello → hELLO) | P1 | ✅ Done |
| `case toggle` | Toggle first char case | P1 | ✅ Done |
| `case detect` | Detect case type | P1 | ✅ Done |
| `case all` | Show all conversions | P1 | ✅ Done |
| `text reverse` | Reverse text | P2 | |
| `text dedup` | Remove duplicate lines | P2 | |
| `text trim` | Remove empty lines/spaces | P2 | |

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

| Command | Description | Priority | Status |
|---------|-------------|----------|--------|
| `brdoc cpf generate` | Generate valid CPF | P1 | ✅ Done |
| `brdoc cpf validate` | Validate CPF | P1 | ✅ Done |
| `brdoc cpf format` | Format CPF (XXX.XXX.XXX-XX) | P1 | ✅ Done |
| `brdoc cnpj generate` | Generate valid CNPJ (alphanumeric) | P1 | ✅ Done |
| `brdoc cnpj validate` | Validate CNPJ | P1 | ✅ Done |
| `brdoc cnpj format` | Format CNPJ | P1 | ✅ Done |

Reference: https://github.com/inovacc/brdoc

### Go Struct Tag Tools (tagfixer integration)

| Command | Description | Priority | Status |
|---------|-------------|----------|--------|
| `tagfixer` | Fix struct tags in Go files | P1 | ✅ Done |
| `tagfixer analyze` | Analyze tag usage patterns | P1 | ✅ Done |
| `tagfixer --case camel` | Convert to camelCase | P1 | ✅ Done |
| `tagfixer --case snake` | Convert to snake_case | P1 | ✅ Done |
| `tagfixer --case kebab` | Convert to kebab-case | P1 | ✅ Done |
| `tagfixer --case pascal` | Convert to PascalCase | P1 | ✅ Done |
| `tagfixer --tags json,yaml` | Fix multiple tag types | P1 | ✅ Done |
| `tagfixer --dry-run` | Preview changes | P1 | ✅ Done |
| `tagfixer --json` | JSON output | P1 | ✅ Done |

Implemented directly in omni (not external dependency)

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

### HTTP Client (curlie-inspired)

Reference: https://github.com/rs/curlie

| Command | Description | Priority |
|---------|-------------|----------|
| `curl` | HTTP client with httpie-like output | P1 |
| `curl GET` | HTTP GET request | P1 |
| `curl POST` | HTTP POST with body | P1 |
| `curl PUT` | HTTP PUT request | P1 |
| `curl DELETE` | HTTP DELETE request | P1 |
| `curl HEAD` | HTTP HEAD request | P2 |
| `curl PATCH` | HTTP PATCH request | P2 |

```bash
# Simple GET request
omni curl https://api.example.com/users

# POST with JSON body
omni curl POST https://api.example.com/users name=John email=john@example.com

# With headers
omni curl https://api.example.com/users Authorization:"Bearer token"

# Form data
omni curl -f POST https://api.example.com/upload file@./data.txt

# JSON output (raw response)
omni curl --json https://api.example.com/users

# Verbose mode
omni curl -v https://api.example.com/users

# Custom headers
omni curl -H "Accept: application/xml" https://api.example.com/data
```

Features:
- [ ] httpie-like syntax for headers and data (key:value, key=value)
- [ ] Colored/formatted output for JSON responses
- [ ] Support for file uploads (@file syntax)
- [ ] Cookie handling
- [ ] Follow redirects
- [ ] Timeout configuration
- [ ] Proxy support
- [ ] TLS/SSL options
- [ ] Pure Go implementation (no exec)

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

## Phase 9 – Database & Code Generation Tools

Database management CLIs and code scaffolding utilities.

### Database Tools

| Command | Description | Priority | Status |
|---------|-------------|----------|--------|
| `bbolt info` | Display database page size | P0 | ✅ Done |
| `bbolt stats` | Show database statistics | P0 | ✅ Done |
| `bbolt buckets` | List all buckets | P0 | ✅ Done |
| `bbolt keys` | List keys in bucket | P0 | ✅ Done |
| `bbolt get` | Get value for key | P0 | ✅ Done |
| `bbolt put` | Store key-value pair | P0 | ✅ Done |
| `bbolt delete` | Delete key | P0 | ✅ Done |
| `bbolt dump` | Dump bucket contents | P0 | ✅ Done |
| `bbolt compact` | Compact database | P0 | ✅ Done |
| `bbolt check` | Verify integrity | P0 | ✅ Done |
| `bbolt pages` | List database pages | P1 | ✅ Done |
| `bbolt page` | Hex dump of page | P1 | ✅ Done |
| `bbolt create-bucket` | Create new bucket | P0 | ✅ Done |
| `bbolt delete-bucket` | Delete bucket | P0 | ✅ Done |
| `sqlite stats` | Show database statistics | P0 | ✅ Done |
| `sqlite tables` | List all tables | P0 | ✅ Done |
| `sqlite schema` | Show table schema | P0 | ✅ Done |
| `sqlite columns` | Show table columns | P0 | ✅ Done |
| `sqlite indexes` | List all indexes | P0 | ✅ Done |
| `sqlite query` | Execute SQL query | P0 | ✅ Done |
| `sqlite vacuum` | Optimize database | P0 | ✅ Done |
| `sqlite check` | Verify integrity | P0 | ✅ Done |
| `sqlite dump` | Export as SQL | P0 | ✅ Done |
| `sqlite import` | Import SQL file | P0 | ✅ Done |

### Server Lifecycle Management

| Command | Description | Priority | Status |
|---------|-------------|----------|--------|
| `server init` | Initialize server with clean database | P0 | |
| `server start` | Start server instance | P0 | |
| `server stop` | Stop running server instance | P0 | |
| `server status` | Check if server is running | P0 | |
| `server restart` | Stop and start server | P1 | |

#### Init Behavior

When `server init` is called:
1. Check if server instance is running
2. If running: stop server, delete database, start server
3. If not running: delete database, start server

```bash
# Initialize with fresh database (auto-handles running server)
omni server init

# Manual workflow equivalent
omni server status          # Check if running
omni server stop            # Stop if running
omni sqlite drop mydb.sqlite  # Delete database
omni server start           # Start fresh
```

### BoltDB CLI Examples

```bash
omni bbolt stats mydb.bolt
omni bbolt buckets mydb.bolt
omni bbolt keys mydb.bolt users
omni bbolt get mydb.bolt users user1
omni bbolt put mydb.bolt config version 1.0.0
omni bbolt compact mydb.bolt mydb-compact.bolt
```

### SQLite CLI Examples

```bash
omni sqlite stats mydb.sqlite
omni sqlite tables mydb.sqlite
omni sqlite query mydb.sqlite "SELECT * FROM users"
omni sqlite dump mydb.sqlite > backup.sql
omni sqlite import mydb.sqlite backup.sql
```

### Code Generation Tools

| Command | Description | Priority | Status |
|---------|-------------|----------|--------|
| `generate cobra init` | Generate Cobra CLI scaffold | P0 | ✅ Done |
| `generate cobra add` | Add new command to Cobra app | P0 | ✅ Done |
| `generate handler` | Generate HTTP handler boilerplate | P1 | |
| `generate grpc` | Generate gRPC service scaffold | P1 | |
| `generate repository` | Generate repository pattern code | P1 | |
| `generate test` | Generate test file scaffold | P1 | |
| `generate mock` | Generate mock implementations | P2 | |

### Protocol Buffer Tools (buf integration)

Reference: https://github.com/bufbuild/buf

| Command | Description | Priority | Status |
|---------|-------------|----------|--------|
| `buf lint` | Lint protobuf files for best practices | P0 | |
| `buf format` | Format protobuf files | P0 | |
| `buf breaking` | Detect breaking changes in protobuf | P0 | |
| `buf build` | Build/compile protobuf files | P0 | |
| `buf generate` | Generate code from protobuf | P1 | |
| `buf mod init` | Initialize buf module | P1 | |
| `buf mod update` | Update buf dependencies | P1 | |
| `buf export` | Export protobuf files | P2 | |
| `buf convert` | Convert between protobuf formats | P2 | |

### Buf CLI Examples

```bash
# Lint protobuf files
omni buf lint proto/

# Format protobuf files
omni buf format proto/ --write

# Check for breaking changes
omni buf breaking proto/ --against .git#branch=main

# Build protobuf
omni buf build proto/

# Generate code
omni buf generate proto/

# Initialize buf module
omni buf mod init

# Update dependencies
omni buf mod update
```

### Buf Features

- [ ] Protobuf linting with 40+ configurable rules
- [ ] Breaking change detection (source and wire level)
- [ ] Code generation with plugin support
- [ ] File formatting and standardization
- [ ] buf.yaml / buf.gen.yaml configuration support
- [ ] Remote module/dependency support
- [ ] JSON, MSVS, JUnit output formats
- [ ] Integration with Buf Schema Registry (BSR)

### Generator CLI Examples

```bash
# Generate new Cobra CLI application
omni generate cobra init myapp --module github.com/user/myapp

# Add command to existing Cobra app
omni generate cobra add serve --parent root
omni generate cobra add user --parent root
omni generate cobra add list --parent user

# Generate with options
omni generate cobra init myapp --module github.com/user/myapp --viper --license MIT

# Generate handler with test
omni generate handler --name UserHandler --path ./handlers
omni generate test ./handlers/user.go
```

### Cobra Generator Features

- [x] Project initialization with go.mod
- [x] Command scaffolding with parent/child relationships
- [x] Viper integration for configuration
- [x] Persistent and local flags setup
- [x] License file generation (MIT, Apache-2.0, BSD-3)
- [x] README generation with usage examples
- [x] Makefile / Taskfile generation
- [x] GitHub Actions CI workflow
- [x] goreleaser configuration
- [x] Config file support (~/.cobra.yaml) compatible with cobra-cli
- [x] Service pattern with inovacc/config integration

### Generator Template System

```go
// internal/cli/generate/template.go
type CobraAppConfig struct {
    ModuleName  string   // github.com/user/myapp
    AppName     string   // myapp
    Description string   // App description
    Author      string   // Author name
    License     string   // MIT, Apache-2.0, BSD-3
    UseViper    bool     // Include viper for config
    Commands    []string // Initial commands to generate
}

func GenerateCobraApp(dir string, cfg CobraAppConfig) error
func AddCommand(dir string, name, parent string) error
```

### Generated Project Structure

```
myapp/
├── cmd/
│   ├── root.go          # Root command
│   ├── serve.go         # Generated subcommand
│   └── version.go       # Version command
├── internal/
│   └── config/
│       └── config.go    # Viper configuration (optional)
├── main.go              # Entry point
├── go.mod               # Go module
├── go.sum
├── Taskfile.yml         # Task runner
├── Makefile             # Make targets
├── README.md            # Documentation
├── LICENSE              # License file
└── .github/
    └── workflows/
        └── ci.yml       # GitHub Actions
```

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
| `xxd` | `encoding/hex` + custom formatting |

---

## Phase 10 – Backup & Archival (Restic-Inspired)

Advanced backup, encryption, and deduplication features inspired by [restic](https://github.com/restic/restic).

### Core Backup Features

| Command | Description | Priority | Status |
|---------|-------------|----------|--------|
| `backup create` | Create encrypted, deduplicated backup | P1 | |
| `backup restore` | Restore files from backup | P1 | |
| `backup list` | List snapshots | P1 | |
| `backup diff` | Show differences between snapshots | P2 | |
| `backup check` | Verify backup integrity | P1 | |
| `backup prune` | Remove old snapshots | P2 | |
| `backup mount` | Mount backup as filesystem (FUSE) | P3 | |

### Encryption & Security

| Feature | Description | Priority | Go Implementation |
|---------|-------------|----------|-------------------|
| AES-256 encryption | Content encryption at rest | P1 | `crypto/aes` |
| Poly1305-AES MAC | Authentication | P1 | `golang.org/x/crypto/poly1305` |
| scrypt KDF | Key derivation from password | P1 | `golang.org/x/crypto/scrypt` |
| Repository lock | Prevent concurrent modifications | P2 | File-based locking |

### Deduplication & Chunking

| Feature | Description | Priority |
|---------|-------------|----------|
| Content-aware chunking | Variable-size chunks based on content | P1 |
| Content-addressable storage | Blobs identified by hash | P1 |
| Pack files | Group blobs for efficient storage | P2 |
| Zstd compression | Configurable compression levels | P1 |

### Reusable Patterns from Restic

**Pattern Matching (from `internal/filter`):**
```go
// Advanced glob patterns with:
// - Recursive wildcards (**)
// - Negation patterns (!)
// - Directory-only patterns (dir/)
filter.Match(patterns, path) bool
```

**Tree Walker (from `internal/walker`):**
```go
// Visitor pattern for tree traversal
walker.Walk(ctx, tree, func(path string, node *Node) error {
    // Process each node
    return walker.Continue
})
```

**Backend Abstraction:**
```go
// Pluggable storage backend interface
type Backend interface {
    Save(ctx, handle, reader) error
    Load(ctx, handle, length, offset, fn) error
    Remove(ctx, handle) error
    List(ctx, fileType, fn) error
}
```

### Backend Support

| Backend | Description | Priority |
|---------|-------------|----------|
| Local filesystem | Default backend | P0 |
| S3/MinIO | AWS S3 compatible storage | P1 |
| SFTP | SSH file transfer | P2 |
| Azure Blob | Azure cloud storage | P2 |
| Google Cloud Storage | GCS backend | P2 |
| REST server | HTTP-based backend | P3 |
| Rclone | Use rclone as backend | P3 |

### Backup CLI Examples

```bash
# Initialize repository
omni backup init /path/to/repo

# Create backup
omni backup create /path/to/repo /path/to/data

# List snapshots
omni backup list /path/to/repo

# Restore snapshot
omni backup restore /path/to/repo latest /path/to/restore

# Check integrity
omni backup check /path/to/repo

# With S3 backend
omni backup init s3:bucket-name/prefix
omni backup create s3:bucket-name/prefix ./data
```

### Code Quality Patterns to Adopt

From restic's architecture:
- **Options struct with AddFlags()** - Consistent flag handling across commands
- **Context propagation** - All operations accept `context.Context`
- **Error wrapping** - Preserve stack traces with `pkg/errors`
- **Backend composition** - Wrap backends with Retry, Cache, Logger, Limiter
- **Iterator channels** - Lazy evaluation for large datasets

---

## Phase 11 – Cloud CLI (Exact Behavior Matching)

**Goal:** Match the exact behavior, command structure, flags, and output formats of AWS CLI, Azure CLI, and GCP gcloud.

### Design Principles

| Principle | Description |
|-----------|-------------|
| **Command compatibility** | Same command names and structure as official CLIs |
| **Flag compatibility** | Same flags with same short/long forms |
| **Output compatibility** | Same output formats (json, text, table, yaml) |
| **Query support** | JMESPath query support (`--query`) |
| **Profile/credential handling** | Same credential file locations and formats |
| **Script compatibility** | Drop-in replacement for existing scripts |

### Global Options (All Clouds)

```bash
# AWS pattern
omni aws [options] <service> <operation> [parameters]
--region          # AWS region
--profile         # Credential profile
--output          # json | text | table | yaml
--query           # JMESPath query
--endpoint-url    # Custom endpoint (LocalStack, etc.)
--debug           # Debug logging
--no-paginate     # Disable pagination

# Azure pattern
omni az [options] <group> <subgroup> <operation> [parameters]
--subscription    # Azure subscription
--output          # json | jsonc | table | tsv | yaml | yamlc | none
--query           # JMESPath query
--verbose         # Verbose output
--debug           # Debug logging

# GCP pattern
omni gcloud [options] <group> <command> [parameters]
--project         # GCP project
--account         # GCP account
--format          # json | yaml | text | table | csv | value | get
--filter          # Resource filter
--quiet           # Suppress prompts
```

---

### AWS CLI – Complete Service Coverage

#### Core Services (P0 - MVP)

| Service | Commands | Status |
|---------|----------|--------|
| **sts** | get-caller-identity, assume-role, get-session-token | ✅ Done |
| **s3** | ls, cp, mv, rm, mb, rb, sync, presign | ✅ Partial |
| **s3api** | get-object, put-object, list-objects-v2, delete-object, head-object | |
| **ec2** | describe-instances, start-instances, stop-instances, describe-vpcs, describe-security-groups | ✅ Done |
| **iam** | get-user, get-role, list-users, list-roles, list-policies, get-policy | ✅ Done |
| **ssm** | get-parameter, put-parameter, get-parameters-by-path, delete-parameter | ✅ Done |

#### Extended Services (P1)

| Service | Commands | Priority |
|---------|----------|----------|
| **lambda** | list-functions, invoke, create-function, update-function-code, delete-function | P1 |
| **dynamodb** | list-tables, scan, query, get-item, put-item, delete-item, create-table | P1 |
| **sns** | list-topics, create-topic, publish, subscribe, delete-topic | P1 |
| **sqs** | list-queues, create-queue, send-message, receive-message, delete-message | P1 |
| **cloudwatch** | get-metric-data, put-metric-data, describe-alarms | P1 |
| **logs** | describe-log-groups, filter-log-events, create-log-group | P1 |

#### Infrastructure Services (P2)

| Service | Commands | Priority |
|---------|----------|----------|
| **cloudformation** | create-stack, delete-stack, describe-stacks, list-stacks | P2 |
| **ecs** | list-clusters, describe-services, run-task, list-tasks | P2 |
| **eks** | list-clusters, describe-cluster, update-kubeconfig | P2 |
| **rds** | describe-db-instances, create-db-instance, delete-db-instance | P2 |
| **route53** | list-hosted-zones, change-resource-record-sets | P2 |
| **secretsmanager** | get-secret-value, create-secret, update-secret | P2 |

#### AWS CLI Examples (Exact Matching)

```bash
# These should work identically to official AWS CLI
omni aws s3 ls
omni aws s3 ls s3://my-bucket/prefix/ --recursive
omni aws s3 cp file.txt s3://my-bucket/ --acl public-read
omni aws s3 sync ./local s3://my-bucket/backup --delete
omni aws s3 presign s3://my-bucket/file.txt --expires-in 3600

omni aws ec2 describe-instances --filters "Name=instance-state-name,Values=running"
omni aws ec2 describe-instances --query "Reservations[*].Instances[*].[InstanceId,State.Name]" --output table

omni aws lambda invoke --function-name myFunc --payload '{"key":"value"}' output.json
omni aws dynamodb scan --table-name MyTable --filter-expression "status = :s" --expression-attribute-values '{":s":{"S":"active"}}'

omni aws cloudformation deploy --template-file template.yaml --stack-name mystack --capabilities CAPABILITY_IAM
```

---

### Azure CLI – Complete Service Coverage

#### Authentication & Account (P0)

| Command | Description | Status |
|---------|-------------|--------|
| `az login` | Interactive login (browser) | |
| `az login --service-principal` | Service principal login | |
| `az logout` | Log out | |
| `az account list` | List subscriptions | |
| `az account set` | Set active subscription | |
| `az account show` | Show current subscription | |

#### Resource Management (P0)

| Command | Description | Status |
|---------|-------------|--------|
| `az group list` | List resource groups | |
| `az group create` | Create resource group | |
| `az group delete` | Delete resource group | |
| `az resource list` | List resources | |
| `az tag list` | List tags | |

#### Compute (P1)

| Command | Description | Status |
|---------|-------------|--------|
| `az vm list` | List VMs | |
| `az vm create` | Create VM | |
| `az vm start` | Start VM | |
| `az vm stop` | Stop VM | |
| `az vm delete` | Delete VM | |
| `az vm show` | Show VM details | |
| `az vmss list` | List VM scale sets | |

#### Storage (P1)

| Command | Description | Status |
|---------|-------------|--------|
| `az storage account list` | List storage accounts | |
| `az storage account create` | Create storage account | |
| `az storage container list` | List blob containers | |
| `az storage blob list` | List blobs | |
| `az storage blob upload` | Upload blob | |
| `az storage blob download` | Download blob | |
| `az storage blob delete` | Delete blob | |

#### Key Vault (P1)

| Command | Description | Status |
|---------|-------------|--------|
| `az keyvault list` | List key vaults | |
| `az keyvault secret list` | List secrets | |
| `az keyvault secret show` | Get secret value | |
| `az keyvault secret set` | Set secret | |
| `az keyvault key list` | List keys | |

#### Kubernetes (P2)

| Command | Description | Status |
|---------|-------------|--------|
| `az aks list` | List AKS clusters | |
| `az aks create` | Create AKS cluster | |
| `az aks get-credentials` | Get kubeconfig | |
| `az aks scale` | Scale node pool | |

#### Azure CLI Examples (Exact Matching)

```bash
# These should work identically to official Azure CLI
omni az login
omni az account list --output table
omni az group create --name myRG --location eastus

omni az vm create --resource-group myRG --name myVM --image Ubuntu2204 --admin-username azureuser --generate-ssh-keys
omni az vm list --output table --query "[].{Name:name, State:powerState}"

omni az storage account create --name mystorageacct --resource-group myRG --sku Standard_LRS
omni az storage blob upload --account-name mystorageacct --container-name mycontainer --file ./local.txt --name remote.txt

omni az keyvault secret set --vault-name myVault --name mySecret --value "secret-value"
omni az keyvault secret show --vault-name myVault --name mySecret --query value -o tsv
```

---

### GCP gcloud – Complete Service Coverage

#### Authentication & Config (P0)

| Command | Description | Status |
|---------|-------------|--------|
| `gcloud auth login` | Interactive login | |
| `gcloud auth activate-service-account` | Service account auth | |
| `gcloud auth list` | List accounts | |
| `gcloud config set` | Set config property | |
| `gcloud config list` | List config | |
| `gcloud projects list` | List projects | |

#### Compute Engine (P1)

| Command | Description | Status |
|---------|-------------|--------|
| `gcloud compute instances list` | List instances | |
| `gcloud compute instances create` | Create instance | |
| `gcloud compute instances start` | Start instance | |
| `gcloud compute instances stop` | Stop instance | |
| `gcloud compute instances delete` | Delete instance | |
| `gcloud compute zones list` | List zones | |
| `gcloud compute regions list` | List regions | |

#### Cloud Storage (P1)

| Command | Description | Status |
|---------|-------------|--------|
| `gcloud storage buckets list` | List buckets | |
| `gcloud storage buckets create` | Create bucket | |
| `gcloud storage ls` | List objects | |
| `gcloud storage cp` | Copy objects | |
| `gcloud storage rm` | Remove objects | |
| `gcloud storage cat` | Output object contents | |

#### GKE (P2)

| Command | Description | Status |
|---------|-------------|--------|
| `gcloud container clusters list` | List GKE clusters | |
| `gcloud container clusters create` | Create cluster | |
| `gcloud container clusters get-credentials` | Get kubeconfig | |
| `gcloud container clusters delete` | Delete cluster | |

#### IAM (P1)

| Command | Description | Status |
|---------|-------------|--------|
| `gcloud iam service-accounts list` | List service accounts | |
| `gcloud iam service-accounts create` | Create service account | |
| `gcloud iam service-accounts keys create` | Create key | |
| `gcloud iam roles list` | List roles | |

#### Cloud Functions / Cloud Run (P2)

| Command | Description | Status |
|---------|-------------|--------|
| `gcloud functions list` | List functions | |
| `gcloud functions deploy` | Deploy function | |
| `gcloud run services list` | List Cloud Run services | |
| `gcloud run deploy` | Deploy to Cloud Run | |

#### GCP gcloud Examples (Exact Matching)

```bash
# These should work identically to official gcloud CLI
omni gcloud auth login
omni gcloud config set project my-project
omni gcloud projects list --format="table(projectId,name,projectNumber)"

omni gcloud compute instances list --filter="status=RUNNING"
omni gcloud compute instances create my-vm --zone=us-central1-a --machine-type=e2-medium --image-family=debian-11 --image-project=debian-cloud

omni gcloud storage buckets create gs://my-bucket --location=us
omni gcloud storage cp ./local.txt gs://my-bucket/
omni gcloud storage ls gs://my-bucket/ --recursive

omni gcloud container clusters create my-cluster --zone=us-central1-a --num-nodes=3
omni gcloud container clusters get-credentials my-cluster --zone=us-central1-a

omni gcloud iam service-accounts create my-sa --display-name="My Service Account"
```

---

---

### Security Profile Architecture (Based on clonr)

**Goal:** Secure credential storage for all cloud providers with hardware-backed encryption.

#### Profile Model

```go
// internal/cli/cloud/profile/model.go
type CloudProfile struct {
    Name           string       `json:"name"`            // "my-aws-prod"
    Provider       string       `json:"provider"`        // "aws", "azure", "gcp"
    Region         string       `json:"region"`          // "us-east-1"
    AccountID      string       `json:"account_id"`      // AWS account, Azure sub, GCP project
    RoleArn        string       `json:"role_arn,omitempty"`   // AWS assume-role
    TenantID       string       `json:"tenant_id,omitempty"`  // Azure
    SourceProfile  string       `json:"source_profile,omitempty"` // For role chaining
    TokenStorage   TokenStorage `json:"token_storage"`   // "encrypted" or "open"
    Default        bool         `json:"default"`
    CreatedAt      time.Time    `json:"created_at"`
    LastUsedAt     time.Time    `json:"last_used_at"`
    CacheExpiry    *time.Time   `json:"cache_expiry,omitempty"`
}

type TokenStorage string
const (
    TokenStorageEncrypted TokenStorage = "encrypted"
    TokenStorageOpen      TokenStorage = "open"       // Fallback when TPM unavailable
)
```

#### File-Based Storage Structure

```
~/.omni/
├── config.json                    # Global config + default profiles
├── master.key                     # TPM-sealed or software master key
└── profiles/
    ├── aws/
    │   ├── default.json           # Profile metadata (no secrets)
    │   ├── default.enc            # Encrypted credentials
    │   ├── prod.json
    │   ├── prod.enc
    │   └── staging.json
    ├── azure/
    │   ├── dev.json
    │   ├── dev.enc
    │   └── prod.json
    └── gcp/
        ├── myproject.json
        └── myproject.enc

# File permissions: 0600 (user read/write only)
```

#### Profile File Format

```json
// ~/.omni/profiles/aws/prod.json (metadata - no secrets)
{
  "name": "prod",
  "provider": "aws",
  "region": "us-east-1",
  "account_id": "123456789012",
  "role_arn": "",
  "source_profile": "",
  "token_storage": "encrypted",
  "default": true,
  "created_at": "2026-02-04T10:00:00Z",
  "last_used_at": "2026-02-04T18:00:00Z"
}

// ~/.omni/profiles/aws/prod.enc (encrypted credentials)
// Binary: ENC:<nonce:12><ciphertext><tag:16>
// Contains encrypted JSON of AWSCredentials struct
```

#### Global Config

```json
// ~/.omni/config.json
{
  "version": 1,
  "defaults": {
    "aws": "prod",
    "azure": "dev",
    "gcp": "myproject"
  },
  "key_rotation_days": 30,
  "cache_ttl_seconds": 3600
}
```

#### Master Key Storage

```go
// ~/.omni/master.key format
// TPM-sealed: binary blob from TPM seal operation
// Software fallback: PBKDF2-derived key from machine ID

type MasterKeyFile struct {
    Version   int       `json:"version"`
    KeyType   string    `json:"key_type"`   // "tpm", "software"
    SealedKey []byte    `json:"sealed_key"` // Encrypted master key
    CreatedAt time.Time `json:"created_at"`
    RotatedAt time.Time `json:"rotated_at"`
}
```

#### Credential Encryption

```go
// internal/cli/cloud/crypto/crypto.go

// Multi-layer encryption:
// 1. TPM-sealed master key (hardware-backed when available)
// 2. Per-profile derived key: SHA256(master || provider:profile_name)
// 3. AES-256-GCM encryption with random 12-byte nonce

// Token prefixes for storage type detection
const (
    OpenPrefix = "OPEN:"  // Plaintext fallback (no TPM)
    EncPrefix  = "ENC:"   // AES-256-GCM encrypted
)

func EncryptCredentials(creds []byte, profileName, provider string) ([]byte, error)
func DecryptCredentials(encrypted []byte, profileName, provider string) ([]byte, error)

// Per-profile key derivation (isolation: one profile compromise doesn't expose others)
func deriveKey(masterKey []byte, profileName, provider string) []byte {
    suffix := []byte(profileName + ":" + provider)
    data := append(masterKey, suffix...)
    hash := sha256.Sum256(data)
    return hash[:]
}
```

#### Provider-Specific Credentials

```go
// AWS credentials
type AWSCredentials struct {
    AccessKeyID     string `json:"access_key_id"`
    SecretAccessKey string `json:"secret_access_key"`
    SessionToken    string `json:"session_token,omitempty"`  // For STS
    RoleArn         string `json:"role_arn,omitempty"`       // For assume-role
}

// Azure credentials
type AzureCredentials struct {
    TenantID       string `json:"tenant_id"`
    ClientID       string `json:"client_id"`
    ClientSecret   string `json:"client_secret,omitempty"`
    Certificate    []byte `json:"certificate,omitempty"`
    SubscriptionID string `json:"subscription_id"`
}

// GCP credentials
type GCPCredentials struct {
    Type           string `json:"type"`                     // "service_account"
    ProjectID      string `json:"project_id"`
    PrivateKeyID   string `json:"private_key_id"`
    PrivateKey     string `json:"private_key"`
    ClientEmail    string `json:"client_email"`
    ClientID       string `json:"client_id"`
    // ... full service account JSON structure
}
```

#### Profile Management Commands

```bash
# Profile CRUD
omni cloud profile add <name> --provider <aws|azure|gcp> [options]
omni cloud profile list [--provider <provider>]
omni cloud profile show <name>
omni cloud profile use <name>               # Set as default
omni cloud profile delete <name>
omni cloud profile validate <name>          # Check credential validity

# AWS-specific profile creation
omni cloud profile add myaws --provider aws --region us-east-1
  # Interactive: prompts for Access Key ID and Secret Access Key

omni cloud profile add myaws --provider aws \
    --access-key-id AKIA... \
    --secret-access-key ...        # Direct input (use with caution)

omni cloud profile add myaws-role --provider aws \
    --role-arn arn:aws:iam::123456789:role/MyRole \
    --source-profile myaws         # Assume role from another profile

# Azure-specific profile creation
omni cloud profile add myazure --provider azure
  # Launches browser for interactive login (OAuth device flow)

omni cloud profile add myazure-sp --provider azure \
    --tenant-id ... --client-id ... --client-secret ...  # Service principal

# GCP-specific profile creation
omni cloud profile add mygcp --provider gcp
  # Launches browser for interactive login

omni cloud profile add mygcp-sa --provider gcp \
    --service-account-key /path/to/key.json  # Service account

# Credential rotation
omni cloud profile rotate <name>            # Force rotation
omni cloud profile cache clear              # Clear cached STS tokens

# Environment variable support
export OMNI_CLOUD_PROFILE=myaws             # Override default
omni aws s3 ls                              # Uses myaws profile
omni aws s3 ls --profile other-profile      # Explicit override
```

#### Profile Service Architecture

```go
// internal/cli/cloud/profile/service.go
type ProfileService struct {
    baseDir string        // ~/.omni
    crypto  CryptoService
}

// File operations
func (ps *ProfileService) AddProfile(profile *CloudProfile, creds any) error {
    // 1. Write profile metadata to ~/.omni/profiles/{provider}/{name}.json
    // 2. Encrypt credentials and write to ~/.omni/profiles/{provider}/{name}.enc
    // 3. Set file permissions to 0600
}

func (ps *ProfileService) GetProfile(name, provider string) (*CloudProfile, error) {
    // Read from ~/.omni/profiles/{provider}/{name}.json
}

func (ps *ProfileService) GetProfileCredentials(name, provider string) (any, error) {
    // 1. Read ~/.omni/profiles/{provider}/{name}.enc
    // 2. Decrypt with profile-specific derived key
    // 3. Return typed credentials (AWSCredentials, AzureCredentials, etc.)
}

func (ps *ProfileService) GetDefaultProfile(provider string) (*CloudProfile, error) {
    // 1. Read ~/.omni/config.json for default profile name
    // 2. Load that profile
}

func (ps *ProfileService) SetDefaultProfile(name, provider string) error {
    // Update ~/.omni/config.json defaults section
}

func (ps *ProfileService) DeleteProfile(name, provider string) error {
    // Remove both .json and .enc files
}

func (ps *ProfileService) ListProfiles(provider string) ([]*CloudProfile, error) {
    // List all .json files in ~/.omni/profiles/{provider}/
    // Or all providers if provider is empty
}

func (ps *ProfileService) ValidateProfile(name, provider string) error
func (ps *ProfileService) RotateCredentials(name, provider string, newCreds any) error
```

#### File Store Implementation

```go
// internal/cli/cloud/profile/store.go
type FileStore struct {
    baseDir string  // ~/.omni
}

func NewFileStore() (*FileStore, error) {
    home, _ := os.UserHomeDir()
    baseDir := filepath.Join(home, ".omni")

    // Create directory structure
    for _, provider := range []string{"aws", "azure", "gcp"} {
        dir := filepath.Join(baseDir, "profiles", provider)
        os.MkdirAll(dir, 0700)
    }

    return &FileStore{baseDir: baseDir}, nil
}

func (fs *FileStore) profilePath(provider, name string) string {
    return filepath.Join(fs.baseDir, "profiles", provider, name+".json")
}

func (fs *FileStore) credentialsPath(provider, name string) string {
    return filepath.Join(fs.baseDir, "profiles", provider, name+".enc")
}

func (fs *FileStore) SaveProfile(profile *CloudProfile) error {
    path := fs.profilePath(profile.Provider, profile.Name)
    data, _ := json.MarshalIndent(profile, "", "  ")
    return os.WriteFile(path, data, 0600)  // User read/write only
}

func (fs *FileStore) SaveCredentials(provider, name string, encrypted []byte) error {
    path := fs.credentialsPath(provider, name)
    return os.WriteFile(path, encrypted, 0600)
}

func (fs *FileStore) LoadProfile(provider, name string) (*CloudProfile, error) {
    path := fs.profilePath(provider, name)
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var profile CloudProfile
    json.Unmarshal(data, &profile)
    return &profile, nil
}

func (fs *FileStore) LoadCredentials(provider, name string) ([]byte, error) {
    path := fs.credentialsPath(provider, name)
    return os.ReadFile(path)
}

func (fs *FileStore) DeleteProfile(provider, name string) error {
    os.Remove(fs.profilePath(provider, name))
    os.Remove(fs.credentialsPath(provider, name))
    return nil
}

func (fs *FileStore) ListProfiles(provider string) ([]string, error) {
    dir := filepath.Join(fs.baseDir, "profiles", provider)
    entries, _ := os.ReadDir(dir)
    var names []string
    for _, e := range entries {
        if strings.HasSuffix(e.Name(), ".json") {
            names = append(names, strings.TrimSuffix(e.Name(), ".json"))
        }
    }
    return names, nil
}
```

#### Security Features

| Feature | Description |
|---------|-------------|
| **TPM-sealed master key** | Hardware-backed encryption when available |
| **Per-profile isolation** | Compromise of one profile doesn't expose others |
| **AES-256-GCM** | Authenticated encryption with integrity check |
| **Automatic rotation** | Configurable credential rotation interval |
| **Secure deletion** | Keystore cascade delete when profile removed |
| **Token caching** | Cache STS tokens with automatic refresh |
| **Audit trail** | Timestamps for creation, access, rotation |
| **Software fallback** | Encrypted storage even without TPM |

#### Integration with CLI Commands

```go
// Example: AWS command using profile
func runS3Ls(cmd *cobra.Command, args []string) error {
    // 1. Get profile (from flag, env var, or default)
    profileName := getProfileName(cmd)  // --profile flag or OMNI_CLOUD_PROFILE

    // 2. Load credentials (auto-decrypted)
    profile, creds, err := profileService.GetProfileWithCredentials(profileName)
    if err != nil {
        return err
    }

    // 3. Create AWS config with credentials
    cfg, err := awscommon.LoadConfigWithCredentials(ctx, creds.(*AWSCredentials))

    // 4. Execute command
    client := s3.NewClient(cfg, ...)
    return client.Ls(ctx, args[0], opts)
}
```

---

### Cloud CLI Architecture

```
internal/cli/
├── cloud/                  # Shared cloud infrastructure
│   ├── profile/            # Profile management
│   │   ├── model.go        # CloudProfile struct
│   │   ├── service.go      # ProfileService (business logic)
│   │   ├── store.go        # File-based storage operations
│   │   └── cli.go          # Interactive profile selector (TUI)
│   ├── crypto/             # Encryption (from clonr pattern)
│   │   ├── crypto.go       # AES-256-GCM encrypt/decrypt
│   │   ├── master.go       # Master key management
│   │   ├── tpm.go          # TPM-sealed master key (when available)
│   │   └── software.go     # Software fallback (PBKDF2)
│   ├── output/             # Output formatting
│   │   ├── format.go       # JSON, table, yaml, tsv formatters
│   │   └── query.go        # JMESPath support (--query)
│   └── cache/              # Token caching
│       └── cache.go        # STS/temporary token cache

# User data directory
~/.omni/
├── config.json             # Global config + default profiles
├── master.key              # Encrypted master key
├── cache/                  # Cached STS tokens
│   └── aws_prod_sts.json   # Temporary tokens with expiry
└── profiles/
    ├── aws/
    │   ├── default.json    # Profile metadata
    │   └── default.enc     # Encrypted credentials
    ├── azure/
    └── gcp/
├── aws/                    # AWS SDK v2
│   ├── aws.go              # Config, credentials, output formatter
│   ├── query.go            # JMESPath query support
│   ├── s3/                 # S3 + S3API operations ✅
│   ├── ec2/                # EC2 operations ✅
│   ├── iam/                # IAM operations ✅
│   ├── sts/                # STS operations ✅
│   ├── ssm/                # SSM operations ✅
│   ├── lambda/             # Lambda operations
│   ├── dynamodb/           # DynamoDB operations
│   └── ...
├── azure/                  # Azure SDK for Go
│   ├── azure.go            # Config, auth, output formatter
│   ├── query.go            # JMESPath query support
│   ├── account/            # Account/subscription management
│   ├── group/              # Resource groups
│   ├── vm/                 # Virtual machines
│   ├── storage/            # Blob storage
│   ├── keyvault/           # Key Vault
│   └── aks/                # AKS
└── gcp/                    # Google Cloud SDK for Go
    ├── gcp.go              # Config, auth, output formatter
    ├── filter.go           # Resource filter support
    ├── auth/               # Authentication
    ├── compute/            # Compute Engine
    ├── storage/            # Cloud Storage
    ├── container/          # GKE
    └── iam/                # IAM

cmd/
├── aws.go                  # aws command root
├── aws_*.go                # aws subcommands (per service)
├── az.go                   # az command root
├── az_*.go                 # az subcommands (per group)
├── gcloud.go               # gcloud command root
└── gcloud_*.go             # gcloud subcommands (per group)
```

### Output Format Support

```go
// internal/cli/cloud/output.go
type OutputFormat string

const (
    OutputJSON    OutputFormat = "json"
    OutputJSONC   OutputFormat = "jsonc"   // Azure: colored JSON
    OutputText    OutputFormat = "text"
    OutputTable   OutputFormat = "table"
    OutputTSV     OutputFormat = "tsv"
    OutputYAML    OutputFormat = "yaml"
    OutputYAMLC   OutputFormat = "yamlc"   // Azure: colored YAML
    OutputCSV     OutputFormat = "csv"     // GCP
    OutputValue   OutputFormat = "value"   // GCP: plain values
)

// JMESPath query support
func QueryJSON(data any, query string) (any, error)
```

### Testing with Local Emulators

| Cloud | Emulator | Docker Image | Services |
|-------|----------|--------------|----------|
| AWS | LocalStack | `localstack/localstack` ✅ | S3, EC2, IAM, STS, SSM, Lambda, DynamoDB, SNS, SQS |
| Azure | Azurite | `mcr.microsoft.com/azure-storage/azurite` | Blob, Queue, Table storage |
| Azure | CosmosDB | `mcr.microsoft.com/cosmosdb/linux/azure-cosmos-emulator` | CosmosDB |
| GCP | Storage | `fsouza/fake-gcs-server` | Cloud Storage |
| GCP | Pub/Sub | `google/cloud-sdk` with emulator | Pub/Sub |
| GCP | Firestore | `google/cloud-sdk` with emulator | Firestore |

### Docker Compose for Testing

```yaml
# test/cloud/docker-compose.yml
services:
  localstack:
    image: localstack/localstack:latest
    ports:
      - "4566:4566"
    environment:
      - SERVICES=s3,ec2,iam,sts,ssm,lambda,dynamodb,sns,sqs

  azurite:
    image: mcr.microsoft.com/azure-storage/azurite
    ports:
      - "10000:10000"  # Blob
      - "10001:10001"  # Queue
      - "10002:10002"  # Table

  fake-gcs:
    image: fsouza/fake-gcs-server
    ports:
      - "4443:4443"
    command: ["-scheme", "http"]
```

### Implementation Priority

| Phase | Cloud | Services | Priority |
|-------|-------|----------|----------|
| 11.1 | AWS | Complete S3, S3API, Lambda, DynamoDB | P0 |
| 11.2 | AWS | SNS, SQS, CloudWatch, Logs | P1 |
| 11.3 | AWS | CloudFormation, ECS, EKS, RDS | P2 |
| 11.4 | Azure | Auth, Account, Group, VM | P1 |
| 11.5 | Azure | Storage, KeyVault | P1 |
| 11.6 | Azure | AKS, SQL, CosmosDB | P2 |
| 11.7 | GCP | Auth, Config, Projects | P1 |
| 11.8 | GCP | Compute, Storage | P1 |
| 11.9 | GCP | GKE, IAM, Functions | P2 |

---

## Test Coverage

**Current:** 30.9% (overall, includes vendored buf packages) | **Target:** 80%

### Omni-owned pkg/ Packages

| Package | Coverage | Status |
|---------|----------|--------|
| pkg/encoding | 100.0% | Excellent |
| pkg/twig/models | 100.0% | Excellent |
| pkg/twig/expander | 98.1% | Excellent |
| pkg/video/m3u8 | 96.8% | Excellent |
| pkg/twig/comparer | 96.3% | Excellent |
| pkg/textutil/diff | 95.2% | Excellent |
| pkg/textutil | 93.7% | Excellent |
| pkg/idgen | 91.4% | Excellent |
| pkg/search/rg | 90.4% | Excellent |
| pkg/hashutil | 88.5% | Good |
| pkg/cssfmt | 87.3% | Good |
| pkg/cryptutil | 85.3% | Good |
| pkg/figlet | 82.9% | Good |
| pkg/twig/scanner | 82.4% | Good |
| pkg/pipeline | 81.6% | Good |
| pkg/twig/formatter | 80.4% | Good |
| pkg/sqlfmt | 79.2% | Acceptable |
| pkg/htmlfmt | 77.9% | Acceptable |
| pkg/search/grep | 77.4% | Acceptable |
| pkg/jsonutil | 67.5% | Needs improvement |
| pkg/video/utils | 58.4% | Needs improvement |
| pkg/video/format | 50.0% | Needs improvement |
| pkg/twig | 44.3% | Needs improvement |
| pkg/video/extractor | 5.7% | No meaningful tests |
| pkg/video (root) | 0.0% | No tests |
| pkg/video/cache | 0.0% | No tests |
| pkg/video/downloader | 0.0% | No tests |
| pkg/video/jsinterp | 0.0% | No tests |
| pkg/video/nethttp | 0.0% | No tests |
| pkg/twig/builder | 0.0% | No tests |
| pkg/twig/parser | 0.0% | No tests |
