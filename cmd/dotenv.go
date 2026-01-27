package cmd

import (
	"os"

	"github.com/inovacc/goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// dotenvCmd represents the dotenv command
var dotenvCmd = &cobra.Command{
	Use:   "dotenv [OPTION]... [FILE]...",
	Short: "Load environment variables from .env files",
	Long: `Parse and display environment variables from .env files.

With no FILE, reads from .env in the current directory.

  -e, --export    output as export statements (for shell sourcing)
  -q, --quiet     suppress warnings
  -x, --expand    expand variables in values

The .env file format:
  # Comments start with #
  KEY=value
  KEY="quoted value"
  KEY='single quoted'
  export KEY=value    # export prefix is optional

Examples:
  goshell dotenv                    # display vars from .env
  goshell dotenv .env.local         # display vars from specific file
  goshell dotenv -e                 # output as export statements
  eval $(goshell dotenv -e)         # load vars into shell
  goshell dotenv -x                 # expand ${VAR} references`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.DotenvOptions{}

		opts.Export, _ = cmd.Flags().GetBool("export")
		opts.Quiet, _ = cmd.Flags().GetBool("quiet")
		opts.Expand, _ = cmd.Flags().GetBool("expand")

		return cli.RunDotenv(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(dotenvCmd)

	dotenvCmd.Flags().BoolP("export", "e", false, "output as export statements")
	dotenvCmd.Flags().BoolP("quiet", "q", false, "suppress warnings")
	dotenvCmd.Flags().BoolP("expand", "x", false, "expand variables in values")
}
