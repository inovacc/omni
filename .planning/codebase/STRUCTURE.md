# Codebase Structure

**Analysis Date:** 2026-04-11

## Directory Layout

```
omni/
├── cmd/                    # Cobra CLI command definitions (thin wrappers)
│   ├── root.go             # Root command, logging, error handling, exit codes
│   ├── *.go                # 160+ command files (head.go, tail.go, grep.go, etc.)
│   └── init()              # Args preprocessing for GNU compatibility
├── internal/
│   ├── cli/
│   │   ├── cmderr/         # Error classification & exit code mapping
│   │   ├── command/        # Unified Command interface & Registry
│   │   ├── input/          # Input source abstraction (stdin, files)
│   │   ├── <command>/      # 160+ command packages (one per command)
│   │   │   ├── <cmd>.go    # Options struct & Run() function
│   │   │   ├── <cmd>_test.go
│   │   │   ├── <cmd>_unix.go   # Optional: Unix-specific code
│   │   │   └── <cmd>_windows.go # Optional: Windows-specific code
│   │   ├── pipe/           # Command chaining with variable substitution
│   │   ├── pipeline/       # Streaming text processing (grep, sort, head, etc.)
│   │   ├── repo/           # Repository analyzer for LLM context generation
│   │   ├── scaffolding/    # Code generators (Cobra, handler, repository, test)
│   │   └── logger/         # Structured JSON logging (slog)
│   ├── flags/              # Flag to environment variable export
│   └── version/            # Version info
├── pkg/                    # Reusable Go libraries (importable by external projects)
│   ├── idgen/              # UUID, ULID, KSUID, Nanoid, Snowflake generation
│   ├── hashutil/           # MD5, SHA1, SHA256, SHA512, CRC32, CRC64 hashing
│   ├── jsonutil/           # jq-style JSON query engine
│   ├── encoding/           # Base64, Base32, Base58 encode/decode
│   ├── cryptutil/          # AES-256-GCM encryption/decryption
│   ├── sqlfmt/             # SQL format/minify/validate
│   ├── cssfmt/             # CSS format/minify/validate
│   ├── htmlfmt/            # HTML format/minify/validate
│   ├── textutil/           # Sort, uniq, trim operations
│   │   └── diff/           # Unified diff format & JSON comparison
│   ├── search/
│   │   ├── grep/           # Pattern matching with options
│   │   └── rg/             # Gitignore-aware parallel search
│   ├── pipeline/           # Streaming multi-stage text processing
│   ├── twig/               # Directory tree scanning & comparison
│   │   ├── scanner/        # Parallel directory walking with MaxFiles
│   │   ├── formatter/      # ASCII tree & JSON output
│   │   └── comparer/       # Tree snapshot comparison with move detection
│   ├── figlet/             # FIGlet ASCII art font rendering
│   ├── userdirs/           # XDG user directory paths
│   ├── video/              # Pure Go youtube-dl/yt-dlp port
│   │   ├── extractor/      # Site-specific metadata extractors (YouTube, generic)
│   │   ├── downloader/     # HTTP/HLS download engines
│   │   ├── format/         # Format sorting & selection
│   │   ├── m3u8/           # HLS manifest parser
│   │   ├── nethttp/        # HTTP client with retry & cookie jar
│   │   ├── jsinterp/       # goja JS runtime (YouTube signature decryption)
│   │   ├── cache/          # Filesystem cache (XDG paths)
│   │   └── utils/          # HTML parse, URL, filename sanitize
│   ├── cobra/              # Cobra CLI helpers
│   │   └── helper/output/  # Output formatting (text/JSON/table)
│   └── private/            # Private internal utilities
├── tests/                  # Integration tests & golden master tests
│   ├── golden/             # Golden master test snapshots
│   │   ├── golden_tests.yaml     # Test definitions
│   │   └── snapshots/      # .stdout files for regression testing
│   ├── scripts/
│   │   ├── test_video.py   # Black-box video tests (vs yt-dlp)
│   │   └── helpers.py      # Test utilities
│   └── __pycache__/        # Python test cache
├── tools/
│   ├── golden/             # Full golden master testing system
│   │   ├── golden_tests.yaml     # Shared test registry
│   │   ├── src/golden/           # Full engine (11 modules)
│   │   └── golden_masters/       # SHA-256 baselines (gitignored)
│   └── scripts/
├── docs/                   # Documentation
│   ├── ROADMAP.md          # Features, phases, milestones
│   ├── BACKLOG.md          # Tech debt, future work
│   ├── ISSUES.md           # Known bugs
│   ├── ARCHITECTURE.md     # Mermaid diagrams
│   ├── GOLDEN_MASTER_TESTING.md
│   ├── adr/                # Architecture Decision Records
│   └── plans/              # Implementation plans
├── docker/                 # Docker build files
├── .github/workflows/      # GitHub Actions CI/CD
├── Taskfile.yml            # Task automation (test, build, lint, etc.)
├── go.mod                  # Module definition & dependencies
├── go.sum                  # Dependency checksums
├── main.go                 # Entry point (delegates to cmd.Execute())
└── CLAUDE.md               # Project conventions & patterns
```

## Directory Purposes

**cmd/:**
- Purpose: Cobra CLI command definitions
- Contains: Command structs, flag definitions, args preprocessing
- Key files: `root.go` (error handling & exit codes), `<cmd>.go` for each command
- Pattern: Thin wrappers that delegate to `internal/cli/<cmd>.Run()`
- Dependency: Imports `internal/cli/<cmd>` for business logic

**internal/cli/:**
- Purpose: Command I/O glue layer
- Contains: One directory per command with Options & Run function
- Pattern: Each Run function accepts `(io.Writer, io.Reader, []string, Options)` and calls `pkg/` logic
- Key shared: `cmderr/` (error classification), `command/` (registry), `input/` (stdin/file abstraction)

**internal/cli/cmderr/:**
- Purpose: Unified error classification
- Contains: Sentinel errors (ErrNotFound, ErrPermission, etc.)
- Maps errors to exit codes: 1=NotFound, 2=InvalidInput, 3=Permission, 4=IO, 5=Timeout, 6=Unsupported
- Pattern: Use `cmderr.Wrap(sentinel, "context")` to add context without losing error type

**internal/cli/command/:**
- Purpose: Registry-based command dispatch
- Contains: Command interface, Registry (thread-safe map), adapters
- Usage: `pipe` command uses Registry to dispatch subcommands
- Pattern: `reg.Register("cmd", AdaptWriterArgs(func(w io.Writer, args []string) error { ... }))`

**internal/cli/input/:**
- Purpose: Unified input handling (stdin, files)
- Contains: Source struct with Name and Reader
- Pattern: `input.Open(args, r)` returns sources, `input.CloseAll(sources)` for cleanup
- Handles: `-` → stdin, multiple files, missing files with appropriate errors

**pkg/:**
- Purpose: Reusable libraries importable by external projects
- Contains: Pure Go implementations without io.Writer output
- Pattern: Core logic only; I/O handled in `internal/cli/` layer
- Examples: `pkg/idgen.GenerateUUID()`, `pkg/hashutil.Hash()`, `pkg/jsonutil.Query()`

**pkg/video/:**
- Purpose: Pure Go youtube-dl/yt-dlp port
- Contains: Extractors (YouTube, generic), downloaders (HTTP, HLS), format selectors
- Key files: `extractor/youtube/` (InnerTube API), `downloader/hls.go` (HLS segment download)
- Database: SQLite `channel.db` for incremental channel downloads

**pkg/twig/:**
- Purpose: Directory tree scanning with comparison
- Contains: `scanner/` (parallel walking, MaxFiles), `formatter/` (ASCII/JSON), `comparer/` (move detection)
- Pattern: `twig.WithMaxFiles(10000)`, `twig.WithParallel(8)`, `twig.WithProgressCallback()`

**pkg/pipeline/:**
- Purpose: Streaming text processing (constant memory)
- Contains: 20+ stages (grep, sort, head, cut, tr, sed, etc.) connected via io.Pipe
- Pattern: `pipeline.New(grep, sort, uniq, head).Run(ctx, stdin, stdout)`

**tests/:**
- Purpose: Integration & golden master regression tests
- Contains: YAML test definitions, snapshot comparisons
- Key files: `golden_tests.yaml`, `snapshots/` (per-command stdout files)
- Pattern: Run `task test:golden:update` to regenerate on intentional output changes

**docs/:**
- Purpose: Project documentation
- Key files: ROADMAP.md (phases), BACKLOG.md (tech debt), ISSUES.md (bugs), ARCHITECTURE.md (diagrams)

## Key File Locations

**Entry Points:**
- `main.go`: Entry point (delegates to cmd.Execute())
- `cmd/root.go`: Root command with logging, error handling, exit code mapping

**Configuration:**
- `Taskfile.yml`: Task automation
- `go.mod`: Module definition & dependencies
- `.github/workflows/test.yml`: CI/CD pipeline

**Core Logic:**
- `pkg/<domain>/*.go`: Business logic (hashutil, jsonutil, encoding, etc.)
- `internal/cli/<cmd>/<cmd>.go`: I/O wrapper for each command

**Error Handling:**
- `internal/cli/cmderr/cmderr.go`: Sentinel errors & exit code mapping

**Command Dispatch:**
- `internal/cli/command/command.go`: Unified Command interface & Registry
- `internal/cli/pipe/pipe.go`: Pipe command with variable substitution
- `internal/cli/pipeline/pipeline.go`: Streaming text processing

**Testing:**
- `tests/golden/golden_tests.yaml`: Test definitions
- `tests/golden/snapshots/<cmd>.stdout`: Expected output snapshots
- `tests/scripts/test_video.py`: Black-box video tests

**Documentation:**
- `docs/ROADMAP.md`: Features & milestones
- `docs/BACKLOG.md`: Tech debt & deprecations
- `docs/ARCHITECTURE.md`: Architecture diagrams

## Naming Conventions

**Files:**
- `<cmd>.go`: Command implementation in `internal/cli/<cmd>/`
- `<cmd>_test.go`: Tests for command
- `<cmd>_unix.go`: Unix-specific code (build tag: `//go:build unix`)
- `<cmd>_windows.go`: Windows-specific code (build tag: `//go:build windows`)

**Directories:**
- `cmd/`: Cobra command wrappers
- `internal/cli/<cmd>/`: Command I/O layer (one dir per command)
- `pkg/<domain>/`: Reusable libraries (no I/O, importable)
- `internal/cli/<shared>/`: Shared utilities (cmderr, command, input, logger)

**Packages:**
- Match directory name: `package head`, `package pipe`, `package cmderr`
- Descriptive names: `idgen`, `hashutil`, `jsonutil`, `sqlfmt`

**Functions:**
- `Run*()`: Command entry point (e.g., `RunHead()`, `RunGrep()`)
- `With*()`: Functional options (e.g., `WithIndent()`, `WithMaxFiles()`)
- `Parse*()`: Parser functions (e.g., `ParseM3U8()`)
- `Match*()`: Predicate functions (e.g., `MatchesGlob()`)

**Interfaces:**
- `Command`: Unified command dispatch interface
- `Stage`: Pipeline stage interface
- `Extractor`: Video extractor interface
- `Option`: Functional option function type

## Where to Add New Code

**New Command (full flow):**
1. Create `internal/cli/<cmd>/<cmd>.go` with Options struct and `Run()` function
2. Create `internal/cli/<cmd>/<cmd>_test.go` with table-driven tests
3. Create `cmd/<cmd>.go` with Cobra command definition
4. Register in `cmd/init()`: `rootCmd.AddCommand(<cmd>Cmd)`
5. Add golden tests to `tests/golden/golden_tests.yaml` and `tools/golden/golden_tests.yaml`
6. Run `task test:golden:update && task golden:record` to generate snapshots
7. Update `CLAUDE.md` command categories

**New Package/Library:**
- Create `pkg/<domain>/` directory
- Export pure functions (no io.Writer output)
- Use functional options: `WithFlag(val) Option`
- Write tests in `*_test.go` files
- Document via godoc comments

**Shared Utilities:**
- Add to `internal/cli/<utility>/` if command-specific
- Add to `pkg/cobra/helper/` if general CLI helper
- Add to `internal/cli/cmderr/` if error-related

**Platform-Specific Code:**
- Create `<file>_unix.go` and `<file>_windows.go`
- Use `//go:build unix` and `//go:build windows` tags
- Keep interface in main file, implementations in variants

**Tests:**
- Unit: `*_test.go` files in same directory as code
- Integration: `tests/` directory
- Golden: Add entry to `tests/golden/golden_tests.yaml`
- Black-box: Add to `tests/scripts/test_*.py` for cross-tool comparison

## Special Directories

**bin/:**
- Purpose: Compiled binaries
- Generated: Yes (created by `task build`)
- Committed: No (in .gitignore)

**pkg/private/:**
- Purpose: Private internal utilities not exported
- Generated: No
- Committed: Yes
- Usage: Vendored code, internal utilities

**.github/workflows/:**
- Purpose: CI/CD pipeline definitions
- Generated: No
- Committed: Yes
- Key: `test.yml` runs golangci-lint, gofmt, govulncheck, go test -race

**testing/__pycache__/, tools/golden/src/golden/__pycache__/:**
- Purpose: Python test cache
- Generated: Yes
- Committed: No (in .gitignore)

**tools/golden/golden_masters/:**
- Purpose: SHA-256 baseline manifests for golden testing
- Generated: Yes
- Committed: No (gitignored)
- Created by: `task golden:record`

---

*Structure analysis: 2026-04-11*
