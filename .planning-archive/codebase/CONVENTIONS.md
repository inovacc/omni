# Coding Conventions

**Analysis Date:** 2026-04-11

## Naming Patterns

**Files:**
- Command packages: `internal/cli/<command>/` (lowercase, single word or hyphenated)
- Library packages: `pkg/<domain>/` (e.g., `pkg/video/`, `pkg/twig/`)
- Test files: `<name>_test.go` (always suffix test, never prefix)
- Platform-specific: `<name>_unix.go`, `<name>_windows.go` with `//go:build` tags
- Options structs: `<Command>Options` (PascalCase, e.g., `HeadOptions`, `HashOptions`)
- Result structs: `<Command>Result` or `<Commands>Result` for plurals (e.g., `HashResult`, `HashesResult`)
- Exported functions: `Run<Command>` or `Run` (e.g., `RunHead()`, `RunHash()`)
- Unexported helpers: lowercase with underscores (e.g., `verifyChecksums()`, `parseCommands()`)

**Functions:**
- All exported functions: PascalCase starting with action verb (e.g., `RunHash`, `WithIndent`, `FormatBytes`)
- CLI wrappers in `cmd/`: Cobra command variables named `<command>Cmd` (e.g., `headCmd`, `lsCmd`)
- Package init functions: Standard `init()` (for flag preprocessing, command registration)
- Receiver names: 1-2 letters (e.g., `func (f *Formatter) Format()`)

**Variables:**
- Options: CamelCase (e.g., `tmpDir`, `testFile`, `showHeaders`)
- Boolean flags: prefix with action or state (e.g., `wantError`, `isJSON`, `showHeaders`)
- Package-level: all lowercase (e.g., `defaultHeadLines`, `omniBanner`)
- Constants: ALL_CAPS (e.g., `DefaultHeadLines`, `ErrNotFound`)

**Types:**
- Structs: PascalCase (e.g., `HashOptions`, `ColorScheme`, `StreamMessage`)
- Interfaces: PascalCase (e.g., `Command`, `Extractor`, `Formatter`)
- Sentinel errors: PascalCase prefixed with `Err` (e.g., `ErrNotFound`, `ErrInvalidInput`)

## Code Style

**Formatting:**
- Tool: `gofmt` (Go standard)
- Linter: `golangci-lint` with `govet` enabled
- Configuration: `.golangci.yml` with minimal rules (excludes generated code, relaxed comment checks)
- No custom formatter config detected

**Linting:**
- Enabled: `govet` only (strict mode)
- Excluded: generated code in `pkg/buf`
- Run: `golangci-lint run --fix ./... --timeout=5m`

**Line length:** No hard limit enforced; most lines stay under 100 characters

**Indentation:** Tabs (Go standard, enforced by gofmt)

## Import Organization

**Order:**
1. Standard library imports (`fmt`, `io`, `os`, etc.)
2. Blank line
3. External packages (`github.com/...`)
4. Blank line
5. Local packages (`github.com/inovacc/omni/...`)

**Pattern:**
```go
import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/cobra/helper/output"
)
```

**Path Aliases:** None detected (use full import paths)

**Package Docstrings:** Always present at package level. Multi-line for complex packages:
```go
// Package cmderr provides a unified error model for CLI commands.
// Commands return errors instead of calling os.Exit directly.
// The root command maps errors to exit codes.
package cmderr
```

## Error Handling

**Sentinel Errors (cmderr):**
Use `internal/cli/cmderr` package for all error classification:

| Sentinel | Use For |
|----------|---------|
| `cmderr.ErrNotFound` | File/resource not found (exit code 1) |
| `cmderr.ErrInvalidInput` | Bad flags, missing operands, parse errors (exit code 2) |
| `cmderr.ErrPermission` | Permission denied (exit code 3) |
| `cmderr.ErrIO` | I/O errors (exit code 4) |
| `cmderr.ErrTimeout` | Timeouts (exit code 5) |
| `cmderr.ErrUnsupported` | Unsupported operations (exit code 6) |
| `cmderr.ErrConflict` | Verification failures, sort disorder (exit code 1) |

**Pattern - Classify errors from os/io:**
```go
// in internal/cli/<command>/<command>.go
if errors.Is(err, os.ErrNotExist) {
	return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("head: %s", err))
}
if errors.Is(err, os.ErrPermission) {
	return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("head: %s", err))
}
return fmt.Errorf("head: %w", err)  // fallback for unclassified errors
```

**Pattern - Validation errors:**
```go
return cmderr.Wrap(cmderr.ErrInvalidInput, "path clean: missing operand")
```

**Pattern - Silent exit (grep-style "no match"):**
```go
return cmderr.SilentExit(1)  // no stderr message, just exit code
```

**Error comparison:** Always use `errors.Is()` and `errors.As()`, never `==`:
```go
// ✅ Correct
if errors.Is(err, sql.ErrNoRows) { ... }

// ❌ Wrong — breaks with wrapped errors
if err == sql.ErrNoRows { ... }
```

**Error wrapping:** Always include context:
```go
return fmt.Errorf("head: %w", err)  // adds command name prefix
```

**Printing errors:** Write to stderr, don't panic:
```go
_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)  // stderr for errors
_, _ = fmt.Fprintln(w, output)                     // stdout for results
```

## Logging

**Framework:** `internal/logger` package (structured logging to file)
- Used only for debug/execution tracking, not in command logic
- Commands accept `io.Writer` for stdout (testable)
- Errors written to stderr explicitly

**Pattern:**
```go
log := logger.Init(cmd.Name())
if log.IsActive() {
	stdout, stderr := log.StartExecution(cmd.Name(), args, cmd.OutOrStdout(), cmd.ErrOrStderr())
	cmd.SetOut(stdout)  // wrap stdout to capture
}
```

## Comments

**When to Comment:**
- Package-level docstrings (required for all packages)
- Exported function/type documentation (required by golangci-lint)
- Non-obvious logic or workarounds
- Complex algorithms or edge cases

**Pattern:**
```go
// RunHead executes the head command.
// r is the default input reader (used when args is empty or contains "-").
func RunHead(w io.Writer, r io.Reader, args []string, opts HeadOptions) error {
```

**JSDoc/TSDoc:** Not used (Go uses comment-to-code format above declarations)

**Linting:** Comments with leading uppercase enforced by govet

## Function Design

**Size:** Keep under 200 lines per function; extract helpers at 100+ lines

**Parameters:**
- Max 3-4 parameters (use Options struct for many config values)
- Order: context/writer first, then inputs, then options
- Example: `func RunHead(w io.Writer, r io.Reader, args []string, opts HeadOptions) error`

**Return Values:**
- Always return `error` as last value
- For result data, use struct (e.g., `HashResult`) with JSON tags
- Multiple returns only for (value, error) pairs

**Pattern - Options struct:**
```go
type HeadOptions struct {
	Lines        int           // -n: number of lines
	Bytes        int           // -c: number of bytes
	Quiet        bool          // -q: never print headers
	OutputFormat output.Format // output format (text/json/table)
}
```

**Pattern - Functional options (for pkg/ libraries):**
```go
type Option func(*Options)

func WithIndent(s string) Option {
	return func(o *Options) { o.Indent = s }
}

func Format(input string, opts ...Option) string {
	o := &Options{Indent: "  "}  // defaults
	for _, opt := range opts {
		opt(o)  // apply options
	}
	// ...
}
```

## Module Design

**Exports:** Minimize public surface. Export only:
- Options structs
- Result structs
- Run functions
- High-level public APIs

**Unexported helpers:** All implementation details lowercase

**Barrel Files:** Minimal use; only when grouping related types
- Example: `pkg/video/types/types.go` exports `VideoInfo`, `Format`, etc.

**Package Layout (command):**
```
internal/cli/<command>/
├── <command>.go         # Options struct, Run function
├── <command>_test.go    # Unit tests
├── <command>_unix.go    # Platform-specific (optional)
└── <command>_windows.go # Platform-specific (optional)
```

**Package Layout (library):**
```
pkg/<domain>/
├── <domain>.go          # Main API
├── <domain>_test.go     # Tests
├── options.go           # Functional options
├── types.go             # Exported types
├── helpers.go           # Unexported helpers
└── subpackage/          # Related functionality
```

## Output Patterns

**Mute unused returns:**
```go
_, _ = fmt.Fprintln(w, output)  // intentionally ignore written bytes and error
```

**Use defers for cleanup:**
```go
defer func() {
	_ = file.Close()  // mute close error
}()

// OR handle close errors:
defer func() {
	if err := file.Close(); err != nil {
		log.Printf("close failed: %v", err)
	}
}()
```

**Use io.Writer for testability:**
```go
// All commands accept io.Writer for output
func RunHead(w io.Writer, r io.Reader, args []string, opts HeadOptions) error {
	_, _ = fmt.Fprintln(w, line)  // write to provided writer, not stdout
}

// Tests can capture output in bytes.Buffer
var buf bytes.Buffer
err := RunHead(&buf, nil, args, opts)
output := buf.String()
```

**JSON output:**
```go
f := output.New(w, opts.OutputFormat)  // output helper that handles JSON/table/text
if f.IsJSON() {
	return json.NewEncoder(w).Encode(result)
}
// plain text output
```

## Platform-Specific Code

**Build tags:**
```go
//go:build unix
// +build unix

package df

// Unix-specific implementation
```

**File naming:** `<name>_unix.go`, `<name>_windows.go`, `<name>_darwin.go`

**Shared interface:** Define interface in main file, implement in platform-specific files

**Example structure:**
```
internal/cli/df/
├── df.go           // Interface, default implementation
├── df_unix.go      // //go:build unix
└── df_windows.go   // //go:build windows
```

---

*Convention analysis: 2026-04-11*
