package cmd

import (
	"github.com/inovacc/omni/internal/cli/strings"
	"github.com/spf13/cobra"
)

var (
	stringsMinLength int
	stringsOffset    string
)

var stringsCmd = &cobra.Command{
	Use:   "strings [OPTION]... [FILE]...",
	Short: "Print the printable strings in files",
	Long: `Print the sequences of printable characters in files.

  -n, --bytes=MIN   print sequences of at least MIN characters (default 4)
  -t, --radix=TYPE  print offset in TYPE: d=decimal, o=octal, x=hex

Examples:
  omni strings binary.exe         # extract strings from binary
  omni strings -n 8 file.bin      # strings of at least 8 chars
  omni strings -t x program       # show hex offsets`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		opts := strings.StringsOptions{
			MinLength: stringsMinLength,
			Offset:    stringsOffset,
			JSON:      jsonOutput,
		}

		return strings.RunStrings(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(stringsCmd)

	stringsCmd.Flags().IntVarP(&stringsMinLength, "bytes", "n", 4, "minimum string length")
	stringsCmd.Flags().StringVarP(&stringsOffset, "radix", "t", "", "print offset (d/o/x)")
	stringsCmd.Flags().Bool("json", false, "output as JSON")
}
