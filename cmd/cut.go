package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/cut"
	"github.com/spf13/cobra"
)

// cutCmd represents the cut command
var cutCmd = &cobra.Command{
	Use:   "cut [OPTION]... [FILE]...",
	Short: "Remove sections from each line of files",
	Long: `Print selected parts of lines from each FILE to standard output.

With no FILE, or when FILE is -, read standard input.

Mandatory arguments to long options are mandatory for short options too.
  -b, --bytes=LIST        select only these bytes
  -c, --characters=LIST   select only these characters
  -d, --delimiter=DELIM   use DELIM instead of TAB for field delimiter
  -f, --fields=LIST       select only these fields
  -s, --only-delimited    do not print lines not containing delimiters
      --complement        complement the set of selected bytes, characters or fields
      --output-delimiter=STRING  use STRING as the output delimiter

Use one, and only one of -b, -c or -f.  Each LIST is made up of one
range, or many ranges separated by commas.  Each range is one of:
  N     N'th byte, character or field, counted from 1
  N-    from N'th byte, character or field, to end of line
  N-M   from N'th to M'th (included) byte, character or field
  -M    from first to M'th (included) byte, character or field`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cut.CutOptions{}

		opts.Bytes, _ = cmd.Flags().GetString("bytes")
		opts.Characters, _ = cmd.Flags().GetString("characters")
		opts.Fields, _ = cmd.Flags().GetString("fields")
		opts.Delimiter, _ = cmd.Flags().GetString("delimiter")
		opts.OnlyDelim, _ = cmd.Flags().GetBool("only-delimited")
		opts.OutputDelim, _ = cmd.Flags().GetString("output-delimiter")
		opts.Complement, _ = cmd.Flags().GetBool("complement")

		return cut.RunCut(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(cutCmd)

	cutCmd.Flags().StringP("bytes", "b", "", "select only these bytes")
	cutCmd.Flags().StringP("characters", "c", "", "select only these characters")
	cutCmd.Flags().StringP("fields", "f", "", "select only these fields")
	cutCmd.Flags().StringP("delimiter", "d", "", "use DELIM instead of TAB for field delimiter")
	cutCmd.Flags().BoolP("only-delimited", "s", false, "do not print lines not containing delimiters")
	cutCmd.Flags().String("output-delimiter", "", "use STRING as the output delimiter")
	cutCmd.Flags().Bool("complement", false, "complement the set of selected bytes, characters or fields")
}
