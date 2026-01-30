package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/inovacc/omni/internal/cli/watch"
	"github.com/spf13/cobra"
)

// watchCmd represents the watch command
var watchCmd = &cobra.Command{
	Use:   "watch [OPTION]... COMMAND",
	Short: "Execute a program periodically, showing output fullscreen",
	Long: `Execute a command repeatedly, displaying its output.

Note: Since omni doesn't execute external commands, this version
monitors files or directories for changes.

  -n, --interval=SECS   seconds to wait between updates (default 2)
  -d, --differences     highlight differences between successive updates
  -t, --no-title        turn off header showing the command and time
  -b, --beep            beep if command has a non-zero exit
  -e, --errexit         exit if command has a non-zero exit
  -p, --precise         attempt run command in precise intervals
  -c, --color           interpret ANSI color and style sequences

Examples:
  omni watch -n 1 file myfile.txt    # Watch a file for changes
  omni watch -n 5 dir ./logs         # Watch a directory`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := watch.WatchOptions{}

		interval, _ := cmd.Flags().GetFloat64("interval")
		opts.Interval = time.Duration(interval * float64(time.Second))
		opts.Differences, _ = cmd.Flags().GetBool("differences")
		opts.NoTitle, _ = cmd.Flags().GetBool("no-title")
		opts.BeepOnError, _ = cmd.Flags().GetBool("beep")
		opts.ExitOnError, _ = cmd.Flags().GetBool("errexit")
		opts.Precise, _ = cmd.Flags().GetBool("precise")
		opts.Color, _ = cmd.Flags().GetBool("color")

		// Create context that cancels on SIGINT/SIGTERM
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			cancel()
		}()

		// Handle different watch modes
		if len(args) >= 2 {
			switch args[0] {
			case "file":
				return watch.WatchFile(ctx, args[1], func(path string) error {
					info, _ := os.Stat(path)
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[%s] File modified: %s (size: %d)\n",
						time.Now().Format("15:04:05"), path, info.Size())
					return nil
				}, opts.Interval)
			case "dir":
				return watch.WatchDir(ctx, args[1], func(event, path string) error {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s: %s\n",
						time.Now().Format("15:04:05"), event, path)
					return nil
				}, opts.Interval)
			}
		}

		// Default: just show periodic message
		return watch.RunWatch(ctx, cmd.OutOrStdout(), func() (string, error) {
			return fmt.Sprintf("Watching: %s\n", args), nil
		}, opts)
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.Flags().Float64P("interval", "n", 2, "seconds to wait between updates")
	watchCmd.Flags().BoolP("differences", "d", false, "highlight differences between updates")
	watchCmd.Flags().BoolP("no-title", "t", false, "turn off header")
	watchCmd.Flags().BoolP("beep", "b", false, "beep if command has a non-zero exit")
	watchCmd.Flags().BoolP("errexit", "e", false, "exit if command has a non-zero exit")
	watchCmd.Flags().BoolP("precise", "p", false, "attempt run command in precise intervals")
	watchCmd.Flags().BoolP("color", "c", false, "interpret ANSI color sequences")
}
