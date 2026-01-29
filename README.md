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

## Library Usage

All commands are available as importable Go packages:

```go
import (
    "github.com/inovacc/omni/internal/cli/ls"
    "github.com/inovacc/omni/internal/cli/hash"
    "github.com/inovacc/omni/internal/cli/uuid"
    "github.com/inovacc/omni/internal/cli/random"
    "github.com/inovacc/omni/internal/cli/dotenv"
    "github.com/inovacc/omni/internal/cli/jq"
)

// List files
ls.RunLs(os.Stdout, []string{"."}, &ls.Options{All: true, Long: true})

// Hash a file
hash.RunSHA256Sum(os.Stdout, []string{"file.bin"}, &hash.Options{})

// Generate UUID
u := uuid.NewUUID()

// Generate random password
password := random.Password(20)

// Load .env file
dotenv.Load(".env")

// Parse JSON
jq.RunJq(os.Stdout, []string{".name", "data.json"}, &jq.Options{Raw: true})
```

## Project Structure

```
omni/
├── cmd/                    # Cobra CLI commands (100+ commands)
│   ├── root.go
│   ├── ls.go
│   ├── grep.go
│   ├── sqlite.go
│   ├── bbolt.go
│   ├── generate.go
│   └── ...
├── internal/
│   ├── cli/               # Library implementations (80+ packages)
│   │   ├── ls/
│   │   ├── grep/
│   │   ├── jq/
│   │   ├── sqlite/        # SQLite operations
│   │   ├── bbolt/         # BoltDB operations
│   │   ├── generate/      # Code generation
│   │   │   └── templates/ # Cobra templates
│   │   └── ...
│   ├── flags/             # Feature flags system
│   ├── logger/            # KSUID-based logging with query support
│   └── twig/              # Tree visualization module
│       ├── scanner/       # Directory scanning
│       ├── formatter/     # Output formatting
│       ├── builder/       # Structure creation
│       ├── parser/        # Tree parsing
│       └── models/        # Data models
├── include/               # Template reference files
│   └── cobra/             # Cobra app templates
├── docs/                  # Documentation
│   ├── ROADMAP.md
│   ├── COMMANDS.md
│   └── BACKLOG.md
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
```

## Contributing

Contributions are welcome! See the documentation in `docs/`:

- [ROADMAP.md](docs/ROADMAP.md) - Implementation phases and planned features
- [BACKLOG.md](docs/BACKLOG.md) - Future work and technical debt
- [REUSABILITY.md](docs/REUSABILITY.md) - Code consolidation opportunities
- [COMMANDS.md](docs/COMMANDS.md) - Full command reference

## License

MIT License - see [LICENSE](LICENSE) for details.
