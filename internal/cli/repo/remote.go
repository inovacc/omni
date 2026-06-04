package repo

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// cloneToTemp clones a remote repository to a temporary directory.
// It tries gh first, then falls back to git clone.
// Returns the temp directory path (caller must clean up) and any error.
func cloneToTemp(target string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "omni-repo-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}

	// Normalize target to a clone-friendly URL
	cloneURL := normalizeRemote(target)

	// Reject a normalized target that looks like a flag before any process is
	// invoked: such a value would reach git/gh as an option (CWE-88 argument
	// injection). Both argv builders below also enforce this, but guarding here
	// keeps the rejection independent of which tool (gh/git) is on PATH.
	if strings.HasPrefix(cloneURL, "-") {
		_ = os.RemoveAll(tmpDir)
		return "", cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("repo: refusing clone target that looks like a flag: %q", cloneURL))
	}

	// Try gh first
	if ghPath, err := exec.LookPath("gh"); err == nil {
		ghArgs, argErr := ghCloneArgs(cloneURL, tmpDir)
		if argErr != nil {
			_ = os.RemoveAll(tmpDir)
			return "", argErr
		}

		cmd := exec.Command(ghPath, ghArgs...)

		var stderr bytes.Buffer

		cmd.Stderr = &stderr

		if err := cmd.Run(); err == nil {
			return tmpDir, nil
		}
	}

	// Fall back to git clone
	gitPath, err := exec.LookPath("git")
	if err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", fmt.Errorf("neither gh nor git found in PATH")
	}

	gitArgs, argErr := gitCloneArgs(cloneURL, tmpDir)
	if argErr != nil {
		_ = os.RemoveAll(tmpDir)
		return "", argErr
	}

	cmd := exec.Command(gitPath, gitArgs...)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", fmt.Errorf("git clone failed: %s", strings.TrimSpace(stderr.String()))
	}

	return tmpDir, nil
}

// ghCloneArgs builds the argv for `gh repo clone`, rejecting a clone URL that
// begins with "-" (which gh would otherwise treat as a flag — CWE-88 argument
// injection) and inserting a "--" option terminator immediately before the URL.
func ghCloneArgs(cloneURL, dest string) ([]string, error) {
	if strings.HasPrefix(cloneURL, "-") {
		return nil, cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("repo: refusing clone target that looks like a flag: %q", cloneURL))
	}

	return []string{"repo", "clone", "--", cloneURL, dest, "--", "--depth=1"}, nil
}

// gitCloneArgs builds the argv for `git clone`, rejecting a clone URL that
// begins with "-" (which git would otherwise treat as a flag — CWE-88 argument
// injection) and inserting a "--" option terminator immediately before the URL.
func gitCloneArgs(cloneURL, dest string) ([]string, error) {
	// Ensure we have a full URL for git clone.
	if !strings.HasPrefix(cloneURL, "https://") && !strings.HasPrefix(cloneURL, "git@") {
		if strings.HasPrefix(cloneURL, "-") {
			return nil, cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("repo: refusing clone target that looks like a flag: %q", cloneURL))
		}

		cloneURL = "https://github.com/" + cloneURL
	}

	if strings.HasPrefix(cloneURL, "-") {
		return nil, cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("repo: refusing clone target that looks like a flag: %q", cloneURL))
	}

	return []string{"clone", "--depth=1", "--", cloneURL, dest}, nil
}

// normalizeRemote converts various remote formats to a gh-friendly reference.
func normalizeRemote(target string) string {
	// Already a full URL
	if strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "git@") {
		return target
	}

	// github.com/owner/repo -> owner/repo
	target = strings.TrimPrefix(target, "github.com/")

	// Remove trailing .git
	target = strings.TrimSuffix(target, ".git")

	return target
}
