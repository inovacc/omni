package cmd

import (
	"fmt"

	"github.com/inovacc/omni/internal/flags"
	"github.com/spf13/cobra"
)

var loggerCmd = &cobra.Command{
	Use:   "logger",
	Short: "Configure omni command logging",
	Long: `Configure omni command logging by outputting shell export statements.

Usage with eval to set environment variables:
  eval "$(omni logger --path /path/to/omni.log)"

To disable logging:
  eval "$(omni logger --disable)"

Environment variables set:
  OMNI_LOG_ENABLED - Set to "true" to enable logging
  OMNI_LOG_PATH    - Path to the log file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logPath, _ := cmd.Flags().GetString("path")
		disable, _ := cmd.Flags().GetBool("disable")
		status, _ := cmd.Flags().GetBool("status")

		if err := flags.IgnoreCommand("logger"); err != nil {
			return err
		}

		if status {
			return printStatus(cmd)
		}

		if disable {
			return flags.DisableFeature("logger")
		}

		if logPath == "" {
			return fmt.Errorf("--path is required (or use --disable to turn off logging)")
		}

		return flags.EnableFeature("logger", logPath)
	},
}

func printStatus(cmd *cobra.Command) error {
	logPath := flags.GetFeatureData("logger")

	if logPath == "" {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Logging: disabled (not configured)")
		return nil
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Logging: Enabled\n")
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Log path: %s\n", logPath)

	return nil
}

func init() {
	rootCmd.AddCommand(loggerCmd)

	loggerCmd.Flags().StringP("path", "p", "", "Path to the log file")
	loggerCmd.Flags().BoolP("disable", "d", false, "Disable logging (unset environment variables)")
	loggerCmd.Flags().BoolP("status", "s", false, "Show current logging status")
}
