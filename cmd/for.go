package cmd

import (
	"os"
	"strconv"
	"strings"

	"github.com/inovacc/omni/internal/cli/forloop"
	"github.com/spf13/cobra"
)

var forCmd = &cobra.Command{
	Use:   "for",
	Short: "Loop and execute commands",
	Long: `Loop over items and execute commands for each.

Subcommands:
  range   Loop over a numeric range
  each    Loop over a list of items
  lines   Loop over lines from stdin or file
  split   Loop over items split by delimiter
  glob    Loop over files matching a pattern

Variable substitution:
  $item or ${item}   Current item value
  $i or ${i}         Current index (0-based)
  $n or ${n}         Current line number (1-based)
  $file or ${file}   Current file path

Examples:
  omni for range 1 5 -- echo $i
  omni for each a b c -- echo "Item: $item"
  omni for lines file.txt -- echo "Line: $line"
  omni for split "," "a,b,c" -- echo $item
  omni for glob "*.txt" -- cat $file`,
}

var forRangeCmd = &cobra.Command{
	Use:   "range START END [STEP] -- COMMAND",
	Short: "Loop over a numeric range",
	Long: `Loop from START to END (inclusive) with optional STEP.

Variable: $i or ${i}

Examples:
  omni for range 1 5 -- echo $i
  omni for range 10 0 -2 -- echo $i
  omni for range 1 100 -- echo "Number: $i"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find the -- separator
		sepIdx := -1
		for i, arg := range args {
			if arg == "--" {
				sepIdx = i
				break
			}
		}

		if sepIdx < 2 {
			return cmd.Usage()
		}

		start, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		end, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}

		step := 0
		if sepIdx > 2 {
			step, err = strconv.Atoi(args[2])
			if err != nil {
				return err
			}
		}

		command := strings.Join(args[sepIdx+1:], " ")

		variable, _ := cmd.Flags().GetString("var")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		opts := forloop.Options{
			Variable: variable,
			DryRun:   dryRun,
		}

		return forloop.RunRange(cmd.OutOrStdout(), start, end, step, command, opts)
	},
}

var forEachCmd = &cobra.Command{
	Use:   "each ITEM... -- COMMAND",
	Short: "Loop over a list of items",
	Long: `Loop over each item in the provided list.

Variable: $item or ${item}

Examples:
  omni for each apple banana cherry -- echo "Fruit: $item"
  omni for each *.go -- echo "File: $item"
  omni for each a b c --var=x -- echo "Value: $x"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sepIdx := -1
		for i, arg := range args {
			if arg == "--" {
				sepIdx = i
				break
			}
		}

		if sepIdx < 1 {
			return cmd.Usage()
		}

		items := args[:sepIdx]
		command := strings.Join(args[sepIdx+1:], " ")

		variable, _ := cmd.Flags().GetString("var")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		opts := forloop.Options{
			Variable: variable,
			DryRun:   dryRun,
		}

		return forloop.RunEach(cmd.OutOrStdout(), items, command, opts)
	},
}

var forLinesCmd = &cobra.Command{
	Use:   "lines [FILE] -- COMMAND",
	Short: "Loop over lines from stdin or file",
	Long: `Loop over each line from stdin or a file.

Variables:
  $line or ${line}   Current line content
  $n or ${n}         Current line number (1-based)

Examples:
  cat file.txt | omni for lines -- echo "Line $n: $line"
  omni for lines input.txt -- echo "$n: $line"
  omni for lines --var=x -- echo "Got: $x"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sepIdx := -1
		for i, arg := range args {
			if arg == "--" {
				sepIdx = i
				break
			}
		}

		if sepIdx < 0 {
			return cmd.Usage()
		}

		variable, _ := cmd.Flags().GetString("var")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		opts := forloop.Options{
			Variable: variable,
			DryRun:   dryRun,
		}

		command := strings.Join(args[sepIdx+1:], " ")

		// Check if a file is specified before --
		if sepIdx > 0 {
			file, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer func() { _ = file.Close() }()

			return forloop.RunLines(cmd.OutOrStdout(), file, command, opts)
		}

		return forloop.RunLines(cmd.OutOrStdout(), os.Stdin, command, opts)
	},
}

var forSplitCmd = &cobra.Command{
	Use:   "split DELIMITER INPUT -- COMMAND",
	Short: "Loop over items split by delimiter",
	Long: `Split input by delimiter and loop over each item.

Variables:
  $item or ${item}   Current item
  $i or ${i}         Current index (0-based)

Examples:
  omni for split "," "a,b,c" -- echo "Item: $item"
  omni for split ":" "$PATH" -- echo "Dir: $item"
  omni for split "\\n" "$(cat file.txt)" -- process $item`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sepIdx := -1
		for i, arg := range args {
			if arg == "--" {
				sepIdx = i
				break
			}
		}

		if sepIdx < 2 {
			return cmd.Usage()
		}

		delimiter := args[0]
		input := args[1]
		command := strings.Join(args[sepIdx+1:], " ")

		variable, _ := cmd.Flags().GetString("var")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		opts := forloop.Options{
			Variable: variable,
			DryRun:   dryRun,
		}

		return forloop.RunSplit(cmd.OutOrStdout(), input, delimiter, command, opts)
	},
}

var forGlobCmd = &cobra.Command{
	Use:   "glob PATTERN -- COMMAND",
	Short: "Loop over files matching a pattern",
	Long: `Loop over files matching a glob pattern.

Variable: $file or ${file}

Patterns:
  *.txt       All .txt files in current directory
  **/*.go     All .go files recursively
  src/*.js    All .js files in src/

Examples:
  omni for glob "*.txt" -- cat $file
  omni for glob "**/*.go" -- wc -l $file
  omni for glob "src/*.js" --dry-run -- echo $file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sepIdx := -1
		for i, arg := range args {
			if arg == "--" {
				sepIdx = i
				break
			}
		}

		if sepIdx < 1 {
			return cmd.Usage()
		}

		pattern := args[0]
		command := strings.Join(args[sepIdx+1:], " ")

		variable, _ := cmd.Flags().GetString("var")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		opts := forloop.Options{
			Variable: variable,
			DryRun:   dryRun,
		}

		return forloop.RunGlob(cmd.OutOrStdout(), pattern, command, opts)
	},
}

func init() {
	rootCmd.AddCommand(forCmd)

	forCmd.AddCommand(forRangeCmd)
	forCmd.AddCommand(forEachCmd)
	forCmd.AddCommand(forLinesCmd)
	forCmd.AddCommand(forSplitCmd)
	forCmd.AddCommand(forGlobCmd)

	// Common flags for all subcommands
	for _, cmd := range []*cobra.Command{forRangeCmd, forEachCmd, forLinesCmd, forSplitCmd, forGlobCmd} {
		cmd.Flags().String("var", "", "variable name to use")
		cmd.Flags().Bool("dry-run", false, "print commands without executing")
	}
}
