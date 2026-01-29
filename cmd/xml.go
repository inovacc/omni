package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/xmlfmt"
	"github.com/spf13/cobra"
)

var xmlCmd = &cobra.Command{
	Use:   "xml [FILE]",
	Short: "XML utilities (format, validate)",
	Long: `XML utilities for formatting and validation.

When called directly, formats XML (same as 'xml fmt').

Subcommands:
  fmt         Format/beautify XML
  validate    Validate XML syntax

Examples:
  omni xml file.xml
  omni xml fmt file.xml
  omni xml validate file.xml
  omni xml "<root><item>value</item></root>"
  cat file.xml | omni xml
  omni xml --minify file.xml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default behavior is formatting
		opts := xmlfmt.Options{}
		opts.Minify, _ = cmd.Flags().GetBool("minify")
		opts.Indent, _ = cmd.Flags().GetString("indent")

		return xmlfmt.Run(os.Stdout, args, opts)
	},
}

var xmlFmtCmd = &cobra.Command{
	Use:   "fmt [FILE]",
	Short: "Format XML",
	Long: `Format and beautify XML.

Reads XML from a file, argument, or stdin and outputs formatted XML
with proper indentation. Use --minify to remove whitespace.

Examples:
  omni xml fmt file.xml
  omni xml fmt "<root><item>value</item></root>"
  cat file.xml | omni xml fmt
  omni xml fmt --minify file.xml
  omni xml fmt --indent "    " file.xml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := xmlfmt.Options{}
		opts.Minify, _ = cmd.Flags().GetBool("minify")
		opts.Indent, _ = cmd.Flags().GetString("indent")

		return xmlfmt.Run(os.Stdout, args, opts)
	},
}

var xmlValidateCmd = &cobra.Command{
	Use:   "validate [FILE...]",
	Short: "Validate XML syntax",
	Long: `Validate XML syntax for one or more files.

Checks that the input is well-formed XML.

Examples:
  omni xml validate file.xml
  omni xml validate *.xml
  cat file.xml | omni xml validate
  omni xml validate --json file.xml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := xmlfmt.ValidateOptions{}
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return xmlfmt.RunValidate(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(xmlCmd)
	xmlCmd.AddCommand(xmlFmtCmd)
	xmlCmd.AddCommand(xmlValidateCmd)

	// Flags for root xml command (formatting)
	xmlCmd.Flags().BoolP("minify", "m", false, "minify XML (remove whitespace)")
	xmlCmd.Flags().StringP("indent", "i", "  ", "indentation string")

	// Flags for xml fmt subcommand
	xmlFmtCmd.Flags().BoolP("minify", "m", false, "minify XML (remove whitespace)")
	xmlFmtCmd.Flags().StringP("indent", "i", "  ", "indentation string")

	// Flags for xml validate subcommand
	xmlValidateCmd.Flags().Bool("json", false, "output as JSON")
}
