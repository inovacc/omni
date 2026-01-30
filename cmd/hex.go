package cmd

import (
	"github.com/inovacc/omni/internal/cli/hexenc"
	"github.com/spf13/cobra"
)

var hexCmd = &cobra.Command{
	Use:   "hex",
	Short: "Hexadecimal encoding and decoding utilities",
	Long: `Hexadecimal encoding and decoding utilities.

Subcommands:
  encode    Encode text to hexadecimal
  decode    Decode hexadecimal to text

Examples:
  omni hex encode "hello"
  omni hex decode "68656c6c6f"
  echo "test" | omni hex encode`,
}

var hexEncodeCmd = &cobra.Command{
	Use:   "encode [TEXT]",
	Short: "Encode text to hexadecimal",
	Long: `Encode text to hexadecimal representation.

Each byte is converted to its two-character hex representation.

Examples:
  omni hex encode "hello"              # Output: 68656c6c6f
  omni hex encode --upper "hello"      # Output: 68656C6C6F
  echo "test" | omni hex encode        # Read from stdin
  omni hex encode file.txt             # Read from file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := hexenc.Options{}
		opts.JSON, _ = cmd.Flags().GetBool("json")
		opts.Uppercase, _ = cmd.Flags().GetBool("upper")

		return hexenc.RunEncode(cmd.OutOrStdout(), args, opts)
	},
}

var hexDecodeCmd = &cobra.Command{
	Use:   "decode [HEX]",
	Short: "Decode hexadecimal to text",
	Long: `Decode hexadecimal string back to text.

Accepts hex strings with or without separators (spaces, colons, dashes).

Examples:
  omni hex decode "68656c6c6f"         # Output: hello
  omni hex decode "68:65:6c:6c:6f"     # With colons
  omni hex decode "68 65 6c 6c 6f"     # With spaces
  echo "74657374" | omni hex decode    # Read from stdin`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := hexenc.Options{}
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return hexenc.RunDecode(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(hexCmd)
	hexCmd.AddCommand(hexEncodeCmd)
	hexCmd.AddCommand(hexDecodeCmd)

	hexEncodeCmd.Flags().Bool("json", false, "output as JSON")
	hexEncodeCmd.Flags().BoolP("upper", "u", false, "use uppercase hex letters")

	hexDecodeCmd.Flags().Bool("json", false, "output as JSON")
}
