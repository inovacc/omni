package cmd

import (
	"github.com/inovacc/omni/internal/cli/column"
	"github.com/spf13/cobra"
)

// columCmd represents the column command
var columCmd = &cobra.Command{
	Use:     "column [OPTION]... [FILE]...",
	Aliases: []string{"colum"},
	Short:   "Columnate lists",
	Long: `Format input into multiple columns.

With no FILE, or when FILE is -, read standard input.

  -t, --table            determine column count based on input
  -s, --separator=STRING delimiter characters for -t option
  -o, --output-separator=STRING  output separator for table mode
  -c, --columns=N        output width in characters (default 80)
  -x, --fillrows         fill rows before columns
  -R, --right            right-align columns`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := column.ColumnOptions{}

		opts.Table, _ = cmd.Flags().GetBool("table")
		opts.Separator, _ = cmd.Flags().GetString("separator")
		opts.OutputSep, _ = cmd.Flags().GetString("output-separator")
		opts.Columns, _ = cmd.Flags().GetInt("columns")
		opts.FillRows, _ = cmd.Flags().GetBool("fillrows")
		opts.Right, _ = cmd.Flags().GetBool("right")
		opts.JSON, _ = cmd.Flags().GetBool("json")
		opts.ColumnHeaders, _ = cmd.Flags().GetString("headers")

		return column.RunColumn(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(columCmd)

	columCmd.Flags().BoolP("table", "t", false, "determine column count based on input")
	columCmd.Flags().StringP("separator", "s", "", "delimiter characters for -t option")
	columCmd.Flags().StringP("output-separator", "o", "", "output separator for table mode")
	columCmd.Flags().IntP("columns", "c", 80, "output width in characters")
	columCmd.Flags().BoolP("fillrows", "x", false, "fill rows before columns")
	columCmd.Flags().BoolP("right", "R", false, "right-align columns")
	columCmd.Flags().BoolP("json", "J", false, "output as JSON")
	columCmd.Flags().StringP("headers", "H", "", "column headers (comma-separated)")
}
