package cmd

import (
	"github.com/inovacc/omni/internal/cli/mkdir"
	"github.com/spf13/cobra"
)

// mkdirCmd represents the mkdir command
var mkdirCmd = &cobra.Command{
	Use:   "mkdir [directory...]",
	Short: "Create directories",
	Long:  `Create the DIRECTORY(ies), if they do not already exist.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		parents, _ := cmd.Flags().GetBool("parents")
		return mkdir.RunMkdir(args, mkdir.Options{Parents: parents})
	},
}

func init() {
	rootCmd.AddCommand(mkdirCmd)

	mkdirCmd.Flags().BoolP("parents", "p", false, "no error if existing, make parent directories as needed")
}
