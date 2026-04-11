# Exit-code changes introduced by Phase 1

Each wave appends rows documenting commands whose exit code changed from
un-classified (exit 1) to a classified cmderr sentinel. Used as input to
v1.0 release notes per CONTEXT.md Decision 6.

| Wave | Command | Before | After (sentinel → code) | Notes |
|------|---------|--------|-------------------------|-------|
| A | kill (no args) | 1 | ErrInvalidInput → 2 | Usage error now classified |
| A | kill (bad pid, e.g. `kill not-a-pid`) | 1 | ErrInvalidInput → 2 | Non-numeric pid |
| A | kill (unknown signal via `-s`) | 1 | ErrInvalidInput → 2 | e.g. `kill -s BOGUSX 1` |
| A | kill (nonexistent pid, e.g. `kill 999999999`) | 1 | ErrNotFound → 1 | Same exit; sentinel now set for downstream callers |
| A | kill (permission denied on `process.Signal`) | 1 | ErrPermission → 3 | Was unclassified; EPERM/os.ErrPermission now mapped |
| A | kill (windows, unsupported POSIX signal e.g. `-USR1`) | 1 | ErrUnsupported → 6 | Locked message: `kill: signal SIG%s not supported on windows (INT/KILL/TERM only)`; pins POLISH-09 format |
| A | sort (file not found) | 1 | ErrNotFound → 1 | No-op: already wrapped in `internal/cli/text/text.go` (RunSort). Ledger entry for `internal/cli/sort/` referred to a library-only helper with no CLI wiring. |
| A | sort (permission denied) | 1 | ErrPermission → 3 | Already wrapped in `text.RunSort`; exit code shifts from 1 to 3. |
| A | sort (`-c` disorder) | 1 | ErrConflict → 1 | Already wrapped in `text.RunSort`; exit code unchanged. |
| A | env (write failure, e.g. broken pipe) | 0 (silently muted) | ErrIO → 4 | Previously `_, _ = fmt.Fprint(...)` silently dropped write errors; now classified. |
| A | date (write failure, e.g. broken pipe) | 0 (silently muted) | ErrIO → 4 | Previously `_, _ = fmt.Fprintln(...)` silently dropped write errors; now classified. |
| B | cssfmt / `css validate` (unbalanced/unclosed) | 1 | ErrInvalidInput → 2 | Was `fmt.Errorf("validation failed")`; now cmderr.ErrInvalidInput. |
| B | cssfmt / `css format` (file not found) | 1 | ErrNotFound → 1 | Previously raw `fmt.Errorf`; now classified via errors.Is(os.ErrNotExist). |
| B | cssfmt / `css format` (permission denied) | 1 | ErrPermission → 3 | Exit code shifts from 1 to 3. |
| B | cssfmt / `css format` (write failure) | 0 (silently muted) | ErrIO → 4 | Previously `_, _ = fmt.Fprintln(...)`; now classified. |
| B | htmlfmt / `html validate` (empty/invalid) | 1 | ErrInvalidInput → 2 | Was `fmt.Errorf("validation failed")`; now cmderr.ErrInvalidInput. |
| B | htmlfmt / `html format` (parse failure from pkghtml) | 1 | ErrInvalidInput → 2 | Was raw `fmt.Errorf("html: %w")`; now classified. |
| B | htmlfmt / `html format` (file not found) | 1 | ErrNotFound → 1 | Previously unclassified; now classified. |
| B | htmlfmt / `html format` (permission denied) | 1 | ErrPermission → 3 | Exit code shifts from 1 to 3. |
| B | htmlfmt / `html format` (write failure) | 0 (silently muted) | ErrIO → 4 | Previously muted; now classified. |
| B | sqlfmt / `sql validate` (parse error) | 1 | ErrInvalidInput → 2 | Was `fmt.Errorf("validation failed")`; now cmderr.ErrInvalidInput. |
| B | sqlfmt / `sql format` (file not found) | 1 | ErrNotFound → 1 | Previously unclassified; now classified. |
| B | sqlfmt / `sql format` (permission denied) | 1 | ErrPermission → 3 | Exit code shifts from 1 to 3. |
| B | sqlfmt / `sql format` (write failure) | 0 (silently muted) | ErrIO → 4 | Previously muted; now classified. |
| B | xmlutil / `xml tojson` (invalid XML) | 1 | ErrInvalidInput → 2 | Was raw `fmt.Errorf("json: invalid XML")`; now classified. |
| B | xmlutil / `xml fromjson` (invalid JSON) | 1 | ErrInvalidInput → 2 | Was raw `fmt.Errorf("xml: invalid JSON")`; now classified. |
| B | xmlutil / `xml tojson|fromjson` (file not found) | 1 | ErrNotFound → 1 | Previously unclassified; now classified. |
| B | xmlutil / `xml tojson|fromjson` (permission denied) | 1 | ErrPermission → 3 | Exit code shifts from 1 to 3. |
| B | xmlutil / `xml tojson|fromjson` (write failure) | 0 (silently muted) | ErrIO → 4 | Previously `_, _ = fmt.Fprint(...)`; now classified. |
| C | ps (invalid sort key, e.g. `ps --sort=bogus`) | 1 | ErrInvalidInput → 2 | New validation; unknown sort keys now classified. |
| C | ps (unix, `/proc` read permission denied) | 1 | ErrPermission → 3 | Exit code shifts from 1 to 3. |
| C | ps (unix, `/proc` read I/O failure) | 1 | ErrIO → 4 | Exit code shifts from 1 to 4. |
| C | ps (windows, `-u` user filter) | 1 | ErrUnsupported → 6 | Locked message: `ps: field user not supported on windows`; pins POLISH-09 format. |
| C | ps (windows, snapshot API failure) | 1 | ErrIO → 4 | CreateToolhelp32Snapshot/Process32First failures now classified. |
| C | pkill (empty pattern) | 1 | ErrInvalidInput → 2 | Was raw `fmt.Errorf`; now classified. |
| C | pkill (invalid regex pattern) | 1 | ErrInvalidInput → 2 | Was `fmt.Errorf("pkill: invalid pattern: %w")`; now classified. |
| C | pkill (invalid signal name/number) | 1 | ErrInvalidInput → 2 | Was `fmt.Errorf("pkill: invalid signal: %s")`; now classified. |
| C | pkill (no match) | 0 | SilentExit(1) → 1 | Previously returned nil (exit 0); now SilentExit(1) per pkill canonical Pattern 4. |
| C | pkill (process listing failure) | 1 | ErrIO → 4 | `process.Processes()` failure now classified. |
| C | pkill (permission denied on signal send) | 1 | ErrPermission → 3 | EPERM on `proc.Signal()` now classified in Result.Error. |
| C | df (unix, path not found) | 1 | ErrNotFound → 1 | Statfs on missing path; exit code unchanged but sentinel set. |
| C | df (unix, permission denied) | 1 | ErrPermission → 3 | Exit code shifts from 1 to 3. |
| C | df (unix, I/O failure) | 1 | ErrIO → 4 | Exit code shifts from 1 to 4. |
| C | df (windows, invalid path string) | 1 | ErrInvalidInput → 2 | UTF16PtrFromString failure; exit code shifts from 1 to 2. |
| C | df (windows, GetDiskFreeSpaceExW failure) | 1 | ErrIO → 4 | Exit code shifts from 1 to 4. |
| C | du (root path not found) | 1 | ErrNotFound → 1 | os.Lstat missing path; exit code unchanged but sentinel set. |
| C | du (root path permission denied) | 1 | ErrPermission → 3 | Exit code shifts from 1 to 3. |
| C | du (root path I/O failure) | 1 | ErrIO → 4 | Exit code shifts from 1 to 4. |
| C | free (unix, sysinfo syscall failure) | 1 | ErrIO → 4 | Exit code shifts from 1 to 4. |
| C | free (windows, GlobalMemoryStatusEx failure) | 1 | ErrIO → 4 | Exit code shifts from 1 to 4. |
| C | yes (write failure, e.g. broken pipe to `head`) | 0 (silently muted) | ErrIO → 4 | Previously `return nil` on any write error; now classified. EPIPE is the expected termination path for `yes \| head`; exit 4 is now the correct sentinel. Document: callers piping yes through head should ignore ErrIO (exit 4) as normal termination. |
| C | uname (write failure, e.g. broken pipe) | 0 (silently muted) | ErrIO → 4 | Previously `_, _ = fmt.Fprintln(...)` silently dropped write errors; now classified. |
| C | lsof (process listing / network query, permission denied) | 1 | ErrPermission → 3 | os.ErrPermission from gopsutil process.Processes() or net.Connections(); exit code shifts from 1 to 3. |
| C | lsof (process listing / network query, I/O failure) | 1 | ErrIO → 4 | Non-permission errors from gopsutil; exit code shifts from 1 to 4. |
| C | ss (JSON write failure) | 0 (silently muted) | ErrIO → 4 | f.Print() errors were previously unchecked; now classified. |
| C | ss (socket print, permission denied) | 1 | ErrPermission → 3 | os.ErrPermission from printSockets(); exit code shifts from 1 to 3. |
| C | ss (socket print, I/O failure) | 1 | ErrIO → 4 | Non-permission errors from printSockets(); exit code shifts from 1 to 4. |
| D | scaffold / WriteTemplate (create/write failure) | 1 | ErrIO → 4 | fs.Create or t.Execute failure now classified; exit code shifts from 1 to 4. |
| D | scaffold / WriteTemplate (template parse failure) | 1 | ErrInvalidInput → 2 | template.New().Parse() failure now classified; exit code shifts from 1 to 2. |
| D | scaffold / WriteLicense (unknown license type) | 1 | ErrInvalidInput → 2 | Unknown license string now classified; exit code shifts from 1 to 2. |
| D | scaffold cobra init (missing --module flag) | 1 | ErrInvalidInput → 2 | Was raw fmt.Errorf; now classified. |
| D | scaffold cobra init (MkdirAll failure) | 1 | ErrIO → 4 | fs.MkdirAll failure now classified; exit code shifts from 1 to 4. |
| D | scaffold cobra add / add-tools (go.mod not found) | 1 | ErrNotFound → 1 | afero.ReadFile failure on go.mod now classified; exit code unchanged but sentinel set. |
| D | scaffold cobra add / add-tools (go.mod parse failure) | 1 | ErrInvalidInput → 2 | Empty module name from go.mod now classified; exit code shifts from 1 to 2. |
| D | scaffold cobra add / add-tools (cmd dir not found) | 1 | ErrNotFound → 1 | fs.Stat failure on cmd/{appName} now classified; exit code unchanged but sentinel set. |
| D | scaffold cobra add / add-tools (file already exists) | 1 | ErrConflict → 1 | Duplicate command/tool file now classified; exit code unchanged but sentinel set. |
| D | scaffold cobra add (empty command name) | 1 | ErrInvalidInput → 2 | Was raw fmt.Errorf; now classified. |
| D | scaffold handler (empty name) | 1 | ErrInvalidInput → 2 | Was raw fmt.Errorf; now classified. |
| D | scaffold handler (MkdirAll failure) | 1 | ErrIO → 4 | fs.MkdirAll failure now classified; exit code shifts from 1 to 4. |
| D | scaffold repository (empty name) | 1 | ErrInvalidInput → 2 | Was raw fmt.Errorf; now classified. |
| D | scaffold repository (MkdirAll failure) | 1 | ErrIO → 4 | fs.MkdirAll failure now classified; exit code shifts from 1 to 4. |
| D | scaffold test (empty source path) | 1 | ErrInvalidInput → 2 | Was raw fmt.Errorf; now classified. |
| D | scaffold test (source file not found) | 1 | ErrNotFound → 1 | os.IsNotExist check now classified; exit code unchanged but sentinel set. |
| D | scaffold test (Go parse failure) | 1 | ErrInvalidInput → 2 | parser.ParseFile failure now classified; exit code shifts from 1 to 2. |
| D | scaffold test (no exported functions) | 1 | ErrInvalidInput → 2 | Was raw fmt.Errorf; now classified. |
| D | scaffold mcp (empty name) | 1 | ErrInvalidInput → 2 | Was raw fmt.Errorf; now classified. |
| D | scaffold mcp (invalid transport) | 1 | ErrInvalidInput → 2 | Unknown transport string now classified; exit code shifts from 1 to 2. |
| D | scaffold mcp (go.mod not found for module detection) | 1 | ErrNotFound → 1 | afero.ReadFile failure on go.mod now classified; exit code unchanged but sentinel set. |
| D | scaffold mcp (MkdirAll failure) | 1 | ErrIO → 4 | fs.MkdirAll failure now classified; exit code shifts from 1 to 4. |
| D | project (path not found) | 1 | ErrNotFound → 1 | os.ErrNotExist from os.Stat now classified; exit code unchanged but sentinel set. |
| D | project (permission denied on path) | 1 | ErrPermission → 3 | os.ErrPermission from os.Stat now classified; exit code shifts from 1 to 3. |
| D | project (I/O failure on stat) | 1 | ErrIO → 4 | Other os.Stat errors now classified; exit code shifts from 1 to 4. |
| D | repo analyze (path not found) | 1 | ErrNotFound → 1 | os.ErrNotExist from resolvePath now classified; exit code unchanged but sentinel set. |
| D | repo analyze (permission denied on path) | 1 | ErrPermission → 3 | os.ErrPermission from resolvePath now classified; exit code shifts from 1 to 3. |
| D | repo analyze (clone failure for remote target) | 1 | ErrIO → 4 | cloneToTemp failure now classified; exit code shifts from 1 to 4. |
| D | repo analyze (output file create failure, -o flag) | 1 | ErrIO → 4 | os.Create failure on output file now classified; exit code shifts from 1 to 4. |
