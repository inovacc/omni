package cmd

import (
	"github.com/inovacc/omni/internal/cli/hash"
	"github.com/spf13/cobra"
)

// crc32sumCmd represents the crc32sum command
var crc32sumCmd = &cobra.Command{
	Use:   "crc32sum [OPTION]... [FILE]...",
	Short: "Compute and check CRC32 checksums",
	Long: `Print or check CRC32 (IEEE polynomial) checksums.

With no FILE, or when FILE is -, read standard input.

  -c, --check   read checksums from FILE and check them
  -b, --binary  read in binary mode
      --quiet   don't print OK for each verified file
      --status  don't output anything, status code shows success
  -w, --warn    warn about improperly formatted checksum lines

Examples:
  omni crc32sum file.txt           # compute CRC32
  omni crc32sum -c checksums.txt   # verify checksums
  omni crc32sum file1 file2        # hash multiple files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := hash.HashOptions{Algorithm: "crc32"}

		opts.Check, _ = cmd.Flags().GetBool("check")
		opts.Binary, _ = cmd.Flags().GetBool("binary")
		opts.Quiet, _ = cmd.Flags().GetBool("quiet")
		opts.Status, _ = cmd.Flags().GetBool("status")
		opts.Warn, _ = cmd.Flags().GetBool("warn")

		return hash.RunCRC32Sum(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(crc32sumCmd)

	crc32sumCmd.Flags().BoolP("check", "c", false, "read checksums from FILE and check them")
	crc32sumCmd.Flags().BoolP("binary", "b", false, "read in binary mode")
	crc32sumCmd.Flags().Bool("quiet", false, "don't print OK for verified files")
	crc32sumCmd.Flags().Bool("status", false, "don't output anything, use status code")
	crc32sumCmd.Flags().BoolP("warn", "w", false, "warn about improperly formatted lines")
}
