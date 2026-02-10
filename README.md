# omni

[![Test](https://github.com/inovacc/omni/actions/workflows/test.yml/badge.svg)](https://github.com/inovacc/omni/actions/workflows/test.yml)

A cross-platform, Go-native replacement for common shell utilities, designed for Taskfile, CI/CD, and enterprise environments.

## Features

- **No exec** - Never spawns external processes
- **Pure Go** - Standard library first, minimal dependencies
- **Cross-platform** - Linux, macOS, Windows
- **Library + CLI** - Use as commands or import as Go packages
- **Safe defaults** - Destructive operations require explicit flags
- **Unix compatible** - GNU-style flags for find (`-name`), head/tail (`-20`)

## Installation

```bash
go install github.com/inovacc/omni@latest
```

Or build from source:
```bash
git clone https://github.com/inovacc/omni.git
cd omni
task build
```

## Quick Start

```bash
# File operations
omni ls -la
omni cat file.txt
omni cp -r src/ dst/
omni rm -rf temp/

# Text processing
omni grep -i "pattern" file.txt
omni sed 's/old/new/g' file.txt
omni awk '{print $1}' data.txt
omni jq '.name' data.json
omni yq '.items[]' config.yaml

# System info
omni ps -a
omni df -h
omni free -h
omni uptime

# Utilities
omni sha256sum file.bin
omni base64 -d encoded.txt
omni uuid -n 5
omni random -t password -l 20

# Encryption
echo "secret" | omni encrypt -p mypass -a
omni decrypt -p mypass -a secret.enc
```

## Command Categories

### Core Commands
| Command | Description |
|---------|-------------|
| `ls` | List directory contents |
| `pwd` | Print working directory |
| `cat` | Concatenate and print files |
| `date` | Print current date/time |
| `dirname` | Strip last path component |
| `basename` | Strip directory from path |
| `realpath` | Print resolved absolute path |

### File Operations
| Command | Description |
|---------|-------------|
| `cp` | Copy files and directories |
| `mv` | Move/rename files |
| `rm` | Remove files/directories |
| `mkdir` | Create directories |
| `rmdir` | Remove empty directories |
| `touch` | Update file timestamps |
| `stat` | Display file status |
| `ln` | Create links |
| `readlink` | Print symlink target |
| `chmod` | Change file permissions |
| `chown` | Change file ownership |

### Text Processing
| Command | Description |
|---------|-------------|
| `grep` | Search for patterns |
| `egrep` | Extended regex grep |
| `fgrep` | Fixed string grep |
| `head` | Output first lines |
| `tail` | Output last lines |
| `sort` | Sort lines |
| `uniq` | Filter duplicate lines |
| `wc` | Word/line/byte count |
| `cut` | Extract fields |
| `tr` | Translate characters |
| `nl` | Number lines |
| `paste` | Merge lines |
| `tac` | Reverse line order |
| `column` | Columnate lists |
| `fold` | Wrap lines |
| `join` | Join files on field |
| `sed` | Stream editor |
| `awk` | Pattern scanning |
| `shuf` | Shuffle lines |
| `split` | Split file into pieces |
| `rev` | Reverse lines |
| `comm` | Compare sorted files |
| `cmp` | Compare bytes |
| `strings` | Print printable strings |

### Search
| Command | Description |
|---------|-------------|
| `rg` | Ripgrep-style search (gitignore, parallel, JSON/NDJSON) |
| `find` | Find files by name/type/size (GNU-compatible flags) |

### System Information
| Command | Description |
|---------|-------------|
| `env` | Print environment |
| `whoami` | Print current user |
| `id` | Print user/group IDs |
| `uname` | Print system info |
| `uptime` | Show system uptime |
| `free` | Display memory info |
| `df` | Show disk usage |
| `du` | Estimate file space |
| `ps` | List processes |
| `kill` | Send signals |
| `time` | Time a command |

### Flow Control
| Command | Description |
|---------|-------------|
| `xargs` | Build arguments |
| `watch` | Execute repeatedly |
| `yes` | Output repeatedly |
| `pipe` | Chain omni commands with variable substitution |
| `pipeline` | Streaming text processing engine (constant memory) |

### Archive & Compression
| Command | Description |
|---------|-------------|
| `tar` | Create/extract tar archives |
| `zip` | Create zip archives |
| `unzip` | Extract zip archives |
| `gzip`/`gunzip`/`zcat` | Gzip compression |
| `bzip2`/`bunzip2`/`bzcat` | Bzip2 compression |
| `xz`/`unxz`/`xzcat` | XZ compression |

### Hash & Encoding
| Command | Description |
|---------|-------------|
| `hash` | Compute file hashes (md5, sha1, sha256, sha512, crc32, crc64) |
| `sha256sum` | SHA256 checksum |
| `sha512sum` | SHA512 checksum |
| `md5sum` | MD5 checksum |
| `crc32sum` | CRC32 checksum (IEEE polynomial) |
| `crc64sum` | CRC64 checksum (ECMA polynomial) |
| `base64` | Base64 encode/decode |
| `base32` | Base32 encode/decode |
| `base58` | Base58 encode/decode |

### Data Processing
| Command | Description |
|---------|-------------|
| `jq` | JSON processor |
| `yq` | YAML processor |
| `dotenv` | Parse .env files |
| `json` | JSON conversions (tostruct, tocsv, fromcsv, toxml, fromxml) |
| `csv` | CSV processing |
| `xml` | XML validate/tojson/fromjson |
| `toml` | TOML validate |
| `yaml` | YAML tostruct/validate |

### Data Formatting
| Command | Description |
|---------|-------------|
| `sql fmt/minify/validate` | SQL formatting |
| `css fmt/minify/validate` | CSS formatting |
| `html fmt/minify/validate` | HTML formatting |

### Case Conversion
| Command | Description |
|---------|-------------|
| `case snake` | Convert to snake_case |
| `case camel` | Convert to camelCase |
| `case kebab` | Convert to kebab-case |
| `case pascal` | Convert to PascalCase |

### Security & Random
| Command | Description |
|---------|-------------|
| `encrypt` | AES-256-GCM encryption |
| `decrypt` | AES-256-GCM decryption |
| `uuid` | Generate UUIDs |
| `random` | Generate random values |

### Video Download
| Command | Description |
|---------|-------------|
| `video download` | Download video from URL |
| `video info` | Show video metadata as JSON |
| `video list-formats` | List available formats |
| `video search` | Search YouTube |
| `video extractors` | List supported sites |

### TUI Pagers
| Command | Description |
|---------|-------------|
| `less` | View file with scrolling |
| `more` | View file page by page |

### Comparison
| Command | Description |
|---------|-------------|
| `diff` | Compare files line by line |

### Cloud & DevOps
| Command | Description |
|---------|-------------|
| `kubectl` / `k` | Full kubectl integration (local source) |
| `terraform` / `tf` | Terraform CLI wrapper |
| `aws` | AWS CLI (EC2, S3, IAM, STS, SSM) |

### Git Hacks
| Command | Alias | Description |
|---------|-------|-------------|
| `git quick-commit` | `gqc` | Stage all + commit |
| `git branch-clean` | `gbc` | Delete merged branches |
| `git undo` | - | Soft reset HEAD~1 |
| `git log-graph` | `git lg` | Pretty log with graph |
| `git status` | `git st` | Short status |

### Kubectl Hacks
| Command | Description |
|---------|-------------|
| `kga` | Get all resources |
| `klf` | Follow pod logs |
| `keb` | Exec into pod |
| `kpf` | Port forward |
| `kge` | Events sorted by time |
| `kcs`/`kns` | Switch context/namespace |

### Database Tools
| Command | Description |
|---------|-------------|
| `sqlite` | SQLite database management (pure Go) |
| `bbolt` | BoltDB key-value store management |

### Code Generation
| Command | Description |
|---------|-------------|
| `generate cobra` | Generate Cobra CLI applications |

### Project Analyzer
| Command | Description |
|---------|-------------|
| `project info` | Full project overview (type, deps, git, docs) |
| `project deps` | Dependency analysis |
| `project docs` | Documentation status check |
| `project git` | Git repository info |
| `project health` | Health score (0-100) with grade |

### Tooling
| Command | Description |
|---------|-------------|
| `lint` | Check Taskfiles for portability |
| `logger` | Configure command logging |

## Database Tools

### SQLite CLI

Pure Go SQLite management (no CGO required):

```bash
# Database info
omni sqlite stats mydb.sqlite
omni sqlite tables mydb.sqlite
omni sqlite schema mydb.sqlite users
omni sqlite columns mydb.sqlite users
omni sqlite indexes mydb.sqlite

# Query execution
omni sqlite query mydb.sqlite "SELECT * FROM users"
omni sqlite query mydb.sqlite "SELECT * FROM users" --json
omni sqlite query mydb.sqlite "SELECT * FROM users" --header

# Maintenance
omni sqlite vacuum mydb.sqlite
omni sqlite check mydb.sqlite
omni sqlite dump mydb.sqlite > backup.sql
omni sqlite import mydb.sqlite backup.sql
```

### BBolt CLI

BoltDB key-value store management:

```bash
# Database info
omni bbolt stats mydb.bolt
omni bbolt buckets mydb.bolt
omni bbolt keys mydb.bolt mybucket

# Key-value operations
omni bbolt get mydb.bolt mybucket mykey
omni bbolt put mydb.bolt mybucket mykey "value"
omni bbolt delete mydb.bolt mybucket mykey

# Maintenance
omni bbolt compact mydb.bolt compacted.bolt
omni bbolt check mydb.bolt
omni bbolt dump mydb.bolt
```

## Code Generation

### Cobra CLI Generator

Generate production-ready Cobra CLI applications:

```bash
# Basic project
omni generate cobra init myapp --module github.com/user/myapp

# With Viper configuration
omni generate cobra init myapp --module github.com/user/myapp --viper

# Full project with CI/CD (goreleaser, workflows, linting)
omni generate cobra init myapp --module github.com/user/myapp --full

# With service pattern (inovacc/config)
omni generate cobra init myapp --module github.com/user/myapp --service

# Add new command to existing project
omni generate cobra add serve
omni generate cobra add config --parent root
```

**Configuration file** (`~/.cobra.yaml`):
```yaml
author: Your Name <email@example.com>
license: MIT
useViper: true
full: true
```

Manage config:
```bash
omni generate cobra config --show
omni generate cobra config --init --author "John Doe" --license MIT
```

## Video Download

Download videos from YouTube and other platforms, pure Go (no FFmpeg required):

```bash
# Download video (best quality)
omni video download "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

# Download worst quality (smallest file)
omni video download -f worst "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

# Custom output filename
omni video download -o "%(title)s.%(ext)s" "https://www.youtube.com/watch?v=..."

# Show video metadata as JSON
omni video info "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

# List available formats
omni video list-formats "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

# Search YouTube
omni video search "golang tutorial"

# Resume a partial download
omni video download -c "https://www.youtube.com/watch?v=..."

# Rate-limited download
omni video download --rate-limit 1M "https://www.youtube.com/watch?v=..."
```

**Supported sites:** YouTube (videos, playlists, search), Generic (direct URLs, `<video>` tags, og:video)

**Protocols:** HTTPS direct download, HLS/M3U8 (with AES-128 decryption)

**Format selectors:** `best`, `worst`, `bestvideo`, `bestaudio`, format ID, `best[height<=720]`

## Project Analyzer

Analyze any codebase directory for project type, languages, dependencies, git info, and health:

```bash
# Full project overview
omni project info

# Analyze a different project
omni project info /path/to/project

# JSON or Markdown output
omni project info --json
omni project health --markdown

# Individual checks
omni project deps                # Dependency analysis
omni project docs                # Documentation status
omni project git -n 5            # Git info (5 recent commits)
omni project health              # Health score (0-100, A-F grade)
```

**Supported ecosystems:** Go, Node.js, Python, Rust, Java, Ruby, PHP, .NET, C/C++, Elixir, Haskell

**Health checks:** README, LICENSE, .gitignore, CI/CD, tests, linter config, git clean, CONTRIBUTING, docs/, CHANGELOG, .editorconfig, build automation

## Library Usage

Core logic is available as importable Go packages under `pkg/`:

```go
import (
    "github.com/inovacc/omni/pkg/idgen"
    "github.com/inovacc/omni/pkg/hashutil"
    "github.com/inovacc/omni/pkg/jsonutil"
    "github.com/inovacc/omni/pkg/encoding"
    "github.com/inovacc/omni/pkg/cryptutil"
    "github.com/inovacc/omni/pkg/sqlfmt"
    "github.com/inovacc/omni/pkg/textutil"
    "github.com/inovacc/omni/pkg/search/grep"
)

// Generate UUID v7
id, _ := idgen.GenerateUUID(idgen.WithUUIDVersion(idgen.V7))

// Hash a file
hash, _ := hashutil.HashFile("file.bin", hashutil.SHA256)

// Query JSON
result, _ := jsonutil.QueryString(`{"name":"omni"}`, ".name")

// Base64 encode
encoded := encoding.Base64Encode([]byte("hello world"))

// AES-256-GCM encryption
ciphertext, _ := cryptutil.Encrypt([]byte("secret"), "password", cryptutil.WithBase64())

// Format SQL
formatted := sqlfmt.Format("select * from users where id=1")

// Sort lines
lines := []string{"banana", "apple", "cherry"}
textutil.SortLines(lines, textutil.WithReverse())

// Search with grep
matches := grep.Search(lines, "app")

// Download video
import "github.com/inovacc/omni/pkg/video"
client, _ := video.New(video.WithFormat("best"))
_ = client.Download(ctx, "https://www.youtube.com/watch?v=...")
```

### Available Packages

| Package | Import | Description |
|---------|--------|-------------|
| `pkg/idgen` | `idgen` | UUID v4/v7, ULID, KSUID, Nanoid, Snowflake |
| `pkg/hashutil` | `hashutil` | MD5, SHA1, SHA256, SHA512, CRC32, CRC64 file/string/reader hashing |
| `pkg/jsonutil` | `jsonutil` | JSON query engine (jq-style filters) |
| `pkg/encoding` | `encoding` | Base64, Base32, Base58 encode/decode |
| `pkg/cryptutil` | `cryptutil` | AES-256-GCM encrypt/decrypt with PBKDF2 |
| `pkg/sqlfmt` | `sqlfmt` | SQL format, minify, validate, tokenize |
| `pkg/cssfmt` | `cssfmt` | CSS format, minify, validate, parse |
| `pkg/htmlfmt` | `htmlfmt` | HTML format, minify, validate |
| `pkg/textutil` | `textutil` | Sort, Uniq, Trim text processing |
| `pkg/textutil/diff` | `diff` | Compute diffs, compare JSON, unified format |
| `pkg/search/grep` | `grep` | Pattern search with regex/fixed/word options |
| `pkg/search/rg` | `rg` | Gitignore parsing, file type matching, binary detection |
| `pkg/pipeline` | `pipeline` | Streaming text processing engine (grep, sort, head, etc.) |
| `pkg/twig` | `twig` | Directory tree scanning, formatting, comparison |
| `pkg/figlet` | `figlet` | FIGlet font parser and ASCII art text renderer |
| `pkg/video` | `video` | Video download engine (YouTube, HLS, generic) |

## Project Structure

```
omni/
├── cmd/                    # Cobra CLI commands (150+ commands)
│   ├── root.go
│   ├── ls.go
│   ├── grep.go
│   └── ...
├── pkg/                    # Reusable Go libraries (importable by external projects)
│   ├── idgen/              # UUID, ULID, KSUID, Nanoid, Snowflake
│   ├── hashutil/           # File/string/reader hashing
│   ├── jsonutil/           # JSON query engine
│   ├── encoding/           # Base64, Base32, Base58
│   ├── cryptutil/          # AES-256-GCM encryption
│   ├── sqlfmt/             # SQL formatting and validation
│   ├── cssfmt/             # CSS formatting and validation
│   ├── htmlfmt/            # HTML formatting and validation
│   ├── textutil/           # Sort, Uniq, Trim
│   │   └── diff/           # Diff computation and JSON comparison
│   ├── search/
│   │   ├── grep/           # Pattern search
│   │   └── rg/             # Gitignore parsing, file type matching
│   ├── pipeline/           # Streaming text processing engine
│   ├── figlet/             # FIGlet font parser and ASCII art
│   ├── twig/               # Tree scanning, formatting, comparison
│   │   ├── scanner/
│   │   ├── formatter/
│   │   ├── comparer/
│   │   └── ...
│   └── video/              # Video download engine
│       ├── extractor/      # Site-specific extractors (YouTube, Generic)
│       ├── downloader/     # HTTP and HLS download engines
│       ├── format/         # Format sorting and selection
│       ├── m3u8/           # HLS manifest parser
│       ├── nethttp/        # HTTP client with cookies/proxy
│       ├── jsinterp/       # JS execution (YouTube signatures)
│       ├── cache/          # Filesystem cache
│       ├── utils/          # HTML, URL, sanitize helpers
│       └── types/          # Shared data structures
├── internal/
│   ├── cli/               # CLI wrappers (I/O, flags, stdin handling)
│   │   ├── ls/
│   │   ├── grep/
│   │   ├── jq/
│   │   └── ...
│   ├── flags/             # Feature flags system
│   └── logger/            # KSUID-based logging
├── include/               # Template reference files
├── docs/                  # Documentation
└── main.go
```

### Test Coverage

- **Overall Coverage:** 30.5% (includes vendored buf packages) | **Omni-owned:** 51.6%
- **Omni pkg/ Average:** ~75% (16 of 31 packages above 80%)
- **Total Test Cases:** 700+
- **Black-box Tests:** 15 Python test suites
- **Golden Master Tests:** 81 snapshot tests across 11 categories

## Platform Support

| Command | Linux | macOS | Windows |
|---------|:-----:|:-----:|:-------:|
| Most commands | ✅ | ✅ | ✅ |
| `chmod` | ✅ | ✅ | ⚠️ Limited |
| `chown` | ✅ | ✅ | ❌ |
| `ps` | ✅ | ✅ | ✅ |
| `df` | ✅ | ✅ | ✅ |
| `free` | ✅ | ✅ | ✅ |
| `uptime` | ✅ | ✅ | ✅ |

## Command Logging

Enable command logging for debugging:

```bash
# Enable logging (Linux/macOS)
eval "$(omni logger --path /tmp/omni-logs)"

# Enable logging (Windows PowerShell)
Invoke-Expression (omni logger --path C:\temp\omni-logs)

# Check status
omni logger --status

# View all logs
omni logger --viewer

# Disable logging
eval "$(omni logger --disable)"
```

Log output is structured JSON with command, args, timestamp, and PID.

### Query Logging

With logging enabled, SQLite queries are automatically logged:

```bash
# Queries logged with timing and row counts
omni sqlite query mydb.sqlite "SELECT * FROM users"

# Include result data in logs (use with caution)
omni sqlite query mydb.sqlite "SELECT * FROM users" --log-data
```

Log entry example:
```json
{
  "msg": "query_result",
  "database": "mydb.sqlite",
  "query": "SELECT * FROM users",
  "status": "success",
  "rows": 10,
  "duration_ms": 25,
  "timestamp": "2026-01-29T12:00:00Z"
}
```

## Use with Taskfile

omni is designed to work with [Taskfile](https://taskfile.dev/):

```yaml
version: '3'

tasks:
  build:
    cmds:
      - omni mkdir -p dist
      - go build -o dist/app .
      - omni sha256sum dist/app > dist/checksums.txt

  clean:
    cmds:
      - omni rm -rf dist

  deploy:
    cmds:
      - omni tar -czvf release.tar.gz dist/
```

## Testing

Run tests with race detection:

```bash
go test -race -cover ./...
```

Or use Taskfile:

```bash
task test               # Unit tests with coverage
task test:blackbox      # Black-box tests (14 Python suites)
task test:golden        # Golden master tests (81 snapshot tests)
task test:golden:update # Regenerate snapshots after intentional changes
task lint               # Linting

# Docker-based testing
task docker:test:linux:blackbox  # Black-box in Linux container
task docker:test:golden          # Golden tests in Linux container
task docker:test:video           # Video comparison (omni vs yt-dlp)
```

## Contributing

Contributions are welcome! See the documentation in `docs/`:

- [ROADMAP.md](docs/ROADMAP.md) - Implementation phases and planned features
- [BACKLOG.md](docs/BACKLOG.md) - Future work and technical debt
- [ARCHITECTURE.md](docs/ARCHITECTURE.md) - System architecture with Mermaid diagrams
- [FEATURES.md](docs/FEATURES.md) - Feature tracker and roadmap
- [MILESTONES.md](docs/MILESTONES.md) - Version milestones and coverage
- [CONTRIBUTORS.md](docs/CONTRIBUTORS.md) - Contributing guide
- [COMMANDS.md](docs/COMMANDS.md) - Full command reference
- [GOLDEN_MASTER_TESTING.md](docs/GOLDEN_MASTER_TESTING.md) - Golden master testing guide
- [REUSABILITY.md](docs/REUSABILITY.md) - Code consolidation opportunities

## License

MIT License - see [LICENSE](LICENSE) for details.
