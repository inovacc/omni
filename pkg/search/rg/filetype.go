package rg

import (
	"bytes"
	"path/filepath"
	"strings"
)

// FileTypeExtensions maps file type names to extensions
var FileTypeExtensions = map[string][]string{
	"go":         {".go"},
	"js":         {".js", ".mjs", ".cjs"},
	"ts":         {".ts", ".tsx", ".mts", ".cts"},
	"py":         {".py", ".pyi", ".pyw"},
	"rust":       {".rs"},
	"c":          {".c", ".h"},
	"cpp":        {".cpp", ".cc", ".cxx", ".hpp", ".hh", ".hxx", ".c++", ".h++"},
	"java":       {".java"},
	"rb":         {".rb", ".rake", ".gemspec"},
	"php":        {".php", ".phtml", ".php3", ".php4", ".php5"},
	"sh":         {".sh", ".bash", ".zsh", ".fish"},
	"json":       {".json", ".jsonl"},
	"yaml":       {".yaml", ".yml"},
	"toml":       {".toml"},
	"xml":        {".xml", ".xsd", ".xsl", ".xslt"},
	"html":       {".html", ".htm", ".xhtml"},
	"css":        {".css", ".scss", ".sass", ".less"},
	"md":         {".md", ".markdown"},
	"sql":        {".sql"},
	"proto":      {".proto"},
	"dockerfile": {"Dockerfile", ".dockerfile"},
	"make":       {"Makefile", "GNUmakefile", "makefile", ".mk"},
	"txt":        {".txt"},
}

// MatchesFileType checks if a file path matches the given include/exclude type filters.
func MatchesFileType(path string, include, exclude []string) bool {
	if len(include) == 0 && len(exclude) == 0 {
		return true
	}

	ext := strings.ToLower(filepath.Ext(path))
	base := filepath.Base(path)

	// Check exclusions first
	for _, t := range exclude {
		if exts, ok := FileTypeExtensions[t]; ok {
			for _, e := range exts {
				if ext == e || base == e {
					return false
				}
			}
		}
	}

	// If no includes specified, accept all (that weren't excluded)
	if len(include) == 0 {
		return true
	}

	// Check inclusions
	for _, t := range include {
		if exts, ok := FileTypeExtensions[t]; ok {
			for _, e := range exts {
				if ext == e || base == e {
					return true
				}
			}
		}
	}

	return false
}

// MatchesGlob checks if a path matches the given glob patterns.
// Supports negation patterns prefixed with "!".
func MatchesGlob(path string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}

	// Check if all patterns are negations
	allNegations := true

	for _, pattern := range patterns {
		if !strings.HasPrefix(pattern, "!") {
			allNegations = false

			break
		}
	}

	for _, pattern := range patterns {
		negate := strings.HasPrefix(pattern, "!")
		if negate {
			pattern = pattern[1:]
		}

		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if !matched {
			// Also try matching full path
			matched, _ = filepath.Match(pattern, path)
		}

		if matched {
			if negate {
				return false // Explicitly excluded
			}

			return true // Explicitly included
		}
	}

	// If no pattern matched and all were negations, include the file
	// If no pattern matched and some were inclusions, exclude the file
	return allNegations
}

// IsBinary checks if data contains null bytes, indicating it's binary content.
func IsBinary(data []byte) bool {
	return bytes.Contains(data, []byte{0})
}
