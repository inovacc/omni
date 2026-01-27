package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(moveCmd)
}

var moveCmd = &cobra.Command{
	Use:   "move",
	Short: "Alias for mv",
	RunE: func(cmd *cobra.Command, args []string) error {
		return mvCmd.RunE(cmd, args)
	},
}
