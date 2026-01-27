package cmd

import (
	"os"

	"github.com/inovacc/goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// chmodCmd represents the chmod command
var chmodCmd = &cobra.Command{
	Use:   "chmod [OPTION]... MODE[,MODE]... FILE...",
	Short: "Change file mode bits",
	Long: `Change the mode of each FILE to MODE.

MODE can be:
  - Octal number (e.g., 755, 644)
  - Symbolic mode (e.g., u+x, go-w, a=rw)

Symbolic mode format: [ugoa][+-=][rwx]
  u = user, g = group, o = others, a = all
  + = add, - = remove, = = set exactly
  r = read, w = write, x = execute

Options:
  -R, --recursive  change files and directories recursively
  -v, --verbose    output a diagnostic for every file processed
  -c, --changes    like verbose but report only when a change is made
  -f, --silent     suppress most error messages
      --reference  use RFILE's mode instead of MODE values`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.ChmodOptions{}

		opts.Recursive, _ = cmd.Flags().GetBool("recursive")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		opts.Changes, _ = cmd.Flags().GetBool("changes")
		opts.Silent, _ = cmd.Flags().GetBool("silent")
		opts.Reference, _ = cmd.Flags().GetString("reference")

		return cli.RunChmod(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(chmodCmd)

	chmodCmd.Flags().BoolP("recursive", "R", false, "change files and directories recursively")
	chmodCmd.Flags().BoolP("verbose", "v", false, "output a diagnostic for every file processed")
	chmodCmd.Flags().BoolP("changes", "c", false, "like verbose but report only when a change is made")
	chmodCmd.Flags().BoolP("silent", "f", false, "suppress most error messages")
	chmodCmd.Flags().String("reference", "", "use RFILE's mode instead of MODE values")
}
