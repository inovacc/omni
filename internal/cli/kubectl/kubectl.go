// Package kubectl provides Kubernetes CLI functionality for omni.
// It wraps the official k8s.io/kubectl package to provide kubectl-compatible commands.
package kubectl

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/kubectl/pkg/cmd"
)

// NewKubectlCommand creates the full kubectl command tree.
// This returns a Cobra command that includes all kubectl subcommands.
func NewKubectlCommand() *cobra.Command {
	return cmd.NewDefaultKubectlCommand()
}

// NewKubectlCommandWithStreams creates kubectl with custom IO streams.
func NewKubectlCommandWithStreams(streams genericiooptions.IOStreams) *cobra.Command {
	return cmd.NewDefaultKubectlCommandWithArgs(cmd.KubectlOptions{
		IOStreams: streams,
	})
}

// DefaultIOStreams returns the default IO streams using os.Stdin/Stdout/Stderr.
func DefaultIOStreams() genericiooptions.IOStreams {
	return genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
}

// Run executes a kubectl command with the given arguments.
func Run(args []string) error {
	kubectlCmd := NewKubectlCommand()
	kubectlCmd.SetArgs(args)
	return kubectlCmd.Execute()
}

// RunWithStreams executes a kubectl command with custom IO streams.
func RunWithStreams(args []string, streams genericiooptions.IOStreams) error {
	kubectlCmd := NewKubectlCommandWithStreams(streams)
	kubectlCmd.SetArgs(args)
	return kubectlCmd.Execute()
}
