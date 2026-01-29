package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/diff"
	"github.com/spf13/cobra"
)

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff [OPTION]... FILE1 FILE2",
	Short: "Compare files line by line",
	Long: `Compare files line by line.

Output format:
  - Unified diff (default): shows context with +/- markers
  - Side-by-side (-y): shows files in parallel columns
  - Brief (-q): only report if files differ

Special modes:
  --json    Compare JSON files structurally
  -r        Recursively compare directories

Examples:
  omni diff file1.txt file2.txt
  omni diff -u 5 old.txt new.txt         # 5 lines of context
  omni diff -y file1.txt file2.txt       # side-by-side
  omni diff -q dir1/ dir2/               # brief comparison
  omni diff --json config1.json config2.json
  omni diff -r dir1/ dir2/               # recursive`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := diff.DiffOptions{}

		opts.Unified, _ = cmd.Flags().GetInt("unified")
		opts.Side, _ = cmd.Flags().GetBool("side-by-side")
		opts.Brief, _ = cmd.Flags().GetBool("brief")
		opts.IgnoreCase, _ = cmd.Flags().GetBool("ignore-case")
		opts.IgnoreSpace, _ = cmd.Flags().GetBool("ignore-space-change")
		opts.IgnoreBlank, _ = cmd.Flags().GetBool("ignore-blank-lines")
		opts.Recursive, _ = cmd.Flags().GetBool("recursive")
		opts.JSON, _ = cmd.Flags().GetBool("json")
		opts.Color, _ = cmd.Flags().GetBool("color")
		opts.Width, _ = cmd.Flags().GetInt("width")
		opts.SuppressCommon, _ = cmd.Flags().GetBool("suppress-common-lines")

		return diff.RunDiff(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().IntP("unified", "u", 3, "output NUM lines of unified context")
	diffCmd.Flags().BoolP("side-by-side", "y", false, "output in two columns")
	diffCmd.Flags().BoolP("brief", "q", false, "report only when files differ")
	diffCmd.Flags().BoolP("ignore-case", "i", false, "ignore case differences")
	diffCmd.Flags().BoolP("ignore-space-change", "b", false, "ignore changes in amount of white space")
	diffCmd.Flags().BoolP("ignore-blank-lines", "B", false, "ignore changes where lines are all blank")
	diffCmd.Flags().BoolP("recursive", "r", false, "recursively compare subdirectories")
	diffCmd.Flags().Bool("json", false, "compare as JSON files")
	diffCmd.Flags().Bool("color", false, "colorize the output")
	diffCmd.Flags().IntP("width", "W", 130, "output at most NUM columns")
	diffCmd.Flags().Bool("suppress-common-lines", false, "do not output common lines in side-by-side")
}
