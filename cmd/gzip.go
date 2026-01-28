package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli"
	"github.com/spf13/cobra"
)

var (
	gzipDecompress bool
	gzipKeep       bool
	gzipForce      bool
	gzipStdout     bool
	gzipVerbose    bool
	gzipLevel      int
)

var gzipCmd = &cobra.Command{
	Use:   "gzip [OPTION]... [FILE]...",
	Short: "Compress or decompress files",
	Long: `Compress or decompress FILEs using gzip format.

  -d, --decompress   decompress
  -k, --keep         keep original files
  -f, --force        force overwrite
  -c, --stdout       write to stdout
  -v, --verbose      verbose mode
  -1 to -9           compression level (default 6)

Examples:
  omni gzip file.txt           # compress to file.txt.gz
  omni gzip -d file.txt.gz     # decompress
  omni gzip -k file.txt        # keep original
  omni gzip -c file.txt > out.gz  # write to stdout`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.GzipOptions{
			Decompress: gzipDecompress,
			Keep:       gzipKeep,
			Force:      gzipForce,
			Stdout:     gzipStdout,
			Verbose:    gzipVerbose,
			Level:      gzipLevel,
		}

		return cli.RunGzip(os.Stdout, args, opts)
	},
}

var gunzipCmd = &cobra.Command{
	Use:   "gunzip [OPTION]... [FILE]...",
	Short: "Decompress gzip files",
	Long: `Decompress FILEs in gzip format.

Equivalent to gzip -d.

Examples:
  omni gunzip file.txt.gz      # decompress
  omni gunzip -k file.txt.gz   # keep original`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.GzipOptions{
			Decompress: true,
			Keep:       gzipKeep,
			Force:      gzipForce,
			Stdout:     gzipStdout,
			Verbose:    gzipVerbose,
		}

		return cli.RunGunzip(os.Stdout, args, opts)
	},
}

var zcatCmd = &cobra.Command{
	Use:   "zcat [FILE]...",
	Short: "Decompress and print gzip files",
	Long: `Decompress and print FILEs to stdout.

Equivalent to gzip -dc.

Examples:
  omni zcat file.txt.gz        # print decompressed content
  omni zcat file.gz | grep x   # decompress and grep`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.RunZcat(os.Stdout, args)
	},
}

func init() {
	rootCmd.AddCommand(gzipCmd)
	rootCmd.AddCommand(gunzipCmd)
	rootCmd.AddCommand(zcatCmd)

	gzipCmd.Flags().BoolVarP(&gzipDecompress, "decompress", "d", false, "decompress")
	gzipCmd.Flags().BoolVarP(&gzipKeep, "keep", "k", false, "keep original files")
	gzipCmd.Flags().BoolVarP(&gzipForce, "force", "f", false, "force overwrite")
	gzipCmd.Flags().BoolVarP(&gzipStdout, "stdout", "c", false, "write to stdout")
	gzipCmd.Flags().BoolVarP(&gzipVerbose, "verbose", "v", false, "verbose mode")
	gzipCmd.Flags().IntVarP(&gzipLevel, "fast", "1", 0, "compress faster")
	gzipCmd.Flags().IntVarP(&gzipLevel, "best", "9", 0, "compress better")

	gunzipCmd.Flags().BoolVarP(&gzipKeep, "keep", "k", false, "keep original files")
	gunzipCmd.Flags().BoolVarP(&gzipForce, "force", "f", false, "force overwrite")
	gunzipCmd.Flags().BoolVarP(&gzipStdout, "stdout", "c", false, "write to stdout")
	gunzipCmd.Flags().BoolVarP(&gzipVerbose, "verbose", "v", false, "verbose mode")
}
