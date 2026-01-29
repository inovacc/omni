package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/archive"
	"github.com/spf13/cobra"
)

// zipCmd represents the zip command
var zipCmd = &cobra.Command{
	Use:   "zip [OPTION]... ZIPFILE FILE...",
	Short: "Package and compress files into a zip archive",
	Long: `Create a zip archive from files and directories.

  -v, --verbose     verbose output
  -r, --recursive   recurse into directories (default for directories)
  -C, --directory   change to directory before adding files

Examples:
  omni zip archive.zip file1.txt file2.txt   # create zip
  omni zip archive.zip dir/                   # zip directory
  omni zip -v archive.zip file.txt           # verbose output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Help()
		}

		opts := archive.ArchiveOptions{
			Create: true,
			File:   args[0],
		}

		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		opts.Directory, _ = cmd.Flags().GetString("directory")

		return archive.RunZip(os.Stdout, args[1:], opts)
	},
}

func init() {
	rootCmd.AddCommand(zipCmd)

	zipCmd.Flags().BoolP("verbose", "v", false, "verbose output")
	zipCmd.Flags().BoolP("recursive", "r", false, "recurse into directories")
	zipCmd.Flags().StringP("directory", "C", "", "change to directory before adding")
}
