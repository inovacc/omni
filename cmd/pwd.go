/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"goshell/pkg/cli"

	"github.com/spf13/cobra"
)

// pwdCmd represents the pwd command
var pwdCmd = &cobra.Command{
	Use:   "pwd",
	Short: "Print working directory",
	Long:  `Print the full filename of the current working directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.RunPwd()
	},
}

func init() {
	rootCmd.AddCommand(pwdCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pwdCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pwdCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
