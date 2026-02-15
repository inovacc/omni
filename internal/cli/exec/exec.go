// Package exec provides a safe external command wrapper with pre-flight credential checks.
package exec

import (
	"fmt"
	"io"
	"os"
	osexec "os/exec"
	"strings"
	"text/tabwriter"

	"github.com/inovacc/omni/internal/cli/output"
)

// Options configures the exec behavior.
type Options struct {
	Force        bool          // skip credential checks
	Strict       bool          // abort if credentials missing
	DryRun       bool          // show checks without executing
	NoPrompt     bool          // don't prompt, just warn and proceed
	OutputFormat output.Format // output format (text/json)
}

// Run executes a command with pre-flight credential detection.
func Run(w io.Writer, command string, args []string, opts Options) error {
	if command == "" {
		return fmt.Errorf("no command specified")
	}

	if opts.Force && !opts.DryRun {
		return execute(command, args)
	}

	registry := NewDetectorRegistry()
	detectors := registry.Match(command)

	var results []CredentialStatus
	for _, d := range detectors {
		status := d(command, args)
		if status.Needed {
			results = append(results, status)
		}
	}

	if opts.DryRun {
		f := output.New(w, opts.OutputFormat)
		if f.IsJSON() {
			return f.Print(buildDryRunResult(command, args, results))
		}
		printResults(w, command, args, results)
		return nil
	}

	var missing []CredentialStatus
	for _, r := range results {
		if !r.Present {
			missing = append(missing, r)
		}
	}

	if len(missing) == 0 {
		return execute(command, args)
	}

	// Print warning table
	printMissing(w, missing)

	if opts.Strict {
		return fmt.Errorf("aborting: missing credentials (strict mode)")
	}

	if !opts.NoPrompt {
		_, _ = fmt.Fprint(w, "\nContinue anyway? [y/N] ")
		var answer string
		_, _ = fmt.Scanln(&answer)
		if !strings.HasPrefix(strings.ToLower(answer), "y") {
			return fmt.Errorf("aborted by user")
		}
	}

	return execute(command, args)
}

func execute(command string, args []string) error {
	cmd := osexec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func printResults(w io.Writer, command string, args []string, results []CredentialStatus) {
	_, _ = fmt.Fprintf(w, "Dry run: omni exec %s %s\n\n", command, strings.Join(args, " "))

	if len(results) == 0 {
		_, _ = fmt.Fprintln(w, "No credential checks matched for this command.")
		return
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "TOOL\tSTATUS\tDETAILS")
	_, _ = fmt.Fprintln(tw, "----\t------\t-------")
	for _, r := range results {
		status := "OK"
		details := "Credentials found"
		if !r.Present {
			status = "MISSING"
			details = strings.Join(r.Missing, ", ")
		}
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\n", r.Tool, status, details)
	}
	_ = tw.Flush()
}

// DryRunResult represents dry-run output for JSON mode
type DryRunResult struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Checks  []CredentialCheck `json:"checks"`
}

// CredentialCheck represents a single credential check result
type CredentialCheck struct {
	Tool       string   `json:"tool"`
	Status     string   `json:"status"`
	Missing    []string `json:"missing,omitempty"`
	Suggestion string   `json:"suggestion,omitempty"`
}

func buildDryRunResult(command string, args []string, results []CredentialStatus) DryRunResult {
	r := DryRunResult{Command: command, Args: args}
	for _, s := range results {
		check := CredentialCheck{Tool: s.Tool}
		if s.Present {
			check.Status = "ok"
		} else {
			check.Status = "missing"
			check.Missing = s.Missing
			check.Suggestion = s.Suggestion
		}
		r.Checks = append(r.Checks, check)
	}
	return r
}

func printMissing(w io.Writer, missing []CredentialStatus) {
	_, _ = fmt.Fprintln(w, "\nWarning: missing credentials detected")
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "TOOL\tMISSING\tSUGGESTION")
	_, _ = fmt.Fprintln(tw, "----\t-------\t----------")
	for _, m := range missing {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\n", m.Tool, strings.Join(m.Missing, ", "), m.Suggestion)
	}
	_ = tw.Flush()
}
