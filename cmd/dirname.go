package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/dirname"
	"github.com/spf13/cobra"
)

// dirnameCmd represents the dirname command
var dirnameCmd = &cobra.Command{
	Use:   "dirname [path...]",
	Short: "Strip last component from file name",
	Long:  `Output each NAME with its last non-slash component and trailing slashes removed.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := dirname.DirnameOptions{}
		opts.JSON, _ = cmd.Flags().GetBool("json")
		return dirname.RunDirname(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(dirnameCmd)

	dirnameCmd.Flags().Bool("json", false, "output as JSON")
}
