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
	Long: `remove is an alias for the rm command.

Usage:
  omni remove FILE [FILE...]

See 'omni rm --help' for full options.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return rmCmd.RunE(cmd, args)
	},
}
