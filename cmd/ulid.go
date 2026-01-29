package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/ulid"
	"github.com/spf13/cobra"
)

// ulidCmd represents the ulid command
var ulidCmd = &cobra.Command{
	Use:   "ulid [OPTION]...",
	Short: "Generate Universally Unique Lexicographically Sortable Identifiers",
	Long: `Generate ULIDs (Universally Unique Lexicographically Sortable Identifiers).

ULIDs are 26-character, Crockford's base32-encoded identifiers that are:
- 128-bit compatible with UUID
- Lexicographically sortable
- Case insensitive
- URL-safe (no special characters)

Structure: 48-bit timestamp (ms) + 80-bit randomness

  -n, --count=N   generate N ULIDs (default 1)
  -l, --lower     output in lowercase
  --json          output as JSON

Examples:
  omni ulid                   # generate one ULID
  omni ulid -n 5              # generate 5 ULIDs
  omni ulid -l                # lowercase output
  omni ulid --json            # JSON output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := ulid.Options{}

		opts.Count, _ = cmd.Flags().GetInt("count")
		opts.Lower, _ = cmd.Flags().GetBool("lower")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return ulid.RunULID(os.Stdout, opts)
	},
}

func init() {
	rootCmd.AddCommand(ulidCmd)

	ulidCmd.Flags().IntP("count", "n", 1, "generate N ULIDs")
	ulidCmd.Flags().BoolP("lower", "l", false, "output in lowercase")
	ulidCmd.Flags().Bool("json", false, "output as JSON")
}
