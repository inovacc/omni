# Design: `omni scaffold mcp`

**Date:** 2026-03-01
**Status:** Approved

## Summary

Add MCP server scaffolding to omni. Generates `internal/mcp/` with server setup, example tool/resource, debug logging, and a Cobra command in `cmd/`.

## CLI

```bash
omni scaffold mcp <name> [flags]

Flags:
  --transport stdio|sse|http-stream   (default: stdio)
  --addr string                        (default: ":8080", for sse/http-stream)
  --log-level info|debug|trace         (default: info)
  --log-file string                    (write log to file)
  --module string                      (auto-detected from go.mod)
  --json                               (JSON output)
```

## Generated Files

```
internal/mcp/
├── server.go       # NewServer(), RegisterTools(), RegisterResources(), Run()
├── tools.go        # GreetInput + handler (example tool)
├── resources.go    # Info resource handler (example resource)
└── debug.go        # DebugLogger: logs JSON-RPC at debug/trace levels

cmd/<appname>/
└── cmd_mcp.go      # Cobra "mcp serve" command with all flags
```

## Components

### server.go
- `NewServer(name, version, logger)` — creates MCP server, registers tools/resources
- `Run(ctx, server, transport, addr)` — starts server with selected transport

### tools.go
- `GreetInput` struct with `json` + `jsonschema` tags
- Handler returns `TextContent`

### resources.go
- Static `info` resource returning server metadata

### debug.go
- **info**: startup/shutdown only
- **debug**: pretty-printed JSON-RPC requests/responses
- **trace**: raw bytes hex dump + JSON
- `--log-file`: `io.MultiWriter(stderr, file)`

### cmd_mcp.go
- `mcp serve` subcommand
- Builds `slog.Logger` from `--log-level` and `--log-file`
- Calls `mcp.NewServer()` + `mcp.Run()`

## Scaffolder Structure

```
internal/cli/scaffolding/mcp/
├── mcp.go              # RunMCPInit(fs, w, opts)
└── templates/
    └── templates.go    # All templates
```

## Dependencies

- `github.com/modelcontextprotocol/go-sdk/mcp` (added to generated go.mod)

## Wiring

- New `scaffoldMCPCmd` in `cmd/scaffold.go`
- Follows exact pattern of handler/repository scaffolders
