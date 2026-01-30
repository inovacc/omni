package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// pipelineCmd represents the pipeline command
var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Internal streaming pipeline engine",
	Long:  `Internal streaming pipeline engine for chaining commands.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "pipeline: not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pipelineCmd)
}
