package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli/text"

	"github.com/spf13/cobra"
)

// sortCmd represents the sort command
var sortCmd = &cobra.Command{
	Use:   "sort [option]... [file]...",
	Short: "Sort lines of text files",
	Long: `Write sorted concatenation of all FILE(s) to standard output.

With no FILE, or when FILE is -, read standard input.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := text.SortOptions{}

		opts.Reverse, _ = cmd.Flags().GetBool("reverse")
		opts.Numeric, _ = cmd.Flags().GetBool("numeric-sort")
		opts.Unique, _ = cmd.Flags().GetBool("unique")
		opts.IgnoreCase, _ = cmd.Flags().GetBool("ignore-case")
		opts.IgnoreLeading, _ = cmd.Flags().GetBool("ignore-leading-blanks")
		opts.Dictionary, _ = cmd.Flags().GetBool("dictionary-order")
		opts.Key, _ = cmd.Flags().GetString("key")
		opts.FieldSep, _ = cmd.Flags().GetString("field-separator")
		opts.Check, _ = cmd.Flags().GetBool("check")
		opts.Stable, _ = cmd.Flags().GetBool("stable")
		opts.Output, _ = cmd.Flags().GetString("output")

		return text.RunSort(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(sortCmd)

	// Ordering options
	sortCmd.Flags().BoolP("reverse", "r", false, "reverse the result of comparisons")
	sortCmd.Flags().BoolP("numeric-sort", "n", false, "compare according to string numerical value")
	sortCmd.Flags().BoolP("ignore-case", "f", false, "fold lower case to upper case characters")
	sortCmd.Flags().BoolP("ignore-leading-blanks", "b", false, "ignore leading blanks")
	sortCmd.Flags().BoolP("dictionary-order", "d", false, "consider only blanks and alphanumeric characters")

	// Other options
	sortCmd.Flags().BoolP("unique", "u", false, "with -c, check for strict ordering; without -c, output only the first of an equal run")
	sortCmd.Flags().StringP("key", "k", "", "sort via a key")
	sortCmd.Flags().StringP("field-separator", "t", "", "use SEP instead of non-blank to blank transition")
	sortCmd.Flags().BoolP("check", "c", false, "check for sorted input; do not sort")
	sortCmd.Flags().BoolP("stable", "s", false, "stabilize sort by disabling last-resort comparison")
	sortCmd.Flags().StringP("output", "o", "", "write result to FILE instead of standard output")
}
