# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Buf is a CLI tool for working with Protocol Buffers. It provides linting, breaking change detection, code generation, and integration with the Buf Schema Registry (BSR).

## Common Commands

```bash
# Build and install
make install              # Install all binaries (buf, protoc-gen-buf-breaking, protoc-gen-buf-lint)

# Linting
make lint                 # Run all linters (golangci-lint, bufstyle, godoclint, govulncheck)
make shortlint            # Run linters excluding long-running ones

# Testing
make test                 # Run all tests
make testrace             # Run tests with race detector
make shorttest            # Run tests excluding long-running ones (uses -test.short)
go test ./path/to/pkg     # Run tests for a specific package
go test -run TestName ./path/to/pkg  # Run a single test

# Formatting
make gofmtmodtidy         # Format code, run goimports, tidy go.mod

# Code generation
make godata               # Generate static Go data files from protos

# Coverage
make cover                # Generate coverage report (opens HTML)
```

## Architecture

### Dependency Layers (Enforced by Go `internal/` visibility)

The codebase uses a 2-layer architecture with Go's built-in `internal/` visibility:

```
cmd/              →  internal/buf/  →  internal/pkg/
(binaries)           (buf-specific)    (generic utils)
```

- `internal/pkg/` packages cannot depend on anything outside `internal/pkg/`
- `internal/buf/` cannot be used by `internal/pkg/`
- `cmd/` packages are self-contained entry points
- `internal/gen/` contains generated code

### Key Packages

- `internal/buf/bufcli` - CLI version and utilities (version constant in `bufcli.go`)
- `internal/buf/bufctl` - Control flow and command execution
- `internal/buf/bufcheck` - Lint and breaking change rule engines
- `internal/buf/bufconfig` - Configuration parsing (`buf.yaml`)
- `internal/buf/bufimage` - Protocol buffer image representation
- `internal/buf/bufmodule` - Module/dependency management
- `internal/protocompile` - Protobuf compiler (inlined from `github.com/bufbuild/protocompile`)
- `github.com/bufbuild/buf/internal/thread` (`.local-deps/thread/`) - Concurrency utilities (use instead of errgroup)
- `internal/pkg/protoencoding` - Proto marshal/unmarshal (use instead of proto.Marshal)
- `internal/pkg/osext` - OS utilities (use instead of os.Getwd/os.Chdir)
- `internal/pkg/standard/xos/xexec` - Command execution (use instead of exec.Command)

## Forbidden Patterns

These are enforced by golangci-lint (forbidigo):

| Forbidden                            | Use Instead                       |
|--------------------------------------|-----------------------------------|
| `errgroup.*`                         | `github.com/bufbuild/buf/internal/thread.Parallelize` |
| `exec.Cmd`, `exec.Command*`          | `internal/pkg/standard/xos/xexec` |
| `os.Rename`                          | (doesn't work across filesystems) |
| `os.Getwd`, `os.Chdir`               | `internal/pkg/osext`              |
| `fmt.Print*`, `log.*`, `print*`      | Structured logging                |
| `proto.Marshal*`, `proto.Unmarshal*` | `internal/pkg/protoencoding`      |
| `proto.Clone`                        | `proto.CloneOf`                   |

## Code Style

### Nolint Directives

`//nolint` comments are **banned** in this codebase. All lint exceptions must be configured in `.golangci.yml` under the `exclusions.rules` section.

### License Headers

All Go files must have Apache 2.0 license headers (2020-2025). The `license-header` tool enforces this.

### Import Aliases

Required aliases (enforced by linter):

```go
import (
    imagev1 "github.com/bufbuild/buf/internal/gen/proto/go/buf/alpha/image/v1"
    modulev1 "github.com/bufbuild/buf/internal/gen/proto/go/buf/alpha/module/v1"
    registryv1alpha1 "github.com/bufbuild/buf/internal/gen/proto/go/buf/alpha/registry/v1alpha1"
)
```

## Version Management

The CLI version is stored in `internal/buf/bufcli/bufcli.go`:

```go
const Version = "1.64.1-dev"
```

Update version with:

```bash
make updateversion VERSION=x.y.z
```

## Inlined Dependencies

External dependencies are inlined directly into the module tree to eliminate `buf.build` from the build and avoid duplicate proto registration panics.

### Proto `.pb.go` Files — Critical Rules

- **Always use the original `.pb.go` files** from BSR (`go mod download` the module, then copy). Never regenerate them with `protoc-gen-go` unless you can guarantee identical raw descriptor bytes — even minor differences cause `slice bounds out of range [-1:]` panics in `protobuf/internal/filedesc.unmarshalSeed`.
- Never have two separate Go packages that register the same `.proto` file name.

### Current Inlined Dependencies

| Original | Internal Path |
|---|---|
| `buf.build/.../protovalidate/.../buf/validate` | `internal/gen/proto/ext/buf/validate` |
| `buf.build/.../protodescriptor/.../buf/descriptor/v1` | `internal/gen/proto/ext/buf/descriptor/v1` |
| `buf.build/.../pluginrpc/.../pluginrpc/v1` | `internal/gen/proto/ext/pluginrpc/v1` |
| `github.com/bufbuild/protocompile` | `internal/protocompile` |
| `github.com/bufbuild/buf/internal/thread` | `.local-deps/thread` (via `replace`) |

---

## Planned Features

### Offline Mode (Implemented)

The 4 public provider constructors in `internal/buf/bufcli/cache.go` (`NewModuleDataProvider`, `NewCommitProvider`, `NewPluginDataProvider`, `NewPolicyDataProvider`) are hardcoded to offline-only mode. They use cache-backed stores with offline delegates (defined in `internal/buf/bufcli/offline_delegate.go`) that return an error on cache miss instead of making network calls. No `BUF_OFFLINE` env var is needed.

The private `new*` helper functions in `cache.go` and the controller (`controller.go`) still use network-backed delegates for operations like `buf dep update` that require network access.

**Key files:**

- `internal/buf/bufcli/cache.go` - Public constructors use offline delegates
- `internal/buf/bufcli/offline_delegate.go` - Offline delegate types that error on cache miss
- `internal/buf/bufcli/controller.go` - Still uses network-backed private constructors
