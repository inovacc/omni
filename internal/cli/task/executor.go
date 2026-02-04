package task

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Executor handles task execution
type Executor struct {
	w          io.Writer
	tf         *Taskfile
	opts       Options
	resolver   *DependencyResolver
	executed   map[string]bool
	cmdRunner  CommandRunner
}

// CommandRunner is the interface for running omni commands
type CommandRunner interface {
	Run(ctx context.Context, w io.Writer, args []string) error
}

// DefaultCommandRunner runs omni commands via the command registry
type DefaultCommandRunner struct{}

// NewExecutor creates a new task executor
func NewExecutor(w io.Writer, tf *Taskfile, opts Options) *Executor {
	return &Executor{
		w:         w,
		tf:        tf,
		opts:      opts,
		resolver:  NewDependencyResolver(tf),
		executed:  make(map[string]bool),
		cmdRunner: &DefaultCommandRunner{},
	}
}

// SetCommandRunner sets a custom command runner (for testing)
func (e *Executor) SetCommandRunner(runner CommandRunner) {
	e.cmdRunner = runner
}

// ListTasks prints available tasks
func (e *Executor) ListTasks() error {
	names := e.tf.ListTaskNames()
	sort.Strings(names)

	_, _ = fmt.Fprintln(e.w, "Available tasks:")
	for _, name := range names {
		task := e.tf.Tasks[name]
		if task.Desc != "" {
			_, _ = fmt.Fprintf(e.w, "  %-20s %s\n", name, task.Desc)
		} else {
			_, _ = fmt.Fprintf(e.w, "  %s\n", name)
		}
	}

	return nil
}

// ShowSummary shows detailed summary of tasks
func (e *Executor) ShowSummary(taskNames []string) error {
	for _, name := range taskNames {
		task := e.tf.GetTask(name)
		if task == nil {
			return fmt.Errorf("task %q not found", name)
		}

		_, _ = fmt.Fprintf(e.w, "Task: %s\n", name)
		if task.Desc != "" {
			_, _ = fmt.Fprintf(e.w, "Description: %s\n", task.Desc)
		}
		if task.Summary != "" {
			_, _ = fmt.Fprintf(e.w, "Summary:\n%s\n", task.Summary)
		}
		if len(task.Deps) > 0 {
			deps := make([]string, len(task.Deps))
			for i, d := range task.Deps {
				deps[i] = d.Task
			}
			_, _ = fmt.Fprintf(e.w, "Dependencies: %s\n", strings.Join(deps, ", "))
		}
		if len(task.Cmds) > 0 {
			_, _ = fmt.Fprintln(e.w, "Commands:")
			for _, cmd := range task.Cmds {
				if cmd.Task != "" {
					_, _ = fmt.Fprintf(e.w, "  - task: %s\n", cmd.Task)
				} else {
					_, _ = fmt.Fprintf(e.w, "  - %s\n", cmd.Cmd)
				}
			}
		}
		_, _ = fmt.Fprintln(e.w)
	}
	return nil
}

// RunTask executes a task and its dependencies
func (e *Executor) RunTask(ctx context.Context, name string) error {
	// Validate dependencies
	if err := e.resolver.ValidateDeps(name); err != nil {
		return err
	}

	// Get execution order
	order, err := e.resolver.ResolveDeps(name)
	if err != nil {
		return err
	}

	// Execute in order
	for _, taskName := range order {
		if err := e.executeTask(ctx, taskName); err != nil {
			return err
		}
	}

	return nil
}

// executeTask executes a single task (without dependencies)
func (e *Executor) executeTask(ctx context.Context, name string) error {
	// Skip if already executed
	if e.executed[name] {
		return nil
	}

	task := e.tf.GetTask(name)
	if task == nil {
		return fmt.Errorf("task %q not found", name)
	}

	// Check status (up-to-date check) unless force
	if !e.opts.Force && len(task.Status) > 0 {
		upToDate, err := e.checkStatus(ctx, task)
		if err == nil && upToDate {
			if e.opts.Verbose {
				_, _ = fmt.Fprintf(e.w, "task: %s is up to date\n", name)
			}
			e.executed[name] = true
			return nil
		}
	}

	// Print task name
	if !e.opts.Silent && !task.Silent {
		_, _ = fmt.Fprintf(e.w, "task: %s\n", name)
	}

	// Create variable resolver
	resolver := NewVarResolver(e.tf.Vars, task.Vars, e.tf.Env)

	// Collect deferred commands
	var deferredCmds []Command

	// Execute commands
	for _, cmd := range task.Cmds {
		if cmd.Defer {
			deferredCmds = append(deferredCmds, cmd)
			continue
		}

		if err := e.executeCommand(ctx, cmd, resolver, task.Silent); err != nil {
			// Execute deferred commands before returning error
			e.executeDeferredCommands(ctx, deferredCmds, resolver, task.Silent)
			if !cmd.IgnoreError {
				return fmt.Errorf("task %s: %w", name, err)
			}
		}
	}

	// Execute deferred commands
	e.executeDeferredCommands(ctx, deferredCmds, resolver, task.Silent)

	e.executed[name] = true
	return nil
}

// executeDeferredCommands runs deferred commands in reverse order
func (e *Executor) executeDeferredCommands(ctx context.Context, cmds []Command, resolver *VarResolver, silent bool) {
	for i := len(cmds) - 1; i >= 0; i-- {
		_ = e.executeCommand(ctx, cmds[i], resolver, silent)
	}
}

// executeCommand executes a single command
func (e *Executor) executeCommand(ctx context.Context, cmd Command, resolver *VarResolver, taskSilent bool) error {
	// Handle task reference
	if cmd.Task != "" {
		return e.RunTask(ctx, cmd.Task)
	}

	// Expand variables
	cmdStr := resolver.Expand(cmd.Cmd)

	// Validate that it's an omni command
	if !isOmniCommand(cmdStr) {
		return fmt.Errorf("only omni commands are supported: %s", cmdStr)
	}

	// Print command (unless silent)
	silent := taskSilent || cmd.Silent || e.opts.Silent
	if !silent && e.opts.Verbose {
		_, _ = fmt.Fprintf(e.w, "  $ %s\n", cmdStr)
	}

	// Dry run: don't actually execute
	if e.opts.DryRun {
		_, _ = fmt.Fprintf(e.w, "  [dry-run] %s\n", cmdStr)
		return nil
	}

	// Parse and execute
	args := parseCommand(cmdStr)
	if len(args) == 0 {
		return nil
	}

	// Remove "omni" prefix if present
	if args[0] == "omni" {
		args = args[1:]
	}

	return e.cmdRunner.Run(ctx, e.w, args)
}

// checkStatus checks if a task is up-to-date
func (e *Executor) checkStatus(ctx context.Context, task *Task) (bool, error) {
	// Status commands should all succeed for task to be up-to-date
	// Since we can't exec external commands, we only support omni commands for status
	resolver := NewVarResolver(e.tf.Vars, task.Vars, e.tf.Env)

	for _, statusCmd := range task.Status {
		cmdStr := resolver.Expand(statusCmd)
		if !isOmniCommand(cmdStr) {
			// Can't check status with non-omni command
			return false, fmt.Errorf("status check requires omni command: %s", cmdStr)
		}

		args := parseCommand(cmdStr)
		if len(args) == 0 {
			continue
		}

		if args[0] == "omni" {
			args = args[1:]
		}

		// Suppress output for status checks
		if err := e.cmdRunner.Run(ctx, io.Discard, args); err != nil {
			return false, nil // Status check failed, task needs to run
		}
	}

	return true, nil
}

// isOmniCommand checks if a command is an omni command
// Valid: "omni echo hello", "echo hello" (implicit omni), "ls"
// Invalid: "/bin/ls", "bash -c 'ls'" (external commands)
func isOmniCommand(cmd string) bool {
	cmd = strings.TrimSpace(cmd)

	// Explicit omni prefix
	if strings.HasPrefix(cmd, "omni ") {
		return true
	}

	// Check for external command indicators
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return true
	}

	firstWord := parts[0]

	// Path-based commands are external
	if strings.Contains(firstWord, "/") || strings.Contains(firstWord, "\\") {
		return false
	}

	// Known external shell commands
	externalCmds := map[string]bool{
		"bash": true, "sh": true, "zsh": true, "fish": true,
		"powershell": true, "pwsh": true, "cmd": true,
		"python": true, "python3": true, "ruby": true, "perl": true,
		"node": true, "npm": true, "npx": true,
		"go": true, "cargo": true, "rustc": true,
		"make": true, "cmake": true, "msbuild": true,
		"docker": true, "kubectl": true,
		"git": true, // git should use omni commands
	}

	if externalCmds[firstWord] {
		return false
	}

	// Assume other single commands are omni subcommands
	return true
}

// parseCommand splits a command string into arguments
// Handles basic quoting
func parseCommand(cmd string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)

	cmd = strings.TrimSpace(cmd)

	for i := 0; i < len(cmd); i++ {
		c := cmd[i]

		if inQuote {
			if c == quoteChar {
				inQuote = false
			} else {
				current.WriteByte(c)
			}
		} else {
			switch c {
			case '"', '\'':
				inQuote = true
				quoteChar = c
			case ' ', '\t':
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
			default:
				current.WriteByte(c)
			}
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

// Run executes an omni command (default implementation)
// This will be overridden to use the actual omni command registry
func (r *DefaultCommandRunner) Run(ctx context.Context, w io.Writer, args []string) error {
	// This is a placeholder - the actual implementation will hook into
	// omni's command registry to execute commands internally
	return fmt.Errorf("command execution not implemented: %v", args)
}
