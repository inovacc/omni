package cmd

import (
	"github.com/inovacc/omni/internal/cli/tr"
	"github.com/spf13/cobra"
)

// trCmd represents the tr command
var trCmd = &cobra.Command{
	Use:   "tr [OPTION]... SET1 [SET2]",
	Short: "Translate or delete characters",
	Long: `Translate, squeeze, and/or delete characters from standard input,
writing to standard output.

  -c, --complement    use the complement of SET1
  -d, --delete        delete characters in SET1, do not translate
  -s, --squeeze-repeats  replace each sequence of a repeated character
                         that is listed in the last SET, with a single
                         occurrence of that character
  -t, --truncate-set1 first truncate SET1 to length of SET2

SETs are specified as strings of characters.  Most represent themselves.
Interpreted sequences are:

  \n     new line
  \r     return
  \t     horizontal tab
  \\     backslash

  CHAR1-CHAR2  all characters from CHAR1 to CHAR2 in ascending order

  [:alnum:]    all letters and digits
  [:alpha:]    all letters
  [:digit:]    all digits
  [:lower:]    all lower case letters
  [:upper:]    all upper case letters
  [:space:]    all horizontal or vertical whitespace
  [:blank:]    all horizontal whitespace
  [:punct:]    all punctuation characters
  [:graph:]    all printable characters, not including space
  [:print:]    all printable characters, including space
  [:xdigit:]   all hexadecimal digits`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := tr.TrOptions{}

		opts.Complement, _ = cmd.Flags().GetBool("complement")
		opts.Delete, _ = cmd.Flags().GetBool("delete")
		opts.Squeeze, _ = cmd.Flags().GetBool("squeeze-repeats")
		opts.Truncate, _ = cmd.Flags().GetBool("truncate-set1")

		set1 := args[0]
		set2 := ""
		if len(args) > 1 {
			set2 = args[1]
		}

		return tr.RunTr(cmd.OutOrStdout(), cmd.InOrStdin(), set1, set2, opts)
	},
}

func init() {
	rootCmd.AddCommand(trCmd)

	trCmd.Flags().BoolP("complement", "c", false, "use the complement of SET1")
	trCmd.Flags().BoolP("delete", "d", false, "delete characters in SET1, do not translate")
	trCmd.Flags().BoolP("squeeze-repeats", "s", false, "replace repeated characters with single occurrence")
	trCmd.Flags().BoolP("truncate-set1", "t", false, "first truncate SET1 to length of SET2")
}
