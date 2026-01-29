package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/htmlenc"
	"github.com/spf13/cobra"
)

var htmlCmd = &cobra.Command{
	Use:   "html",
	Short: "HTML encoding and decoding utilities",
	Long: `HTML encoding and decoding utilities.

Subcommands:
  encode    HTML encode text (escape special characters)
  decode    HTML decode text (unescape entities)

Examples:
  omni html encode "<script>alert('xss')</script>"
  omni html decode "&lt;div&gt;content&lt;/div&gt;"
  echo "<p>hello</p>" | omni html encode`,
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

		return htmlenc.RunEncode(os.Stdout, args, opts)
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

		return htmlenc.RunDecode(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(htmlCmd)
	htmlCmd.AddCommand(htmlEncodeCmd)
	htmlCmd.AddCommand(htmlDecodeCmd)

	htmlEncodeCmd.Flags().Bool("json", false, "output as JSON")
	htmlDecodeCmd.Flags().Bool("json", false, "output as JSON")
}
