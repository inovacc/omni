package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	"github.com/inovacc/omni/internal/cli/scaffolding"
	mcptpl "github.com/inovacc/omni/internal/cli/scaffolding/mcp/templates"
)

// MCPOptions configures MCP server generation
type MCPOptions struct {
	Module    string // Go module path (auto-detected from go.mod if empty)
	Transport string // Default transport: stdio, sse, http-stream (default: stdio)
	Addr      string // Default listen address for SSE/HTTP (default: ":8080")
}

// MCPResult represents the result of MCP generation
type MCPResult struct {
	Status       string   `json:"status"`
	Name         string   `json:"name"`
	Transport    string   `json:"transport"`
	FilesCreated []string `json:"files_created"`
}

// RunMCPInit generates MCP server code for an existing Cobra project
func RunMCPInit(w io.Writer, fs afero.Fs, name string, opts MCPOptions, genOpts scaffolding.Options) error {
	if name == "" {
		return fmt.Errorf("scaffold: MCP server name is required")
	}

	// Set defaults
	if opts.Transport == "" {
		opts.Transport = "stdio"
	}

	if opts.Addr == "" {
		opts.Addr = ":8080"
	}

	// Validate transport
	switch opts.Transport {
	case "stdio", "sse", "http-stream":
	default:
		return fmt.Errorf("scaffold: invalid transport %q (must be stdio, sse, or http-stream)", opts.Transport)
	}

	// Detect module from go.mod if not provided
	if opts.Module == "" {
		mod, err := detectModule(fs)
		if err != nil {
			return fmt.Errorf("scaffold: --module is required (could not detect from go.mod: %w)", err)
		}
		opts.Module = mod
	}

	// Detect app name from module
	appName := filepath.Base(opts.Module)

	// Prepare template data
	data := mcptpl.TemplateData{
		Module:    opts.Module,
		AppName:   appName,
		Name:      name,
		Transport: opts.Transport,
		Addr:      opts.Addr,
	}

	var filesCreated []string

	// Create internal/mcp/ directory
	mcpDir := filepath.Join("internal", "mcp")
	if err := fs.MkdirAll(mcpDir, 0755); err != nil {
		return fmt.Errorf("scaffold: failed to create directory %s: %w", mcpDir, err)
	}

	// Generate files
	files := []struct {
		path string
		tmpl string
	}{
		{filepath.Join(mcpDir, "server.go"), mcptpl.ServerTemplate},
		{filepath.Join(mcpDir, "tools.go"), mcptpl.ToolsTemplate},
		{filepath.Join(mcpDir, "resources.go"), mcptpl.ResourcesTemplate},
		{filepath.Join(mcpDir, "debug.go"), mcptpl.DebugTemplate},
	}

	for _, f := range files {
		if err := scaffolding.WriteTemplate(fs, f.path, f.tmpl, data); err != nil {
			return fmt.Errorf("scaffold: failed to create %s: %w", f.path, err)
		}
		filesCreated = append(filesCreated, f.path)
	}

	// Generate cmd file
	cmdDir := filepath.Join("cmd", appName)
	if err := fs.MkdirAll(cmdDir, 0755); err != nil {
		return fmt.Errorf("scaffold: failed to create directory %s: %w", cmdDir, err)
	}

	cmdPath := filepath.Join(cmdDir, "cmd_mcp.go")
	if err := scaffolding.WriteTemplate(fs, cmdPath, mcptpl.CmdTemplate, data); err != nil {
		return fmt.Errorf("scaffold: failed to create %s: %w", cmdPath, err)
	}
	filesCreated = append(filesCreated, cmdPath)

	if genOpts.JSON {
		result := MCPResult{
			Status:       "created",
			Name:         name,
			Transport:    opts.Transport,
			FilesCreated: filesCreated,
		}
		return json.NewEncoder(w).Encode(result)
	}

	_, _ = fmt.Fprintf(w, "Created MCP server: %s\n", name)
	_, _ = fmt.Fprintf(w, "Transport: %s\n", opts.Transport)

	if opts.Transport != "stdio" {
		_, _ = fmt.Fprintf(w, "Address: %s\n", opts.Addr)
	}

	_, _ = fmt.Fprintln(w, "\nFiles created:")
	for _, f := range filesCreated {
		_, _ = fmt.Fprintf(w, "  - %s\n", f)
	}

	_, _ = fmt.Fprintln(w, "\nNext steps:")
	_, _ = fmt.Fprintln(w, "  1. Add github.com/modelcontextprotocol/go-sdk to go.mod:")
	_, _ = fmt.Fprintln(w, "     go get github.com/modelcontextprotocol/go-sdk")
	_, _ = fmt.Fprintln(w, "  2. Add tools and resources in internal/mcp/")
	_, _ = fmt.Fprintf(w, "  3. Run: go run ./cmd/%s mcp serve\n", appName)

	return nil
}

// detectModule reads the module path from go.mod in the current directory
func detectModule(fs afero.Fs) (string, error) {
	data, err := afero.ReadFile(fs, "go.mod")
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}

	return "", fmt.Errorf("module directive not found in go.mod")
}
