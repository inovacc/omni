package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/id"
	"github.com/spf13/cobra"
)

// idCmd represents the id command
var idCmd = &cobra.Command{
	Use:   "id [OPTION]... [USER]",
	Short: "Print user and group information",
	Long: `Print user and group information for the specified USER,
or (when USER omitted) for the current user.

  -g, --group   print only the effective group ID
  -G, --groups  print all group IDs
  -n, --name    print a name instead of a number, for -ugG
  -r, --real    print the real ID instead of the effective ID, with -ugG
  -u, --user    print only the effective user ID`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := id.IDOptions{}

		opts.User, _ = cmd.Flags().GetBool("user")
		opts.Group, _ = cmd.Flags().GetBool("group")
		opts.Groups, _ = cmd.Flags().GetBool("groups")
		opts.Name, _ = cmd.Flags().GetBool("name")
		opts.Real, _ = cmd.Flags().GetBool("real")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		if len(args) > 0 {
			opts.Username = args[0]
		}

		return id.RunID(os.Stdout, opts)
	},
}

func init() {
	rootCmd.AddCommand(idCmd)

	idCmd.Flags().BoolP("user", "u", false, "print only the effective user ID")
	idCmd.Flags().BoolP("group", "g", false, "print only the effective group ID")
	idCmd.Flags().BoolP("groups", "G", false, "print all group IDs")
	idCmd.Flags().BoolP("name", "n", false, "print a name instead of a number")
	idCmd.Flags().BoolP("real", "r", false, "print the real ID instead of the effective ID")
	idCmd.Flags().Bool("json", false, "output as JSON")
}
