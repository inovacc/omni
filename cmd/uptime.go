package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli/uptime"

	"github.com/spf13/cobra"
)

// uptimeCmd represents the uptime command
var uptimeCmd = &cobra.Command{
	Use:   "uptime [OPTION]...",
	Short: "Tell how long the system has been running",
	Long: `Print the current time, how long the system has been running,
how many users are currently logged on, and the system load averages
for the past 1, 5, and 15 minutes.

  -p, --pretty   show uptime in pretty format
  -s, --since    system up since, in yyyy-mm-dd HH:MM:SS format`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := uptime.UptimeOptions{}

		opts.Pretty, _ = cmd.Flags().GetBool("pretty")
		opts.Since, _ = cmd.Flags().GetBool("since")

		return uptime.RunUptime(os.Stdout, opts)
	},
}

func init() {
	rootCmd.AddCommand(uptimeCmd)

	uptimeCmd.Flags().BoolP("pretty", "p", false, "show uptime in pretty format")
	uptimeCmd.Flags().BoolP("since", "s", false, "system up since")
}
