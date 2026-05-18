# External Integrations

**Analysis Date:** 2026-04-11

## APIs & External Services

**AWS (Amazon Web Services):**
- Service: Multi-service AWS integration
  - SDK/Client: `github.com/aws/aws-sdk-go-v2` v1.41.1
  - Services: EC2, IAM, S3, SSM, STS
  - Auth: AWS credentials (env vars, profiles, IAM roles)
  - Implementation: `internal/cli/aws/` with subpackages for each service
  - Custom endpoint support for LocalStack and other AWS-compatible services

**YouTube & Video Platforms:**
- Service: Video download and metadata extraction
  - SDK/Client: `github.com/dop251/goja` (pure Go JS runtime for signature decryption)
  - Extractors: YouTube, HLS/M3U8, generic HTTP/OG:video fallback
  - Implementation: `pkg/video/extractor/` (registry-based, init()-loaded)
  - Authentication: YouTube InnerTube API (no authentication required, uses client configs)
  - Features: Playlist support, channel enumeration, format selection, resume capability
  - Caching: File cache via `pkg/video/cache/` (XDG paths)

**Kubernetes (kubectl):**
- Service: Kubernetes cluster management
  - SDK/Client: `k8s.io/kubectl` v0.35.0
  - Libraries: `k8s.io/cli-runtime`, `k8s.io/client-go`, `k8s.io/api`
  - Auth: kubeconfig, KUBECONFIG env var, in-cluster auth
  - Implementation: `cmd/kubectl.go` wrapper; full kubectl commands delegated to k8s.io/kubectl
  - Supported: Pod management, deployments, services, ConfigMaps, Secrets, logs, port-forward, exec

**HashiCorp Vault:**
- Service: Secrets management and encryption
  - SDK/Client: `github.com/hashicorp/vault/api` v1.22.0
  - Auth: Token-based (VAULT_TOKEN), AppRole, Kubernetes auth
  - Implementation: `internal/cli/vault/vault.go`
  - Configuration: VAULT_ADDR, VAULT_TOKEN, VAULT_NAMESPACE env vars
  - Operations: Read/write secrets, PKI, encryption-as-a-service

## Data Storage

**Databases:**
- SQLite (via `modernc.org/sqlite` v1.44.3)
  - Usage: Video channel metadata database (`internal/cli/video/channel.go`)
  - Schema: Stores channel info, incremental download tracking
  - Location: Channel folder → `channel.db`
  - Operations: Incremental upsert, query by channel URL
  - Implementation: `internal/cli/video/channeldb.go`

- BoltDB (via `go.etcd.io/bbolt` v1.4.3)
  - Usage: Pure Go embedded key-value store
  - Implementation: `internal/cli/bbolt/` wrapper command
  - Features: Transactions, buckets, key scanning

**File Storage:**
- Local filesystem only
  - XDG Base Directory support: `~/.cache/omni`, `~/.config/omni`
  - Video cache: Uses XDG paths via `pkg/video/cache/`
  - No cloud blob storage integration (S3 integration is for AWS CLI, not internal storage)

**Caching:**
- Filesystem cache (video downloader)
  - Location: `~/.cache/omni/video/` (XDG-compliant)
  - Purpose: Fragment state, retry metadata, session cookies
  - Implementation: `pkg/video/cache/`, `pkg/video/downloader/fragment.go`

## Authentication & Identity

**AWS Authentication:**
- Provider: AWS IAM
- Implementation: `internal/cli/aws/aws.go:LoadConfig()`
- Methods:
  - Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
  - AWS profiles (via `~/.aws/credentials`, `~/.aws/config`)
  - IAM roles (EC2 instance metadata)
  - Custom omni cloud profiles (OMNI_CLOUD_PROFILE, `--profile omni:name`)
- Endpoint override: Support for LocalStack and other AWS-compatible services

**Kubernetes Authentication:**
- Provider: kubeconfig + in-cluster auth
- Implementation: `k8s.io/client-go` standard mechanisms
- Kubeconfig: KUBECONFIG env var, `~/.kube/config` default
- In-cluster: Service account tokens for pod-based auth

**Vault Authentication:**
- Provider: HashiCorp Vault
- Implementation: `internal/cli/vault/vault.go:New()`
- Methods:
  - Token-based (VAULT_TOKEN)
  - AppRole (role_id, secret_id)
  - Kubernetes auth (JWT)
- Configuration: VAULT_ADDR, VAULT_TOKEN, VAULT_NAMESPACE, TLS skip option

**YouTube/Video Platform:**
- Authentication: None required for public content
- InnerTube API: Multiple client configs (Android VR preferred to avoid PoToken requirement)
- Channel cookies: Stored via `pkg/video/nethttp/cookies.go` (Netscape format)

## Monitoring & Observability

**Error Tracking:**
- None detected (no Sentry, Rollbar, etc.)

**Logs:**
- Approach: Structured logging via `log/slog` (mentioned in CLAUDE.md)
- Output: stderr for CLI, JSON for structured data
- Implementation: Commands write to `io.Writer` (stdout) and log to stderr

**Metrics & Performance:**
- Video downloader: Progress tracking and speed estimation
  - Implementation: `pkg/video/downloader/progress.go`
  - Tracks: Download speed, ETA, bytes downloaded, percentage
  - Output: Real-time terminal progress or silent mode (`--no-progress`)

## CI/CD & Deployment

**Hosting:**
- GitHub (repository: `github.com/inovacc/omni`)
- Binary releases via GitHub Releases

**Build & Release Pipeline:**
- Tool: GoReleaser v2
- Config: `.goreleaser.yaml`
- Triggers: Git tags
- Targets: Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64, arm64)
- Output: tar.gz (Linux/macOS), zip (Windows)
- Pre-hooks: `go mod tidy`

**CI Pipeline:**
- GitHub Actions (`.github/workflows/test.yml`)
- Reusable workflow: `inovacc/workflows/.github/workflows/reusable-go-check.yml`
- Checks:
  - Unit tests with race detection (15m timeout)
  - Linting via golangci-lint
  - Vulnerability scanning via govulncheck
  - Total timeout: 30 minutes

**Testing Infrastructure:**
- Docker Compose: `docker/docker-compose.test.yml`
- Test targets:
  - Unit tests (Go race detector)
  - Black-box tests (Python-based)
  - Golden master tests (snapshot comparison)
  - Video comparison tests (omni vs yt-dlp)
- Dockerfile multi-stage builds for Linux, Windows, video testing

**Taskfile Automation:**
- Task runner: `Taskfile.yml` (v3)
- Key tasks:
  - `task build` - Compile binary
  - `task test` - Run all tests with coverage
  - `task test:integration` - Integration tests
  - `task test:blackbox` - Black-box tests
  - `task test:golden` - Golden master tests
  - `task docker:test:*` - Docker-based testing

## Environment Configuration

**Required Environment Variables:**
- AWS operations:
  - `AWS_PROFILE` - Profile name
  - `AWS_REGION` - Region
  - `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` - Credentials (alternative to profile)
- Vault operations:
  - `VAULT_ADDR` - Server address
  - `VAULT_TOKEN` - Auth token
  - `VAULT_NAMESPACE` - Optional namespace
- Kubernetes:
  - `KUBECONFIG` - Path to kubeconfig
- Video:
  - No env vars required for basic operation
- Build:
  - `CGO_ENABLED` - Set to 0 for cross-platform static builds
  - `GOOS`, `GOARCH` - For cross-compilation

**Optional Environment Variables:**
- `OMNI_CLOUD_PROFILE` - Custom cloud profile name (alternative to `--profile omni:name`)

**Secrets Location:**
- AWS: `~/.aws/credentials`, `~/.aws/config`
- Kubernetes: `~/.kube/config`
- Vault: VAULT_TOKEN env var or ~/.vault-token
- YouTube session: `~/.cache/omni/video/cookies.txt` (Netscape format)
- General: No .env files checked in; secrets via env vars only

## Webhooks & Callbacks

**Incoming:**
- None detected

**Outgoing:**
- None detected

## Notable Implementation Details

**YouTube Signature Decryption:**
- Method: Pure Go JavaScript interpreter (goja)
- Fallback: Multiple InnerTube client configs (Android VR first)
- Purpose: Extract player JavaScript and decrypt signature/nsig parameters
- Cache: Decrypted signatures stored to avoid re-downloading player JS

**AWS Cloud Profiles:**
- Custom integration: `internal/cli/cloud/profile/` for omni-specific profiles
- Prefix notation: `--profile omni:profilename` for custom profiles
- Fallback: Standard AWS SDK profiles if no omni prefix

**Video Download Resume:**
- Fragment state: Persisted to `~/.cache/omni/video/`
- `.part` files: Incomplete downloads
- Range requests: HTTP Range header for resume capability

**Kubernetes Integration:**
- Direct delegation to k8s.io/kubectl (no custom logic)
- Supports all kubectl subcommands and flags
- Kubeconfig resolution: Standard KUBECONFIG env var, ~/.kube/config

---

*Integration audit: 2026-04-11*
