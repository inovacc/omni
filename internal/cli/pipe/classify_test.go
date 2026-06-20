package pipe

import (
	"bytes"
	"errors"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/spf13/cobra"
)

// TestExecuteCommand_UnknownCommand_IsInvalidInput verifies that dispatching an
// unknown command name through the Cobra fallback path is a usage/argument
// error (exit 2), classified as cmderr.ErrInvalidInput.
func TestExecuteCommand_UnknownCommand_IsInvalidInput(t *testing.T) {
	root := &cobra.Command{Use: "omni"}
	root.AddCommand(&cobra.Command{
		Use: "real",
		RunE: func(_ *cobra.Command, _ []string) error {
			return nil
		},
	})
	registry := NewRegistry(root)

	var stdout bytes.Buffer
	err := executeCommand(registry, []string{"definitely-not-a-command"}, nil, &stdout)
	if !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Fatalf("unknown command: want ErrInvalidInput, got %v", err)
	}
}

// TestExecuteCobraCommand_BadFlag_IsInvalidInput verifies that a flag-parse
// failure on a known command is a usage error (exit 2).
func TestExecuteCobraCommand_BadFlag_IsInvalidInput(t *testing.T) {
	root := &cobra.Command{Use: "omni"}
	root.AddCommand(&cobra.Command{
		Use: "real",
		RunE: func(_ *cobra.Command, _ []string) error {
			return nil
		},
	})
	registry := NewRegistry(root)

	var stdout bytes.Buffer
	err := executeCobraCommand(registry, []string{"real", "--no-such-flag"}, nil, &stdout)
	if !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Fatalf("bad flag: want ErrInvalidInput, got %v", err)
	}
}
