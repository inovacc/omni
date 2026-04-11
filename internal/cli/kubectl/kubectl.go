// Package kubectl provides Kubernetes CLI functionality for omni.
// It wraps the official k8s.io/kubectl package to provide kubectl-compatible commands.
package kubectl

import (
	"os"

	"github.com/inovacc/omni/internal/cli/cmderr"
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

	if err := kubectlCmd.Execute(); err != nil {
		// Pitfall 3: avoid double-wrapping if already a cmderr sentinel.
		// kubectl's cobra errors are plain errors; wrap at the boundary.
		return cmderr.Wrap(cmderr.ErrIO, "kubectl: "+err.Error())
	}

	return nil
}

// RunWithStreams executes a kubectl command with custom IO streams.
func RunWithStreams(args []string, streams genericiooptions.IOStreams) error {
	kubectlCmd := NewKubectlCommandWithStreams(streams)
	kubectlCmd.SetArgs(args)

	if err := kubectlCmd.Execute(); err != nil {
		return cmderr.Wrap(cmderr.ErrIO, "kubectl: "+err.Error())
	}

	return nil
}
