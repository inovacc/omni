# ADR-0002: Internalize buf Protobuf Tooling

**Status:** Proposed
**Date:** 2026-02-18
**Decision:** Use `github.com/bufbuild/protocompile` as a direct dependency instead of internalizing the full `github.com/bufbuild/buf` repository.

---

## Context

omni provides `buf` subcommands (lint, format, build, breaking, generate, mod init, mod update, ls-files) as pure-Go replacements for the buf CLI. The original implementation vendored `pkg/buf/pkg/bufapi` but this was removed during the module flatten (commit `a48b52a`). Currently:

- **`buf lint`** — Works via a hand-written proto3 lexer/parser (`proto.go`, ~1100 LOC) with 28 rules; 6 are stubs
- **`buf format`** — Only collapses consecutive blank lines; no real formatting
- **`buf build`** — Complete stub ("compilation engine not available")
- **`buf breaking`** — Complete stub ("not available")
- **`buf generate`** — Shells out to `protoc` (violates no-exec principle for remote plugins)
- **`buf mod init`** / **`buf ls-files`** — Fully functional (no compiler needed)
- **`buf mod update`** — Stub (needs BSR network access)

## Analysis

### Full buf Repository (github.com/bufbuild/buf)

| Field | Value |
|-------|-------|
| Module | `github.com/bufbuild/buf` |
| Upstream commit | `1de5bd86c535754bb9cb147361a457ed140de59d` |
| Total Go files | ~1,008 |
| Total LOC | ~200,000–350,000 |
| Packages | ~150+ |
| Direct deps | 47 |
| License | Apache 2.0 |

The full buf repo is massive. All library code lives under `private/` (not importable). Key packages:

- `private/buf/bufformat/` — AST-based proto formatter (requires protocompile)
- `private/bufpkg/bufcheck/` — Lint + breaking client interface with pluginrpc-based rule dispatch
- `private/bufpkg/bufimage/` — Image abstraction (compiled FileDescriptorSet)
- `private/bufpkg/bufconfig/` — buf.yaml/buf.gen.yaml parsing
- `private/bufpkg/bufprotosource/` — Rich source-aware descriptor interfaces

Heavy dependencies include: ConnectRPC, Docker SDK, CEL (google/cel-go), wazero (WASM runtime), LSP, gRPC, OCI containers, OAuth2. These are completely unnecessary for omni's needs.

### protocompile (github.com/bufbuild/protocompile)

| Field | Value |
|-------|-------|
| Module | `github.com/bufbuild/protocompile` |
| Latest version | v0.14.1 |
| Go files | ~100 |
| Purpose | Pure-Go protobuf parser, linker, compiler |
| Direct deps | ~12 (4 net-new for omni) |
| License | Apache 2.0 |

protocompile provides the entire proto compilation pipeline:

| Stage | Package | Purpose |
|-------|---------|---------|
| Parse | `parser/` | Source → AST (`ast.FileNode`) |
| Descriptor | `parser/` | AST → `descriptorpb.FileDescriptorProto` |
| Link | `linker/` | Cross-file type resolution, validation |
| Source info | `sourceinfo/` | Source location/comment mapping |
| Walk | `walk/` | Descriptor hierarchy traversal |
| Well-known types | `wellknownimports/` | Embedded `google/protobuf/*.proto` |

**Net new production deps for omni (only 4):**
- `buf.build/gen/go/bufbuild/protodescriptor/protocolbuffers/go` — Generated descriptor types
- `github.com/tidwall/btree` — Symbol table internals
- `github.com/petermattis/goid` — Goroutine ID for parallel compiler
- `github.com/bmatcuk/doublestar/v4` — Glob matching for imports

Already in omni: `google.golang.org/protobuf`, `golang.org/x/sync`, `golang.org/x/exp`, `github.com/rivo/uniseg`

## Decision

**Strategy: Keep external dependency (`protocompile`) + selective code copy from `buf` for formatter/lint rules.**

Rationale:
1. **The full buf repo is too large** (1000+ files, 47 deps) and lives entirely under `private/` — internalizing it would be a fork, not a vendoring
2. **protocompile is the core engine** that powers all of buf's compilation — it's small, focused, and designed as a library
3. **Only 4 net-new deps** vs. ~40+ from the full buf repo
4. **Replaces the hand-written parser** (`proto.go`, ~1100 LOC) with a battle-tested compiler used by the real buf
5. **Enables real implementations** of build, breaking, format, and improved lint

What to copy from the full buf repo (into `pkg/buf/`):
- Formatter logic from `private/buf/bufformat/` — adapted to work with protocompile AST directly
- Breaking change detection rule logic from `private/bufpkg/bufcheck/bufcheckserver/internal/bufcheckserverhandle/` — adapted to use `bufprotosource`-style patterns on protocompile descriptors
- Lint rule patterns from the same package — to upgrade the existing 28 rules + implement the 6 COMMENT stubs

What NOT to copy:
- BSR/registry client code (`bufregistryapi/`, `bufconnect/`, `bufmoduleapi/`)
- Plugin execution (`bufprotopluginexec/`, wasm support)
- CLI framework (`bufcli/`, `bufcobra/`)
- Controller/orchestration (`bufctl/`)
- Generated proto stubs (`gen/proto/`)
- LSP server (`buflsp/`)
- Docker/OCI support (`bufremoteplugin/`)

## Implementation Plan

### Phase 1: Add protocompile dependency and implement `buf build`

1. Add `github.com/bufbuild/protocompile` to `go.mod`
2. Create `pkg/buf/compile/` — thin wrapper around protocompile's `Compiler`
3. Implement real `RunBuild`:
   - Compile proto files via protocompile
   - Serialize `FileDescriptorSet` to JSON or binary
   - Write to `-o` output path or stdout
4. Update tests

### Phase 2: Implement `buf breaking`

1. Create `pkg/buf/breaking/` — breaking change detection rules
2. Copy/adapt rule logic from buf's `bufcheckserverhandle` breaking handlers
3. Compare current vs. against `FileDescriptorSet` descriptors
4. Implement the 8 documented rules: FILE_NO_DELETE, PACKAGE_NO_DELETE, MESSAGE_NO_DELETE, FIELD_NO_DELETE, FIELD_SAME_TYPE, ENUM_NO_DELETE, SERVICE_NO_DELETE, RPC_NO_DELETE
5. Update `RunBreaking` and tests

### Phase 3: Upgrade `buf format`

1. Create `pkg/buf/format/` — AST-based proto formatter
2. Port formatter logic from buf's `bufformat` package
3. Use protocompile's `parser.Parse()` → `ast.FileNode` for proper AST round-tripping
4. Replace `cleanupBlankLines` with real formatting (indentation, spacing, import sorting)
5. Update `RunFormat` and tests

### Phase 4: Upgrade `buf lint`

1. Migrate lint rules from the hand-written `ProtoFile` AST to protocompile's `protoreflect.FileDescriptor`
2. Implement the 6 COMMENT rule stubs using protocompile's source info (line/column/comments)
3. Implement remaining stubs (DIRECTORY_SAME_PACKAGE, PACKAGE_SAME_DIRECTORY, ONEOF_LOWER_SNAKE_CASE, RPC_REQUEST_RESPONSE_UNIQUE)
4. Wire `--error-format` (JSON, github-actions) into `RunLint`
5. Wire `buf.yaml` config loading into `RunLint` (currently ignored)
6. Retire `proto.go` (hand-written parser)

### Phase 5: Cleanup

1. Delete `internal/cli/buf/proto.go` (replaced by protocompile)
2. Run `go mod tidy`
3. Update documentation
4. Remove `.tmp/buf`

## Risk Assessment

- **License compatibility**: Apache 2.0 (buf/protocompile) is compatible with BSD 3-Clause (omni)
- **Maintenance burden**: protocompile is actively maintained by Buf Technologies; as an external dep, we get updates for free
- **Complexity**: No CGO, no assembly, no build tags — pure Go
- **Breaking changes risk**: protocompile follows semver; pinning to v0.14.x is safe

## Consequences

- omni gains a real protobuf compiler engine
- `buf build` produces real `FileDescriptorSet` output
- `buf breaking` detects actual breaking changes
- `buf format` properly reformats proto files
- `buf lint` operates on compiled descriptors instead of text scanning
- ~4 new small dependencies added to go.mod
- ~1100 lines of hand-written parser can be deleted
