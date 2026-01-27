package cmd

import (
	"github.com/inovacc/goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// mvCmd represents the mv command
var mvCmd = &cobra.Command{
	Use:     "mv [source...] [destination]",
	Aliases: []string{"move"},
	Short:   "Move (rename) files",
	Long:    `Rename SOURCE to DEST, or move SOURCE(s) to DIRECTORY.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.RunMove(args)
	},
}

func init() {
	rootCmd.AddCommand(mvCmd)
}
