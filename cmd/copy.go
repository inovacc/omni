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
	Long: `copy is an alias for the cp command.

Usage:
  omni copy SOURCE DEST

See 'omni cp --help' for full options.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cpCmd.RunE(cmd, args)
	},
}
