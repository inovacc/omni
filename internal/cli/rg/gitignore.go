package rg

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// MatchResult indicates the result of matching a path against gitignore patterns
type MatchResult int

const (
	// NoMatch means no pattern matched
	NoMatch MatchResult = iota
	// Ignore means the path should be ignored
	Ignore
	// Include means the path was re-included via negation
	Include
)

// Pattern represents a single gitignore pattern
type Pattern struct {
	Original   string         // Original pattern string
	Regex      *regexp.Regexp // Compiled regex
	Negation   bool           // starts with !
	DirOnly    bool           // ends with /
	Anchored   bool           // starts with / or contains /
	DoubleGlob bool           // contains **
}

// Gitignore represents a collection of gitignore patterns from a specific location
type Gitignore struct {
	patterns []Pattern
	basePath string // Directory containing the gitignore file
}

// GitignoreSet represents a collection of gitignore files
type GitignoreSet struct {
	gitignores []*Gitignore
	basePath   string // Base path for relative path matching
}

// NewGitignoreSet creates a new GitignoreSet loading patterns from multiple sources
func NewGitignoreSet(searchDir string) *GitignoreSet {
	absSearchDir, _ := filepath.Abs(searchDir)
	gs := &GitignoreSet{
		gitignores: make([]*Gitignore, 0),
		basePath:   absSearchDir,
	}

	// 1. Load global gitignore (~/.config/git/ignore)
	if globalPath := getGlobalGitignorePath(); globalPath != "" {
		if gi := loadGitignoreFile(globalPath, ""); gi != nil {
			gs.gitignores = append(gs.gitignores, gi)
		}
	}

	// 2. Find git root and load .git/info/exclude
	if gitRoot := findGitRoot(searchDir); gitRoot != "" {
		excludePath := filepath.Join(gitRoot, ".git", "info", "exclude")
		if gi := loadGitignoreFile(excludePath, gitRoot); gi != nil {
			gs.gitignores = append(gs.gitignores, gi)
		}
	}

	// 3. Load .gitignore and .ignore files walking up from searchDir
	gs.loadIgnoreFilesFromHierarchy(searchDir)

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
func (gs *GitignoreSet) loadIgnoreFilesFromHierarchy(dir string) {
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
			gs.gitignores = append(gs.gitignores, gi)
		}

		// Load .ignore (ripgrep-specific)
		ignorePath := filepath.Join(d, ".ignore")
		if gi := loadGitignoreFile(ignorePath, d); gi != nil {
			gs.gitignores = append(gs.gitignores, gi)
		}
	}
}

// loadGitignoreFile loads patterns from a single ignore file
func loadGitignoreFile(path, basePath string) *Gitignore {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	gi := &Gitignore{
		patterns: make([]Pattern, 0),
		basePath: basePath,
	}

	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimRight(line, "\r")

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle escaped # at start
		if strings.HasPrefix(line, "\\#") {
			line = line[1:]
		}

		pattern := parsePattern(line)
		if pattern != nil {
			gi.patterns = append(gi.patterns, *pattern)
		}
	}

	if len(gi.patterns) == 0 {
		return nil
	}

	return gi
}

// parsePattern parses a gitignore pattern string into a Pattern
func parsePattern(line string) *Pattern {
	p := &Pattern{
		Original: line,
	}

	pattern := line

	// Handle negation
	if strings.HasPrefix(pattern, "!") {
		p.Negation = true
		pattern = pattern[1:]
	}

	// Handle escaped !
	if strings.HasPrefix(pattern, "\\!") {
		pattern = pattern[1:]
	}

	// Handle directory-only patterns
	if strings.HasSuffix(pattern, "/") {
		p.DirOnly = true
		pattern = strings.TrimSuffix(pattern, "/")
	}

	// Handle anchored patterns
	if strings.HasPrefix(pattern, "/") {
		p.Anchored = true
		pattern = strings.TrimPrefix(pattern, "/")
	} else if strings.Contains(pattern, "/") {
		// Patterns with / in the middle are also anchored
		p.Anchored = true
	}

	// Check for double glob
	p.DoubleGlob = strings.Contains(pattern, "**")

	// Convert to regex
	regex := patternToRegex(pattern)

	re, err := regexp.Compile(regex)
	if err != nil {
		return nil
	}

	p.Regex = re

	return p
}

// patternToRegex converts a gitignore glob pattern to a regex
func patternToRegex(pattern string) string {
	var result strings.Builder

	result.WriteString("^")

	i := 0

	for i < len(pattern) {
		ch := pattern[i]

		switch ch {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				// ** matches any path segments
				if i+2 < len(pattern) && pattern[i+2] == '/' {
					// **/ at start or middle
					result.WriteString("(?:.*(?:/|$))?")

					i += 3

					continue
				} else if i == 0 || (i > 0 && pattern[i-1] == '/') {
					// ** at end or as entire segment
					result.WriteString(".*")

					i += 2

					continue
				}
			}
			// Single * - match anything except /
			result.WriteString("[^/]*")
		case '?':
			// Match any single character except /
			result.WriteString("[^/]")
		case '[':
			// Character class
			j := i + 1

			if j < len(pattern) && pattern[j] == '!' {
				j++
			}

			if j < len(pattern) && pattern[j] == ']' {
				j++
			}

			for j < len(pattern) && pattern[j] != ']' {
				j++
			}

			if j >= len(pattern) {
				// No closing bracket, treat literally
				result.WriteString(regexp.QuoteMeta("["))
			} else {
				// Copy character class, converting ! to ^
				class := pattern[i+1 : j]
				if strings.HasPrefix(class, "!") {
					class = "^" + class[1:]
				}

				result.WriteString("[")
				result.WriteString(class)
				result.WriteString("]")

				i = j
			}
		case '.', '+', '(', ')', '{', '}', '^', '$', '|':
			// Escape regex metacharacters
			result.WriteString("\\")
			result.WriteByte(ch)
		case '\\':
			// Escape next character
			if i+1 < len(pattern) {
				i++
				result.WriteString(regexp.QuoteMeta(string(pattern[i])))
			}
		default:
			result.WriteByte(ch)
		}

		i++
	}

	result.WriteString("$")

	return result.String()
}

// Match checks if a path should be ignored
// Returns Ignore, Include (negation), or NoMatch
func (gs *GitignoreSet) Match(path string, isDir bool) MatchResult {
	// Make path relative to the base path if it's absolute
	absPath, err := filepath.Abs(path)
	if err == nil && gs.basePath != "" {
		if rel, err := filepath.Rel(gs.basePath, absPath); err == nil {
			path = rel
		}
	}

	// Normalize path separators
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, "./")

	result := NoMatch

	// Check all gitignore files - later ones override earlier
	for _, gi := range gs.gitignores {
		if r := gi.match(path, isDir); r != NoMatch {
			result = r
		}
	}

	return result
}

// match checks a path against patterns in a single gitignore file
func (gi *Gitignore) match(path string, isDir bool) MatchResult {
	result := NoMatch

	// Get the path relative to the gitignore base
	relPath := path
	if gi.basePath != "" {
		absPath, err := filepath.Abs(path)
		if err == nil {
			absBase, err := filepath.Abs(gi.basePath)
			if err == nil {
				if rel, err := filepath.Rel(absBase, absPath); err == nil {
					relPath = filepath.ToSlash(rel)
				}
			}
		}
	}

	// Check each pattern - last match wins
	for _, p := range gi.patterns {
		if p.matchPath(relPath, path, isDir) {
			if p.Negation {
				result = Include
			} else {
				result = Ignore
			}
		}
	}

	return result
}

// matchPath checks if a single pattern matches the path
func (p *Pattern) matchPath(relPath, fullPath string, isDir bool) bool {
	// Directory-only patterns only match directories
	if p.DirOnly && !isDir {
		return false
	}

	// Try matching the full relative path
	if p.Regex.MatchString(relPath) {
		return true
	}

	// For non-anchored patterns, also try matching just the basename
	if !p.Anchored {
		base := filepath.Base(fullPath)
		if p.Regex.MatchString(base) {
			return true
		}
	}

	// For patterns without /, also try matching any path component
	if !p.Anchored && !strings.Contains(p.Original, "/") {
		parts := strings.Split(relPath, "/")
		if slices.ContainsFunc(parts, func(part string) bool {
			return p.Regex.MatchString(part)
		}) {
			return true
		}
	}

	return false
}

// ShouldIgnore is a convenience method that returns true if the path should be ignored
func (gs *GitignoreSet) ShouldIgnore(path string, isDir bool) bool {
	return gs.Match(path, isDir) == Ignore
}

// AddCommonIgnores adds common patterns that should always be ignored
func (gs *GitignoreSet) AddCommonIgnores() {
	commonPatterns := []string{
		".git",
		"node_modules",
		"__pycache__",
		".idea",
		".vscode",
	}

	gi := &Gitignore{
		patterns: make([]Pattern, 0, len(commonPatterns)),
		basePath: gs.basePath,
	}

	for _, pat := range commonPatterns {
		if p := parsePattern(pat); p != nil {
			gi.patterns = append(gi.patterns, *p)
		}
	}

	// Prepend so actual gitignore files can override
	gs.gitignores = append([]*Gitignore{gi}, gs.gitignores...)
}
