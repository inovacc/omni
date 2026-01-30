package cmd

import (
	"github.com/inovacc/omni/internal/cli/ls"
	"github.com/spf13/cobra"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls [file...]",
	Short: "List directory contents",
	Long: `List information about the FILEs (the current directory by default).
Sort entries alphabetically if none of -tSU is specified.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := ls.Options{}

		opts.All, _ = cmd.Flags().GetBool("all")
		opts.AlmostAll, _ = cmd.Flags().GetBool("almost-all")
		opts.LongFormat, _ = cmd.Flags().GetBool("long")
		opts.HumanReadble, _ = cmd.Flags().GetBool("human-readable")
		opts.OnePerLine, _ = cmd.Flags().GetBool("one")
		opts.Recursive, _ = cmd.Flags().GetBool("recursive")
		opts.Reverse, _ = cmd.Flags().GetBool("reverse")
		opts.SortByTime, _ = cmd.Flags().GetBool("time")
		opts.SortBySize, _ = cmd.Flags().GetBool("size")
		opts.NoSort, _ = cmd.Flags().GetBool("no-sort")
		opts.Directory, _ = cmd.Flags().GetBool("directory")
		opts.Classify, _ = cmd.Flags().GetBool("classify")
		opts.Inode, _ = cmd.Flags().GetBool("inode")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return ls.Run(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)

	lsCmd.Flags().BoolP("all", "a", false, "do not ignore entries starting with .")
	lsCmd.Flags().BoolP("almost-all", "A", false, "do not list implied . and ..")
	lsCmd.Flags().BoolP("long", "l", false, "use a long listing format")
	lsCmd.Flags().BoolP("human-readable", "H", false, "with -l, print sizes in human readable format")
	lsCmd.Flags().BoolP("one", "1", false, "list one file per line")
	lsCmd.Flags().BoolP("recursive", "R", false, "list subdirectories recursively")
	lsCmd.Flags().BoolP("reverse", "r", false, "reverse order while sorting")
	lsCmd.Flags().BoolP("time", "t", false, "sort by modification time, newest first")
	lsCmd.Flags().BoolP("size", "S", false, "sort by file size, largest first")
	lsCmd.Flags().BoolP("no-sort", "U", false, "do not sort; list entries in directory order")
	lsCmd.Flags().BoolP("directory", "d", false, "list directories themselves, not their contents")
	lsCmd.Flags().BoolP("classify", "F", false, "append indicator (*/=>@|) to entries")
	lsCmd.Flags().BoolP("inode", "i", false, "print the index number of each file")
	lsCmd.Flags().Bool("json", false, "output in JSON format")
}
