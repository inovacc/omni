package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli"

	"github.com/spf13/cobra"
)

// whoamiCmd represents the whoami command
var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Print effective username",
	Long:  `Print the user name associated with the current effective user ID.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.RunWhoami(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}
