// Package hacks provides Git shortcut commands for common operations.
package hacks

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// Sanctioned exec exception: this package's purpose is to orchestrate the
// external git tool. Permitted under the no-exec invariant — see
// docs/architecture/patterns.md § "No-exec invariant: scope & sanctioned exceptions".

// QuickCommit stages all changes and commits with a message.
// Equivalent to: git add -A && git commit -m "message"
func QuickCommit(message string, addAll bool) error {
	if addAll {
		if err := runGitCommand("add", "-A"); err != nil {
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("git: failed to stage files: %v", err))
		}
	}

	if err := runGitCommand("commit", "-m", message); err != nil {
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("git: failed to commit: %v", err))
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
		return nil, cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("git: failed to list merged branches: %v", err))
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
		return "", cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("git: not a git repository or failed to get branch: %v", err))
	}

	return strings.TrimSpace(out), nil
}

func runGitCommand(args ...string) error {
	cmd := exec.Command("git", args...) //nolint:gosec // sanctioned exec exception (see package doc)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("git: command failed with exit code %d", exitErr.ExitCode()))
		}
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("git: %v", err))
	}

	return nil
}

func runGitCommandOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...) //nolint:gosec // sanctioned exec exception (see package doc)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("git: %s (exit %d)", stderr.String(), exitErr.ExitCode()))
		}
		return "", cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("git: %v: %s", err, stderr.String()))
	}

	return stdout.String(), nil
}
