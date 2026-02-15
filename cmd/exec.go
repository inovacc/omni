package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/exec"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec <command> [args...]",
	Short: "Run external commands with credential pre-flight checks",
	Long: `Safe wrapper for external commands that can't be reimplemented in Go.

Before executing, omni inspects the command to detect missing credentials
(registry tokens, cloud provider keys, kubeconfig, etc.) and warns you.

Examples:
  omni exec pnpm install
  omni exec --force docker build .
  omni exec --dry-run aws s3 ls
  omni exec --strict kubectl get pods`,
	DisableFlagParsing: false,
	Args:               cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		strict, _ := cmd.Flags().GetBool("strict")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		noPrompt, _ := cmd.Flags().GetBool("no-prompt")

		return exec.Run(os.Stdout, args[0], args[1:], exec.Options{
			Force:    force,
			Strict:   strict,
			DryRun:   dryRun,
			NoPrompt: noPrompt,
		})
	},
}

func init() {
	execCmd.Flags().BoolP("force", "f", false, "Skip credential checks, execute immediately")
	execCmd.Flags().Bool("strict", false, "Abort if any credentials are missing (CI mode)")
	execCmd.Flags().Bool("dry-run", false, "Show credential checks without executing")
	execCmd.Flags().Bool("no-prompt", false, "Don't prompt, just warn and proceed")

	rootCmd.AddCommand(execCmd)
}
