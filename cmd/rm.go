package cmd

import (
	"github.com/inovacc/omni/internal/cli/rm"
	"github.com/spf13/cobra"
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:     "rm [file...]",
	Aliases: []string{"remove"},
	Short:   "Remove files or directories",
	Long: `Remove the FILE(s).

Protected paths (system directories, SSH keys, credentials, etc.) cannot be
deleted without explicit override flags. Use --force for non-critical
protected paths, or --no-preserve-root for critical system paths.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		recursive, _ := cmd.Flags().GetBool("recursive")
		force, _ := cmd.Flags().GetBool("force")
		noPreserveRoot, _ := cmd.Flags().GetBool("no-preserve-root")
		return rm.RunRm(args, rm.RmOptions{
			Recursive:      recursive,
			Force:          force,
			NoPreserveRoot: noPreserveRoot,
		})
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)

	rmCmd.Flags().BoolP("recursive", "r", false, "remove directories and their contents recursively")
	rmCmd.Flags().BoolP("force", "f", false, "ignore nonexistent files and arguments, never prompt")
	rmCmd.Flags().Bool("no-preserve-root", false, "do not treat protected paths specially (dangerous)")
}
