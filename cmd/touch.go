package cmd

import (
	"github.com/inovacc/omni/internal/cli/stat"
	"github.com/spf13/cobra"
)

// touchCmd represents the touch command
var touchCmd = &cobra.Command{
	Use:   "touch [file...]",
	Short: "Update the access and modification times of each FILE to the current time",
	Long:  `Update the access and modification times of each FILE to the current time. A FILE argument that does not exist is created empty.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return stat.RunTouch(args, stat.TouchOptions{})
	},
}

func init() {
	rootCmd.AddCommand(touchCmd)
}
