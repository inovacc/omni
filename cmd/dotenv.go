package cmd

import (
	"github.com/inovacc/omni/internal/cli/dotenv"
	"github.com/spf13/cobra"
)

// dotenvCmd represents the dotenv command
var dotenvCmd = &cobra.Command{
	Use:   "dotenv [OPTION]... [FILE]...",
	Short: "Load environment variables from .env files",
	Long: `Parse and display environment variables from .env files.

With no FILE, reads from .env in the current directory.

  -e, --export      output as export statements (for shell sourcing)
  -s, --shell TYPE  target shell (auto, bash, zsh, fish, powershell, cmd, nushell)
  -q, --quiet       suppress warnings
  -x, --expand      expand variables in values

The .env file format:
  # Comments start with #
  KEY=value
  KEY="quoted value"
  KEY='single quoted'
  export KEY=value    # export prefix is optional

Shell export formats:
  bash/zsh:    export KEY="value"
  powershell:  $env:KEY = "value"
  cmd:         set KEY=value
  fish:        set -gx KEY "value"
  nushell:     $env.KEY = "value"

Examples:
  omni dotenv                    # display vars from .env
  omni dotenv .env.local         # display vars from specific file
  omni dotenv -e                 # output as export statements (auto-detect shell)
  omni dotenv -e -s powershell   # output for PowerShell
  omni dotenv -e -s fish         # output for Fish shell

Load into shell:
  Bash/Zsh:    eval $(omni dotenv -e)
  PowerShell:  omni dotenv -e -s powershell | Invoke-Expression
  Fish:        omni dotenv -e -s fish | source
  CMD:         for /f "tokens=*" %i in ('omni dotenv -e -s cmd') do %i`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := dotenv.DotenvOptions{}

		opts.Export, _ = cmd.Flags().GetBool("export")
		opts.Quiet, _ = cmd.Flags().GetBool("quiet")
		opts.Expand, _ = cmd.Flags().GetBool("expand")

		shellStr, _ := cmd.Flags().GetString("shell")
		opts.Shell = dotenv.ShellType(shellStr)

		return dotenv.RunDotenv(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(dotenvCmd)

	dotenvCmd.Flags().BoolP("export", "e", false, "output as export statements")
	dotenvCmd.Flags().StringP("shell", "s", "auto", "target shell (auto, bash, zsh, fish, powershell, cmd, nushell)")
	dotenvCmd.Flags().BoolP("quiet", "q", false, "suppress warnings")
	dotenvCmd.Flags().BoolP("expand", "x", false, "expand variables in values")
}
