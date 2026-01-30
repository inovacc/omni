package cmd

import (
	"github.com/inovacc/omni/internal/cli/tomlutil"
	"github.com/spf13/cobra"
)

var tomlCmd = &cobra.Command{
	Use:   "toml",
	Short: "TOML utilities",
	Long: `TOML utilities for validation and formatting.

Subcommands:
  validate    Validate TOML syntax
  fmt         Format/beautify TOML

Examples:
  omni toml validate config.toml
  omni toml fmt config.toml`,
}

var tomlValidateCmd = &cobra.Command{
	Use:   "validate [FILE...]",
	Short: "Validate TOML syntax",
	Long: `Validate TOML syntax for one or more files.

Checks that the input is valid TOML.

Examples:
  omni toml validate config.toml
  omni toml validate *.toml
  cat config.toml | omni toml validate
  omni toml validate --json config.toml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := tomlutil.ValidateOptions{}
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return tomlutil.RunValidate(cmd.OutOrStdout(), args, opts)
	},
}

var tomlFmtCmd = &cobra.Command{
	Use:   "fmt [FILE]",
	Short: "Format TOML",
	Long: `Format and beautify TOML.

Parses TOML and outputs it with consistent formatting.

Examples:
  omni toml fmt config.toml
  cat config.toml | omni toml fmt
  omni toml fmt --indent 4 config.toml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := tomlutil.FormatOptions{}
		opts.Indent, _ = cmd.Flags().GetInt("indent")

		return tomlutil.RunFormat(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(tomlCmd)
	tomlCmd.AddCommand(tomlValidateCmd)
	tomlCmd.AddCommand(tomlFmtCmd)

	tomlValidateCmd.Flags().Bool("json", false, "output as JSON")

	tomlFmtCmd.Flags().IntP("indent", "i", 2, "indentation width")
}
