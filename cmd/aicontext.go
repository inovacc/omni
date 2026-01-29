package cmd

import (
	"fmt"
	"os"

	"github.com/inovacc/omni/internal/cli/aicontext"
	"github.com/spf13/cobra"
)

var aicontextCmd = &cobra.Command{
	Use:   "aicontext",
	Short: "Generate AI-optimized context documentation",
	Long: `Generate comprehensive, AI-optimized documentation about the omni application.

Unlike cmdtree which shows a visual tree, aicontext produces a detailed context
document designed for AI consumption, including:

  - Application overview and design principles
  - Command categories with descriptions
  - Complete command reference with all flags
  - Library API usage examples
  - Architecture documentation

Examples:

  # Generate markdown documentation (default)
  omni aicontext

  # Generate JSON for programmatic use
  omni aicontext --json

  # Generate compact output (no examples or long descriptions)
  omni aicontext --compact

  # Filter to a specific category
  omni aicontext --category "Text Processing"

  # Write to a file
  omni aicontext --output context.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := aicontext.Options{}

		opts.JSON, _ = cmd.Flags().GetBool("json")
		opts.Compact, _ = cmd.Flags().GetBool("compact")
		opts.Category, _ = cmd.Flags().GetString("category")
		outputFile, _ := cmd.Flags().GetString("output")

		var w = os.Stdout

		if outputFile != "" {
			f, err := os.Create(outputFile)
			if err != nil {
				return fmt.Errorf("aicontext: %w", err)
			}

			defer func() {
				_ = f.Close()
			}()

			w = f
		}

		return aicontext.RunAIContext(w, rootCmd, opts)
	},
}

func init() {
	rootCmd.AddCommand(aicontextCmd)

	aicontextCmd.Flags().Bool("json", false, "output as structured JSON")
	aicontextCmd.Flags().Bool("compact", false, "omit examples and long descriptions")
	aicontextCmd.Flags().String("category", "", "filter to specific category")
	aicontextCmd.Flags().StringP("output", "o", "", "write to file instead of stdout")
}
