package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/base"
	"github.com/spf13/cobra"
)

// base58Cmd represents the base58 command
var base58Cmd = &cobra.Command{
	Use:   "base58 [OPTION]... [FILE]",
	Short: "Base58 encode or decode data (Bitcoin alphabet)",
	Long: `Base58 encode or decode FILE, or standard input, to standard output.

Uses Bitcoin/IPFS alphabet: 123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz

With no FILE, or when FILE is -, read standard input.

  -d, --decode    decode data

Examples:
  echo "hello" | omni base58           # encode
  omni base58 -d encoded.txt           # decode`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := base.BaseOptions{}

		opts.Decode, _ = cmd.Flags().GetBool("decode")

		return base.RunBase58(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(base58Cmd)

	base58Cmd.Flags().BoolP("decode", "d", false, "decode data")
}
