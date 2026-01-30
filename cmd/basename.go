package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/basename"
	"github.com/spf13/cobra"
)

// basenameCmd represents the basename command
var basenameCmd = &cobra.Command{
	Use:   "basename NAME [SUFFIX]",
	Short: "Strip directory and suffix from file names",
	Long: `Print NAME with any leading directory components removed.
If specified, also remove a trailing SUFFIX.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := basename.BasenameOptions{}
		opts.Suffix, _ = cmd.Flags().GetString("suffix")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		// Handle suffix from positional argument (traditional usage)
		names := args
		if opts.Suffix == "" && len(args) > 1 {
			opts.Suffix = args[1]
			names = args[:1]
		}

		return basename.RunBasename(cmd.OutOrStdout(), names, opts)
	},
}

func init() {
	rootCmd.AddCommand(basenameCmd)

	basenameCmd.Flags().StringP("suffix", "s", "", "remove a trailing SUFFIX")
	basenameCmd.Flags().Bool("json", false, "output as JSON")
}
