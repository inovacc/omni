package cmd

import (
	"github.com/inovacc/omni/internal/cli/kubectl"
	"github.com/spf13/cobra"
)

var kubectlCmd = &cobra.Command{
	Use:                "kubectl",
	Aliases:            []string{"k"},
	Short:              "Kubernetes CLI",
	Long: `Kubernetes command-line tool integrated into omni.

This is a full integration of kubectl, supporting all kubectl commands and flags.
You can use 'omni kubectl' or the shorter alias 'omni k'.

Examples:
  omni kubectl get pods
  omni k get pods -A
  omni k describe node mynode
  omni k logs -f mypod
  omni k exec -it mypod -- /bin/sh
  omni k apply -f manifest.yaml`,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return kubectl.Run(args)
	},
}

func init() {
	rootCmd.AddCommand(kubectlCmd)
}
