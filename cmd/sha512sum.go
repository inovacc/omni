package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/hash"
	"github.com/spf13/cobra"
)

// sha512sumCmd represents the sha512sum command
var sha512sumCmd = &cobra.Command{
	Use:   "sha512sum [OPTION]... [FILE]...",
	Short: "Compute and check SHA512 message digest",
	Long: `Print or check SHA512 (512-bit) checksums.

With no FILE, or when FILE is -, read standard input.

  -c, --check   read checksums from FILE and check them
  -b, --binary  read in binary mode
      --quiet   don't print OK for each verified file
      --status  don't output anything, status code shows success
  -w, --warn    warn about improperly formatted checksum lines

Examples:
  omni sha512sum file.txt           # compute hash
  omni sha512sum -c checksums.txt   # verify checksums`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := hash.HashOptions{Algorithm: "sha512"}

		opts.Check, _ = cmd.Flags().GetBool("check")
		opts.Binary, _ = cmd.Flags().GetBool("binary")
		opts.Quiet, _ = cmd.Flags().GetBool("quiet")
		opts.Status, _ = cmd.Flags().GetBool("status")
		opts.Warn, _ = cmd.Flags().GetBool("warn")

		return hash.RunSHA512Sum(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(sha512sumCmd)

	sha512sumCmd.Flags().BoolP("check", "c", false, "read checksums from FILE and check them")
	sha512sumCmd.Flags().BoolP("binary", "b", false, "read in binary mode")
	sha512sumCmd.Flags().Bool("quiet", false, "don't print OK for verified files")
	sha512sumCmd.Flags().Bool("status", false, "don't output anything, use status code")
	sha512sumCmd.Flags().BoolP("warn", "w", false, "warn about improperly formatted lines")
}
