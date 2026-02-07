# Code Reusability Analysis

This document analyzes the omni codebase to identify reusable patterns, code consolidation opportunities, and potential refactoring targets.

---

## Current Architecture Overview

### Project Structure
```
omni/
├── cmd/                    # 98 Cobra CLI commands
├── pkg/                    # 12 reusable Go libraries (importable externally)
│   ├── idgen/              # UUID, ULID, KSUID, Nanoid, Snowflake
│   ├── hashutil/           # MD5, SHA256, SHA512 hashing
│   ├── jsonutil/           # jq-style JSON query engine
│   ├── encoding/           # Base64, Base32, Base58
│   ├── cryptutil/          # AES-256-GCM encrypt/decrypt
│   ├── sqlfmt/             # SQL format/minify/validate
│   ├── cssfmt/             # CSS format/minify/validate
│   ├── htmlfmt/            # HTML format/minify/validate
│   ├── textutil/           # Sort, Uniq, Trim + diff/
│   ├── search/grep/        # Pattern search with options
│   ├── search/rg/          # Gitignore parsing, file type matching
│   └── twig/               # Tree scanning, formatting, comparison
├── internal/
│   ├── cli/               # CLI wrappers (delegates to pkg/ for core logic)
│   ├── flags/             # Feature flags system
│   └── logger/            # KSUID-based logging
└── main.go
```

### Statistics
- **Command Files:** 98 in `cmd/`
- **Pkg Libraries:** 12 in `pkg/` (externally importable)
- **CLI Packages:** 79 in `internal/cli/` (thin wrappers)
- **Test Coverage:** 97.7% (86/88 packages)
- **Total Test Cases:** 700+

---

## Existing Shared Packages

### 1. `internal/cli/fs` - File System Operations
**Current Functions:**
- `Cd(path)` - Change directory
- `Chmod(path, mode)` - Change permissions
- `Mkdir(path, perm, parents)` - Create directory
- `Rmdir(path)` - Remove directory
- `Touch(path)` - Create/update timestamps
- `Rm(path, recursive)` - Remove files
- `Copy(src, dst)` - Copy file
- `Move(src, dst)` - Move/rename file
- `Stat(path)` - File info
- `Lstat(path)` - File info (no symlink follow)
- `IsNotExist(err)` - Check error type

**Usage:** Used by `cp`, `mv`, `rm`, `mkdir`, `rmdir`, `touch`, `chmod`, `stat`

### 2. `internal/cli/path` - Path Utilities
**Current Functions:**
- `Realpath(path)` - Resolve absolute path with symlinks
- `Dirname(path)` - Extract directory component
- `Basename(path)` - Extract file name component
- `Join(paths...)` - Join path components

**Usage:** Used by `realpath`, `dirname`, `basename`

### 3. `internal/cli/text` - Text Processing
**Current Functions:**
- `RunSort(w, args, opts)` - Sort command
- `RunUniq(w, args, opts)` - Uniq command
- `Sort(lines)` - Simple sort helper
- `Uniq(lines)` - Simple unique helper
- `TrimLines(lines)` - Trim whitespace helper

**Structs:**
- `SortOptions` - 11 configuration fields
- `UniqOptions` - 9 configuration fields

**Usage:** Used by `sort`, `uniq`

### 4. `internal/cli/pager` - TUI Paging
**Functions:** Terminal-based file viewing with scrolling

**Usage:** Used by `less`, `more`

### 5. `internal/cli/timeutil` - Time Utilities
**Functions:** Time formatting and parsing helpers

**Usage:** Used by `date`, `time`, `uptime`

### 6. `internal/cli/crypt` - Encryption
**Functions:** AES-256-GCM encryption/decryption

**Usage:** Used by `encrypt`, `decrypt`

---

## Code Patterns Analysis

### Consistent Command Pattern
All commands follow the signature:
```go
func Run<Command>(w io.Writer, args []string, opts *<Command>Options) error
```

**Examples Found (50+):**
- `RunCat(w, args, opts)` - cat package
- `RunGrep(w, pattern, args, opts)` - grep package
- `RunSort(w, args, opts)` - text package
- `RunHash(w, args, opts)` - hash package

### Duplicated Patterns

#### File Reading Pattern (35 occurrences)
```go
scanner := bufio.NewScanner(r)
for scanner.Scan() {
    line := scanner.Text()
    // process line
}
return scanner.Err()
```

**Found in:** cat, grep, head, tail, wc, nl, sort, uniq, cut, paste, join, sed, awk, fold, column, tac, rev, shuf, split, comm, diff

#### File Opening Pattern (55 occurrences)
```go
f, err := os.Open(path)
if err != nil {
    return fmt.Errorf("command: %s: %w", path, err)
}
defer func() { _ = f.Close() }()
```

**Found in:** Nearly all file-processing commands

#### Stdin Fallback Pattern
```go
var r io.Reader = os.Stdin
if len(args) > 0 && args[0] != "-" {
    f, err := os.Open(args[0])
    // ...
    r = f
}
```

**Found in:** cat, grep, head, tail, wc, sort, uniq, cut, nl, sed, awk

---

## Consolidation Opportunities

### Priority 1: Input Reader Helper

**Proposed Package:** `internal/cli/input`

```go
package input

// Reader wraps stdin or file input with consistent error handling
type Reader struct {
    r         io.Reader
    closer    io.Closer
    filename  string
}

// OpenArgs opens files from args or falls back to stdin
func OpenArgs(args []string) ([]*Reader, error) {
    if len(args) == 0 {
        return []*Reader{{r: os.Stdin, filename: "(stdin)"}}, nil
    }
    // ...
}

// Lines returns a line scanner for the reader
func (r *Reader) Lines() *LineScanner

// Close closes the underlying file if applicable
func (r *Reader) Close() error
```

**Impact:** Could simplify 27+ packages

### Priority 2: Line Processing Pipeline

**Proposed Package:** `internal/cli/pipeline`

```go
package pipeline

// Stage represents a processing step
type Stage func([]string) ([]string, error)

// Run executes stages in sequence
func Run(input []string, stages ...Stage) ([]string, error)

// Common stages
func Filter(fn func(string) bool) Stage
func Map(fn func(string) string) Stage
func Sort(opts SortOptions) Stage
func Unique(opts UniqOptions) Stage
```

**Impact:** Could unify text processing in sort, uniq, grep, sed, awk, tr, cut

### Priority 3: Output Formatter

**Proposed Package:** `internal/cli/output`

```go
package output

type Format int

const (
    FormatText Format = iota
    FormatJSON
    FormatTable
)

// Formatter handles output in various formats
type Formatter struct {
    w      io.Writer
    format Format
}

// Print outputs data in the configured format
func (f *Formatter) Print(data any) error

// PrintLines outputs lines with optional line numbers
func (f *Formatter) PrintLines(lines []string, numbered bool) error
```

**Impact:** Could standardize JSON output across all commands

### Priority 4: Options Pattern Consolidation

Many Options structs share common fields:

**Common Fields:**
- `Quiet bool` - Suppress output
- `Verbose bool` - Extra output
- `DryRun bool` - Preview only
- `Recursive bool` - Process directories
- `Force bool` - Override protections

**Proposed:**
```go
package options

// Common embeddable options
type Common struct {
    Quiet   bool
    Verbose bool
}

type FileOps struct {
    Common
    Recursive bool
    Force     bool
    DryRun    bool
}
```

---

## Refactoring Candidates

### High Priority (Reduce Duplication)

| Package | Current Lines | After Refactor | Savings |
|---------|---------------|----------------|---------|
| `cat` | 160 | 80 | 50% |
| `head` | 150 | 70 | 53% |
| `tail` | 200 | 90 | 55% |
| `wc` | 180 | 80 | 56% |
| `nl` | 120 | 60 | 50% |
| `grep` | 310 | 180 | 42% |

**Total Estimated Reduction:** ~400 lines

### Medium Priority (Improve Maintainability)

| Package | Issue | Recommendation |
|---------|-------|----------------|
| `text` | Has both Sort and Uniq in one file | Keep as is (cohesive) |
| `hash` | Multiple hash algorithms | Extract to `internal/cli/hash/algo` |
| `archive` | tar/zip/gzip in one file (500+ lines) | Split into `tar.go`, `zip.go`, `gzip.go` |
| `base` | base32/58/64 in one file | Keep as is (similar logic) |

### Low Priority (Future Improvements)

| Package | Issue | Recommendation |
|---------|-------|----------------|
| Platform files | `*_unix.go`, `*_windows.go` scattered | Create `internal/platform` abstraction |
| Error messages | Inconsistent formatting | Create `internal/errors` package |
| Tests | Some duplicated test helpers | Create `internal/testutil` package |

---

## Merge Candidates (Similar Functionality)

### 1. Encoding Commands
**Current:** `base32`, `base58`, `base64` (3 separate cmd files)
**Internal:** Single `internal/cli/base` package (correct)
**Recommendation:** Keep merged in internal, separate cmd files for discoverability

### 2. Hash Commands
**Current:** `hash`, `md5sum`, `sha256sum`, `sha512sum` (4 cmd files)
**Internal:** Single `internal/cli/hash` package (correct)
**Recommendation:** Keep as is

### 3. Grep Variants
**Current:** `grep`, `egrep`, `fgrep` (3 cmd files)
**Internal:** Single `internal/cli/grep` package with options
**Recommendation:** Keep as is (Unix compatibility)

### 4. Archive Commands
**Current:** `tar`, `zip`, `unzip`, `gzip`, `bzip2`, `xz` (6 cmd files)
**Internal:** `archive`, `gzip`, `bzip2`, `xz` (4 packages)
**Recommendation:** Consider consolidating `gzip`, `bzip2`, `xz` into `compression` package

---

## Implementation Roadmap

### Phase 1: Create Input Helper (2-3 hours)
1. Create `internal/cli/input/input.go`
2. Implement `Reader`, `OpenArgs`, `Lines`
3. Update 5 packages as proof-of-concept: `cat`, `head`, `tail`, `wc`, `nl`
4. Add tests

### Phase 2: Create Pipeline Package (4-6 hours)
1. Create `internal/cli/pipeline/pipeline.go`
2. Implement `Stage`, `Run`, common stages
3. Refactor `sort`, `uniq` to use pipeline
4. Add tests

### Phase 3: Refactor Remaining Packages (8-10 hours)
1. Update remaining text processing packages
2. Update grep, sed, awk to use shared input handling
3. Ensure backwards compatibility

### Phase 4: Add Output Formatter (3-4 hours)
1. Create `internal/cli/output/output.go`
2. Implement JSON/Text/Table formats
3. Update commands with `--json` flag

---

## Metrics Summary

| Metric | Current | After Consolidation |
|--------|---------|---------------------|
| Total Packages | 79 | 82 (+3 shared) |
| Duplicated Patterns | ~400 lines | ~50 lines |
| Shared Utilities | 5 packages | 8 packages |
| Avg Package Size | ~150 lines | ~120 lines |

---

## Recommendations

1. **Immediate:** Create `internal/cli/input` package to reduce file handling duplication
2. **Short-term:** Create `internal/cli/output` for consistent JSON output
3. **Medium-term:** Create `internal/cli/pipeline` for text processing chain
4. **Long-term:** Create `internal/platform` for cross-platform abstractions

---

## Related Documents

- [ROADMAP.md](ROADMAP.md) - Implementation phases
- [BACKLOG.md](BACKLOG.md) - Future work items
- [COMMANDS.md](COMMANDS.md) - Command reference
