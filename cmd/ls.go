package cmd

import (
	"github.com/inovacc/goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// lsCmd represents the ls command
var lsCmd = &cobra.Command{
	Use:   "ls [directory]",
	Short: "List directory contents",
	Long:  `List information about the FILEs (the current directory by default).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.RunLs(args)
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// lsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// lsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
