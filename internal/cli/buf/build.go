package buf

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/inovacc/omni/pkg/buf/pkg/bufapi"
)

// RunBuild builds proto files using the real buf compilation engine.
func RunBuild(w io.Writer, dir string, opts BuildOptions) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}

	if opts.Output != "" {
		// Write image to file
		outputFormat := "json"
		if strings.HasSuffix(opts.Output, ".bin") {
			outputFormat = "bin"
		}

		var buf bytes.Buffer

		fileCount, err := bufapi.BuildDir(context.Background(), &buf, absDir, outputFormat)
		if err != nil {
			return fmt.Errorf("buf: %w", err)
		}

		if fileCount == 0 {
			_, _ = fmt.Fprintln(w, "No proto files found")

			return nil
		}

		if err := os.WriteFile(opts.Output, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("buf: failed to write output: %w", err)
		}

		_, _ = fmt.Fprintf(w, "Built %d file(s) to %s\n", fileCount, opts.Output)
	} else {
		// No output file — write JSON image to stdout
		fileCount, err := bufapi.BuildDir(context.Background(), w, absDir, "json")
		if err != nil {
			return fmt.Errorf("buf: %w", err)
		}

		if fileCount == 0 {
			_, _ = fmt.Fprintln(w, "No proto files found")

			return nil
		}
	}

	return nil
}

// RunBreaking checks for breaking changes between dir and opts.Against.
// Uses the real buf breaking change detection engine (protocompile + bufcheck rules).
func RunBreaking(w io.Writer, dir string, opts BreakingOptions) error {
	if opts.Against == "" {
		return fmt.Errorf("buf: --against is required")
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}

	absAgainst, err := filepath.Abs(opts.Against)
	if err != nil {
		absAgainst = opts.Against
	}

	format := opts.ErrorFormat
	if format == "" {
		format = "text"
	}

	if err := bufapi.BreakingDir(context.Background(), w, absDir, absAgainst, format); err != nil {
		return fmt.Errorf("buf: %w", err)
	}
	return nil
}
