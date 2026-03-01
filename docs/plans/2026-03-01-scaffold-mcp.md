# Scaffold MCP Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `omni scaffold mcp <name>` to generate MCP server code with stdio/SSE/HTTP-stream transports and debug logging.

**Architecture:** New scaffolder at `internal/cli/scaffolding/mcp/` following handler/repository pattern. Templates generate `internal/mcp/` (server, tools, resources, debug) + `cmd/<app>/cmd_mcp.go`.

**Tech Stack:** Go templates, afero filesystem, `github.com/modelcontextprotocol/go-sdk/mcp`

---

### Task 1: Create MCP templates

**Files:**
- Create: `internal/cli/scaffolding/mcp/templates/templates.go`

**Step 1:** Create templates file with TemplateData struct and all 5 templates (server.go, tools.go, resources.go, debug.go, cmd_mcp.go).

**Step 2:** Verify it compiles: `go build ./internal/cli/scaffolding/mcp/...`

**Step 3:** Commit: `feat(scaffold): add MCP server templates`

---

### Task 2: Create MCP scaffolder

**Files:**
- Create: `internal/cli/scaffolding/mcp/mcp.go`

**Step 1:** Write `RunMCPInit(w, fs, name, opts, genOpts)` following handler.go pattern exactly.

**Step 2:** Verify: `go build ./internal/cli/scaffolding/mcp/...`

**Step 3:** Commit: `feat(scaffold): add MCP scaffolder`

---

### Task 3: Create test for MCP scaffolder

**Files:**
- Create: `internal/cli/scaffolding/mcp/mcp_test.go`

**Step 1:** Write table-driven tests using afero.MemMapFs: default opts, each transport, JSON output, missing name error.

**Step 2:** Run: `go test -v ./internal/cli/scaffolding/mcp/...`

**Step 3:** Commit: `test(scaffold): add MCP scaffolder tests`

---

### Task 4: Wire into cmd/scaffold.go

**Files:**
- Modify: `cmd/scaffold.go`

**Step 1:** Add `scaffoldMCPCmd` with flags: `--transport`, `--addr`, `--log-level`, `--log-file`, `--module`, detect module from go.mod.

**Step 2:** Register in init(): `scaffoldCmd.AddCommand(scaffoldMCPCmd)`

**Step 3:** Verify: `go build . && go run . scaffold mcp --help`

**Step 4:** Commit: `feat(scaffold): wire MCP command into scaffold`

---

### Task 5: Update docs

**Files:**
- Modify: `CLAUDE.md` — add mcp to scaffold section

**Step 1:** Update scaffold command list and examples.

**Step 2:** Commit: `docs: add scaffold mcp to CLAUDE.md`
