package cmd

import (
	"github.com/inovacc/omni/internal/cli/snowflake"
	"github.com/spf13/cobra"
)

// snowflakeCmd represents the snowflake command
var snowflakeCmd = &cobra.Command{
	Use:   "snowflake [OPTION]...",
	Short: "Generate Twitter Snowflake-style IDs",
	Long: `Generate Snowflake IDs - distributed, time-sortable unique identifiers.

Snowflake IDs are 64-bit integers with embedded timestamp:
- 1 bit: unused (sign)
- 41 bits: timestamp (milliseconds, ~69 years)
- 10 bits: worker ID (0-1023)
- 12 bits: sequence (0-4095 per ms per worker)

Features:
- Roughly time-ordered
- Distributed generation (with worker IDs)
- ~4 million IDs per second per worker

  -n, --count=N     generate N Snowflake IDs (default 1)
  -w, --worker=N    worker ID (0-1023, default 0)
  --json            output as JSON

Examples:
  omni snowflake                 # generate one Snowflake ID
  omni snowflake -n 5            # generate 5 IDs
  omni snowflake -w 42           # use worker ID 42
  omni snowflake --json          # JSON output`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := snowflake.Options{}

		opts.Count, _ = cmd.Flags().GetInt("count")
		opts.WorkerID, _ = cmd.Flags().GetInt64("worker")
		opts.JSON, _ = cmd.Flags().GetBool("json")

		return snowflake.RunSnowflake(cmd.OutOrStdout(), opts)
	},
}

func init() {
	rootCmd.AddCommand(snowflakeCmd)

	snowflakeCmd.Flags().IntP("count", "n", 1, "generate N Snowflake IDs")
	snowflakeCmd.Flags().Int64P("worker", "w", 0, "worker ID (0-1023)")
	snowflakeCmd.Flags().Bool("json", false, "output as JSON")
}
