package cmd

import (
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
		opts := whoami.WhoamiOptions{}
		opts.OutputFormat = getOutputOpts(cmd).GetFormat()
		return whoami.RunWhoami(cmd.OutOrStdout(), opts)
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)

}
