package cmd

import (
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()

	// Finalize logging with the actual command error
	log := logger.Get()
	if log != nil && log.IsActive() {
		log.EndExecution(err)
		_ = log.Close()
	}

	if err != nil {
		os.Exit(cmderr.ExitCodeFor(err))
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().Bool("json", false, "output as JSON")
	rootCmd.PersistentFlags().Bool("table", false, "output as aligned table")
}
