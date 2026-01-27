package cmd

import (
	"github.com/inovacc/goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// basenameCmd represents the basename command
var basenameCmd = &cobra.Command{
	Use:   "basename NAME [SUFFIX]",
	Short: "Strip directory and suffix from file names",
	Long:  `Print NAME with any leading directory components removed. If specified, also remove a trailing SUFFIX.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.RunBasename(args)
	},
}

func init() {
	rootCmd.AddCommand(basenameCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// basenameCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// basenameCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
