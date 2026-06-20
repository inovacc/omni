package cmd

import (
	copy2 "github.com/inovacc/omni/internal/cli/copy"
	"github.com/spf13/cobra"
)

// mvCmd represents the mv command
var mvCmd = &cobra.Command{
	Use:     "mv [source...] [destination]",
	Aliases: []string{"move"},
	Short:   "Move (rename) files",
	Long: `Rename SOURCE to DEST, or move SOURCE(s) to DIRECTORY.

Examples:
  omni mv old.txt new.txt      # rename a file
  omni mv a.txt b.txt dir/     # move multiple files into a directory
  omni move src dest           # move/rename (alias)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return copy2.RunMove(args, copy2.MoveOptions{})
	},
}

func init() {
	rootCmd.AddCommand(mvCmd)
}
