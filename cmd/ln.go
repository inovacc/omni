package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli"

	"github.com/spf13/cobra"
)

// lnCmd represents the ln command
var lnCmd = &cobra.Command{
	Use:   "ln [OPTION]... TARGET LINK_NAME",
	Short: "Make links between files",
	Long: `Create a link to TARGET with the name LINK_NAME.
Create hard links by default, symbolic links with --symbolic.

  -s, --symbolic     make symbolic links instead of hard links
  -f, --force        remove existing destination files
  -n, --no-dereference  treat LINK_NAME as a normal file if it is a symlink
  -v, --verbose      print name of each linked file
  -b, --backup       make a backup of each existing destination file
  -r, --relative     create symbolic links relative to link location`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.LnOptions{}

		opts.Symbolic, _ = cmd.Flags().GetBool("symbolic")
		opts.Force, _ = cmd.Flags().GetBool("force")
		opts.NoClobber, _ = cmd.Flags().GetBool("no-dereference")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		opts.Backup, _ = cmd.Flags().GetBool("backup")
		opts.Relative, _ = cmd.Flags().GetBool("relative")

		return cli.RunLn(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(lnCmd)

	lnCmd.Flags().BoolP("symbolic", "s", false, "make symbolic links instead of hard links")
	lnCmd.Flags().BoolP("force", "f", false, "remove existing destination files")
	lnCmd.Flags().BoolP("no-dereference", "n", false, "treat LINK_NAME as a normal file if it is a symlink")
	lnCmd.Flags().BoolP("verbose", "v", false, "print name of each linked file")
	lnCmd.Flags().BoolP("backup", "b", false, "make a backup of each existing destination file")
	lnCmd.Flags().BoolP("relative", "r", false, "create symbolic links relative to link location")
}
