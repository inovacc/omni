package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/arch"
	"github.com/spf13/cobra"
)

var archJSON bool

var archCmd = &cobra.Command{
	Use:   "arch",
	Short: "Print machine architecture",
	Long: `Print the machine hardware name (similar to uname -m).

Examples:
  omni arch    # x86_64, aarch64, etc.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := arch.ArchOptions{JSON: archJSON}
		return arch.RunArch(cmd.OutOrStdout(), opts)
	},
}

func init() {
	rootCmd.AddCommand(archCmd)

	archCmd.Flags().BoolVar(&archJSON, "json", false, "output as JSON")
}
