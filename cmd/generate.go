package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/inovacc/omni/internal/cli/generate"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Code generation utilities",
	Long: `generate provides code generation utilities for scaffolding applications.

Subcommands:
  cobra init    Initialize a new Cobra CLI application
  cobra add     Add a new command to an existing Cobra application
  cobra config  Manage cobra generator configuration
  handler       Generate HTTP handler
  repository    Generate database repository
  test          Generate tests for a Go source file

Configuration:
  Default values can be set in ~/.cobra.yaml (compatible with cobra-cli).
  Command-line flags override config file values.

Examples:
  omni generate cobra init myapp --module github.com/user/myapp
  omni generate cobra add serve --parent root
  omni generate cobra config --show
  omni generate handler user --method GET,POST --framework chi
  omni generate repository user --entity User --table users
  omni generate test internal/cli/foo/foo.go`,
}

var generateCobraCmd = &cobra.Command{
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
  init    Initialize a new Cobra CLI application
  add     Add a new command to an existing application
  config  Manage generator configuration`,
}

var generateCobraInitCmd = &cobra.Command{
	Use:   "init <directory>",
	Short: "Initialize a new Cobra CLI application",
	Long: `Initialize a new Cobra CLI application with all necessary scaffolding.

Configuration:
  Reads defaults from ~/.cobra.yaml if present.
  Command-line flags override config file values.

Creates (basic mode):
  - main.go          Entry point
  - cmd/root.go      Root command
  - cmd/version.go   Version command
  - go.mod           Go module
  - README.md        Documentation
  - Taskfile.yml     Task runner
  - .gitignore       Git ignore file
  - .editorconfig    Editor configuration
  - LICENSE          License file (optional)

With --viper:
  - internal/config/config.go  Viper configuration

With --service:
  - internal/parameters/config.go  Service parameters
  - internal/service/service.go    Service handler (uses inovacc/config)

With --full (includes all above plus):
  - .goreleaser.yaml              GoReleaser configuration
  - .golangci.yml                 GolangCI-Lint configuration (v2)
  - tools.go                      Build tool dependencies
  - .github/workflows/build.yml   GitHub Actions build workflow
  - .github/workflows/test.yml    GitHub Actions test workflow
  - .github/workflows/release.yaml GitHub Actions release workflow

Examples:
  omni generate cobra init myapp --module github.com/user/myapp
  omni generate cobra init ./apps/cli --module github.com/user/cli --viper
  omni generate cobra init myapp --module github.com/user/myapp --license MIT --author "John Doe"
  omni generate cobra init myapp --module github.com/user/myapp --full --service`,
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
		full, _ := cmd.Flags().GetBool("full")

		// Build options from flags
		opts := generate.CobraInitOptions{
			Module:      module,
			AppName:     appName,
			Description: description,
			Author:      author,
			License:     license,
			UseViper:    useViper,
			UseService:  useService,
			Full:        full,
		}

		// Load config file and merge with flags
		cfg, configPath, err := generate.LoadCobraConfig()
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

		return generate.RunCobraInit(cmd.OutOrStdout(), dir, opts, generate.Options{JSON: jsonOutput})
	},
}

var generateCobraAddCmd = &cobra.Command{
	Use:   "add <command-name>",
	Short: "Add a new command to an existing Cobra application",
	Long: `Add a new command to an existing Cobra CLI application.

Creates a new command file in the cmd/ directory with the proper structure.

Examples:
  omni generate cobra add serve
  omni generate cobra add serve --parent root
  omni generate cobra add list --parent user --description "List all users"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		parent, _ := cmd.Flags().GetString("parent")
		description, _ := cmd.Flags().GetString("description")
		dir, _ := cmd.Flags().GetString("dir")

		if dir == "" {
			dir, _ = os.Getwd()
		}

		return generate.RunCobraAdd(cmd.OutOrStdout(), dir, generate.CobraAddOptions{
			Name:        args[0],
			Parent:      parent,
			Description: description,
		}, generate.Options{JSON: jsonOutput})
	},
}

var generateCobraConfigCmd = &cobra.Command{
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
  omni generate cobra config --show
  omni generate cobra config --init
  omni generate cobra config --init --author "John Doe" --license MIT`,
	RunE: func(cmd *cobra.Command, args []string) error {
		showConfig, _ := cmd.Flags().GetBool("show")
		initConfig, _ := cmd.Flags().GetBool("init")

		if showConfig {
			cfg, configPath, err := generate.LoadCobraConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if configPath == "" {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No configuration file found.")
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "\nTo create one, run:")
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "  omni generate cobra config --init")
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

			cfg := &generate.CobraConfig{
				Author:     author,
				License:    license,
				UseViper:   useViper,
				UseService: useService,
				Full:       full,
			}

			configPath := generate.DefaultConfigPath()
			if configPath == "" {
				return fmt.Errorf("could not determine home directory")
			}

			// Check if file already exists
			if _, err := os.Stat(configPath); err == nil {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Config file already exists: %s\n", configPath)
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Use --show to view current configuration.")
				return nil
			}

			if err := generate.WriteDefaultConfig(configPath, cfg); err != nil {
				return fmt.Errorf("failed to write config: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created config file: %s\n", configPath)
			return nil
		}

		return cmd.Help()
	},
}

var generateHandlerCmd = &cobra.Command{
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
  omni generate handler user
  omni generate handler user --method GET,POST --framework chi
  omni generate handler user --dir handlers --package handlers
  omni generate handler product --middleware --framework gin`,
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

		opts := generate.HandlerOptions{
			Package:    pkg,
			Dir:        dir,
			Methods:    methods,
			Path:       path,
			Middleware: middleware,
			Framework:  framework,
		}

		return generate.RunHandlerInit(cmd.OutOrStdout(), args[0], opts, generate.Options{JSON: jsonOutput})
	},
}

var generateRepositoryCmd = &cobra.Command{
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
  omni generate repository user
  omni generate repository user --entity User --table users
  omni generate repository product --db mysql
  omni generate repository order --interface=false`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		pkg, _ := cmd.Flags().GetString("package")
		dir, _ := cmd.Flags().GetString("dir")
		entity, _ := cmd.Flags().GetString("entity")
		table, _ := cmd.Flags().GetString("table")
		db, _ := cmd.Flags().GetString("db")
		iface, _ := cmd.Flags().GetBool("interface")

		opts := generate.RepositoryOptions{
			Package:   pkg,
			Dir:       dir,
			Entity:    entity,
			Table:     table,
			DB:        db,
			Interface: iface,
		}

		return generate.RunRepositoryInit(cmd.OutOrStdout(), args[0], opts, generate.Options{JSON: jsonOutput})
	},
}

var generateTestCmd = &cobra.Command{
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
  omni generate test internal/cli/foo/foo.go
  omni generate test pkg/service/user.go --parallel
  omni generate test handler.go --table=false
  omni generate test service.go --benchmark --mock`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		table, _ := cmd.Flags().GetBool("table")
		parallel, _ := cmd.Flags().GetBool("parallel")
		mock, _ := cmd.Flags().GetBool("mock")
		benchmark, _ := cmd.Flags().GetBool("benchmark")
		fuzz, _ := cmd.Flags().GetBool("fuzz")

		opts := generate.TestOptions{
			Table:     table,
			Parallel:  parallel,
			Mock:      mock,
			Benchmark: benchmark,
			Fuzz:      fuzz,
		}

		return generate.RunTestInit(cmd.OutOrStdout(), args[0], opts, generate.Options{JSON: jsonOutput})
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.AddCommand(generateCobraCmd)
	generateCmd.AddCommand(generateHandlerCmd)
	generateCmd.AddCommand(generateRepositoryCmd)
	generateCmd.AddCommand(generateTestCmd)
	generateCobraCmd.AddCommand(generateCobraInitCmd)
	generateCobraCmd.AddCommand(generateCobraAddCmd)
	generateCobraCmd.AddCommand(generateCobraConfigCmd)

	// Persistent flags for generate command
	generateCmd.PersistentFlags().Bool("json", false, "output as JSON")

	// Flags for cobra init
	generateCobraInitCmd.Flags().StringP("module", "m", "", "Go module path (required)")
	generateCobraInitCmd.Flags().StringP("name", "n", "", "application name (defaults to directory name)")
	generateCobraInitCmd.Flags().StringP("description", "d", "", "application description")
	generateCobraInitCmd.Flags().StringP("author", "a", "", "author name")
	generateCobraInitCmd.Flags().StringP("license", "l", "", "license type (MIT, Apache-2.0, BSD-3)")
	generateCobraInitCmd.Flags().Bool("viper", false, "include viper for configuration")
	generateCobraInitCmd.Flags().Bool("service", false, "include service pattern with inovacc/config")
	generateCobraInitCmd.Flags().Bool("full", false, "full project with goreleaser, workflows, etc.")
	_ = generateCobraInitCmd.MarkFlagRequired("module")

	// Flags for cobra add
	generateCobraAddCmd.Flags().StringP("parent", "p", "root", "parent command")
	generateCobraAddCmd.Flags().StringP("description", "d", "", "command description")
	generateCobraAddCmd.Flags().String("dir", "", "project directory (defaults to current directory)")

	// Flags for cobra config
	generateCobraConfigCmd.Flags().Bool("show", false, "show current configuration")
	generateCobraConfigCmd.Flags().Bool("init", false, "create a new configuration file")
	generateCobraConfigCmd.Flags().StringP("author", "a", "", "author name for config")
	generateCobraConfigCmd.Flags().StringP("license", "l", "", "license type for config")
	generateCobraConfigCmd.Flags().Bool("viper", false, "set useViper in config")
	generateCobraConfigCmd.Flags().Bool("service", false, "set useService in config")
	generateCobraConfigCmd.Flags().Bool("full", false, "set full in config")

	// Flags for handler
	generateHandlerCmd.Flags().StringP("package", "p", "handler", "package name")
	generateHandlerCmd.Flags().StringP("dir", "d", "internal/handler", "output directory")
	generateHandlerCmd.Flags().StringP("method", "m", "GET,POST,PUT,DELETE", "HTTP methods (comma-separated)")
	generateHandlerCmd.Flags().String("path", "", "URL path pattern")
	generateHandlerCmd.Flags().Bool("middleware", false, "include middleware support")
	generateHandlerCmd.Flags().StringP("framework", "f", "stdlib", "framework: stdlib, chi, gin, echo")

	// Flags for repository
	generateRepositoryCmd.Flags().StringP("package", "p", "repository", "package name")
	generateRepositoryCmd.Flags().StringP("dir", "d", "internal/repository", "output directory")
	generateRepositoryCmd.Flags().StringP("entity", "e", "", "entity struct name")
	generateRepositoryCmd.Flags().StringP("table", "t", "", "database table name")
	generateRepositoryCmd.Flags().String("db", "postgres", "database type: postgres, mysql, sqlite")
	generateRepositoryCmd.Flags().Bool("interface", true, "generate interface")

	// Flags for test
	generateTestCmd.Flags().Bool("table", true, "generate table-driven tests")
	generateTestCmd.Flags().Bool("parallel", false, "add t.Parallel() calls")
	generateTestCmd.Flags().Bool("mock", false, "generate mock setup")
	generateTestCmd.Flags().Bool("benchmark", false, "include benchmark tests")
	generateTestCmd.Flags().Bool("fuzz", false, "include fuzz tests")
}
