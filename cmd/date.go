package cmd

import (
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/date"
	"github.com/spf13/cobra"
)

// dateCmd represents the date command
var dateCmd = &cobra.Command{
	Use:   "date [+FORMAT]",
	Short: "Print the current date and time",
	Long: `Display the current time in the given FORMAT, or set the system date.

FORMAT controls the output. Interpreted sequences are:
  %Y   year
  %m   month (01..12)
  %d   day of month (01..31)
  %H   hour (00..23)
  %M   minute (00..59)
  %S   second (00..60)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := date.DateOptions{}

		opts.UTC, _ = cmd.Flags().GetBool("utc")
		opts.ISO, _ = cmd.Flags().GetBool("iso-8601")

		// Check if format is provided as argument (like +%Y-%m-%d)
		if len(args) > 0 && len(args[0]) > 0 && args[0][0] == '+' {
			opts.Format = convertDateFormat(args[0][1:])
		}

		return date.RunDate(os.Stdout, opts)
	},
}

func init() {
	rootCmd.AddCommand(dateCmd)

	dateCmd.Flags().BoolP("utc", "u", false, "print Coordinated Universal Time (UTC)")
	dateCmd.Flags().Bool("iso-8601", false, "output date/time in ISO 8601 format")
}

// convertDateFormat converts strftime-style format to Go's time format
func convertDateFormat(format string) string {
	replacements := map[string]string{
		"%Y": "2006",
		"%m": "01",
		"%d": "02",
		"%H": "15",
		"%M": "04",
		"%S": "05",
		"%y": "06",
		"%b": "Jan",
		"%B": "January",
		"%a": "Mon",
		"%A": "Monday",
		"%p": "PM",
		"%Z": "MST",
		"%z": "-0700",
		"%n": "\n",
		"%t": "\t",
		"%%": "%",
	}

	result := format
	for k, v := range replacements {
		result = replaceAll(result, k, v)
	}

	return result
}

func replaceAll(s, old, replacement string) string {
	var result strings.Builder

	for i := 0; i < len(s); {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result.WriteString(replacement)

			i += len(old)
		} else {
			result.WriteString(string(s[i]))
			i++
		}
	}

	return result.String()
}
