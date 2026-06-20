package pipe

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/spf13/pflag"
)

// executeCommand executes a single omni command.
// It tries the unified command.Registry first (if available), then falls back to Cobra dispatch.
func executeCommand(registry *CommandRegistry, cmdParts []string, stdin io.Reader, stdout io.Writer) error {
	if registry == nil {
		return fmt.Errorf("command registry not initialized")
	}

	cmdName := cmdParts[0]
	cmdArgs := cmdParts[1:]

	// Try unified Registry first
	if registry.Unified != nil {
		if cmd, ok := registry.Unified.Get(cmdName); ok {
			r := stdin
			if r == nil {
				r = strings.NewReader("")
			}
			return cmd.Run(context.Background(), stdout, r, cmdArgs)
		}
	}

	// Fall back to Cobra dispatch
	return executeCobraCommand(registry, cmdParts, stdin, stdout)
}

// executeCobraCommand dispatches via Cobra's command tree.
func executeCobraCommand(registry *CommandRegistry, cmdParts []string, stdin io.Reader, stdout io.Writer) error {
	if registry.RootCmd == nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("unknown command: %s", cmdParts[0]))
	}

	// Find the command
	cmd, args, err := registry.RootCmd.Find(cmdParts)
	if err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("unknown command: %s", cmdParts[0]))
	}

	// Check if we found a valid command (not root)
	if cmd == registry.RootCmd {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("unknown command: %s", cmdParts[0]))
	}

	// Set up input/output using Cobra's built-in methods
	if stdin != nil {
		cmd.SetIn(stdin)
	}

	cmd.SetOut(stdout)
	cmd.SetErr(stdout)

	// Reset flags to defaults
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
	})

	// Parse flags
	if err := cmd.ParseFlags(args); err != nil {
		return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("parsing flags: %s", err))
	}

	// Get remaining args after flags
	cmdArgs := cmd.Flags().Args()

	// Execute directly via RunE
	if cmd.RunE != nil {
		return cmd.RunE(cmd, cmdArgs)
	}

	if cmd.Run != nil {
		cmd.Run(cmd, cmdArgs)
		return nil
	}

	return fmt.Errorf("command %s has no run function", cmdParts[0])
}
