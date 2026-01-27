package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli"

	"github.com/spf13/cobra"
)

// base64Cmd represents the base64 command
var base64Cmd = &cobra.Command{
	Use:   "base64 [OPTION]... [FILE]",
	Short: "Base64 encode or decode data",
	Long: `Base64 encode or decode FILE, or standard input, to standard output.

With no FILE, or when FILE is -, read standard input.

  -d, --decode    decode data
  -w, --wrap=N    wrap encoded lines after N characters (default 76, 0 = no wrap)
  -i, --ignore-garbage  ignore non-alphabet characters when decoding

Examples:
  echo "hello" | omni base64           # encode
  echo "aGVsbG8K" | omni base64 -d     # decode
  omni base64 file.bin                 # encode file
  omni base64 -d encoded.txt           # decode file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.BaseOptions{}

		opts.Decode, _ = cmd.Flags().GetBool("decode")
		opts.Wrap, _ = cmd.Flags().GetInt("wrap")
		opts.IgnoreGarbage, _ = cmd.Flags().GetBool("ignore-garbage")

		return cli.RunBase64(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(base64Cmd)

	base64Cmd.Flags().BoolP("decode", "d", false, "decode data")
	base64Cmd.Flags().IntP("wrap", "w", 76, "wrap encoded lines after N characters (0 = no wrap)")
	base64Cmd.Flags().BoolP("ignore-garbage", "i", false, "ignore non-alphabet characters when decoding")
}
