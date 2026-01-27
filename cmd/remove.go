package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Alias for rm",
	RunE: func(cmd *cobra.Command, args []string) error {
		return rmCmd.RunE(cmd, args)
	},
}
