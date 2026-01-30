package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/hash"
	"github.com/spf13/cobra"
)

// md5sumCmd represents the md5sum command
var md5sumCmd = &cobra.Command{
	Use:   "md5sum [OPTION]... [FILE]...",
	Short: "Compute and check MD5 message digest",
	Long: `Print or check MD5 (128-bit) checksums.

With no FILE, or when FILE is -, read standard input.

  -c, --check   read checksums from FILE and check them
  -b, --binary  read in binary mode
      --quiet   don't print OK for each verified file
      --status  don't output anything, status code shows success
  -w, --warn    warn about improperly formatted checksum lines

Note: MD5 is cryptographically broken and should not be used for security.
Use SHA256 for secure hashing.

Examples:
  omni md5sum file.txt           # compute hash
  omni md5sum -c checksums.txt   # verify checksums`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := hash.HashOptions{Algorithm: "md5"}

		opts.Check, _ = cmd.Flags().GetBool("check")
		opts.Binary, _ = cmd.Flags().GetBool("binary")
		opts.Quiet, _ = cmd.Flags().GetBool("quiet")
		opts.Status, _ = cmd.Flags().GetBool("status")
		opts.Warn, _ = cmd.Flags().GetBool("warn")

		return hash.RunMD5Sum(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(md5sumCmd)

	md5sumCmd.Flags().BoolP("check", "c", false, "read checksums from FILE and check them")
	md5sumCmd.Flags().BoolP("binary", "b", false, "read in binary mode")
	md5sumCmd.Flags().Bool("quiet", false, "don't print OK for verified files")
	md5sumCmd.Flags().Bool("status", false, "don't output anything, use status code")
	md5sumCmd.Flags().BoolP("warn", "w", false, "warn about improperly formatted lines")
}
