// Package cli defines the Command interface for omni commands.
// New commands should implement this interface. Existing commands
// are not required to adopt it.
package cli

import (
	"context"
	"io"
)

// Command is the unified interface for omni commands.
type Command interface {
	Run(ctx context.Context, w io.Writer, r io.Reader, args []string) error
}
