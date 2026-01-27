package cmd

import (
	"os"

	"github.com/inovacc/goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// headCmd represents the head command
var headCmd = &cobra.Command{
	Use:   "head [option]... [file]...",
	Short: "Output the first part of files",
	Long: `Print the first 10 lines of each FILE to standard output.
With more than one FILE, precede each with a header giving the file name.
With no FILE, or when FILE is -, read standard input.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.HeadOptions{}

		opts.Lines, _ = cmd.Flags().GetInt("lines")
		opts.Bytes, _ = cmd.Flags().GetInt("bytes")
		opts.Quiet, _ = cmd.Flags().GetBool("quiet")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")

		return cli.RunHead(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(headCmd)

	headCmd.Flags().IntP("lines", "n", 10, "print the first NUM lines instead of the first 10")
	headCmd.Flags().IntP("bytes", "c", 0, "print the first NUM bytes of each file")
	headCmd.Flags().BoolP("quiet", "q", false, "never print headers giving file names")
	headCmd.Flags().BoolP("verbose", "v", false, "always print headers giving file names")
}
