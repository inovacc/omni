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
  k8s         Format YAML with Kubernetes conventions
  tostruct    Convert YAML to Go struct definition

Examples:
  omni yaml validate config.yaml
  omni yaml fmt config.yaml
  omni yaml fmt --sort-keys config.yaml
  omni yaml k8s deployment.yaml
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
Supports multi-document YAML files.

Examples:
  omni yaml fmt config.yaml
  omni yaml fmt --sort-keys config.yaml
  omni yaml fmt --remove-empty config.yaml
  omni yaml fmt -i config.yaml              # in-place edit
  cat config.yaml | omni yaml fmt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := yamlutil.FormatOptions{}
		opts.Indent, _ = cmd.Flags().GetInt("indent")
		opts.JSON, _ = cmd.Flags().GetBool("json")
		opts.SortKeys, _ = cmd.Flags().GetBool("sort-keys")
		opts.RemoveEmpty, _ = cmd.Flags().GetBool("remove-empty")
		opts.InPlace, _ = cmd.Flags().GetBool("in-place")

		return yamlutil.RunFormat(cmd.OutOrStdout(), args, opts)
	},
}

var yamlK8sCmd = &cobra.Command{
	Use:     "k8s [FILE]",
	Aliases: []string{"kubernetes", "kube"},
	Short:   "Format YAML with Kubernetes conventions",
	Long: `Format YAML with Kubernetes-specific key ordering.

Orders keys according to Kubernetes conventions:
  - Top level: apiVersion, kind, metadata, spec, status
  - Metadata: name, namespace, labels, annotations

Supports multi-document YAML files (---, multiple resources).

Examples:
  omni yaml k8s deployment.yaml
  omni yaml k8s --remove-empty deployment.yaml
  omni yaml k8s -i deployment.yaml           # in-place edit
  cat manifest.yaml | omni yaml k8s`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := yamlutil.K8sFormatOptions{}
		opts.Indent, _ = cmd.Flags().GetInt("indent")
		opts.RemoveEmpty, _ = cmd.Flags().GetBool("remove-empty")
		opts.InPlace, _ = cmd.Flags().GetBool("in-place")

		return yamlutil.RunK8sFormat(cmd.OutOrStdout(), args, opts)
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
	yamlCmd.AddCommand(yamlK8sCmd)
	yamlCmd.AddCommand(yamlToStructCmd)

	yamlValidateCmd.Flags().Bool("json", false, "output as JSON")
	yamlValidateCmd.Flags().Bool("strict", false, "fail on unknown fields")

	// fmt flags
	yamlFmtCmd.Flags().Int("indent", 2, "indentation width")
	yamlFmtCmd.Flags().Bool("json", false, "output as JSON instead of YAML")
	yamlFmtCmd.Flags().Bool("sort-keys", false, "sort keys alphabetically")
	yamlFmtCmd.Flags().Bool("remove-empty", false, "remove empty/null values")
	yamlFmtCmd.Flags().BoolP("in-place", "i", false, "modify file in place")

	// k8s flags
	yamlK8sCmd.Flags().Int("indent", 2, "indentation width")
	yamlK8sCmd.Flags().Bool("remove-empty", false, "remove empty/null values")
	yamlK8sCmd.Flags().BoolP("in-place", "i", false, "modify file in place")

	// tostruct flags
	yamlToStructCmd.Flags().StringP("name", "n", "Root", "struct name")
	yamlToStructCmd.Flags().StringP("package", "p", "main", "package name")
	yamlToStructCmd.Flags().Bool("inline", false, "inline nested structs")
	yamlToStructCmd.Flags().Bool("omitempty", false, "add omitempty to all fields")
}
