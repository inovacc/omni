package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/readlink"
	"github.com/spf13/cobra"
)

// readlinkCmd represents the readlink command
var readlinkCmd = &cobra.Command{
	Use:   "readlink [OPTION]... FILE...",
	Short: "Print resolved symbolic links or canonical file names",
	Long: `Print value of a symbolic link or canonical file name.

  -f, --canonicalize            canonicalize by following every symlink
  -e, --canonicalize-existing   canonicalize, all components must exist
  -m, --canonicalize-missing    canonicalize without requirements on existence
  -n, --no-newline              do not output the trailing delimiter
  -q, --quiet                   suppress most error messages
  -s, --silent                  suppress most error messages
  -v, --verbose                 report error messages
  -z, --zero                    end each output line with NUL, not newline`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := readlink.ReadlinkOptions{}

		opts.Canonicalize, _ = cmd.Flags().GetBool("canonicalize")
		opts.CanonicalizeExisting, _ = cmd.Flags().GetBool("canonicalize-existing")
		opts.CanonicalizeMissing, _ = cmd.Flags().GetBool("canonicalize-missing")
		opts.NoNewline, _ = cmd.Flags().GetBool("no-newline")
		opts.Quiet, _ = cmd.Flags().GetBool("quiet")
		opts.Silent, _ = cmd.Flags().GetBool("silent")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		opts.Zero, _ = cmd.Flags().GetBool("zero")

		return readlink.RunReadlink(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(readlinkCmd)

	readlinkCmd.Flags().BoolP("canonicalize", "f", false, "canonicalize by following every symlink")
	readlinkCmd.Flags().BoolP("canonicalize-existing", "e", false, "canonicalize, all components must exist")
	readlinkCmd.Flags().BoolP("canonicalize-missing", "m", false, "canonicalize without requirements on existence")
	readlinkCmd.Flags().BoolP("no-newline", "n", false, "do not output the trailing delimiter")
	readlinkCmd.Flags().BoolP("quiet", "q", false, "suppress most error messages")
	readlinkCmd.Flags().BoolP("silent", "s", false, "suppress most error messages")
	readlinkCmd.Flags().BoolP("verbose", "v", false, "report error messages")
	readlinkCmd.Flags().BoolP("zero", "z", false, "end each output line with NUL, not newline")
}
