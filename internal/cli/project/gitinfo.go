package project

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// analyzeGit gathers git repository information.
func analyzeGit(dir string, limit int) *GitReport {
	report := &GitReport{}

	// Check if it's a git repo
	gitDir := filepath.Join(dir, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return report
	}

	report.IsRepo = true

	if limit <= 0 {
		limit = 10
	}

	// Current branch
	if out, err := runGitInDir(dir, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		report.Branch = strings.TrimSpace(out)
	}

	// Remote
	if out, err := runGitInDir(dir, "remote"); err == nil {
		remote := strings.TrimSpace(out)
		if remote != "" {
			// Take the first remote
			parts := strings.SplitN(remote, "\n", 2)
			report.Remote = parts[0]

			// Get remote URL
			if url, err := runGitInDir(dir, "remote", "get-url", report.Remote); err == nil {
				report.RemoteURL = sanitizeURL(strings.TrimSpace(url))
			}
		}
	}

	// Clean status
	if out, err := runGitInDir(dir, "status", "--porcelain"); err == nil {
		report.Clean = strings.TrimSpace(out) == ""
	}

	// Ahead/behind
	if report.Branch != "" && report.Remote != "" {
		tracking := report.Remote + "/" + report.Branch
		if out, err := runGitInDir(dir, "rev-list", "--left-right", "--count", report.Branch+"..."+tracking); err == nil {
			parts := strings.Fields(strings.TrimSpace(out))
			if len(parts) == 2 {
				report.Ahead, _ = strconv.Atoi(parts[0])
				report.Behind, _ = strconv.Atoi(parts[1])
			}
		}
	}

	// Recent commits
	format := "--pretty=format:%h %s"
	if out, err := runGitInDir(dir, "log", format, fmt.Sprintf("-n%d", limit)); err == nil {
		lines := strings.TrimSpace(out)
		if lines != "" {
			report.RecentCommits = strings.Split(lines, "\n")
		}
	}

	// Tags (most recent first)
	if out, err := runGitInDir(dir, "tag", "--sort=-version:refname"); err == nil {
		tags := strings.TrimSpace(out)
		if tags != "" {
			all := strings.Split(tags, "\n")
			if len(all) > limit {
				all = all[:limit]
			}

			report.Tags = all
		}
	}

	// Total commits
	if out, err := runGitInDir(dir, "rev-list", "--count", "HEAD"); err == nil {
		report.TotalCommits, _ = strconv.Atoi(strings.TrimSpace(out))
	}

	// Contributors
	if out, err := runGitInDir(dir, "shortlog", "-sn", "--all"); err == nil {
		lines := strings.TrimSpace(out)
		if lines != "" {
			report.Contributors = len(strings.Split(lines, "\n"))
		}
	}

	return report
}

// sanitizeURL strips credentials from URLs (e.g., https://user:token@host/...).
func sanitizeURL(rawURL string) string {
	if strings.Contains(rawURL, "@") && strings.Contains(rawURL, "://") {
		// https://user:pass@host/path -> https://host/path
		parts := strings.SplitN(rawURL, "://", 2)
		if len(parts) == 2 {
			if idx := strings.Index(parts[1], "@"); idx >= 0 {
				return parts[0] + "://" + parts[1][idx+1:]
			}
		}
	}

	return rawURL
}

func runGitInDir(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s: %w", stderr.String(), err)
	}

	return stdout.String(), nil
}

// RunGit runs the git info subcommand.
func RunGit(w io.Writer, args []string, opts Options) error {
	dir, err := resolvePath(args)
	if err != nil {
		return fmt.Errorf("project git: %w", err)
	}

	report := analyzeGit(dir, opts.Limit)

	if !report.IsRepo {
		_, _ = fmt.Fprintln(w, "Not a git repository.")
		return nil
	}

	if opts.JSON {
		return formatGitJSON(w, report)
	}

	if opts.Markdown {
		return formatGitMarkdown(w, report)
	}

	return formatGitText(w, report)
}
