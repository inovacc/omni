// Package twig provides a high-level API for scanning, parsing, formatting, and building directory structures.
//
// # Quick Start
//
// Generate a tree from a directory:
//
//	t := tree.NewTree()
//	output, err := t.Generate(ctx, "/path/to/scan")
//	fmt.Println(output)
//
// Create a directory structure from a tree format:
//
//	t := tree.NewTree()
//	result, err := t.CreateFromString(ctx, treeFormat, "/target/path")
//
// # Configuration
//
// Use options to customize behavior:
//
//	t := tree.NewTree(
//	    tree.WithMaxDepth(3),
//	    tree.WithShowHidden(true),
//	    tree.WithIgnorePatterns("node_modules", ".git"),
//	)
package twig

import (
	"context"
	"io"

	"github.com/inovacc/omni/pkg/twig/builder"
	"github.com/inovacc/omni/pkg/twig/formatter"
	"github.com/inovacc/omni/pkg/twig/models"
	"github.com/inovacc/omni/pkg/twig/parser"
	"github.com/inovacc/omni/pkg/twig/scanner"
)

// Tree provides a high-level API for directory tree operations.
type Tree struct {
	scanConfig   *scanner.ScanConfig
	formatConfig *formatter.FormatConfig
	buildConfig  *builder.BuildConfig
}

// GenerateResult contains the output and statistics from a Generate operation.
type GenerateResult struct {
	Output string
	Stats  *models.TreeStats
	Root   *models.Node
}

// CreateResult contains the result of a Create operation.
type CreateResult struct {
	BuildResult *builder.BuildResult
	Root        *models.Node
}

// NewTree creates a new Tree instance with the default configuration.
// Use options to customize behavior.
func NewTree(opts ...TreeOption) *Tree {
	t := &Tree{
		scanConfig:   scanner.DefaultConfig(),
		formatConfig: formatter.DefaultFormatConfig(),
		buildConfig:  builder.DefaultBuildConfig(),
	}

	// Apply options
	for _, opt := range opts {
		opt(t)
	}

	return t
}

// Generate scans a directory and returns its tree representation as a string.
// The output format is compatible with the Unix 'tree' command.
//
// Example:
//
//	t := tree.NewTree()
//	output, err := t.Generate(ctx, "/path/to/scan")
//	if err != nil {
//	    return err
//	}
//	fmt.Println(output)
func (t *Tree) Generate(ctx context.Context, path string) (string, error) {
	result, err := t.GenerateWithStats(ctx, path)
	if err != nil {
		return "", err
	}

	return result.Output, nil
}

// GenerateWithStats scans a directory and returns the tree representation along with statistics.
//
// Example:
//
//	result, err := t.GenerateWithStats(ctx, "/path/to/scan")
//	fmt.Println(result.Output)
//	fmt.Printf("Total files: %d, Total dirs: %d\n",
//	    result.Stats.TotalFiles, result.Stats.TotalDirs)
func (t *Tree) GenerateWithStats(ctx context.Context, path string) (*GenerateResult, error) {
	// Scan directory
	s := scanner.NewScanner(t.scanConfig)

	root, err := s.Scan(ctx, path)
	if err != nil {
		return nil, err
	}

	// Calculate statistics
	stats := models.CalculateStats(root)

	// Format tree
	f := formatter.NewFormatter(t.formatConfig)

	var output string
	if t.formatConfig.JSONOutput {
		output, err = f.FormatJSON(root, stats)
		if err != nil {
			return nil, err
		}
	} else {
		output = f.FormatSimple(root)
	}

	return &GenerateResult{
		Output: output,
		Stats:  stats,
		Root:   root,
	}, nil
}

// GenerateJSON scans a directory and returns JSON representation with optional stats.
//
// Example:
//
//	output, err := t.GenerateJSON(ctx, "/path/to/scan", true)
//	fmt.Println(output)
func (t *Tree) GenerateJSON(ctx context.Context, path string, includeStats bool) (string, error) {
	// Scan directory
	s := scanner.NewScanner(t.scanConfig)

	root, err := s.Scan(ctx, path)
	if err != nil {
		return "", err
	}

	// Calculate statistics if requested
	var stats *models.TreeStats
	if includeStats {
		stats = models.CalculateStats(root)
	}

	// Format as JSON
	f := formatter.NewFormatter(t.formatConfig)

	return f.FormatJSON(root, stats)
}

// GenerateJSONStream scans a directory and writes the tree as streaming NDJSON to the writer.
// Each line is a self-contained JSON object with types: "begin", "node", "stats", "end".
func (t *Tree) GenerateJSONStream(ctx context.Context, path string, w io.Writer) error {
	s := scanner.NewScanner(t.scanConfig)

	root, err := s.Scan(ctx, path)
	if err != nil {
		return err
	}

	stats := models.CalculateStats(root)

	f := formatter.NewFormatter(t.formatConfig)

	return f.FormatJSONStream(w, root, stats)
}

// Create creates a directory structure from a tree format reader.
// The reader should contain tree format text (e.g., from a file or string).
//
// Example:
//
//	file, _ := os.Open("structure.txt")
//	defer file.Close()
//	result, err := t.Create(ctx, file, "/target/path")
func (t *Tree) Create(ctx context.Context, input io.Reader, targetPath string) (*CreateResult, error) {
	// Parse tree structure
	p := parser.NewParser()

	root, err := p.Parse(input)
	if err != nil {
		return nil, err
	}

	// Build structure
	b := builder.NewBuilder(t.buildConfig)

	buildResult, err := b.Build(ctx, root, targetPath)
	if err != nil {
		return nil, err
	}

	return &CreateResult{
		BuildResult: buildResult,
		Root:        root,
	}, nil
}

// CreateFromString creates a directory structure from a tree format string.
// This is a convenience method for Create that accepts a string input.
//
// Example:
//
//	treeFormat := `project/
//	├── src/
//	│   └── main.go
//	└── README.md`
//	result, err := t.CreateFromString(ctx, treeFormat, "/target/path")
func (t *Tree) CreateFromString(ctx context.Context, treeFormat string, targetPath string) (*CreateResult, error) {
	p := parser.NewParser()

	root, err := p.ParseString(treeFormat)
	if err != nil {
		return nil, err
	}

	b := builder.NewBuilder(t.buildConfig)

	buildResult, err := b.Build(ctx, root, targetPath)
	if err != nil {
		return nil, err
	}

	return &CreateResult{
		BuildResult: buildResult,
		Root:        root,
	}, nil
}

// Parse parses a tree format string and returns the root node without creating files.
// This is useful for validation or manipulation before building.
//
// Example:
//
//	root, err := t.Parse(treeFormat)
//	// Modify root...
//	b := builder.NewBuilder(config)
//	b.Build(ctx, root, "/target/path")
func (t *Tree) Parse(treeFormat string) (*models.Node, error) {
	p := parser.NewParser()
	return p.ParseString(treeFormat)
}

// ParseReader parses tree format from a reader and returns the root node.
func (t *Tree) ParseReader(input io.Reader) (*models.Node, error) {
	p := parser.NewParser()
	return p.Parse(input)
}

// Format converts a node tree to a formatted string.
// This is useful when you have a node tree and want to output it.
//
// Example:
//
//	root := models.NewNode("project", "project", true)
//	// Build tree...
//	output := t.Format(root)
func (t *Tree) Format(root *models.Node) string {
	f := formatter.NewFormatter(t.formatConfig)
	return f.FormatSimple(root)
}

// Scan scans a directory and returns the node tree without formatting.
// This is useful when you want to manipulate the tree before formatting.
//
// Example:
//
//	root, err := t.Scan(ctx, "/path/to/scan")
//	// Modify root...
//	output := t.Format(root)
func (t *Tree) Scan(ctx context.Context, path string) (*models.Node, error) {
	s := scanner.NewScanner(t.scanConfig)
	return s.Scan(ctx, path)
}

// Build creates the directory structure from a node tree.
// This is useful when you've constructed or modified a tree programmatically.
//
// Example:
//
//	root := models.NewNode("project", "project", true)
//	// Build tree structure...
//	result, err := t.Build(ctx, root, "/target/path")
func (t *Tree) Build(ctx context.Context, root *models.Node, targetPath string) (*builder.BuildResult, error) {
	b := builder.NewBuilder(t.buildConfig)
	return b.Build(ctx, root, targetPath)
}
