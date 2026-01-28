//nolint:errcheck,forbidigo // Borrowed code from twig
package builder

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/inovacc/omni/internal/twig/models"
)

// Sentinel errors for builder operations
var (
	ErrInvalidTargetPath = errors.New("invalid target path")
	ErrTargetExists      = errors.New("target directory already exists")
	ErrTypeMismatch      = errors.New("type mismatch")
	ErrPathTraversal     = errors.New("path traversal detected")
	ErrInvalidCharacters = errors.New("invalid characters in path")
)

// StructureBuilder defines the interface for building directory structures
type StructureBuilder interface {
	Build(ctx context.Context, root *models.Node, targetPath string) (*BuildResult, error)
}

// BuildConfig holds configuration for building
type BuildConfig struct {
	DryRun          bool
	Overwrite       bool
	SkipExisting    bool
	AbortOnConflict bool
	Verbose         bool
}

// DefaultBuildConfig returns default build configuration
func DefaultBuildConfig() *BuildConfig {
	return &BuildConfig{
		DryRun:          false,
		Overwrite:       false,
		SkipExisting:    false,
		AbortOnConflict: false,
		Verbose:         false,
	}
}

// Builder creates physical files and directories from node structure
type Builder struct {
	config *BuildConfig
}

// Ensure Builder implements StructureBuilder
var _ StructureBuilder = (*Builder)(nil)

// NewBuilder creates a new builder and returns a StructureBuilder interface
func NewBuilder(config *BuildConfig) StructureBuilder {
	if config == nil {
		config = DefaultBuildConfig()
	}

	return &Builder{config: config}
}

// BuildResult contains the result of a build operation
type BuildResult struct {
	Created []string
	Skipped []string
	Errors  []error
	DryRun  bool
}

// Build creates the directory structure at targetPath with context support for cancellation
func (b *Builder) Build(ctx context.Context, root *models.Node, targetPath string) (*BuildResult, error) {
	// Check context before starting
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := &BuildResult{
		Created: make([]string, 0),
		Skipped: make([]string, 0),
		Errors:  make([]error, 0),
		DryRun:  b.config.DryRun,
	}

	// Validate a target path
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return nil, errors.Join(ErrInvalidTargetPath, err)
	}

	// Check if a target exists
	if _, err := os.Stat(absPath); err == nil {
		// Target exists
		if b.config.AbortOnConflict {
			return nil, fmt.Errorf("%w: %s", ErrTargetExists, absPath)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("error checking target path: %w", err)
	}

	// Create a target directory if it doesn't exist
	if !b.config.DryRun {
		if err := os.MkdirAll(absPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create target directory: %w", err)
		}
	}

	// Build the structure
	if err := b.buildNode(ctx, root, absPath, result); err != nil {
		return result, err
	}

	return result, nil
}

func (b *Builder) buildNode(ctx context.Context, node *models.Node, currentPath string, result *BuildResult) error {
	// Check for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if node == nil {
		return nil
	}

	// Build children
	for _, child := range node.Children {
		childPath := filepath.Join(currentPath, child.Name)

		// Check if already exists
		exists := false
		if info, err := os.Stat(childPath); err == nil {
			exists = true

			// Check type mismatch
			if child.IsDir != info.IsDir() {
				err := fmt.Errorf("%w at %s: expected %s, found %s",
					ErrTypeMismatch,
					childPath,
					b.nodeType(child.IsDir),
					b.nodeType(info.IsDir()))
				result.Errors = append(result.Errors, err)

				continue
			}

			// Handle existing items
			if b.config.AbortOnConflict {
				return fmt.Errorf("item already exists: %s", childPath)
			}

			if b.config.SkipExisting {
				result.Skipped = append(result.Skipped, childPath)
				if b.config.Verbose {
					fmt.Printf("Skipped (exists): %s\n", childPath)
				}

				continue
			}

			if !b.config.Overwrite && !child.IsDir {
				result.Skipped = append(result.Skipped, childPath)
				if b.config.Verbose {
					fmt.Printf("Skipped (exists): %s\n", childPath)
				}

				continue
			}
		}

		// Validate path (security check)
		if err := b.validatePath(childPath, currentPath); err != nil {
			result.Errors = append(result.Errors, err)
			continue
		}

		if child.IsDir {
			// Create directory
			if !b.config.DryRun {
				if err := os.MkdirAll(childPath, 0755); err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("failed to create directory %s: %w", childPath, err))
					continue
				}
			}

			if !exists {
				result.Created = append(result.Created, childPath)
			}

			if b.config.Verbose {
				action := "Created"
				if b.config.DryRun {
					action = "[DRY RUN] Would create"
				}

				fmt.Printf("%s directory: %s\n", action, childPath)
			}

			// Recursively build children
			if err := b.buildNode(ctx, child, currentPath, result); err != nil {
				return err
			}
		} else {
			// Create file
			if !b.config.DryRun {
				// Create a parent directory if needed
				parentDir := filepath.Dir(childPath)
				if err := os.MkdirAll(parentDir, 0755); err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("failed to create parent directory for %s: %w", childPath, err))
					continue
				}

				// Create an empty file
				file, err := os.Create(childPath)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("failed to create file %s: %w", childPath, err))
					continue
				}

				// Write comment as content if present
				if child.Comment != "" {
					content := fmt.Sprintf("# %s\n", child.Comment)
					if _, err := file.WriteString(content); err != nil {
						file.Close()

						result.Errors = append(result.Errors, fmt.Errorf("failed to write to file %s: %w", childPath, err))

						continue
					}
				}

				file.Close()
			}

			if !exists {
				result.Created = append(result.Created, childPath)
			}

			if b.config.Verbose {
				action := "Created"
				if b.config.DryRun {
					action = "[DRY RUN] Would create"
				}

				fmt.Printf("%s file: %s\n", action, childPath)
			}
		}
	}

	return nil
}

func (b *Builder) nodeType(isDir bool) string {
	if isDir {
		return "directory"
	}

	return "file"
}

func (b *Builder) validatePath(path, basePath string) error {
	// Prevent path traversal attacks
	absPath, err := filepath.Abs(path)
	if err != nil {
		return errors.Join(ErrInvalidTargetPath, err)
	}

	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return errors.Join(ErrInvalidTargetPath, err)
	}

	// Check if a path is under a base path
	if !strings.HasPrefix(absPath, absBase) {
		return fmt.Errorf("%w: %s", ErrPathTraversal, path)
	}

	// Check for invalid characters (basic check)
	name := filepath.Base(path)
	if strings.ContainsAny(name, "<>:\"|?*") {
		return fmt.Errorf("%w: %s", ErrInvalidCharacters, name)
	}

	return nil
}

// PrintResult prints a formatted summary of the build operation results to stdout,
// including counts of created items, skipped items, and errors. If errors occurred,
// they are listed individually. For dry runs, displays a notice that no changes were made.
func PrintResult(result *BuildResult) {
	if result.DryRun {
		fmt.Println("\n=== DRY RUN - No changes were made ===")
	}

	fmt.Println("\nBuild Result:")
	fmt.Printf("  • Created: %d items\n", len(result.Created))
	fmt.Printf("  • Skipped: %d items\n", len(result.Skipped))
	fmt.Printf("  • Errors: %d\n", len(result.Errors))

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")

		for _, err := range result.Errors {
			fmt.Printf("  ✗ %v\n", err)
		}
	}
}
