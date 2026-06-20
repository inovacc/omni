package cmd

import (
	"github.com/spf13/cobra"

	"github.com/inovacc/omni/pkg/procutil"
)

var pypsCmd = &cobra.Command{
	Use:   "pyps",
	Short: "List and signal running Python processes",
	Long: `Enumerate Python processes by matching executable basename (python, python3, python3.X, pythonw).

Examples:
  omni pyps                                # list Python processes
  omni pyps --json
  omni pyps kill 12345
  omni pyps kill python --recursive --yes  # signal every python process
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRuntimePsList(cmd, procutil.RuntimePython)
	},
}

var pypsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Python processes",
	Long: `List running Python processes.

Examples:
  omni pyps list                  # list Python processes
  omni pyps list --json           # list as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error { return runRuntimePsList(cmd, procutil.RuntimePython) },
}

var pypsKillCmd = &cobra.Command{
	Use:   "kill <pid|name>",
	Short: "Signal one or more Python processes",
	Long: `Signal a Python process by PID or by name.

Examples:
  omni pyps kill 12345            # signal by PID
  omni pyps kill python --recursive --yes  # signal every match`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRuntimePsKill(cmd, procutil.RuntimePython, args[0])
	},
}

func init() {
	registerRuntimePsFlags(pypsCmd, pypsListCmd, pypsKillCmd)
	pypsCmd.AddCommand(pypsListCmd, pypsKillCmd)
	rootCmd.AddCommand(pypsCmd)
}
