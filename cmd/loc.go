package cmd

import (
	"github.com/inovacc/omni/internal/cli/loc"
	"github.com/spf13/cobra"
)

var locCmd = &cobra.Command{
	Use:   "loc [PATH]...",
	Short: "Count lines of code by language",
	Long: `Count lines of code, comments, and blanks by programming language.

Similar to tokei, cloc, or sloccount. Automatically detects language by
file extension and counts code, comments, and blank lines.

Supported languages: Go, Rust, JavaScript, TypeScript, Python, Java, C, C++,
C#, Ruby, PHP, Swift, Kotlin, Scala, Shell, Lua, SQL, HTML, CSS, and more.

Default excludes: .git, node_modules, vendor, __pycache__, .idea, .vscode,
target, build, dist

Examples:
  omni loc                         # count in current directory
  omni loc ./src                   # count in specific directory
  omni loc --json .                # output as JSON
  omni loc --exclude test .        # exclude "test" directory
  omni loc --hidden .              # include hidden files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := loc.Options{}

		opts.Exclude, _ = cmd.Flags().GetStringSlice("exclude")
		opts.Hidden, _ = cmd.Flags().GetBool("hidden")
		opts.OutputFormat = getOutputOpts(cmd).GetFormat()

		return loc.RunLoc(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(locCmd)

	locCmd.Flags().StringSliceP("exclude", "e", nil, "directories to exclude")
	locCmd.Flags().Bool("hidden", false, "include hidden files")
}
