package cmd

import (
	copy2 "github.com/inovacc/omni/internal/cli/copy"
	"github.com/spf13/cobra"
)

// mvCmd represents the mv command
var mvCmd = &cobra.Command{
	Use:     "mv [source...] [destination]",
	Aliases: []string{"move"},
	Short:   "Move (rename) files",
	Long:    `Rename SOURCE to DEST, or move SOURCE(s) to DIRECTORY.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return copy2.RunMove(args, copy2.MoveOptions{})
	},
}

func init() {
	rootCmd.AddCommand(mvCmd)
}
