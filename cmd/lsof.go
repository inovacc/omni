package cmd

import (
	"github.com/inovacc/omni/internal/cli/lsof"
	"github.com/spf13/cobra"
)

// lsofCmd represents the lsof command
var lsofCmd = &cobra.Command{
	Use:   "lsof [OPTIONS]",
	Short: "List open files and network connections",
	Long: `List information about files and network connections opened by processes.

Shows network connections for all processes by default. Use filters to narrow down results.

Options:
  -p PID      show files for specific process ID
  -u USER     show files for specific user
  -i          show network connections only
  -i:PORT     show connections using specific port
  -c CMD      filter by command name prefix
  -t          show TCP connections only
  -U          show UDP connections only
  -4          show only IPv4
  -6          show only IPv6

Examples:
  omni lsof                    # show all network connections
  omni lsof -i                 # show network files only
  omni lsof -i:80              # show connections on port 80
  omni lsof -i:443 -t          # TCP connections on port 443
  omni lsof -p 1234            # show files for PID 1234
  omni lsof -u root            # show files for root user
  omni lsof -c nginx           # show files for nginx processes
  omni lsof -j                 # output as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := lsof.Options{}
		opts.PID, _ = cmd.Flags().GetInt("pid")
		opts.User, _ = cmd.Flags().GetString("user")
		opts.Network, _ = cmd.Flags().GetBool("network")
		opts.Command, _ = cmd.Flags().GetString("command")
		opts.NoHeaders, _ = cmd.Flags().GetBool("no-headers")
		opts.JSON, _ = cmd.Flags().GetBool("json")
		opts.IPv4, _ = cmd.Flags().GetBool("ipv4")
		opts.IPv6, _ = cmd.Flags().GetBool("ipv6")
		opts.Listen, _ = cmd.Flags().GetBool("listen")
		opts.Established, _ = cmd.Flags().GetBool("established")

		// Handle -i:PORT syntax
		portFilter, _ := cmd.Flags().GetInt("port")
		if portFilter > 0 {
			opts.Port = portFilter
			opts.Network = true
		}

		// Protocol filter
		tcpOnly, _ := cmd.Flags().GetBool("tcp")
		udpOnly, _ := cmd.Flags().GetBool("udp")
		if tcpOnly {
			opts.Protocol = "tcp"
		} else if udpOnly {
			opts.Protocol = "udp"
		}

		return lsof.Run(cmd.OutOrStdout(), opts)
	},
}

func init() {
	rootCmd.AddCommand(lsofCmd)

	lsofCmd.Flags().IntP("pid", "p", 0, "show files for specific process ID")
	lsofCmd.Flags().StringP("user", "u", "", "show files for specific user")
	lsofCmd.Flags().BoolP("network", "i", false, "show network connections only")
	lsofCmd.Flags().Int("port", 0, "filter by port number (use with -i)")
	lsofCmd.Flags().StringP("command", "c", "", "filter by command name prefix")
	lsofCmd.Flags().BoolP("tcp", "t", false, "show TCP connections only")
	lsofCmd.Flags().BoolP("udp", "U", false, "show UDP connections only")
	lsofCmd.Flags().BoolP("ipv4", "4", false, "show only IPv4")
	lsofCmd.Flags().BoolP("ipv6", "6", false, "show only IPv6")
	lsofCmd.Flags().BoolP("listen", "l", false, "show only listening sockets")
	lsofCmd.Flags().BoolP("established", "e", false, "show only established connections")
	lsofCmd.Flags().BoolP("no-headers", "n", false, "don't print headers")
	lsofCmd.Flags().BoolP("json", "j", false, "output as JSON")
}
