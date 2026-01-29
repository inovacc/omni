package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/whoami"
	"github.com/spf13/cobra"
)

// whoamiCmd represents the whoami command
var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Print effective username",
	Long:  `Print the user name associated with the current effective user ID.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return whoami.RunWhoami(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}
