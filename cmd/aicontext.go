package cmd

import (
	"fmt"
	"os"

	"github.com/inovacc/omni/internal/cli/aicontext"
	"github.com/inovacc/omni/internal/flags"
	"github.com/spf13/cobra"
)

var aicontextCmd = &cobra.Command{
	Use:   "aicontext",
	Short: "Generate AI context for coding agents",
	Long: `Generate concise context for AI coding agents.

Examples:
  omni aicontext              # Markdown output
  omni aicontext --json       # JSON output
  omni aicontext -c text      # Filter by category
  omni aicontext -o ctx.md    # Write to file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := aicontext.Options{}
		opts.JSON, _ = cmd.Flags().GetBool("json")
		opts.Category, _ = cmd.Flags().GetString("category")
		opts.NoStructure, _ = cmd.Flags().GetBool("no-structure")
		outputFile, _ := cmd.Flags().GetString("output")

		if err := flags.IgnoreCommand("aicontext"); err != nil {
			return err
		}

		w := cmd.OutOrStdout()

		if outputFile != "" {
			f, err := os.Create(outputFile)
			if err != nil {
				return fmt.Errorf("aicontext: %w", err)
			}
			defer func() { _ = f.Close() }()
			w = f
		}

		return aicontext.RunAIContext(w, rootCmd, opts)
	},
}

func init() {
	rootCmd.AddCommand(aicontextCmd)
	aicontextCmd.Flags().BoolP("json", "j", false, "JSON output")
	aicontextCmd.Flags().StringP("category", "c", "", "filter category")
	aicontextCmd.Flags().StringP("output", "o", "", "write to file")
	aicontextCmd.Flags().Bool("no-structure", false, "omit project structure")
}
