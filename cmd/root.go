package cmd

import (
	"github.com/inovacc/omni/internal/flags"
	"github.com/inovacc/omni/internal/logger"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "omni",
	Short: "Go-native replacement for common shell utilities",
	Long: `omni is a cross-platform, safe, Go-native replacement for common shell utilities,
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
			log.LogCommand(args)
		}
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		log := logger.Get()
		if log != nil {
			return log.Close()
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
