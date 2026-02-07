package rg

import (
	"os"
	"path/filepath"

	pkgrg "github.com/inovacc/omni/pkg/search/rg"
)

// MatchResult indicates the result of matching a path against gitignore patterns
type MatchResult = pkgrg.MatchResult

const (
	// NoMatch means no pattern matched
	NoMatch = pkgrg.NoMatch
	// Ignore means the path should be ignored
	Ignore = pkgrg.Ignore
	// Include means the path was re-included via negation
	Include = pkgrg.Include
)

// Pattern is an alias for the pkg type
type Pattern = pkgrg.Pattern

// Gitignore is an alias for the pkg type
type Gitignore = pkgrg.Gitignore

// GitignoreSet is an alias for the pkg type
type GitignoreSet = pkgrg.GitignoreSet

// NewGitignoreSet creates a new GitignoreSet loading patterns from multiple sources
func NewGitignoreSet(searchDir string) *GitignoreSet {
	absSearchDir, _ := filepath.Abs(searchDir)
	gs := pkgrg.NewGitignoreSet(absSearchDir)

	// 1. Load global gitignore (~/.config/git/ignore)
	if globalPath := getGlobalGitignorePath(); globalPath != "" {
		if gi := loadGitignoreFile(globalPath, ""); gi != nil {
			gs.AddGitignore(gi)
		}
	}

	// 2. Find git root and load .git/info/exclude
	if gitRoot := findGitRoot(searchDir); gitRoot != "" {
		excludePath := filepath.Join(gitRoot, ".git", "info", "exclude")
		if gi := loadGitignoreFile(excludePath, gitRoot); gi != nil {
			gs.AddGitignore(gi)
		}
	}

	// 3. Load .gitignore and .ignore files walking up from searchDir
	loadIgnoreFilesFromHierarchy(gs, searchDir)

	return gs
}

// getGlobalGitignorePath returns the path to the global gitignore file
func getGlobalGitignorePath() string {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		path := filepath.Join(xdgConfig, "git", "ignore")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Fall back to ~/.config/git/ignore
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	path := filepath.Join(home, ".config", "git", "ignore")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	// Also check ~/.gitignore_global (common alternative)
	path = filepath.Join(home, ".gitignore_global")
	if _, err := os.Stat(path); err == nil {
		return path
	}

	return ""
}

// findGitRoot finds the root of the git repository containing dir
func findGitRoot(dir string) string {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}

	current := absDir

	for {
		gitDir := filepath.Join(current, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return current
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}

		current = parent
	}

	return ""
}

// loadIgnoreFilesFromHierarchy loads .gitignore and .ignore files from dir up to root
func loadIgnoreFilesFromHierarchy(gs *GitignoreSet, dir string) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return
	}

	// Collect directories from root to dir (we want root patterns first)
	var dirs []string

	current := absDir

	for {
		dirs = append([]string{current}, dirs...)

		parent := filepath.Dir(current)
		if parent == current {
			break
		}

		current = parent
	}

	// Load patterns from root to dir (later files override earlier)
	for _, d := range dirs {
		// Load .gitignore
		gitignorePath := filepath.Join(d, ".gitignore")
		if gi := loadGitignoreFile(gitignorePath, d); gi != nil {
			gs.AddGitignore(gi)
		}

		// Load .ignore (ripgrep-specific)
		ignorePath := filepath.Join(d, ".ignore")
		if gi := loadGitignoreFile(ignorePath, d); gi != nil {
			gs.AddGitignore(gi)
		}
	}
}

// loadGitignoreFile loads patterns from a single ignore file
func loadGitignoreFile(path, basePath string) *Gitignore {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	return pkgrg.ParseGitignore(string(data), basePath)
}

// parsePattern delegates to pkg/search/rg
func parsePattern(line string) *Pattern {
	return pkgrg.ParsePattern(line)
}

// patternToRegex delegates to pkg/search/rg
func patternToRegex(pattern string) string {
	return pkgrg.PatternToRegex(pattern)
}

