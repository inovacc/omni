package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/inovacc/omni/internal/cli/task"
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task [TASK...]",
	Short: "Run tasks defined in Taskfile.yml",
	Long: `A task runner that executes tasks defined in Taskfile.yml.

By default, only omni internal commands are supported. Use --allow-external
to enable execution of external shell commands (golangci-lint, go, npm, etc).

Examples:
  # List available tasks
  omni task --list

  # Run the default task
  omni task

  # Run a specific task
  omni task build

  # Run multiple tasks
  omni task build test

  # Run with external commands enabled
  omni task --allow-external lint

  # Dry run (show commands without executing)
  omni task --dry-run build

  # Force run even if up-to-date
  omni task --force build

  # Show task summary
  omni task --summary build

Taskfile Format:
  version: '3'

  vars:
    BUILD_DIR: ./build

  tasks:
    default:
      deps: [build]

    build:
      desc: Build the project
      cmds:
        - omni mkdir -p {{.BUILD_DIR}}
        - omni cp -r src/* {{.BUILD_DIR}}/

    lint:
      desc: Run linter (requires --allow-external)
      cmds:
        - golangci-lint run --fix ./...

    clean:
      desc: Clean build artifacts
      cmds:
        - omni rm -rf {{.BUILD_DIR}}

Supported Features:
  - Task dependencies (deps)
  - Variable expansion ({{.VAR}})
  - Task includes (includes)
  - Status checks for up-to-date detection
  - Deferred commands
  - Task aliases
  - External commands (with --allow-external)

Limitations:
  - Dynamic variables (sh:) are not supported`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := task.Options{}

		opts.Taskfile, _ = cmd.Flags().GetString("taskfile")
		opts.Dir, _ = cmd.Flags().GetString("dir")
		opts.List, _ = cmd.Flags().GetBool("list")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		opts.DryRun, _ = cmd.Flags().GetBool("dry-run")
		opts.Force, _ = cmd.Flags().GetBool("force")
		opts.Silent, _ = cmd.Flags().GetBool("silent")
		opts.Summary, _ = cmd.Flags().GetBool("summary")
		opts.AllowExternal, _ = cmd.Flags().GetBool("allow-external")

		// Create context that cancels on SIGINT/SIGTERM
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			cancel()
		}()

		return task.Run(ctx, cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(taskCmd)

	taskCmd.Flags().StringP("taskfile", "t", "", "path to Taskfile.yml")
	taskCmd.Flags().StringP("dir", "d", "", "working directory")
	taskCmd.Flags().BoolP("list", "l", false, "list available tasks")
	taskCmd.Flags().BoolP("verbose", "v", false, "verbose output")
	taskCmd.Flags().Bool("dry-run", false, "print commands without executing")
	taskCmd.Flags().BoolP("force", "f", false, "force run even if up-to-date")
	taskCmd.Flags().BoolP("silent", "s", false, "suppress output")
	taskCmd.Flags().Bool("summary", false, "show task summary")
	taskCmd.Flags().Bool("allow-external", false, "allow external (non-omni) commands")

	// Register the command runner factory
	task.CommandRunnerFactory = func(dir string, allowExternal bool) task.CommandRunner {
		omniRunner := task.NewCobraCommandRunner(rootCmd)
		if allowExternal {
			return task.NewHybridCommandRunner(omniRunner, dir)
		}

		return omniRunner
	}
}
