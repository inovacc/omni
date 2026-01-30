package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/ps"
	"github.com/spf13/cobra"
)

// psCmd represents the ps command
var psCmd = &cobra.Command{
	Use:   "ps [OPTION]...",
	Short: "Report a snapshot of current processes",
	Long: `Display information about active processes.

  -a             show processes for all users
  -f             full-format listing
  -l             long format
  -u USER        show processes for specified user
  -p PID         show process with specified PID
  --no-headers   don't print header line
  --sort COL     sort by column (pid, cpu, mem, time)
  -j, --json     output as JSON
  --go           show only Go processes (detected via gops)

Go processes are automatically detected and marked. Use --go to filter
only Go processes, or -j/--json to see the is_go field in output.

Note: On Linux, reads /proc filesystem. On Windows, uses Win32 API.

Examples:
  omni ps                 # show current user's processes
  omni ps -a              # show all processes
  omni ps -f              # full format listing
  omni ps -p 1234         # show specific process
  omni ps --sort cpu      # sort by CPU usage
  omni ps -j              # output as JSON
  omni ps --go            # show only Go processes
  omni ps --go -j         # Go processes as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := ps.Options{}

		opts.All, _ = cmd.Flags().GetBool("all")
		opts.Full, _ = cmd.Flags().GetBool("full")
		opts.Long, _ = cmd.Flags().GetBool("long")
		opts.User, _ = cmd.Flags().GetString("user")
		opts.Pid, _ = cmd.Flags().GetInt("pid")
		opts.NoHeaders, _ = cmd.Flags().GetBool("no-headers")
		opts.Sort, _ = cmd.Flags().GetString("sort")
		opts.JSON, _ = cmd.Flags().GetBool("json")
		opts.GoOnly, _ = cmd.Flags().GetBool("go")

		// Check for aux style (positional arg)
		for _, arg := range args {
			if arg == "aux" {
				opts.Aux = true
				opts.All = true
			}
		}

		return ps.Run(cmd.OutOrStdout(), opts)
	},
}

func init() {
	rootCmd.AddCommand(psCmd)

	psCmd.Flags().BoolP("all", "a", false, "show processes for all users")
	psCmd.Flags().BoolP("full", "f", false, "full-format listing")
	psCmd.Flags().BoolP("long", "l", false, "long format")
	psCmd.Flags().StringP("user", "u", "", "show processes for specified user")
	psCmd.Flags().IntP("pid", "p", 0, "show process with specified PID")
	psCmd.Flags().Bool("no-headers", false, "don't print header line")
	psCmd.Flags().String("sort", "", "sort by column (pid, cpu, mem, time)")
	psCmd.Flags().BoolP("json", "j", false, "output as JSON")
	psCmd.Flags().Bool("go", false, "show only Go processes")
}
