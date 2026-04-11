// Package repo provides repository analysis optimized for LLM consumption.
package repo

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// Options configures the repo analyze command behavior.
type Options struct {
	JSON     bool   // output as JSON
	Compact  bool   // shorter output for smaller context windows
	Sections string // comma-separated section filter
	Output   string // write to file instead of stdout
}

// sectionSet returns the set of requested sections, or nil for all.
func (o Options) sectionSet() map[string]bool {
	if o.Sections == "" {
		return nil
	}

	set := make(map[string]bool)

	for part := range strings.SplitSeq(o.Sections, ",") {
		s := strings.TrimSpace(part)
		if s != "" {
			set[s] = true
		}
	}

	return set
}

// wantSection returns true if the section should be included.
func (o Options) wantSection(name string) bool {
	set := o.sectionSet()
	if set == nil {
		return true
	}

	return set[name]
}

// resolvePath resolves a target path to an absolute directory.
// It handles ".", absolute paths, and relative paths.
func resolvePath(target string) (string, error) {
	if target == "" || target == "." {
		abs, err := filepath.Abs(".")
		if err != nil {
			return "", cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("repo: invalid path: %v", err))
		}
		return abs, nil
	}

	abs, err := filepath.Abs(target)
	if err != nil {
		return "", cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("repo: invalid path %q: %v", target, err))
	}

	info, err := os.Stat(abs)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("repo: path not found: %s", abs))
		}
		if errors.Is(err, os.ErrPermission) {
			return "", cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("repo: permission denied: %s", abs))
		}
		return "", cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("repo: stat %s: %v", abs, err))
	}

	if !info.IsDir() {
		return filepath.Dir(abs), nil
	}

	return abs, nil
}

// isRemote returns true if the target looks like a remote repository reference.
func isRemote(target string) bool {
	if strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "http://") {
		return true
	}

	if strings.HasPrefix(target, "git@") {
		return true
	}

	// owner/repo shorthand (exactly one slash, no dots in first segment except github.com style)
	if strings.Contains(target, "/") && !strings.HasPrefix(target, "/") && !strings.HasPrefix(target, ".") {
		parts := strings.SplitN(target, "/", 3)
		if len(parts) == 2 {
			return true
		}

		// github.com/owner/repo
		if len(parts) == 3 && strings.Contains(parts[0], ".") {
			return true
		}
	}

	return false
}
