package cmd

import (
	"github.com/inovacc/omni/internal/cli/nl"
	"github.com/spf13/cobra"
)

// nlCmd represents the nl command
var nlCmd = &cobra.Command{
	Use:   "nl [OPTION]... [FILE]...",
	Short: "Number lines of files",
	Long: `Write each FILE to standard output, with line numbers added.

With no FILE, or when FILE is -, read standard input.

  -b, --body-numbering=STYLE      use STYLE for numbering body lines
  -n, --number-format=FORMAT      insert line numbers according to FORMAT
  -s, --number-separator=STRING   add STRING after line number
  -v, --starting-line-number=N    first line number on each logical page
  -i, --line-increment=N          line number increment at each line
  -w, --number-width=N            use N columns for line numbers

STYLE is one of:
  a      number all lines
  t      number only nonempty lines (default for body)
  n      number no lines

FORMAT is one of:
  ln     left justified, no leading zeros
  rn     right justified, no leading zeros (default)
  rz     right justified, leading zeros`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := nl.NlOptions{}

		opts.BodyNumbering, _ = cmd.Flags().GetString("body-numbering")
		opts.NumberFormat, _ = cmd.Flags().GetString("number-format")
		opts.NumberSep, _ = cmd.Flags().GetString("number-separator")
		opts.StartingNumber, _ = cmd.Flags().GetInt("starting-line-number")
		opts.Increment, _ = cmd.Flags().GetInt("line-increment")
		opts.NumberWidth, _ = cmd.Flags().GetInt("number-width")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return nl.RunNl(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(nlCmd)

	nlCmd.Flags().StringP("body-numbering", "b", "t", "use STYLE for numbering body lines")
	nlCmd.Flags().StringP("number-format", "n", "rn", "insert line numbers according to FORMAT")
	nlCmd.Flags().StringP("number-separator", "s", "\t", "add STRING after line number")
	nlCmd.Flags().IntP("starting-line-number", "v", 1, "first line number")
	nlCmd.Flags().IntP("line-increment", "i", 1, "line number increment")
	nlCmd.Flags().IntP("number-width", "w", 6, "use N columns for line numbers")
	nlCmd.Flags().Bool("json", false, "output as JSON")
}
