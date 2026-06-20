# Backlog

Prioritized items for future development.

> Last updated: 2026-05-24

---

## Completed

All core, file, text, system, archive, encoding, hashing, data, formatting, search, flow,
security, pager, comparison, and tooling commands are implemented (160+ commands).

Completed integrations: Kubernetes, Terraform, Vault, AWS (EC2/S3/IAM/STS/SSM).
Completed hacks: Git (12 shortcuts), Kubectl (17 shortcuts).
Completed engines: `pipe` (Cobra dispatch), `pipeline` (streaming io.Pipe stages).

See CLAUDE.md for the full command inventory.

---

## High Priority (P0)

### Core Infrastructure
- [x] Unified `Command` interface contract (`internal/cli/command/` — interface, Registry, adapters)
- [x] Consistent error model with exit codes (`cmderr` 100% adoption — all internal/cli/ commands, Phase 1 Apr 2026)
- [x] Add `--json` flag to remaining commands that lack it
- [x] Unified output formatter (text/json/table)
- [x] cmderr rollout — all commands adopted (Phase 1 complete, Apr 2026)
- [x] Migrate `pipe` command to use `command.Registry` for dispatch (with Cobra fallback)
- [x] Expand pipe Registry to 24 commands (awk, fold, column, paste, xxd, grep, tr, hash, base64, base32, caseconv, strings, shuf added)

### cmderr Phase 2 follow-ups (deferred from Phase 1)

- [x] **[DONE 2026-06-04] Static-analysis rule for raw `os.ErrX` returns:** Implemented as `tools/cmderrlint` — a pure-Go (go/ast, no `golang.org/x/tools`, honoring the lean-`go.mod` invariant — a `go/analysis` analyzer would pull `x/tools`) regression guard, wired blocking into `task lint:cmderr` + the `test.yml` quality job. Found and fixed a real bug (`lsof` network-mode permission errors exited 1 instead of 3). Not a golangci-lint plugin by design; runs as a standalone vet-style check.
- [x] **[DONE 2026-06-04] Cross-command exit-code matrix:** Satisfied by the `cmderr_wave_a/b/z(_risky)` golden categories (per-command exit-code pinning across many commands) + `TestExitCodeFor` (full sentinel→code unit matrix incl. `WithExitCode`/`SilentExit`) + `docs/EXIT-CODES.md` (the human-readable contract).
- [x] **[DONE] `cmderr.Is<Class>()` convenience helpers:** Present in `internal/cli/cmderr/cmderr.go` — `IsNotFound`, `IsInvalidInput`, `IsPermission`, `IsIO`, `IsConflict`, `IsTimeout`, `IsUnsupported`.
- [x] **[DONE] `docs/EXIT-CODES.md`:** Exists with the full exit-code ↔ sentinel ↔ `Is*` table; referenced by `CONTRIBUTING.md`.

### no-exec items — RESOLVED via boundary decision (2026-06-03)

Resolved by the no-exec boundary decision: the invariant governs *utility reimplementations*; commands whose **purpose** is to orchestrate an external tool are sanctioned exceptions. See `docs/architecture/patterns.md` § "No-exec invariant: scope & sanctioned exceptions" and `docs/quality/HARDENING.md` § "Resolution status".

- [x] **[RESOLVED — accepted] `internal/cli/exec/exec.go`:** ACCEPTED as a sanctioned exec wrapper (the launcher *is* the feature). Documented, not a violation. (was: no-exec-violation, Plan 14.)
- [x] **[RESOLVED — accepted] `internal/cli/repo/remote.go`:** ACCEPTED as a sanctioned exec wrapper (`git`/`gh` clone orchestration). Documented. (was: no-exec-violation, Plan 13.) Optional future enhancement: pure-Go clone via `go-git` — nice-to-have, not required.

### Hardening deferrals (2026-06-04 second-pass audit)

Open items from `docs/quality/HARDENING-2026-06-04.md` (all 19 confirmed findings of that audit are FIXED; these are the second-pass triage residue):

- [ ] **[DECISION / no-exec] `internal/cli/video/auth.go` headless-Chrome exec sink:** the YouTube cookie extractor spawns Chrome via `exec.Command`, which is NOT in the documented sanctioned-exec allowlist (`exec`/`forloop`/`task`/`terraform`/`repo`/git-gh-hacks/`buf generate`/machineid/uname). Maintainer decision needed: sanction + document it (the launcher *is* the feature, like `exec`/`repo`) or remove the path. The CDP read-cap and dial-SSRF defenses around it are already fixed.
- [ ] **[BACKLOG / LOW] `dotenv` `FormatExport` cmd-`set` escaping:** the `eval "$(omni dotenv -e)"` path: the cmd.exe `set` branch does no escaping and the KEY is never escaped in any shell branch. LOW (operator-owned `.env` trust boundary); a robust per-shell key+value escaper (incl. cmd quoting) is the fix.
- [ ] **[BACKLOG] second-pass audit completion:** the verify workflow hit a `StructuredOutput` tooling error; re-run a focused audit of the YAML alias-bomb (`yaml.Unmarshal` of untrusted Taskfile/buf configs) and bespoke recursive-parser depth (`jsonfmt`/`tomlutil`/`cssfmt`/`htmlfmt`/`sqlfmt`) — speculative, unconfirmed. Also apply the scaffolding name-traversal guard to the `cobra`/`mcp` scaffolders for parity.

### Hardening deferrals (from security/robustness audit, 2026-06-03)

Surfaced by the hardening sweep on `harden/audit-fixes`; full context in `docs/quality/HARDENING.md`.

- [ ] **[DEPRECATION] Remove `encoding.Base58Decode` after 2026-07-19:** replaced by `Base58DecodeStrict` (reports invalid input). All internal callers migrated; remove in a dedicated cleanup commit.
- [ ] **[BACKLOG / DEPRECATION] crypto-02 — versioned ciphertext envelope:** `pkg/cryptutil` PBKDF2 default iteration count (100k) is below current OWASP guidance, but cannot be raised: the iteration count is NOT stored in the ciphertext envelope, so bumping the default would make existing default-cost blobs undecryptable. Needs a versioned envelope (store the iteration count + version byte) → dual-read old/new → raise default → cleanup after the 30-day deprecation window. The floor-validation (crypto-01) was already applied.
- [ ] **[BACKLOG / no-exec] machineid_darwin — remove `ioreg` dependency:** macOS `getMachineID` still spawns `ioreg` to read `IOPlatformUUID`; kept as a documented machine-identity exec exception because there is no pure-Go/no-cgo path to the *same* value and the ID feeds the master-key KDF (changing it bricks existing `master.key`). Future fix requires a pure-Go-reachable UUID source OR a `master.key` re-encryption migration on identifier change.
- [x] **[DONE 2026-06-03] `internal/cli/exec` test hermeticity:** Root cause — on Windows `os.UserHomeDir()` reads `USERPROFILE`, not `HOME`, so tests setting only `HOME` still saw the dev's real `~/.aws`/`~/.npmrc`/`~/.netrc`. Fixed test-only: set `USERPROFILE` alongside `HOME` to a `t.TempDir()`, isolate `PATH`, create a fake `.netrc` in the temp HOME, and assert the strict-mode `cmderr.ErrInvalidInput` sentinel fires before any exec. Production code confirmed correct.
- [x] **[DONE 2026-06-03] `internal/cli/dotenv` export-format test:** Pinned `t.Setenv("SHELL", "/bin/bash")` so `DetectShell()` selects the POSIX branch deterministically on every OS (SHELL is checked before `runtime.GOOS`). Test-only; production behavior unchanged.

### Phase 06 — scan reachability (deferred 2026-06-03)

- [ ] **[BACKLOG] `contrib/govulncheck-scan` — reachability source scanning:** `omni scan source <dir>` (report only vulns whose symbol is actually called) was dropped from v1.0 (ADR-0008). Its only implementation, `golang.org/x/vuln`, (a) loads packages via `x/tools/go/packages` which execs `go list` (violates the no-exec rule) and (b) pulls `golang.org/x/vuln` + `golang.org/x/tools` into the main `go.mod` via MVS even behind a build tag (violates the lean-`go.mod` rule, ADR-0007). Future home: a self-contained `contrib/govulncheck-scan` module (mirroring `contrib/sigstore-verify`) under a sanctioned `go list` exec exception, decided in a follow-up ADR. Until then `omni scan source` returns `cmderr.ErrUnsupported` (exit 6).

### Pre-existing flaky tests

- [x] **[DONE 2026-06-03] `internal/cli/exec/detector_test.go::TestDetectGo_PrivateWithNetrc`:** Fixed as part of the exec test-hermeticity work above — now sets `USERPROFILE` and creates a fake `.netrc` in the temp HOME, so it is deterministic on Windows too.

---

## Medium Priority (P1)

### Data Formatting ✅
- [x] `yaml fmt` - YAML formatter/beautifier (indentation, key sorting, remove empty)
- [x] `yaml k8s` - Kubernetes YAML formatter (standard key ordering, multi-document)

### GitHub Hacks ✅
- [x] `gh-pr-checkout` - Checkout PR by number
- [x] `gh-pr-diff` - Show PR diff locally
- [x] `gh-pr-approve` - Quick approve PR
- [x] `gh-issue-mine` - List issues assigned to me
- [x] `gh-repo-clone-org` - Clone all repos from org
- [x] `gh-actions-rerun` - Rerun failed workflow

### Cloud & DevOps Integrations

#### Consul Integration
- [ ] `consul members` - List cluster members
- [ ] `consul kv` - Key-value store operations
- [ ] `consul services` - Service catalog operations

#### Nomad Integration
- [ ] `nomad job` - Job management
- [ ] `nomad node` - Node operations
- [ ] `nomad alloc` - Allocation operations

#### Packer Integration
- [ ] `packer build` - Build images
- [ ] `packer validate` - Validate templates
- [ ] `packer fmt` - Format templates

---

## Medium-Low Priority (P1.5)

### Feature Completeness (relocated from ISSUES.md, 2026-06-03)
- [ ] `sed` does not implement the full GNU sed feature set (multi-line, hold space, branching) — only basic substitution patterns today.
- [ ] `awk` does not implement the complete AWK language specification — covers common patterns only.
- [ ] `rg` fidelity gaps: binary file detection is heuristic (null-byte check) and may misclassify files; `.gitignore` nested-negation edge cases may differ from ripgrep.
- [ ] Video download limits: YouTube signature decryption is fragile (depends on goja JS runtime; YouTube player-JS changes require updates); no SAMPLE-AES HLS (only AES-128-CBC); no FFmpeg merge for video+audio (note: FFmpeg merge conflicts with the no-exec invariant).

### Video Enhancements
- [x] `omni video download <ID>` — shortcut to download by bare 11-char YouTube ID (auto-resolves to full URL)
- [ ] Add `--description` flag to include video description in output/sidecar

### Tree Enhancements
- [ ] `omni tree` optimize with multi-analyzer architecture
  - Spawn multiple analyzers writing to temp files, then assemble final structure
  - Add `--longest-path` flag to find and display the longest path in the tree

### Curl Enhancements ✅
- [x] `omni curl --json <url>` — auto-pretty-print JSON responses via global `--json` flag

### Tool Analysis & Enhancement
- [ ] Audit all 155+ tools for consistency, missing features, edge cases
- [ ] `omni sqlite` — enhance with interactive mode, query history, output formats

### Windows Features
- [ ] `omni regedit` — Windows registry mapping/viewer
  - Read/write/list registry keys, export/import, search

### Note Enhancements
- [ ] `omni note open` — open temporary file in `$EDITOR`, then copy content into the note file on save/close

### P2P Chat
- [ ] `omni mini chat` — peer-to-peer chat using gossip protocol
  - Invite key logic extracted from Syncthing discovery/relay model
  - Encrypted, decentralized, no server required

---

## Low Priority (P2)

### Advanced Features
- [ ] Plugin system
- [ ] WASM build target
- [ ] Embedded mode for other tools
- [ ] Filter DSL (`--where` conditions)

### AI Code Generation
- [ ] `ai generate` - AI assistant for code generation
  - Configurable system prompts, multiple providers
  - Context-aware code generation
  - Code review, explain, refactor modes

---

## Technical Debt

- [ ] Standardize flag naming across commands
- [ ] Add context.Context to all long-running operations (Command interface provides this path)
- [~] Improve error messages with actionable suggestions (cmderr sentinels provide structured exit codes)
- [x] Split large packages: archive.go → archive.go + tar.go + zip.go; pipe.go → pipe.go + parse.go + substitute.go + execute.go (Mar 2026)
- [ ] Migrate remaining Run signatures to `command.Command` interface incrementally

---

## Testing

### Current Status (May 2026)
- **Total Test Cases:** ~800+ tests across all packages (+92 added in May 2026 for procutil/obfuscate/procmetrics/gopsagent/gopsclient/runtimeps)
- **Overall Coverage:** 59.4% (includes vendored buf packages after flattening)
- **Omni-owned pkg/ avg:** ~78% (24 of 31 packages above 80%; new gops packages drag the avg slightly until backlog item is addressed)
- **New (May 2026):** pkg/procmetrics 93.8%, pkg/gopsagent 59.1%, pkg/procutil 57.4%, pkg/obfuscate 55.6%, internal/cli/runtimeps 34.9%, internal/gopsclient 31.3%
- **Packages with new tests (Feb 2026):** twig/builder (58.9%), twig/parser (79.1%), video/jsinterp, video/downloader (progress, fragment, selector), video/nethttp (cookies, SAPISID), video/extractor (helpers, M3U8), video/options

### Recently Resolved
- [x] scaffold cobra generates `cmd/{appName}/` structure instead of `main.go` + `cmd/` (Mar 2026)
- [x] Tests for twig/builder (58.9%) and twig/parser (79.1%) — completed Feb 2026
- [x] scaffolding refactor — `generate` renamed to `scaffold`, reorganized into domain subpackages (Feb 2026)
- [x] cmderr batches 4-5 — 20 more commands adopted (Feb 2026)
- [x] afero refactor — scaffolding packages accept `afero.Fs` for in-memory testing (Feb 2026)
- [x] cmderr batches 6-7 — 24 more commands adopted, total now 73 (Mar 2026)
- [x] pipe registry expanded to 18 commands with unified dispatch (Mar 2026)
- [x] archive.go split into archive.go + tar.go + zip.go (Mar 2026)
- [x] pipe.go split into pipe.go + parse.go + substitute.go + execute.go (Mar 2026)
- [x] pipeline CLI wrapper now propagates context.Context (Mar 2026)
- [x] cmderr batch 8 — 11 more commands adopted: uuid, random, caseconv, jwt, note, jsonfmt, htmlenc, tomlutil, xmlfmt, pwd, exist (Mar 2026)
- [x] rg package threaded with context.Context for cancellation support (Mar 2026)
- [x] pipe Registry expanded to 24 commands with hash, base64, base32, caseconv, strings, shuf (Mar 2026)
- [x] Runtime-aware process tools shipped: `omni gops` (10 subcommands) + `nodeps`/`pyps`/`javaps` (May 2026)
- [x] Embeddable runtime agent shipped as `pkg/gopsagent` with HMAC challenge + notify-on-startup config (May 2026)
- [x] Cobra scaffolder: `--platform-split` (cmd_<name>_{windows,darwin,unix}.go) + `--daemon` (full PID-file service template with systemd/launchd/SCM install) (May 2026)

### Remaining
- [ ] **Coverage gap in new gops packages (May 2026):** `internal/gopsclient` 31.3%, `internal/cli/runtimeps` 34.9%, `pkg/obfuscate` 55.6%, `pkg/procutil` 57.4%, `pkg/gopsagent` 59.1%. Target 80% per project policy. `pkg/procmetrics` already at 93.8%. Lower numbers reflect TUI/cobra-flag code paths and platform-conditional kill impls that need OS-specific test fixtures.
- [ ] **Coverage gap (relocated from ISSUES.md, 2026-06-03):** `pkg/video/extractor/generic` has no unit tests (P2); `pkg/video/extractor/youtube` has minimal tests (~4.0%, P2). Target 80% per project policy.
- [ ] Platform-specific tests (Windows edge cases, symlinks, permissions)
- [ ] Large file (>1GB) handling tests
- [ ] Benchmarks vs GNU tools (sort, grep, file operations)
- [x] Golden tests with expected output files (117 tests, 13 categories including buf/protobuf)
- [ ] CI coverage threshold enforcement (80%)

---

## Documentation

- [ ] Full command reference (man-page style)
- [ ] Library API documentation (godoc)
- [ ] Migration guide from shell scripts
- [ ] Taskfile integration examples

---

## CI/CD

### Done
- [x] GitHub Actions test workflow (lint, fmt, vulncheck, test -race)
- [x] Release workflow

### Remaining
- [ ] Multi-platform builds (Linux, macOS, Windows)
- [ ] Multi-arch builds (amd64, arm64)
- [ ] Automated releases with goreleaser
- [ ] Code coverage reporting + badges
- [ ] Coverage threshold enforcement

---

## Cross-Platform Notes

| Command | Linux | macOS | Windows | Notes |
|---------|:-----:|:-----:|:-------:|-------|
| `chmod` | ✅ | ✅ | ⚠️ | Limited permission model |
| `chown` | ✅ | ✅ | ❌ | Not applicable |
| `df` | ✅ | ✅ | ⚠️ | Different syscalls |
| `free` | ✅ | ⚠️ | ⚠️ | /proc vs sysctl |
| `ps` | ✅ | ✅ | ⚠️ | Different APIs |
| `ln -s` | ✅ | ✅ | ⚠️ | Requires admin |
| `kill` | ✅ | ✅ | ⚠️ | Signal handling differs |

---

## Version Milestones

### v0.1.0 - MVP ✅
Core commands, JSON output, basic docs

### v0.2.0 - File Operations ✅
cp, mv, rm, mkdir, stat, touch, ln, readlink, chmod, chown

### v0.3.0 - Text Processing ✅
grep, head, tail, sort, uniq, wc, cut, tr, sed, awk, rg, pipeline

### v0.4.0 - System Info ✅
df, du, env, whoami, uname, time, uptime, free, ps, kill, id, lsof

### v0.5.0 - Test Coverage ✅
97.7% package coverage (86/88), 700+ test cases

### v0.6.0 - Cloud & DevOps ✅
Kubernetes, Terraform, Vault, AWS, Git hacks, Kubectl hacks

### v0.7.0 - Engines & Media ✅
pipe (Cobra dispatch), pipeline (streaming), video (pure Go youtube-dl), buf (protobuf)

### v1.0.0 - Production Ready
- All P0 infrastructure items complete
- Full documentation
- 80%+ overall test coverage
- CI coverage enforcement
