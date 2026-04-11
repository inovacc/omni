package yes

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// RunYes repeatedly outputs a string until killed
func RunYes(ctx context.Context, w io.Writer, args []string) error {
	output := "y"
	if len(args) > 0 {
		output = strings.Join(args, " ")
	}

	// Handle signals for a graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGPIPE)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-sigCh:
			return nil
		default:
			// Classify once per Pitfall 7: do not allocate per-iteration on broken pipe.
			if _, err := fmt.Fprintln(w, output); err != nil {
				return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("yes: write: %s", err))
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
	for i := range count {
		result[i] = output
	}

	return result
}
