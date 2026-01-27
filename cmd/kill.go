package cmd

import (
	"os"

	"github.com/inovacc/goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// killCmd represents the kill command
var killCmd = &cobra.Command{
	Use:   "kill [OPTION]... PID...",
	Short: "Send a signal to a process",
	Long: `Send the specified signal to the specified processes.

  -s, --signal=SIGNAL  specify the signal to be sent
  -l, --list           list signal names
  -v, --verbose        report successful signals

Signal can be specified by name (e.g., HUP, KILL, TERM) or number.
Common signals:
   1) SIGHUP       2) SIGINT       3) SIGQUIT
   9) SIGKILL     15) SIGTERM (default)

Examples:
  goshell kill 1234           # send SIGTERM to process 1234
  goshell kill -9 1234        # send SIGKILL to process 1234
  goshell kill -s HUP 1234    # send SIGHUP to process 1234
  goshell kill -l             # list all signal names`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.KillOptions{}

		opts.Signal, _ = cmd.Flags().GetString("signal")
		opts.List, _ = cmd.Flags().GetBool("list")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")

		return cli.RunKill(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(killCmd)

	killCmd.Flags().StringP("signal", "s", "", "specify the signal to be sent")
	killCmd.Flags().BoolP("list", "l", false, "list signal names")
	killCmd.Flags().BoolP("verbose", "v", false, "report successful signals")
}
