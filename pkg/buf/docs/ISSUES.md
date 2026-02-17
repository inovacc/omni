# Known Issues

## ~~rawDesc go_package mismatch after restructuring~~ [RESOLVED]

Proto files have been regenerated with correct `internal/` `go_package` paths.

## ~~Offline mode cache miss errors are generic~~ [RESOLVED]

Offline delegate errors already include actionable guidance (e.g., "run 'buf dep update' with network access first").

## Test plugin binaries must be pre-installed

- **Severity:** Low (test-only)
- **Details:** Several integration tests require test plugin binaries (`buf-plugin-suffix`, `buf-plugin-panic`, etc.) to be in PATH. These are built from `internal/buf/bufcheck/internal/cmd/` but must be installed separately before running tests. `make installtest` handles this.
- **Workaround:** Run `make installtest` before `go test ./...`.

## protoc include path required for some tests

- **Severity:** Low (test-only, platform-specific)
- **Details:** `protoc-gen-buf-lint` tests require protoc's well-known type includes (e.g., `google/protobuf/any.proto`) at specific paths. On Windows, the path may differ from the expected location.
- **Workaround:** Install protoc and ensure its include directory is accessible.
