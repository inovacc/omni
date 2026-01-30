package cmd

import (
	"github.com/inovacc/omni/internal/cli/ps"
	"github.com/spf13/cobra"
)

// topCmd represents the top command
var topCmd = &cobra.Command{
	Use:   "top",
	Short: "Display system processes sorted by resource usage",
	Long: `Display system processes sorted by CPU or memory usage.

  -n NUM         number of processes to show (default 10)
  --sort COL     sort by column: cpu (default), mem, pid
  -j, --json     output as JSON
  --go           show only Go processes

Note: This is a snapshot view. For real-time monitoring, use system top.

Examples:
  omni top                 # show top 10 by CPU
  omni top -n 20           # show top 20 processes
  omni top --sort mem      # sort by memory
  omni top --go            # show top Go processes
  omni top -j              # output as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error {
		n, _ := cmd.Flags().GetInt("num")
		sortBy, _ := cmd.Flags().GetString("sort")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		goOnly, _ := cmd.Flags().GetBool("go")

		opts := ps.Options{
			All:    true,
			JSON:   jsonOutput,
			GoOnly: goOnly,
			Sort:   sortBy,
		}

		if opts.Sort == "" {
			opts.Sort = "cpu"
		}

		return ps.RunTop(cmd.OutOrStdout(), opts, n)
	},
}

func init() {
	rootCmd.AddCommand(topCmd)

	topCmd.Flags().IntP("num", "n", 10, "number of processes to show")
	topCmd.Flags().String("sort", "cpu", "sort by column: cpu, mem, pid")
	topCmd.Flags().BoolP("json", "j", false, "output as JSON")
	topCmd.Flags().Bool("go", false, "show only Go processes")
}
