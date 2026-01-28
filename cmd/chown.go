package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli/chown"

	"github.com/spf13/cobra"
)

// chownCmd represents the chown command
var chownCmd = &cobra.Command{
	Use:   "chown [OPTION]... OWNER[:GROUP] FILE...",
	Short: "Change file owner and group",
	Long: `Change the owner and/or group of each FILE to OWNER and/or GROUP.

OWNER can be specified as:
  - User name (e.g., root)
  - Numeric user ID (e.g., 0)
  - OWNER:GROUP to change both
  - OWNER: to change owner and set group to owner's login group
  - :GROUP to change only the group

Options:
  -R, --recursive   operate on files and directories recursively
  -v, --verbose     output a diagnostic for every file processed
  -c, --changes     like verbose but report only when a change is made
  -f, --silent      suppress most error messages
  -h, --no-dereference  affect symbolic links instead of referenced file
      --reference   use RFILE's owner and group
      --preserve-root  fail to operate recursively on '/'`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := chown.ChownOptions{}

		opts.Recursive, _ = cmd.Flags().GetBool("recursive")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		opts.Changes, _ = cmd.Flags().GetBool("changes")
		opts.Silent, _ = cmd.Flags().GetBool("silent")
		opts.NoDereference, _ = cmd.Flags().GetBool("no-dereference")
		opts.Reference, _ = cmd.Flags().GetString("reference")
		opts.PreserveRoot, _ = cmd.Flags().GetBool("preserve-root")

		return chown.RunChown(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(chownCmd)

	chownCmd.Flags().BoolP("recursive", "R", false, "operate on files and directories recursively")
	chownCmd.Flags().BoolP("verbose", "v", false, "output a diagnostic for every file processed")
	chownCmd.Flags().BoolP("changes", "c", false, "like verbose but report only when a change is made")
	chownCmd.Flags().BoolP("silent", "f", false, "suppress most error messages")
	chownCmd.Flags().BoolP("no-dereference", "h", false, "affect symbolic links instead of referenced file")
	chownCmd.Flags().String("reference", "", "use RFILE's owner and group")
	chownCmd.Flags().Bool("preserve-root", false, "fail to operate recursively on '/'")
}
