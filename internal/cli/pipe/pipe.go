package pipe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Options configure the pipe command behavior
type Options struct {
	JSON      bool   // --json: output pipeline result as JSON
	Separator string // --sep: command separator (default "|")
	Verbose   bool   // --verbose: show intermediate steps
}

// Result represents the pipeline execution result
type Result struct {
	Commands []CommandResult `json:"commands"`
	Output   string          `json:"output"`
	Success  bool            `json:"success"`
	Error    string          `json:"error,omitempty"`
}

// CommandResult represents a single command's execution
type CommandResult struct {
	Command string `json:"command"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// CommandRegistry holds references to available commands
type CommandRegistry struct {
	RootCmd *cobra.Command
}

// NewRegistry creates a new command registry
func NewRegistry(rootCmd *cobra.Command) *CommandRegistry {
	return &CommandRegistry{RootCmd: rootCmd}
}

// Run executes a pipeline of commands
func Run(w io.Writer, args []string, opts Options, registry *CommandRegistry) error {
	if len(args) == 0 {
		return fmt.Errorf("pipe: no commands provided")
	}

	// Parse commands from args
	commands := parseCommands(args, opts.Separator)
	if len(commands) == 0 {
		return fmt.Errorf("pipe: no commands provided")
	}

	result := Result{
		Commands: make([]CommandResult, 0, len(commands)),
		Success:  true,
	}

	// Execute pipeline
	var input bytes.Buffer

	for i, cmdStr := range commands {
		cmdResult := CommandResult{Command: cmdStr}

		// Parse command and arguments
		cmdParts := parseCommandLine(cmdStr)
		if len(cmdParts) == 0 {
			continue
		}

		// Execute command
		var output bytes.Buffer

		err := executeCommand(registry, cmdParts, &input, &output)
		if err != nil {
			cmdResult.Error = err.Error()
			result.Success = false
			result.Error = fmt.Sprintf("command %d failed: %s", i+1, err.Error())
			result.Commands = append(result.Commands, cmdResult)

			break
		}

		cmdResult.Output = output.String()
		result.Commands = append(result.Commands, cmdResult)

		// Verbose output
		if opts.Verbose {
			_, _ = fmt.Fprintf(w, "=== Step %d: %s ===\n", i+1, cmdStr)
			_, _ = fmt.Fprintln(w, output.String())
		}

		// Pass output to next command's input
		input = output
	}

	result.Output = input.String()

	// Output result
	if opts.JSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		return enc.Encode(result)
	}

	if !opts.Verbose {
		_, _ = fmt.Fprint(w, result.Output)
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Error)
	}

	return nil
}

// parseCommands splits args into individual commands
func parseCommands(args []string, separator string) []string {
	if separator == "" {
		separator = "|"
	}

	// Join all args for processing
	joined := strings.Join(args, " ")

	// Check for brace syntax: {cmd1}, {cmd2}, {cmd3}
	if strings.Contains(joined, "{") && strings.Contains(joined, "}") {
		return parseBraceCommands(joined)
	}

	// Check if single arg with separators
	if len(args) == 1 && strings.Contains(args[0], separator) {
		parts := strings.Split(args[0], separator)
		commands := make([]string, 0, len(parts))

		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				commands = append(commands, p)
			}
		}

		return commands
	}

	// Check if args contain separator as separate element
	hasSeparator := slices.Contains(args, separator)

	if hasSeparator {
		var (
			commands []string
			current  []string
		)

		for _, arg := range args {
			if arg == separator {
				if len(current) > 0 {
					commands = append(commands, strings.Join(current, " "))
					current = nil
				}
			} else {
				current = append(current, arg)
			}
		}

		if len(current) > 0 {
			commands = append(commands, strings.Join(current, " "))
		}

		return commands
	}

	// If multiple args and at least some have spaces, treat each as a separate command
	// This handles: omni pipe "cat file.txt" "grep pattern" "sort"
	if len(args) > 1 {
		someHaveSpaces := false

		for _, arg := range args {
			if strings.Contains(arg, " ") {
				someHaveSpaces = true

				break
			}
		}

		if someHaveSpaces {
			return args
		}
	}

	// Otherwise, join all args as a single command, then try to split by separator
	if strings.Contains(joined, separator) {
		parts := strings.Split(joined, separator)
		commands := make([]string, 0, len(parts))

		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				commands = append(commands, p)
			}
		}

		return commands
	}

	// Single command - join all args as one command
	return []string{strings.Join(args, " ")}
}

// parseBraceCommands parses commands in {cmd1}, {cmd2} format
func parseBraceCommands(input string) []string {
	var (
		commands []string
		current  strings.Builder
	)

	depth := 0

	for _, r := range input {
		switch r {
		case '{':
			depth++

			if depth == 1 {
				continue // Don't include opening brace
			}
		case '}':
			depth--

			if depth == 0 {
				cmd := strings.TrimSpace(current.String())
				if cmd != "" {
					commands = append(commands, cmd)
				}

				current.Reset()

				continue // Don't include closing brace
			}
		case ',':
			if depth == 0 {
				continue // Skip commas between commands
			}
		}

		if depth > 0 {
			current.WriteRune(r)
		}
	}

	// Handle any remaining content (in case of missing closing brace)
	if current.Len() > 0 {
		cmd := strings.TrimSpace(current.String())
		if cmd != "" {
			commands = append(commands, cmd)
		}
	}

	return commands
}

// parseCommandLine splits a command string into parts, respecting quotes
func parseCommandLine(cmdLine string) []string {
	var (
		parts   []string
		current strings.Builder
		inQuote rune
		escaped bool
	)

	for _, r := range cmdLine {
		if escaped {
			current.WriteRune(r)

			escaped = false

			continue
		}

		if r == '\\' {
			escaped = true

			continue
		}

		if inQuote != 0 {
			if r == inQuote {
				inQuote = 0
			} else {
				current.WriteRune(r)
			}

			continue
		}

		if r == '"' || r == '\'' {
			inQuote = r

			continue
		}

		if r == ' ' || r == '\t' {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}

			continue
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// executeCommand executes a single omni command using cmd.SetIn for stdin
func executeCommand(registry *CommandRegistry, cmdParts []string, stdin io.Reader, stdout io.Writer) error {
	if registry == nil || registry.RootCmd == nil {
		return fmt.Errorf("command registry not initialized")
	}

	// Find the command
	cmd, args, err := registry.RootCmd.Find(cmdParts)
	if err != nil {
		return fmt.Errorf("unknown command: %s", cmdParts[0])
	}

	// Check if we found a valid command (not root)
	if cmd == registry.RootCmd {
		return fmt.Errorf("unknown command: %s", cmdParts[0])
	}

	// Set up input/output using Cobra's built-in methods
	// Commands now use cmd.InOrStdin() which will return this reader
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
		return fmt.Errorf("parsing flags: %w", err)
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

// RunWithInput executes a pipeline with initial input
func RunWithInput(w io.Writer, input string, args []string, opts Options, registry *CommandRegistry) error {
	if len(args) == 0 {
		return fmt.Errorf("pipe: no commands provided")
	}

	// Parse commands
	commands := parseCommands(args, opts.Separator)
	if len(commands) == 0 {
		return fmt.Errorf("pipe: no commands provided")
	}

	result := Result{
		Commands: make([]CommandResult, 0, len(commands)),
		Success:  true,
	}

	// Start with provided input
	inputBuf := bytes.NewBufferString(input)

	for i, cmdStr := range commands {
		cmdResult := CommandResult{Command: cmdStr}

		cmdParts := parseCommandLine(cmdStr)
		if len(cmdParts) == 0 {
			continue
		}

		var output bytes.Buffer

		err := executeCommand(registry, cmdParts, inputBuf, &output)
		if err != nil {
			cmdResult.Error = err.Error()
			result.Success = false
			result.Error = fmt.Sprintf("command %d failed: %s", i+1, err.Error())
			result.Commands = append(result.Commands, cmdResult)

			break
		}

		cmdResult.Output = output.String()
		result.Commands = append(result.Commands, cmdResult)

		if opts.Verbose {
			_, _ = fmt.Fprintf(w, "=== Step %d: %s ===\n", i+1, cmdStr)
			_, _ = fmt.Fprintln(w, output.String())
		}

		inputBuf = &output
	}

	result.Output = inputBuf.String()

	if opts.JSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		return enc.Encode(result)
	}

	if !opts.Verbose {
		_, _ = fmt.Fprint(w, result.Output)
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Error)
	}

	return nil
}
