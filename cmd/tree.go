package cmd

import (
	"strings"

	"github.com/inovacc/omni/internal/cli/tree"
	"github.com/spf13/cobra"
)

var treeCmd = &cobra.Command{
	Use:   "tree [path]",
	Short: "Display directory tree structure",
	Long: `Display a tree visualization of directory contents.

Examples:
  omni tree                          # current directory
  omni tree /path/to/dir             # specific directory
  omni tree -a                       # show hidden files
  omni tree -d 3                     # limit depth to 3
  omni tree -i "node_modules,.git"   # ignore patterns
  omni tree --dirs-only              # show only directories
  omni tree -s                       # show statistics
  omni tree --json                   # output as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := tree.TreeOptions{}

		opts.All, _ = cmd.Flags().GetBool("all")
		opts.DirsOnly, _ = cmd.Flags().GetBool("dirs-only")
		opts.Depth, _ = cmd.Flags().GetInt("depth")
		opts.NoDirSlash, _ = cmd.Flags().GetBool("no-dir-slash")
		opts.Stats, _ = cmd.Flags().GetBool("stats")
		opts.Hash, _ = cmd.Flags().GetBool("hash")
		opts.JSON, _ = cmd.Flags().GetBool("json")
		opts.NoColor, _ = cmd.Flags().GetBool("no-color")
		opts.Size, _ = cmd.Flags().GetBool("size")
		opts.Date, _ = cmd.Flags().GetBool("date")

		ignoreStr, _ := cmd.Flags().GetString("ignore")
		if ignoreStr != "" {
			opts.Ignore = strings.Split(ignoreStr, ",")
		}

		return tree.RunTree(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(treeCmd)

	treeCmd.Flags().BoolP("all", "a", false, "show hidden files")
	treeCmd.Flags().Bool("dirs-only", false, "show only directories")
	treeCmd.Flags().IntP("depth", "d", -1, "maximum depth to scan (-1 for unlimited)")
	treeCmd.Flags().StringP("ignore", "i", "", "patterns to ignore (comma-separated)")
	treeCmd.Flags().Bool("no-dir-slash", false, "don't add trailing slash to directory names")
	treeCmd.Flags().BoolP("stats", "s", false, "show statistics")
	treeCmd.Flags().Bool("hash", false, "show SHA256 hash for files")
	treeCmd.Flags().BoolP("json", "j", false, "output as JSON format")
	treeCmd.Flags().Bool("no-color", false, "disable colored output")
	treeCmd.Flags().Bool("size", false, "show file sizes")
	treeCmd.Flags().Bool("date", false, "show modification dates")
}
