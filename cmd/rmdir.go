package cmd

import (
	"github.com/inovacc/goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// rmdirCmd represents the rmdir command
var rmdirCmd = &cobra.Command{
	Use:   "rmdir [directory...]",
	Short: "Remove empty directories",
	Long:  `Remove the DIRECTORY(ies), if they are empty.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.RunRmdir(args)
	},
}

func init() {
	rootCmd.AddCommand(rmdirCmd)
}
