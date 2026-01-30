package cmd

import (
	"github.com/inovacc/omni/internal/cli/kill"
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
  -j, --json           output as JSON

Signal can be specified by name (e.g., HUP, KILL, TERM) or number.
Common signals:
   1) SIGHUP       2) SIGINT       3) SIGQUIT
   9) SIGKILL     15) SIGTERM (default)

Examples:
  omni kill 1234           # send SIGTERM to process 1234
  omni kill -9 1234        # send SIGKILL to process 1234
  omni kill -s HUP 1234    # send SIGHUP to process 1234
  omni kill -l             # list all signal names
  omni kill -l -j          # list signals as JSON
  omni kill -j 1234        # kill with JSON output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := kill.KillOptions{}

		opts.Signal, _ = cmd.Flags().GetString("signal")
		opts.List, _ = cmd.Flags().GetBool("list")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return kill.RunKill(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(killCmd)

	killCmd.Flags().StringP("signal", "s", "", "specify the signal to be sent")
	killCmd.Flags().BoolP("list", "l", false, "list signal names")
	killCmd.Flags().BoolP("verbose", "v", false, "report successful signals")
	killCmd.Flags().BoolP("json", "j", false, "output as JSON")
}
