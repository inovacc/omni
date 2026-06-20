package cmd

import (
	"github.com/spf13/cobra"

	"github.com/inovacc/omni/pkg/procutil"
)

var nodepsCmd = &cobra.Command{
	Use:   "nodeps",
	Short: "List and signal running Node.js processes",
	Long: `Enumerate Node.js processes by matching executable basename (node, nodejs).

Examples:
  omni nodeps                              # list Node.js processes
  omni nodeps --json
  omni nodeps kill 12345
  omni nodeps kill node --recursive --yes  # signal every node process
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRuntimePsList(cmd, procutil.RuntimeNode)
	},
}

var nodepsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Node.js processes",
	Long: `List running Node.js processes.

Examples:
  omni nodeps list                # list Node.js processes
  omni nodeps list --json         # list as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error { return runRuntimePsList(cmd, procutil.RuntimeNode) },
}

var nodepsKillCmd = &cobra.Command{
	Use:   "kill <pid|name>",
	Short: "Signal one or more Node.js processes",
	Long: `Signal a Node.js process by PID or by name.

Examples:
  omni nodeps kill 12345          # signal by PID
  omni nodeps kill node --recursive --yes  # signal every match`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRuntimePsKill(cmd, procutil.RuntimeNode, args[0])
	},
}

func init() {
	registerRuntimePsFlags(nodepsCmd, nodepsListCmd, nodepsKillCmd)
	nodepsCmd.AddCommand(nodepsListCmd, nodepsKillCmd)
	rootCmd.AddCommand(nodepsCmd)
}
