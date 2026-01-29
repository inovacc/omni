package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/grep"
	"github.com/spf13/cobra"
)

// grepCmd represents the grep command
var grepCmd = &cobra.Command{
	Use:   "grep [options] PATTERN [FILE...]",
	Short: "Print lines that match patterns",
	Long: `Search for PATTERN in each FILE.
When FILE is '-', read standard input.
With no FILE, read '.' if recursive; otherwise, read standard input.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := grep.GrepOptions{}

		opts.IgnoreCase, _ = cmd.Flags().GetBool("ignore-case")
		opts.InvertMatch, _ = cmd.Flags().GetBool("invert-match")
		opts.LineNumber, _ = cmd.Flags().GetBool("line-number")
		opts.Count, _ = cmd.Flags().GetBool("count")
		opts.FilesWithMatch, _ = cmd.Flags().GetBool("files-with-matches")
		opts.FilesNoMatch, _ = cmd.Flags().GetBool("files-without-match")
		opts.OnlyMatching, _ = cmd.Flags().GetBool("only-matching")
		opts.Quiet, _ = cmd.Flags().GetBool("quiet")
		opts.WithFilename, _ = cmd.Flags().GetBool("with-filename")
		opts.NoFilename, _ = cmd.Flags().GetBool("no-filename")
		opts.ExtendedRegexp, _ = cmd.Flags().GetBool("extended-regexp")
		opts.FixedStrings, _ = cmd.Flags().GetBool("fixed-strings")
		opts.WordRegexp, _ = cmd.Flags().GetBool("word-regexp")
		opts.LineRegexp, _ = cmd.Flags().GetBool("line-regexp")
		opts.Context, _ = cmd.Flags().GetInt("context")
		opts.BeforeContext, _ = cmd.Flags().GetInt("before-context")
		opts.AfterContext, _ = cmd.Flags().GetInt("after-context")
		opts.MaxCount, _ = cmd.Flags().GetInt("max-count")
		opts.Recursive, _ = cmd.Flags().GetBool("recursive")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		pattern := args[0]
		files := args[1:]

		return grep.RunGrep(os.Stdout, pattern, files, opts)
	},
}

func init() {
	rootCmd.AddCommand(grepCmd)

	// Pattern matching options
	grepCmd.Flags().BoolP("extended-regexp", "E", false, "interpret PATTERN as an extended regular expression")
	grepCmd.Flags().BoolP("fixed-strings", "F", false, "interpret PATTERN as fixed strings")
	grepCmd.Flags().BoolP("ignore-case", "i", false, "ignore case distinctions in patterns and data")
	grepCmd.Flags().BoolP("word-regexp", "w", false, "match only whole words")
	grepCmd.Flags().BoolP("line-regexp", "x", false, "match only whole lines")

	// Matching control
	grepCmd.Flags().BoolP("invert-match", "v", false, "select non-matching lines")
	grepCmd.Flags().IntP("max-count", "m", 0, "stop after NUM matches")

	// Output control
	grepCmd.Flags().BoolP("count", "c", false, "only print a count of matching lines per FILE")
	grepCmd.Flags().BoolP("files-with-matches", "l", false, "only print FILE names containing matches")
	grepCmd.Flags().BoolP("files-without-match", "L", false, "only print FILE names not containing matches")
	grepCmd.Flags().BoolP("line-number", "n", false, "prefix each line of output with line number")
	grepCmd.Flags().BoolP("only-matching", "o", false, "show only nonempty parts of lines that match")
	grepCmd.Flags().BoolP("quiet", "q", false, "suppress all normal output")
	grepCmd.Flags().BoolP("with-filename", "H", false, "print file name with output lines")
	grepCmd.Flags().Bool("no-filename", false, "suppress the file name prefix on output")

	// Context control
	grepCmd.Flags().IntP("after-context", "A", 0, "print NUM lines of trailing context")
	grepCmd.Flags().IntP("before-context", "B", 0, "print NUM lines of leading context")
	grepCmd.Flags().IntP("context", "C", 0, "print NUM lines of output context")

	// File and directory selection
	grepCmd.Flags().BoolP("recursive", "r", false, "search directories recursively")

	// Output format
	grepCmd.Flags().Bool("json", false, "output as JSON")
}
