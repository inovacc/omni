package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/which"
	"github.com/spf13/cobra"
)

var whichAll bool

var whichCmd = &cobra.Command{
	Use:   "which [OPTION]... COMMAND...",
	Short: "Locate a command",
	Long: `Write the full path of COMMAND(s) to standard output.

  -a, --all   print all matching executables in PATH, not just the first

Examples:
  omni which go              # /usr/local/go/bin/go
  omni which python python3  # locate multiple commands
  omni which -a python       # show all python executables`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := which.WhichOptions{
			All: whichAll,
		}

		return which.RunWhich(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(whichCmd)

	whichCmd.Flags().BoolVarP(&whichAll, "all", "a", false, "print all matches")
}
