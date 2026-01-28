# Tree Package API

High-level API for directory tree operations: scanning, parsing, formatting, and building.

## Installation

```bash
go get github.com/inovacc/twig/pkg/tree
```

## Quick Start

### Generate a Tree from Directory

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/inovacc/twig/pkg/tree"
)

func main() {
    t := tree.NewTree()
    output, err := t.Generate(context.Background(), "/path/to/scan")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(output)
}
```

**Output:**

```
project/
├── src/
│   ├── main.go
│   └── utils.go
├── go.mod
└── README.md
```

### Create Directory Structure

```go
treeFormat := `my-app/
├── cmd/
│   └── main.go
├── pkg/
│   └── lib.go
└── go.mod`

t := tree.NewTree()
result, err := t.CreateFromString(context.Background(), treeFormat, "./my-app")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Created %d items\n", len(result.BuildResult.Created))
```

## Configuration with Options

Use the options pattern to customize behavior:

```go
t := tree.NewTree(
    tree.WithMaxDepth(3),                              // Limit depth
    tree.WithShowHidden(true),                         // Show hidden files
    tree.WithIgnorePatterns("node_modules", ".git"),   // Ignore patterns
    tree.WithDirSlash(false),                          // No trailing slashes
    tree.WithColors(true),                             // Enable colors
)
```

### Available Options

#### Scanner Options

- `WithMaxDepth(int)` - Maximum depth to scan (-1 for unlimited)
- `WithShowHidden(bool)` - Show hidden files/directories
- `WithIgnorePatterns(...string)` - Patterns to ignore (e.g., `"*.log"`, `"node_modules"`)
- `WithDirsOnly(bool)` - Show only directories
- `WithScanConfig(*scanner.ScanConfig)` - Custom scan configuration

#### Formatter Options

- `WithColors(bool)` - Enable colored output
- `WithDirSlash(bool)` - Add trailing slash to directory names
- `WithShowSize(bool)` - Show file sizes
- `WithShowDate(bool)` - Show modification dates
- `WithFormatConfig(*formatter.FormatConfig)` - Custom format configuration

#### Builder Options

- `WithDryRun(bool)` - Preview without creating
- `WithOverwrite(bool)` - Overwrite existing files
- `WithSkipExisting(bool)` - Skip existing files/directories
- `WithAbortOnConflict(bool)` - Abort if target exists
- `WithVerbose(bool)` - Verbose output
- `WithBuildConfig(*builder.BuildConfig)` - Custom build configuration

## API Reference

### Core Methods

#### Generate Methods

```go
// Generate scans a directory and returns formatted tree
func (t *Tree) Generate(ctx context.Context, path string) (string, error)

// GenerateWithStats returns tree with statistics
func (t *Tree) GenerateWithStats(ctx context.Context, path string) (*GenerateResult, error)
```

**Example:**

```go
result, err := t.GenerateWithStats(ctx, "/path")
fmt.Println(result.Output)
fmt.Printf("Files: %d, Dirs: %d\n",
    result.Stats.TotalFiles,
    result.Stats.TotalDirs)
```

#### Create Methods

```go
// Create from io.Reader
func (t *Tree) Create(ctx context.Context, input io.Reader, targetPath string) (*CreateResult, error)

// CreateFromString from string
func (t *Tree) CreateFromString(ctx context.Context, treeFormat string, targetPath string) (*CreateResult, error)
```

**Example:**

```go
result, err := t.CreateFromString(ctx, treeFormat, "/target")
fmt.Printf("Created: %d\n", len(result.BuildResult.Created))
fmt.Printf("Skipped: %d\n", len(result.BuildResult.Skipped))
fmt.Printf("Errors: %d\n", len(result.BuildResult.Errors))
```

#### Low-Level Methods

```go
// Scan directory without formatting
func (t *Tree) Scan(ctx context.Context, path string) (*models.Node, error)

// Parse tree format without creating
func (t *Tree) Parse(treeFormat string) (*models.Node, error)
func (t *Tree) ParseReader(input io.Reader) (*models.Node, error)

// Format node tree to string
func (t *Tree) Format(root *models.Node) string

// Build directory structure from node
func (t *Tree) Build(ctx context.Context, root *models.Node, targetPath string) (*builder.BuildResult, error)
```

## Advanced Usage Examples

### Custom Filtering

```go
t := tree.NewTree(
    tree.WithIgnorePatterns(
        "*.exe",
        "*.dll",
        "node_modules",
        ".git",
        "__pycache__",
    ),
    tree.WithMaxDepth(5),
    tree.WithShowHidden(false),
)

output, _ := t.Generate(ctx, ".")
```

### Dry Run Before Creating

```go
// First, do a dry run
t := tree.NewTree(tree.WithDryRun(true), tree.WithVerbose(true))
result, err := t.CreateFromString(ctx, treeFormat, "/target")

fmt.Printf("Would create %d items\n", len(result.BuildResult.Created))

// If satisfied, create for real
t = tree.NewTree(tree.WithVerbose(true))
result, err = t.CreateFromString(ctx, treeFormat, "/target")
```

### Manipulate Tree Programmatically

```go
// Scan a directory
root, err := t.Scan(ctx, "/source")

// Modify the tree (e.g., filter out certain files)
filterTree(root, func(n *models.Node) bool {
    return !strings.HasPrefix(n.Name, "test_")
})

// Format modified tree
output := t.Format(root)
fmt.Println(output)

// Or build it to a new location
result, err := t.Build(ctx, root, "/target")
```

### Context Cancellation

```go
// Create cancellable context
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Operations respect context cancellation
output, err := t.Generate(ctx, "/large/directory")
if errors.Is(err, context.DeadlineExceeded) {
    fmt.Println("Operation timed out")
}
```

### Error Handling with Sentinel Errors

```go
import "github.com/inovacc/twig/pkg/tree/scanner"

_, err := t.Scan(ctx, "/path")
if errors.Is(err, scanner.ErrPathNotFound) {
    fmt.Println("Path does not exist")
} else if errors.Is(err, scanner.ErrPermissionDenied) {
    fmt.Println("Permission denied")
}
```

## Integration Example

### Generate Project Documentation

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/inovacc/twig/pkg/tree"
)

func generateProjectDocs(projectPath string) error {
    t := tree.NewTree(
        tree.WithMaxDepth(4),
        tree.WithIgnorePatterns(".git", "node_modules", "vendor"),
        tree.WithDirSlash(true),
        tree.WithColors(false), // For markdown
    )

    result, err := t.GenerateWithStats(context.Background(), projectPath)
    if err != nil {
        return err
    }

    // Write to README.md
    readme := fmt.Sprintf(`# Project Structure

Total Files: %d | Total Directories: %d

\`\`\`
%s
\`\`\`
`, result.Stats.TotalFiles, result.Stats.TotalDirs, result.Output)

    return os.WriteFile("STRUCTURE.md", []byte(readme), 0644)
}
```

### Scaffold New Projects

```go
func scaffoldGoProject(name, path string) error {
    template := fmt.Sprintf(`%s/
├── cmd/
│   └── main.go
├── pkg/
│   └── lib.go
├── internal/
├── tests/
├── go.mod
├── go.sum
├── README.md
├── .gitignore
└── Makefile`, name)

    t := tree.NewTree(
        tree.WithVerbose(true),
        tree.WithSkipExisting(true),
    )

    result, err := t.CreateFromString(context.Background(), template, path)
    if err != nil {
        return err
    }

    fmt.Printf("✓ Created %d items\n", len(result.BuildResult.Created))
    return nil
}
```

## Type Reference

### GenerateResult

```go
type GenerateResult struct {
    Output string              // Formatted tree output
    Stats  *models.TreeStats  // Statistics
    Root   *models.Node       // Root node
}
```

### CreateResult

```go
type CreateResult struct {
    BuildResult *builder.BuildResult  // Build details
    Root        *models.Node          // Parsed root node
}
```

### TreeStats

```go
type TreeStats struct {
    TotalDirs  int  // Total directories
    TotalFiles int  // Total files
    MaxDepth   int  // Maximum depth
}
```

### BuildResult

```go
type BuildResult struct {
    Created []string  // Paths of created items
    Skipped []string  // Paths of skipped items
    Errors  []error   // Errors encountered
    DryRun  bool      // Was this a dry run
}
```

## See Also

- [Examples](../../examples/basic-usage/main.go) - Complete working examples
- [Scanner Package](scanner) - Low-level directory scanning
- [Parser Package](parser) - Tree format parsing
- [Formatter Package](formatter) - Tree formatting
- [Builder Package](builder) - Directory structure building
- [Models Package](models) - Data models

## License

MIT License - See LICENSE file for details
