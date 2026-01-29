package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/seq"
	"github.com/spf13/cobra"
)

var (
	seqSeparator  string
	seqFormat     string
	seqEqualWidth bool
)

var seqCmd = &cobra.Command{
	Use:   "seq [OPTION]... LAST or seq [OPTION]... FIRST LAST or seq [OPTION]... FIRST INCREMENT LAST",
	Short: "Print a sequence of numbers",
	Long: `Print numbers from FIRST to LAST, in steps of INCREMENT.

  -s, --separator=STRING  use STRING to separate numbers (default: \n)
  -f, --format=FORMAT     use printf style floating-point FORMAT
  -w, --equal-width       equalize width by padding with leading zeros

Examples:
  omni seq 5               # print 1 2 3 4 5
  omni seq 2 5             # print 2 3 4 5
  omni seq 1 2 10          # print 1 3 5 7 9
  omni seq -w 1 10         # print 01 02 ... 10
  omni seq -s ', ' 1 5     # print 1, 2, 3, 4, 5
  omni seq 0.5 0.1 1.0     # print 0.5 0.6 0.7 0.8 0.9 1.0`,
	Args: cobra.RangeArgs(1, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := seq.SeqOptions{
			Separator:  seqSeparator,
			Format:     seqFormat,
			EqualWidth: seqEqualWidth,
		}

		return seq.RunSeq(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(seqCmd)

	seqCmd.Flags().StringVarP(&seqSeparator, "separator", "s", "", "use STRING to separate numbers")
	seqCmd.Flags().StringVarP(&seqFormat, "format", "f", "", "use printf style FORMAT")
	seqCmd.Flags().BoolVarP(&seqEqualWidth, "equal-width", "w", false, "equalize width with leading zeros")
}
