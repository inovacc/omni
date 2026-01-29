package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/yamlutil"
	"github.com/spf13/cobra"
)

var yamlCmd = &cobra.Command{
	Use:   "yaml",
	Short: "YAML utilities",
	Long: `YAML utilities for validation and formatting.

Subcommands:
  validate    Validate YAML syntax
  fmt         Format/beautify YAML

Examples:
  omni yaml validate config.yaml
  omni yaml fmt config.yaml`,
}

var yamlValidateCmd = &cobra.Command{
	Use:   "validate [FILE...]",
	Short: "Validate YAML syntax",
	Long: `Validate YAML syntax for one or more files.

Checks that the input is valid YAML. Supports multi-document YAML files.

Examples:
  omni yaml validate config.yaml
  omni yaml validate *.yaml
  omni yaml validate --strict config.yaml
  cat config.yaml | omni yaml validate
  omni yaml validate --json config.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := yamlutil.ValidateOptions{}
		opts.JSON, _ = cmd.Flags().GetBool("json")
		opts.Strict, _ = cmd.Flags().GetBool("strict")

		return yamlutil.RunValidate(os.Stdout, args, opts)
	},
}

var yamlFmtCmd = &cobra.Command{
	Use:   "fmt [FILE]",
	Short: "Format YAML",
	Long: `Format and beautify YAML.

Parses YAML and outputs it with consistent formatting.

Examples:
  omni yaml fmt config.yaml
  cat config.yaml | omni yaml fmt
  omni yaml fmt --indent 4 config.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := yamlutil.FormatOptions{}
		opts.Indent, _ = cmd.Flags().GetInt("indent")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return yamlutil.RunFormat(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(yamlCmd)
	yamlCmd.AddCommand(yamlValidateCmd)
	yamlCmd.AddCommand(yamlFmtCmd)

	yamlValidateCmd.Flags().Bool("json", false, "output as JSON")
	yamlValidateCmd.Flags().Bool("strict", false, "fail on unknown fields")

	yamlFmtCmd.Flags().IntP("indent", "i", 2, "indentation width")
	yamlFmtCmd.Flags().Bool("json", false, "output as JSON instead of YAML")
}
