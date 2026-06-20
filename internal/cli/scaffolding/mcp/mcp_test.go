package mcp

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/internal/cli/scaffolding"
)

func TestRunMCPInit(t *testing.T) {
	tests := []struct {
		name      string
		mcpName   string
		opts      MCPOptions
		genOpts   scaffolding.Options
		gomod     string
		wantErr   bool
		errMsg    string
		wantFiles []string
	}{
		{
			name:    "default options with module flag",
			mcpName: "myserver",
			opts:    MCPOptions{Module: "github.com/user/myapp"},
			gomod:   "module github.com/user/myapp\n\ngo 1.22\n",
			wantFiles: []string{
				"internal/mcp/server.go",
				"internal/mcp/tools.go",
				"internal/mcp/resources.go",
				"internal/mcp/debug.go",
				"cmd/myapp/cmd_mcp.go",
			},
		},
		{
			name:    "auto-detect module from go.mod",
			mcpName: "myserver",
			opts:    MCPOptions{},
			gomod:   "module github.com/user/coolapp\n\ngo 1.22\n",
			wantFiles: []string{
				"internal/mcp/server.go",
				"cmd/coolapp/cmd_mcp.go",
			},
		},
		{
			name:    "sse transport",
			mcpName: "api",
			opts:    MCPOptions{Module: "github.com/user/myapp", Transport: "sse", Addr: ":9090"},
			gomod:   "module github.com/user/myapp\n\ngo 1.22\n",
			wantFiles: []string{
				"internal/mcp/server.go",
				"cmd/myapp/cmd_mcp.go",
			},
		},
		{
			name:    "http-stream transport",
			mcpName: "api",
			opts:    MCPOptions{Module: "github.com/user/myapp", Transport: "http-stream"},
			gomod:   "module github.com/user/myapp\n\ngo 1.22\n",
			wantFiles: []string{
				"internal/mcp/server.go",
			},
		},
		{
			name:    "empty name",
			mcpName: "",
			opts:    MCPOptions{Module: "github.com/user/myapp"},
			gomod:   "module github.com/user/myapp\n\ngo 1.22\n",
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name:    "invalid transport",
			mcpName: "myserver",
			opts:    MCPOptions{Module: "github.com/user/myapp", Transport: "grpc"},
			gomod:   "module github.com/user/myapp\n\ngo 1.22\n",
			wantErr: true,
			errMsg:  "invalid transport",
		},
		{
			name:    "no module and no go.mod",
			mcpName: "myserver",
			opts:    MCPOptions{},
			gomod:   "",
			wantErr: true,
			errMsg:  "--module is required",
		},
		{
			name:    "json output",
			mcpName: "myserver",
			opts:    MCPOptions{Module: "github.com/user/myapp"},
			genOpts: scaffolding.Options{JSON: true},
			gomod:   "module github.com/user/myapp\n\ngo 1.22\n",
			wantFiles: []string{
				"internal/mcp/server.go",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tt.gomod != "" {
				_ = afero.WriteFile(fs, "go.mod", []byte(tt.gomod), 0644)
			}

			var buf bytes.Buffer
			err := RunMCPInit(&buf, fs, tt.mcpName, tt.opts, tt.genOpts)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify files exist
			for _, f := range tt.wantFiles {
				exists, err := afero.Exists(fs, f)
				if err != nil {
					t.Fatalf("error checking file %s: %v", f, err)
				}
				if !exists {
					t.Errorf("expected file %s to exist", f)
				}
			}

			// Verify JSON output if applicable
			if tt.genOpts.JSON {
				var result MCPResult
				if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
					t.Fatalf("invalid JSON output: %v", err)
				}
				if result.Status != "created" {
					t.Errorf("expected status 'created', got %q", result.Status)
				}
				if result.Name != tt.mcpName {
					t.Errorf("expected name %q, got %q", tt.mcpName, result.Name)
				}
			}
		})
	}
}

func TestRunMCPInit_FileContents(t *testing.T) {
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "go.mod", []byte("module github.com/user/myapp\n\ngo 1.22\n"), 0644)

	var buf bytes.Buffer
	err := RunMCPInit(&buf, fs, "testserver", MCPOptions{
		Module:    "github.com/user/myapp",
		Transport: "sse",
		Addr:      ":9090",
	}, scaffolding.Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify server.go contains transport switch
	serverContent, _ := afero.ReadFile(fs, "internal/mcp/server.go")
	if !strings.Contains(string(serverContent), "StdioTransport") {
		t.Error("server.go should contain StdioTransport")
	}
	if !strings.Contains(string(serverContent), "NewSSEHandler") {
		t.Error("server.go should contain NewSSEHandler")
	}
	if !strings.Contains(string(serverContent), "NewStreamableHTTPHandler") {
		t.Error("server.go should contain NewStreamableHTTPHandler")
	}

	// Verify cmd file references correct module
	cmdContent, _ := afero.ReadFile(fs, "cmd/myapp/cmd_mcp.go")
	if !strings.Contains(string(cmdContent), "github.com/user/myapp/internal/mcp") {
		t.Error("cmd_mcp.go should import the module's internal/mcp package")
	}
	if !strings.Contains(string(cmdContent), `"sse"`) {
		t.Error("cmd_mcp.go should have sse as default transport")
	}
	if !strings.Contains(string(cmdContent), `":9090"`) {
		t.Error("cmd_mcp.go should have :9090 as default addr")
	}

	// Verify tools.go contains greet tool
	toolsContent, _ := afero.ReadFile(fs, "internal/mcp/tools.go")
	if !strings.Contains(string(toolsContent), "greet") {
		t.Error("tools.go should contain greet tool")
	}

	// Verify debug.go contains NewLogger
	debugContent, _ := afero.ReadFile(fs, "internal/mcp/debug.go")
	if !strings.Contains(string(debugContent), "NewLogger") {
		t.Error("debug.go should contain NewLogger")
	}
}

func TestRunMCPInit_CompileGeneratedCode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create real temp directory
	tmpDir := t.TempDir()

	// Use real filesystem
	fs := afero.NewOsFs()

	module := "github.com/test/mcpapp"
	appName := "mcpapp"

	// Generate MCP server files
	var buf bytes.Buffer
	err := RunMCPInit(&buf, afero.NewBasePathFs(fs, tmpDir), "testserver", MCPOptions{
		Module:    module,
		Transport: "stdio",
	}, scaffolding.Options{})
	if err != nil {
		t.Fatalf("RunMCPInit failed: %v", err)
	}

	// Create go.mod
	gomod := "module " + module + "\n\ngo 1.22\n"
	if err := afero.WriteFile(fs, filepath.Join(tmpDir, "go.mod"), []byte(gomod), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	// Create main.go with rootCmd so cmd_mcp.go compiles
	mainGo := `package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "` + appName + `",
	Short: "Test MCP app",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
`
	cmdDir := filepath.Join(tmpDir, "cmd", appName)
	if err := afero.WriteFile(fs, filepath.Join(cmdDir, appName+".go"), []byte(mainGo), 0644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	// Run go mod tidy to fetch dependencies
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = tmpDir
	tidyOut, err := tidy.CombinedOutput()
	if err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, tidyOut)
	}

	// Compile the generated code
	build := exec.Command("go", "build", "./cmd/"+appName)
	build.Dir = tmpDir
	buildOut, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, buildOut)
	}

	// Verify internal/mcp/ also compiles independently
	buildMCP := exec.Command("go", "build", "./internal/mcp/...")
	buildMCP.Dir = tmpDir
	buildMCPOut, err := buildMCP.CombinedOutput()
	if err != nil {
		t.Fatalf("go build ./internal/mcp/... failed: %v\n%s", err, buildMCPOut)
	}

	t.Log("Generated MCP project compiles successfully")
}

func TestRunMCPInit_RejectsNameTraversal(t *testing.T) {
	for _, bad := range []string{"../evil", "sub/evil"} {
		fs := afero.NewMemMapFs()
		var buf bytes.Buffer
		err := RunMCPInit(&buf, fs, bad, MCPOptions{Module: "github.com/x/y"}, scaffolding.Options{})
		if !errors.Is(err, cmderr.ErrInvalidInput) {
			t.Fatalf("name %q: want ErrInvalidInput, got %v", bad, err)
		}
	}
}

func TestDetectModule(t *testing.T) {
	tests := []struct {
		name    string
		gomod   string
		want    string
		wantErr bool
	}{
		{
			name:  "standard module",
			gomod: "module github.com/user/myapp\n\ngo 1.22\n",
			want:  "github.com/user/myapp",
		},
		{
			name:    "no go.mod",
			gomod:   "",
			wantErr: true,
		},
		{
			name:    "no module directive",
			gomod:   "go 1.22\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			if tt.gomod != "" {
				_ = afero.WriteFile(fs, "go.mod", []byte(tt.gomod), 0644)
			}

			got, err := detectModule(fs)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
