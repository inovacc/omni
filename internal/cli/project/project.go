// Package project provides a general-purpose project analyzer command.
package project

import (
	"os"
	"path/filepath"
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
		return "", err
	}

	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}

	if !info.IsDir() {
		return filepath.Dir(abs), nil
	}

	return abs, nil
}
