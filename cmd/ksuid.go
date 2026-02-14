package cmd

import (
	"github.com/inovacc/omni/internal/cli/ksuid"
	"github.com/spf13/cobra"
)

// ksuidCmd represents the ksuid command
var ksuidCmd = &cobra.Command{
	Use:   "ksuid [OPTION]...",
	Short: "Generate K-Sortable Unique IDentifiers",
	Long: `Generate KSUIDs (K-Sortable Unique IDentifiers).

KSUIDs are 27-character, base62-encoded identifiers that are:
- Globally unique
- Naturally sortable by generation time
- URL-safe and case-sensitive

Structure: 4-byte timestamp + 16-byte random payload

  -n, --count=N   generate N KSUIDs (default 1)
  --json          output as JSON

Examples:
  omni ksuid                  # generate one KSUID
  omni ksuid -n 5             # generate 5 KSUIDs
  omni ksuid --json           # JSON output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := ksuid.Options{}

		opts.Count, _ = cmd.Flags().GetInt("count")
		opts.OutputFormat = getOutputOpts(cmd).GetFormat()

		return ksuid.RunKSUID(cmd.OutOrStdout(), opts)
	},
}

func init() {
	rootCmd.AddCommand(ksuidCmd)

	ksuidCmd.Flags().IntP("count", "n", 1, "generate N KSUIDs")
}
