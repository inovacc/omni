package cli

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
)

// NohupOptions configures the nohup command behavior
type NohupOptions struct {
	OutputFile string // output file (default: nohup.out)
}

// RunNohup runs a command immune to hangups
// Note: In Go, we can't truly detach a process like Unix nohup does.
// This implementation ignores SIGHUP and redirects output, which provides
// similar behavior for internal commands.
func RunNohup(w io.Writer, args []string, opts NohupOptions) error {
	if len(args) == 0 {
		return fmt.Errorf("nohup: missing operand")
	}

	// Ignore SIGHUP signal
	signal.Ignore(syscall.SIGHUP)

	// Determine output file
	outputFile := opts.OutputFile
	if outputFile == "" {
		outputFile = "nohup.out"
	}

	// Check if stdout is a terminal
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		// Stdout is a terminal, redirect to file
		f, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			// Try home directory
			home, _ := os.UserHomeDir()
			if home != "" {
				outputFile = home + "/nohup.out"
				f, err = os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			}
			if err != nil {
				return fmt.Errorf("nohup: cannot create output file: %w", err)
			}
		}
		defer func() {
			_ = f.Close()
		}()

		_, _ = fmt.Fprintf(os.Stderr, "nohup: ignoring input and appending output to '%s'\n", outputFile)
		w = f
	}

	// Since we can't exec external commands, this is limited to printing a message
	// In a real implementation, this would use syscall.Exec or os/exec
	_, _ = fmt.Fprintf(w, "nohup: would run: %v\n", args)
	_, _ = fmt.Fprintln(w, "nohup: note: goshell cannot spawn external processes")
	_, _ = fmt.Fprintln(w, "nohup: use system nohup for external commands")

	return nil
}

// SetupNohup sets up the environment for nohup-like behavior
// This can be called at the start of a long-running command to make it
// immune to hangups
func SetupNohup() {
	signal.Ignore(syscall.SIGHUP)
}
