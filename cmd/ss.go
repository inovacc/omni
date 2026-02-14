package cmd

import (
	"github.com/inovacc/omni/internal/cli/ss"
	"github.com/spf13/cobra"
)

// ssCmd represents the ss command
var ssCmd = &cobra.Command{
	Use:   "ss [OPTIONS]",
	Short: "Display socket statistics",
	Long: `Display socket statistics. Similar to the Linux ss command.

Shows information about TCP, UDP, and Unix sockets including state,
local/remote addresses, and optionally process information.

Options:
  -a          display all sockets
  -l          display listening sockets only
  -t          display TCP sockets
  -u          display UDP sockets
  -x          display Unix sockets
  -p          show process using socket
  -n          don't resolve service names
  -4          display only IPv4 sockets
  -6          display only IPv6 sockets
  -s          print summary statistics
  -e          show extended socket info
  --state     filter by state (established, listen, time_wait, etc.)

Examples:
  omni ss -l              # show listening sockets
  omni ss -t              # show TCP sockets
  omni ss -tl             # show TCP listening sockets
  omni ss -tlp            # show TCP listening sockets with process info
  omni ss -a              # show all sockets
  omni ss -s              # show summary statistics
  omni ss --state listen  # filter by state
  omni ss -4              # show only IPv4
  omni ss -j              # output as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := ss.Options{}
		opts.All, _ = cmd.Flags().GetBool("all")
		opts.Listening, _ = cmd.Flags().GetBool("listening")
		opts.TCP, _ = cmd.Flags().GetBool("tcp")
		opts.UDP, _ = cmd.Flags().GetBool("udp")
		opts.Unix, _ = cmd.Flags().GetBool("unix")
		opts.Processes, _ = cmd.Flags().GetBool("processes")
		opts.Numeric, _ = cmd.Flags().GetBool("numeric")
		opts.IPv4, _ = cmd.Flags().GetBool("ipv4")
		opts.IPv6, _ = cmd.Flags().GetBool("ipv6")
		opts.Summary, _ = cmd.Flags().GetBool("summary")
		opts.Extended, _ = cmd.Flags().GetBool("extended")
		opts.NoHeaders, _ = cmd.Flags().GetBool("no-header")
		opts.OutputFormat = getOutputOpts(cmd).GetFormat()
		opts.State, _ = cmd.Flags().GetString("state")

		return ss.Run(cmd.OutOrStdout(), opts)
	},
}

// netstatCmd is an alias for ss with netstat-like defaults
var netstatCmd = &cobra.Command{
	Use:   "netstat [OPTIONS]",
	Short: "Display network connections (alias for ss)",
	Long: `Display network connections. Alias for ss command.

Shows information about network connections including state,
local/remote addresses, and optionally process information.

Examples:
  omni netstat -a         # show all connections
  omni netstat -l         # show listening sockets
  omni netstat -t         # show TCP connections
  omni netstat -p         # show process info
  omni netstat -an        # all connections, numeric`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := ss.Options{
			TCP: true,
			UDP: true,
		}
		opts.All, _ = cmd.Flags().GetBool("all")
		opts.Listening, _ = cmd.Flags().GetBool("listening")
		opts.TCP, _ = cmd.Flags().GetBool("tcp")
		opts.UDP, _ = cmd.Flags().GetBool("udp")
		opts.Processes, _ = cmd.Flags().GetBool("processes")
		opts.Numeric, _ = cmd.Flags().GetBool("numeric")
		opts.OutputFormat = getOutputOpts(cmd).GetFormat()

		return ss.Run(cmd.OutOrStdout(), opts)
	},
}

func init() {
	rootCmd.AddCommand(ssCmd)
	rootCmd.AddCommand(netstatCmd)

	// ss flags
	ssCmd.Flags().BoolP("all", "a", false, "display all sockets")
	ssCmd.Flags().BoolP("listening", "l", false, "display listening sockets only")
	ssCmd.Flags().BoolP("tcp", "t", false, "display TCP sockets")
	ssCmd.Flags().BoolP("udp", "u", false, "display UDP sockets")
	ssCmd.Flags().BoolP("unix", "x", false, "display Unix sockets")
	ssCmd.Flags().BoolP("processes", "p", false, "show process using socket")
	ssCmd.Flags().BoolP("numeric", "n", false, "don't resolve service names")
	ssCmd.Flags().BoolP("ipv4", "4", false, "display only IPv4 sockets")
	ssCmd.Flags().BoolP("ipv6", "6", false, "display only IPv6 sockets")
	ssCmd.Flags().BoolP("summary", "s", false, "print summary statistics")
	ssCmd.Flags().BoolP("extended", "e", false, "show extended socket info")
	ssCmd.Flags().Bool("no-header", false, "don't print headers")
	ssCmd.Flags().String("state", "", "filter by state (established, listen, time_wait, etc.)")

	// netstat flags (subset)
	netstatCmd.Flags().BoolP("all", "a", false, "display all sockets")
	netstatCmd.Flags().BoolP("listening", "l", false, "display listening sockets only")
	netstatCmd.Flags().BoolP("tcp", "t", false, "display TCP sockets")
	netstatCmd.Flags().BoolP("udp", "u", false, "display UDP sockets")
	netstatCmd.Flags().BoolP("processes", "p", false, "show process using socket")
	netstatCmd.Flags().BoolP("numeric", "n", false, "don't resolve service names")
}
