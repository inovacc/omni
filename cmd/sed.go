package cmd

import (
	"os"

	"github.com/inovacc/goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// sedCmd represents the sed command
var sedCmd = &cobra.Command{
	Use:   "sed [OPTION]... {script} [FILE]...",
	Short: "Stream editor for filtering and transforming text",
	Long: `Sed is a stream editor. A stream editor is used to perform basic text
transformations on an input stream (a file or input from a pipeline).

  -e script      add the script to the commands to be executed
  -i[SUFFIX]     edit files in place (makes backup if SUFFIX supplied)
  -n             suppress automatic printing of pattern space
  -E, -r         use extended regular expressions

Supported commands:
  s/regexp/replacement/flags  substitute
  d                           delete pattern space
  p                           print pattern space
  q                           quit

Examples:
  goshell sed 's/old/new/' file.txt        # replace first occurrence
  goshell sed 's/old/new/g' file.txt       # replace all occurrences
  goshell sed -i.bak 's/foo/bar/g' file    # in-place edit with backup
  goshell sed '/pattern/d' file.txt        # delete matching lines
  goshell sed -n '/pattern/p' file.txt     # print only matching lines`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.SedOptions{}

		exprs, _ := cmd.Flags().GetStringSlice("expression")
		opts.Expression = exprs
		opts.InPlace, _ = cmd.Flags().GetBool("in-place")
		opts.InPlaceExt, _ = cmd.Flags().GetString("in-place-suffix")
		opts.Quiet, _ = cmd.Flags().GetBool("quiet")
		opts.Extended, _ = cmd.Flags().GetBool("regexp-extended")

		return cli.RunSed(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(sedCmd)

	sedCmd.Flags().StringSliceP("expression", "e", nil, "add the script to the commands to be executed")
	sedCmd.Flags().BoolP("in-place", "i", false, "edit files in place")
	sedCmd.Flags().String("in-place-suffix", "", "backup suffix for in-place edit")
	sedCmd.Flags().BoolP("quiet", "n", false, "suppress automatic printing of pattern space")
	sedCmd.Flags().BoolP("regexp-extended", "E", false, "use extended regular expressions")
	sedCmd.Flags().BoolP("r", "r", false, "use extended regular expressions (alias)")
}
