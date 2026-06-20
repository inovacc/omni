<!-- Extracted from CLAUDE.md leanness pass, 2026-05-24 -->
# Architecture

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
│   └── userdirs/       # XDG user directory paths
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

Use build tags for platform-specific implementations. Two acceptable layouts:

**Single file with build tag** — use when only one platform needs a divergent impl and the file is small (<~80 lines):

```go
//go:build !windows
// or
//go:build unix
```

**Split files (default for new scaffolds)** — `_windows.go` + `_darwin.go` + `_unix.go` (where `_unix.go` covers Linux/BSD via `//go:build unix && !darwin`). Use when each platform diverges meaningfully:

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

##### Scaffolding daemon-style apps (`--daemon`)

For long-running services that need self-management (PID file, foreground/background, OS service install) — modeled after the weaver pattern — pass `--daemon` to `omni scaffold cobra init`:

```bash
omni scaffold cobra init weaverd --module github.com/user/weaverd --daemon
# Generates:
#   internal/serverinfo/serverinfo.go    - PID + version JSON state under os.UserConfigDir(),
#                                          stale-PID detection via gopsutil
#   cmd/weaverd/cmd_server.go            - Cobra subcommands: server {start,stop,restart,status,install,uninstall}
#                                          plus --foreground flag on start
#   cmd/weaverd/server.go                - Shared: serverStart/Stop/Restart/Status/Install/Uninstall + daemonize()
#                                          self-execs with WEAVERD_DAEMON_CHILD=1 to detach from parent
#   cmd/weaverd/server_unix.go           - //go:build !windows: setSysProcAttr (Setsid),
#                                          stopProcess (SIGTERM), isPrivileged (uid 0), elevateAndRerun (sudo)
#   cmd/weaverd/server_systemd.go        - //go:build !windows && !darwin: systemd unit install/uninstall
#   cmd/weaverd/server_darwin.go         - //go:build darwin: launchd plist install/uninstall
#   cmd/weaverd/server_windows.go        - //go:build windows: SCM install/uninstall, taskkill, UAC elevate
```

**What's different from `--service`:**
- `--service` registers with the OS service manager via `kardianos/service` — start/stop happen through systemd/launchd/SCM.
- `--daemon` is **self-supervising**: writes its own PID file, validates the recorded PID is actually our binary (handles crashes that leave stale state), and re-execs itself with an env-var marker to fork into the background.
- The two are **mutually exclusive** (both would register a lifecycle command group); the scaffolder rejects `--service --daemon`.

**How to fill in your server logic**: edit `runServe()` in `cmd/<app>/server.go`. It MUST call `serverinfo.Write()` once ready to serve and `serverinfo.Remove()` on exit (the scaffolded stub does both).

**Privilege elevation**: `server install`/`uninstall` auto-elevate via `sudo` (Unix) or `runas` (Windows) if not already privileged — no need to wrap the command yourself.

**Foreground mode**: `<app> server start --foreground` skips daemonization (useful when running under a service manager that expects PID 1 to be the daemon, like `systemctl`'s `Type=simple` units). The generated systemd unit and launchd plist already pass `--foreground`.

##### Scaffolding platform-split commands

`omni scaffold cobra add` accepts `--platform-split` to emit the three-file layout automatically alongside the shared Cobra registration file:

```bash
omni scaffold cobra add daemon --platform-split
# Creates:
#   cmd/<app>/cmd_daemon.go          - shared: Cobra registration, RunE -> runDaemon(cmd, args)
#   cmd/<app>/cmd_daemon_windows.go  - //go:build windows         + func runDaemon
#   cmd/<app>/cmd_daemon_darwin.go   - //go:build darwin          + func runDaemon
#   cmd/<app>/cmd_daemon_unix.go     - //go:build unix && !darwin + func runDaemon
```

The shared file's `RunE` delegates to `run<Name>(cmd, args)`; each platform file supplies exactly one implementation, so the Go build picks the correct one per OS at compile time. Pre-flight check rejects the scaffold if any of the four target files already exists.

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

**Commands adopted (ALL):** every command in `internal/cli/` returns classified cmderr sentinels. 100% adoption completed in Phase 1 (April 2026).

**Exit-code contract:** v1.0 is the first stable exit-code contract. Changes from this point forward follow the CLAUDE.md breaking-change protocol.

See `.planning/phases/01-cmderr-migration-completion/EXIT-CODE-CHANGES.md` for the Phase 1 transition ledger.

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

#### No-exec invariant: scope & sanctioned exceptions

omni's foundational **NO-EXEC** rule (CLAUDE.md Design Principle #1) governs **utility reimplementations**. Any command that *re-implements* a shell utility — `cat`, `grep`, `tar`, `unzip`, `sed`, `find`, `ls`, `df`, `ps`, hashing, encoding, etc. — MUST be pure Go and MUST NEVER shell out. Spawning an external process to do work that omni claims to provide natively is a violation: it breaks the "single static binary, no external processes" core value, ties the command to an undeclared runtime dependency on `$PATH`, and forfeits determinism/portability. New utility commands have zero tolerance for `os/exec`.

A small, explicitly enumerated set of commands exist **specifically to orchestrate external tools**. For these, running an external process IS the command's purpose — there is no pure-Go equivalent to reimplement, because the thing being driven is itself the external tool. These are **sanctioned exceptions** that MAY import and use `os/exec`:

| Command | Orchestrates | Notes |
|---------|--------------|-------|
| `exec` | an arbitrary operator-supplied command | the launcher *is* the feature; stdio inherited from the operator |
| `forloop` (`omni for`) | a per-iteration command template | must use argv-array invocation, never a shell string |
| `task` | task-runner command lines | prefer the in-process Command registry; never a shell fallback |
| `terraform` (`omni tf`) | the `terraform` binary | external prerequisite documented |
| git hacks (`omni git ...` / `omni gh ...`) | `git` / `gh` binaries | args passed as argv, IDs parsed as ints |
| `repo` | `git` / `gh` for remote clone | argv invocation only |
| `buf generate` (local plugins) | `protoc` / local codegen plugins | args from operator-authored `buf.gen.yaml` |

**These are the ONLY allowed exec sites.** Rules for sanctioned exceptions:

- **Injection-safe always.** Pass arguments as an argv slice (`exec.Command(bin, args...)`), NEVER interpolate untrusted input into a shell string (`sh -c` / `cmd /C`). A loop value, filename, manifest field, or config-derived token concatenated into a shell command line is a command-injection sink and is forbidden.
- **No PATH-hijack surprises.** Resolving an unqualified binary name via `exec.LookPath` inherits `$PATH`; treat a missing tool as a clear, classified error (`cmderr.ErrUnsupported`), not an opaque exec failure.
- **No new exec sites.** Adding `os/exec` to any package not in the table above is a NO-EXEC violation and must be rejected in review. If a genuinely new orchestrator is needed, it must be added to this table with explicit justification, kept argv-only, and documented in `docs/EXTERNAL_SOURCES.md`.
- **Platform helpers are not an excuse.** Build-tagged platform code (`_darwin.go`, `_windows.go`) is held to the same standard — deriving a value the binary needs (machine ID, kernel version) must use pure-Go (`golang.org/x/sys`) sources, not a spawned OS utility.

See `docs/quality/HARDENING.md` for the per-finding resolution status of historical exec sites.

---

