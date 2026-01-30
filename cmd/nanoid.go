package cmd

import (
	"github.com/inovacc/omni/internal/cli/nanoid"
	"github.com/spf13/cobra"
)

// nanoidCmd represents the nanoid command
var nanoidCmd = &cobra.Command{
	Use:   "nanoid [OPTION]...",
	Short: "Generate compact, URL-safe unique IDs",
	Long: `Generate NanoIDs - compact, URL-safe, unique string identifiers.

NanoIDs are:
- Shorter than UUID (21 chars vs 36)
- URL-safe (using A-Za-z0-9_-)
- Cryptographically secure
- Customizable length and alphabet

Default: 21 characters from URL-safe alphabet (64 chars)

  -n, --count=N     generate N NanoIDs (default 1)
  -l, --length=N    length of NanoID (default 21)
  -a, --alphabet=S  custom alphabet
  --json            output as JSON

Examples:
  omni nanoid                    # generate one NanoID (21 chars)
  omni nanoid -n 5               # generate 5 NanoIDs
  omni nanoid -l 10              # shorter 10-char NanoID
  omni nanoid -a "0123456789"    # numeric only
  omni nanoid --json             # JSON output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := nanoid.Options{}

		opts.Count, _ = cmd.Flags().GetInt("count")
		opts.Length, _ = cmd.Flags().GetInt("length")
		opts.Alphabet, _ = cmd.Flags().GetString("alphabet")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return nanoid.RunNanoID(cmd.OutOrStdout(), opts)
	},
}

func init() {
	rootCmd.AddCommand(nanoidCmd)

	nanoidCmd.Flags().IntP("count", "n", 1, "generate N NanoIDs")
	nanoidCmd.Flags().IntP("length", "l", 21, "length of NanoID")
	nanoidCmd.Flags().StringP("alphabet", "a", "", "custom alphabet")
	nanoidCmd.Flags().Bool("json", false, "output as JSON")
}
