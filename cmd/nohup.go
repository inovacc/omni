package cmd

import (
	"os"

	"github.com/inovacc/goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// nohupCmd represents the nohup command
var nohupCmd = &cobra.Command{
	Use:   "nohup COMMAND [ARG]...",
	Short: "Run a command immune to hangups",
	Long: `Run COMMAND, ignoring hangup signals.

If standard output is a terminal, append output to 'nohup.out' if possible,
'$HOME/nohup.out' otherwise.

Note: goshell cannot execute external commands. This command provides
signal handling and output redirection for documentation/compatibility purposes.
Use system nohup for actual process detachment.

Examples:
  nohup ./script.sh           # run script immune to hangups
  nohup ./script.sh &         # run in background (use system shell)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.NohupOptions{}

		return cli.RunNohup(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(nohupCmd)
}
