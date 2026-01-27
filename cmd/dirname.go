package cmd

import (
	"os"

	"github.com/inovacc/goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// dirnameCmd represents the dirname command
var dirnameCmd = &cobra.Command{
	Use:   "dirname [path...]",
	Short: "Strip last component from file name",
	Long:  `Output each NAME with its last non-slash component and trailing slashes removed.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.RunDirname(os.Stdout, args)
	},
}

func init() {
	rootCmd.AddCommand(dirnameCmd)
}
