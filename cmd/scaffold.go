package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/inovacc/omni/internal/cli/scaffolding"
	scaffoldcobra "github.com/inovacc/omni/internal/cli/scaffolding/cobra"
	"github.com/inovacc/omni/internal/cli/scaffolding/handler"
	scaffoldmcp "github.com/inovacc/omni/internal/cli/scaffolding/mcp"
	"github.com/inovacc/omni/internal/cli/scaffolding/repository"
	"github.com/inovacc/omni/internal/cli/scaffolding/testgen"
)

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Code scaffolding utilities",
	Long: `scaffold provides code generation utilities for scaffolding applications.

Subcommands:
  cobra init    Initialize a new Cobra CLI application
  cobra add     Add a new command to an existing Cobra application
  cobra config  Manage cobra generator configuration
  handler       Generate HTTP handler
  repository    Generate database repository
  test          Generate tests for a Go source file
  mcp           Generate MCP server with tools, resources, and debug logging

Configuration:
  Default values can be set in ~/.cobra.yaml (compatible with cobra-cli).
  Command-line flags override config file values.

Examples:
  omni scaffold cobra init myapp --module github.com/user/myapp
  omni scaffold cobra add serve --parent root
  omni scaffold cobra config --show
  omni scaffold handler user --method GET,POST --framework chi
  omni scaffold repository user --entity User --table users
  omni scaffold test internal/cli/foo/foo.go
  omni scaffold mcp myserver --transport sse --addr :9090`,
}

var scaffoldCobraCmd = &cobra.Command{
	Use:   "cobra",
	Short: "Cobra CLI application generator",
	Long: `Generate Cobra CLI applications and commands.

Configuration file (~/.cobra.yaml):
  author: Your Name <email@example.com>
  license: MIT
  useViper: true
  useService: false
  full: false

Subcommands:
  init       Initialize a new Cobra CLI application
  add        Add a new command to an existing application
  add-tools  Add cmdtree and aicontext to an existing project
  config     Manage generator configuration

Examples:
  omni scaffold cobra init myapp --module github.com/user/myapp
  omni scaffold cobra add serve --parent root`,
}

var scaffoldCobraInitCmd = &cobra.Command{
	Use:   "init <directory>",
	Short: "Initialize a new Cobra CLI application",
	Long: `Initialize a new Cobra CLI application with all necessary scaffolding.

Configuration:
  Reads defaults from ~/.cobra.yaml if present.
  Command-line flags override config file values.

Creates (basic mode):
  - cmd/{name}/{name}.go       Entry point + root command
  - cmd/{name}/cmd_version.go  Version command
  - cmd/{name}/cmd_cmdtree.go  Command tree utility
  - go.mod                     Go module
  - README.md                  Documentation
  - Taskfile.yml               Task runner
  - .gitignore                 Git ignore file
  - .editorconfig              Editor configuration
  - LICENSE                    License file (optional)

With --aicontext:
  - cmd/{name}/cmd_aicontext.go AI context generator for coding agents

With --viper:
  - internal/config/config.go  Viper configuration

With --service:
  - internal/parameters/config.go  Service parameters
  - internal/service/service.go    Service handler (uses inovacc/config)

With --daemon (mutually exclusive with --service):
  - internal/serverinfo/serverinfo.go  PID/version JSON state with stale-PID detection
  - cmd/{name}/cmd_server.go           Cobra subcommands: start/stop/restart/status/install/uninstall
  - cmd/{name}/server.go               Shared start/stop/daemonize logic + --foreground flag
  - cmd/{name}/server_unix.go          Unix helpers (setSysProcAttr, stopProcess, isPrivileged, elevateAndRerun)
  - cmd/{name}/server_systemd.go       systemd unit install/uninstall (Linux/BSD)
  - cmd/{name}/server_darwin.go        launchd plist install/uninstall (macOS)
  - cmd/{name}/server_windows.go       Windows SCM install/uninstall + helpers

With --full (includes all above plus):
  - .goreleaser.yaml               GoReleaser configuration
  - .golangci.yml                  GolangCI-Lint configuration (v2)
  - tools.go                       Build tool dependencies
  - .github/workflows/build.yml    GitHub Actions build workflow
  - .github/workflows/test.yml     GitHub Actions test workflow
  - .github/workflows/release.yaml GitHub Actions release workflow

Examples:
  omni scaffold cobra init myapp --module github.com/user/myapp
  omni scaffold cobra init ./apps/cli --module github.com/user/cli --viper
  omni scaffold cobra init myapp --module github.com/user/myapp --license MIT --author "John Doe"
  omni scaffold cobra init myapp --module github.com/user/myapp --full --service
  omni scaffold cobra init myapp --module github.com/user/myapp --aicontext
  omni scaffold cobra init weaverd --module github.com/user/weaverd --daemon   # self-daemonizing server`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		module, _ := cmd.Flags().GetString("module")
		appName, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		author, _ := cmd.Flags().GetString("author")
		license, _ := cmd.Flags().GetString("license")
		useViper, _ := cmd.Flags().GetBool("viper")
		useService, _ := cmd.Flags().GetBool("service")
		useDaemon, _ := cmd.Flags().GetBool("daemon")
		full, _ := cmd.Flags().GetBool("full")
		aicontext, _ := cmd.Flags().GetBool("aicontext")

		// Build options from flags
		opts := scaffoldcobra.CobraInitOptions{
			Module:      module,
			AppName:     appName,
			Description: description,
			Author:      author,
			License:     license,
			UseViper:    useViper,
			UseService:  useService,
			UseDaemon:   useDaemon,
			Full:        full,
			AIContext:   aicontext,
		}

		// Load config file and merge with flags
		cfg, configPath, err := scaffoldcobra.LoadCobraConfig()
		if err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to load config from %s: %v\n", configPath, err)
		} else if configPath != "" && !jsonOutput {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Using config file: %s\n", configPath)
		}

		// Track which flags were explicitly set
		flagsSet := make(map[string]bool)
		cmd.Flags().Visit(func(f *pflag.Flag) {
			flagsSet[f.Name] = true
		})

		// Merge config with flags (flags take precedence)
		cfg.MergeWithFlags(&opts, flagsSet)

		dir := args[0]
		if !filepath.IsAbs(dir) {
			cwd, _ := os.Getwd()
			dir = filepath.Join(cwd, dir)
		}

		return scaffoldcobra.RunCobraInit(cmd.OutOrStdout(), afero.NewOsFs(), dir, opts, scaffolding.Options{JSON: jsonOutput})
	},
}

var scaffoldCobraAddCmd = &cobra.Command{
	Use:   "add <command-name>",
	Short: "Add a new command to an existing Cobra application",
	Long: `Add a new command to an existing Cobra CLI application.

Creates a new command file in the cmd/{appName}/ directory with cmd_ prefix.

Examples:
  omni scaffold cobra add serve
  omni scaffold cobra add serve --parent root
  omni scaffold cobra add list --parent user --description "List all users"
  omni scaffold cobra add daemon --platform-split   # emits shared + _windows.go + _darwin.go + _unix.go`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		parent, _ := cmd.Flags().GetString("parent")
		description, _ := cmd.Flags().GetString("description")
		dir, _ := cmd.Flags().GetString("dir")
		platformSplit, _ := cmd.Flags().GetBool("platform-split")

		if dir == "" {
			dir, _ = os.Getwd()
		}

		return scaffoldcobra.RunCobraAdd(cmd.OutOrStdout(), afero.NewOsFs(), dir, scaffoldcobra.CobraAddOptions{
			Name:          args[0],
			Parent:        parent,
			Description:   description,
			PlatformSplit: platformSplit,
		}, scaffolding.Options{JSON: jsonOutput})
	},
}

var scaffoldCobraAddToolsCmd = &cobra.Command{
	Use:   "add-tools",
	Short: "Add cmdtree and aicontext to an existing Cobra project",
	Long: `Add cmdtree and aicontext commands to an existing Cobra CLI project.

Creates:
  - cmd/{appName}/cmd_cmdtree.go   Command tree utility

With --aicontext:
  - cmd/{appName}/cmd_aicontext.go AI context generator for coding agents

Examples:
  omni scaffold cobra add-tools
  omni scaffold cobra add-tools --aicontext
  omni scaffold cobra add-tools --dir /path/to/project`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		aicontext, _ := cmd.Flags().GetBool("aicontext")
		dir, _ := cmd.Flags().GetString("dir")

		if dir == "" {
			dir, _ = os.Getwd()
		}

		return scaffoldcobra.RunCobraAddTools(cmd.OutOrStdout(), afero.NewOsFs(), dir, scaffoldcobra.AddToolsOptions{
			AIContext: aicontext,
		}, scaffolding.Options{JSON: jsonOutput})
	},
}

var scaffoldCobraConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage cobra generator configuration",
	Long: `Manage the cobra generator configuration file.

The configuration file is stored at ~/.cobra.yaml and is compatible
with cobra-cli's configuration format.

Available options in config file:
  author: Your Name <email@example.com>
  license: MIT | Apache-2.0 | BSD-3
  useViper: true | false
  useService: true | false
  full: true | false

Examples:
  omni scaffold cobra config --show
  omni scaffold cobra config --init
  omni scaffold cobra config --init --author "John Doe" --license MIT`,
	RunE: func(cmd *cobra.Command, args []string) error {
		showConfig, _ := cmd.Flags().GetBool("show")
		initConfig, _ := cmd.Flags().GetBool("init")

		if showConfig {
			cfg, configPath, err := scaffoldcobra.LoadCobraConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if configPath == "" {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No configuration file found.")
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "\nTo create one, run:")
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  omni scaffold cobra config --init")
				return nil
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Config file: %s\n\n", configPath)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "author: %s\n", cfg.Author)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "license: %s\n", cfg.License)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "useViper: %v\n", cfg.UseViper)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "useService: %v\n", cfg.UseService)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "full: %v\n", cfg.Full)
			return nil
		}

		if initConfig {
			author, _ := cmd.Flags().GetString("author")
			license, _ := cmd.Flags().GetString("license")
			useViper, _ := cmd.Flags().GetBool("viper")
			useService, _ := cmd.Flags().GetBool("service")
			full, _ := cmd.Flags().GetBool("full")

			cfg := &scaffoldcobra.CobraConfig{
				Author:     author,
				License:    license,
				UseViper:   useViper,
				UseService: useService,
				Full:       full,
			}

			configPath := scaffoldcobra.DefaultConfigPath()
			if configPath == "" {
				return fmt.Errorf("could not determine home directory")
			}

			// Check if file already exists
			if _, err := os.Stat(configPath); err == nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Config file already exists: %s\n", configPath)
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Use --show to view current configuration.")
				return nil
			}

			if err := scaffoldcobra.WriteDefaultConfig(afero.NewOsFs(), configPath, cfg); err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created config file: %s\n", configPath)
			return nil
		}

		return cmd.Help()
	},
}

var scaffoldHandlerCmd = &cobra.Command{
	Use:   "handler <name>",
	Short: "Generate HTTP handler",
	Long: `Generate an HTTP handler with the specified name.

Supports multiple frameworks: stdlib, chi, gin, echo

  -p, --package      Package name (default: "handler")
  -d, --dir          Output directory (default: "internal/handler")
  -m, --method       HTTP methods: GET,POST,PUT,DELETE,PATCH (default: GET,POST,PUT,DELETE)
  --path             URL path pattern
  --middleware       Include middleware support
  -f, --framework    Framework: stdlib, chi, gin, echo (default: stdlib)

Examples:
  omni scaffold handler user
  omni scaffold handler user --method GET,POST --framework chi
  omni scaffold handler user --dir handlers --package handlers
  omni scaffold handler product --middleware --framework gin`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		pkg, _ := cmd.Flags().GetString("package")
		dir, _ := cmd.Flags().GetString("dir")
		methodsStr, _ := cmd.Flags().GetString("method")
		path, _ := cmd.Flags().GetString("path")
		middleware, _ := cmd.Flags().GetBool("middleware")
		framework, _ := cmd.Flags().GetString("framework")

		var methods []string
		if methodsStr != "" {
			methods = strings.Split(methodsStr, ",")
		}

		opts := handler.HandlerOptions{
			Package:    pkg,
			Dir:        dir,
			Methods:    methods,
			Path:       path,
			Middleware: middleware,
			Framework:  framework,
		}

		return handler.RunHandlerInit(cmd.OutOrStdout(), afero.NewOsFs(), args[0], opts, scaffolding.Options{JSON: jsonOutput})
	},
}

var scaffoldRepositoryCmd = &cobra.Command{
	Use:   "repository <name>",
	Short: "Generate database repository",
	Long: `Generate a database repository with the specified name.

  -p, --package      Package name (default: "repository")
  -d, --dir          Output directory (default: "internal/repository")
  -e, --entity       Entity struct name (default: capitalized name)
  -t, --table        Database table name (default: lowercase name + "s")
  --db               Database type: postgres, mysql, sqlite (default: postgres)
  --interface        Generate interface (default: true)

Examples:
  omni scaffold repository user
  omni scaffold repository user --entity User --table users
  omni scaffold repository product --db mysql
  omni scaffold repository order --interface=false`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		pkg, _ := cmd.Flags().GetString("package")
		dir, _ := cmd.Flags().GetString("dir")
		entity, _ := cmd.Flags().GetString("entity")
		table, _ := cmd.Flags().GetString("table")
		db, _ := cmd.Flags().GetString("db")
		iface, _ := cmd.Flags().GetBool("interface")

		opts := repository.RepositoryOptions{
			Package:   pkg,
			Dir:       dir,
			Entity:    entity,
			Table:     table,
			DB:        db,
			Interface: iface,
		}

		return repository.RunRepositoryInit(cmd.OutOrStdout(), afero.NewOsFs(), args[0], opts, scaffolding.Options{JSON: jsonOutput})
	},
}

var scaffoldTestCmd = &cobra.Command{
	Use:   "test <file.go>",
	Short: "Generate tests for a Go source file",
	Long: `Generate test stubs for exported functions in a Go source file.

Parses the input file and generates test functions for all exported
functions and methods.

  --table           Generate table-driven tests (default: true)
  --parallel        Add t.Parallel() calls
  --mock            Generate mock setup
  --benchmark       Include benchmark tests
  --fuzz            Include fuzz tests (Go 1.18+)

Examples:
  omni scaffold test internal/cli/foo/foo.go
  omni scaffold test pkg/service/user.go --parallel
  omni scaffold test handler.go --table=false
  omni scaffold test service.go --benchmark --mock`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		table, _ := cmd.Flags().GetBool("table")
		parallel, _ := cmd.Flags().GetBool("parallel")
		mock, _ := cmd.Flags().GetBool("mock")
		benchmark, _ := cmd.Flags().GetBool("benchmark")
		fuzz, _ := cmd.Flags().GetBool("fuzz")

		opts := testgen.TestOptions{
			Table:     table,
			Parallel:  parallel,
			Mock:      mock,
			Benchmark: benchmark,
			Fuzz:      fuzz,
		}

		return testgen.RunTestInit(cmd.OutOrStdout(), afero.NewOsFs(), args[0], opts, scaffolding.Options{JSON: jsonOutput})
	},
}

var scaffoldMCPCmd = &cobra.Command{
	Use:   "mcp <name>",
	Short: "Generate MCP server",
	Long: `Generate an MCP (Model Context Protocol) server with tools, resources, and debug logging.

Generates:
  - internal/mcp/server.go       Server setup with transport selection
  - internal/mcp/tools.go        Example greet tool
  - internal/mcp/resources.go    Example info resource
  - internal/mcp/debug.go        Logger with debug/trace levels
  - cmd/{appName}/cmd_mcp.go     Cobra command with mcp serve

Transports:
  stdio        Standard I/O (default, for CLI integration)
  sse          Server-Sent Events (legacy, HTTP-based)
  http-stream  Streamable HTTP (recommended for remote servers)

Examples:
  omni scaffold mcp myserver
  omni scaffold mcp myserver --transport sse --addr :9090
  omni scaffold mcp myserver --transport http-stream
  omni scaffold mcp myserver --module github.com/user/myapp`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		module, _ := cmd.Flags().GetString("module")
		transport, _ := cmd.Flags().GetString("transport")
		addr, _ := cmd.Flags().GetString("addr")

		opts := scaffoldmcp.MCPOptions{
			Module:    module,
			Transport: transport,
			Addr:      addr,
		}

		return scaffoldmcp.RunMCPInit(cmd.OutOrStdout(), afero.NewOsFs(), args[0], opts, scaffolding.Options{JSON: jsonOutput})
	},
}

func init() {
	rootCmd.AddCommand(scaffoldCmd)
	scaffoldCmd.AddCommand(scaffoldCobraCmd)
	scaffoldCmd.AddCommand(scaffoldHandlerCmd)
	scaffoldCmd.AddCommand(scaffoldRepositoryCmd)
	scaffoldCmd.AddCommand(scaffoldTestCmd)
	scaffoldCmd.AddCommand(scaffoldMCPCmd)
	scaffoldCobraCmd.AddCommand(scaffoldCobraInitCmd)
	scaffoldCobraCmd.AddCommand(scaffoldCobraAddCmd)
	scaffoldCobraCmd.AddCommand(scaffoldCobraAddToolsCmd)
	scaffoldCobraCmd.AddCommand(scaffoldCobraConfigCmd)

	// Persistent flags for scaffold command
	scaffoldCmd.PersistentFlags().Bool("json", false, "output as JSON")

	// Flags for cobra init
	scaffoldCobraInitCmd.Flags().StringP("module", "m", "", "Go module path (required)")
	scaffoldCobraInitCmd.Flags().StringP("name", "n", "", "application name (defaults to directory name)")
	scaffoldCobraInitCmd.Flags().StringP("description", "d", "", "application description")
	scaffoldCobraInitCmd.Flags().StringP("author", "a", "", "author name")
	scaffoldCobraInitCmd.Flags().StringP("license", "l", "", "license type (MIT, Apache-2.0, BSD-3)")
	scaffoldCobraInitCmd.Flags().Bool("viper", false, "include viper for configuration")
	scaffoldCobraInitCmd.Flags().Bool("service", false, "include OS service pattern (kardianos/service)")
	scaffoldCobraInitCmd.Flags().Bool("daemon", false, "include self-daemonizing PID-file pattern with server start/stop/restart/status/install/uninstall (mutually exclusive with --service)")
	scaffoldCobraInitCmd.Flags().Bool("full", false, "full project with goreleaser, workflows, etc.")
	scaffoldCobraInitCmd.Flags().Bool("aicontext", false, "include aicontext command for AI coding agents")
	_ = scaffoldCobraInitCmd.MarkFlagRequired("module")

	// Flags for cobra add
	scaffoldCobraAddCmd.Flags().StringP("parent", "p", "root", "parent command")
	scaffoldCobraAddCmd.Flags().StringP("description", "d", "", "command description")
	scaffoldCobraAddCmd.Flags().String("dir", "", "project directory (defaults to current directory)")
	scaffoldCobraAddCmd.Flags().Bool("platform-split", false, "emit platform-split files: cmd_<name>_{windows,darwin,unix}.go alongside the shared file")

	// Flags for cobra add-tools
	scaffoldCobraAddToolsCmd.Flags().Bool("aicontext", false, "include aicontext command for AI coding agents")
	scaffoldCobraAddToolsCmd.Flags().String("dir", "", "project directory (defaults to current directory)")

	// Flags for cobra config
	scaffoldCobraConfigCmd.Flags().Bool("show", false, "show current configuration")
	scaffoldCobraConfigCmd.Flags().Bool("init", false, "create a new configuration file")
	scaffoldCobraConfigCmd.Flags().StringP("author", "a", "", "author name for config")
	scaffoldCobraConfigCmd.Flags().StringP("license", "l", "", "license type for config")
	scaffoldCobraConfigCmd.Flags().Bool("viper", false, "set useViper in config")
	scaffoldCobraConfigCmd.Flags().Bool("service", false, "set useService in config")
	scaffoldCobraConfigCmd.Flags().Bool("full", false, "set full in config")

	// Flags for handler
	scaffoldHandlerCmd.Flags().StringP("package", "p", "handler", "package name")
	scaffoldHandlerCmd.Flags().StringP("dir", "d", "internal/handler", "output directory")
	scaffoldHandlerCmd.Flags().StringP("method", "m", "GET,POST,PUT,DELETE", "HTTP methods (comma-separated)")
	scaffoldHandlerCmd.Flags().String("path", "", "URL path pattern")
	scaffoldHandlerCmd.Flags().Bool("middleware", false, "include middleware support")
	scaffoldHandlerCmd.Flags().StringP("framework", "f", "stdlib", "framework: stdlib, chi, gin, echo")

	// Flags for repository
	scaffoldRepositoryCmd.Flags().StringP("package", "p", "repository", "package name")
	scaffoldRepositoryCmd.Flags().StringP("dir", "d", "internal/repository", "output directory")
	scaffoldRepositoryCmd.Flags().StringP("entity", "e", "", "entity struct name")
	scaffoldRepositoryCmd.Flags().StringP("table", "t", "", "database table name")
	scaffoldRepositoryCmd.Flags().String("db", "postgres", "database type: postgres, mysql, sqlite")
	scaffoldRepositoryCmd.Flags().Bool("interface", true, "generate interface")

	// Flags for test
	scaffoldTestCmd.Flags().Bool("table", true, "generate table-driven tests")
	scaffoldTestCmd.Flags().Bool("parallel", false, "add t.Parallel() calls")
	scaffoldTestCmd.Flags().Bool("mock", false, "generate mock setup")
	scaffoldTestCmd.Flags().Bool("benchmark", false, "include benchmark tests")
	scaffoldTestCmd.Flags().Bool("fuzz", false, "include fuzz tests")

	// Flags for mcp
	scaffoldMCPCmd.Flags().StringP("module", "m", "", "Go module path (auto-detected from go.mod)")
	scaffoldMCPCmd.Flags().String("transport", "stdio", "transport type: stdio, sse, http-stream")
	scaffoldMCPCmd.Flags().String("addr", ":8080", "listen address (for sse/http-stream)")
}
