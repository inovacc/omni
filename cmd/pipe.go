package cmd

import (
	"bufio"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/pipe"
	"github.com/spf13/cobra"
)

var pipeCmd = &cobra.Command{
	Use:   "pipe {CMD}, {CMD}, ... | CMD | CMD",
	Short: "Chain omni commands without shell pipes",
	Long: `Chain multiple omni commands together, passing output from one to the next.

This allows creating pipelines of omni commands without using shell pipes,
making scripts more portable and avoiding shell-specific behavior.

Commands can be separated by:
  - Curly braces with commas: {cmd1}, {cmd2}, {cmd3} (recommended)
  - The | character: cmd1 | cmd2 | cmd3
  - A custom separator with --sep
  - As separate quoted arguments

The first command can read from stdin if no file is specified.

Examples:
  # Using braces (recommended - clearest syntax)
  omni pipe '{ls -la}', '{grep .go}', '{wc -l}'
  omni pipe '{cat file.txt}', '{sort}', '{uniq}'
  omni pipe '{cat data.json}', '{jq .users[]}'

  # Using | separator (quote the whole thing)
  omni pipe "cat file.txt | grep pattern | sort | uniq"

  # Using separate arguments with | between
  omni pipe cat file.txt \| grep error \| sort \| uniq -c

  # Using custom separator
  omni pipe --sep "->" "cat file.txt -> grep error -> sort"

  # Multiple quoted commands
  omni pipe "cat file.txt" "grep pattern" "sort" "uniq"

  # With stdin
  echo "hello world" | omni pipe '{grep hello}', '{wc -l}'

  # Verbose mode to see intermediate results
  omni pipe -v '{cat file.txt}', '{head -10}', '{sort}'

  # JSON output with pipeline metadata
  omni pipe --json '{cat file.txt}', '{wc -l}'

Supported commands include all omni commands:
  cat, grep, head, tail, sort, uniq, wc, cut, tr, sed, awk,
  base64, hex, json, jq, yq, curl, and many more.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := pipe.Options{}
		opts.JSON, _ = cmd.Flags().GetBool("json")
		opts.Separator, _ = cmd.Flags().GetString("sep")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")

		registry := pipe.NewRegistry(rootCmd)

		// Check if we have stdin input
		stat, _ := os.Stdin.Stat()
		hasStdin := (stat.Mode() & os.ModeCharDevice) == 0

		if hasStdin && len(args) > 0 {
			// Read stdin and pass as initial input
			scanner := bufio.NewScanner(os.Stdin)
			var lines []string

			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}

			input := strings.Join(lines, "\n")
			if len(lines) > 0 {
				input += "\n"
			}

			return pipe.RunWithInput(cmd.OutOrStdout(), input, args, opts, registry)
		}

		return pipe.Run(cmd.OutOrStdout(), args, opts, registry)
	},
}

func init() {
	rootCmd.AddCommand(pipeCmd)

	pipeCmd.Flags().Bool("json", false, "output result as JSON with metadata")
	pipeCmd.Flags().StringP("sep", "s", "|", "command separator")
	pipeCmd.Flags().BoolP("verbose", "v", false, "show intermediate results")
}
