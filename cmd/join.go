package cmd

import (
	"os"

	"github.com/inovacc/goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// joinCmd represents the join command
var joinCmd = &cobra.Command{
	Use:   "join [OPTION]... FILE1 FILE2",
	Short: "Join lines of two files on a common field",
	Long: `For each pair of input lines with identical join fields, write a line to
standard output. The default join field is the first, delimited by blanks.

  -1 FIELD       join on this FIELD of file 1
  -2 FIELD       join on this FIELD of file 2
  -t CHAR        use CHAR as input and output field separator
  -i             ignore differences in case when comparing fields
  -a FILENUM     also print unpairable lines from file FILENUM
  -v FILENUM     print only unpairable lines from file FILENUM
  -e EMPTY       replace missing fields with EMPTY

When FILE1 or FILE2 is -, read standard input.

Examples:
  goshell join file1.txt file2.txt           # join on first field
  goshell join -1 2 -2 1 file1.txt file2.txt # join field 2 of file1 with field 1 of file2
  goshell join -t ',' data1.csv data2.csv    # join CSV files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.JoinOptions{}

		opts.Field1, _ = cmd.Flags().GetInt("1")
		opts.Field2, _ = cmd.Flags().GetInt("2")
		opts.Separator, _ = cmd.Flags().GetString("t")
		opts.IgnoreCase, _ = cmd.Flags().GetBool("i")
		opts.Empty, _ = cmd.Flags().GetString("e")

		unpaired, _ := cmd.Flags().GetInt("a")
		switch unpaired {
		case 1:
			opts.Unpaired1 = true
		case 2:
			opts.Unpaired2 = true
		}

		onlyUnpaired, _ := cmd.Flags().GetInt("v")
		switch onlyUnpaired {
		case 1:
			opts.OnlyUnpaired1 = true
		case 2:
			opts.OnlyUnpaired2 = true
		}

		return cli.RunJoin(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(joinCmd)

	joinCmd.Flags().IntP("1", "1", 1, "join on this FIELD of file 1")
	joinCmd.Flags().IntP("2", "2", 1, "join on this FIELD of file 2")
	joinCmd.Flags().StringP("t", "t", "", "use CHAR as input and output field separator")
	joinCmd.Flags().BoolP("i", "i", false, "ignore case when comparing fields")
	joinCmd.Flags().StringP("e", "e", "", "replace missing fields with EMPTY")
	joinCmd.Flags().IntP("a", "a", 0, "also print unpairable lines from file FILENUM (1 or 2)")
	joinCmd.Flags().IntP("v", "v", 0, "print only unpairable lines from file FILENUM (1 or 2)")
}
