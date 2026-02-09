package cmd

import (
	"github.com/inovacc/omni/internal/cli/project"
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Analyze project structure, dependencies, and health",
	Long: `Project analyzer that inspects any codebase directory and outputs
a structured report covering project type detection, language identification,
dependency parsing, git status, and documentation checks.

Subcommands:
  info     Full project overview (type, deps, git, docs)
  deps     Dependency analysis only
  docs     Documentation status check
  git      Git repository info
  health   Health score (0-100) with grade

Examples:
  omni project info
  omni project info --json
  omni project deps /path/to/project
  omni project health --markdown
  omni project git`,
}

var projectInfoCmd = &cobra.Command{
	Use:   "info [path]",
	Short: "Full project overview",
	Long: `Analyze the project directory and show a complete overview including
project type, languages, dependencies, git info, and documentation status.

Examples:
  omni project info
  omni project info --json
  omni project info --markdown
  omni project info /path/to/project`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return project.RunInfo(cmd.OutOrStdout(), args, projectOpts(cmd))
	},
}

var projectDepsCmd = &cobra.Command{
	Use:   "deps [path]",
	Short: "Dependency analysis",
	Long: `Analyze project dependencies. Supports Go (go.mod), Node.js (package.json),
Python (requirements.txt, pyproject.toml), Rust (Cargo.toml), Java (pom.xml,
build.gradle), Ruby (Gemfile), PHP (composer.json), and .NET (*.csproj).

Examples:
  omni project deps
  omni project deps --json
  omni project deps /path/to/project`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return project.RunDeps(cmd.OutOrStdout(), args, projectOpts(cmd))
	},
}

var projectDocsCmd = &cobra.Command{
	Use:   "docs [path]",
	Short: "Documentation status check",
	Long: `Check documentation status of the project: README, LICENSE, CHANGELOG,
CONTRIBUTING, CI/CD configs, linter configs, and more.

Examples:
  omni project docs
  omni project docs --json
  omni project docs /path/to/project`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return project.RunDocs(cmd.OutOrStdout(), args, projectOpts(cmd))
	},
}

var projectGitCmd = &cobra.Command{
	Use:   "git [path]",
	Short: "Git repository info",
	Long: `Show git repository information including branch, remote, status,
recent commits, tags, and contributor count.

Examples:
  omni project git
  omni project git --json
  omni project git -n 20`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return project.RunGit(cmd.OutOrStdout(), args, projectOpts(cmd))
	},
}

var projectHealthCmd = &cobra.Command{
	Use:   "health [path]",
	Short: "Health score (0-100) with grade",
	Long: `Compute a project health score from 0 to 100 based on presence of
README, LICENSE, CI/CD, tests, linter config, and other best practices.

Grade scale: A (90+), B (80+), C (70+), D (60+), F (<60)

Examples:
  omni project health
  omni project health --json
  omni project health --markdown`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return project.RunHealth(cmd.OutOrStdout(), args, projectOpts(cmd))
	},
}

func projectOpts(cmd *cobra.Command) project.Options {
	opts := project.Options{}
	opts.JSON, _ = cmd.Flags().GetBool("json")
	opts.Markdown, _ = cmd.Flags().GetBool("markdown")
	opts.Verbose, _ = cmd.Flags().GetBool("verbose")
	opts.Limit, _ = cmd.Flags().GetInt("limit")
	return opts
}

func init() {
	rootCmd.AddCommand(projectCmd)

	projectCmd.AddCommand(projectInfoCmd)
	projectCmd.AddCommand(projectDepsCmd)
	projectCmd.AddCommand(projectDocsCmd)
	projectCmd.AddCommand(projectGitCmd)
	projectCmd.AddCommand(projectHealthCmd)

	projectCmd.PersistentFlags().Bool("json", false, "output as JSON")
	projectCmd.PersistentFlags().Bool("markdown", false, "output as Markdown")
	projectCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	projectCmd.PersistentFlags().IntP("limit", "n", 10, "limit for lists (e.g., recent commits)")
}
