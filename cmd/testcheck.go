package cmd

import (
	"os"

	"github.com/inovacc/omni/internal/cli/testcheck"
	"github.com/spf13/cobra"
)

var testcheckOpts testcheck.Options

// testcheckCmd represents the testcheck command
var testcheckCmd = &cobra.Command{
	Use:   "testcheck [directory]",
	Short: "Check test coverage for Go packages",
	Long: `Scan a directory for Go packages and report which have tests.

By default, only shows packages WITHOUT tests. Use --all to show all packages.

Examples:
  omni testcheck ./pkg/cli/           # Check packages in pkg/cli
  omni testcheck .                    # Check current directory
  omni testcheck --all ./pkg/         # Show all packages
  omni testcheck --summary ./pkg/     # Show only summary
  omni testcheck --json ./pkg/        # Output as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := "."
		if len(args) > 0 {
			dir = args[0]
		}

		return testcheck.Run(os.Stdout, dir, testcheckOpts)
	},
}

func init() {
	rootCmd.AddCommand(testcheckCmd)

	testcheckCmd.Flags().BoolVarP(&testcheckOpts.JSON, "json", "j", false, "output as JSON")
	testcheckCmd.Flags().BoolVarP(&testcheckOpts.ShowAll, "all", "a", false, "show all packages (default shows only missing)")
	testcheckCmd.Flags().BoolVarP(&testcheckOpts.Summary, "summary", "s", false, "show only summary")
	testcheckCmd.Flags().BoolVarP(&testcheckOpts.Verbose, "verbose", "v", false, "show test file names")
}
