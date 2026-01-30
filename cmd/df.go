package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/df"
	"github.com/spf13/cobra"
)

// dfCmd represents the df command
var dfCmd = &cobra.Command{
	Use:   "df [OPTION]... [FILE]...",
	Short: "Report file system disk space usage",
	Long: `Show information about the file system on which each FILE resides,
or all file systems by default.

  -h, --human-readable  print sizes in human readable format (e.g., 1K 234M 2G)
  -i, --inodes          list inode information instead of block usage
  -B, --block-size=SIZE scale sizes by SIZE before printing them
      --total           produce a grand total
  -t, --type=TYPE       limit listing to file systems of type TYPE
  -x, --exclude-type=TYPE  exclude file systems of type TYPE
  -l, --local           limit listing to local file systems
  -P, --portability     use the POSIX output format`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := df.DFOptions{}

		opts.HumanReadable, _ = cmd.Flags().GetBool("human-readable")
		opts.Inodes, _ = cmd.Flags().GetBool("inodes")
		opts.BlockSize, _ = cmd.Flags().GetInt64("block-size")
		opts.Total, _ = cmd.Flags().GetBool("total")
		opts.Type, _ = cmd.Flags().GetString("type")
		opts.ExcludeType, _ = cmd.Flags().GetString("exclude-type")
		opts.Local, _ = cmd.Flags().GetBool("local")
		opts.Portability, _ = cmd.Flags().GetBool("portability")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return df.RunDF(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(dfCmd)

	dfCmd.Flags().BoolP("human-readable", "H", false, "print sizes in human readable format")
	dfCmd.Flags().BoolP("inodes", "i", false, "list inode information instead of block usage")
	dfCmd.Flags().Int64P("block-size", "B", 0, "scale sizes by SIZE before printing them")
	dfCmd.Flags().Bool("total", false, "produce a grand total")
	dfCmd.Flags().StringP("type", "t", "", "limit listing to file systems of type TYPE")
	dfCmd.Flags().StringP("exclude-type", "x", "", "exclude file systems of type TYPE")
	dfCmd.Flags().BoolP("local", "l", false, "limit listing to local file systems")
	dfCmd.Flags().BoolP("portability", "P", false, "use the POSIX output format")
	dfCmd.Flags().Bool("json", false, "output as JSON")
}
