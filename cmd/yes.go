package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/inovacc/omni/internal/cli/yes"
	"github.com/spf13/cobra"
)

// yesCmd represents the yes command
var yesCmd = &cobra.Command{
	Use:   "yes [STRING]...",
	Short: "Output a string repeatedly until killed",
	Long: `Repeatedly output a line with all specified STRING(s), or 'y'.

Examples:
  omni yes              # outputs 'y' forever
  omni yes hello        # outputs 'hello' forever
  omni yes | head -5    # outputs 5 'y' lines`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE)
		go func() {
			<-sigCh
			cancel()
		}()

		return yes.RunYes(ctx, cmd.OutOrStdout(), args)
	},
}

func init() {
	rootCmd.AddCommand(yesCmd)
}
