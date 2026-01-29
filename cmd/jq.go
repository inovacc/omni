package cmd

import (
	"github.com/inovacc/omni/internal/cli/jq"
	"github.com/spf13/cobra"
)

// jqCmd represents the jq command
var jqCmd = &cobra.Command{
	Use:   "jq [OPTION]... FILTER [FILE]...",
	Short: "Command-line JSON processor",
	Long: `jq is a lightweight command-line JSON processor.

This is a simplified implementation supporting common operations:
  .           identity (output input unchanged)
  .field      access object field
  .field.sub  nested field access
  .[n]        array index
  .[]         iterate array elements
  keys        get object/array keys
  length      get length
  type        get type name

  -r          output raw strings (no quotes)
  -c          compact output
  -s          slurp: read all inputs into array
  -n          null input
  --tab       use tabs for indentation

Examples:
  echo '{"name":"John"}' | omni jq '.name'
  echo '[1,2,3]' | omni jq '.[]'
  echo '{"a":{"b":1}}' | omni jq '.a.b'
  omni jq -r '.name' data.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := jq.JqOptions{}

		opts.Raw, _ = cmd.Flags().GetBool("raw-output")
		opts.Compact, _ = cmd.Flags().GetBool("compact-output")
		opts.Slurp, _ = cmd.Flags().GetBool("slurp")
		opts.NullInput, _ = cmd.Flags().GetBool("null-input")
		opts.Tab, _ = cmd.Flags().GetBool("tab")
		opts.Sort, _ = cmd.Flags().GetBool("sort-keys")

		return jq.RunJq(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(jqCmd)

	jqCmd.Flags().BoolP("raw-output", "r", false, "output raw strings")
	jqCmd.Flags().BoolP("compact-output", "c", false, "compact output")
	jqCmd.Flags().BoolP("slurp", "s", false, "read all inputs into array")
	jqCmd.Flags().BoolP("null-input", "n", false, "don't read any input")
	jqCmd.Flags().Bool("tab", false, "use tabs for indentation")
	jqCmd.Flags().BoolP("sort-keys", "S", false, "sort object keys")
}
