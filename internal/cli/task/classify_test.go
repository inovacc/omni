package task

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/spf13/cobra"
)

// TestCobraCommandRunner_UnknownCommand_IsInvalidInput verifies that invoking an
// unknown command through the Cobra command runner is a usage/argument error
// (exit 2), classified as cmderr.ErrInvalidInput.
func TestCobraCommandRunner_UnknownCommand_IsInvalidInput(t *testing.T) {
	root := &cobra.Command{Use: "omni"}
	runner := NewCobraCommandRunner(root)

	var buf bytes.Buffer
	err := runner.Run(context.Background(), &buf, []string{"definitely-not-a-command"})
	if !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Fatalf("unknown command: want ErrInvalidInput, got %v", err)
	}
}

// TestExecutor_UnknownExternalCommand_IsInvalidInput verifies that an
// unmapped, non-external command in a task step is a usage error (exit 2).
func TestExecutor_UnknownExternalCommand_IsInvalidInput(t *testing.T) {
	tf := &Taskfile{
		Version: "3",
		Tasks: map[string]*Task{
			"build": {Cmds: []Command{{Cmd: "this-is-not-an-omni-command --x"}}},
		},
	}

	var buf bytes.Buffer
	e := NewExecutor(&buf, tf, Options{AllowExternal: false})
	err := e.RunTask(context.Background(), "build")
	if !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Fatalf("unknown external command: want ErrInvalidInput, got %v", err)
	}
}
