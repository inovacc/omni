package cmd

import (
	"github.com/inovacc/omni/internal/cli/hmac"
	"github.com/spf13/cobra"
)

// hmacCmd represents the hmac command
var hmacCmd = &cobra.Command{
	Use:   "hmac [OPTION]... [MESSAGE]",
	Short: "Compute a keyed-hash message authentication code (HMAC)",
	Long: `Compute the HMAC of a message using a shared secret key.

With no MESSAGE argument, read the message from standard input.

  -a, --algorithm ALG  hash algorithm: sha256 (default), sha1, sha512
  -k, --key KEY        shared secret key (required)

Examples:
  omni hmac -k secret "the message"      # HMAC-SHA256 of "the message"
  omni hmac -a sha512 -k secret msg      # HMAC-SHA512
  echo -n "the message" | omni hmac -k secret`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := hmac.HMACOptions{}

		opts.Algorithm, _ = cmd.Flags().GetString("algorithm")
		opts.Key, _ = cmd.Flags().GetString("key")

		return hmac.RunHMAC(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(hmacCmd)

	hmacCmd.Flags().StringP("algorithm", "a", "sha256", "hash algorithm (sha256, sha1, sha512)")
	hmacCmd.Flags().StringP("key", "k", "", "shared secret key (required)")
}
