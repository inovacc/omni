package cmd

import (
	"github.com/inovacc/omni/internal/cli/realpath"
	"github.com/spf13/cobra"
)

// realpathCmd represents the realpath command
var realpathCmd = &cobra.Command{
	Use:   "realpath [path...]",
	Short: "Print the resolved path",
	Long:  `Print the resolved absolute file name; all components must exist.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := realpath.RealpathOptions{OutputFormat: getOutputOpts(cmd).GetFormat()}
		return realpath.RunRealpath(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(realpathCmd)
}
