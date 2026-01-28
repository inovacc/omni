package cmd

import (
	"github.com/inovacc/omni/pkg/cli"
	"github.com/spf13/cobra"
)

var sleepCmd = &cobra.Command{
	Use:   "sleep NUMBER[SUFFIX]...",
	Short: "Delay for a specified amount of time",
	Long: `Pause for NUMBER seconds. SUFFIX may be:
  s   seconds (default)
  m   minutes
  h   hours
  d   days

NUMBER may be a decimal.

Examples:
  omni sleep 5           # sleep 5 seconds
  omni sleep 0.5         # sleep 0.5 seconds
  omni sleep 1m          # sleep 1 minute
  omni sleep 1h 30m      # sleep 1.5 hours`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.RunSleep(args)
	},
}

func init() {
	rootCmd.AddCommand(sleepCmd)
}
