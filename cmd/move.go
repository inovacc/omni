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
	Long: `move is an alias for the mv command.

Usage:
  omni move SOURCE DEST

See 'omni mv --help' for full options.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mvCmd.RunE(cmd, args)
	},
}
