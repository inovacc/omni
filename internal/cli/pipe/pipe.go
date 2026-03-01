package pipe

import (
	"bytes"
	"fmt"
	"io"

	"github.com/inovacc/omni/internal/cli/command"
	"github.com/inovacc/omni/pkg/cobra/helper/output"
	"github.com/spf13/cobra"
)

// Options configure the pipe command behavior
type Options struct {
	OutputFormat output.Format // output format (text/json)
	Separator    string        // --sep: command separator (default "|")
	Verbose      bool          // --verbose: show intermediate steps
	VarName      string        // --var: variable name for output substitution (default "OUT")
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

// CommandRegistry holds references to available commands.
// It tries the unified command.Registry first, then falls back to Cobra dispatch.
type CommandRegistry struct {
	RootCmd  *cobra.Command
	Unified  *command.Registry
}

// NewRegistry creates a new command registry with Cobra fallback only.
func NewRegistry(rootCmd *cobra.Command) *CommandRegistry {
	return &CommandRegistry{RootCmd: rootCmd}
}

// NewRegistryWithUnified creates a registry that tries unified commands first.
func NewRegistryWithUnified(rootCmd *cobra.Command, unified *command.Registry) *CommandRegistry {
	return &CommandRegistry{RootCmd: rootCmd, Unified: unified}
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
		// Apply variable substitution from previous output
		prevOutput := input.String()
		substitutedCmds, isIteration := substituteVariables(cmdStr, prevOutput, opts.VarName)

		for _, subCmd := range substitutedCmds {
			cmdResult := CommandResult{Command: subCmd}

			// Parse command and arguments
			cmdParts := parseCommandLine(subCmd)
			if len(cmdParts) == 0 {
				continue
			}

			// Execute command
			var output bytes.Buffer

			// For iteration, don't pass stdin (each command is independent)
			var cmdInput io.Reader
			if !isIteration {
				cmdInput = &input
			}

			err := executeCommand(registry, cmdParts, cmdInput, &output)
			if err != nil {
				cmdResult.Error = err.Error()
				result.Success = false
				result.Error = fmt.Sprintf("command %d failed: %s", i+1, err.Error())
				result.Commands = append(result.Commands, cmdResult)

				if !isIteration {
					break
				}

				continue
			}

			cmdResult.Output = output.String()
			result.Commands = append(result.Commands, cmdResult)

			// Verbose output
			if opts.Verbose {
				_, _ = fmt.Fprintf(w, "=== Step %d: %s ===\n", i+1, subCmd)
				_, _ = fmt.Fprintln(w, output.String())
			}

			// Pass output to next command's input (only for non-iteration)
			if !isIteration {
				input = output
			}
		}

		if !result.Success && !isIteration {
			break
		}
	}

	result.Output = input.String()

	// Output result
	f := output.New(w, opts.OutputFormat)
	if f.IsJSON() {
		return f.Print(result)
	}

	if !opts.Verbose {
		_, _ = fmt.Fprint(w, result.Output)
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Error)
	}

	return nil
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
		// Apply variable substitution from previous output
		prevOutput := inputBuf.String()
		substitutedCmds, isIteration := substituteVariables(cmdStr, prevOutput, opts.VarName)

		for _, subCmd := range substitutedCmds {
			cmdResult := CommandResult{Command: subCmd}

			cmdParts := parseCommandLine(subCmd)
			if len(cmdParts) == 0 {
				continue
			}

			var output bytes.Buffer

			// For iteration, don't pass stdin (each command is independent)
			var cmdInput io.Reader
			if !isIteration {
				cmdInput = inputBuf
			}

			err := executeCommand(registry, cmdParts, cmdInput, &output)
			if err != nil {
				cmdResult.Error = err.Error()
				result.Success = false
				result.Error = fmt.Sprintf("command %d failed: %s", i+1, err.Error())
				result.Commands = append(result.Commands, cmdResult)

				if !isIteration {
					break
				}

				continue
			}

			cmdResult.Output = output.String()
			result.Commands = append(result.Commands, cmdResult)

			if opts.Verbose {
				_, _ = fmt.Fprintf(w, "=== Step %d: %s ===\n", i+1, subCmd)
				_, _ = fmt.Fprintln(w, output.String())
			}

			// Pass output to next command's input (only for non-iteration)
			if !isIteration {
				inputBuf = &output
			}
		}

		if !result.Success && !isIteration {
			break
		}
	}

	result.Output = inputBuf.String()

	f := output.New(w, opts.OutputFormat)
	if f.IsJSON() {
		return f.Print(result)
	}

	if !opts.Verbose {
		_, _ = fmt.Fprint(w, result.Output)
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Error)
	}

	return nil
}
