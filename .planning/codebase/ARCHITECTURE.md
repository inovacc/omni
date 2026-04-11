# Architecture

**Analysis Date:** 2026-04-11

## Pattern Overview

**Overall:** Hexagonal/Clean Architecture with Cobra CLI framework

**Key Characteristics:**
- Thin CLI command wrappers in `cmd/` that delegate to reusable packages
- Core logic abstracted into `pkg/` libraries importable by external projects
- I/O glue layer in `internal/cli/` that bridges Cobra and pkg logic
- Unified Command interface for registry-based dispatch
- Error classification via sentinel values with mapped exit codes
- All commands accept `io.Writer` and `io.Reader` for testability

## Layers

**Presentation Layer (cmd/):**
- Purpose: Cobra CLI command definitions and flag parsing
- Location: `cmd/*.go` (160+ command files)
- Contains: Command definitions, flag configuration, args preprocessing
- Depends on: `internal/cli/<command>` packages for business logic
- Used by: Cobra CLI framework (main entry point)
- Pattern: Commands are thin wrappers that parse flags and delegate to `internal/cli/<cmd>/Run()` functions

**I/O Glue Layer (internal/cli/<command>/):**
- Purpose: Handle input/output operations, orchestrate stdin/stdout, manage file I/O
- Location: `internal/cli/<command>/` (160+ subdirectories, one per command)
- Contains: Options structs, Run functions with `(io.Writer, io.Reader, []string)` signatures
- Depends on: `pkg/` libraries for core logic, `internal/cli/cmderr` for error handling
- Used by: `cmd/` wrappers and pipe command
- Pattern: Each command has a Run function that coordinates I/O, calls pkg logic, and formats output

**Business Logic Layer (pkg/):**
- Purpose: Reusable, testable command implementations
- Location: `pkg/<domain>/` (18+ subdirectories)
- Contains: Core algorithms (hashing, encryption, encoding, formatting), data structures, processing engines
- Depends on: Go stdlib only (no Cobra, no io.Writer output)
- Used by: `internal/cli/` packages and external Go projects
- Pattern: Pure functions with options using functional option pattern (`WithFlag(value) Option`)

**Infrastructure Layer:**
- Purpose: Cross-cutting concerns and utilities
- Location: `internal/cli/cmderr/`, `internal/cli/command/`, `internal/cli/input/`, `pkg/cobra/`
- Contains: Error handling, command registry, input source management, output formatting

## Data Flow

**Command Execution Path:**

1. User runs `omni <cmd> <args>`
2. `main.go` → `cmd.Execute()` (root command)
3. Cobra parses flags and routes to specific `cmd/<cmd>.go`
4. `cmd/<cmd>.go` constructs Options struct from flags
5. `cmd/<cmd>.go` calls `internal/cli/<cmd>.Run(w, r, args, opts)`
6. `internal/cli/<cmd>` handles I/O:
   - Opens input sources via `input.Open(args, r)`
   - Calls pkg logic (e.g., `pkg/hash.Hash(data)`)
   - Formats output via `output.New(w, format).Render()`
7. Returns error (classified via cmderr sentinels)
8. `cmd/root.go` maps error to exit code via `cmderr.ExitCodeFor(err)`

**Pipe Command Special Flow:**

1. User runs `omni pipe 'cmd1' 'cmd2' 'cmd3'`
2. `internal/cli/pipe.Run()` parses command strings
3. Applies variable substitution (`$OUT` → last output line)
4. For each substituted command:
   - Routes to `executeCommand()` which tries:
     a. Unified `command.Registry` (if available)
     b. Falls back to Cobra via `cobradispatch.ExecuteCommand()`
   - Chains stdin/stdout between commands
5. Returns aggregated Result with all intermediate outputs

**Error Handling:**

- Commands return errors instead of calling `os.Exit()`
- Errors are classified via sentinel values in `internal/cli/cmderr/`
- `cmderr.ExitCodeFor(err)` in root command maps errors to exit codes:
  - `ErrNotFound` → 1
  - `ErrInvalidInput` → 2
  - `ErrPermission` → 3
  - `ErrIO` → 4
  - `ErrTimeout` → 5
  - `ErrUnsupported` → 6
  - `SilentExit(code)` → code without printing error message

## Key Abstractions

**Command Interface:**
- Purpose: Unified interface for command dispatch in pipe/pipeline
- Location: `internal/cli/command/command.go`
- Pattern: `Command interface { Run(ctx, w, io.Writer, r io.Reader, args []string) error }`
- Adapters: `AdaptWriterArgs()`, `AdaptWriterReaderArgs()`, `AdaptFull()` for legacy Run signatures
- Registry: Thread-safe map of command name → Command implementation

**Options Pattern:**
- Purpose: Flexible, type-safe command configuration
- Example: `head.HeadOptions{Lines: 10, Quiet: false}`
- pkg pattern: `func (opt Option) Apply(*Options)` for functional options
- Cobra pattern: Parse flags → populate Options struct → pass to Run function

**Input Source Abstraction:**
- Purpose: Unified handling of stdin, files, and multiple input sources
- Location: `internal/cli/input/input.go`
- Pattern: `input.Open(args, r)` returns `[]Source` where Source has Name and Reader
- Handles: `-` (stdin), multiple files, missing files with appropriate error classification

**Output Formatting:**
- Purpose: Flexible output (text/JSON/table) without duplicating logic
- Location: `pkg/cobra/helper/output/`
- Pattern: `output.New(w, format)` → methods like `Render()`, `IsJSON()`
- Usage: Commands check `f.IsJSON()` to decide between text and JSON output

**Error Wrapping:**
- Purpose: Add context while preserving sentinel error type
- Location: `internal/cli/cmderr/cmderr.go`
- Pattern: `cmderr.Wrap(sentinel, "context message")` returns `fmt.Errorf("context: %w", sentinel)`
- Effect: `errors.Is(err, ErrNotFound)` still matches wrapped errors

## Entry Points

**Main Entry Point:**
- Location: `main.go`
- Triggers: `omni` binary execution
- Responsibilities: Delegates to `cmd.Execute()`

**Root Command:**
- Location: `cmd/root.go`
- Triggers: All subcommands route through here
- Responsibilities:
  - Sets up logger and command execution context
  - Parses persistent flags (`--json`, `--table`)
  - Maps errors to exit codes
  - Silences Cobra usage/error output

**Command Preprocessors:**
- Location: `cmd/<cmd>.go init()` functions
- Purpose: Rewrite `os.Args` before Cobra parsing for GNU compatibility
- Examples:
  - `head`/`tail`: Convert `-20` to `-n 20`
  - `find`: Convert `-name` to `--name`

## Cross-Cutting Concerns

**Logging:**
- Framework: `log/slog` (structured JSON)
- Location: `internal/logger/`
- Pattern: Initialized in `cmd/root.go` PersistentPreRun
- Output: Only when `OMNI_LOGGING` env var set

**Validation:**
- Pattern: Input validation in `internal/cli/<cmd>/` before calling pkg logic
- Example: `head` validates Lines/Bytes are non-negative
- Error type: `cmderr.ErrInvalidInput` for bad arguments

**Authentication:**
- Cloud commands (AWS, kubectl, vault): Use environment-based auth
- Location: `internal/cli/aws/`, `internal/cli/kubectl/`, `internal/cli/vault/`
- Pattern: Load config from `~/.aws/credentials`, `~/.kube/config`, `~/.vault-token`

**File Operations:**
- Pattern: Use `pkg/twig/scanner` for directory walking with parallelism and progress
- Pattern: Use `internal/cli/input` for unified input source handling
- Pattern: Always defer file closes with `defer func() { _ = f.Close() }()`

---

*Architecture analysis: 2026-04-11*
