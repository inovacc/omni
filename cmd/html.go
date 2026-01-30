package cmd

import (
	"github.com/inovacc/omni/internal/cli/htmlenc"
	"github.com/inovacc/omni/internal/cli/htmlfmt"
	"github.com/spf13/cobra"
)

var htmlCmd = &cobra.Command{
	Use:   "html",
	Short: "HTML utilities (format, encode, decode)",
	Long: `HTML utilities for formatting, encoding, and decoding.

Subcommands:
  fmt       Format/beautify HTML
  minify    Minify HTML
  validate  Validate HTML syntax
  encode    HTML encode text (escape special characters)
  decode    HTML decode text (unescape entities)

Examples:
  omni html fmt file.html
  omni html minify file.html
  omni html validate file.html
  omni html encode "<script>alert('xss')</script>"
  omni html decode "&lt;div&gt;content&lt;/div&gt;"`,
}

var htmlEncodeCmd = &cobra.Command{
	Use:   "encode [TEXT]",
	Short: "HTML encode text",
	Long: `HTML encode text by escaping special characters.

Converts characters like <, >, &, ", and ' to their HTML entity equivalents.

Examples:
  omni html encode "<script>alert('xss')</script>"
  omni html encode "Tom & Jerry"
  echo "<div>" | omni html encode`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := htmlenc.Options{}
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return htmlenc.RunEncode(cmd.OutOrStdout(), args, opts)
	},
}

var htmlDecodeCmd = &cobra.Command{
	Use:   "decode [TEXT]",
	Short: "HTML decode text",
	Long: `HTML decode text by unescaping HTML entities.

Converts HTML entities like &lt;, &gt;, &amp;, &quot; back to their original characters.

Examples:
  omni html decode "&lt;script&gt;"
  omni html decode "Tom &amp; Jerry"
  echo "&lt;div&gt;" | omni html decode`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := htmlenc.Options{}
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return htmlenc.RunDecode(cmd.OutOrStdout(), args, opts)
	},
}

var htmlFmtCmd = &cobra.Command{
	Use:     "fmt [FILE]",
	Aliases: []string{"format", "beautify"},
	Short:   "Format/beautify HTML",
	Long: `Format HTML with proper indentation.

  -i, --indent=STR     indentation string (default "  ")
  --sort-attrs         sort attributes alphabetically

Examples:
  omni html fmt file.html
  omni html fmt "<div><p>text</p></div>"
  cat file.html | omni html fmt
  omni html fmt --sort-attrs file.html`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := htmlfmt.Options{}
		opts.Indent, _ = cmd.Flags().GetString("indent")
		opts.SortAttrs, _ = cmd.Flags().GetBool("sort-attrs")

		return htmlfmt.Run(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

var htmlMinifyCmd = &cobra.Command{
	Use:     "minify [FILE]",
	Aliases: []string{"min", "compact"},
	Short:   "Minify HTML",
	Long: `Minify HTML by removing unnecessary whitespace and comments.

Examples:
  omni html minify file.html
  cat file.html | omni html minify`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return htmlfmt.RunMinify(cmd.OutOrStdout(), cmd.InOrStdin(), args, htmlfmt.Options{})
	},
}

var htmlValidateCmd = &cobra.Command{
	Use:     "validate [FILE]",
	Aliases: []string{"check", "lint"},
	Short:   "Validate HTML syntax",
	Long: `Validate HTML syntax.

Exit codes:
  0  Valid HTML
  1  Invalid HTML or error

  --json    output result as JSON

Examples:
  omni html validate file.html
  omni html validate "<div><p>text</p></div>"
  omni html validate --json file.html`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := htmlfmt.ValidateOptions{}
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return htmlfmt.RunValidate(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(htmlCmd)
	htmlCmd.AddCommand(htmlEncodeCmd)
	htmlCmd.AddCommand(htmlDecodeCmd)
	htmlCmd.AddCommand(htmlFmtCmd)
	htmlCmd.AddCommand(htmlMinifyCmd)
	htmlCmd.AddCommand(htmlValidateCmd)

	htmlEncodeCmd.Flags().Bool("json", false, "output as JSON")
	htmlDecodeCmd.Flags().Bool("json", false, "output as JSON")

	// html fmt flags
	htmlFmtCmd.Flags().StringP("indent", "i", "  ", "indentation string")
	htmlFmtCmd.Flags().Bool("sort-attrs", false, "sort attributes alphabetically")

	// html validate flags
	htmlValidateCmd.Flags().Bool("json", false, "output as JSON")
}
