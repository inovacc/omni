package cmd

import (
	"github.com/inovacc/omni/internal/cli/pager"
	"github.com/spf13/cobra"
)

// moreCmd represents the more command
var moreCmd = &cobra.Command{
	Use:   "more [OPTION]... [FILE]",
	Short: "View file contents page by page",
	Long: `Display file contents one screen at a time.

more is a simpler pager than less - it's designed to show content
and quit when reaching the end.

Navigation:
  Space, Enter    Scroll down one page
  q               Quit

Examples:
  omni more file.txt
  omni more -n file.txt     # with line numbers
  cat file.txt | omni more  # from stdin`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := pager.PagerOptions{}

		opts.LineNumbers, _ = cmd.Flags().GetBool("line-numbers")
		opts.Quit = true // more traditionally quits at end

		return pager.RunMore(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(moreCmd)

	moreCmd.Flags().BoolP("line-numbers", "n", false, "show line numbers")
}
