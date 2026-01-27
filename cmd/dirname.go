package cmd

import (
	"github.com/inovacc/goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// dirnameCmd represents the dirname command
var dirnameCmd = &cobra.Command{
	Use:   "dirname [path...]",
	Short: "Strip last component from file name",
	Long:  `Output each NAME with its last non-slash component and trailing slashes removed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.RunDirname(args)
	},
}

func init() {
	rootCmd.AddCommand(dirnameCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dirnameCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dirnameCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
