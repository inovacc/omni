package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/xmlfmt"
	"github.com/inovacc/omni/internal/cli/xmlutil"
	"github.com/spf13/cobra"
)

var xmlCmd = &cobra.Command{
	Use:   "xml [FILE]",
	Short: "XML utilities (format, validate, convert)",
	Long: `XML utilities for formatting, validation, and conversion.

When called directly, formats XML (same as 'xml fmt').

Subcommands:
  fmt         Format/beautify XML
  validate    Validate XML syntax
  tojson      Convert XML to JSON
  fromjson    Convert JSON to XML

Examples:
  omni xml file.xml
  omni xml fmt file.xml
  omni xml validate file.xml
  omni xml "<root><item>value</item></root>"
  cat file.xml | omni xml
  omni xml --minify file.xml
  omni xml tojson file.xml
  omni xml fromjson file.json`,
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

var xmlToJSONCmd = &cobra.Command{
	Use:     "tojson [FILE]",
	Aliases: []string{"json", "2json"},
	Short:   "Convert XML to JSON",
	Long: `Convert XML data to JSON format.

  --attr-prefix=STR    prefix for attributes in JSON (default "-")
  --text-key=STR       key for text content (default "#text")

Examples:
  omni xml tojson file.xml
  cat file.xml | omni xml tojson
  omni xml tojson --attr-prefix=@ file.xml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := xmlutil.FromXMLOptions{}

		opts.AttrPrefix, _ = cmd.Flags().GetString("attr-prefix")
		opts.TextKey, _ = cmd.Flags().GetString("text-key")

		return xmlutil.RunFromXML(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

var xmlFromJSONCmd = &cobra.Command{
	Use:     "fromjson [FILE]",
	Aliases: []string{"from-json", "json2xml"},
	Short:   "Convert JSON to XML",
	Long: `Convert JSON data to XML format.

  -r, --root=NAME      root element name (default "root")
  -i, --indent=STR     indentation string (default "  ")
  --item-tag=NAME      tag for array items (default "item")
  --attr-prefix=STR    prefix for attributes (default "-")

Examples:
  omni xml fromjson file.json
  echo '{"name":"John"}' | omni xml fromjson
  omni xml fromjson -r person file.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := xmlutil.ToXMLOptions{}

		opts.Root, _ = cmd.Flags().GetString("root")
		opts.Indent, _ = cmd.Flags().GetString("indent")
		opts.ItemTag, _ = cmd.Flags().GetString("item-tag")
		opts.AttrPrefix, _ = cmd.Flags().GetString("attr-prefix")

		return xmlutil.RunToXML(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(xmlCmd)
	xmlCmd.AddCommand(xmlFmtCmd)
	xmlCmd.AddCommand(xmlValidateCmd)
	xmlCmd.AddCommand(xmlToJSONCmd)
	xmlCmd.AddCommand(xmlFromJSONCmd)

	// Flags for root xml command (formatting)
	xmlCmd.Flags().BoolP("minify", "m", false, "minify XML (remove whitespace)")
	xmlCmd.Flags().StringP("indent", "i", "  ", "indentation string")

	// Flags for xml fmt subcommand
	xmlFmtCmd.Flags().BoolP("minify", "m", false, "minify XML (remove whitespace)")
	xmlFmtCmd.Flags().StringP("indent", "i", "  ", "indentation string")

	// Flags for xml validate subcommand
	xmlValidateCmd.Flags().Bool("json", false, "output as JSON")

	// Flags for xml tojson subcommand
	xmlToJSONCmd.Flags().String("attr-prefix", "-", "prefix for attributes in JSON")
	xmlToJSONCmd.Flags().String("text-key", "#text", "key for text content")

	// Flags for xml fromjson subcommand
	xmlFromJSONCmd.Flags().StringP("root", "r", "root", "root element name")
	xmlFromJSONCmd.Flags().StringP("indent", "i", "  ", "indentation string")
	xmlFromJSONCmd.Flags().String("item-tag", "item", "tag for array items")
	xmlFromJSONCmd.Flags().String("attr-prefix", "-", "prefix for attributes")
}
