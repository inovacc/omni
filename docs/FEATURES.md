# Feature Tracker

> Last updated: February 2026

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

### Code Generation
- **Status:** Complete
- Cobra CLI application scaffolding
- Handler, repository, test code generators

### Project Analyzer
- **Status:** Complete
- Project type detection (11 ecosystems)
- Dependency analysis, documentation status
- Git info, health scoring (0-100, A-F grade)

### Reusable pkg/ Libraries
- **Status:** Complete (16 packages)
- idgen, hashutil, jsonutil, encoding, cryptutil, sqlfmt, cssfmt, htmlfmt
- textutil, search/grep, search/rg, twig, pipeline, video, figlet, buf

### Testing Infrastructure
- **Status:** Complete
- Go unit tests (700+ test cases)
- Python black-box test suite (15 test scripts)
- Golden master tests (81 snapshot tests, 11 categories)
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
| Consistent exit codes | Standardize error codes across all commands | Low |

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
