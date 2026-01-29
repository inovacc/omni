package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/grep"
	"github.com/spf13/cobra"
)

// fgrepCmd represents the fgrep command (grep -F)
var fgrepCmd = &cobra.Command{
	Use:   "fgrep [options] PATTERN [FILE...]",
	Short: "Print lines that match patterns (fixed strings)",
	Long: `Search for PATTERN in each FILE using fixed strings (no regex).
This is equivalent to 'grep -F'.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := grep.GrepOptions{
			FixedStrings: true, // fgrep always uses fixed strings
		}

		opts.IgnoreCase, _ = cmd.Flags().GetBool("ignore-case")
		opts.InvertMatch, _ = cmd.Flags().GetBool("invert-match")
		opts.LineNumber, _ = cmd.Flags().GetBool("line-number")
		opts.Count, _ = cmd.Flags().GetBool("count")
		opts.FilesWithMatch, _ = cmd.Flags().GetBool("files-with-matches")
		opts.OnlyMatching, _ = cmd.Flags().GetBool("only-matching")
		opts.Quiet, _ = cmd.Flags().GetBool("quiet")
		opts.Context, _ = cmd.Flags().GetInt("context")

		pattern := args[0]
		files := args[1:]

		return grep.RunGrep(os.Stdout, pattern, files, opts)
	},
}

func init() {
	rootCmd.AddCommand(fgrepCmd)

	fgrepCmd.Flags().BoolP("ignore-case", "i", false, "ignore case distinctions")
	fgrepCmd.Flags().BoolP("invert-match", "v", false, "select non-matching lines")
	fgrepCmd.Flags().BoolP("line-number", "n", false, "prefix each line with line number")
	fgrepCmd.Flags().BoolP("count", "c", false, "only print a count of matching lines")
	fgrepCmd.Flags().BoolP("files-with-matches", "l", false, "only print FILE names with matches")
	fgrepCmd.Flags().BoolP("only-matching", "o", false, "show only matched parts")
	fgrepCmd.Flags().BoolP("quiet", "q", false, "suppress all normal output")
	fgrepCmd.Flags().IntP("context", "C", 0, "print NUM lines of context")
}
