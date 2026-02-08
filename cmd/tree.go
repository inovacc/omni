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
  omni tree -L 3                     # limit depth to 3
  omni tree -d 3                     # limit depth to 3 (alias for -L)
  omni tree -i "node_modules,.git"   # ignore patterns
  omni tree --dirs-only              # show only directories
  omni tree -s                       # show statistics
  omni tree --json                   # output as JSON
  omni tree --json-stream            # streaming NDJSON output
  omni tree -t 8                     # use 8 parallel workers
  omni tree --max-files 10000        # cap at 10000 items
  omni tree --compare a.json b.json  # compare two snapshots`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := tree.TreeOptions{}

		opts.All, _ = cmd.Flags().GetBool("all")
		opts.DirsOnly, _ = cmd.Flags().GetBool("dirs-only")
		opts.Depth, _ = cmd.Flags().GetInt("depth")

		// -L overrides --depth/-d if explicitly set
		if cmd.Flags().Changed("level") {
			opts.Depth, _ = cmd.Flags().GetInt("level")
		}
		opts.NoDirSlash, _ = cmd.Flags().GetBool("no-dir-slash")
		opts.Stats, _ = cmd.Flags().GetBool("stats")
		opts.Hash, _ = cmd.Flags().GetBool("hash")
		opts.JSON, _ = cmd.Flags().GetBool("json")
		opts.JSONStream, _ = cmd.Flags().GetBool("json-stream")
		opts.NoColor, _ = cmd.Flags().GetBool("no-color")
		opts.Size, _ = cmd.Flags().GetBool("size")
		opts.Date, _ = cmd.Flags().GetBool("date")
		opts.MaxFiles, _ = cmd.Flags().GetInt("max-files")
		opts.MaxHashSize, _ = cmd.Flags().GetInt64("max-hash-size")
		opts.Threads, _ = cmd.Flags().GetInt("threads")
		opts.DetectMoves, _ = cmd.Flags().GetBool("detect-moves")

		compareFiles, _ := cmd.Flags().GetStringSlice("compare")
		if len(compareFiles) == 2 {
			opts.Compare = compareFiles
		}

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
	treeCmd.Flags().IntP("level", "L", -1, "maximum depth level (alias for --depth)")
	treeCmd.Flags().StringP("ignore", "i", "", "patterns to ignore (comma-separated)")
	treeCmd.Flags().Bool("no-dir-slash", false, "don't add trailing slash to directory names")
	treeCmd.Flags().BoolP("stats", "s", false, "show statistics")
	treeCmd.Flags().Bool("hash", false, "show SHA256 hash for files")
	treeCmd.Flags().BoolP("json", "j", false, "output as JSON format")
	treeCmd.Flags().Bool("json-stream", false, "streaming NDJSON output (one JSON object per line)")
	treeCmd.Flags().Bool("no-color", false, "disable colored output")
	treeCmd.Flags().Bool("size", false, "show file sizes")
	treeCmd.Flags().Bool("date", false, "show modification dates")
	treeCmd.Flags().Int("max-files", 0, "maximum number of files to scan (0 = unlimited)")
	treeCmd.Flags().Int64("max-hash-size", 0, "skip hashing files larger than N bytes (0 = unlimited)")
	treeCmd.Flags().IntP("threads", "t", 0, "number of parallel workers (0 = auto, 1 = sequential)")
	treeCmd.Flags().StringSlice("compare", nil, "compare two JSON tree snapshots")
	treeCmd.Flags().Bool("detect-moves", true, "detect moved files when comparing (default true)")
}
