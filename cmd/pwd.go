package cmd

import (
	"os"

	"github.com/inovacc/omni/pkg/cli"

	"github.com/spf13/cobra"
)

// pwdCmd represents the pwd command
var pwdCmd = &cobra.Command{
	Use:   "pwd",
	Short: "Print working directory",
	Long:  `Print the full filename of the current working directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cli.RunPwd(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(pwdCmd)
}
