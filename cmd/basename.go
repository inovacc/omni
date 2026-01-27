package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli"

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
		suffix, _ := cmd.Flags().GetString("suffix")

		// Handle suffix from positional argument (traditional usage)
		names := args
		if suffix == "" && len(args) > 1 {
			suffix = args[1]
			names = args[:1]
		}

		return cli.RunBasename(os.Stdout, names, suffix)
	},
}

func init() {
	rootCmd.AddCommand(basenameCmd)

	basenameCmd.Flags().StringP("suffix", "s", "", "remove a trailing SUFFIX")
}
