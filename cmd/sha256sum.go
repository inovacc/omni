package cmd

import (
	"github.com/inovacc/omni/internal/cli/hash"
	"github.com/spf13/cobra"
)

// sha256sumCmd represents the sha256sum command
var sha256sumCmd = &cobra.Command{
	Use:   "sha256sum [OPTION]... [FILE]...",
	Short: "Compute and check SHA256 message digest",
	Long: `Print or check SHA256 (256-bit) checksums.

With no FILE, or when FILE is -, read standard input.

  -c, --check   read checksums from FILE and check them
  -b, --binary  read in binary mode
      --quiet   don't print OK for each verified file
      --status  don't output anything, status code shows success
  -w, --warn    warn about improperly formatted checksum lines

Examples:
  omni sha256sum file.txt           # compute hash
  omni sha256sum -c checksums.txt   # verify checksums`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := hash.HashOptions{Algorithm: "sha256"}

		opts.Check, _ = cmd.Flags().GetBool("check")
		opts.Binary, _ = cmd.Flags().GetBool("binary")
		opts.Quiet, _ = cmd.Flags().GetBool("quiet")
		opts.Status, _ = cmd.Flags().GetBool("status")
		opts.Warn, _ = cmd.Flags().GetBool("warn")

		return hash.RunSHA256Sum(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(sha256sumCmd)

	sha256sumCmd.Flags().BoolP("check", "c", false, "read checksums from FILE and check them")
	sha256sumCmd.Flags().BoolP("binary", "b", false, "read in binary mode")
	sha256sumCmd.Flags().Bool("quiet", false, "don't print OK for verified files")
	sha256sumCmd.Flags().Bool("status", false, "don't output anything, use status code")
	sha256sumCmd.Flags().BoolP("warn", "w", false, "warn about improperly formatted lines")
}
