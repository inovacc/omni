package rg

import (
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
	Patterns []Pattern
	BasePath string // Directory containing the gitignore file
}

// GitignoreSet represents a collection of gitignore files
type GitignoreSet struct {
	Gitignores []*Gitignore
	BasePath   string // Base path for relative path matching
}

// NewGitignoreSet creates a new empty GitignoreSet for the given base directory
func NewGitignoreSet(basePath string) *GitignoreSet {
	absPath, _ := filepath.Abs(basePath)
	return &GitignoreSet{
		Gitignores: make([]*Gitignore, 0),
		BasePath:   absPath,
	}
}

// AddGitignore adds a Gitignore to the set
func (gs *GitignoreSet) AddGitignore(gi *Gitignore) {
	if gi != nil {
		gs.Gitignores = append(gs.Gitignores, gi)
	}
}

// PrependGitignore adds a Gitignore at the beginning of the set (lower priority)
func (gs *GitignoreSet) PrependGitignore(gi *Gitignore) {
	if gi != nil {
		gs.Gitignores = append([]*Gitignore{gi}, gs.Gitignores...)
	}
}

// ParseGitignore parses gitignore content into a Gitignore struct
func ParseGitignore(content, basePath string) *Gitignore {
	gi := &Gitignore{
		Patterns: make([]Pattern, 0),
		BasePath: basePath,
	}

	for line := range strings.SplitSeq(content, "\n") {
		line = strings.TrimRight(line, "\r")

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle escaped # at start
		if strings.HasPrefix(line, "\\#") {
			line = line[1:]
		}

		pattern := ParsePattern(line)
		if pattern != nil {
			gi.Patterns = append(gi.Patterns, *pattern)
		}
	}

	if len(gi.Patterns) == 0 {
		return nil
	}

	return gi
}

// ParsePattern parses a gitignore pattern string into a Pattern
func ParsePattern(line string) *Pattern {
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
	regex := PatternToRegex(pattern)

	re, err := regexp.Compile(regex)
	if err != nil {
		return nil
	}

	p.Regex = re

	return p
}

// PatternToRegex converts a gitignore glob pattern to a regex
func PatternToRegex(pattern string) string {
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

// Match checks if a path should be ignored.
// Returns Ignore, Include (negation), or NoMatch.
func (gs *GitignoreSet) Match(path string, isDir bool) MatchResult {
	// Make path relative to the base path if it's absolute
	absPath, err := filepath.Abs(path)
	if err == nil && gs.BasePath != "" {
		if rel, err := filepath.Rel(gs.BasePath, absPath); err == nil {
			path = rel
		}
	}

	// Normalize path separators
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, "./")

	result := NoMatch

	// Check all gitignore files - later ones override earlier
	for _, gi := range gs.Gitignores {
		if r := gi.Match(path, isDir); r != NoMatch {
			result = r
		}
	}

	return result
}

// Match checks a path against patterns in a single gitignore file
func (gi *Gitignore) Match(path string, isDir bool) MatchResult {
	result := NoMatch

	// Get the path relative to the gitignore base
	relPath := path
	if gi.BasePath != "" {
		absPath, err := filepath.Abs(path)
		if err == nil {
			absBase, err := filepath.Abs(gi.BasePath)
			if err == nil {
				if rel, err := filepath.Rel(absBase, absPath); err == nil {
					relPath = filepath.ToSlash(rel)
				}
			}
		}
	}

	// Check each pattern - last match wins
	for _, p := range gi.Patterns {
		if p.MatchPath(relPath, path, isDir) {
			if p.Negation {
				result = Include
			} else {
				result = Ignore
			}
		}
	}

	return result
}

// MatchPath checks if a single pattern matches the path
func (p *Pattern) MatchPath(relPath, fullPath string, isDir bool) bool {
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
		Patterns: make([]Pattern, 0, len(commonPatterns)),
		BasePath: gs.BasePath,
	}

	for _, pat := range commonPatterns {
		if p := ParsePattern(pat); p != nil {
			gi.Patterns = append(gi.Patterns, *p)
		}
	}

	// Prepend so actual gitignore files can override
	gs.PrependGitignore(gi)
}
