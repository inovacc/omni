package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli/free"

	"github.com/spf13/cobra"
)

// freeCmd represents the free command
var freeCmd = &cobra.Command{
	Use:   "free [OPTION]...",
	Short: "Display amount of free and used memory in the system",
	Long: `Display the total amount of free and used physical and swap memory
in the system, as well as the buffers and caches used by the kernel.

  -b, --bytes         show output in bytes
  -k, --kibibytes     show output in kibibytes (default)
  -m, --mebibytes     show output in mebibytes
  -g, --gibibytes     show output in gibibytes
  -h, --human         show human-readable output
  -w, --wide          wide output
  -t, --total         show total for RAM + swap`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := free.FreeOptions{}

		opts.Bytes, _ = cmd.Flags().GetBool("bytes")
		opts.Kibibytes, _ = cmd.Flags().GetBool("kibibytes")
		opts.Mebibytes, _ = cmd.Flags().GetBool("mebibytes")
		opts.Gibibytes, _ = cmd.Flags().GetBool("gibibytes")
		opts.Human, _ = cmd.Flags().GetBool("human")
		opts.Wide, _ = cmd.Flags().GetBool("wide")
		opts.Total, _ = cmd.Flags().GetBool("total")

		return free.RunFree(os.Stdout, opts)
	},
}

func init() {
	rootCmd.AddCommand(freeCmd)

	freeCmd.Flags().BoolP("bytes", "b", false, "show output in bytes")
	freeCmd.Flags().BoolP("kibibytes", "k", false, "show output in kibibytes")
	freeCmd.Flags().BoolP("mebibytes", "m", false, "show output in mebibytes")
	freeCmd.Flags().BoolP("gibibytes", "g", false, "show output in gibibytes")
	freeCmd.Flags().BoolP("human", "H", false, "show human-readable output")
	freeCmd.Flags().BoolP("wide", "w", false, "wide output")
	freeCmd.Flags().BoolP("total", "t", false, "show total for RAM + swap")
}
