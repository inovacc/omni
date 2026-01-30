package cmd

import (
	"strings"

	"github.com/inovacc/omni/internal/cli/tagfixer"
	"github.com/spf13/cobra"
)

// tagfixerCmd represents the tagfixer command
var tagfixerCmd = &cobra.Command{
	Use:   "tagfixer [PATH]",
	Short: "Fix and standardize Go struct tags",
	Long: `Fix and standardize struct tags in Go files.

Supports multiple casing conventions: camelCase, PascalCase, kebab-case, snake_case.

Case Types:
  camel   - camelCase (default)
  pascal  - PascalCase
  snake   - snake_case
  kebab   - kebab-case

Flags:
  -c, --case        Target case type (default: camel)
  -t, --tags        Tag types to fix (default: json)
  -r, --recursive   Process directories recursively
  -d, --dry-run     Preview changes without writing
  -a, --analyze     Analyze mode - generate report only
  -v, --verbose     Verbose output
  --json            Output as JSON

Examples:
  omni tagfixer                           # Fix json tags in current dir (camelCase)
  omni tagfixer ./pkg                     # Fix in specific directory
  omni tagfixer -c snake                  # Use snake_case
  omni tagfixer -c kebab                  # Use kebab-case
  omni tagfixer -t json,yaml,xml          # Fix multiple tag types
  omni tagfixer -d                        # Dry-run (preview)
  omni tagfixer -a                        # Analyze only
  omni tagfixer -a -v                     # Detailed analysis
  omni tagfixer --json                    # JSON output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := tagfixer.Options{}

		caseStr, _ := cmd.Flags().GetString("case")
		if caseStr != "" {
			caseType, err := tagfixer.ParseCaseType(caseStr)
			if err != nil {
				return err
			}
			opts.Case = caseType
		}

		tagsStr, _ := cmd.Flags().GetString("tags")
		if tagsStr != "" {
			opts.Tags = strings.Split(tagsStr, ",")
		}

		opts.DryRun, _ = cmd.Flags().GetBool("dry-run")
		opts.Recursive, _ = cmd.Flags().GetBool("recursive")
		opts.Analyze, _ = cmd.Flags().GetBool("analyze")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		if len(args) > 0 {
			opts.Path = args[0]
		}

		return tagfixer.RunTagFixer(cmd.OutOrStdout(), opts)
	},
}

// tagfixerAnalyzeCmd is an alias for tagfixer --analyze
var tagfixerAnalyzeCmd = &cobra.Command{
	Use:     "analyze [PATH]",
	Aliases: []string{"report"},
	Short:   "Analyze struct tag usage patterns",
	Long: `Analyze Go files to understand current struct tag patterns.

Generates a report showing:
  - Total files, structs, and fields analyzed
  - Tag type statistics (json, yaml, xml, etc.)
  - Case type distribution
  - Consistency score (0-100%)
  - Recommended case type based on existing patterns

Examples:
  omni tagfixer analyze                   # Analyze current directory
  omni tagfixer analyze ./pkg             # Analyze specific directory
  omni tagfixer analyze -r                # Recursive analysis
  omni tagfixer analyze -v                # Verbose (show all files)
  omni tagfixer analyze --json            # JSON output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := tagfixer.Options{
			Analyze: true,
		}

		tagsStr, _ := cmd.Flags().GetString("tags")
		if tagsStr != "" {
			opts.Tags = strings.Split(tagsStr, ",")
		}

		opts.Recursive, _ = cmd.Flags().GetBool("recursive")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		if len(args) > 0 {
			opts.Path = args[0]
		}

		return tagfixer.RunTagFixer(cmd.OutOrStdout(), opts)
	},
}

func init() {
	rootCmd.AddCommand(tagfixerCmd)

	// Add subcommand
	tagfixerCmd.AddCommand(tagfixerAnalyzeCmd)

	// Main command flags
	tagfixerCmd.Flags().StringP("case", "c", "camel", "target case type (camel, pascal, snake, kebab)")
	tagfixerCmd.Flags().StringP("tags", "t", "json", "comma-separated tag types to fix")
	tagfixerCmd.Flags().BoolP("recursive", "r", true, "process directories recursively")
	tagfixerCmd.Flags().BoolP("dry-run", "d", false, "preview changes without writing")
	tagfixerCmd.Flags().BoolP("analyze", "a", false, "analyze mode - generate report only")
	tagfixerCmd.Flags().BoolP("verbose", "v", false, "verbose output")
	tagfixerCmd.Flags().Bool("json", false, "output as JSON")

	// Analyze subcommand flags
	tagfixerAnalyzeCmd.Flags().StringP("tags", "t", "json,yaml,xml", "comma-separated tag types to analyze")
	tagfixerAnalyzeCmd.Flags().BoolP("recursive", "r", true, "process directories recursively")
	tagfixerAnalyzeCmd.Flags().BoolP("verbose", "v", false, "verbose output (show all files)")
	tagfixerAnalyzeCmd.Flags().Bool("json", false, "output as JSON")
}
