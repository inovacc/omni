package cmd

import (
	"fmt"
	"time"

	"github.com/inovacc/omni/internal/cli/output"
	"github.com/inovacc/omni/internal/cli/timecmd"
	"github.com/spf13/cobra"
)

// timeCmd represents the time command
var timeCmd = &cobra.Command{
	Use:   "time",
	Short: "Time a simple command or give resource usage",
	Long: `The time utility executes and times the specified utility. After the
utility finishes, time writes to the standard error stream, the total
time elapsed.

Note: Since omni doesn't execute external commands, this command
provides timing utilities and can measure internal operations.

Examples:
  omni time sleep 2    # Time a sleep operation
  omni time            # Just show current time info`,
	RunE: func(cmd *cobra.Command, args []string) error {
		outFmt := getOutputOpts(cmd).GetFormat()
		w := cmd.OutOrStdout()

		if len(args) == 0 {
			now := time.Now()
			f := output.New(w, outFmt)
			if f.IsJSON() {
				return f.Print(map[string]any{
					"time":      now.Format(time.RFC3339),
					"unix":      now.Unix(),
					"unix_nano": now.UnixNano(),
				})
			}

			_, _ = fmt.Fprintf(w, "Current time: %s\n", now.Format(time.RFC3339))
			_, _ = fmt.Fprintf(w, "Unix timestamp: %d\n", now.Unix())
			_, _ = fmt.Fprintf(w, "Unix nano: %d\n", now.UnixNano())

			return nil
		}

		// If first arg is "sleep", do a timed sleep
		if args[0] == "sleep" && len(args) > 1 {
			duration, err := time.ParseDuration(args[1] + "s")
			if err != nil {
				duration, err = time.ParseDuration(args[1])
				if err != nil {
					return fmt.Errorf("invalid duration: %s", args[1])
				}
			}

			_, err = timecmd.RunTimeJSON(cmd.ErrOrStderr(), outFmt, func() error {
				time.Sleep(duration)
				return nil
			})

			return err
		}

		return fmt.Errorf("time: cannot execute external commands, use 'time sleep N' for timing")
	},
}

func init() {
	rootCmd.AddCommand(timeCmd)
}
