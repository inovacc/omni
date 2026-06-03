package task

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
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

// Run executes a command via the system shell.
//
// Security (supply-02): values reaching this runner can originate from task
// config (task names, args derived from config) and are NOT necessarily
// user-authored shell text. To prevent shell injection from interpolated
// values, only a SINGLE arg is treated as a literal shell command line (the
// sanctioned "run a shell command" feature). When multiple argv elements are
// supplied, the non-leading elements are delivered to the program as inert
// data, never re-parsed or expanded by the shell, so an element such as
// "foo; rm -rf /", "$(reboot)", or (on Windows) "x>victim" / "a&&calc" cannot
// inject additional commands:
//
//   - POSIX: `sh -c 'exec "$0" "$@"' prog arg...` binds the elements to the
//     positional parameters, so the shell does no word-splitting, globbing, or
//     substitution on them — they reach the program verbatim.
//   - Windows: cmd.exe re-parses `& | > < ^` even inside an already-tokenized
//     argument, because syscall.EscapeArg only quotes a space/tab/quote (a
//     SPACE-FREE metacharacter argument is passed unquoted). The runner instead
//     invokes `cmd /V:ON /C` and references each non-leading element via DELAYED
//     expansion (`!OMNI_ARGn!`) sourced from the environment, so the value
//     expands AFTER tokenization and its metacharacters are inert. See the
//     multi-arg branch below for details and the `!`-literal limitation.
func (r *ShellCommandRunner) Run(ctx context.Context, w io.Writer, args []string) error {
	if len(args) == 0 {
		return nil
	}

	// Create command based on OS.
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		if len(args) == 1 {
			// Single arg: an explicit shell command line.
			cmd = exec.CommandContext(ctx, "cmd", "/C", args[0])
		} else {
			// argv form: pass non-leading args as inert data via the process
			// environment, referenced with DELAYED expansion (`!var!`).
			//
			// supply-02 (Windows): os/exec quotes an argument only when it
			// contains a space/tab/quote (syscall.EscapeArg); it does NOT escape
			// cmd.exe metacharacters (& | > < ^). A SPACE-FREE element such as
			// `x>file` or `a&&calc` would therefore be passed UNQUOTED and
			// cmd.exe would reparse the separators as command syntax (injection).
			//
			// The fix mirrors forloop supply-01: invoke `cmd /V:ON /C` and refer
			// to each non-leading element as `!OMNI_ARGn!`. Delayed expansion
			// substitutes the value AFTER the line is tokenized, so the value
			// lands as a single inert token and its metacharacters are never
			// reparsed as command syntax. The leading element (the program to
			// run) is kept as a literal token so it still resolves as a command.
			//
			// Known limitation: delayed expansion treats `!` as special, so a
			// non-leading value containing a literal `!` may be altered. This is
			// the same accepted tradeoff as forloop: injection safety outweighs
			// faithful handling of a literal `!` in untrusted argv values.
			var b strings.Builder

			b.WriteString(args[0])

			env := os.Environ()

			for i, a := range args[1:] {
				name := fmt.Sprintf("OMNI_ARG%d", i)
				env = append(env, name+"="+a)

				b.WriteString(" !")
				b.WriteString(name)
				b.WriteString("!")
			}

			cmd = exec.CommandContext(ctx, "cmd", "/V:ON", "/C", b.String())
			cmd.Env = env
		}
	} else {
		if len(args) == 1 {
			// Single arg: an explicit shell command line.
			cmd = exec.CommandContext(ctx, "sh", "-c", args[0])
		} else {
			// argv form: run the first element as the command and bind the
			// remaining elements to positional parameters so the shell does
			// no word-splitting, globbing, or substitution on them.
			// `sh -c 'exec "$0" "$@"' prog arg1 arg2 ...`
			shArgs := append([]string{"-c", `exec "$0" "$@"`}, args...)
			cmd = exec.CommandContext(ctx, "sh", shArgs...)
		}
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
