package pipeline

import (
	"context"
	"fmt"
	"io"
	"sync"
)

// Pipeline chains multiple stages together, connecting them via io.Pipe.
type Pipeline struct {
	stages []Stage
}

// New creates a pipeline with the given stages.
func New(stages ...Stage) *Pipeline {
	return &Pipeline{stages: stages}
}

// Add appends stages to the pipeline.
func (p *Pipeline) Add(stages ...Stage) *Pipeline {
	p.stages = append(p.stages, stages...)
	return p
}

// Stages returns the current stages (for inspection/testing).
func (p *Pipeline) Stages() []Stage {
	return p.stages
}

// Run executes the pipeline, reading from in and writing to out.
// Each stage runs in its own goroutine, connected by io.Pipe.
func (p *Pipeline) Run(ctx context.Context, in io.Reader, out io.Writer) error {
	if len(p.stages) == 0 {
		// No stages: copy input to output
		_, err := io.Copy(out, in)
		return err
	}

	if len(p.stages) == 1 {
		return p.stages[0].Process(ctx, in, out)
	}

	// Create pipes between stages
	// stage[0] reads from `in`, writes to pipes[0]
	// stage[i] reads from pipes[i-1], writes to pipes[i]
	// stage[n-1] reads from pipes[n-2], writes to `out`
	pipes := make([]*io.PipeWriter, len(p.stages)-1)
	readers := make([]*io.PipeReader, len(p.stages)-1)

	for i := range pipes {
		readers[i], pipes[i] = io.Pipe()
	}

	var (
		wg   sync.WaitGroup
		errs = make([]error, len(p.stages))
	)

	// Launch all stages
	for i, stage := range p.stages {
		wg.Add(1)

		go func(idx int, s Stage) {
			defer wg.Done()

			var r io.Reader
			var w io.Writer

			if idx == 0 {
				r = in
			} else {
				r = readers[idx-1]
			}

			if idx == len(p.stages)-1 {
				w = out
			} else {
				w = pipes[idx]
			}

			err := s.Process(ctx, r, w)

			// Close the write end of pipe when done
			if idx < len(p.stages)-1 {
				if err != nil {
					_ = pipes[idx].CloseWithError(err)
				} else {
					_ = pipes[idx].Close()
				}
			}

			// Close the read end if we stopped early (e.g. head)
			if idx > 0 {
				_ = readers[idx-1].Close()
			}

			errs[idx] = err
		}(i, stage)
	}

	wg.Wait()

	// Return the first meaningful error
	for i, err := range errs {
		if err != nil && err != io.ErrClosedPipe {
			return fmt.Errorf("stage %d (%s): %w", i, p.stages[i].Name(), err)
		}
	}

	return nil
}
