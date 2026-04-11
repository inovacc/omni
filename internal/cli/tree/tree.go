package tree

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/cobra/helper/output"
	twig2 "github.com/inovacc/omni/pkg/twig"
	"github.com/inovacc/omni/pkg/twig/comparer"
	"github.com/inovacc/omni/pkg/twig/models"
)

// TreeOptions configures the tree command behavior
type TreeOptions struct {
	All          bool          // -a: show hidden files
	DirsOnly     bool          // -d: show only directories
	Depth        int           // --depth: maximum depth (-1 for unlimited)
	Ignore       []string      // -i: patterns to ignore
	NoDirSlash   bool          // don't add trailing slash to directories
	Stats        bool          // -s: show statistics
	Hash         bool          // --hash: show file hashes
	JSON         bool          // -j: output as JSON (local flag, kept for -j shorthand)
	JSONStream   bool          // --json-stream: streaming NDJSON output
	OutputFormat output.Format // global output format
	NoColor      bool          // --no-color: disable colors
	Size         bool          // --size: show file sizes
	Date         bool          // --date: show modification dates
	MaxFiles     int           // --max-files: cap total scanned items
	MaxHashSize  int64         // --max-hash-size: skip hashing large files
	Threads      int           // -t/--threads: parallel workers
	Compare      []string      // --compare: two JSON files to compare
	DetectMoves  bool          // --detect-moves: detect moved files in compare
}

// RunTree executes the tree command
func RunTree(w io.Writer, args []string, opts TreeOptions) error {
	// Handle compare mode
	if len(opts.Compare) == 2 {
		return runCompare(w, opts)
	}

	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	// Merge local -j flag with global --json
	useJSON := opts.JSON || opts.OutputFormat == output.FormatJSON

	// Build tree options
	var treeOpts []twig2.TreeOption

	treeOpts = append(treeOpts, twig2.WithMaxDepth(opts.Depth))
	treeOpts = append(treeOpts, twig2.WithShowHidden(opts.All))
	treeOpts = append(treeOpts, twig2.WithDirsOnly(opts.DirsOnly))
	treeOpts = append(treeOpts, twig2.WithDirSlash(!opts.NoDirSlash))
	treeOpts = append(treeOpts, twig2.WithColors(!opts.NoColor))
	treeOpts = append(treeOpts, twig2.WithJSONOutput(useJSON))
	treeOpts = append(treeOpts, twig2.WithShowHash(opts.Hash))
	treeOpts = append(treeOpts, twig2.WithShowSize(opts.Size))
	treeOpts = append(treeOpts, twig2.WithShowDate(opts.Date))

	if opts.MaxFiles > 0 {
		treeOpts = append(treeOpts, twig2.WithMaxFiles(opts.MaxFiles))
	}

	if opts.MaxHashSize > 0 {
		treeOpts = append(treeOpts, twig2.WithMaxHashSize(opts.MaxHashSize))
	}

	if opts.Threads > 0 {
		treeOpts = append(treeOpts, twig2.WithParallel(opts.Threads))
	}

	if len(opts.Ignore) > 0 {
		treeOpts = append(treeOpts, twig2.WithIgnorePatterns(opts.Ignore...))
	}

	t := twig2.NewTree(treeOpts...)

	// Handle streaming JSON output
	if opts.JSONStream {
		return t.GenerateJSONStream(context.Background(), path, w)
	}

	if opts.Stats {
		result, err := t.GenerateWithStats(context.Background(), path)
		if err != nil {
			return classifyTreeError("tree", err)
		}

		_, _ = fmt.Fprint(w, result.Output)
		_, _ = fmt.Fprintf(w, "\n%d directories, %d files\n", result.Stats.TotalDirs, result.Stats.TotalFiles)
	} else {
		out, err := t.Generate(context.Background(), path)
		if err != nil {
			return classifyTreeError("tree", err)
		}

		// Trim trailing newline if present (Generate adds one)
		out = strings.TrimSuffix(out, "\n")
		if _, err := fmt.Fprintln(w, out); err != nil {
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("tree: write: %s", err))
		}
	}

	return nil
}

// classifyTreeError maps twig/scanner errors to cmderr sentinels at the CLI boundary.
func classifyTreeError(cmd string, err error) error {
	switch {
	case errors.Is(err, os.ErrNotExist):
		return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("%s: %s", cmd, err))
	case errors.Is(err, os.ErrPermission):
		return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("%s: %s", cmd, err))
	default:
		return fmt.Errorf("%s: %w", cmd, err)
	}
}

// runCompare compares two JSON tree snapshots
func runCompare(w io.Writer, opts TreeOptions) error {
	left, err := loadTreeJSON(opts.Compare[0])
	if err != nil {
		return err // already classified by loadTreeJSON
	}

	right, err := loadTreeJSON(opts.Compare[1])
	if err != nil {
		return err // already classified by loadTreeJSON
	}

	cfg := comparer.CompareConfig{
		DetectMoves: opts.DetectMoves,
	}

	result := comparer.Compare(left.Tree, right.Tree, cfg)
	result.LeftPath = opts.Compare[0]
	result.RightPath = opts.Compare[1]

	if opts.JSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		if err := enc.Encode(result); err != nil {
			return fmt.Errorf("tree compare: %w", err)
		}

		return nil
	}

	// Human-readable output
	for _, c := range result.Changes {
		switch c.Type {
		case comparer.Added:
			_, _ = fmt.Fprintf(w, "+ %s\n", c.Path)
		case comparer.Removed:
			_, _ = fmt.Fprintf(w, "- %s\n", c.Path)
		case comparer.Modified:
			_, _ = fmt.Fprintf(w, "~ %s\n", c.Path)
		case comparer.Moved:
			_, _ = fmt.Fprintf(w, "> %s (from %s)\n", c.Path, c.OldPath)
		}
	}

	_, _ = fmt.Fprintf(w, "\n%d added, %d removed, %d modified, %d moved\n",
		result.Summary.Added, result.Summary.Removed, result.Summary.Modified, result.Summary.Moved)

	return nil
}

// loadTreeJSON reads a JSON tree snapshot file and validates it
func loadTreeJSON(path string) (*models.JSONOutput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("tree compare: %s: %s", path, err))
		}
		if errors.Is(err, os.ErrPermission) {
			return nil, cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("tree compare: %s: %s", path, err))
		}
		return nil, fmt.Errorf("tree compare: reading %s: %w", path, err)
	}

	var out models.JSONOutput
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("tree compare: parsing %s: %s", path, err))
	}

	if out.Tree == nil {
		return nil, cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("tree compare: %s: missing 'tree' field", path))
	}

	return &out, nil
}
