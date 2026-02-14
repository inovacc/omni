package cmd

import (
	"github.com/inovacc/omni/internal/cli/stat"
	"github.com/spf13/cobra"
)

// statCmd represents the stat command
var statCmd = &cobra.Command{
	Use:   "stat [file...]",
	Short: "Display file or file system status",
	Long:  `Display file or file system status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return stat.RunStat(cmd.OutOrStdout(), args, stat.StatOptions{OutputFormat: getOutputOpts(cmd).GetFormat()})
	},
}

func init() {
	rootCmd.AddCommand(statCmd)
}
