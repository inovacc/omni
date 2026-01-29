package cmd

import (
	"github.com/inovacc/omni/internal/cli/cron"
	"github.com/spf13/cobra"
)

var cronCmd = &cobra.Command{
	Use:   "cron EXPRESSION",
	Short: "Parse and explain cron expressions",
	Long: `Parse cron expressions and display human-readable explanations.

Cron expression format: MINUTE HOUR DAY MONTH WEEKDAY

Field values:
  MINUTE   0-59
  HOUR     0-23
  DAY      1-31
  MONTH    1-12 (or names: jan, feb, etc.)
  WEEKDAY  0-7 (0 and 7 = Sunday, or names: sun, mon, etc.)

Special characters:
  *    Any value
  ,    List separator (e.g., 1,3,5)
  -    Range (e.g., 1-5)
  /    Step (e.g., */5)

Aliases:
  @yearly    Once a year (0 0 1 1 *)
  @monthly   Once a month (0 0 1 * *)
  @weekly    Once a week (0 0 * * 0)
  @daily     Once a day (0 0 * * *)
  @hourly    Once an hour (0 * * * *)

Examples:
  omni cron "*/15 * * * *"              # Every 15 minutes
  omni cron "0 9 * * 1-5"               # 9 AM on weekdays
  omni cron "0 0 1 * *"                 # First day of month at midnight
  omni cron "30 4 1,15 * *"             # 4:30 AM on 1st and 15th
  omni cron "@daily"                    # Every day at midnight
  omni cron --next 5 "0 */2 * * *"      # Show next 5 runs`,
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOutput, _ := cmd.Flags().GetBool("json")
		next, _ := cmd.Flags().GetInt("next")
		validate, _ := cmd.Flags().GetBool("validate")

		opts := cron.Options{
			JSON:     jsonOutput,
			Next:     next,
			Validate: validate,
		}

		return cron.Run(cmd.OutOrStdout(), args, opts)
	},
}

func init() {
	rootCmd.AddCommand(cronCmd)

	cronCmd.Flags().Bool("json", false, "output as JSON")
	cronCmd.Flags().Int("next", 0, "show next N scheduled runs")
	cronCmd.Flags().Bool("validate", false, "only validate the expression")
}
