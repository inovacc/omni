// Package hacks provides GitHub CLI shortcut commands for common operations.
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

// TODO:no-exec-violation: this package uses os/exec.Command("gh", ...) which violates
// CLAUDE.md "No exec" design principle. Tracked for Plan 17 / backlog cleanup.

// PRCheckout checks out a pull request by number.
func PRCheckout(number int) error {
	return runGhCommand("pr", "checkout", fmt.Sprintf("%d", number))
}

// PRDiff shows the diff for a pull request.
func PRDiff(number int) (string, error) {
	return runGhCommandOutput("pr", "diff", fmt.Sprintf("%d", number))
}

// PRApprove approves a pull request.
func PRApprove(number int) error {
	return runGhCommand("pr", "review", fmt.Sprintf("%d", number), "--approve")
}

// IssueMine lists issues assigned to the current user.
func IssueMine() (string, error) {
	return runGhCommandOutput("issue", "list", "--assignee", "@me")
}

// RepoCloneOrg clones all repositories in an organization.
func RepoCloneOrg(org string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 100
	}

	out, err := runGhCommandOutput("repo", "list", org, "--limit", fmt.Sprintf("%d", limit), "--json", "name", "--jq", ".[].name")
	if err != nil {
		return nil, cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("gh: failed to list repos for org %q: %v", org, err))
	}

	var cloned []string

	for name := range strings.SplitSeq(strings.TrimSpace(out), "\n") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		repo := fmt.Sprintf("%s/%s", org, name)
		if err := runGhCommand("repo", "clone", repo); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "gh: failed to clone %s: %v\n", repo, err)

			continue
		}

		cloned = append(cloned, repo)
	}

	return cloned, nil
}

// ActionsRerun re-runs a workflow run by ID.
func ActionsRerun(runID int) error {
	return runGhCommand("run", "rerun", fmt.Sprintf("%d", runID))
}

func runGhCommand(args ...string) error {
	cmd := exec.Command("gh", args...) //nolint:gosec // TODO:no-exec-violation
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			switch exitErr.ExitCode() {
			case 4: // gh uses exit 4 for auth errors
				return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("gh: not authenticated (exit %d)", exitErr.ExitCode()))
			default:
				return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("gh: command failed with exit code %d", exitErr.ExitCode()))
			}
		}
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("gh: %v", err))
	}

	return nil
}

func runGhCommandOutput(args ...string) (string, error) {
	cmd := exec.Command("gh", args...) //nolint:gosec // TODO:no-exec-violation

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			switch exitErr.ExitCode() {
			case 4: // gh auth error
				return "", cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("gh: not authenticated (exit %d): %s", exitErr.ExitCode(), stderr.String()))
			default:
				return "", cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("gh: %s (exit %d)", stderr.String(), exitErr.ExitCode()))
			}
		}
		return "", cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("gh: %v: %s", err, stderr.String()))
	}

	return stdout.String(), nil
}
