package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli"
	"github.com/spf13/cobra"
)

var (
	splitLines      int
	splitBytes      string
	splitSuffixLen  int
	splitNumericSfx bool
	splitVerbose    bool
)

var splitCmd = &cobra.Command{
	Use:   "split [OPTION]... [FILE [PREFIX]]",
	Short: "Split a file into pieces",
	Long: `Output pieces of FILE to PREFIXaa, PREFIXab, ...;
default PREFIX is 'x'.

  -l, --lines=NUMBER   put NUMBER lines per output file
  -b, --bytes=SIZE     put SIZE bytes per output file
  -a, --suffix-length  generate suffixes of length N (default 2)
  -d, --numeric-suffixes  use numeric suffixes instead of alphabetic
      --verbose        print a diagnostic just before each output file is opened

SIZE may have a suffix: K=1024, M=1024*1024, G=1024*1024*1024

Examples:
  omni split file.txt              # split into 1000-line files
  omni split -l 100 file.txt       # split into 100-line files
  omni split -b 1M file.bin        # split into 1MB files
  omni split -d file.txt chunk_    # use numeric suffixes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.SplitOptions{
			Lines:       splitLines,
			Bytes:       splitBytes,
			Suffix:      splitSuffixLen,
			NumericSufx: splitNumericSfx,
			Verbose:     splitVerbose,
		}

		return cli.RunSplit(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(splitCmd)

	splitCmd.Flags().IntVarP(&splitLines, "lines", "l", 0, "put NUMBER lines per output file")
	splitCmd.Flags().StringVarP(&splitBytes, "bytes", "b", "", "put SIZE bytes per output file")
	splitCmd.Flags().IntVarP(&splitSuffixLen, "suffix-length", "a", 2, "generate suffixes of length N")
	splitCmd.Flags().BoolVarP(&splitNumericSfx, "numeric-suffixes", "d", false, "use numeric suffixes")
	splitCmd.Flags().BoolVar(&splitVerbose, "verbose", false, "print diagnostic for each output file")
}
