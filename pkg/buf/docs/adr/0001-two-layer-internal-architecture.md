# ADR-0001: Two-Layer Internal Architecture

## Status

Accepted

## Date

2026-02-15

## Context

The buf CLI originally used a 4-layer dependency structure under `private/`:

```
cmd/ → private/buf/ → private/bufpkg/ → private/pkg/
```

This was enforced by the `bandeps` tool since Go's `private/` directory convention has no special compiler support. The layers were:
- `private/pkg/` — generic utilities (47 packages)
- `private/bufpkg/` — buf-specific utilities (18 packages)
- `private/buf/` — buf CLI logic (18 packages)
- `cmd/` — binary entry points (3 binaries)

Problems with this approach:
1. **No compiler enforcement**: `private/` has no meaning to the Go toolchain. The `bandeps` linter was required to catch violations, but only ran during CI.
2. **Unnecessary granularity**: The distinction between `private/buf/` and `private/bufpkg/` added complexity without clear benefit. Many packages in `bufpkg/` were tightly coupled to `buf/` packages.
3. **Confusing naming**: `private/` is not a Go convention; `internal/` is the standard.

## Decision

Restructure to a 2-layer architecture using Go's built-in `internal/` visibility:

```
cmd/ → internal/buf/ → internal/pkg/
```

- Rename `private/` to `internal/` for compiler-enforced visibility
- Merge `bufpkg/` packages into `buf/` (no naming conflicts existed)
- Simplify `bandeps` rules from 5 to 4 (the `bufpkg` intermediate layer rule is removed)

## Consequences

### Positive
- Go compiler enforces visibility: external consumers cannot import `internal/` packages
- Simpler mental model: only 2 layers instead of 4
- `bandeps` becomes optional (compiler does the heavy lifting)
- 18 packages moved without any naming conflicts

### Negative
- Large diff (~1,300 Go files modified for import path changes)
- Generated `.pb.go` files require careful handling: rawDesc binary descriptors contain embedded `go_package` strings that must not be modified by import rewriting
- Historical git blame is harder to trace across the rename

### Risks Mitigated
- Binary protobuf descriptors in `.pb.go` rawDesc constants were preserved intact by restoring original files and only updating Go import statements
- `go build ./...` and `go vet ./...` verified clean after each phase
