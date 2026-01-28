package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli/file"
	"github.com/spf13/cobra"
)

var (
	fileBrief     bool
	fileMimeType  bool
	fileNoDeref   bool
	fileSeparator string
)

var fileCmd = &cobra.Command{
	Use:   "file [OPTION]... FILE...",
	Short: "Determine file type",
	Long: `Determine the type of each FILE.

  -b, --brief           do not prepend filenames to output
  -i, --mime            output MIME type strings
  -h, --no-dereference  don't follow symlinks
  -F, --separator       use string as separator instead of ':'

Examples:
  omni file image.png          # PNG image data
  omni file -i document.pdf    # application/pdf
  omni file -b script.sh       # output type only
  omni file *                  # check all files`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := file.FileOptions{
			Brief:     fileBrief,
			MimeType:  fileMimeType,
			NoDeref:   fileNoDeref,
			Separator: fileSeparator,
		}

		return file.RunFile(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(fileCmd)

	fileCmd.Flags().BoolVarP(&fileBrief, "brief", "b", false, "do not prepend filenames")
	fileCmd.Flags().BoolVarP(&fileMimeType, "mime", "i", false, "output MIME type")
	fileCmd.Flags().BoolVarP(&fileNoDeref, "no-dereference", "L", false, "don't follow symlinks")
	fileCmd.Flags().StringVarP(&fileSeparator, "separator", "F", ":", "use string as separator")
}
