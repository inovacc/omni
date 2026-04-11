// Package project provides a general-purpose project analyzer command.
package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// Options configures the project command behavior.
type Options struct {
	JSON     bool // output as JSON
	Markdown bool // output as Markdown
	Verbose  bool // verbose output
	Limit    int  // limit for lists (e.g., recent commits)
}

// resolvePath returns the absolute directory path from args or current dir.
func resolvePath(args []string) (string, error) {
	dir := "."
	if len(args) > 0 && args[0] != "" {
		dir = args[0]
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("project: invalid path %q: %v", dir, err))
	}

	info, err := os.Stat(abs)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("project: path not found: %s", abs))
		}
		if errors.Is(err, os.ErrPermission) {
			return "", cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("project: permission denied: %s", abs))
		}
		return "", cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("project: stat %s: %v", abs, err))
	}

	if !info.IsDir() {
		return filepath.Dir(abs), nil
	}

	return abs, nil
}
