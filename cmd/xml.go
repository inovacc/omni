package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/xmlfmt"
	"github.com/spf13/cobra"
)

var xmlCmd = &cobra.Command{
	Use:   "xml [FILE]",
	Short: "Format and pretty-print XML",
	Long: `Format and pretty-print XML.

Reads XML from a file, argument, or stdin and outputs formatted XML
with proper indentation. Use --minify to remove whitespace.

Examples:
  omni xml file.xml
  omni xml "<root><item>value</item></root>"
  cat file.xml | omni xml
  omni xml --minify file.xml
  omni xml --indent "    " file.xml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := xmlfmt.Options{}
		opts.Minify, _ = cmd.Flags().GetBool("minify")
		opts.Indent, _ = cmd.Flags().GetString("indent")

		return xmlfmt.Run(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(xmlCmd)

	xmlCmd.Flags().BoolP("minify", "m", false, "minify XML (remove whitespace)")
	xmlCmd.Flags().StringP("indent", "i", "  ", "indentation string")
}
