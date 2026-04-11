// Package task implements a task runner for omni-only commands.
// It parses Taskfile.yml format and executes only omni internal commands.
package task

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// Options configures the task runner
type Options struct {
	Taskfile      string // Path to Taskfile.yml (default: current directory)
	Dir           string // Working directory
	List          bool   // List available tasks
	Verbose       bool   // Verbose output
	DryRun        bool   // Show commands without executing
	Force         bool   // Force run even if up-to-date
	Silent        bool   // Suppress output
	Summary       bool   // Show task summary/description
	AllowExternal bool   // Allow external (non-omni) commands
}

// DefaultTaskfiles lists the default taskfile names to search for
var DefaultTaskfiles = []string{
	"Taskfile.yml",
	"Taskfile.yaml",
	"taskfile.yml",
	"taskfile.yaml",
}

// CommandRunnerFactory is a function that creates a command runner
// This allows the cmd package to inject the Cobra command runner
// The dir parameter is the working directory for external commands
var CommandRunnerFactory func(dir string, allowExternal bool) CommandRunner

// Run executes the task runner
func Run(ctx context.Context, w io.Writer, taskNames []string, opts Options) error {
	// Find taskfile
	taskfilePath, err := findTaskfile(opts.Taskfile, opts.Dir)
	if err != nil {
		return fmt.Errorf("task: %w", err)
	}

	// Parse taskfile
	tf, err := ParseTaskfile(taskfilePath)
	if err != nil {
		return fmt.Errorf("task: %w", err)
	}

	// Set base directory from taskfile location
	if opts.Dir == "" {
		opts.Dir = filepath.Dir(taskfilePath)
	}

	// Create executor
	exec := NewExecutor(w, tf, opts)

	// Set command runner if factory is available
	if CommandRunnerFactory != nil {
		exec.SetCommandRunner(CommandRunnerFactory(opts.Dir, opts.AllowExternal))
	}

	// List tasks if requested
	if opts.List {
		return exec.ListTasks()
	}

	// Show summary if requested
	if opts.Summary && len(taskNames) > 0 {
		return exec.ShowSummary(taskNames)
	}

	// Run default task if none specified
	if len(taskNames) == 0 {
		if tf.Tasks["default"] != nil {
			taskNames = []string{"default"}
		} else {
			return cmderr.Wrap(cmderr.ErrInvalidInput, "task: no task specified and no default task found")
		}
	}

	// Execute tasks
	for _, name := range taskNames {
		if err := exec.RunTask(ctx, name); err != nil {
			return err
		}
	}

	return nil
}

// findTaskfile searches for a taskfile in the given directory
func findTaskfile(path, dir string) (string, error) {
	// If explicit path given, use it
	if path != "" {
		if _, err := os.Stat(path); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return "", cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("task: taskfile not found: %s", path))
			}
			if errors.Is(err, os.ErrPermission) {
				return "", cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("task: %s", err))
			}
			return "", fmt.Errorf("task: %w", err)
		}

		return path, nil
	}

	// Search in directory
	searchDir := dir
	if searchDir == "" {
		var err error

		searchDir, err = os.Getwd()
		if err != nil {
			return "", cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("task: getwd: %s", err))
		}
	}

	// Look for default taskfiles
	for _, name := range DefaultTaskfiles {
		path := filepath.Join(searchDir, name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("task: no taskfile found in %s", searchDir))
}
