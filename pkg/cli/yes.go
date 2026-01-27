package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// RunYes repeatedly outputs a string until killed
func RunYes(ctx context.Context, w io.Writer, args []string) error {
	output := "y"
	if len(args) > 0 {
		output = strings.Join(args, " ")
	}

	// Handle signals for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-sigCh:
			return nil
		default:
			_, err := fmt.Fprintln(w, output)
			if err != nil {
				// Likely a broken pipe, exit gracefully
				return nil
			}
		}
	}
}

// Yes is a simple yes function for scripting
func Yes(output string, count int) []string {
	if output == "" {
		output = "y"
	}
	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = output
	}
	return result
}
