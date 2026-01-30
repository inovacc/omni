package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/xz"
	"github.com/spf13/cobra"
)

var (
	xzDecompress bool
	xzKeep       bool
	xzForce      bool
	xzStdout     bool
	xzVerbose    bool
	xzList       bool
)

var xzCmd = &cobra.Command{
	Use:   "xz [OPTION]... [FILE]...",
	Short: "Compress or decompress xz files",
	Long: `Compress or decompress FILEs using xz format.

Note: Full xz support requires external library. Basic info/listing is supported.

  -d, --decompress   decompress
  -k, --keep         keep original files
  -f, --force        force overwrite
  -c, --stdout       write to stdout
  -v, --verbose      verbose mode
  -l, --list         list compressed file info

Examples:
  omni xz -l file.xz           # list info
  omni xz -d file.txt.xz       # decompress (requires external lib)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := xz.XzOptions{
			Decompress: xzDecompress,
			Keep:       xzKeep,
			Force:      xzForce,
			Stdout:     xzStdout,
			Verbose:    xzVerbose,
			List:       xzList,
		}

		return xz.RunXz(cmd.OutOrStdout(), args, opts)
	},
}

var unxzCmd = &cobra.Command{
	Use:   "unxz [OPTION]... [FILE]...",
	Short: "Decompress xz files",
	Long: `Decompress FILEs in xz format.

Equivalent to xz -d.

Note: Full decompression requires external library.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := xz.XzOptions{
			Decompress: true,
			Keep:       xzKeep,
			Force:      xzForce,
			Stdout:     xzStdout,
			Verbose:    xzVerbose,
		}

		return xz.RunUnxz(cmd.OutOrStdout(), args, opts)
	},
}

var xzcatCmd = &cobra.Command{
	Use:   "xzcat [FILE]...",
	Short: "Decompress and print xz files",
	Long: `Decompress and print FILEs to stdout.

Equivalent to xz -dc.

Note: Full decompression requires external library.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return xz.RunXzcat(cmd.OutOrStdout(), args)
	},
}

func init() {
	rootCmd.AddCommand(xzCmd)
	rootCmd.AddCommand(unxzCmd)
	rootCmd.AddCommand(xzcatCmd)

	xzCmd.Flags().BoolVarP(&xzDecompress, "decompress", "d", false, "decompress")
	xzCmd.Flags().BoolVarP(&xzKeep, "keep", "k", false, "keep original files")
	xzCmd.Flags().BoolVarP(&xzForce, "force", "f", false, "force overwrite")
	xzCmd.Flags().BoolVarP(&xzStdout, "stdout", "c", false, "write to stdout")
	xzCmd.Flags().BoolVarP(&xzVerbose, "verbose", "v", false, "verbose mode")
	xzCmd.Flags().BoolVarP(&xzList, "list", "l", false, "list compressed file info")

	unxzCmd.Flags().BoolVarP(&xzKeep, "keep", "k", false, "keep original files")
	unxzCmd.Flags().BoolVarP(&xzForce, "force", "f", false, "force overwrite")
	unxzCmd.Flags().BoolVarP(&xzStdout, "stdout", "c", false, "write to stdout")
	unxzCmd.Flags().BoolVarP(&xzVerbose, "verbose", "v", false, "verbose mode")
}
