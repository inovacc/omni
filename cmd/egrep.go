package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/grep"
	"github.com/spf13/cobra"
)

// egrepCmd represents the egrep command (grep -E)
var egrepCmd = &cobra.Command{
	Use:   "egrep [options] PATTERN [FILE...]",
	Short: "Print lines that match patterns (extended regexp)",
	Long: `Search for PATTERN in each FILE using extended regular expressions.
This is equivalent to 'grep -E'.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := grep.GrepOptions{
			ExtendedRegexp: true, // egrep always uses extended regexp
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
	rootCmd.AddCommand(egrepCmd)

	egrepCmd.Flags().BoolP("ignore-case", "i", false, "ignore case distinctions")
	egrepCmd.Flags().BoolP("invert-match", "v", false, "select non-matching lines")
	egrepCmd.Flags().BoolP("line-number", "n", false, "prefix each line with line number")
	egrepCmd.Flags().BoolP("count", "c", false, "only print a count of matching lines")
	egrepCmd.Flags().BoolP("files-with-matches", "l", false, "only print FILE names with matches")
	egrepCmd.Flags().BoolP("only-matching", "o", false, "show only matched parts")
	egrepCmd.Flags().BoolP("quiet", "q", false, "suppress all normal output")
	egrepCmd.Flags().IntP("context", "C", 0, "print NUM lines of context")
}
