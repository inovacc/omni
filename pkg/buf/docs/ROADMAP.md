# Roadmap

## Phase 1: Core CLI [COMPLETE]

- [x] Protobuf linting with configurable rules
- [x] Breaking change detection
- [x] Code generation orchestration (`buf generate`)
- [x] Protobuf compilation and image building
- [x] Multi-module workspace support
- [x] Configuration system (`buf.yaml`, `buf.gen.yaml`, `buf.lock`)
- [x] Input fetching (local dirs, git repos, archives, BSR modules)
- [x] Content-addressable storage for modules
- [x] `protoc` compatibility mode (`buf alpha protoc`)
- [x] Format command (`buf format`)
- [x] Export command (`buf export`)
- [x] Convert command (`buf convert`)

## Phase 2: Registry Integration [COMPLETE]

- [x] Buf Schema Registry (BSR) authentication
- [x] Module push/pull from BSR
- [x] Dependency management (`buf dep update`, `buf dep prune`)
- [x] Module labels, commits, and versioning
- [x] Organization and user management
- [x] Plugin registry integration
- [x] Remote plugin execution
- [x] SDK info commands
- [x] Webhook management

## Phase 3: Advanced Features [COMPLETE]

- [x] Language Server Protocol (LSP) support
- [x] Custom check plugins (lint + breaking)
- [x] Policy management (`buf policy push/update/prune`)
- [x] Plugin management (`buf plugin push/update/prune`)
- [x] Studio agent for BSR studio
- [x] `buf curl` for RPC invocation
- [x] WASM plugin support
- [x] Docker-based remote plugin execution

## Phase 4: Architecture Modernization [COMPLETE]

- [x] Inline buf.build dependencies into module tree
- [x] Offline-only provider constructors (cache-backed, no network by default)
- [x] Restructure to 2-layer `internal/` architecture
- [x] Remove bandeps in favor of Go's built-in `internal/` visibility
- [x] Evaluate extracting reusable packages to standalone modules
  - Best candidates: `thread` (0 internal deps, 9 dependents), `dag` (2 deps, 4 dependents)
  - Good candidates: `protoencoding` (2 deps, 17 dependents)
  - Deferred: `storage` (9 internal deps, 36 dependents — too coupled)
  - `protovalidate` is inlined from upstream; extraction = revert to external dep
  - Actual extraction tracked in BACKLOG.md

## Overall Progress: ~98%

The project is mature and production-ready. Architecture modernization is complete; remaining work is incremental tech debt and potential package extractions.
