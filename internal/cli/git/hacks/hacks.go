// Package hacks provides Git shortcut commands for common operations.
package hacks

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// QuickCommit stages all changes and commits with a message.
// Equivalent to: git add -A && git commit -m "message"
func QuickCommit(message string, addAll bool) error {
	if addAll {
		if err := runGitCommand("add", "-A"); err != nil {
			return fmt.Errorf("failed to stage files: %w", err)
		}
	}

	if err := runGitCommand("commit", "-m", message); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// BranchClean deletes local branches that have been merged into the current branch.
// Returns the list of deleted branches.
func BranchClean(dryRun bool) ([]string, error) {
	// Get current branch (to verify we're in a git repo)
	_, err := getCurrentBranch()
	if err != nil {
		return nil, err
	}

	// Get merged branches
	out, err := runGitCommandOutput("branch", "--merged")
	if err != nil {
		return nil, fmt.Errorf("failed to list merged branches: %w", err)
	}

	var deleted []string

	lines := strings.SplitSeq(strings.TrimSpace(out), "\n")

	for line := range lines {
		branch := strings.TrimSpace(line)
		// Skip current branch (marked with *)
		if strings.HasPrefix(branch, "*") {
			continue
		}
		// Skip main/master/develop branches
		if branch == "main" || branch == "master" || branch == "develop" {
			continue
		}
		// Skip empty
		if branch == "" {
			continue
		}

		if dryRun {
			deleted = append(deleted, branch)
		} else {
			if err := runGitCommand("branch", "-d", branch); err == nil {
				deleted = append(deleted, branch)
			}
		}
	}

	return deleted, nil
}

// Undo undoes the last commit, keeping changes staged.
// Equivalent to: git reset --soft HEAD~1
func Undo() error {
	return runGitCommand("reset", "--soft", "HEAD~1")
}

// AmendNoEdit amends the last commit without editing the message.
// Equivalent to: git commit --amend --no-edit
func AmendNoEdit() error {
	return runGitCommand("commit", "--amend", "--no-edit")
}

// StashStaged stashes only staged changes.
func StashStaged(message string) error {
	args := []string{"stash", "push", "--staged"}
	if message != "" {
		args = append(args, "-m", message)
	}

	return runGitCommand(args...)
}

// LogGraph shows a pretty log with graph.
func LogGraph(count int) (string, error) {
	args := []string{
		"log",
		"--oneline",
		"--graph",
		"--decorate",
		"--all",
	}
	if count > 0 {
		args = append(args, fmt.Sprintf("-n%d", count))
	}

	return runGitCommandOutput(args...)
}

// DiffWords shows word-level diff.
func DiffWords(args ...string) (string, error) {
	cmdArgs := append([]string{"diff", "--word-diff"}, args...)
	return runGitCommandOutput(cmdArgs...)
}

// BlameLine shows blame for a specific line range.
func BlameLine(file string, startLine, endLine int) (string, error) {
	return runGitCommandOutput(
		"blame",
		fmt.Sprintf("-L%d,%d", startLine, endLine),
		file,
	)
}

// Status returns the current git status.
func Status() (string, error) {
	return runGitCommandOutput("status", "-sb")
}

// Push pushes to the remote with current branch.
func Push(force bool) error {
	args := []string{"push"}
	if force {
		args = append(args, "--force-with-lease")
	}

	return runGitCommand(args...)
}

// PullRebase pulls from the remote with rebase.
func PullRebase() error {
	return runGitCommand("pull", "--rebase")
}

// FetchAll fetches all remotes with prune.
func FetchAll() error {
	return runGitCommand("fetch", "--all", "--prune")
}

// Helper functions

func getCurrentBranch() (string, error) {
	out, err := runGitCommandOutput("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return strings.TrimSpace(out), nil
}

func runGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func runGitCommandOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}

	return stdout.String(), nil
}
