package rg

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Options configures the rg command behavior
type Options struct {
	IgnoreCase     bool     // -i: case insensitive search
	SmartCase      bool     // -S: smart case (case insensitive if pattern is lowercase)
	WordRegexp     bool     // -w: match whole words only
	LineNumber     bool     // -n: show line numbers (default true)
	Count          bool     // -c: only show count of matches
	FilesWithMatch bool     // -l: only show file names with matches
	InvertMatch    bool     // -v: show non-matching lines
	Context        int      // -C: lines of context (before and after)
	Before         int      // -B: lines before match
	After          int      // -A: lines after match
	Types          []string // -t: file types to include
	TypesNot       []string // -T: file types to exclude
	Glob           []string // -g: glob patterns to include
	Hidden         bool     // --hidden: search hidden files
	NoIgnore       bool     // --no-ignore: don't respect gitignore
	MaxCount       int      // -m: max matches per file
	MaxDepth       int      // --max-depth: max directory depth
	FollowSymlinks bool     // -L: follow symlinks
	JSON           bool     // --json: JSON output
	NoHeading      bool     // --no-heading: no file name headings
	OnlyMatching   bool     // -o: only show matching part
	Quiet          bool     // -q: quiet mode, exit on first match
	Fixed          bool     // -F: treat pattern as literal string
}

// Match represents a single match result
type Match struct {
	Path       string `json:"path"`
	LineNumber int    `json:"line_number"`
	Column     int    `json:"column,omitempty"`
	Line       string `json:"line"`
	Match      string `json:"match,omitempty"`
}

// FileResult represents matches in a single file
type FileResult struct {
	Path    string  `json:"path"`
	Matches []Match `json:"matches"`
	Count   int     `json:"count"`
}

// Result represents the complete search result
type Result struct {
	Files      []FileResult `json:"files"`
	TotalFiles int          `json:"total_files"`
	TotalMatch int          `json:"total_matches"`
}

// fileTypeExtensions maps file type names to extensions
var fileTypeExtensions = map[string][]string{
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

// Run executes the rg command
func Run(w io.Writer, pattern string, paths []string, opts Options) error {
	if pattern == "" {
		return fmt.Errorf("rg: no pattern provided")
	}

	if len(paths) == 0 {
		paths = []string{"."}
	}

	// Build regex pattern
	regexPattern := pattern
	if opts.Fixed {
		regexPattern = regexp.QuoteMeta(pattern)
	}

	if opts.WordRegexp {
		regexPattern = `\b` + regexPattern + `\b`
	}

	flags := ""
	if opts.IgnoreCase {
		flags = "(?i)"
	} else if opts.SmartCase && pattern == strings.ToLower(pattern) {
		flags = "(?i)"
	}

	re, err := regexp.Compile(flags + regexPattern)
	if err != nil {
		return fmt.Errorf("rg: invalid pattern: %w", err)
	}

	// Load gitignore patterns
	var gitignorePatterns []string
	if !opts.NoIgnore {
		gitignorePatterns = loadGitignore(".")
	}

	result := Result{
		Files: make([]FileResult, 0),
	}

	// Search each path
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			if !opts.Quiet {
				_, _ = fmt.Fprintf(w, "rg: %s: %v\n", path, err)
			}

			continue
		}

		if info.IsDir() {
			err = searchDir(w, path, re, opts, gitignorePatterns, &result, 0)
		} else {
			err = searchFile(w, path, re, opts, &result)
		}

		if err != nil {
			if !opts.Quiet {
				_, _ = fmt.Fprintf(w, "rg: %v\n", err)
			}
		}

		if opts.Quiet && result.TotalMatch > 0 {
			break
		}
	}

	// Output JSON if requested
	if opts.JSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		return enc.Encode(result)
	}

	return nil
}

func searchDir(w io.Writer, dir string, re *regexp.Regexp, opts Options, gitignore []string, result *Result, depth int) error {
	if opts.MaxDepth > 0 && depth > opts.MaxDepth {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		name := entry.Name()
		path := filepath.Join(dir, name)

		// Skip hidden files unless --hidden
		if !opts.Hidden && strings.HasPrefix(name, ".") {
			continue
		}

		// Check gitignore
		if !opts.NoIgnore && isIgnored(path, gitignore) {
			continue
		}

		if entry.IsDir() {
			if opts.FollowSymlinks || entry.Type()&os.ModeSymlink == 0 {
				if err := searchDir(w, path, re, opts, gitignore, result, depth+1); err != nil {
					if !opts.Quiet {
						_, _ = fmt.Fprintf(w, "rg: %s: %v\n", path, err)
					}
				}
			}

			continue
		}

		// Check file type filters
		if !matchesFileType(path, opts.Types, opts.TypesNot) {
			continue
		}

		// Check glob patterns
		if !matchesGlob(path, opts.Glob) {
			continue
		}

		if err := searchFile(w, path, re, opts, result); err != nil {
			if !opts.Quiet {
				_, _ = fmt.Fprintf(w, "rg: %s: %v\n", path, err)
			}
		}

		if opts.Quiet && result.TotalMatch > 0 {
			return nil
		}
	}

	return nil
}

func searchFile(w io.Writer, path string, re *regexp.Regexp, opts Options, result *Result) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	// Check if binary
	buf := make([]byte, 512)

	n, _ := file.Read(buf)
	if n > 0 && isBinary(buf[:n]) {
		return nil // Skip binary files
	}

	// Reset file position
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)

	var (
		lineNum     int
		matches     []Match
		matchCount  int
		beforeLines []string
		afterNeeded int
	)

	beforeContext := opts.Before

	afterContext := opts.After
	if opts.Context > 0 {
		beforeContext = opts.Context
		afterContext = opts.Context
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check for match
		found := re.MatchString(line)
		if opts.InvertMatch {
			found = !found
		}

		if found {
			matchCount++
			result.TotalMatch++

			if opts.MaxCount > 0 && matchCount > opts.MaxCount {
				break
			}

			if !opts.Count && !opts.FilesWithMatch && !opts.Quiet {
				match := Match{
					Path:       path,
					LineNumber: lineNum,
					Line:       line,
				}

				if opts.OnlyMatching {
					match.Match = re.FindString(line)
				}

				matches = append(matches, match)

				// Print context before
				if !opts.JSON && beforeContext > 0 {
					for _, bl := range beforeLines {
						printLine(w, path, 0, bl, opts, true)
					}

					beforeLines = nil
				}

				if !opts.JSON {
					printLine(w, path, lineNum, line, opts, false)
				}

				afterNeeded = afterContext
			}
		} else {
			// Handle context
			if afterNeeded > 0 && !opts.JSON && !opts.Count && !opts.FilesWithMatch {
				printLine(w, path, lineNum, line, opts, true)

				afterNeeded--
			}

			if beforeContext > 0 {
				beforeLines = append(beforeLines, line)
				if len(beforeLines) > beforeContext {
					beforeLines = beforeLines[1:]
				}
			}
		}
	}

	if matchCount > 0 {
		result.TotalFiles++

		fileResult := FileResult{
			Path:    path,
			Matches: matches,
			Count:   matchCount,
		}
		result.Files = append(result.Files, fileResult)

		if opts.FilesWithMatch && !opts.JSON {
			_, _ = fmt.Fprintln(w, path)
		}

		if opts.Count && !opts.JSON {
			_, _ = fmt.Fprintf(w, "%s:%d\n", path, matchCount)
		}
	}

	return scanner.Err()
}

func printLine(w io.Writer, path string, lineNum int, line string, opts Options, isContext bool) {
	sep := ":"
	if isContext {
		sep = "-"
	}

	if opts.NoHeading {
		if opts.LineNumber && lineNum > 0 {
			_, _ = fmt.Fprintf(w, "%s%s%d%s%s\n", path, sep, lineNum, sep, line)
		} else {
			_, _ = fmt.Fprintf(w, "%s%s%s\n", path, sep, line)
		}
	} else {
		if opts.LineNumber && lineNum > 0 {
			_, _ = fmt.Fprintf(w, "%d%s%s\n", lineNum, sep, line)
		} else {
			_, _ = fmt.Fprintln(w, line)
		}
	}
}

func matchesFileType(path string, include, exclude []string) bool {
	if len(include) == 0 && len(exclude) == 0 {
		return true
	}

	ext := strings.ToLower(filepath.Ext(path))
	base := filepath.Base(path)

	// Check exclusions first
	for _, t := range exclude {
		if exts, ok := fileTypeExtensions[t]; ok {
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
		if exts, ok := fileTypeExtensions[t]; ok {
			for _, e := range exts {
				if ext == e || base == e {
					return true
				}
			}
		}
	}

	return false
}

func matchesGlob(path string, patterns []string) bool {
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

func loadGitignore(dir string) []string {
	var patterns []string

	// Check for .gitignore in current and parent directories
	current := dir

	for {
		gitignorePath := filepath.Join(current, ".gitignore")

		if data, err := os.ReadFile(gitignorePath); err == nil {
			for line := range strings.SplitSeq(string(data), "\n") {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "#") {
					patterns = append(patterns, line)
				}
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}

		current = parent
	}

	// Add common ignored patterns
	patterns = append(patterns, ".git", "node_modules", "vendor", "__pycache__", ".idea", ".vscode")

	return patterns
}

func isIgnored(path string, patterns []string) bool {
	base := filepath.Base(path)

	for _, pattern := range patterns {
		if pattern == base {
			return true
		}

		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}

		// Check if any path component matches
		for part := range strings.SplitSeq(filepath.ToSlash(path), "/") {
			if part == pattern {
				return true
			}

			if matched, _ := filepath.Match(pattern, part); matched {
				return true
			}
		}
	}

	return false
}

func isBinary(data []byte) bool {
	return bytes.Contains(data, []byte{0})
}
