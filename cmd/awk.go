package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// awkCmd represents the awk command
var awkCmd = &cobra.Command{
	Use:   "awk",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("awk called")
	},
}

func init() {
	rootCmd.AddCommand(awkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// awkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// awkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
