package cmd

import (
	"github.com/inovacc/omni/internal/cli/echo"
	"github.com/spf13/cobra"
)

var echoOpts echo.EchoOptions

// echoCmd represents the echo command
var echoCmd = &cobra.Command{
	Use:   "echo [STRING]...",
	Short: "Display a line of text",
	Long: `Echo the STRING(s) to standard output.

Examples:
  omni echo Hello World     # outputs 'Hello World'
  omni echo -n "no newline" # outputs without trailing newline
  omni echo -e "tab\there"  # outputs with tab character`,
	RunE: func(cmd *cobra.Command, args []string) error {
		echoOpts.OutputFormat = getOutputOpts(cmd).GetFormat()
		return echo.RunEcho(cmd.OutOrStdout(), args, echoOpts)
	},
}

func init() {
	rootCmd.AddCommand(echoCmd)

	echoCmd.Flags().BoolVarP(&echoOpts.NoNewline, "no-newline", "n", false, "do not output the trailing newline")
	echoCmd.Flags().BoolVarP(&echoOpts.EnableEscapes, "escape", "e", false, "enable interpretation of backslash escapes")
	echoCmd.Flags().BoolVarP(&echoOpts.DisableEscapes, "no-escape", "E", false, "disable interpretation of backslash escapes (default)")
}
