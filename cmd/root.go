package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "goshell",
	Short: "Go-native replacement for common shell utilities",
	Long: `goshell is a cross-platform, safe, Go-native replacement for common shell utilities,
designed for Taskfile, CI/CD, and enterprise environments.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}
