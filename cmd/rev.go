package cmd

import (
	"github.com/inovacc/omni/internal/cli/rev"
	"github.com/spf13/cobra"
)

var revCmd = &cobra.Command{
	Use:   "rev [FILE]...",
	Short: "Reverse lines characterwise",
	Long: `Reverse the characters in each line of FILE(s) or standard input.

Examples:
  echo "hello" | omni rev     # olleh
  omni rev file.txt           # reverse each line in file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := rev.RevOptions{}
		opts.OutputFormat = getOutputOpts(cmd).GetFormat()
		return rev.RunRev(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(revCmd)

}
