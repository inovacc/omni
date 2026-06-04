package cmd

import (
	"github.com/inovacc/omni/internal/cli/reprocheck"
	"github.com/spf13/cobra"
)

var reprocheckCmd = &cobra.Command{
	Use:   "reprocheck --a FILE [--a FILE...] --b FILE [--b FILE...]",
	Short: "Fail if any A/B build artifact pair differs (reproducible-build gate)",
	Long: `Compare two index-aligned sets of built artifacts by sha256 and exit
non-zero (cmderr.ErrConflict) on any drift. Used by the v1.0 release dual-build
job to dogfood reproducibility across all six targets.

Example:
  omni reprocheck --a buildA/omni-linux-amd64 --b buildB/omni-linux-amd64`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		a, _ := cmd.Flags().GetStringArray("a")
		b, _ := cmd.Flags().GetStringArray("b")
		return reprocheck.Run(cmd.OutOrStdout(), reprocheck.Options{A: a, B: b})
	},
}

func init() {
	reprocheckCmd.Flags().StringArray("a", nil, "artifact path from build A (repeatable)")
	reprocheckCmd.Flags().StringArray("b", nil, "artifact path from build B (repeatable)")
	rootCmd.AddCommand(reprocheckCmd)
}
