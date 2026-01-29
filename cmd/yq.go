package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/yq"
	"github.com/spf13/cobra"
)

// yqCmd represents the yq command
var yqCmd = &cobra.Command{
	Use:   "yq [OPTION]... FILTER [FILE]...",
	Short: "Command-line YAML processor",
	Long: `yq is a lightweight command-line YAML processor.

Uses the same filter syntax as jq:
  .           identity
  .field      access field
  .field.sub  nested access
  .[n]        array index
  .[]         iterate array

  -r          output raw strings
  -c          compact JSON output
  -o json     output as JSON
  -o yaml     output as YAML (default)
  -n          null input

Examples:
  omni yq '.name' config.yaml
  omni yq -o json '.' config.yaml    # convert YAML to JSON
  echo "name: John" | omni yq '.name'
  omni yq '.items[]' data.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := yq.YqOptions{}

		opts.Raw, _ = cmd.Flags().GetBool("raw-output")
		opts.Compact, _ = cmd.Flags().GetBool("compact-output")
		opts.NullInput, _ = cmd.Flags().GetBool("null-input")
		opts.Indent, _ = cmd.Flags().GetInt("indent")

		outputFormat, _ := cmd.Flags().GetString("output-format")
		if outputFormat == "json" {
			opts.OutputJSON = true
		} else {
			opts.OutputYAML = true
		}

		return yq.RunYq(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(yqCmd)

	yqCmd.Flags().BoolP("raw-output", "r", false, "output raw strings")
	yqCmd.Flags().BoolP("compact-output", "c", false, "compact output")
	yqCmd.Flags().BoolP("null-input", "n", false, "don't read any input")
	yqCmd.Flags().StringP("output-format", "o", "yaml", "output format (yaml or json)")
	yqCmd.Flags().IntP("indent", "I", 2, "indentation level")
}
