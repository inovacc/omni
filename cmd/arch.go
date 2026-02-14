package cmd

import (
	"github.com/inovacc/omni/internal/cli/arch"
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
		opts := arch.ArchOptions{OutputFormat: getOutputOpts(cmd).GetFormat()}
		return arch.RunArch(cmd.OutOrStdout(), opts)
	},
}

func init() {
	rootCmd.AddCommand(archCmd)
}
