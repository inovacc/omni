package cmd

import (
	"os"

	"github.com/inovacc/goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// envCmd represents the env command
var envCmd = &cobra.Command{
	Use:   "env [NAME...]",
	Short: "Print environment variables",
	Long: `Print the values of the specified environment variables.
If no NAME is specified, print all environment variables.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.EnvOptions{}

		opts.NullTerminated, _ = cmd.Flags().GetBool("null")
		opts.Unset, _ = cmd.Flags().GetString("unset")
		opts.Ignore, _ = cmd.Flags().GetBool("ignore-environment")

		return cli.RunEnv(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(envCmd)

	envCmd.Flags().BoolP("null", "0", false, "end each output line with NUL, not newline")
	envCmd.Flags().StringP("unset", "u", "", "remove variable from the environment")
	envCmd.Flags().BoolP("ignore-environment", "i", false, "start with an empty environment")
}
