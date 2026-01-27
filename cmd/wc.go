package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli"

	"github.com/spf13/cobra"
)

// wcCmd represents the wc command
var wcCmd = &cobra.Command{
	Use:   "wc [option]... [file]...",
	Short: "Print newline, word, and byte counts for each file",
	Long: `Print newline, word, and byte counts for each FILE, and a total line if
more than one FILE is specified. A word is a non-zero-length sequence of
characters delimited by white space.

With no FILE, or when FILE is -, read standard input.

The options below may be used to select which counts are printed, always in
the following order: newline, word, character, byte, maximum line length.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cli.WCOptions{}

		opts.Lines, _ = cmd.Flags().GetBool("lines")
		opts.Words, _ = cmd.Flags().GetBool("words")
		opts.Bytes, _ = cmd.Flags().GetBool("bytes")
		opts.Chars, _ = cmd.Flags().GetBool("chars")
		opts.MaxLineLen, _ = cmd.Flags().GetBool("max-line-length")

		return cli.RunWC(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(wcCmd)

	wcCmd.Flags().BoolP("lines", "l", false, "print the newline counts")
	wcCmd.Flags().BoolP("words", "w", false, "print the word counts")
	wcCmd.Flags().BoolP("bytes", "c", false, "print the byte counts")
	wcCmd.Flags().BoolP("chars", "m", false, "print the character counts")
	wcCmd.Flags().BoolP("max-line-length", "L", false, "print the maximum display width")
}
