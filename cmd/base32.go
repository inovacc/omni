package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli"

	"github.com/spf13/cobra"
)

// base32Cmd represents the base32 command
var base32Cmd = &cobra.Command{
	Use:   "base32 [OPTION]... [FILE]",
	Short: "Base32 encode or decode data",
	Long: `Base32 encode or decode FILE, or standard input, to standard output.

With no FILE, or when FILE is -, read standard input.

  -d, --decode    decode data
  -w, --wrap=N    wrap encoded lines after N characters (default 76, 0 = no wrap)

Examples:
  echo "hello" | omni base32           # encode
  echo "NBSWY3DP" | omni base32 -d     # decode
  omni base32 file.bin                 # encode file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.BaseOptions{}

		opts.Decode, _ = cmd.Flags().GetBool("decode")
		opts.Wrap, _ = cmd.Flags().GetInt("wrap")

		return cli.RunBase32(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(base32Cmd)

	base32Cmd.Flags().BoolP("decode", "d", false, "decode data")
	base32Cmd.Flags().IntP("wrap", "w", 76, "wrap encoded lines after N characters (0 = no wrap)")
}
