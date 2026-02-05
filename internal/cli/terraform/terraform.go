// Package terraform provides Terraform CLI wrapper for omni.
// It wraps the terraform binary as a subprocess for now.
// Future: Direct integration via local source code.
package terraform

import (
	"fmt"
	"os"
	"os/exec"
)

// Run executes a terraform command with the given arguments.
func Run(args []string) error {
	cmd := exec.Command("terraform", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// Init initializes a Terraform working directory.
func Init(upgrade, reconfigure bool) error {
	args := []string{"init"}
	if upgrade {
		args = append(args, "-upgrade")
	}

	if reconfigure {
		args = append(args, "-reconfigure")
	}

	return Run(args)
}

// Plan creates an execution plan.
func Plan(out string, vars map[string]string, varFiles []string, destroy bool) error {
	args := []string{"plan"}
	if out != "" {
		args = append(args, "-out="+out)
	}

	if destroy {
		args = append(args, "-destroy")
	}

	for k, v := range vars {
		args = append(args, "-var", fmt.Sprintf("%s=%s", k, v))
	}

	for _, f := range varFiles {
		args = append(args, "-var-file="+f)
	}

	return Run(args)
}

// Apply applies the changes.
func Apply(planFile string, autoApprove bool, vars map[string]string, varFiles []string) error {
	args := []string{"apply"}
	if autoApprove {
		args = append(args, "-auto-approve")
	}

	for k, v := range vars {
		args = append(args, "-var", fmt.Sprintf("%s=%s", k, v))
	}

	for _, f := range varFiles {
		args = append(args, "-var-file="+f)
	}

	if planFile != "" {
		args = append(args, planFile)
	}

	return Run(args)
}

// Destroy destroys the managed infrastructure.
func Destroy(autoApprove bool, vars map[string]string, varFiles []string) error {
	args := []string{"destroy"}
	if autoApprove {
		args = append(args, "-auto-approve")
	}

	for k, v := range vars {
		args = append(args, "-var", fmt.Sprintf("%s=%s", k, v))
	}

	for _, f := range varFiles {
		args = append(args, "-var-file="+f)
	}

	return Run(args)
}

// Validate validates the configuration.
func Validate() error {
	return Run([]string{"validate"})
}

// Fmt formats the configuration files.
func Fmt(check, recursive, diff bool) error {
	args := []string{"fmt"}
	if check {
		args = append(args, "-check")
	}

	if recursive {
		args = append(args, "-recursive")
	}

	if diff {
		args = append(args, "-diff")
	}

	return Run(args)
}

// Output shows output values.
func Output(name string, json bool) error {
	args := []string{"output"}
	if json {
		args = append(args, "-json")
	}

	if name != "" {
		args = append(args, name)
	}

	return Run(args)
}

// State management commands

// StateList lists resources in the state.
func StateList(addresses ...string) error {
	args := append([]string{"state", "list"}, addresses...)
	return Run(args)
}

// StateShow shows a resource in the state.
func StateShow(address string) error {
	return Run([]string{"state", "show", address})
}

// StateMv moves a resource in the state.
func StateMv(source, destination string, dryRun bool) error {
	args := []string{"state", "mv"}
	if dryRun {
		args = append(args, "-dry-run")
	}

	args = append(args, source, destination)

	return Run(args)
}

// StateRm removes a resource from the state.
func StateRm(addresses []string, dryRun bool) error {
	args := []string{"state", "rm"}
	if dryRun {
		args = append(args, "-dry-run")
	}

	args = append(args, addresses...)

	return Run(args)
}

// StatePull pulls and outputs the state.
func StatePull() error {
	return Run([]string{"state", "pull"})
}

// StatePush pushes a local state file to remote.
func StatePush(path string, force bool) error {
	args := []string{"state", "push"}
	if force {
		args = append(args, "-force")
	}

	args = append(args, path)

	return Run(args)
}

// Workspace management commands

// WorkspaceList lists workspaces.
func WorkspaceList() error {
	return Run([]string{"workspace", "list"})
}

// WorkspaceNew creates a new workspace.
func WorkspaceNew(name string) error {
	return Run([]string{"workspace", "new", name})
}

// WorkspaceSelect selects a workspace.
func WorkspaceSelect(name string) error {
	return Run([]string{"workspace", "select", name})
}

// WorkspaceDelete deletes a workspace.
func WorkspaceDelete(name string, force bool) error {
	args := []string{"workspace", "delete"}
	if force {
		args = append(args, "-force")
	}

	args = append(args, name)

	return Run(args)
}

// WorkspaceShow shows the current workspace.
func WorkspaceShow() error {
	return Run([]string{"workspace", "show"})
}

// Import imports existing infrastructure into state.
func Import(address, id string, vars map[string]string, varFiles []string) error {
	args := make([]string, 0, 1+2*len(vars)+len(varFiles)+2)

	args = append(args, "import")
	for k, v := range vars {
		args = append(args, "-var", fmt.Sprintf("%s=%s", k, v))
	}

	for _, f := range varFiles {
		args = append(args, "-var-file="+f)
	}

	args = append(args, address, id)

	return Run(args)
}

// Taint marks a resource for recreation.
func Taint(address string) error {
	return Run([]string{"taint", address})
}

// Untaint removes the taint from a resource.
func Untaint(address string) error {
	return Run([]string{"untaint", address})
}

// Refresh refreshes the state.
func Refresh(vars map[string]string, varFiles []string) error {
	args := make([]string, 0, 1+2*len(vars)+len(varFiles))

	args = append(args, "refresh")
	for k, v := range vars {
		args = append(args, "-var", fmt.Sprintf("%s=%s", k, v))
	}

	for _, f := range varFiles {
		args = append(args, "-var-file="+f)
	}

	return Run(args)
}

// Providers lists and manages providers.
func Providers() error {
	return Run([]string{"providers"})
}

// ProvidersLock locks provider versions.
func ProvidersLock(platforms ...string) error {
	args := make([]string, 0, 2+len(platforms))

	args = append(args, "providers", "lock")
	for _, p := range platforms {
		args = append(args, "-platform="+p)
	}

	return Run(args)
}

// Graph generates a dependency graph.
func Graph(plan string, drawCycles bool) error {
	args := []string{"graph"}
	if plan != "" {
		args = append(args, "-plan="+plan)
	}

	if drawCycles {
		args = append(args, "-draw-cycles")
	}

	return Run(args)
}

// Show shows a plan file or state.
func Show(path string, json bool) error {
	args := []string{"show"}
	if json {
		args = append(args, "-json")
	}

	if path != "" {
		args = append(args, path)
	}

	return Run(args)
}

// Version shows the terraform version.
func Version() error {
	return Run([]string{"version"})
}

// Get downloads modules.
func Get(update bool) error {
	args := []string{"get"}
	if update {
		args = append(args, "-update")
	}

	return Run(args)
}

// ForceUnlock releases a stuck state lock.
func ForceUnlock(lockID string, force bool) error {
	args := []string{"force-unlock"}
	if force {
		args = append(args, "-force")
	}

	args = append(args, lockID)

	return Run(args)
}

// Console starts an interactive console.
func Console() error {
	return Run([]string{"console"})
}

// Test runs tests.
func Test() error {
	return Run([]string{"test"})
}
