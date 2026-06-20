package cmd

import (
	"github.com/inovacc/omni/internal/cli/dirname"
	"github.com/spf13/cobra"
)

// dirnameCmd represents the dirname command
var dirnameCmd = &cobra.Command{
	Use:   "dirname [path...]",
	Short: "Strip last component from file name",
	Long: `Output each NAME with its last non-slash component and trailing slashes removed.

Examples:
  omni dirname /usr/bin/sort      # prints "/usr/bin"
  omni dirname src/main.go        # prints "src"`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := dirname.DirnameOptions{OutputFormat: getOutputOpts(cmd).GetFormat()}
		return dirname.RunDirname(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(dirnameCmd)
}
