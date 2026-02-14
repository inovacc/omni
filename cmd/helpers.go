package cmd

import (
	"github.com/inovacc/omni/internal/cli/output"
	"github.com/spf13/cobra"
)

// getOutputOpts reads the global --json and --table flags from the command
// and returns an output.Options. During migration, commands that still
// declare local --json flags will shadow the persistent one — Cobra resolves
// local flags first, so there is no breakage.
func getOutputOpts(cmd *cobra.Command) output.Options {
	j, _ := cmd.Flags().GetBool("json")
	tbl, _ := cmd.Flags().GetBool("table")

	return output.Options{JSON: j, Table: tbl}
}
