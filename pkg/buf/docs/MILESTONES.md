# Milestones

## v1.64.0 (Latest Release)

- Policy registry commands (`buf registry policy`)
- LSP map support improvements
- LSP server version logging

## v1.63.0

- Plugin registry v1beta1 support
- Additional breaking change rules

## v1.62.x

- Edition 2023/2024 support improvements
- Protovalidate integration for lint rules

## v1.60.0 - v1.61.0

- Policy management (`buf policy push/update/prune`)
- Plugin management (`buf plugin push/update/prune`)
- Custom check plugin framework

## v1.50.0 - v1.59.0

- LSP server implementation
- WASM plugin execution
- Studio agent improvements
- Module v1 API migration

## v1.0.0 - v1.49.0

- Core CLI: lint, breaking, generate, format, export, convert
- BSR integration: push, pull, dependency management
- Workspace support
- protoc compatibility mode
- `buf curl` for RPC invocation

## Next (v1.65.0+)

- Architecture: 2-layer `internal/` structure (complete)
- Offline-only provider constructors (complete)
- Dependency inlining for self-contained builds (complete)
- Package extraction: `thread` (planned — see BACKLOG.md), `dag` (deferred to P3)
- Tech debt: controller.go functionOptions refactor (complete), flag name plumbing (complete)
