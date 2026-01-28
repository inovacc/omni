package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli/stat"

	"github.com/spf13/cobra"
)

// statCmd represents the stat command
var statCmd = &cobra.Command{
	Use:   "stat [file...]",
	Short: "Display file or file system status",
	Long:  `Display file or file system status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonMode, _ := cmd.Flags().GetBool("json")
		return stat.RunStat(os.Stdout, args, stat.StatOptions{JSON: jsonMode})
	},
}

func init() {
	rootCmd.AddCommand(statCmd)
	statCmd.Flags().Bool("json", false, "Output in JSON format")
}
