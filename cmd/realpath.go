package cmd

import (
	"github.com/inovacc/goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// realpathCmd represents the realpath command
var realpathCmd = &cobra.Command{
	Use:   "realpath [path...]",
	Short: "Print the resolved path",
	Long:  `Print the resolved absolute file name; all components must exist.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.RunRealpath(args)
	},
}

func init() {
	rootCmd.AddCommand(realpathCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// realpathCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// realpathCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
