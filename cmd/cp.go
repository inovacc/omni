package cmd

import (
	copy2 "github.com/inovacc/omni/internal/cli/copy"
	"github.com/spf13/cobra"
)

// cpCmd represents the cp command
var cpCmd = &cobra.Command{
	Use:     "cp [source...] [destination]",
	Aliases: []string{"copy"},
	Short:   "Copy files and directories",
	Long:    `Copy SOURCE to DEST, or multiple SOURCE(s) to DIRECTORY.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return copy2.RunCopy(args, copy2.CopyOptions{})
	},
}

func init() {
	rootCmd.AddCommand(cpCmd)
}
