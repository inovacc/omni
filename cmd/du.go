package cmd

import (
	"github.com/inovacc/omni/internal/cli/du"
	"github.com/spf13/cobra"
)

// duCmd represents the du command
var duCmd = &cobra.Command{
	Use:   "du [OPTION]... [FILE]...",
	Short: "Estimate file space usage",
	Long: `Summarize disk usage of each FILE, recursively for directories.

  -a, --all             write counts for all files, not just directories
  -b, --bytes           equivalent to --apparent-size --block-size=1
  -c, --total           produce a grand total
  -h, --human-readable  print sizes in human readable format (e.g., 1K 234M 2G)
  -s, --summarize       display only a total for each argument
  -d, --max-depth=N     print the total for a directory only if it is N or fewer
                        levels below the command line argument
  -x, --one-file-system skip directories on different file systems
      --apparent-size   print apparent sizes, rather than disk usage
  -0, --null            end each output line with NUL, not newline
  -B, --block-size=SIZE scale sizes by SIZE before printing them`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := du.DUOptions{}

		opts.All, _ = cmd.Flags().GetBool("all")
		opts.ByteCount, _ = cmd.Flags().GetBool("bytes")
		opts.Total, _ = cmd.Flags().GetBool("total")
		opts.HumanReadable, _ = cmd.Flags().GetBool("human-readable")
		opts.SummarizeOnly, _ = cmd.Flags().GetBool("summarize")
		opts.MaxDepth, _ = cmd.Flags().GetInt("max-depth")
		opts.OneFileSystem, _ = cmd.Flags().GetBool("one-file-system")
		opts.ApparentSize, _ = cmd.Flags().GetBool("apparent-size")
		opts.NullTerminator, _ = cmd.Flags().GetBool("null")
		opts.BlockSize, _ = cmd.Flags().GetInt64("block-size")
		opts.OutputFormat = getOutputOpts(cmd).GetFormat()

		return du.RunDU(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(duCmd)

	duCmd.Flags().BoolP("all", "a", false, "write counts for all files, not just directories")
	duCmd.Flags().BoolP("bytes", "b", false, "equivalent to --apparent-size --block-size=1")
	duCmd.Flags().BoolP("total", "c", false, "produce a grand total")
	duCmd.Flags().BoolP("human-readable", "H", false, "print sizes in human readable format")
	duCmd.Flags().BoolP("summarize", "s", false, "display only a total for each argument")
	duCmd.Flags().IntP("max-depth", "d", 0, "print total for directory only if N or fewer levels deep")
	duCmd.Flags().BoolP("one-file-system", "x", false, "skip directories on different file systems")
	duCmd.Flags().Bool("apparent-size", false, "print apparent sizes, rather than disk usage")
	duCmd.Flags().BoolP("null", "0", false, "end each output line with NUL, not newline")
	duCmd.Flags().Int64P("block-size", "B", 0, "scale sizes by SIZE before printing them")

}
