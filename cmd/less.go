package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/pager"
	"github.com/spf13/cobra"
)

// lessCmd represents the less command
var lessCmd = &cobra.Command{
	Use:   "less [OPTION]... [FILE]",
	Short: "View file contents with scrolling",
	Long: `Display file contents one screen at a time with scrolling support.

Navigation:
  j, Down, Enter  Scroll down one line
  k, Up           Scroll up one line
  Space, PgDn     Scroll down one page
  PgUp            Scroll up one page
  g, Home         Go to beginning
  G, End          Go to end
  /               Search forward
  n               Next search match
  N               Previous search match
  h               Show help
  q               Quit

Examples:
  omni less file.txt
  omni less -N file.txt     # with line numbers
  omni less -S file.txt     # chop long lines
  cat file.txt | omni less  # from stdin`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := pager.PagerOptions{}

		opts.LineNumbers, _ = cmd.Flags().GetBool("LINE-NUMBERS")
		opts.NoInit, _ = cmd.Flags().GetBool("no-init")
		opts.Quit, _ = cmd.Flags().GetBool("quit-if-one-screen")
		opts.IgnoreCase, _ = cmd.Flags().GetBool("ignore-case")
		opts.Chop, _ = cmd.Flags().GetBool("chop-long-lines")
		opts.Raw, _ = cmd.Flags().GetBool("raw-control-chars")

		return pager.RunLess(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(lessCmd)

	lessCmd.Flags().BoolP("LINE-NUMBERS", "N", false, "show line numbers")
	lessCmd.Flags().BoolP("no-init", "X", false, "don't clear screen on start")
	lessCmd.Flags().BoolP("quit-if-one-screen", "F", false, "quit if content fits on one screen")
	lessCmd.Flags().BoolP("ignore-case", "i", false, "case-insensitive search")
	lessCmd.Flags().BoolP("chop-long-lines", "S", false, "truncate long lines")
	lessCmd.Flags().BoolP("raw-control-chars", "R", false, "show raw control characters")
}
