//nolint:errcheck,nilerr // Borrowed code from twig
package scanner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/inovacc/omni/pkg/tree/models"
)

// Sentinel errors for scanner operations
var (
	ErrPathNotFound     = errors.New("path not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrInvalidPath      = errors.New("invalid path")
)

// DirectoryScanner defines the interface for scanning directory structures
type DirectoryScanner interface {
	Scan(ctx context.Context, rootPath string) (*models.Node, error)
}

// ScanConfig holds a configuration for scanning
type ScanConfig struct {
	MaxDepth       int
	ShowHidden     bool
	IgnorePatterns []string
	DirsOnly       bool
	ShowHash       bool // Calculate file hashes
}

// DefaultConfig returns default scanning configuration
func DefaultConfig() *ScanConfig {
	return &ScanConfig{
		MaxDepth:   -1, // unlimited
		ShowHidden: false,
		IgnorePatterns: []string{
			".git",
			"node_modules",
			".DS_Store",
			"__pycache__",
			"*.pyc",
			".idea",
			".vscode",
		},
		DirsOnly: false,
	}
}

// Scanner scans directory structures
type Scanner struct {
	config *ScanConfig
}

// Ensure Scanner implements DirectoryScanner
var _ DirectoryScanner = (*Scanner)(nil)

// NewScanner creates a new scanner with a given config and returns a DirectoryScanner interface
func NewScanner(config *ScanConfig) DirectoryScanner {
	if config == nil {
		config = DefaultConfig()
	}

	return &Scanner{config: config}
}

// Scan scans a directory and returns the root node with context support for cancellation
func (s *Scanner) Scan(ctx context.Context, rootPath string) (*models.Node, error) {
	// Check context before starting
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, errors.Join(ErrInvalidPath, err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Join(ErrPathNotFound, err)
		}

		if os.IsPermission(err) {
			return nil, errors.Join(ErrPermissionDenied, err)
		}

		return nil, err
	}

	root := models.NewNode(filepath.Base(absPath), absPath, info.IsDir())
	root.FileInfo = info

	if info.IsDir() {
		if err := s.scanDir(ctx, root, 0); err != nil {
			return nil, err
		}
	}

	return root, nil
}

func (s *Scanner) scanDir(ctx context.Context, parent *models.Node, currentDepth int) error {
	// Check for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	// Check max depth
	if s.config.MaxDepth >= 0 && currentDepth >= s.config.MaxDepth {
		return nil
	}

	entries, err := os.ReadDir(parent.Path)
	if err != nil {
		// Permission denied or other error, skip this directory
		return nil
	}

	for _, entry := range entries {
		// Skip hidden files if configured
		if !s.config.ShowHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Check ignore patterns
		if s.shouldIgnore(entry.Name()) {
			continue
		}

		fullPath := filepath.Join(parent.Path, entry.Name())

		info, err := entry.Info()
		if err != nil {
			continue // Skip if we can't get info
		}

		isDir := entry.IsDir()

		// Skip files if DirsOnly is enabled
		if s.config.DirsOnly && !isDir {
			continue
		}

		child := models.NewNode(entry.Name(), fullPath, isDir)
		child.FileInfo = info

		// Calculate hash for files if enabled
		if s.config.ShowHash && !isDir {
			hash, err := s.calculateFileHash(fullPath)
			if err == nil {
				child.Hash = hash
			}
		}

		parent.AddChild(child)

		// Recursively scan directories
		if isDir {
			if err := s.scanDir(ctx, child, currentDepth+1); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Scanner) shouldIgnore(name string) bool {
	for _, pattern := range s.config.IgnorePatterns {
		matched, err := filepath.Match(pattern, name)
		if err == nil && matched {
			return true
		}
	}

	return false
}

// calculateFileHash calculates the SHA256 hash of a file
func (s *Scanner) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
