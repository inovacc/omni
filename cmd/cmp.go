package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli"
	"github.com/spf13/cobra"
)

var (
	cmpSilent     bool
	cmpVerbose    bool
	cmpPrintBytes bool
	cmpSkipBytes  int64
	cmpMaxBytes   int64
)

var cmpCmd = &cobra.Command{
	Use:   "cmp [OPTION]... FILE1 FILE2",
	Short: "Compare two files byte by byte",
	Long: `Compare two files byte by byte.

  -s, --silent   suppress all normal output
  -l, --verbose  output byte numbers and differing byte values
  -b, --print-bytes  print differing bytes
  -i, --ignore-initial=SKIP  skip first SKIP bytes
  -n, --bytes=LIMIT  compare at most LIMIT bytes

Exit status:
  0  files are identical
  1  files differ
  2  trouble

Examples:
  omni cmp file1.bin file2.bin     # compare files
  omni cmp -s file1 file2          # silent, check exit status
  omni cmp -l file1 file2          # show all differences`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.CmpOptions{
			Silent:     cmpSilent,
			Verbose:    cmpVerbose,
			PrintBytes: cmpPrintBytes,
			SkipBytes1: cmpSkipBytes,
			SkipBytes2: cmpSkipBytes,
			MaxBytes:   cmpMaxBytes,
		}

		result, err := cli.RunCmp(os.Stdout, args, opts)
		if err != nil {
			return err
		}

		if result == cli.CmpDiffer {
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cmpCmd)

	cmpCmd.Flags().BoolVarP(&cmpSilent, "silent", "s", false, "suppress all output")
	cmpCmd.Flags().BoolVarP(&cmpVerbose, "verbose", "l", false, "output byte numbers and values")
	cmpCmd.Flags().BoolVarP(&cmpPrintBytes, "print-bytes", "b", false, "print differing bytes")
	cmpCmd.Flags().Int64VarP(&cmpSkipBytes, "ignore-initial", "i", 0, "skip first SKIP bytes")
	cmpCmd.Flags().Int64VarP(&cmpMaxBytes, "bytes", "n", 0, "compare at most LIMIT bytes")
}
