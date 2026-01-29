package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/urlenc"
	"github.com/spf13/cobra"
)

var urlCmd = &cobra.Command{
	Use:   "url",
	Short: "URL encoding and decoding utilities",
	Long: `URL encoding and decoding utilities.

Subcommands:
  encode    URL encode text
  decode    URL decode text

Examples:
  omni url encode "hello world"
  omni url decode "hello%20world"
  echo "test string" | omni url encode
  omni url encode --component "a=b&c=d"`,
}

var urlEncodeCmd = &cobra.Command{
	Use:   "encode [TEXT]",
	Short: "URL encode text",
	Long: `URL encode text for safe use in URLs.

By default uses path encoding. Use --component for query string encoding
which is more aggressive (encodes more characters).

Examples:
  omni url encode "hello world"           # Output: hello%20world
  omni url encode --component "a=b&c=d"   # Output: a%3Db%26c%3Dd
  echo "test" | omni url encode           # Read from stdin
  omni url encode file.txt                # Read from file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := urlenc.Options{}
		opts.Component, _ = cmd.Flags().GetBool("component")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return urlenc.RunEncode(os.Stdout, args, opts)
	},
}

var urlDecodeCmd = &cobra.Command{
	Use:   "decode [TEXT]",
	Short: "URL decode text",
	Long: `URL decode percent-encoded text.

By default uses path decoding. Use --component for query string decoding.

Examples:
  omni url decode "hello%20world"         # Output: hello world
  omni url decode --component "a%3Db"     # Output: a=b
  echo "test%20string" | omni url decode  # Read from stdin`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := urlenc.Options{}
		opts.Component, _ = cmd.Flags().GetBool("component")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return urlenc.RunDecode(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(urlCmd)
	urlCmd.AddCommand(urlEncodeCmd)
	urlCmd.AddCommand(urlDecodeCmd)

	urlEncodeCmd.Flags().BoolP("component", "c", false, "use query component encoding (more aggressive)")
	urlEncodeCmd.Flags().Bool("json", false, "output as JSON")

	urlDecodeCmd.Flags().BoolP("component", "c", false, "use query component decoding")
	urlDecodeCmd.Flags().Bool("json", false, "output as JSON")
}
