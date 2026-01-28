package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli/awk"

	"github.com/spf13/cobra"
)

// awkCmd represents the awk command
var awkCmd = &cobra.Command{
	Use:   "awk [OPTION]... 'program' [FILE]...",
	Short: "Pattern scanning and processing language",
	Long: `Awk scans each input file for lines that match any of a set of patterns.

This is a simplified subset of AWK supporting:
  - Field access: $0 (whole line), $1, $2, etc.
  - Pattern blocks: BEGIN{}, END{}, /regex/{}
  - Print statements: print, print $1, print $1,$2
  - Built-in variable: NF (number of fields)

  -F fs          use fs for the input field separator
  -v var=value   assign value to variable var

Examples:
  omni awk '{print $1}' file.txt          # print first field
  omni awk -F: '{print $1}' /etc/passwd   # use : as separator
  omni awk '/pattern/{print}' file.txt    # print matching lines
  omni awk 'BEGIN{print "start"} {print} END{print "end"}' file
  omni awk '{print $1, $NF}' file.txt     # print first and last field`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := awk.AwkOptions{}

		opts.FieldSeparator, _ = cmd.Flags().GetString("field-separator")

		// Parse -v assignments
		vars, _ := cmd.Flags().GetStringSlice("assign")
		if len(vars) > 0 {
			opts.Variables = make(map[string]string)
			for _, v := range vars {
				// Split on first =
				for i, c := range v {
					if c == '=' {
						opts.Variables[v[:i]] = v[i+1:]
						break
					}
				}
			}
		}

		return awk.RunAwk(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(awkCmd)

	awkCmd.Flags().StringP("field-separator", "F", "", "use FS for the input field separator")
	awkCmd.Flags().StringSliceP("assign", "v", nil, "assign value to variable (var=value)")
}
