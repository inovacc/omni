package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli"
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
		return cli.RunRev(os.Stdout, args)
	},
}

func init() {
	rootCmd.AddCommand(revCmd)
}
