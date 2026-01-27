package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli"

	"github.com/spf13/cobra"
)

// hashCmd represents the hash command
var hashCmd = &cobra.Command{
	Use:   "hash [OPTION]... [FILE]...",
	Short: "Compute and check file hashes",
	Long: `Print or check cryptographic hashes (checksums).

With no FILE, or when FILE is -, read standard input.

  -a, --algorithm ALG  hash algorithm: md5, sha1, sha256 (default), sha512
  -c, --check          read checksums from FILE and check them
  -b, --binary         read in binary mode
  -r, --recursive      hash files recursively in directories
      --quiet          don't print OK for each verified file
      --status         don't output anything, status code shows success
  -w, --warn           warn about improperly formatted checksum lines

Examples:
  omni hash file.txt                    # SHA256 hash
  omni hash -a md5 file.txt             # MD5 hash
  omni hash -r ./dir                    # hash all files in directory
  omni hash -c checksums.txt            # verify checksums
  omni hash file1 file2 > checksums.txt # create checksum file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.HashOptions{}

		opts.Algorithm, _ = cmd.Flags().GetString("algorithm")
		opts.Check, _ = cmd.Flags().GetBool("check")
		opts.Binary, _ = cmd.Flags().GetBool("binary")
		opts.Recursive, _ = cmd.Flags().GetBool("recursive")
		opts.Quiet, _ = cmd.Flags().GetBool("quiet")
		opts.Status, _ = cmd.Flags().GetBool("status")
		opts.Warn, _ = cmd.Flags().GetBool("warn")

		return cli.RunHash(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(hashCmd)

	hashCmd.Flags().StringP("algorithm", "a", "sha256", "hash algorithm (md5, sha1, sha256, sha512)")
	hashCmd.Flags().BoolP("check", "c", false, "read checksums from FILE and check them")
	hashCmd.Flags().BoolP("binary", "b", false, "read in binary mode")
	hashCmd.Flags().BoolP("recursive", "r", false, "hash files recursively")
	hashCmd.Flags().Bool("quiet", false, "don't print OK for verified files")
	hashCmd.Flags().Bool("status", false, "don't output anything, use status code")
	hashCmd.Flags().BoolP("warn", "w", false, "warn about improperly formatted lines")
}
