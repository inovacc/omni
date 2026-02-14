package cmd

import (
	"github.com/inovacc/omni/internal/cli/pwd"
	"github.com/spf13/cobra"
)

// pwdCmd represents the pwd command
var pwdCmd = &cobra.Command{
	Use:   "pwd",
	Short: "Print working directory",
	Long:  `Print the full filename of the current working directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := pwd.PwdOptions{OutputFormat: getOutputOpts(cmd).GetFormat()}
		return pwd.RunPwd(cmd.OutOrStdout(), opts)
	},
}

func init() {
	rootCmd.AddCommand(pwdCmd)
}
