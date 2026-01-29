package cmd

import (
	"github.com/inovacc/omni/internal/cli/rm"
	"github.com/spf13/cobra"
)

// rmdirCmd represents the rmdir command
var rmdirCmd = &cobra.Command{
	Use:   "rmdir [directory...]",
	Short: "Remove empty directories",
	Long: `Remove the DIRECTORY(ies), if they are empty.

Protected paths (system directories, credentials, etc.) cannot be
deleted without the --no-preserve-root flag.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		noPreserveRoot, _ := cmd.Flags().GetBool("no-preserve-root")
		return rm.RunRmdir(args, rm.RmdirOptions{
			NoPreserveRoot: noPreserveRoot,
		})
	},
}

func init() {
	rootCmd.AddCommand(rmdirCmd)

	rmdirCmd.Flags().Bool("no-preserve-root", false, "do not treat protected paths specially (dangerous)")
}
