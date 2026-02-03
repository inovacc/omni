package cmd

import (
	"github.com/inovacc/omni/internal/cli/rg"
	"github.com/spf13/cobra"
)

var rgCmd = &cobra.Command{
	Use:   "rg [OPTIONS] PATTERN [PATH...]",
	Short: "Recursively search for a pattern (ripgrep-style)",
	Long: `Recursively search current directory for a regex pattern.

rg is a line-oriented search tool that recursively searches the current
directory for a regex pattern. By default, rg respects gitignore rules
and automatically skips hidden files/directories and binary files.

This is inspired by ripgrep (https://github.com/BurntSushi/ripgrep).

Examples:
  # Search for pattern in current directory
  omni rg "pattern"

  # Search in specific directory
  omni rg "pattern" ./src

  # Case insensitive search
  omni rg -i "pattern"

  # Search only Go files
  omni rg -t go "func main"

  # Search with context (3 lines before and after)
  omni rg -C 3 "error"

  # Show only filenames with matches
  omni rg -l "TODO"

  # Count matches per file
  omni rg -c "pattern"

  # Include hidden files
  omni rg --hidden "pattern"

  # Don't respect gitignore
  omni rg --no-ignore "pattern"

  # Search for literal string (no regex)
  omni rg -F "func()"

  # JSON output
  omni rg --json "pattern"

  # Glob patterns
  omni rg -g "*.go" -g "!*_test.go" "pattern"

File Types:
  go, js, ts, py, rust, c, cpp, java, rb, php, sh, json, yaml, toml,
  xml, html, css, md, sql, proto, dockerfile, make, txt`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := rg.Options{}

		opts.IgnoreCase, _ = cmd.Flags().GetBool("ignore-case")
		opts.SmartCase, _ = cmd.Flags().GetBool("smart-case")
		opts.WordRegexp, _ = cmd.Flags().GetBool("word-regexp")
		opts.LineNumber, _ = cmd.Flags().GetBool("line-number")
		opts.Count, _ = cmd.Flags().GetBool("count")
		opts.FilesWithMatch, _ = cmd.Flags().GetBool("files-with-matches")
		opts.InvertMatch, _ = cmd.Flags().GetBool("invert-match")
		opts.Context, _ = cmd.Flags().GetInt("context")
		opts.Before, _ = cmd.Flags().GetInt("before-context")
		opts.After, _ = cmd.Flags().GetInt("after-context")
		opts.Types, _ = cmd.Flags().GetStringSlice("type")
		opts.TypesNot, _ = cmd.Flags().GetStringSlice("type-not")
		opts.Glob, _ = cmd.Flags().GetStringSlice("glob")
		opts.Hidden, _ = cmd.Flags().GetBool("hidden")
		opts.NoIgnore, _ = cmd.Flags().GetBool("no-ignore")
		opts.MaxCount, _ = cmd.Flags().GetInt("max-count")
		opts.MaxDepth, _ = cmd.Flags().GetInt("max-depth")
		opts.FollowSymlinks, _ = cmd.Flags().GetBool("follow")
		opts.JSON, _ = cmd.Flags().GetBool("json")
		opts.NoHeading, _ = cmd.Flags().GetBool("no-heading")
		opts.OnlyMatching, _ = cmd.Flags().GetBool("only-matching")
		opts.Quiet, _ = cmd.Flags().GetBool("quiet")
		opts.Fixed, _ = cmd.Flags().GetBool("fixed-strings")

		pattern := args[0]
		paths := args[1:]

		return rg.Run(cmd.OutOrStdout(), pattern, paths, opts)
	},
}

func init() {
	rootCmd.AddCommand(rgCmd)

	// Case sensitivity
	rgCmd.Flags().BoolP("ignore-case", "i", false, "case insensitive search")
	rgCmd.Flags().BoolP("smart-case", "S", false, "smart case (insensitive if pattern is all lowercase)")
	rgCmd.Flags().BoolP("word-regexp", "w", false, "only match whole words")
	rgCmd.Flags().BoolP("fixed-strings", "F", false, "treat pattern as literal string")

	// Output control
	rgCmd.Flags().BoolP("line-number", "n", false, "show line numbers")
	rgCmd.Flags().BoolP("count", "c", false, "only show count of matches per file")
	rgCmd.Flags().BoolP("files-with-matches", "l", false, "only show file names with matches")
	rgCmd.Flags().BoolP("invert-match", "v", false, "show non-matching lines")
	rgCmd.Flags().BoolP("only-matching", "o", false, "show only matching part of line")
	rgCmd.Flags().BoolP("no-heading", "H", false, "don't group matches by file name")
	rgCmd.Flags().BoolP("quiet", "q", false, "quiet mode, exit on first match")
	rgCmd.Flags().Bool("json", false, "output results as JSON")

	// Context
	rgCmd.Flags().IntP("context", "C", 0, "show N lines before and after match")
	rgCmd.Flags().IntP("before-context", "B", 0, "show N lines before match")
	rgCmd.Flags().IntP("after-context", "A", 0, "show N lines after match")

	// File filtering
	rgCmd.Flags().StringSliceP("type", "t", nil, "only search files of TYPE (go, js, py, etc.)")
	rgCmd.Flags().StringSliceP("type-not", "T", nil, "exclude files of TYPE")
	rgCmd.Flags().StringSliceP("glob", "g", nil, "include/exclude files matching GLOB (prefix with ! to exclude)")

	// Directory control
	rgCmd.Flags().Bool("hidden", false, "search hidden files and directories")
	rgCmd.Flags().Bool("no-ignore", false, "don't respect gitignore files")
	rgCmd.Flags().IntP("max-count", "m", 0, "limit matches per file")
	rgCmd.Flags().Int("max-depth", 0, "limit directory traversal depth")
	rgCmd.Flags().BoolP("follow", "L", false, "follow symbolic links")
}
