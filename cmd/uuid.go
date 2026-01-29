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

Versions:
  4  Random UUID (default) - fully random, no ordering
  7  Time-ordered UUID - timestamp + random, sortable

  -v, --version=N UUID version (4 or 7, default 4)
  -n, --count=N   generate N UUIDs (default 1)
  -u, --upper     output in uppercase
  -x, --no-dashes output without dashes
  --json          output as JSON

Examples:
  omni uuid                  # generate one UUID v4
  omni uuid -v 7             # generate time-ordered UUID v7
  omni uuid -n 5             # generate 5 UUIDs
  omni uuid -u               # uppercase output
  omni uuid -x               # no dashes (32 hex chars)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := uuid.UUIDOptions{}

		opts.Version, _ = cmd.Flags().GetInt("version")
		opts.Count, _ = cmd.Flags().GetInt("count")
		opts.Upper, _ = cmd.Flags().GetBool("upper")
		opts.NoDashes, _ = cmd.Flags().GetBool("no-dashes")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return uuid.RunUUID(os.Stdout, opts)
	},
}

func init() {
	rootCmd.AddCommand(uuidCmd)

	uuidCmd.Flags().IntP("version", "v", 4, "UUID version (4 or 7)")
	uuidCmd.Flags().IntP("count", "n", 1, "generate N UUIDs")
	uuidCmd.Flags().BoolP("upper", "u", false, "output in uppercase")
	uuidCmd.Flags().BoolP("no-dashes", "x", false, "output without dashes")
	uuidCmd.Flags().Bool("json", false, "output as JSON")
}
