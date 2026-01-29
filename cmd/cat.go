package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/cat"
	"github.com/spf13/cobra"
)

// catCmd represents the cat command
var catCmd = &cobra.Command{
	Use:   "cat [file...]",
	Short: "Concatenate files and print on the standard output",
	Long: `Concatenate FILE(s) to standard output.
With no FILE, or when FILE is -, read standard input.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := cat.CatOptions{}

		opts.NumberAll, _ = cmd.Flags().GetBool("number")
		opts.NumberNonBlank, _ = cmd.Flags().GetBool("number-nonblank")
		opts.ShowEnds, _ = cmd.Flags().GetBool("show-ends")
		opts.ShowTabs, _ = cmd.Flags().GetBool("show-tabs")
		opts.SqueezeBlank, _ = cmd.Flags().GetBool("squeeze-blank")
		opts.ShowNonPrint, _ = cmd.Flags().GetBool("show-nonprinting")

		// -A is equivalent to -vET
		if showAll, _ := cmd.Flags().GetBool("show-all"); showAll {
			opts.ShowNonPrint = true
			opts.ShowEnds = true
			opts.ShowTabs = true
		}

		// -e is equivalent to -vE
		if e, _ := cmd.Flags().GetBool("e"); e {
			opts.ShowNonPrint = true
			opts.ShowEnds = true
		}

		// -t is equivalent to -vT
		if t, _ := cmd.Flags().GetBool("t"); t {
			opts.ShowNonPrint = true
			opts.ShowTabs = true
		}

		opts.JSON, _ = cmd.Flags().GetBool("json")

		return cat.RunCat(os.Stdout, args, opts)
	},
}

func init() {
	rootCmd.AddCommand(catCmd)

	catCmd.Flags().BoolP("number", "n", false, "number all output lines")
	catCmd.Flags().BoolP("number-nonblank", "b", false, "number nonempty output lines, overrides -n")
	catCmd.Flags().BoolP("show-ends", "E", false, "display $ at end of each line")
	catCmd.Flags().BoolP("show-tabs", "T", false, "display TAB characters as ^I")
	catCmd.Flags().BoolP("squeeze-blank", "s", false, "suppress repeated empty output lines")
	catCmd.Flags().BoolP("show-nonprinting", "v", false, "use ^ and M- notation, except for LFD and TAB")
	catCmd.Flags().BoolP("show-all", "A", false, "equivalent to -vET")
	catCmd.Flags().BoolP("e", "e", false, "equivalent to -vE")
	catCmd.Flags().BoolP("t", "t", false, "equivalent to -vT")
	catCmd.Flags().Bool("json", false, "output as JSON array of lines")
}
