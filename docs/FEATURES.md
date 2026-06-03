# Feature Tracker

> Last updated: 2026-05-24

---

## Completed Features

### Core Shell Utilities (120+ commands)
- **Status:** Complete
- File operations, text processing, system info, archive, compression
- All commands use io.Writer pattern for testability
- Cross-platform (Linux, macOS, Windows)

### Search Engine (rg)
- **Status:** Complete
- Ripgrep-compatible search with gitignore support
- Parallel directory walking, streaming JSON output
- File type filtering, context lines, glob patterns

### Data Processing
- **Status:** Complete
- jq (JSON query), yq (YAML query), dotenv parser
- JSON/CSV/XML bidirectional conversions
- YAML/TOML/XML validation

### Formatting Engines
- **Status:** Complete
- SQL, CSS, HTML format/minify/validate
- Protobuf tooling (buf lint, format, compile, breaking, generate)

### Encoding & Hashing
- **Status:** Complete
- Base64, Base32, Base58, hex, URL, HTML encode/decode
- MD5, SHA1, SHA256, SHA512, CRC32, CRC64 checksums
- AES-256-GCM encryption/decryption
- UUID v4/v7, ULID, KSUID, Nanoid, Snowflake generation

### Video Download Engine
- **Status:** Complete
- Pure Go youtube-dl/yt-dlp port
- YouTube (InnerTube API, signature decryption, playlists)
- HLS/M3U8 with AES-128 decryption
- Format selection, resume support, rate limiting

### Streaming Engines
- **Status:** Complete
- `pipe`: Cobra-dispatched command chaining with variable substitution
- `pipeline`: Streaming io.Pipe stages (20 built-in transforms)

### Cloud & DevOps
- **Status:** Complete
- Kubernetes (kubectl via k8s.io/kubectl), 17 kubectl hacks
- Terraform CLI wrapper (20+ subcommands)
- AWS SDK (EC2, S3, IAM, STS, SSM)
- HashiCorp Vault integration
- Git hacks (12 shortcuts)

### Database Management
- **Status:** Complete
- SQLite (pure Go, no CGO): query, stats, tables, schema, dump, import
- BBolt: buckets, keys, get/put/delete, compact, check

### Code Scaffolding
- **Status:** Complete
- `scaffold cobra init/add/add-tools/config` CLI application scaffolding
- Handler, repository, test code generators
- cmdtree and aicontext template generation
- All scaffolding functions accept `afero.Fs` for filesystem abstraction (in-memory testing)
- `scaffold cobra add --platform-split` emits `cmd_<name>_{windows,darwin,unix}.go` with build tags so each platform supplies exactly one implementation
- `scaffold cobra init --daemon` generates a complete self-daemonizing service (PID-file state via gopsutil, foreground/install/uninstall, systemd unit / launchd plist / Windows SCM)

### Process (runtime-aware)
- **Status:** Complete
- `omni gops` — list, kill, inspect, monitor, obfuscation, top (bubbletea TUI), agent-cmd, trace, profile, stream
- `omni nodeps`, `omni pyps`, `omni javaps` — list + kill for Node/Python/Java processes
- Pure Go via `debug/buildinfo` + `gopsutil`; no exec, no agent required for list/kill/inspect/monitor/obfuscation/top
- Safety: `kill <name>` matching >1 process requires `--recursive`; `--recursive` requires `--yes`
- Embeddable agent in `pkg/gopsagent` — 3 lines for Go programs to expose runtime introspection over loopback TCP; optional HMAC challenge; optional startup notification via `~/.config/gops/config.json` or `GOPS_AGENT_NOTIFY=1`

### Project Analyzer
- **Status:** Complete
- Project type detection (11 ecosystems)
- Dependency analysis, documentation status
- Git info, health scoring (0-100, A-F grade)

### Repository Context Generator
- **Status:** Complete
- `repo analyze` — structured Markdown/JSON context optimized for LLM consumption
- Directory tree, key file contents, entry points, architecture inference
- API surface analysis (exported func counts per pkg/ package)
- Test pattern detection, CI/CD config, config files
- Remote repository support (clones via gh/git to temp dir)
- Section filtering, compact mode, file output

### Reusable pkg/ Libraries
- **Status:** Complete (21 packages)
- idgen, hashutil, jsonutil, encoding, cryptutil, sqlfmt, cssfmt, htmlfmt
- textutil, search/grep, search/rg, twig, pipeline, video, figlet, userdirs, cobra/helper
- procutil, procmetrics, obfuscate, gopsagent (runtime-aware process tools, see Process feature above)

### Testing Infrastructure
- **Status:** Complete
- Go unit tests (700+ test cases)
- Python black-box test suite (15 test scripts)
- Golden master tests (117 tests, 13 categories)
- Docker test environments (Linux, Windows)
- Benchmark suite (Go micro + coreutils macro + modern tools)

---

## Proposed Features

### P1 - High Priority

| Feature | Description | Complexity |
|---------|-------------|------------|
| ~~Golden master testing~~ | ~~Characterization tests capturing command output snapshots~~ | ~~DONE (81 tests, 11 categories)~~ |
| Unified output formatter | text/json/table output modes for all commands | Medium |
| `--json` flag everywhere | Add JSON output to remaining commands | Low |
| Consistent exit codes | Standardize error codes across all commands (49/160+ adopted) | Low |

### P2 - Medium Priority

| Feature | Description | Complexity |
|---------|-------------|------------|
| YAML formatter | `yaml fmt` with key sorting, indentation normalization | Medium |
| K8s YAML formatter | `yaml k8s` with standard key ordering | Medium |
| GitHub hacks | PR checkout, diff, approve, issue management shortcuts | Medium |
| Consul integration | KV store, service catalog, cluster members | High |
| Nomad integration | Job, node, allocation management | High |

### P3 - Low Priority

| Feature | Description | Complexity |
|---------|-------------|------------|
| Plugin system | Extensible command loading | High |
| WASM build target | Run omni in browser/edge environments | High |
| Filter DSL | `--where` conditions for structured output | Medium |
| AI code generation | Context-aware code generation with LLM providers | High |
