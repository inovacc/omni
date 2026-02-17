package repo

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
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

	// Try gh first
	if ghPath, err := exec.LookPath("gh"); err == nil {
		cmd := exec.Command(ghPath, "repo", "clone", cloneURL, tmpDir, "--", "--depth=1")

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

	// Ensure we have a full URL for git clone
	if !strings.HasPrefix(cloneURL, "https://") && !strings.HasPrefix(cloneURL, "git@") {
		cloneURL = "https://github.com/" + cloneURL
	}

	cmd := exec.Command(gitPath, "clone", "--depth=1", cloneURL, tmpDir)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		_ = os.RemoveAll(tmpDir)
		return "", fmt.Errorf("git clone failed: %s", strings.TrimSpace(stderr.String()))
	}

	return tmpDir, nil
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
