// Package pipeline provides a streaming text processing engine.
// Stages are connected via io.Pipe goroutines for memory-efficient,
// line-by-line processing of large files.
package pipeline

import (
	"context"
	"io"
)

// Stage represents a single processing step in a pipeline.
type Stage interface {
	// Process reads from in, transforms the data, and writes to out.
	Process(ctx context.Context, in io.Reader, out io.Writer) error

	// Name returns a human-readable name for the stage.
	Name() string
}
