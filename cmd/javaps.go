package cmd

import (
	"github.com/spf13/cobra"

	"github.com/inovacc/omni/pkg/procutil"
)

var javapsCmd = &cobra.Command{
	Use:   "javaps",
	Short: "List and signal running Java (JVM) processes",
	Long: `Enumerate JVM processes by matching executable basename (java, javaw).

Examples:
  omni javaps                              # list JVM processes
  omni javaps --json
  omni javaps kill 12345
  omni javaps kill java --recursive --yes  # signal every java process
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRuntimePsList(cmd, procutil.RuntimeJava)
	},
}

var javapsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Java (JVM) processes",
	Long: `List running Java (JVM) processes.

Examples:
  omni javaps list                # list JVM processes
  omni javaps list --json         # list as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error { return runRuntimePsList(cmd, procutil.RuntimeJava) },
}

var javapsKillCmd = &cobra.Command{
	Use:   "kill <pid|name>",
	Short: "Signal one or more Java processes",
	Long: `Signal a Java process by PID or by name.

Examples:
  omni javaps kill 12345          # signal by PID
  omni javaps kill java --recursive --yes  # signal every match`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRuntimePsKill(cmd, procutil.RuntimeJava, args[0])
	},
}

func init() {
	registerRuntimePsFlags(javapsCmd, javapsListCmd, javapsKillCmd)
	javapsCmd.AddCommand(javapsListCmd, javapsKillCmd)
	rootCmd.AddCommand(javapsCmd)
}
