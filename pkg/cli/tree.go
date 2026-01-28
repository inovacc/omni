package cli

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/inovacc/omni/pkg/tree"
)

// TreeOptions configures the tree command behavior
type TreeOptions struct {
	All        bool     // -a: show hidden files
	DirsOnly   bool     // -d: show only directories
	Depth      int      // --depth: maximum depth (-1 for unlimited)
	Ignore     []string // -i: patterns to ignore
	NoDirSlash bool     // don't add trailing slash to directories
	Stats      bool     // -s: show statistics
	Hash       bool     // --hash: show file hashes
	JSON       bool     // -j: output as JSON
	NoColor    bool     // --no-color: disable colors
	Size       bool     // --size: show file sizes
	Date       bool     // --date: show modification dates
}

// RunTree executes the tree command
func RunTree(w io.Writer, args []string, opts TreeOptions) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	// Build tree options
	var treeOpts []tree.TreeOption

	treeOpts = append(treeOpts, tree.WithMaxDepth(opts.Depth))
	treeOpts = append(treeOpts, tree.WithShowHidden(opts.All))
	treeOpts = append(treeOpts, tree.WithDirsOnly(opts.DirsOnly))
	treeOpts = append(treeOpts, tree.WithDirSlash(!opts.NoDirSlash))
	treeOpts = append(treeOpts, tree.WithColors(!opts.NoColor))
	treeOpts = append(treeOpts, tree.WithJSONOutput(opts.JSON))
	treeOpts = append(treeOpts, tree.WithShowHash(opts.Hash))
	treeOpts = append(treeOpts, tree.WithShowSize(opts.Size))
	treeOpts = append(treeOpts, tree.WithShowDate(opts.Date))

	if len(opts.Ignore) > 0 {
		treeOpts = append(treeOpts, tree.WithIgnorePatterns(opts.Ignore...))
	}

	t := tree.NewTree(treeOpts...)

	if opts.Stats {
		result, err := t.GenerateWithStats(context.Background(), path)
		if err != nil {
			return fmt.Errorf("tree: %w", err)
		}

		_, _ = fmt.Fprint(w, result.Output)
		_, _ = fmt.Fprintf(w, "\n%d directories, %d files\n", result.Stats.TotalDirs, result.Stats.TotalFiles)
	} else {
		output, err := t.Generate(context.Background(), path)
		if err != nil {
			return fmt.Errorf("tree: %w", err)
		}

		// Trim trailing newline if present (Generate adds one)
		output = strings.TrimSuffix(output, "\n")
		_, _ = fmt.Fprintln(w, output)
	}

	return nil
}
