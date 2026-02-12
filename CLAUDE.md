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
│   └── video/          # Video download engine (YouTube, HLS, generic)
├── internal/cli/       # CLI wrappers (I/O, flags, stdin handling)
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

### Implemented (155+ commands)

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
| **Code Gen** | generate handler, generate repository, generate test |
| **Security** | encrypt, decrypt, uuid, random, jwt decode |
| **Pagers** | less, more |
| **Comparison** | diff |
| **Tooling** | lint, cmdtree, loc, cron, project (info, deps, docs, git, health) |
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

Current coverage: ~30.5% overall, 51.6% omni-owned (~75% avg for pkg/)

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
- `testing/golden/golden_tests.yaml` — Declarative test registry (81 tests, 11 categories)
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
