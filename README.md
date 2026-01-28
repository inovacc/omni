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

Build from source:
```bash
git clone https://github.com/inovacc/omni.git
cd omni
task build
```

Or manually:
```bash
go build -ldflags="-s -w" -o omni .
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

### Tooling
| Command | Description |
|---------|-------------|
| `lint` | Check Taskfiles for portability |

## Library Usage

All commands are available as importable Go packages:

```go
import "github.com/inovacc/omni/pkg/cli"

// List files
cli.RunLs(os.Stdout, []string{"."}, cli.LsOptions{All: true, Long: true})

// Hash a file
cli.RunSHA256Sum(os.Stdout, []string{"file.bin"}, cli.HashOptions{})

// Generate UUID
uuid := cli.NewUUID()

// Generate random password
password := cli.RandomPassword(20)

// Load .env file
cli.LoadDotenv(".env")

// Parse JSON
cli.RunJq(os.Stdout, []string{".name", "data.json"}, cli.JqOptions{Raw: true})
```

## Project Structure

```
omni/
├── cmd/                    # Cobra CLI commands
│   ├── root.go
│   ├── ls.go
│   ├── grep.go
│   └── ...
├── pkg/cli/               # Library implementations
│   ├── ls.go
│   ├── grep.go
│   ├── jq.go
│   └── ...
├── tests/                 # Integration tests
├── docs/                  # Documentation
│   ├── ROADMAP.md
│   ├── COMMANDS.md
│   └── BACKLOG.md
└── main.go
```

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

Contributions are welcome! Please see [ROADMAP.md](docs/ROADMAP.md) for planned features.

## License

MIT License - see [LICENSE](LICENSE) for details.
