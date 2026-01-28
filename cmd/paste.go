package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli/paste"

	"github.com/spf13/cobra"
)

// pasteCmd represents the paste command
var pasteCmd = &cobra.Command{
	Use:   "paste [OPTION]... [FILE]...",
	Short: "Merge lines of files",
	Long: `Write lines consisting of the sequentially corresponding lines from
each FILE, separated by TABs, to standard output.

With no FILE, or when FILE is -, read standard input.

  -d, --delimiters=LIST   reuse characters from LIST instead of TABs
  -s, --serial            paste one file at a time instead of in parallel
  -z, --zero-terminated   line delimiter is NUL, not newline`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := paste.PasteOptions{}

		opts.Delimiters, _ = cmd.Flags().GetString("delimiters")
		opts.Serial, _ = cmd.Flags().GetBool("serial")
		opts.Zero, _ = cmd.Flags().GetBool("zero-terminated")

		return paste.RunPaste(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(pasteCmd)

	pasteCmd.Flags().StringP("delimiters", "d", "", "reuse characters from LIST instead of TABs")
	pasteCmd.Flags().BoolP("serial", "s", false, "paste one file at a time instead of in parallel")
	pasteCmd.Flags().BoolP("zero-terminated", "z", false, "line delimiter is NUL, not newline")
}
