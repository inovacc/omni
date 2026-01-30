package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/xargs"
	"github.com/spf13/cobra"
)

// xargsCmd represents the xargs command
var xargsCmd = &cobra.Command{
	Use:   "xargs [OPTION]... [COMMAND [INITIAL-ARGS]]",
	Short: "Build and execute command lines from standard input",
	Long: `Read items from standard input, delimited by blanks or newlines, and
execute a command for each item.

Note: Since omni doesn't execute external commands, this version
reads and prints arguments from stdin. It can be used to transform
input for piping to other tools.

  -0, --null            input items are separated by a null character
  -d, --delimiter=DELIM  input items are separated by DELIM
  -n, --max-args=MAX    use at most MAX arguments per command line
  -P, --max-procs=MAX   run at most MAX processes at a time
  -r, --no-run-if-empty if there are no arguments, do not run COMMAND
  -t, --verbose         print commands before executing them
  -I REPLACE-STR        replace occurrences of REPLACE-STR in initial args

Examples:
  echo "a b c" | omni xargs        # prints: a b c
  echo -e "a\nb\nc" | omni xargs   # prints: a b c
  echo -e "a\nb\nc" | omni xargs -n 1  # prints each on separate line`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := xargs.XargsOptions{}

		opts.NullInput, _ = cmd.Flags().GetBool("null")
		opts.Delimiter, _ = cmd.Flags().GetString("delimiter")
		opts.MaxArgs, _ = cmd.Flags().GetInt("max-args")
		opts.MaxProcs, _ = cmd.Flags().GetInt("max-procs")
		opts.NoRunEmpty, _ = cmd.Flags().GetBool("no-run-if-empty")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		opts.ReplaceStr, _ = cmd.Flags().GetString("I")

		return xargs.RunXargsWithPrint(cmd.OutOrStdout(), os.Stdin, opts)
	},
}

func init() {
	rootCmd.AddCommand(xargsCmd)

	xargsCmd.Flags().BoolP("null", "0", false, "input items are separated by a null character")
	xargsCmd.Flags().StringP("delimiter", "d", "", "input items are separated by DELIM")
	xargsCmd.Flags().IntP("max-args", "n", 0, "use at most MAX arguments per command line")
	xargsCmd.Flags().IntP("max-procs", "P", 1, "run at most MAX processes at a time")
	xargsCmd.Flags().BoolP("no-run-if-empty", "r", false, "if there are no arguments, do not run")
	xargsCmd.Flags().BoolP("verbose", "t", false, "print commands before executing them")
	xargsCmd.Flags().StringP("I", "I", "", "replace occurrences of REPLACE-STR")
}
