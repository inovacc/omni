package cmd

import (
	"github.com/inovacc/omni/internal/cli/buf"
	"github.com/spf13/cobra"
)

var bufCmd = &cobra.Command{
	Use:   "buf",
	Short: "Protocol buffer utilities (lint, format, compile, generate)",
	Long: `Protocol buffer utilities inspired by buf.build.

Subcommands:
  lint       Lint proto files
  format     Format proto files
  compile    Compile proto files
  breaking   Check for breaking changes
  generate   Generate code from proto files
  mod        Module management (init)
  ls-files   List proto files

Examples:
  omni buf lint
  omni buf format --write
  omni buf compile -o image.bin
  omni buf breaking --against ../v1
  omni buf generate
  omni buf mod init buf.build/org/repo`,
}

var bufLintCmd = &cobra.Command{
	Use:   "lint [DIR]",
	Short: "Lint proto files",
	Long: `Lint proto files for style and structure issues.

Uses buf.yaml configuration if present. Default rules: STANDARD

Flags:
  --error-format=FORMAT  Output format: text, json, github-actions (default: text)
  --exclude-path=PATH    Paths to exclude (can be repeated)
  --config=FILE          Custom config file path

Categories:
  MINIMAL    Minimal set of rules
  BASIC      Basic rules (includes MINIMAL)
  STANDARD   Standard rules (includes BASIC)
  COMMENTS   Comment-related rules

Examples:
  omni buf lint
  omni buf lint ./proto
  omni buf lint --error-format=json
  omni buf lint --exclude-path=vendor`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		opts := buf.LintOptions{}
		opts.ErrorFormat, _ = cmd.Flags().GetString("error-format")
		opts.ExcludePath, _ = cmd.Flags().GetStringSlice("exclude-path")
		opts.Config, _ = cmd.Flags().GetString("config")

		return buf.RunLint(cmd.OutOrStdout(), dir, opts)
	},
}

var bufFormatCmd = &cobra.Command{
	Use:     "format [DIR]",
	Aliases: []string{"fmt"},
	Short:   "Format proto files",
	Long: `Format proto files with consistent style.

Flags:
  -w, --write       Rewrite files in place
  -d, --diff        Display diff instead of formatted output
  --exit-code       Exit with non-zero if files are not formatted

Examples:
  omni buf format
  omni buf format --write
  omni buf format --diff
  omni buf format --exit-code  # for CI`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		opts := buf.FormatOptions{}
		opts.Write, _ = cmd.Flags().GetBool("write")
		opts.Diff, _ = cmd.Flags().GetBool("diff")
		opts.ExitCode, _ = cmd.Flags().GetBool("exit-code")

		return buf.RunFormat(cmd.OutOrStdout(), dir, opts)
	},
}

var bufCompileCmd = &cobra.Command{
	Use:   "compile [DIR]",
	Short: "Compile proto files",
	Long: `Compile proto files and output an image.

Flags:
  -o, --output=FILE      Output file (.bin or .json)
  --exclude-path=PATH    Paths to exclude
  --error-format=FORMAT  Output format: text, json, github-actions

Examples:
  omni buf compile
  omni buf compile -o image.bin
  omni buf compile -o image.json
  omni buf compile --exclude-path=vendor`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		opts := buf.BuildOptions{}
		opts.Output, _ = cmd.Flags().GetString("output")
		opts.ExcludePath, _ = cmd.Flags().GetStringSlice("exclude-path")
		opts.ErrorFormat, _ = cmd.Flags().GetString("error-format")

		return buf.RunBuild(cmd.OutOrStdout(), dir, opts)
	},
}

var bufBreakingCmd = &cobra.Command{
	Use:   "breaking [DIR]",
	Short: "Check for breaking changes",
	Long: `Check for breaking changes against a previous version.

Flags:
  --against=PATH         Source to compare against (required)
  --exclude-path=PATH    Paths to exclude
  --exclude-imports      Don't check imported files
  --error-format=FORMAT  Output format: text, json, github-actions

Breaking change rules:
  FILE_NO_DELETE      Files cannot be deleted
  PACKAGE_NO_DELETE   Packages cannot be changed
  MESSAGE_NO_DELETE   Messages cannot be deleted
  FIELD_NO_DELETE     Fields cannot be deleted
  FIELD_SAME_TYPE     Field types cannot change
  ENUM_NO_DELETE      Enums cannot be deleted
  SERVICE_NO_DELETE   Services cannot be deleted
  RPC_NO_DELETE       RPCs cannot be deleted

Examples:
  omni buf breaking --against ../v1
  omni buf breaking --against ./baseline --error-format=json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		opts := buf.BreakingOptions{}
		opts.Against, _ = cmd.Flags().GetString("against")
		opts.ExcludePath, _ = cmd.Flags().GetStringSlice("exclude-path")
		opts.ExcludeImports, _ = cmd.Flags().GetBool("exclude-imports")
		opts.ErrorFormat, _ = cmd.Flags().GetString("error-format")

		return buf.RunBreaking(cmd.OutOrStdout(), dir, opts)
	},
}

var bufGenerateCmd = &cobra.Command{
	Use:     "generate [DIR]",
	Aliases: []string{"gen"},
	Short:   "Generate code from proto files",
	Long: `Generate code using plugins defined in buf.gen.yaml.

Flags:
  --template=FILE        Alternate buf.gen.yaml location
  -o, --output=DIR       Base output directory
  --include-imports      Include imported files in generation

buf.gen.yaml example:
  version: v1
  plugins:
    - local: protoc-gen-go
      out: gen/go
      opt:
        - paths=source_relative

Examples:
  omni buf generate
  omni buf generate --template=custom.gen.yaml
  omni buf generate -o ./generated`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		opts := buf.GenerateOptions{}
		opts.Template, _ = cmd.Flags().GetString("template")
		opts.Output, _ = cmd.Flags().GetString("output")
		opts.IncludeImports, _ = cmd.Flags().GetBool("include-imports")

		return buf.RunGenerate(cmd.OutOrStdout(), dir, opts)
	},
}

var bufModCmd = &cobra.Command{
	Use:   "mod",
	Short: "Module management commands",
	Long: `Module management commands.

Subcommands:
  init    Initialize a new buf module
  update  Update dependencies`,
}

var bufModInitCmd = &cobra.Command{
	Use:   "init [NAME]",
	Short: "Initialize a new buf module",
	Long: `Initialize a new buf.yaml configuration file.

Examples:
  omni buf mod init
  omni buf mod init buf.build/org/repo`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := cmd.Flags().GetString("dir")
		if dir == "" {
			dir = "."
		}

		moduleName := ""
		if len(args) > 0 {
			moduleName = args[0]
		}

		return buf.RunInit(cmd.OutOrStdout(), dir, moduleName)
	},
}

var bufModUpdateCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"up"},
	Short:   "Update dependencies",
	Long: `Update dependencies listed in buf.yaml.

Note: Full dependency resolution requires network access to BSR.

Examples:
  omni buf mod update`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		return buf.RunDepUpdate(cmd.OutOrStdout(), dir)
	},
}

var bufLsFilesCmd = &cobra.Command{
	Use:   "ls-files [DIR]",
	Short: "List proto files in the module",
	Long: `List all proto files in the module.

Examples:
  omni buf ls-files
  omni buf ls-files ./proto`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		return buf.RunLsFiles(cmd.OutOrStdout(), dir)
	},
}

func init() {
	rootCmd.AddCommand(bufCmd)

	// Add subcommands
	bufCmd.AddCommand(bufLintCmd)
	bufCmd.AddCommand(bufFormatCmd)
	bufCmd.AddCommand(bufCompileCmd)
	bufCmd.AddCommand(bufBreakingCmd)
	bufCmd.AddCommand(bufGenerateCmd)
	bufCmd.AddCommand(bufModCmd)
	bufCmd.AddCommand(bufLsFilesCmd)

	// Add mod subcommands
	bufModCmd.AddCommand(bufModInitCmd)
	bufModCmd.AddCommand(bufModUpdateCmd)

	// buf lint flags
	bufLintCmd.Flags().String("error-format", "text", "output format (text, json, github-actions)")
	bufLintCmd.Flags().StringSlice("exclude-path", nil, "paths to exclude")
	bufLintCmd.Flags().String("config", "", "custom config file path")

	// buf format flags
	bufFormatCmd.Flags().BoolP("write", "w", false, "rewrite files in place")
	bufFormatCmd.Flags().BoolP("diff", "d", false, "display diff")
	bufFormatCmd.Flags().Bool("exit-code", false, "exit with non-zero if files unformatted")

	// buf compile flags
	bufCompileCmd.Flags().StringP("output", "o", "", "output file path")
	bufCompileCmd.Flags().StringSlice("exclude-path", nil, "paths to exclude")
	bufCompileCmd.Flags().String("error-format", "text", "output format (text, json, github-actions)")

	// buf breaking flags
	bufBreakingCmd.Flags().String("against", "", "source to compare against (required)")
	bufBreakingCmd.Flags().StringSlice("exclude-path", nil, "paths to exclude")
	bufBreakingCmd.Flags().Bool("exclude-imports", false, "don't check imported files")
	bufBreakingCmd.Flags().String("error-format", "text", "output format (text, json, github-actions)")

	// buf generate flags
	bufGenerateCmd.Flags().String("template", "", "alternate buf.gen.yaml location")
	bufGenerateCmd.Flags().StringP("output", "o", "", "base output directory")
	bufGenerateCmd.Flags().Bool("include-imports", false, "include imported files")

	// buf mod init flags
	bufModInitCmd.Flags().String("dir", ".", "directory to initialize")
}
