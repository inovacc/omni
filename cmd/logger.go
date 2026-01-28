package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/inovacc/omni/internal/logger"
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

		if status {
			return printStatus(cmd)
		}

		if disable {
			return printDisable(cmd)
		}

		if logPath == "" {
			return fmt.Errorf("--path is required (or use --disable to turn off logging)")
		}

		return printEnable(cmd, logPath)
	},
}

func printEnable(cmd *cobra.Command, logPath string) error {
	if runtime.GOOS == "windows" {
		// PowerShell format
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "$env:%s = \"true\"\n", logger.EnvLogEnabled)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "$env:%s = \"%s\"\n", logger.EnvLogPath, logPath)
	} else {
		// Bash/Zsh format
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "export %s=true\n", logger.EnvLogEnabled)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "export %s=%s\n", logger.EnvLogPath, logPath)
	}

	return nil
}

func printDisable(cmd *cobra.Command) error {
	if runtime.GOOS == "windows" {
		// PowerShell format
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Remove-Item Env:%s -ErrorAction SilentlyContinue\n", logger.EnvLogEnabled)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Remove-Item Env:%s -ErrorAction SilentlyContinue\n", logger.EnvLogPath)
	} else {
		// Bash/Zsh format
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "unset %s\n", logger.EnvLogEnabled)
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "unset %s\n", logger.EnvLogPath)
	}

	return nil
}

func printStatus(cmd *cobra.Command) error {
	enabled := os.Getenv(logger.EnvLogEnabled)
	logPath := os.Getenv(logger.EnvLogPath)

	if enabled == "" && logPath == "" {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Logging: disabled (not configured)")
		return nil
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Logging: %s\n", enabledStr(enabled))
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Log path: %s\n", valueOrNone(logPath))

	return nil
}

func enabledStr(val string) string {
	if val == "true" || val == "1" || val == "yes" {
		return "enabled"
	}

	return "disabled"
}

func valueOrNone(val string) string {
	if val == "" {
		return "(not set)"
	}

	return val
}

func init() {
	rootCmd.AddCommand(loggerCmd)

	loggerCmd.Flags().StringP("path", "p", "", "Path to the log file")
	loggerCmd.Flags().BoolP("disable", "d", false, "Disable logging (unset environment variables)")
	loggerCmd.Flags().BoolP("status", "s", false, "Show current logging status")
}
