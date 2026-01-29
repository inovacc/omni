package cmd

import (
	"github.com/inovacc/omni/internal/cli/yaml2struct"
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
  tostruct    Convert YAML to Go struct definition

Examples:
  omni yaml validate config.yaml
  omni yaml fmt config.yaml
  omni yaml tostruct config.yaml`,
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

		return yamlutil.RunValidate(cmd.OutOrStdout(), args, opts)
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

		return yamlutil.RunFormat(cmd.OutOrStdout(), args, opts)
	},
}

var yamlToStructCmd = &cobra.Command{
	Use:     "tostruct [FILE]",
	Aliases: []string{"2struct", "gostruct"},
	Short:   "Convert YAML to Go struct definition",
	Long: `Convert YAML data to a Go struct definition.

  -n, --name=NAME      struct name (default "Root")
  -p, --package=PKG    package name (default "main")
  --inline             inline nested structs
  --omitempty          add omitempty to all fields

Examples:
  omni yaml tostruct config.yaml
  cat config.yaml | omni yaml tostruct
  omni yaml tostruct -n Config -p models config.yaml
  omni yaml tostruct --omitempty config.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := yaml2struct.Options{}

		opts.Name, _ = cmd.Flags().GetString("name")
		opts.Package, _ = cmd.Flags().GetString("package")
		opts.Inline, _ = cmd.Flags().GetBool("inline")
		opts.OmitEmpty, _ = cmd.Flags().GetBool("omitempty")

		return yaml2struct.RunYAML2Struct(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(yamlCmd)
	yamlCmd.AddCommand(yamlValidateCmd)
	yamlCmd.AddCommand(yamlFmtCmd)
	yamlCmd.AddCommand(yamlToStructCmd)

	yamlValidateCmd.Flags().Bool("json", false, "output as JSON")
	yamlValidateCmd.Flags().Bool("strict", false, "fail on unknown fields")

	yamlFmtCmd.Flags().IntP("indent", "i", 2, "indentation width")
	yamlFmtCmd.Flags().Bool("json", false, "output as JSON instead of YAML")

	// tostruct flags
	yamlToStructCmd.Flags().StringP("name", "n", "Root", "struct name")
	yamlToStructCmd.Flags().StringP("package", "p", "main", "package name")
	yamlToStructCmd.Flags().Bool("inline", false, "inline nested structs")
	yamlToStructCmd.Flags().Bool("omitempty", false, "add omitempty to all fields")
}
