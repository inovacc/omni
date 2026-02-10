package cmd

import (
	pathutil "github.com/inovacc/omni/internal/cli/path"
	"github.com/spf13/cobra"
)

// pathCmd represents the path parent command
var pathCmd = &cobra.Command{
	Use:   "path",
	Short: "Path manipulation utilities",
	Long:  `Clean, resolve, and manipulate file paths using OS-native conventions.`,
}

// pathCleanCmd represents the path clean subcommand
var pathCleanCmd = &cobra.Command{
	Use:   "clean [path...]",
	Short: "Return the shortest equivalent path with OS separators",
	Long:  `Clean returns the shortest path name equivalent to path by purely lexical processing. It applies OS-native separators.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := pathutil.CleanOptions{}
		opts.JSON, _ = cmd.Flags().GetBool("json")
		return pathutil.RunClean(cmd.OutOrStdout(), args, opts)
	},
}

// pathAbsCmd represents the path abs subcommand
var pathAbsCmd = &cobra.Command{
	Use:   "abs [path...]",
	Short: "Return the absolute path",
	Long:  `Abs returns an absolute representation of path. Relative paths are resolved from the current working directory.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := pathutil.AbsOptions{}
		opts.JSON, _ = cmd.Flags().GetBool("json")
		return pathutil.RunAbs(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(pathCmd)

	pathCmd.AddCommand(pathCleanCmd)
	pathCleanCmd.Flags().Bool("json", false, "output as JSON")

	pathCmd.AddCommand(pathAbsCmd)
	pathAbsCmd.Flags().Bool("json", false, "output as JSON")
}
