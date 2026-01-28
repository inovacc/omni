package cmd

import (
	"github.com/inovacc/omni/pkg/cli/copy"

	"github.com/spf13/cobra"
)

// mvCmd represents the mv command
var mvCmd = &cobra.Command{
	Use:     "mv [source...] [destination]",
	Aliases: []string{"move"},
	Short:   "Move (rename) files",
	Long:    `Rename SOURCE to DEST, or move SOURCE(s) to DIRECTORY.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return copy.RunMove(args, copy.MoveOptions{})
	},
}

func init() {
	rootCmd.AddCommand(mvCmd)
}
