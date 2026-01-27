package cmd

import (
	"os"

	"github.com/inovacc/goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// lintCmd represents the lint command
var lintCmd = &cobra.Command{
	Use:   "lint [OPTION]... [FILE|DIR]...",
	Short: "Check Taskfiles for portability issues",
	Long: `Lint Taskfiles for cross-platform portability.

Checks for:
  - Shell commands that should use goshell equivalents
  - Non-portable commands (package managers, OS-specific tools)
  - Bash-specific syntax ([[ ]], <<<, etc.)
  - Hardcoded Unix paths
  - Pipe chains that may fail silently

Severity levels:
  error   - Will likely fail on some platforms
  warning - May cause issues, should be reviewed
  info    - Suggestions for improvement

Examples:
  goshell lint                            # lint Taskfile.yml in current dir
  goshell lint Taskfile.yml               # lint specific file
  goshell lint ./tasks/                   # lint all Taskfiles in directory
  goshell lint --strict Taskfile.yml      # enable strict mode
  goshell lint -q Taskfile.yml            # only show errors`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.LintOptions{}

		opts.Format, _ = cmd.Flags().GetString("format")
		opts.Fix, _ = cmd.Flags().GetBool("fix")
		opts.Strict, _ = cmd.Flags().GetBool("strict")
		opts.Quiet, _ = cmd.Flags().GetBool("quiet")

		return cli.RunLint(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(lintCmd)

	lintCmd.Flags().StringP("format", "f", "text", "output format (text, json)")
	lintCmd.Flags().Bool("fix", false, "auto-fix issues where possible")
	lintCmd.Flags().Bool("strict", false, "enable strict mode (more warnings become errors)")
	lintCmd.Flags().BoolP("quiet", "q", false, "only show errors, not warnings")
}
