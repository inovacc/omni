package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/archive"
	"github.com/spf13/cobra"
)

// tarCmd represents the tar command
var tarCmd = &cobra.Command{
	Use:   "tar [OPTION]... [FILE]...",
	Short: "Create, extract, or list archive files",
	Long: `Manipulate tape archive files.

  -c, --create           create a new archive
  -x, --extract          extract files from an archive
  -t, --list             list the contents of an archive
  -f, --file=ARCHIVE     use archive file ARCHIVE
  -v, --verbose          verbosely list files processed
  -z, --gzip             filter through gzip
  -C, --directory=DIR    change to directory DIR
      --strip-components=N  strip N leading path components

Examples:
  omni tar -cvf archive.tar dir/        # create tar archive
  omni tar -czvf archive.tar.gz dir/    # create gzipped tar
  omni tar -xvf archive.tar             # extract tar archive
  omni tar -xzvf archive.tar.gz         # extract gzipped tar
  omni tar -tvf archive.tar             # list contents
  omni tar -xvf archive.tar -C /dest    # extract to directory`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := archive.ArchiveOptions{}

		opts.Create, _ = cmd.Flags().GetBool("create")
		opts.Extract, _ = cmd.Flags().GetBool("extract")
		opts.List, _ = cmd.Flags().GetBool("list")
		opts.File, _ = cmd.Flags().GetString("file")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		opts.Gzip, _ = cmd.Flags().GetBool("gzip")
		opts.Directory, _ = cmd.Flags().GetString("directory")
		opts.StripComponents, _ = cmd.Flags().GetInt("strip-components")

		return archive.RunTar(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(tarCmd)

	tarCmd.Flags().BoolP("create", "c", false, "create a new archive")
	tarCmd.Flags().BoolP("extract", "x", false, "extract files from an archive")
	tarCmd.Flags().BoolP("list", "t", false, "list the contents of an archive")
	tarCmd.Flags().StringP("file", "f", "", "use archive file ARCHIVE")
	tarCmd.Flags().BoolP("verbose", "v", false, "verbosely list files processed")
	tarCmd.Flags().BoolP("gzip", "z", false, "filter through gzip")
	tarCmd.Flags().StringP("directory", "C", "", "change to directory DIR")
	tarCmd.Flags().Int("strip-components", 0, "strip N leading path components")
}
