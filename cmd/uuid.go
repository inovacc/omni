package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/uuid"
	"github.com/spf13/cobra"
)

// uuidCmd represents the uuid command
var uuidCmd = &cobra.Command{
	Use:   "uuid [OPTION]...",
	Short: "Generate random UUIDs",
	Long: `Generate random UUIDs (Universally Unique Identifiers).

Generates version 4 (random) UUIDs per RFC 4122.

  -n, --count=N   generate N UUIDs (default 1)
  -u, --upper     output in uppercase
  -x, --no-dashes output without dashes

Examples:
  omni uuid                  # generate one UUID
  omni uuid -n 5             # generate 5 UUIDs
  omni uuid -u               # uppercase output
  omni uuid -x               # no dashes (32 hex chars)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := uuid.UUIDOptions{}

		opts.Count, _ = cmd.Flags().GetInt("count")
		opts.Upper, _ = cmd.Flags().GetBool("upper")
		opts.NoDashes, _ = cmd.Flags().GetBool("no-dashes")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return uuid.RunUUID(os.Stdout, opts)
	},
}

func init() {
	rootCmd.AddCommand(uuidCmd)

	uuidCmd.Flags().IntP("count", "n", 1, "generate N UUIDs")
	uuidCmd.Flags().BoolP("upper", "u", false, "output in uppercase")
	uuidCmd.Flags().BoolP("no-dashes", "x", false, "output without dashes")
	uuidCmd.Flags().Bool("json", false, "output as JSON")
}
