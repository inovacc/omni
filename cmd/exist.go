package cmd

import (
	"errors"
	"io"
	"os"

	"github.com/inovacc/omni/internal/cli/exist"
	"github.com/spf13/cobra"
)

var existCmd = &cobra.Command{
	Use:   "exist",
	Short: "Check if files, directories, commands, env vars, processes, or ports exist",
	Long: `Check existence of various targets with proper exit codes for scripting.

Exit status:
  0  target exists
  1  target not found

Examples:
  omni exist file go.mod              # check if regular file exists
  omni exist dir cmd                  # check if directory exists
  omni exist path go.mod              # check if any path exists
  omni exist command go               # check if command is in PATH
  omni exist env PATH                 # check if env var is set
  omni exist process 1234             # check if PID is running
  omni exist port 8080                # check if TCP port is listening
  omni exist -q file go.mod && echo yes  # quiet mode for scripting`,
}

var existFileCmd = &cobra.Command{
	Use:   "file <path>",
	Short: "Check if a regular file exists",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runExist(cmd, args, exist.RunFile)
	},
}

var existDirCmd = &cobra.Command{
	Use:   "dir <path>",
	Short: "Check if a directory exists",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runExist(cmd, args, exist.RunDir)
	},
}

var existPathCmd = &cobra.Command{
	Use:   "path <path>",
	Short: "Check if any path exists (file, dir, symlink)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runExist(cmd, args, exist.RunPath)
	},
}

var existCommandCmd = &cobra.Command{
	Use:   "command <name>",
	Short: "Check if a command exists in PATH",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runExist(cmd, args, exist.RunCommand)
	},
}

var existEnvCmd = &cobra.Command{
	Use:   "env <name>",
	Short: "Check if an environment variable is set",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runExist(cmd, args, exist.RunEnv)
	},
}

var existProcessCmd = &cobra.Command{
	Use:   "process <name|pid>",
	Short: "Check if a process is running",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runExist(cmd, args, exist.RunProcess)
	},
}

var existPortCmd = &cobra.Command{
	Use:   "port <number>",
	Short: "Check if a TCP port is listening",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runExist(cmd, args, exist.RunPort)
	},
}

type existRunFunc func(w io.Writer, target string, opts exist.Options) error

func runExist(cmd *cobra.Command, args []string, fn existRunFunc) error {
	opts := exist.Options{}
	opts.Quiet, _ = cmd.Flags().GetBool("quiet")
	opts.JSON, _ = cmd.Flags().GetBool("json")

	err := fn(cmd.OutOrStdout(), args[0], opts)
	if errors.Is(err, exist.ErrNotFound) {
		os.Exit(1)
	}

	return err
}

func init() {
	rootCmd.AddCommand(existCmd)

	existCmd.PersistentFlags().BoolP("quiet", "q", false, "suppress output")
	existCmd.PersistentFlags().Bool("json", false, "output as JSON")

	existCmd.AddCommand(existFileCmd)
	existCmd.AddCommand(existDirCmd)
	existCmd.AddCommand(existPathCmd)
	existCmd.AddCommand(existCommandCmd)
	existCmd.AddCommand(existEnvCmd)
	existCmd.AddCommand(existProcessCmd)
	existCmd.AddCommand(existPortCmd)
}
