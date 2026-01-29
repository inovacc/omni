package cmd

import (
	"os"
	"path/filepath"

	"github.com/inovacc/omni/internal/cli/generate"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Code generation utilities",
	Long: `generate provides code generation utilities for scaffolding applications.

Subcommands:
  cobra init    Initialize a new Cobra CLI application
  cobra add     Add a new command to an existing Cobra application

Examples:
  omni generate cobra init myapp --module github.com/user/myapp
  omni generate cobra add serve --parent root`,
}

var generateCobraCmd = &cobra.Command{
	Use:   "cobra",
	Short: "Cobra CLI application generator",
	Long: `Generate Cobra CLI applications and commands.

Subcommands:
  init    Initialize a new Cobra CLI application
  add     Add a new command to an existing application`,
}

var generateCobraInitCmd = &cobra.Command{
	Use:   "init <directory>",
	Short: "Initialize a new Cobra CLI application",
	Long: `Initialize a new Cobra CLI application with all necessary scaffolding.

Creates:
  - main.go          Entry point
  - cmd/root.go      Root command
  - cmd/version.go   Version command
  - go.mod           Go module
  - README.md        Documentation
  - Taskfile.yml     Task runner
  - .gitignore       Git ignore file
  - LICENSE          License file (optional)

With --viper:
  - internal/config/config.go  Viper configuration

Examples:
  omni generate cobra init myapp --module github.com/user/myapp
  omni generate cobra init ./apps/cli --module github.com/user/cli --viper
  omni generate cobra init myapp --module github.com/user/myapp --license MIT --author "John Doe"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		module, _ := cmd.Flags().GetString("module")
		appName, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		author, _ := cmd.Flags().GetString("author")
		license, _ := cmd.Flags().GetString("license")
		useViper, _ := cmd.Flags().GetBool("viper")

		dir := args[0]
		if !filepath.IsAbs(dir) {
			cwd, _ := os.Getwd()
			dir = filepath.Join(cwd, dir)
		}

		return generate.RunCobraInit(os.Stdout, dir, generate.CobraInitOptions{
			Module:      module,
			AppName:     appName,
			Description: description,
			Author:      author,
			License:     license,
			UseViper:    useViper,
		}, generate.Options{JSON: jsonOutput})
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

		return generate.RunCobraAdd(os.Stdout, dir, generate.CobraAddOptions{
			Name:        args[0],
			Parent:      parent,
			Description: description,
		}, generate.Options{JSON: jsonOutput})
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.AddCommand(generateCobraCmd)
	generateCobraCmd.AddCommand(generateCobraInitCmd)
	generateCobraCmd.AddCommand(generateCobraAddCmd)

	// Persistent flags for generate command
	generateCmd.PersistentFlags().Bool("json", false, "output as JSON")

	// Flags for cobra init
	generateCobraInitCmd.Flags().StringP("module", "m", "", "Go module path (required)")
	generateCobraInitCmd.Flags().StringP("name", "n", "", "application name (defaults to directory name)")
	generateCobraInitCmd.Flags().StringP("description", "d", "", "application description")
	generateCobraInitCmd.Flags().StringP("author", "a", "", "author name")
	generateCobraInitCmd.Flags().StringP("license", "l", "", "license type (MIT, Apache-2.0, BSD-3)")
	generateCobraInitCmd.Flags().Bool("viper", false, "include viper for configuration")
	_ = generateCobraInitCmd.MarkFlagRequired("module")

	// Flags for cobra add
	generateCobraAddCmd.Flags().StringP("parent", "p", "root", "parent command")
	generateCobraAddCmd.Flags().StringP("description", "d", "", "command description")
	generateCobraAddCmd.Flags().String("dir", "", "project directory (defaults to current directory)")
}
