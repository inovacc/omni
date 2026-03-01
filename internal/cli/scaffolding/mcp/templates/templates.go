package templates

// TemplateData contains all data needed for MCP server template rendering
type TemplateData struct {
	Module    string // Go module path (e.g., github.com/user/myapp)
	AppName   string // Application name
	Name      string // MCP server name
	Transport string // Default transport: stdio, sse, http-stream
	Addr      string // Default address for SSE/HTTP (e.g., ":8080")
}

// ServerTemplate generates internal/mcp/server.go
const ServerTemplate = `package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewServer(name, version string, logger *slog.Logger) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: name, Version: version}, &mcp.ServerOptions{
		Logger: logger,
	})
	RegisterTools(s)
	RegisterResources(s)
	return s
}

func Run(ctx context.Context, s *mcp.Server, transport, addr string, logger *slog.Logger) error {
	switch transport {
	case "stdio":
		return s.Run(ctx, &mcp.StdioTransport{})
	case "sse":
		handler := mcp.NewSSEHandler(func(r *http.Request) *mcp.Server { return s }, nil)
		logger.Info("starting SSE server", "addr", addr)
		return http.ListenAndServe(addr, handler)
	case "http-stream":
		handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server { return s }, nil)
		logger.Info("starting HTTP streamable server", "addr", addr)
		return http.ListenAndServe(addr, handler)
	default:
		return fmt.Errorf("unknown transport: %s", transport)
	}
}
`

// ToolsTemplate generates internal/mcp/tools.go
const ToolsTemplate = `package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GreetInput struct {
	Name string ` + "`" + `json:"name" jsonschema:"description=Name of the person to greet"` + "`" + `
}

func RegisterTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "greet",
		Description: "Greet someone by name",
	}, handleGreet)
}

func handleGreet(_ context.Context, _ *mcp.CallToolRequest, input GreetInput) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Hello, %s!", input.Name)},
		},
	}, nil, nil
}
`

// ResourcesTemplate generates internal/mcp/resources.go
const ResourcesTemplate = `package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterResources(s *mcp.Server) {
	s.AddResource(&mcp.Resource{
		URI:         "info://server",
		Name:        "Server Info",
		Description: "Information about this MCP server",
		MIMEType:    "text/plain",
	}, handleInfo)
}

func handleInfo(_ context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: "info://server", MIMEType: "text/plain", Text: "{{.Name}} MCP Server v0.1.0"},
		},
	}, nil
}
`

// DebugTemplate generates internal/mcp/debug.go
const DebugTemplate = `package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
)

// NewLogger creates a logger with the specified level, writing to stderr and optionally a log file.
func NewLogger(level string, logFile string) (*slog.Logger, func(), error) {
	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "trace":
		slogLevel = slog.LevelDebug - 4 // custom trace level
	default:
		slogLevel = slog.LevelInfo
	}

	var writers []io.Writer
	writers = append(writers, os.Stderr)

	var cleanup func()

	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, nil, fmt.Errorf("open log file: %w", err)
		}
		writers = append(writers, f)
		cleanup = func() { _ = f.Close() }
	} else {
		cleanup = func() {}
	}

	w := io.MultiWriter(writers...)
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slogLevel})
	return slog.New(handler), cleanup, nil
}

// PrettyJSON formats JSON for debug logging.
func PrettyJSON(data any) string {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", data)
	}
	return string(b)
}
`

// CmdTemplate generates cmd/<appname>/cmd_mcp.go
const CmdTemplate = `package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	mcpserver "{{.Module}}/internal/mcp"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP server commands",
}

var mcpServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		transport, _ := cmd.Flags().GetString("transport")
		addr, _ := cmd.Flags().GetString("addr")
		logLevel, _ := cmd.Flags().GetString("log-level")
		logFile, _ := cmd.Flags().GetString("log-file")

		logger, cleanup, err := mcpserver.NewLogger(logLevel, logFile)
		if err != nil {
			return err
		}
		defer cleanup()

		logger.Info("starting MCP server",
			"name", "{{.Name}}",
			"transport", transport,
		)

		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		s := mcpserver.NewServer("{{.Name}}", "v0.1.0", logger)
		return mcpserver.Run(ctx, s, transport, addr, logger)
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.AddCommand(mcpServeCmd)

	mcpServeCmd.Flags().String("transport", "{{.Transport}}", "transport type: stdio, sse, http-stream")
	mcpServeCmd.Flags().String("addr", "{{.Addr}}", "listen address (for sse/http-stream)")
	mcpServeCmd.Flags().String("log-level", "info", "log level: info, debug, trace")
	mcpServeCmd.Flags().String("log-file", "", "log file path (in addition to stderr)")
}
`
