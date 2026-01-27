package cmd

import (
	"os"

	"github.com/inovacc/goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// unzipCmd represents the unzip command
var unzipCmd = &cobra.Command{
	Use:   "unzip [OPTION]... ZIPFILE",
	Short: "Extract files from a zip archive",
	Long: `Extract files from a zip archive.

  -l, --list        list contents without extracting
  -v, --verbose     verbose output
  -d, --directory   extract files into directory
      --strip-components=N  strip N leading path components

Examples:
  goshell unzip archive.zip              # extract to current directory
  goshell unzip -d /dest archive.zip     # extract to specific directory
  goshell unzip -l archive.zip           # list contents
  goshell unzip -v archive.zip           # verbose extraction`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Help()
		}

		opts := cli.ArchiveOptions{
			File: args[0],
		}

		opts.List, _ = cmd.Flags().GetBool("list")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		opts.Directory, _ = cmd.Flags().GetString("directory")
		opts.StripComponents, _ = cmd.Flags().GetInt("strip-components")

		if opts.List {
			return cli.RunArchive(os.Stdout, nil, opts)
		}

		opts.Extract = true
		return cli.RunUnzip(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(unzipCmd)

	unzipCmd.Flags().BoolP("list", "l", false, "list contents without extracting")
	unzipCmd.Flags().BoolP("verbose", "v", false, "verbose output")
	unzipCmd.Flags().StringP("directory", "d", "", "extract files into directory")
	unzipCmd.Flags().Int("strip-components", 0, "strip N leading path components")
}
