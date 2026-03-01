package pipeline

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/inovacc/omni/internal/cli/cmderr"
	pkgpipeline "github.com/inovacc/omni/pkg/pipeline"
)

// Options configures the pipeline command behavior.
type Options struct {
	File    string // -f: input file
	Verbose bool   // -v: show stage names
}

// Run executes the pipeline with the given stage definitions.
func Run(ctx context.Context, w io.Writer, r io.Reader, args []string, opts Options) error {
	if len(args) == 0 {
		return cmderr.Wrap(cmderr.ErrInvalidInput, "pipeline: no stages provided")
	}

	// Parse stage definitions
	stages, err := pkgpipeline.ParseAll(args)
	if err != nil {
		return err
	}

	// Determine input source
	var input = r

	if opts.File != "" {
		f, err := os.Open(opts.File)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("pipeline: %s", err))
			}
			return fmt.Errorf("pipeline: %w", err)
		}

		defer func() { _ = f.Close() }()

		input = f
	}

	if opts.Verbose {
		for i, s := range stages {
			_, _ = fmt.Fprintf(w, "--- stage %d: %s ---\n", i+1, s.Name())
		}
	}

	p := pkgpipeline.New(stages...)

	return p.Run(ctx, input, w)
}
