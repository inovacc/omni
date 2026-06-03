package cmd

// helplint:ignore — Long strings need omni-usage examples added in a future pass.

import (
	"errors"
	"fmt"
	"os"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/internal/flags"
	"github.com/inovacc/omni/internal/logger"
	"github.com/spf13/cobra"
)

const omniBanner = `
  ██████╗ ███╗   ███╗███╗   ██╗██╗
 ██╔═══██╗████╗ ████║████╗  ██║██║
 ██║   ██║██╔████╔██║██╔██╗ ██║██║
 ██║   ██║██║╚██╔╝██║██║╚██╗██║██║
 ╚██████╔╝██║ ╚═╝ ██║██║ ╚████║██║
  ╚═════╝ ╚═╝     ╚═╝╚═╝  ╚═══╝╚═╝
  Shell utilities, rewritten in Go.
`

var rootCmd = &cobra.Command{
	Use:   "omni",
	Short: "Go-native replacement for common shell utilities",
	Long: omniBanner + `omni is a cross-platform, safe, Go-native replacement for common shell utilities,
designed for Taskfile, CI/CD, and enterprise environments.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := flags.ExportFlagsToEnv(); err != nil {
			return
		}

		// Skip logging for commands that shouldn't create log files
		if flags.ShouldIgnoreCommand(cmd.Name()) {
			return
		}

		log := logger.Init(cmd.Name())
		if log.IsActive() {
			// Wrap stdout/stderr to capture output
			stdout, stderr := log.StartExecution(cmd.Name(), args, cmd.OutOrStdout(), cmd.ErrOrStderr())
			cmd.SetOut(stdout)
			cmd.SetErr(stderr)
		}
	},
}

// panicExitCode is the exit code used when a command handler panics. It is
// deliberately distinct from the cmderr sentinel codes (1-6) and from the Go
// runtime's own panic exit code (2, which aliases ErrInvalidInput) so callers
// can tell an internal panic apart from a normal classified failure.
const panicExitCode = 70

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
//
// A deferred recover converts any panic in a command handler into a classified
// error so logging is finalized, the error is printed in the standard
// "Error: <msg>" format, and the process exits with a deterministic code
// instead of leaking a raw runtime stack trace and exit code 2.
func Execute() {
	var err error

	defer func() {
		if r := recover(); r != nil {
			err = cmderr.WithExitCode(fmt.Errorf("panic: %v", r), panicExitCode)
		}
		finalize(err)
	}()

	err = rootCmd.Execute()
}

// finalize finalizes logging with the command/panic error, prints a
// non-silent error to stderr, and exits with the mapped exit code. It is the
// single completion path shared by the normal and panic-recovery flows.
func finalize(err error) {
	// Finalize logging with the actual command error
	log := logger.Get()
	if log != nil && log.IsActive() {
		log.EndExecution(err)
		_ = log.Close()
	}

	if err != nil {
		// Print error unless it's a silent exit (e.g. grep no-match)
		var silent *cmderr.SilentError
		if !errors.As(err, &silent) {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		}

		os.Exit(cmderr.ExitCodeFor(err))
	}
}

func init() {
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().Bool("json", false, "output as JSON")
	rootCmd.PersistentFlags().Bool("table", false, "output as aligned table")
}
