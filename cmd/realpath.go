package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli"

	"github.com/spf13/cobra"
)

// realpathCmd represents the realpath command
var realpathCmd = &cobra.Command{
	Use:   "realpath [path...]",
	Short: "Print the resolved path",
	Long:  `Print the resolved absolute file name; all components must exist.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.RunRealpath(os.Stdout, args)
	},
}

func init() {
	rootCmd.AddCommand(realpathCmd)
}
