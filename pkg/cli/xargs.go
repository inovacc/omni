package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// XargsOptions configures the xargs command behavior
type XargsOptions struct {
	MaxArgs     int    // -n: use at most MAX-ARGS arguments per command line
	MaxProcs    int    // -P: run up to MAX-PROCS processes at a time
	Delimiter   string // -d: use DELIM instead of whitespace
	NullInput   bool   // -0: items are separated by a null, not whitespace
	NoRunEmpty  bool   // -r: do not run if input is empty
	Verbose     bool   // -t: print commands before executing
	Interactive bool   // -p: prompt before running each command
	ReplaceStr  string // -I: replace occurrences of REPLACE-STR in initial args
}

// XargsWorkerFunc is the function type for xargs workers
type XargsWorkerFunc func(args []string) error

// RunXargs reads arguments from input and passes them to a worker function
// Note: omni doesn't exec external commands, so this works with worker functions
func RunXargs(w io.Writer, r io.Reader, initialArgs []string, opts XargsOptions, worker XargsWorkerFunc) error {
	// Parse input into arguments
	args, err := parseXargsInput(r, opts)
	if err != nil {
		return err
	}

	if len(args) == 0 && opts.NoRunEmpty {
		return nil
	}

	// If no worker provided, just print the arguments
	if worker == nil {
		worker = func(a []string) error {
			_, _ = fmt.Fprintln(w, strings.Join(a, " "))
			return nil
		}
	}

	// Determine batch size
	batchSize := len(args)
	if opts.MaxArgs > 0 {
		batchSize = opts.MaxArgs
	}

	// Create batches
	var batches [][]string

	for i := 0; i < len(args); i += batchSize {
		end := min(i+batchSize, len(args))

		batch := args[i:end]

		// Handle replacement string
		if opts.ReplaceStr != "" && len(initialArgs) > 0 {
			var expandedArgs []string

			for _, arg := range initialArgs {
				for _, input := range batch {
					expandedArgs = append(expandedArgs, strings.ReplaceAll(arg, opts.ReplaceStr, input))
				}
			}

			batches = append(batches, expandedArgs)
		} else {
			combined := make([]string, 0, len(initialArgs)+len(batch))
			combined = append(combined, initialArgs...)
			combined = append(combined, batch...)
			batches = append(batches, combined)
		}
	}

	// Execute batches
	maxProcs := 1
	if opts.MaxProcs > 0 {
		maxProcs = opts.MaxProcs
	}

	if maxProcs == 1 {
		// Sequential execution
		for _, batch := range batches {
			if opts.Verbose {
				_, _ = fmt.Fprintf(os.Stderr, "%s\n", strings.Join(batch, " "))
			}

			if err := worker(batch); err != nil {
				return err
			}
		}
	} else {
		// Parallel execution
		var wg sync.WaitGroup

		sem := make(chan struct{}, maxProcs)
		errCh := make(chan error, len(batches))

		for _, batch := range batches {
			wg.Add(1)

			go func(b []string) {
				defer wg.Done()

				sem <- struct{}{}

				defer func() { <-sem }()

				if opts.Verbose {
					_, _ = fmt.Fprintf(os.Stderr, "%s\n", strings.Join(b, " "))
				}

				if err := worker(b); err != nil {
					errCh <- err
				}
			}(batch)
		}

		wg.Wait()
		close(errCh)

		// Return first error if any
		for err := range errCh {
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// parseXargsInput parses input according to xargs options
func parseXargsInput(r io.Reader, opts XargsOptions) ([]string, error) {
	var args []string

	if opts.NullInput {
		// Read null-terminated items
		scanner := bufio.NewScanner(r)
		scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if atEOF && len(data) == 0 {
				return 0, nil, nil
			}

			if i := bytes.IndexByte(data, 0); i >= 0 {
				return i + 1, data[0:i], nil
			}

			if atEOF {
				return len(data), data, nil
			}

			return 0, nil, nil
		})

		for scanner.Scan() {
			if text := scanner.Text(); text != "" {
				args = append(args, text)
			}
		}

		return args, scanner.Err()
	}

	if opts.Delimiter != "" {
		// Read with custom delimiter
		scanner := bufio.NewScanner(r)

		delim := opts.Delimiter
		if len(delim) > 0 {
			delimByte := delim[0]

			scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
				if atEOF && len(data) == 0 {
					return 0, nil, nil
				}

				if i := bytes.IndexByte(data, delimByte); i >= 0 {
					return i + 1, data[0:i], nil
				}

				if atEOF {
					return len(data), data, nil
				}

				return 0, nil, nil
			})
		}

		for scanner.Scan() {
			if text := scanner.Text(); text != "" {
				args = append(args, text)
			}
		}

		return args, scanner.Err()
	}

	// Default: whitespace separated
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		args = append(args, scanner.Text())
	}

	return args, scanner.Err()
}

// RunXargsWithPrint is a convenience function that just prints arguments
func RunXargsWithPrint(w io.Writer, r io.Reader, opts XargsOptions) error {
	return RunXargs(w, r, nil, opts, func(args []string) error {
		_, _ = fmt.Fprintln(w, strings.Join(args, " "))
		return nil
	})
}
