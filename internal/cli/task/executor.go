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
	w         io.Writer
	tf        *Taskfile
	opts      Options
	resolver  *DependencyResolver
	executed  map[string]bool
	cmdRunner CommandRunner
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

	// Check if command can be mapped to an omni built-in
	mappedCmd, isOmni := mapToOmniCommand(cmdStr)

	// If not an omni command and external not allowed, error
	if !isOmni && !e.opts.AllowExternal {
		return fmt.Errorf("unknown command: %s\n\nHint: Use --allow-external to run external commands", firstWord(cmdStr))
	}

	// Print command (unless silent)
	silent := taskSilent || cmd.Silent || e.opts.Silent
	if !silent && e.opts.Verbose {
		if isOmni && !strings.HasPrefix(cmdStr, "omni ") {
			_, _ = fmt.Fprintf(e.w, "  $ %s  (using omni %s)\n", cmdStr, firstWord(cmdStr))
		} else {
			_, _ = fmt.Fprintf(e.w, "  $ %s\n", cmdStr)
		}
	}

	// Dry run: don't actually execute
	if e.opts.DryRun {
		if isOmni && !strings.HasPrefix(cmdStr, "omni ") {
			_, _ = fmt.Fprintf(e.w, "  [dry-run] %s  (using omni %s)\n", cmdStr, firstWord(cmdStr))
		} else {
			_, _ = fmt.Fprintf(e.w, "  [dry-run] %s\n", cmdStr)
		}

		return nil
	}

	// Parse and execute
	args := parseCommand(mappedCmd)
	if len(args) == 0 {
		return nil
	}

	// Remove "omni" prefix if present
	if args[0] == "omni" {
		args = args[1:]
	}

	// If it's an omni command, use the omni runner directly
	// If it's external, the HybridCommandRunner will route to shell
	return e.cmdRunner.Run(ctx, e.w, args)
}

// checkStatus checks if a task is up-to-date
func (e *Executor) checkStatus(ctx context.Context, task *Task) (bool, error) {
	// Status commands should all succeed for task to be up-to-date
	resolver := NewVarResolver(e.tf.Vars, task.Vars, e.tf.Env)

	for _, statusCmd := range task.Status {
		cmdStr := resolver.Expand(statusCmd)
		if !e.opts.AllowExternal && !isOmniCommand(cmdStr) {
			// Can't check status with non-omni command unless external allowed
			return false, fmt.Errorf("status check requires omni command (or use --allow-external): %s", cmdStr)
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
			// Status check failed means task is not up-to-date, not an error
			return false, nil //nolint:nilerr // intentional: failed status check = task needs to run
		}
	}

	return true, nil
}

// firstWord extracts the first word from a command string
func firstWord(cmd string) string {
	parts := strings.Fields(strings.TrimSpace(cmd))
	if len(parts) == 0 {
		return ""
	}

	return parts[0]
}

// omniCommands is the set of commands that omni provides as built-ins.
// These commands will be automatically routed to omni instead of external shell.
var omniCommands = map[string]bool{
	// File operations
	"cat": true, "cp": true, "mv": true, "rm": true, "mkdir": true, "rmdir": true,
	"ls": true, "find": true, "touch": true, "chmod": true, "chown": true,
	"ln": true, "readlink": true, "realpath": true, "stat": true, "file": true,
	"dd": true, "df": true, "du": true,

	// Text processing
	"head": true, "tail": true, "tac": true, "rev": true,
	"grep": true, "egrep": true, "fgrep": true, "rg": true,
	"sed": true, "awk": true, "cut": true, "sort": true, "uniq": true,
	"wc": true, "nl": true, "tr": true, "fold": true, "column": true,
	"join": true, "paste": true, "comm": true, "cmp": true, "diff": true,
	"shuf": true, "split": true, "strings": true,

	// Archive/compression
	"tar": true, "zip": true, "unzip": true,
	"gzip": true, "gunzip": true, "zcat": true,
	"bzip2": true, "bunzip2": true, "bzcat": true,
	"xz": true, "unxz": true, "xzcat": true,

	// Data formats
	"jq": true, "yq": true, "json": true, "yaml": true, "toml": true, "xml": true,
	"csv": true, "html": true, "css": true,

	// Encoding
	"base32": true, "base64": true, "base58": true,
	"md5sum": true, "sha256sum": true, "sha512sum": true, "hash": true,
	"hex": true, "xxd": true,

	// System info
	"echo": true, "printf": true, "env": true, "pwd": true,
	"uname": true, "arch": true, "whoami": true, "id": true,
	"ps": true, "kill": true, "free": true, "uptime": true, "which": true,

	// Utilities
	"date": true, "sleep": true, "seq": true, "yes": true, "tree": true,
	"basename": true, "dirname": true, "xargs": true, "time": true,
	"curl": true, "watch": true,

	// Security
	"encrypt": true, "decrypt": true, "random": true, "uuid": true,

	// Omni-specific
	"loc": true, "dotenv": true,
}

// mapToOmniCommand checks if a command can be handled by omni's built-in
// and returns true if it should be routed to omni instead of external shell.
func mapToOmniCommand(cmd string) (string, bool) {
	cmd = strings.TrimSpace(cmd)

	// Already has omni prefix
	if strings.HasPrefix(cmd, "omni ") {
		return cmd, true
	}

	// Check first word
	first := firstWord(cmd)
	if first == "" {
		return cmd, false
	}

	// Path-based commands should not be mapped
	if strings.Contains(first, "/") || strings.Contains(first, "\\") {
		return cmd, false
	}

	// Check if it's a known omni command
	if omniCommands[first] {
		return cmd, true
	}

	return cmd, false
}

// isOmniCommand checks if a command is an omni command
// Uses the omniCommands map to determine if a command is built-in
func isOmniCommand(cmd string) bool {
	_, isOmni := mapToOmniCommand(cmd)
	return isOmni
}

// parseCommand splits a command string into arguments
// Handles basic quoting
func parseCommand(cmd string) []string {
	var (
		args    []string
		current strings.Builder
	)

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
