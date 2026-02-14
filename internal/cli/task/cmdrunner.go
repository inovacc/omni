package task

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// CobraCommandRunner runs commands using a Cobra root command
type CobraCommandRunner struct {
	rootCmd *cobra.Command
}

// NewCobraCommandRunner creates a command runner from a Cobra root command
func NewCobraCommandRunner(rootCmd *cobra.Command) *CobraCommandRunner {
	return &CobraCommandRunner{rootCmd: rootCmd}
}

// Run executes a command using Cobra
func (r *CobraCommandRunner) Run(ctx context.Context, w io.Writer, args []string) error {
	if len(args) == 0 {
		return nil
	}

	// Create a buffer to capture output
	var stdout, stderr bytes.Buffer

	// Clone the root command to avoid state pollution
	cmd := findSubCommand(r.rootCmd, args[0])
	if cmd == nil {
		return fmt.Errorf("unknown command: %s", args[0])
	}

	// Set output
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	// Execute with remaining args
	cmd.SetArgs(args[1:])
	err := cmd.Execute()

	// Write output to writer
	_, _ = w.Write(stdout.Bytes())
	if stderr.Len() > 0 {
		_, _ = w.Write(stderr.Bytes())
	}

	return err
}

// findSubCommand finds a subcommand by name
func findSubCommand(root *cobra.Command, name string) *cobra.Command {
	for _, cmd := range root.Commands() {
		if cmd.Name() == name || cmd.HasAlias(name) {
			return cmd
		}
	}

	return nil
}

// MockCommandRunner is for testing
type MockCommandRunner struct {
	Commands [][]string
	Outputs  map[string]string
	Errors   map[string]error
}

// NewMockCommandRunner creates a mock command runner
func NewMockCommandRunner() *MockCommandRunner {
	return &MockCommandRunner{
		Commands: make([][]string, 0),
		Outputs:  make(map[string]string),
		Errors:   make(map[string]error),
	}
}

// Run records the command and returns mocked output/error
func (m *MockCommandRunner) Run(ctx context.Context, w io.Writer, args []string) error {
	m.Commands = append(m.Commands, args)

	if len(args) == 0 {
		return nil
	}

	key := args[0]
	if output, ok := m.Outputs[key]; ok {
		_, _ = fmt.Fprint(w, output)
	}

	if err, ok := m.Errors[key]; ok {
		return err
	}

	return nil
}

// SetOutput sets the output for a command
func (m *MockCommandRunner) SetOutput(cmd, output string) {
	m.Outputs[cmd] = output
}

// SetError sets an error for a command
func (m *MockCommandRunner) SetError(cmd string, err error) {
	m.Errors[cmd] = err
}

// ShellCommandRunner runs commands via the system shell
type ShellCommandRunner struct {
	dir string // Working directory
}

// NewShellCommandRunner creates a shell command runner
func NewShellCommandRunner(dir string) *ShellCommandRunner {
	return &ShellCommandRunner{dir: dir}
}

// Run executes a command via the system shell
func (r *ShellCommandRunner) Run(ctx context.Context, w io.Writer, args []string) error {
	if len(args) == 0 {
		return nil
	}

	// Join args into a command string
	cmdStr := strings.Join(args, " ")

	// Create command based on OS
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", cmdStr)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", cmdStr)
	}

	// Set working directory
	if r.dir != "" {
		cmd.Dir = r.dir
	}

	// Capture output
	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	err := cmd.Run()

	// Write output
	_, _ = w.Write(stdout.Bytes())
	if stderr.Len() > 0 {
		_, _ = w.Write(stderr.Bytes())
	}

	return err
}

// HybridCommandRunner routes commands to either omni (Cobra) or shell
type HybridCommandRunner struct {
	omniRunner  CommandRunner
	shellRunner *ShellCommandRunner
}

// NewHybridCommandRunner creates a hybrid command runner
func NewHybridCommandRunner(omniRunner CommandRunner, dir string) *HybridCommandRunner {
	return &HybridCommandRunner{
		omniRunner:  omniRunner,
		shellRunner: NewShellCommandRunner(dir),
	}
}

// Run executes a command, routing to omni or shell as appropriate
func (r *HybridCommandRunner) Run(ctx context.Context, w io.Writer, args []string) error {
	if len(args) == 0 {
		return nil
	}

	// Try omni first
	err := r.omniRunner.Run(ctx, w, args)
	if err == nil {
		return nil
	}

	// If omni command not found, try shell
	if strings.Contains(err.Error(), "unknown command") {
		return r.shellRunner.Run(ctx, w, args)
	}

	return err
}
