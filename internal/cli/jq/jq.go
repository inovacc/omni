package jq

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/jsonutil"
)

// JqOptions configures the jq command behavior
type JqOptions struct {
	Raw        bool // -r: output raw strings (no quotes)
	Compact    bool // -c: compact output (no pretty print)
	Slurp      bool // -s: read entire input into array
	NullInput  bool // -n: don't read any input
	Tab        bool // --tab: use tabs for indentation
	Sort       bool // -S: sort object keys
	Color      bool // -C: colorize output (not implemented)
	Monochrome bool // -M: monochrome output
}

// RunJq executes jq-like JSON processing
// r is the default input reader (used when no files are specified)
func RunJq(w io.Writer, r io.Reader, args []string, opts JqOptions) error {
	filter := "."

	var files []string

	if len(args) > 0 {
		filter = args[0]
		files = args[1:]
	}

	var inputs []any

	if opts.NullInput {
		inputs = []any{nil}
	} else if len(files) == 0 {
		// Read from provided reader (typically stdin)
		data, err := io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("jq: %w", err)
		}

		if opts.Slurp {
			var items []any

			dec := json.NewDecoder(strings.NewReader(string(data)))

			for {
				var v any
				if err := dec.Decode(&v); err == io.EOF {
					break
				} else if err != nil {
					return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("jq: parse error: %s", err))
				}

				items = append(items, v)
			}

			inputs = []any{items}
		} else {
			var v any
			if err := json.Unmarshal(data, &v); err != nil {
				return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("jq: parse error: %s", err))
			}

			inputs = []any{v}
		}
	} else {
		for _, file := range files {
			var (
				data []byte
				err  error
			)

			if file == "-" {
				data, err = io.ReadAll(r)
			} else {
				data, err = os.ReadFile(file)
			}

			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("jq: %s", file))
				}
				return fmt.Errorf("jq: %w", err)
			}

			var v any
			if err := json.Unmarshal(data, &v); err != nil {
				return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("jq: %s: parse error: %s", file, err))
			}

			inputs = append(inputs, v)
		}

		if opts.Slurp {
			inputs = []any{inputs}
		}
	}

	for _, input := range inputs {
		results, err := ApplyJqFilter(input, filter)
		if err != nil {
			return fmt.Errorf("jq: %w", err)
		}

		for _, result := range results {
			if err := outputJqResult(w, result, opts); err != nil {
				return err
			}
		}
	}

	return nil
}

// ApplyJqFilter delegates to the pkg/jsonutil filter engine.
func ApplyJqFilter(input any, filter string) ([]any, error) {
	return jsonutil.ApplyFilter(input, filter)
}

func outputJqResult(w io.Writer, result any, opts JqOptions) error {
	// Raw string output
	if opts.Raw {
		if s, ok := result.(string); ok {
			_, _ = fmt.Fprintln(w, s)
			return nil
		}
	}

	// JSON output
	encoder := json.NewEncoder(w)

	if !opts.Compact {
		if opts.Tab {
			encoder.SetIndent("", "\t")
		} else {
			encoder.SetIndent("", "  ")
		}
	}

	return encoder.Encode(result)
}
