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
├── pkg/                # Reusable Go libraries (importable by external projects)
│   ├── idgen/          # UUID, ULID, KSUID, Nanoid, Snowflake
│   ├── hashutil/       # MD5, SHA1, SHA256, SHA512, CRC32, CRC64 hashing
│   ├── jsonutil/       # jq-style JSON query engine
│   ├── encoding/       # Base64, Base32, Base58 encode/decode
│   ├── cryptutil/      # AES-256-GCM encrypt/decrypt
│   ├── sqlfmt/         # SQL format/minify/validate
│   ├── cssfmt/         # CSS format/minify/validate
│   ├── htmlfmt/        # HTML format/minify/validate
│   ├── textutil/       # Sort, Uniq, Trim + diff/
│   ├── search/grep/    # Pattern search with options
│   ├── search/rg/      # Gitignore parsing, file type matching
│   ├── pipeline/       # Streaming text processing engine
│   ├── figlet/         # FIGlet font parser and ASCII art
│   ├── twig/           # Tree scanning, formatting, comparison
│   ├── userdirs/       # XDG user directory paths
│   └── video/          # Video download engine (YouTube, HLS, generic)
├── internal/cli/       # CLI wrappers (I/O, flags, stdin handling)
│   ├── cmderr/         # Unified error model (sentinels, exit codes)
│   ├── command/        # Unified Command interface + Registry + adapters
│   ├── repo/           # Repository analyzer (LLM-optimized context generation)
│   ├── scaffolding/    # Code scaffolding (cobra, handler, repository, testgen)
│   │   ├── cobra/      # Cobra CLI project generator + config
│   │   ├── handler/    # HTTP handler generator
│   │   ├── repository/ # Database repository generator
│   │   └── testgen/    # Go test generator
│   ├── <command>/      # Each command delegates to pkg/ for core logic
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

#### Args Preprocessor Pattern (GNU Compatibility)

Commands that need Unix-style flag compatibility use `os.Args` preprocessing in `init()`, before Cobra parses flags:

- **`head`/`tail`**: Convert `-NUM` to `-n NUM` (e.g., `head -20` → `head -n 20`)
  - Uses regex `^-(\d+)$` to detect numeric flags
  - Implemented in `cmd/head.go:preprocessHeadArgs()` and `cmd/tail.go:preprocessTailArgs()`

- **`find`**: Convert single-dash GNU flags to double-dash (e.g., `-name` → `--name`)
  - Uses `knownFindFlags` map in `cmd/find.go` listing all valid find flags
  - Implemented in `cmd/find.go:preprocessFindArgs()`
  - Supports: `-name`, `-iname`, `-path`, `-ipath`, `-regex`, `-iregex`, `-type`, `-size`, `-mindepth`, `-maxdepth`, `-mtime`, `-mmin`, `-atime`, `-amin`, `-empty`, `-executable`, `-readable`, `-writable`, `-print0`, `-not`, `-json`

**Pattern**: Check if command name is in `os.Args`, then rewrite args before Cobra's `Execute()`.

#### Pkg Library Pattern

Reusable logic lives in `pkg/` with functional options:

```go
// pkg/sqlfmt/sqlfmt.go
type Option func(*Options)
func WithIndent(s string) Option { return func(o *Options) { o.Indent = s } }
func Format(input string, opts ...Option) string { ... }
```

`internal/cli/` packages delegate to `pkg/` and handle I/O:

```go
// internal/cli/sqlfmt/sqlfmt.go
import pkgsql "github.com/inovacc/omni/pkg/sqlfmt"

func RunFormat(w io.Writer, args []string, opts Options) error {
    result := pkgsql.Format(input, pkgsql.WithIndent(opts.Indent))
    _, _ = fmt.Fprintln(w, result)
    return nil
}
```

#### Unified Command Interface

All new commands should implement the `Command` interface from `internal/cli/command/`:

```go
// Command is the interface all CLI commands should implement.
type Command interface {
    Run(ctx context.Context, w io.Writer, r io.Reader, args []string) error
}
```

**Registry** maps command names to implementations (thread-safe):
```go
reg := command.NewRegistry()
reg.Register("head", command.AdaptWriterReaderArgs(
    func(w io.Writer, r io.Reader, args []string) error {
        return head.RunHead(w, r, args, head.HeadOptions{Lines: 10})
    },
))
cmd, ok := reg.Get("head")
```

**Adapters** bridge existing Run signatures to the Command interface:
- `AdaptWriterArgs(fn)` — for `func(io.Writer, []string) error` (hash, base, archive)
- `AdaptWriterReaderArgs(fn)` — for `func(io.Writer, io.Reader, []string) error` (head, tail, sort)
- `AdaptFull(fn)` — for `func(context.Context, io.Writer, io.Reader, []string) error`

**Migration:** Incrementally adopt — wrap existing Run functions with adapters, register in the Registry. The `pipe` command can use the Registry to dispatch commands.

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

#### Error Handling (cmderr)

All commands should use `internal/cli/cmderr` sentinels for error classification.
The root command (`cmd/root.go`) maps these to exit codes via `cmderr.ExitCodeFor()`.

**Sentinels → Exit Codes:**
| Sentinel | Exit Code | Use For |
|----------|-----------|---------|
| `cmderr.ErrNotFound` | 1 | File/resource not found |
| `cmderr.ErrConflict` | 1 | Verification failures, sort disorder |
| `cmderr.ErrInvalidInput` | 2 | Bad flags, missing operands, parse errors |
| `cmderr.ErrPermission` | 3 | Permission denied |
| `cmderr.ErrIO` | 4 | I/O errors |
| `cmderr.ErrTimeout` | 5 | Timeouts |
| `cmderr.ErrUnsupported` | 6 | Unsupported operations |

**Pattern — classify errors from os/io:**
```go
if errors.Is(err, os.ErrNotExist) {
    return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("head: %s", err))
}
if errors.Is(err, os.ErrPermission) {
    return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("head: %s", err))
}
return fmt.Errorf("head: %w", err) // fallback for unclassified errors
```

**Pattern — validation errors:**
```go
return cmderr.Wrap(cmderr.ErrInvalidInput, "path clean: missing operand")
```

**Pattern — silent exit (grep-style):**
```go
return cmderr.SilentExit(1) // no message, just exit code
```

**Commands adopted (84):** cat, curl, crypt, diff, grep, find, fs, jq, ls, sed, head, tail, text (sort/uniq), hash, path, archive, base, xxd, yq, buf (build/format/lint), bzip2, xz, env, kill, ps, df, du, dotenv, free, pipe, chown, rg, pipeline, file, which, shuf, readlink, sqlite, bbolt, pager, join, cmp, comm, cron, loc, lint, seq, sleep, strings, basename, dirname, realpath, whoami, uptime, id, awk, column, nl, paste, banner, rev, tac, fold, cut, printf, hexenc, urlenc, tr, xargs, watch, yes, uuid, random, caseconv, jwt, note, jsonfmt, htmlenc, tomlutil, xmlfmt, pwd, exist

**Commands NOT yet adopted:** ~76 remaining — adopt in future batches following the same pattern.

**General rules:**
- Always wrap errors with context: `fmt.Errorf("command: %w", err)`
- Write informational errors to stderr: `fmt.Fprintf(os.Stderr, "error: %v\n", err)`
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

### Implemented (160+ commands)

| Category | Commands |
|----------|----------|
| **Core** | ls, pwd, cat, date, dirname, basename, realpath, path (clean, abs), tree, arch, sleep, seq, printf, for |
| **File** | cp, mv, rm, mkdir, rmdir, touch, stat, ln, readlink, chmod, chown, find, dd, file, which |
| **Text** | grep, egrep, fgrep, head, tail, sort, uniq, wc, cut, tr, nl, paste, tac, column, fold, join, sed, awk, shuf, split, rev, comm, cmp, strings |
| **Search** | rg (ripgrep-style search with gitignore support, parallel walking, streaming JSON) |
| **System** | env, whoami, id, uname, uptime, free, df, du, ps, kill, time |
| **Flow** | xargs, watch, yes, pipe, pipeline |
| **Archive** | tar, zip, unzip |
| **Compression** | gzip, gunzip, zcat, bzip2, bunzip2, bzcat, xz, unxz, xzcat |
| **Hash** | hash, sha256sum, sha512sum, md5sum, crc32sum, crc64sum |
| **Encoding** | base64, base32, base58, url encode/decode, html encode/decode, hex encode/decode, xxd |
| **Data** | jq, yq, dotenv, json (tostruct, tocsv, fromcsv, toxml, fromxml), yaml tostruct, yaml validate, toml validate, xml (validate, tojson, fromjson) |
| **Formatting** | sql fmt/minify/validate, html fmt/minify/validate, css fmt/minify/validate |
| **Protobuf** | buf lint, buf format, buf compile, buf breaking, buf generate, buf mod init/update, buf ls-files |
| **Code Gen** | scaffold cobra init/add/add-tools/config, scaffold handler, scaffold repository, scaffold test |
| **Security** | encrypt, decrypt, uuid, random, jwt decode |
| **Pagers** | less, more |
| **Comparison** | diff |
| **Tooling** | lint, cmdtree, loc, cron, project (info, deps, docs, git, health), repo (analyze) |
| **Network** | curl |
| **Video** | video download, video info, video list-formats, video search, video extractors, video channel |
| **Cloud/DevOps** | kubectl (k), terraform (tf), aws |
| **Git Hacks** | git (quick-commit, branch-clean, undo, amend, stash-staged, log-graph, diff-words, blame-line), gqc, gbc |
| **Checks** | exist (file, dir, path, command, env, process, port) |
| **Kubectl Hacks** | kga, klf, keb, kpf, kdp, krr, kge, ktp, ktn, kcs, kns, kwp, kscale, kdebug, kdrain, krun, kconfig |

### Tree Command (Advanced Features)

The `tree` command includes performance optimizations for large codebases and a JSON compare feature:

**Scanner Optimizations:**
- `--max-files N` - Cap total scanned items (0 = unlimited)
- `--max-hash-size N` - Skip hashing files larger than N bytes (0 = unlimited)
- `-t/--threads N` - Parallel workers (0 = auto/NumCPU, 1 = sequential)
- Progress callback API for library consumers

**Streaming Output:**
- `--json-stream` - NDJSON output (one JSON object per line: begin, node, stats, end)

**JSON Compare:**
- `--compare a.json b.json` - Compare two tree JSON snapshots
- `--detect-moves` - Detect moved files by hash matching (default true)
- 5-phase algorithm: flatten → removed → added → moves → modified
- Output: human-readable (`+`/`-`/`~`/`>` prefixes) or `--json`

**Library API (pkg/twig):**
```go
// Performance options
twig.WithMaxFiles(10000)
twig.WithMaxHashSize(50 * 1024 * 1024) // 50MB
twig.WithParallel(8)
twig.WithProgressCallback(func(n int) { fmt.Printf("\r%d files...", n) })

// Streaming
t.GenerateJSONStream(ctx, path, os.Stdout)

// Compare
comparer.Compare(leftJSON, rightJSON, comparer.CompareConfig{DetectMoves: true})
```

**Package Structure:**
- `pkg/twig/scanner/` - Parallel scanning, MaxFiles, MaxHashSize
- `pkg/twig/formatter/` - NDJSON streaming output
- `pkg/twig/comparer/` - JSON snapshot comparison

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

### Pipeline Command (Streaming Engine)

The `pipeline` command is a streaming text processing engine with built-in transform stages connected via `io.Pipe()` goroutines. Unlike `pipe` (which dispatches through Cobra), `pipeline` has zero Cobra overhead per stage and processes line-by-line for constant memory usage.

**Built-in Stages (20):**

*Streaming (constant memory):*
`grep`, `grep-v`, `contains`, `replace`, `head`/`take`, `skip`, `uniq`, `cut`, `tr`, `sed`, `rev`, `nl`, `tee`, `filter`(lib-only), `map`(lib-only)

*Buffering (reads all input):*
`sort`, `tail`, `tac`, `wc`

**Examples:**
```bash
omni pipeline 'grep error' 'sort' 'uniq' 'head 10' < log.txt
omni pipeline -f access.log 'grep 404' 'cut -d" " -f1' 'sort' 'uniq'
omni pipeline -v 'grep -i warning' 'sort -rn' 'head 5'
```

**Library API (`pkg/pipeline`):**
```go
p := pipeline.New(
    &pipeline.Grep{Pattern: "error", IgnoreCase: true},
    &pipeline.Sort{},
    &pipeline.Uniq{},
    &pipeline.Head{N: 10},
)
err := p.Run(ctx, os.Stdin, os.Stdout)
```

**Package Structure:**
- `pkg/pipeline/stage.go` - Stage interface
- `pkg/pipeline/stages.go` - 20 built-in stage implementations
- `pkg/pipeline/pipeline.go` - Orchestrator with io.Pipe chaining
- `pkg/pipeline/parse.go` - CLI string → Stage parser

### Rg Command (Ripgrep-Style Search)

The `rg` command provides ripgrep-compatible search with gitignore support:

```go
type Options struct {
    IgnoreCase      bool   // -i: case insensitive
    Fixed           bool   // -F: literal string (no regex)
    LineNumber      bool   // -n: show line numbers (default: true)
    Count           bool   // -c: count matches only
    FilesWithMatch  bool   // -l: list files with matches
    FilesWithout    bool   // -L: list files without matches
    Hidden          bool   // --hidden: search hidden files
    NoIgnore        bool   // --no-ignore: don't respect gitignore
    MaxCount        int    // -m: max matches per file
    Context         int    // -C: context lines
    Before          int    // -B: lines before match
    After           int    // -A: lines after match
    Glob            string // -g: glob pattern filter
    Type            string // -t: file type filter
    JSON            bool   // --json: JSON output
    JSONStream      bool   // --json-stream: streaming NDJSON output
    Threads         int    // -j/--threads: parallel workers (0 = auto)
}
```

**Gitignore Support:**
- Loads patterns from `~/.config/git/ignore` (global)
- Loads patterns from `.git/info/exclude` (repo-level)
- Loads patterns from `.gitignore` files (hierarchy)
- Loads patterns from `.ignore` files (ripgrep-specific)
- Supports negation patterns (`!pattern`)
- Supports directory-only patterns (`dir/`)
- Supports double-glob patterns (`**/test.go`)

**Performance Features:**
- Parallel directory walking with configurable workers
- Literal string optimization for `-F` flag (uses strings.Index)
- Streaming JSON output for large result sets

**Examples:**
```bash
# Basic search
omni rg "pattern" .

# Case insensitive, show context
omni rg -i -C 2 "error" ./src

# Fixed string search (faster, no regex)
omni rg -F "fmt.Println" .

# Filter by file type
omni rg -t go "func main" .

# JSON output
omni rg --json "TODO" .

# Streaming JSON (NDJSON) for piping
omni rg --json-stream "pattern" . | jq '.data.lines.text'

# Parallel search with 8 workers
omni rg -j 8 "pattern" /large/codebase

# Search hidden files, ignore gitignore
omni rg --hidden --no-ignore "secret" .
```

### Video Download Engine

The `video` command is a pure Go youtube-dl/yt-dlp port under `pkg/video/`:

```
pkg/video/
├── video.go                    # Client orchestrator (Extract, Download, DownloadInfo)
├── options.go                  # Functional options (WithFormat, WithProxy, etc.)
├── types/types.go              # VideoInfo, Format, Fragment, Thumbnail, Metadata map, etc.
├── extractor/                  # Site-specific metadata extractors
│   ├── extractor.go            # Extractor interface + BaseExtractor helpers
│   ├── registry.go             # Register/Match/Names (init()-based registration)
│   ├── youtube/                # YouTube: InnerTube API, signature decryption, playlists, channels
│   └── generic/                # Fallback: direct URLs, <video> tags, og:video
├── downloader/                 # Download engines
│   ├── http.go                 # HTTPS: resume via Range, .part files, retry, rate-limit
│   └── hls.go                  # HLS/M3U8: segment download, AES-128 decrypt, master→media resolution
├── format/                     # Format sorting + selector ("best", "worst", "best[height<=720]")
├── m3u8/parser.go              # HLS manifest parser (master + media playlists, EXT-X-KEY)
├── nethttp/client.go           # HTTP client (cookie jar, proxy, retries, per-request headers)
├── jsinterp/jsinterp.go        # goja JS runtime (YouTube signature/nsig decryption)
├── cache/cache.go              # Filesystem cache (XDG paths)
└── utils/                      # HTML parse, URL join, filename sanitize, traverse_obj
```

**Key patterns:**
- Extractors register via `init()` + `extractor.Register()`; `extractor/all/all.go` blank-imports all extractors
- YouTube uses InnerTube API with multiple client configs (android_vr first — no PoToken needed)
- HLS downloader auto-resolves master playlists to highest-bandwidth variant
- `pkg/video/` is a standalone library importable by external Go projects
- `--complete` flag: forces `best` format + writes `.md` sidecar (title, link, description) via `WriteMarkdown` option
- Markdown sidecar: `writeMarkdown()` in `pkg/video/video.go` creates `<video_name>.md` alongside the video file
- Channel extractor uses InnerTube browse API (`/youtubei/v1/browse`) with `videosTabParams` for Videos tab
- Channel pagination via continuation tokens from `continuationItemRenderer`
- `video channel` stores metadata in SQLite (`channel.db`) inside channel folder for incremental downloads
- `VideoInfo.Metadata` map holds extractor-specific data (subscriber count, avatar URL, etc.)

**CLI wrappers:** `internal/cli/video/` (download.go, info.go, formats.go, output.go, channel.go, channeldb.go) → `cmd/video.go`

**Dependencies:** `github.com/dop251/goja` (pure Go JS runtime for YouTube signatures)

**Testing:**
- Unit tests: `go test ./pkg/video/...` (format, m3u8, utils, registry — 33+ tests)
- Black-box: `testing/scripts/test_video.py` (omni video info/list-formats/download vs yt-dlp)
- Docker: `task docker:test:video` (comparison tests with yt-dlp in container)

---

## Cloud & DevOps Integrations

### Kubernetes (kubectl / k)

Full kubectl integration via `k8s.io/kubectl` package using local source code.

```bash
omni kubectl get pods          # or: omni k get pods
omni k get pods -A
omni k describe node mynode
omni k logs -f mypod
omni k apply -f manifest.yaml
```

**Source:** `B:\shared\personal\repos\kubernetes\kubectl` (local replace directive)

### Terraform (terraform / tf)

Terraform CLI wrapper for infrastructure management.

```bash
omni tf init
omni tf plan -out=plan.tfplan
omni tf apply -auto-approve
omni tf destroy
omni tf state list
omni tf workspace select prod
```

**Commands:** init, plan, apply, destroy, validate, fmt, output, state (list/show/mv/rm), workspace (list/new/select/delete), import, taint, untaint, refresh, graph, console, providers, get, test, show, version

### Git Hacks

Shortcuts for common Git operations.

| Command | Alias | Description |
|---------|-------|-------------|
| `omni git quick-commit -m "msg"` | `omni gqc -m "msg"` | Stage all + commit |
| `omni git branch-clean` | `omni gbc` | Delete merged branches |
| `omni git undo` | - | Soft reset HEAD~1 |
| `omni git amend` | - | Amend --no-edit |
| `omni git stash-staged` | - | Stash staged only |
| `omni git log-graph` | `omni git lg` | Pretty log with graph |
| `omni git diff-words` | - | Word-level diff |
| `omni git blame-line` | - | Blame line range |
| `omni git status` | `omni git st` | Short status |
| `omni git push` | - | Push (--force-with-lease) |
| `omni git pull-rebase` | `omni git pr` | Pull --rebase |
| `omni git fetch-all` | `omni git fa` | Fetch --all --prune |

### Kubectl Hacks

Shortcuts for common Kubernetes operations.

| Command | Description |
|---------|-------------|
| `omni kga` | Get all resources (pods, svc, deploy, etc.) |
| `omni klf <pod>` | Follow logs with timestamps |
| `omni keb <pod>` | Exec bash into pod (falls back to sh) |
| `omni kpf <target> <local:remote>` | Port forward |
| `omni kdp <selector>` | Delete pods by selector |
| `omni krr <deployment>` | Rollout restart deployment |
| `omni kge` | Get events sorted by time |
| `omni ktp` | Top pods by resource |
| `omni ktn` | Top nodes by resource |
| `omni kcs [context]` | Switch/list contexts |
| `omni kns [namespace]` | Switch/list namespaces |
| `omni kwp` | Watch pods continuously |
| `omni kscale <deploy> <n>` | Scale deployment |
| `omni kdebug <pod>` | Debug with ephemeral container |
| `omni kdrain <node>` | Drain node for maintenance |
| `omni krun <name> --image=<img>` | Run one-off pod |
| `omni kconfig` | Show kubeconfig info |

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

Current coverage: ~25.8% overall (~78% avg for omni-owned pkg/ packages; total skewed by vendored buf packages)

### Test Files

| File | Tests |
|------|-------|
| `internal/cli/hash/hash_test.go` | SHA256, SHA512, MD5, CRC32, CRC64 hashing |
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
| `internal/cli/path/path_test.go` | dirname, basename, realpath, clean, abs |
| `internal/cli/pipe/pipe_test.go` | Command parsing, variable substitution |
| `internal/cli/xxd/xxd_test.go` | Hex dump, reverse, plain, include, bits modes |
| `internal/cli/rg/rg_test.go` | Ripgrep search, parallel walking, streaming JSON, gitignore integration |
| `internal/cli/rg/gitignore_test.go` | Gitignore pattern parsing, negation, directory-only, double globs |
| `pkg/idgen/idgen_test.go` | UUID v4/v7, ULID, KSUID, Nanoid, Snowflake generation |
| `pkg/hashutil/hashutil_test.go` | HashFile, HashReader, HashString, HashBytes, CRC32, CRC64 |
| `pkg/jsonutil/jsonutil_test.go` | Query, QueryString, QueryReader, ApplyFilter |
| `pkg/encoding/encoding_test.go` | Base64/32/58 encode/decode, WrapString |
| `pkg/cryptutil/cryptutil_test.go` | Encrypt/decrypt roundtrip, key generation/derivation |
| `pkg/sqlfmt/sqlfmt_test.go` | Format, Minify, Validate, Tokenize, IsKeyword |
| `pkg/cssfmt/cssfmt_test.go` | Format, Minify, Validate, RemoveComments, ParseDeclarations |
| `pkg/htmlfmt/htmlfmt_test.go` | Format, Minify, Validate, IsSelfClosing, CollapseWhitespace |
| `pkg/textutil/textutil_test.go` | Sort, SortLines, UniqueConsecutive, Uniq, TrimLines |
| `pkg/textutil/diff/diff_test.go` | ComputeDiff, FormatUnified, CompareJSON, CompareJSONBytes |
| `pkg/search/grep/grep_test.go` | Search, SearchWithOptions, CompilePattern |
| `pkg/search/rg/rg_test.go` | ParsePattern, GitignoreSet, MatchesFileType, MatchesGlob, IsBinary |
| `pkg/twig/scanner/scanner_test.go` | Directory scanning, MaxFiles, MaxHashSize, parallel scanning, progress callback |
| `pkg/twig/formatter/formatter_test.go` | ASCII tree, JSON, flattened hash, NDJSON streaming |
| `pkg/twig/comparer/comparer_test.go` | Tree comparison: added, removed, modified, moved, detect-moves |
| `pkg/video/format/format_test.go` | Format sorting, HasVideo, HasAudio, resolution |
| `pkg/video/format/selector_test.go` | Format selector parsing ("best", "worst", filter expressions) |
| `pkg/video/m3u8/parser_test.go` | M3U8 manifest parsing (master/media playlists, segments, keys) |
| `pkg/video/utils/*_test.go` | Sanitize, HTML, URL, parse, traverse tests |
| `pkg/video/extractor/registry_test.go` | Extractor registration and matching |
| `pkg/video/extractor/youtube/channel_test.go` | Channel URL matching, duration/view count parsing |
| `internal/cli/video/channel_test.go` | Channel DB init, schema, incremental tracking, upsert |
| `testing/scripts/test_video.py` | Black-box: omni video vs yt-dlp comparison (Docker) |
| `pkg/pipeline/pipeline_test.go` | Orchestrator, multi-stage, context cancel, head drain |
| `pkg/pipeline/parse_test.go` | CLI string parser, sed expressions, flag combinations |
| `internal/cli/pipeline/pipeline_test.go` | CLI wrapper integration tests |
| `internal/cli/exist/exist_test.go` | File, dir, path, command, env, process, port existence checks, JSON/quiet modes |
| `internal/cli/project/project_test.go` | Project detection, deps parsing, health scoring, output formatting |
| `internal/cli/command/command_test.go` | Command interface, Registry (register, get, names, concurrency), adapters |
| `internal/cli/repo/repo_test.go` | Repo analyze: path resolution, remote detection, tree, key files, entry points, architecture, sections, JSON/Markdown output |
| `pkg/video/downloader/progress_test.go` | SpeedTracker, FormatBytes, FormatSpeed, FormatETA, FormatPercent |
| `pkg/video/downloader/fragment_test.go` | SaveFragmentState/LoadFragmentState roundtrip, RemoveFragmentState, AppendToFile |
| `pkg/video/downloader/downloader_test.go` | SelectDownloader type assertions for all protocols |
| `pkg/video/nethttp/cookies_test.go` | LoadNetscapeCookies parsing, roundtrip, malformed lines, nonexistent file |
| `pkg/video/nethttp/sapisidhash_test.go` | ComputeSAPISIDHash format, ExtractSAPISID from cookie jar |
| `pkg/video/extractor/helpers_test.go` | SearchRegex, ParseJSON, ParseM3U8Formats |
| `pkg/video/options_test.go` | applyOptions defaults, With* option composition |

### Golden Master Tests

Characterization tests that capture exact command outputs as snapshots and detect regressions.

**MUST run on every feature/code change:**

```bash
# Verify outputs match snapshots
task test:golden

# After intentional output changes, regenerate snapshots
task test:golden:update

# Update specific category only
task test:golden:update -- --update encoding

# CI pre-flight: verify all snapshot files exist
task test:golden:check

# List all registered tests
python testing/scripts/test_golden.py --list

# Run with verbose diffs
python testing/scripts/test_golden.py --verbose
```

**Adding golden tests for new commands:**
1. Add test case to `testing/golden/golden_tests.yaml` AND `tools/golden/golden_tests.yaml`
2. Run `task test:golden:update` and `task golden:record` to generate snapshots
3. Review the generated `.stdout` files
4. Commit both the YAML entry and snapshot files

**Full-featured system (tools/golden/):**
```bash
task golden:compare             # Compare with SHA-256 fast path
task golden:record              # Record baselines with manifest
task golden:list                # List test cases
task golden:map:table           # Test map as table
task golden:docker:build        # Build Docker image
task golden:docker:compare      # Compare in container
task golden:docker:record       # Record in container
```

**Lightweight system (testing/golden/):**
```bash
task docker:test:golden          # Verify in Linux container
task docker:test:golden:update   # Regenerate and persist to host via volume mount
```

**Structure:**
- `testing/golden_engine.py` — Lightweight engine (auto-discovered by `run_all.py`)
- `testing/golden/golden_tests.yaml` — Declarative test registry (117 tests, 13 categories)
- `testing/golden/snapshots/` — JSON metadata + .stdout sidecars
- `tools/golden/src/golden/` — Full engine (11 modules: manifest, parallel, map, Docker)
- `tools/golden/golden_tests.yaml` — Shared registry (same tests)
- `tools/golden/golden_masters/` — Baselines with SHA-256 manifest (gitignored)
- See `docs/GOLDEN_MASTER_TESTING.md` for full documentation

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

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->

<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
