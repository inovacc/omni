package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli/fold"

	"github.com/spf13/cobra"
)

// foldCmd represents the fold command
var foldCmd = &cobra.Command{
	Use:   "fold [OPTION]... [FILE]...",
	Short: "Wrap each input line to fit in specified width",
	Long: `Wrap input lines in each FILE, writing to standard output.

With no FILE, or when FILE is -, read standard input.

  -w, --width=WIDTH  use WIDTH columns instead of 80
  -b, --bytes        count bytes rather than columns
  -s, --spaces       break at spaces

Examples:
  omni fold -w 40 file.txt     # wrap lines at 40 columns
  omni fold -s -w 72 README    # wrap at spaces, 72 columns`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := fold.FoldOptions{}

		opts.Width, _ = cmd.Flags().GetInt("width")
		opts.Bytes, _ = cmd.Flags().GetBool("bytes")
		opts.Spaces, _ = cmd.Flags().GetBool("spaces")

		return fold.RunFold(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(foldCmd)

	foldCmd.Flags().IntP("width", "w", 80, "use WIDTH columns instead of 80")
	foldCmd.Flags().BoolP("bytes", "b", false, "count bytes rather than columns")
	foldCmd.Flags().BoolP("spaces", "s", false, "break at spaces")
}
