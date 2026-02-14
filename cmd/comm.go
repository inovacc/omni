package cmd

import (
	"github.com/inovacc/omni/internal/cli/comm"
	"github.com/spf13/cobra"
)

var (
	commSuppress1    bool
	commSuppress2    bool
	commSuppress3    bool
	commCheckOrder   bool
	commNoCheckOrder bool
	commOutputDelim  string
	commZeroTerm     bool
)

var commCmd = &cobra.Command{
	Use:   "comm [OPTION]... FILE1 FILE2",
	Short: "Compare two sorted files line by line",
	Long: `Compare sorted files FILE1 and FILE2 line by line.

With no options, produce three-column output:
  Column 1: lines unique to FILE1
  Column 2: lines unique to FILE2
  Column 3: lines common to both files

  -1                 suppress column 1 (lines unique to FILE1)
  -2                 suppress column 2 (lines unique to FILE2)
  -3                 suppress column 3 (lines common to both)
  --check-order      check that input is correctly sorted
  --nocheck-order    do not check input order
  --output-delimiter use STR as output delimiter
  -z, --zero-terminated  line delimiter is NUL

Examples:
  omni comm file1.txt file2.txt        # show all columns
  omni comm -12 file1.txt file2.txt    # show only common lines
  omni comm -3 file1.txt file2.txt     # show only unique lines`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := comm.CommOptions{
			Suppress1:    commSuppress1,
			Suppress2:    commSuppress2,
			Suppress3:    commSuppress3,
			CheckOrder:   commCheckOrder,
			NoCheckOrder: commNoCheckOrder,
			OutputDelim:  commOutputDelim,
			ZeroTerm:     commZeroTerm,
			OutputFormat: getOutputOpts(cmd).GetFormat(),
		}

		return comm.RunComm(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(commCmd)

	commCmd.Flags().BoolVarP(&commSuppress1, "1", "1", false, "suppress column 1")
	commCmd.Flags().BoolVarP(&commSuppress2, "2", "2", false, "suppress column 2")
	commCmd.Flags().BoolVarP(&commSuppress3, "3", "3", false, "suppress column 3")
	commCmd.Flags().BoolVar(&commCheckOrder, "check-order", false, "check input is sorted")
	commCmd.Flags().BoolVar(&commNoCheckOrder, "nocheck-order", false, "do not check input order")
	commCmd.Flags().StringVar(&commOutputDelim, "output-delimiter", "", "use STR as output delimiter")
	commCmd.Flags().BoolVarP(&commZeroTerm, "zero-terminated", "z", false, "line delimiter is NUL")
}
