package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// topCmd represents the top command
var topCmd = &cobra.Command{
	Use:   "top",
	Short: "Display system processes (TUI)",
	Long:  `Display and monitor system processes in a TUI interface.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, _ = fmt.Fprintln(os.Stdout, "top: not yet implemented (requires TUI)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(topCmd)
}
