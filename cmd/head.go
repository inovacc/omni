package cmd

import (
	"os"
	"regexp"

	"github.com/inovacc/omni/internal/cli/head"
	"github.com/spf13/cobra"
)

// numericFlagRegex matches -NUM style arguments (e.g., -80, -100).
var numericFlagRegex = regexp.MustCompile(`^-(\d+)$`)

// headCmd represents the head command
var headCmd = &cobra.Command{
	Use:   "head [option]... [file]...",
	Short: "Output the first part of files",
	Long: `Print the first 10 lines of each FILE to standard output.
With more than one FILE, precede each with a header giving the file name.
With no FILE, or when FILE is -, read standard input.

Numeric shortcuts are supported: -80 is equivalent to -n 80.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := head.HeadOptions{}

		opts.Lines, _ = cmd.Flags().GetInt("lines")
		opts.Bytes, _ = cmd.Flags().GetInt("bytes")
		opts.Quiet, _ = cmd.Flags().GetBool("quiet")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return head.RunHead(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(headCmd)

	headCmd.Flags().IntP("lines", "n", 10, "print the first NUM lines instead of the first 10")
	headCmd.Flags().IntP("bytes", "c", 0, "print the first NUM bytes of each file")
	headCmd.Flags().BoolP("quiet", "q", false, "never print headers giving file names")
	headCmd.Flags().BoolP("verbose", "v", false, "always print headers giving file names")
	headCmd.Flags().Bool("json", false, "output as JSON")

	// Preprocess os.Args to convert -NUM to -n NUM for head command
	preprocessHeadArgs()
}

// preprocessHeadArgs converts -NUM style arguments to -n NUM before Cobra parses them.
func preprocessHeadArgs() {
	if len(os.Args) < 2 {
		return
	}

	// Check if this is a head command
	isHeadCmd := false
	for i, arg := range os.Args {
		if arg == "head" && i > 0 {
			isHeadCmd = true
			break
		}
	}

	if !isHeadCmd {
		return
	}

	// Rewrite -NUM to -n NUM
	newArgs := make([]string, 0, len(os.Args)+1)

	for _, arg := range os.Args {
		if matches := numericFlagRegex.FindStringSubmatch(arg); matches != nil {
			newArgs = append(newArgs, "-n", matches[1])
		} else {
			newArgs = append(newArgs, arg)
		}
	}

	os.Args = newArgs
}
