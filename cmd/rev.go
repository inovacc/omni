package cmd

import (
	"os"

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
		opts.JSON, _ = cmd.Flags().GetBool("json")
		return rev.RunRev(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(revCmd)

	revCmd.Flags().Bool("json", false, "output as JSON")
}
