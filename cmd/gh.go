package cmd

import (
	"fmt"

	ghhacks "github.com/inovacc/omni/internal/cli/gh/hacks"
	"github.com/spf13/cobra"
)

var ghCmd = &cobra.Command{
	Use:   "gh",
	Short: "GitHub CLI shortcuts",
	Long:  `Convenience wrappers around common gh (GitHub CLI) operations.`,
}

var ghPRCheckoutCmd = &cobra.Command{
	Use:   "pr-checkout <number>",
	Short: "Check out a pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := parseIntArg(args[0], "PR number")
		if err != nil {
			return err
		}

		return ghhacks.PRCheckout(n)
	},
}

var ghPRDiffCmd = &cobra.Command{
	Use:   "pr-diff <number>",
	Short: "Show diff for a pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := parseIntArg(args[0], "PR number")
		if err != nil {
			return err
		}

		out, err := ghhacks.PRDiff(n)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprint(cmd.OutOrStdout(), out)

		return nil
	},
}

var ghPRApproveCmd = &cobra.Command{
	Use:   "pr-approve <number>",
	Short: "Approve a pull request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := parseIntArg(args[0], "PR number")
		if err != nil {
			return err
		}

		return ghhacks.PRApprove(n)
	},
}

var ghIssueMineCmd = &cobra.Command{
	Use:   "issue-mine",
	Short: "List issues assigned to you",
	RunE: func(cmd *cobra.Command, args []string) error {
		out, err := ghhacks.IssueMine()
		if err != nil {
			return err
		}

		_, _ = fmt.Fprint(cmd.OutOrStdout(), out)

		return nil
	},
}

var ghRepoCloneOrgCmd = &cobra.Command{
	Use:   "repo-clone-org <org>",
	Short: "Clone all repositories in an organization",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")

		cloned, err := ghhacks.RepoCloneOrg(args[0], limit)
		if err != nil {
			return err
		}

		for _, repo := range cloned {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "cloned: %s\n", repo)
		}

		return nil
	},
}

var ghActionsRerunCmd = &cobra.Command{
	Use:   "actions-rerun <run-id>",
	Short: "Re-run a workflow run",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := parseIntArg(args[0], "run ID")
		if err != nil {
			return err
		}

		return ghhacks.ActionsRerun(n)
	},
}

func parseIntArg(s, label string) (int, error) {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return 0, fmt.Errorf("invalid %s: %q", label, s)
	}

	return n, nil
}

func init() {
	rootCmd.AddCommand(ghCmd)

	ghCmd.AddCommand(ghPRCheckoutCmd)
	ghCmd.AddCommand(ghPRDiffCmd)
	ghCmd.AddCommand(ghPRApproveCmd)
	ghCmd.AddCommand(ghIssueMineCmd)
	ghCmd.AddCommand(ghRepoCloneOrgCmd)
	ghCmd.AddCommand(ghActionsRerunCmd)

	ghRepoCloneOrgCmd.Flags().Int("limit", 100, "max repos to clone")
}
