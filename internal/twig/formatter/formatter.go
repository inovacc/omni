//nolint:forbidigo // Borrowed code from twig
package formatter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/inovacc/omni/internal/twig/models"
	"github.com/xlab/treeprint"
)

// TreeFormatter defines the interface for formatting tree structures
type TreeFormatter interface {
	Format(root *models.Node) string
	FormatSimple(root *models.Node) string
	FormatJSON(root *models.Node, stats *models.TreeStats) (string, error)
}

// FormatConfig holds formatting configuration
type FormatConfig struct {
	ShowColors       bool
	ShowSize         bool
	ShowDate         bool
	ShowDirSlash     bool // Add trailing slash to directory names
	ShowHash         bool // Show file hashes
	FlattenFilesHash bool // Flatten tree and show only files with hashes
	JSONOutput       bool // Output as JSON instead of ASCII tree
}

// DefaultFormatConfig returns default formatting configuration
func DefaultFormatConfig() *FormatConfig {
	return &FormatConfig{
		ShowColors:   true,
		ShowSize:     false,
		ShowDate:     false,
		ShowDirSlash: true, // Show slashes by default (standard convention)
	}
}

// Formatter formats node tree to string representation
type Formatter struct {
	config *FormatConfig
}

// Ensure Formatter implements TreeFormatter
var _ TreeFormatter = (*Formatter)(nil)

// NewFormatter creates a new formatter and returns a TreeFormatter interface
func NewFormatter(config *FormatConfig) TreeFormatter {
	if config == nil {
		config = DefaultFormatConfig()
	}

	return &Formatter{config: config}
}

// Format formats the node tree to ASCII tree format
func (f *Formatter) Format(root *models.Node) string {
	if root == nil {
		return ""
	}

	tree := treeprint.New()

	// Set root name
	rootName := root.Name
	if root.IsDir && f.config.ShowDirSlash {
		rootName += "/"
	}

	tree.SetValue(f.formatName(rootName, root.IsDir))

	// Add children
	f.addChildren(tree, root)

	return tree.String()
}

// FormatSimple formats without using treeprint library (for parsing compatibility)
func (f *Formatter) FormatSimple(root *models.Node) string {
	if root == nil {
		return ""
	}

	// If FlattenFilesHash is enabled, use flattened format
	if f.config.FlattenFilesHash {
		return f.formatFlattened(root)
	}

	var sb strings.Builder

	// Write root
	name := root.Name
	if root.IsDir && f.config.ShowDirSlash {
		name += "/"
	}

	sb.WriteString(name + "\n")

	// Write children
	f.formatSimpleRecursive(&sb, root.Children, "")

	return sb.String()
}

func (f *Formatter) formatSimpleRecursive(sb *strings.Builder, nodes []*models.Node, prefix string) {
	for i, node := range nodes {
		isLast := i == len(nodes)-1

		// Draw the branch
		if isLast {
			sb.WriteString(prefix + "└── ")
		} else {
			sb.WriteString(prefix + "├── ")
		}

		// Draw the name
		name := node.Name
		if node.IsDir && f.config.ShowDirSlash {
			name += "/"
		}

		sb.WriteString(name)

		// Add hash if present and ShowHash is enabled
		if f.config.ShowHash && node.Hash != "" {
			sb.WriteString("  [" + node.Hash + "]")
		}

		// Add comment if present
		if node.Comment != "" {
			sb.WriteString("  # " + node.Comment)
		}

		sb.WriteString("\n")

		// Recursively format children
		if len(node.Children) > 0 {
			newPrefix := prefix
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}

			f.formatSimpleRecursive(sb, node.Children, newPrefix)
		}
	}
}

func (f *Formatter) addChildren(tree treeprint.Tree, node *models.Node) {
	for _, child := range node.Children {
		var branch treeprint.Tree

		childName := child.Name
		if child.IsDir {
			if f.config.ShowDirSlash {
				childName += "/"
			}

			branch = tree.AddBranch(f.formatName(childName, true))
			f.addChildren(branch, child)
		} else {
			tree.AddNode(f.formatName(child.Name, false))
		}
	}
}

func (f *Formatter) formatName(name string, isDir bool) string {
	if !f.config.ShowColors {
		return name
	}

	if isDir {
		return color.BlueString(name)
	}

	return name
}

// formatFlattened formats the tree in a flattened format showing only files with hashes
func (f *Formatter) formatFlattened(root *models.Node) string {
	var sb strings.Builder
	f.collectFilesWithHashes(&sb, root, "")

	return sb.String()
}

// collectFilesWithHashes recursively collects files and their hashes in flattened format
func (f *Formatter) collectFilesWithHashes(sb *strings.Builder, node *models.Node, path string) {
	// Build current path
	currentPath := path
	if currentPath != "" {
		currentPath += "/"
	}

	currentPath += node.Name

	// If it's a file, output it with hash
	if !node.IsDir && node.Hash != "" {
		sb.WriteString(node.Hash)
		sb.WriteString("  ")
		sb.WriteString(currentPath)
		sb.WriteString("\n")
	}

	// Recursively process children
	for _, child := range node.Children {
		f.collectFilesWithHashes(sb, child, currentPath)
	}
}

// PrintStats prints a formatted summary of tree statistics to stdout,
// including total directories, files, maximum depth, and total items.
func PrintStats(stats *models.TreeStats) {
	fmt.Printf("\n")
	fmt.Printf("Summary:\n")
	fmt.Printf("  • Directories: %d\n", stats.TotalDirs)
	fmt.Printf("  • Files: %d\n", stats.TotalFiles)
	fmt.Printf("  • Max depth: %d\n", stats.MaxDepth)
	fmt.Printf("  • Total items: %d\n", stats.TotalDirs+stats.TotalFiles)
}

// FormatJSON formats the node tree and optional stats as JSON.
// If stats is nil, only the tree is included in the output.
func (f *Formatter) FormatJSON(root *models.Node, stats *models.TreeStats) (string, error) {
	if root == nil {
		return "{}", nil
	}

	output := &models.JSONOutput{
		Tree: root.ToJSON(),
	}

	if stats != nil {
		output.Stats = stats.ToJSONStats()
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonBytes) + "\n", nil
}
