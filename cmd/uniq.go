package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli"

	"github.com/spf13/cobra"
)

// uniqCmd represents the uniq command
var uniqCmd = &cobra.Command{
	Use:   "uniq [option]... [input [output]]",
	Short: "Report or omit repeated lines",
	Long: `Filter adjacent matching lines from INPUT (or standard input),
writing to OUTPUT (or standard output).

With no options, matching lines are merged to the first occurrence.

Note: 'uniq' does not detect repeated lines unless they are adjacent.
You may want to sort the input first, or use 'sort -u' without 'uniq'.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.UniqOptions{}

		opts.Count, _ = cmd.Flags().GetBool("count")
		opts.Repeated, _ = cmd.Flags().GetBool("repeated")
		opts.AllRepeated, _ = cmd.Flags().GetBool("all-repeated")
		opts.IgnoreCase, _ = cmd.Flags().GetBool("ignore-case")
		opts.Unique, _ = cmd.Flags().GetBool("unique")
		opts.SkipFields, _ = cmd.Flags().GetInt("skip-fields")
		opts.SkipChars, _ = cmd.Flags().GetInt("skip-chars")
		opts.CheckChars, _ = cmd.Flags().GetInt("check-chars")
		opts.ZeroTerminate, _ = cmd.Flags().GetBool("zero-terminated")

		return cli.RunUniq(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(uniqCmd)

	uniqCmd.Flags().BoolP("count", "c", false, "prefix lines by the number of occurrences")
	uniqCmd.Flags().BoolP("repeated", "d", false, "only print duplicate lines, one for each group")
	uniqCmd.Flags().BoolP("all-repeated", "D", false, "print all duplicate lines")
	uniqCmd.Flags().BoolP("ignore-case", "i", false, "ignore differences in case when comparing")
	uniqCmd.Flags().BoolP("unique", "u", false, "only print unique lines")
	uniqCmd.Flags().IntP("skip-fields", "f", 0, "avoid comparing the first N fields")
	uniqCmd.Flags().IntP("skip-chars", "s", 0, "avoid comparing the first N characters")
	uniqCmd.Flags().IntP("check-chars", "w", 0, "compare no more than N characters in lines")
	uniqCmd.Flags().BoolP("zero-terminated", "z", false, "line delimiter is NUL, not newline")
}
