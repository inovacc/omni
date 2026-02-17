package cmd

import (
	"github.com/inovacc/omni/internal/cli/repo"
	"github.com/spf13/cobra"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Repository analysis tools",
	Long: `Repository analysis tools for generating structured context
about any codebase, optimized for LLM consumption.

Subcommands:
  analyze   Generate comprehensive repository context

Examples:
  omni repo analyze .
  omni repo analyze /path/to/project
  omni repo analyze github.com/owner/repo
  omni repo analyze . --compact
  omni repo analyze . --json
  omni repo analyze . -o context.md
  omni repo analyze . --sections=tree,deps,keys`,
}

var repoAnalyzeCmd = &cobra.Command{
	Use:   "analyze [path|url]",
	Short: "Generate comprehensive repository context",
	Long: `Analyze a repository and produce structured Markdown or JSON context
optimized for LLM consumption. Includes directory tree, key file contents,
dependencies, entry points, architecture patterns, API surface, git info,
test patterns, and CI/CD configuration.

Supports local paths and remote GitHub repositories (cloned to temp dir).

Sections: overview, tree, keys, deps, api, git, tests, ci

Examples:
  omni repo analyze .                          # Local
  omni repo analyze /path/to/project           # Local absolute
  omni repo analyze github.com/owner/repo      # Remote (clones to temp)
  omni repo analyze . --compact                # Shorter output
  omni repo analyze . --json                   # JSON format
  omni repo analyze . -o context.md            # Write to file
  omni repo analyze . --sections=tree,deps     # Specific sections only`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := repo.Options{}
		opts.JSON, _ = cmd.Flags().GetBool("json")
		opts.Compact, _ = cmd.Flags().GetBool("compact")
		opts.Sections, _ = cmd.Flags().GetString("sections")
		opts.Output, _ = cmd.Flags().GetString("output")

		return repo.RunAnalyze(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(repoCmd)
	repoCmd.AddCommand(repoAnalyzeCmd)

	repoAnalyzeCmd.Flags().Bool("compact", false, "shorter output for smaller context windows")
	repoAnalyzeCmd.Flags().String("sections", "", "comma-separated section filter (overview,tree,keys,deps,api,git,tests,ci)")
	repoAnalyzeCmd.Flags().StringP("output", "o", "", "write output to file")
}
