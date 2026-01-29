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
	Long:    `Remove the FILE(s).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		recursive, _ := cmd.Flags().GetBool("recursive")
		force, _ := cmd.Flags().GetBool("force")
		return rm.RunRm(args, rm.RmOptions{Recursive: recursive, Force: force})
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)

	rmCmd.Flags().BoolP("recursive", "r", false, "remove directories and their contents recursively")
	rmCmd.Flags().BoolP("force", "f", false, "ignore nonexistent files and arguments, never prompt")
}
