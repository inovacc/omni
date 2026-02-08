package cmd

import (
	"github.com/inovacc/omni/internal/cli/pipeline"
	"github.com/spf13/cobra"
)

var pipelineCmd = &cobra.Command{
	Use:   "pipeline STAGE [STAGE...]",
	Short: "Streaming text processing engine",
	Long: `Streaming text processing engine with built-in transform stages.

Each stage is a quoted string describing a transform. Stages are connected
via io.Pipe goroutines for memory-efficient, line-by-line processing.

Available stages:
  grep PATTERN       Filter lines matching regex pattern (-i, -v)
  grep-v PATTERN     Filter lines NOT matching pattern
  contains SUBSTR    Filter lines containing literal substring (-i)
  replace OLD NEW    Replace all occurrences of OLD with NEW
  head [N]           Output first N lines (default 10)
  take [N]           Alias for head
  tail [N]           Output last N lines (default 10)
  skip N             Skip first N lines
  sort               Sort lines (-r reverse, -n numeric, -rn both)
  uniq               Remove consecutive duplicate lines (-i)
  cut -dDELIM -fN    Extract fields (-d delimiter, -f fields)
  tr FROM TO         Translate characters
  sed s/PAT/REPL/g   Regex substitution
  rev                Reverse each line
  nl                 Number each line
  tee FILE           Copy output to file and next stage
  tac                Reverse line order
  wc                 Count lines/words/chars (-l, -w, -c)

Examples:
  omni pipeline 'grep error' 'sort' 'uniq' 'head 10' < log.txt
  omni pipeline -f access.log 'grep 404' 'cut -d" " -f1' 'sort' 'uniq'
  omni pipeline -v 'grep -i warning' 'sort -rn' 'head 5'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := pipeline.Options{}
		opts.File, _ = cmd.Flags().GetString("file")
		opts.Verbose, _ = cmd.Flags().GetBool("verbose")

		return pipeline.Run(cmd.OutOrStdout(), cmd.InOrStdin(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(pipelineCmd)

	pipelineCmd.Flags().StringP("file", "f", "", "input file (default: stdin)")
	pipelineCmd.Flags().BoolP("verbose", "v", false, "show stage names before processing")
}
