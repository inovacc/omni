package cmd

import (
	"github.com/inovacc/omni/internal/cli/bzip2"
	"github.com/spf13/cobra"
)

var (
	bzip2Decompress bool
	bzip2Keep       bool
	bzip2Force      bool
	bzip2Stdout     bool
	bzip2Verbose    bool
)

var bzip2Cmd = &cobra.Command{
	Use:   "bzip2 [OPTION]... [FILE]...",
	Short: "Decompress bzip2 files",
	Long: `Decompress FILEs using bzip2 format.

Note: Only decompression is supported (Go stdlib limitation).

  -d, --decompress   decompress (required)
  -k, --keep         keep original files
  -f, --force        force overwrite
  -c, --stdout       write to stdout
  -v, --verbose      verbose mode

Examples:
  omni bzip2 -d file.txt.bz2   # decompress
  omni bzip2 -dk file.txt.bz2  # decompress, keep original`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := bzip2.Bzip2Options{
			Decompress: bzip2Decompress,
			Keep:       bzip2Keep,
			Force:      bzip2Force,
			Stdout:     bzip2Stdout,
			Verbose:    bzip2Verbose,
		}

		return bzip2.RunBzip2(cmd.OutOrStdout(), args, opts)
	},
}

var bunzip2Cmd = &cobra.Command{
	Use:   "bunzip2 [OPTION]... [FILE]...",
	Short: "Decompress bzip2 files",
	Long: `Decompress FILEs in bzip2 format.

Equivalent to bzip2 -d.

Examples:
  omni bunzip2 file.txt.bz2    # decompress
  omni bunzip2 -k file.txt.bz2 # keep original`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := bzip2.Bzip2Options{
			Decompress: true,
			Keep:       bzip2Keep,
			Force:      bzip2Force,
			Stdout:     bzip2Stdout,
			Verbose:    bzip2Verbose,
		}

		return bzip2.RunBunzip2(cmd.OutOrStdout(), args, opts)
	},
}

var bzcatCmd = &cobra.Command{
	Use:   "bzcat [FILE]...",
	Short: "Decompress and print bzip2 files",
	Long: `Decompress and print FILEs to stdout.

Equivalent to bzip2 -dc.

Examples:
  omni bzcat file.txt.bz2      # print decompressed content`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return bzip2.RunBzcat(cmd.OutOrStdout(), args)
	},
}

func init() {
	rootCmd.AddCommand(bzip2Cmd)
	rootCmd.AddCommand(bunzip2Cmd)
	rootCmd.AddCommand(bzcatCmd)

	bzip2Cmd.Flags().BoolVarP(&bzip2Decompress, "decompress", "d", false, "decompress")
	bzip2Cmd.Flags().BoolVarP(&bzip2Keep, "keep", "k", false, "keep original files")
	bzip2Cmd.Flags().BoolVarP(&bzip2Force, "force", "f", false, "force overwrite")
	bzip2Cmd.Flags().BoolVarP(&bzip2Stdout, "stdout", "c", false, "write to stdout")
	bzip2Cmd.Flags().BoolVarP(&bzip2Verbose, "verbose", "v", false, "verbose mode")

	bunzip2Cmd.Flags().BoolVarP(&bzip2Keep, "keep", "k", false, "keep original files")
	bunzip2Cmd.Flags().BoolVarP(&bzip2Force, "force", "f", false, "force overwrite")
	bunzip2Cmd.Flags().BoolVarP(&bzip2Stdout, "stdout", "c", false, "write to stdout")
	bunzip2Cmd.Flags().BoolVarP(&bzip2Verbose, "verbose", "v", false, "verbose mode")
}
