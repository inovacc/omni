package cmd

import (
	"github.com/inovacc/omni/internal/cli/tac"
	"github.com/spf13/cobra"
)

// tacCmd represents the tac command
var tacCmd = &cobra.Command{
	Use:   "tac [OPTION]... [FILE]...",
	Short: "Concatenate and print files in reverse",
	Long: `Write each FILE to standard output, last line first.

With no FILE, or when FILE is -, read standard input.

  -b, --before             attach the separator before instead of after
  -r, --regex              interpret the separator as a regular expression
  -s, --separator=STRING   use STRING as the separator instead of newline`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := tac.TacOptions{}

		opts.Before, _ = cmd.Flags().GetBool("before")
		opts.Regex, _ = cmd.Flags().GetBool("regex")
		opts.Separator, _ = cmd.Flags().GetString("separator")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return tac.RunTac(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(tacCmd)

	tacCmd.Flags().BoolP("before", "b", false, "attach the separator before instead of after")
	tacCmd.Flags().BoolP("regex", "r", false, "interpret the separator as a regular expression")
	tacCmd.Flags().StringP("separator", "s", "", "use STRING as the separator instead of newline")
	tacCmd.Flags().Bool("json", false, "output as JSON")
}
