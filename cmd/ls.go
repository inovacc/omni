package cmd

import (
	"github.com/inovacc/goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls [directory]",
	Short: "List directory contents",
	Long:  `List information about the FILEs (the current directory by default).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonMode, _ := cmd.Flags().GetBool("json")
		return cli.RunLs(args, jsonMode)
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)

	lsCmd.Flags().Bool("json", false, "Output in JSON format")
}
