package cmd

import (
	"github.com/inovacc/omni/pkg/cli"

	"github.com/spf13/cobra"
)

// cpCmd represents the cp command
var cpCmd = &cobra.Command{
	Use:     "cp [source...] [destination]",
	Aliases: []string{"copy"},
	Short:   "Copy files and directories",
	Long:    `Copy SOURCE to DEST, or multiple SOURCE(s) to DIRECTORY.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.RunCopy(args)
	},
}

func init() {
	rootCmd.AddCommand(cpCmd)
}
