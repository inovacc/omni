# omni

[![Test](https://github.com/inovacc/omni/actions/workflows/test.yml/badge.svg)](https://github.com/inovacc/omni/actions/workflows/test.yml)

A cross-platform, Go-native replacement for common shell utilities, designed for Taskfile, CI/CD, and enterprise environments.

## Features

- **No exec** - Never spawns external processes
- **Pure Go** - Standard library first, minimal dependencies
- **Cross-platform** - Linux, macOS, Windows
- **Library + CLI** - Use as commands or import as Go packages
- **Safe defaults** - Destructive operations require explicit flags

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
| `nohup` | Ignore hangup signal |

### Archive & Compression
| Command | Description |
|---------|-------------|
| `tar` | Create/extract tar archives |
| `zip` | Create zip archives |
| `unzip` | Extract zip archives |

### Hash & Encoding
| Command | Description |
|---------|-------------|
| `hash` | Compute file hashes |
| `sha256sum` | SHA256 checksum |
| `sha512sum` | SHA512 checksum |
| `md5sum` | MD5 checksum |
| `base64` | Base64 encode/decode |
| `base32` | Base32 encode/decode |
| `base58` | Base58 encode/decode |

### Data Processing
| Command | Description |
|---------|-------------|
| `jq` | JSON processor |
| `yq` | YAML processor |
| `dotenv` | Parse .env files |

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

### Database Tools
| Command | Description |
|---------|-------------|
| `sqlite` | SQLite database management (pure Go) |
| `bbolt` | BoltDB key-value store management |

### Code Generation
| Command | Description |
|---------|-------------|
| `generate cobra` | Generate Cobra CLI applications |

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
| `pkg/hashutil` | `hashutil` | MD5, SHA256, SHA512 file/string/reader hashing |
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
| `pkg/twig` | `twig` | Directory tree scanning, formatting, comparison |
| `pkg/video` | `video` | Video download engine (YouTube, HLS, generic) |

## Project Structure

```
omni/
├── cmd/                    # Cobra CLI commands (100+ commands)
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

- **Overall Coverage:** 97.7% (86/88 packages have tests)
- **CLI Packages:** 100% (79/79 packages)
- **Total Test Cases:** 700+

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
task test
task lint

# Video comparison tests (omni vs yt-dlp in Docker)
task docker:test:video
```

## Contributing

Contributions are welcome! See the documentation in `docs/`:

- [ROADMAP.md](docs/ROADMAP.md) - Implementation phases and planned features
- [BACKLOG.md](docs/BACKLOG.md) - Future work and technical debt
- [REUSABILITY.md](docs/REUSABILITY.md) - Code consolidation opportunities
- [COMMANDS.md](docs/COMMANDS.md) - Full command reference

## License

MIT License - see [LICENSE](LICENSE) for details.
