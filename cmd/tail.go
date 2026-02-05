package cmd

import (
	"os"
	"regexp"
	"time"

	"github.com/inovacc/omni/internal/cli/tail"
	"github.com/spf13/cobra"
)

// tailNumericFlagRegex matches -NUM style arguments (e.g., -80, -100).
var tailNumericFlagRegex = regexp.MustCompile(`^-(\d+)$`)

// tailCmd represents the tail command
var tailCmd = &cobra.Command{
	Use:   "tail [option]... [file]...",
	Short: "Output the last part of files",
	Long: `Print the last 10 lines of each FILE to standard output.
With more than one FILE, precede each with a header giving the file name.
With no FILE, or when FILE is -, read standard input.

Numeric shortcuts are supported: -80 is equivalent to -n 80.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := tail.TailOptions{}

		opts.Lines, _ = cmd.Flags().GetInt("lines")
		opts.Bytes, _ = cmd.Flags().GetInt("bytes")
		opts.Follow, _ = cmd.Flags().GetBool("follow")
		opts.Quiet, _ = cmd.Flags().GetBool("quiet")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")
		opts.Sleep, _ = cmd.Flags().GetDuration("sleep-interval")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return tail.RunTail(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(tailCmd)

	tailCmd.Flags().IntP("lines", "n", 10, "output the last NUM lines, instead of the last 10")
	tailCmd.Flags().IntP("bytes", "c", 0, "output the last NUM bytes")
	tailCmd.Flags().BoolP("follow", "f", false, "output appended data as the file grows")
	tailCmd.Flags().BoolP("quiet", "q", false, "never output headers giving file names")
	tailCmd.Flags().BoolP("verbose", "v", false, "always output headers giving file names")
	tailCmd.Flags().Duration("sleep-interval", time.Second, "with -f, sleep for approximately N seconds between iterations")
	tailCmd.Flags().Bool("json", false, "output as JSON")

	// Preprocess os.Args to convert -NUM to -n NUM for tail command
	preprocessTailArgs()
}

// preprocessTailArgs converts -NUM style arguments to -n NUM before Cobra parses them.
func preprocessTailArgs() {
	if len(os.Args) < 2 {
		return
	}

	// Check if this is a tail command
	isTailCmd := false
	for i, arg := range os.Args {
		if arg == "tail" && i > 0 {
			isTailCmd = true
			break
		}
	}

	if !isTailCmd {
		return
	}

	// Rewrite -NUM to -n NUM
	newArgs := make([]string, 0, len(os.Args)+1)

	for _, arg := range os.Args {
		if matches := tailNumericFlagRegex.FindStringSubmatch(arg); matches != nil {
			newArgs = append(newArgs, "-n", matches[1])
		} else {
			newArgs = append(newArgs, arg)
		}
	}

	os.Args = newArgs
}
