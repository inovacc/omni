package cmd

import (
	"github.com/inovacc/omni/internal/cli/hash"
	"github.com/spf13/cobra"
)

// crc64sumCmd represents the crc64sum command
var crc64sumCmd = &cobra.Command{
	Use:   "crc64sum [OPTION]... [FILE]...",
	Short: "Compute and check CRC64 checksums",
	Long: `Print or check CRC64 (ECMA polynomial) checksums.

With no FILE, or when FILE is -, read standard input.

  -c, --check   read checksums from FILE and check them
  -b, --binary  read in binary mode
      --quiet   don't print OK for each verified file
      --status  don't output anything, status code shows success
  -w, --warn    warn about improperly formatted checksum lines

Examples:
  omni crc64sum file.txt           # compute CRC64
  omni crc64sum -c checksums.txt   # verify checksums
  omni crc64sum file1 file2        # hash multiple files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := hash.HashOptions{Algorithm: "crc64"}

		opts.Check, _ = cmd.Flags().GetBool("check")
		opts.Binary, _ = cmd.Flags().GetBool("binary")
		opts.Quiet, _ = cmd.Flags().GetBool("quiet")
		opts.Status, _ = cmd.Flags().GetBool("status")
		opts.Warn, _ = cmd.Flags().GetBool("warn")

		return hash.RunCRC64Sum(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(crc64sumCmd)

	crc64sumCmd.Flags().BoolP("check", "c", false, "read checksums from FILE and check them")
	crc64sumCmd.Flags().BoolP("binary", "b", false, "read in binary mode")
	crc64sumCmd.Flags().Bool("quiet", false, "don't print OK for verified files")
	crc64sumCmd.Flags().Bool("status", false, "don't output anything, use status code")
	crc64sumCmd.Flags().BoolP("warn", "w", false, "warn about improperly formatted lines")
}
