package cmd

import (
	"github.com/inovacc/omni/internal/cli/env"
	"github.com/spf13/cobra"
)

// envCmd represents the env command
var envCmd = &cobra.Command{
	Use:   "env [NAME...]",
	Short: "Print environment variables",
	Long: `Print the values of the specified environment variables.
If no NAME is specified, print all environment variables.

Examples:
  omni env                        # print all environment variables
  omni env PATH                   # print a single variable
  omni env HOME USER              # print several variables`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := env.EnvOptions{}

		opts.NullTerminated, _ = cmd.Flags().GetBool("null")
		opts.Unset, _ = cmd.Flags().GetString("unset")
		opts.Ignore, _ = cmd.Flags().GetBool("ignore-environment")
		opts.OutputFormat = getOutputOpts(cmd).GetFormat()

		return env.RunEnv(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(envCmd)

	envCmd.Flags().BoolP("null", "0", false, "end each output line with NUL, not newline")
	envCmd.Flags().StringP("unset", "u", "", "remove variable from the environment")
	envCmd.Flags().BoolP("ignore-environment", "i", false, "start with an empty environment")

}
