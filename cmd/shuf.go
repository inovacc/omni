package cmd

import (
	"github.com/inovacc/omni/internal/cli/shuf"
	"github.com/spf13/cobra"
)

var (
	shufEcho       bool
	shufInputRange string
	shufHeadCount  int
	shufRepeat     bool
	shufZeroTerm   bool
	shufJSON       bool
)

var shufCmd = &cobra.Command{
	Use:   "shuf [OPTION]... [FILE]",
	Short: "Generate random permutations",
	Long: `Write a random permutation of the input lines to standard output.

  -e, --echo          treat each ARG as an input line
  -i, --input-range   treat each number LO through HI as an input line
  -n, --head-count    output at most COUNT lines
  -r, --repeat        output lines can be repeated (with -n)
  -z, --zero-terminated  line delimiter is NUL, not newline

Examples:
  omni shuf file.txt              # shuffle lines of file
  omni shuf -e a b c d            # shuffle arguments
  omni shuf -i 1-10               # shuffle numbers 1-10
  omni shuf -n 5 file.txt         # output 5 random lines
  omni shuf -rn 10 -e yes no      # 10 random picks with repetition`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := shuf.ShufOptions{
			Echo:       shufEcho,
			InputRange: shufInputRange,
			HeadCount:  shufHeadCount,
			Repeat:     shufRepeat,
			ZeroTerm:   shufZeroTerm,
			JSON:       shufJSON,
		}

		return shuf.RunShuf(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(shufCmd)

	shufCmd.Flags().BoolVarP(&shufEcho, "echo", "e", false, "treat each ARG as an input line")
	shufCmd.Flags().StringVarP(&shufInputRange, "input-range", "i", "", "treat each number LO through HI as an input line")
	shufCmd.Flags().IntVarP(&shufHeadCount, "head-count", "n", 0, "output at most COUNT lines")
	shufCmd.Flags().BoolVarP(&shufRepeat, "repeat", "r", false, "output lines can be repeated")
	shufCmd.Flags().BoolVarP(&shufZeroTerm, "zero-terminated", "z", false, "line delimiter is NUL")
	shufCmd.Flags().BoolVar(&shufJSON, "json", false, "output as JSON")
}
