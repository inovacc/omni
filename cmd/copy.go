package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(copyCmd)
}

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Alias for cp",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cpCmd.RunE(cmd, args)
	},
}
