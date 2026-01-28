package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli/arch"
	"github.com/spf13/cobra"
)

var archCmd = &cobra.Command{
	Use:   "arch",
	Short: "Print machine architecture",
	Long: `Print the machine hardware name (similar to uname -m).

Examples:
  omni arch    # x86_64, aarch64, etc.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return arch.RunArch(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(archCmd)
}
