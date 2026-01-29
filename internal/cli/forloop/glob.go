package forloop

import (
	"os"
	"path/filepath"
	"strings"
)

// filepathGlob returns files matching the pattern
// Supports ** for recursive matching
func filepathGlob(pattern string) ([]string, error) {
	// Check if pattern contains **
	if strings.Contains(pattern, "**") {
		return globRecursive(pattern)
	}

	// Standard glob
	return filepath.Glob(pattern)
}

// globRecursive handles ** patterns
func globRecursive(pattern string) ([]string, error) {
	// Split pattern at **
	parts := strings.SplitN(pattern, "**", 2)
	if len(parts) != 2 {
		return filepath.Glob(pattern)
	}

	base := parts[0]
	if base == "" {
		base = "."
	}

	suffix := strings.TrimPrefix(parts[1], string(filepath.Separator))

	var matches []string

	err := filepath.Walk(base, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil //nolint:nilerr // Skip permission errors, continue walking
		}

		if suffix == "" {
			matches = append(matches, path)
			return nil
		}

		// Check if file matches suffix pattern
		if !info.IsDir() {
			matched, _ := filepath.Match(suffix, filepath.Base(path))
			if matched {
				matches = append(matches, path)
			}
		}

		return nil
	})

	return matches, err
}
