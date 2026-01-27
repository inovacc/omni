package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// gopsCmd represents the gops command
var gopsCmd = &cobra.Command{
	Use:   "gops",
	Short: "Display Go process information",
	Long:  `Display information about running Go processes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, _ = fmt.Fprintln(os.Stdout, "gops: not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(gopsCmd)
}
