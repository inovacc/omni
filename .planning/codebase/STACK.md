# Technology Stack

**Analysis Date:** 2026-04-11

## Languages

**Primary:**
- Go 1.25.0 - All core logic, CLI commands, libraries, and utilities

## Runtime

**Environment:**
- Go runtime (no external runtime dependencies)
- Cross-platform: Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64, arm64)

**Package Manager:**
- Go Modules (`go.mod`, `go.sum`)
- Lockfile: `go.sum` present

## Frameworks

**Core CLI:**
- Cobra v1.10.2 - Command-line interface framework (located in `cmd/`)

**Testing:**
- Go testing package (`testing` stdlib)
- Table-driven test patterns throughout

**TUI & Output:**
- Charmbracelet Bubbletea v1.3.10 - Terminal UI framework (used by pagers)
- Charmbracelet Lipgloss v1.1.0 - Terminal styling and ANSI formatting

**Build & Release:**
- GoReleaser v2 - Cross-platform binary building and releases (config: `.goreleaser.yaml`)
- Task v3 - Task automation (config: `Taskfile.yml`)

## Key Dependencies

**Critical:**
- `github.com/spf13/cobra` v1.10.2 - CLI command framework (all commands)
- `github.com/charmbracelet/bubbletea` v1.3.10 - TUI pagers (less, more)
- `github.com/dop251/goja` v0.0.0-20260106131823 - Pure Go JavaScript runtime (YouTube signature decryption in video package)
- `gopkg.in/yaml.v3` v3.0.1 - YAML parsing (yq, dotenv, config parsing)

**Infrastructure & Cloud:**
- `github.com/aws/aws-sdk-go-v2` v1.41.1 - AWS SDK (with services: EC2, IAM, S3, SSM, STS)
- `github.com/hashicorp/vault/api` v1.22.0 - HashiCorp Vault client (secret management)
- `k8s.io/kubectl` v0.35.0 - Kubernetes kubectl client (with `k8s.io/cli-runtime`)

**Databases & Storage:**
- `go.etcd.io/bbolt` v1.4.3 - Pure Go key-value store (embedded database)
- `modernc.org/sqlite` v1.44.3 - Pure Go SQLite driver (channel DB for video downloader, SQLite shell)

**Cryptography & Security:**
- `golang.org/x/crypto` v0.48.0 - Stdlib crypto extensions (PBKDF2 for encryption)
- `github.com/btcsuite/btcd/btcutil` v1.1.6 - Bitcoin utilities (Base58 encoding)

**Code Generation & Parsing:**
- `github.com/bufbuild/protocompile` v0.14.1 - Pure Go protobuf compiler (buf format/lint)
- `google.golang.org/protobuf` v1.36.11 - Protocol Buffers runtime

**ID Generation:**
- `github.com/segmentio/ksuid` v1.0.4 - Sortable unique IDs (KSUID generation)

**Filesystem & Template:**
- `github.com/spf13/afero` v1.15.0 - Filesystem abstraction (testable file operations in scaffolding)
- Standard `text/template` - Code generation templates

**System Monitoring:**
- `github.com/shirou/gopsutil/v3` v3.24.5 - OS/process utilities (ps, df, du, free, uptime)
- `github.com/google/gops` v0.3.29 - Go process analysis

**Utilities:**
- `github.com/BurntSushi/toml` v1.6.0 - TOML parsing
- `github.com/fatih/color` v1.18.0 - Colored terminal output
- `github.com/xlab/treeprint` v1.2.0 - Tree printing (legacy, replaced by `pkg/twig`)

**Standards Library Usage:**
- `os`, `io`, `io/fs` - File operations
- `path/filepath` - Path manipulation
- `regexp` - Pattern matching (grep, sed, rg)
- `sync`, `sync/atomic` - Concurrency (parallel scanning in rg, tree)
- `encoding/json` - JSON operations (jq, output)
- `encoding/xml` - XML parsing/output
- `encoding/csv` - CSV operations
- `archive/tar`, `archive/zip` - Archive operations
- `compress/gzip`, `compress/bzip2` - Compression
- `crypto/*` - Hashing, encryption
- `syscall` - System calls (for df, ps, etc.)
- `go/parser`, `go/ast` - Go code analysis (scaffold testgen)

## Configuration

**Environment Variables:**
- `VAULT_ADDR` - Vault server address (default: `https://127.0.0.1:8200`)
- `VAULT_TOKEN` - Vault authentication token
- `VAULT_NAMESPACE` - Vault namespace
- `AWS_PROFILE` - AWS profile selection
- `AWS_REGION` - AWS region
- `OMNI_CLOUD_PROFILE` - Custom omni cloud profile (alternative to `--profile omni:name`)
- `CGO_ENABLED` - Enables/disables CGO (set to 0 for cross-platform builds)

**Build Configuration:**
- `.golangci.yml` - Linting configuration (golangci-lint)
  - Linters: govet
  - Exclusions: pkg/buf (vendored), generated code
  - Timeout: 10 minutes
- `.goreleaser.yaml` - Release binary building
  - Targets: Linux, Windows, macOS (amd64, arm64)
  - CGO disabled for static binaries
  - Output formats: tar.gz (Linux/macOS), zip (Windows)

**Testing Configuration:**
- `Taskfile.yml` - Task automation for build, test, golden tests
- `docker/docker-compose.test.yml` - Containerized test environments
  - Unit test target (with race detector)
  - Black-box testing (Python-based)
  - Golden master tests (snapshot comparison)
  - Video comparison tests (omni vs yt-dlp)

**CI/CD:**
- `.github/workflows/test.yml` - GitHub Actions test pipeline
  - Reusable workflow from `inovacc/workflows`
  - Runs: tests (15m timeout), linting, vulnerability checks
  - Timeout: 30 minutes total
- `.github/workflows/release.yml` - Release pipeline (GoReleaser)

## Platform Requirements

**Development:**
- Go 1.25.0 or later
- Task 3.x for local automation
- Docker (optional, for containerized tests)
- Python 3.8+ (for black-box tests)

**Production:**
- Target: Linux, macOS, Windows
- Architecture: x86_64 (amd64), ARM64 (arm64)
- No external runtime dependencies (static binaries via CGO_ENABLED=0)
- Supports AWS SDK (for aws commands)
- Supports kubectl (for k8s commands)
- Supports HashiCorp Vault (optional, for vault commands)

---

*Stack analysis: 2026-04-11*
