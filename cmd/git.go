package cmd

import (
	"fmt"

	"github.com/inovacc/omni/internal/cli/git/hacks"
	"github.com/spf13/cobra"
)

var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "Git shortcuts and hacks",
	Long:  `Git shortcut commands for common operations.`,
}

// git quick-commit (gqc)
var gitQuickCommitCmd = &cobra.Command{
	Use:     "quick-commit",
	Aliases: []string{"qc"},
	Short:   "Stage all and commit",
	Long: `Stage all changes and commit with a message.
Equivalent to: git add -A && git commit -m "message"

Examples:
  omni git quick-commit -m "fix bug"
  omni git qc -m "add feature"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		msg, _ := cmd.Flags().GetString("message")
		if msg == "" {
			return fmt.Errorf("commit message is required (-m)")
		}
		addAll, _ := cmd.Flags().GetBool("all")
		return hacks.QuickCommit(msg, addAll)
	},
}

// git branch-clean (gbc)
var gitBranchCleanCmd = &cobra.Command{
	Use:     "branch-clean",
	Aliases: []string{"bc"},
	Short:   "Delete merged branches",
	Long: `Delete local branches that have been merged into the current branch.
Skips main, master, and develop branches.

Examples:
  omni git branch-clean
  omni git bc --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		deleted, err := hacks.BranchClean(dryRun)
		if err != nil {
			return err
		}
		if len(deleted) == 0 {
			fmt.Println("No merged branches to delete")
			return nil
		}
		action := "Deleted"
		if dryRun {
			action = "Would delete"
		}
		for _, branch := range deleted {
			fmt.Printf("%s: %s\n", action, branch)
		}
		return nil
	},
}

// git undo
var gitUndoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Undo last commit (soft reset)",
	Long: `Undo the last commit, keeping changes staged.
Equivalent to: git reset --soft HEAD~1

Examples:
  omni git undo`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return hacks.Undo()
	},
}

// git amend
var gitAmendCmd = &cobra.Command{
	Use:   "amend",
	Short: "Amend last commit without editing",
	Long: `Amend the last commit without editing the message.
Equivalent to: git commit --amend --no-edit

Examples:
  omni git amend`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return hacks.AmendNoEdit()
	},
}

// git stash-staged
var gitStashStagedCmd = &cobra.Command{
	Use:   "stash-staged",
	Short: "Stash only staged changes",
	Long: `Stash only staged changes, leaving unstaged changes in the working directory.

Examples:
  omni git stash-staged
  omni git stash-staged -m "WIP: feature"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		msg, _ := cmd.Flags().GetString("message")
		return hacks.StashStaged(msg)
	},
}

// git log-graph
var gitLogGraphCmd = &cobra.Command{
	Use:     "log-graph",
	Aliases: []string{"lg"},
	Short:   "Pretty log with graph",
	Long: `Show a pretty git log with graph visualization.
Equivalent to: git log --oneline --graph --decorate --all

Examples:
  omni git log-graph
  omni git lg -n 20`,
	RunE: func(cmd *cobra.Command, args []string) error {
		count, _ := cmd.Flags().GetInt("count")
		out, err := hacks.LogGraph(count)
		if err != nil {
			return err
		}
		fmt.Print(out)
		return nil
	},
}

// git diff-words
var gitDiffWordsCmd = &cobra.Command{
	Use:   "diff-words",
	Short: "Word-level diff",
	Long: `Show word-level diff instead of line-level.
Equivalent to: git diff --word-diff

Examples:
  omni git diff-words
  omni git diff-words HEAD~1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		out, err := hacks.DiffWords(args...)
		if err != nil {
			return err
		}
		fmt.Print(out)
		return nil
	},
}

// git blame-line
var gitBlameLineCmd = &cobra.Command{
	Use:   "blame-line <file>",
	Short: "Blame specific line range",
	Long: `Show blame for a specific line range in a file.

Examples:
  omni git blame-line main.go --start 10 --end 20`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		start, _ := cmd.Flags().GetInt("start")
		end, _ := cmd.Flags().GetInt("end")
		if start == 0 || end == 0 {
			return fmt.Errorf("--start and --end flags are required")
		}
		out, err := hacks.BlameLine(args[0], start, end)
		if err != nil {
			return err
		}
		fmt.Print(out)
		return nil
	},
}

// git status (short)
var gitStatusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"st"},
	Short:   "Short status",
	Long: `Show short git status.
Equivalent to: git status -sb

Examples:
  omni git status
  omni git st`,
	RunE: func(cmd *cobra.Command, args []string) error {
		out, err := hacks.Status()
		if err != nil {
			return err
		}
		fmt.Print(out)
		return nil
	},
}

// git push
var gitPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push to remote",
	Long: `Push to the remote repository.

Examples:
  omni git push
  omni git push --force`,
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		return hacks.Push(force)
	},
}

// git pull-rebase
var gitPullRebaseCmd = &cobra.Command{
	Use:     "pull-rebase",
	Aliases: []string{"pr"},
	Short:   "Pull with rebase",
	Long: `Pull from remote with rebase.
Equivalent to: git pull --rebase

Examples:
  omni git pull-rebase
  omni git pr`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return hacks.PullRebase()
	},
}

// git fetch-all
var gitFetchAllCmd = &cobra.Command{
	Use:     "fetch-all",
	Aliases: []string{"fa"},
	Short:   "Fetch all remotes with prune",
	Long: `Fetch all remotes with prune.
Equivalent to: git fetch --all --prune

Examples:
  omni git fetch-all
  omni git fa`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return hacks.FetchAll()
	},
}

func init() {
	rootCmd.AddCommand(gitCmd)

	// quick-commit
	gitQuickCommitCmd.Flags().StringP("message", "m", "", "Commit message (required)")
	gitQuickCommitCmd.Flags().BoolP("all", "a", true, "Stage all changes before commit")
	gitCmd.AddCommand(gitQuickCommitCmd)

	// branch-clean
	gitBranchCleanCmd.Flags().Bool("dry-run", false, "Show branches that would be deleted")
	gitCmd.AddCommand(gitBranchCleanCmd)

	// undo
	gitCmd.AddCommand(gitUndoCmd)

	// amend
	gitCmd.AddCommand(gitAmendCmd)

	// stash-staged
	gitStashStagedCmd.Flags().StringP("message", "m", "", "Stash message")
	gitCmd.AddCommand(gitStashStagedCmd)

	// log-graph
	gitLogGraphCmd.Flags().IntP("count", "n", 0, "Number of commits to show")
	gitCmd.AddCommand(gitLogGraphCmd)

	// diff-words
	gitCmd.AddCommand(gitDiffWordsCmd)

	// blame-line
	gitBlameLineCmd.Flags().Int("start", 0, "Start line number")
	gitBlameLineCmd.Flags().Int("end", 0, "End line number")
	gitCmd.AddCommand(gitBlameLineCmd)

	// status
	gitCmd.AddCommand(gitStatusCmd)

	// push
	gitPushCmd.Flags().Bool("force", false, "Force push (with lease)")
	gitCmd.AddCommand(gitPushCmd)

	// pull-rebase
	gitCmd.AddCommand(gitPullRebaseCmd)

	// fetch-all
	gitCmd.AddCommand(gitFetchAllCmd)
}

// Standalone aliases for quick access
var gqcCmd = &cobra.Command{
	Use:   "gqc",
	Short: "Git quick commit (alias)",
	Long: `Alias for 'git quick-commit'.
Stage all changes and commit with a message.

Examples:
  omni gqc -m "fix bug"`,
	RunE: gitQuickCommitCmd.RunE,
}

var gbcCmd = &cobra.Command{
	Use:   "gbc",
	Short: "Git branch clean (alias)",
	Long: `Alias for 'git branch-clean'.
Delete merged branches.

Examples:
  omni gbc --dry-run`,
	RunE: gitBranchCleanCmd.RunE,
}

func init() {
	// Register standalone aliases
	gqcCmd.Flags().StringP("message", "m", "", "Commit message (required)")
	gqcCmd.Flags().BoolP("all", "a", true, "Stage all changes before commit")
	rootCmd.AddCommand(gqcCmd)

	gbcCmd.Flags().Bool("dry-run", false, "Show branches that would be deleted")
	rootCmd.AddCommand(gbcCmd)
}
