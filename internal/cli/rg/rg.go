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
	"runtime"
	"strings"
	"sync"
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
	JSONStream     bool     // --json-stream: streaming NDJSON output
	NoHeading      bool     // --no-heading: no file name headings
	OnlyMatching   bool     // -o: only show matching part
	Quiet          bool     // -q: quiet mode, exit on first match
	Fixed          bool     // -F: treat pattern as literal string
	Threads        int      // --threads: number of worker threads (0 = auto)

	// New options for ripgrep compatibility
	Color      string   // --color: when to use colors (auto, always, never)
	Colors     []string // --colors: custom color specifications
	Replace    string   // -r/--replace: replacement string for matches
	Multiline  bool     // -U/--multiline: enable multiline matching
	Trim       bool     // --trim: trim leading/trailing whitespace
	ShowColumn bool     // --column: show column numbers
	ByteOffset bool     // -b/--byte-offset: show byte offset (not implemented)
	Stats      bool     // --stats: show search statistics
	Passthru   bool     // --passthru: show all lines, highlighting matches
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

// resultInternal is used during parallel search with mutex protection
type resultInternal struct {
	Result

	mu sync.Mutex
}

// StreamMessage represents a message in NDJSON streaming output
type StreamMessage struct {
	Type string `json:"type"` // "begin", "match", "context", "end", "summary"
	Data any    `json:"data"`
}

// StreamBegin is sent at the start of searching a file
type StreamBegin struct {
	Path string `json:"path"`
}

// StreamLines holds line text for streaming output
type StreamLines struct {
	Text string `json:"text"`
}

// StreamMatch is sent for each match
type StreamMatch struct {
	Path       string      `json:"path"`
	LineNumber int         `json:"line_number"`
	Column     int         `json:"column,omitempty"`
	Lines      StreamLines `json:"lines"`
	Match      string      `json:"match,omitempty"`
}

// StreamContext is sent for context lines
type StreamContext struct {
	Path       string      `json:"path"`
	LineNumber int         `json:"line_number"`
	Lines      StreamLines `json:"lines"`
}

// StreamEnd is sent at the end of searching a file
type StreamEnd struct {
	Path       string `json:"path"`
	MatchCount int    `json:"match_count"`
}

// StreamSummary is sent at the end of all searching
type StreamSummary struct {
	TotalFiles   int `json:"total_files"`
	TotalMatches int `json:"total_matches"`
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

	// For literal/fixed patterns without regex features, we can use a fast path
	useLiteralSearch := opts.Fixed && !opts.WordRegexp && !opts.InvertMatch

	// Build regex pattern (needed even for literal if we need to highlight matches)
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

	// Prepare literal pattern for fast search
	literalPattern := pattern
	if opts.IgnoreCase || (opts.SmartCase && pattern == strings.ToLower(pattern)) {
		literalPattern = strings.ToLower(pattern)
	}

	result := &resultInternal{
		Result: Result{
			Files: make([]FileResult, 0),
		},
	}

	// Determine number of workers
	numWorkers := opts.Threads
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}

	// Create streaming encoder if needed
	var streamEnc *json.Encoder

	var streamMu sync.Mutex

	if opts.JSONStream {
		streamEnc = json.NewEncoder(w)
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

		// Load gitignore patterns for this specific path
		var gitignore *GitignoreSet

		if !opts.NoIgnore {
			searchDir := path
			if !info.IsDir() {
				searchDir = filepath.Dir(path)
			}

			gitignore = NewGitignoreSet(searchDir)
			gitignore.AddCommonIgnores()
		}

		if info.IsDir() {
			if numWorkers > 1 {
				err = searchDirParallel(w, path, re, pattern, literalPattern, useLiteralSearch, opts, gitignore, result, numWorkers, streamEnc, &streamMu)
			} else {
				err = searchDir(w, path, re, pattern, literalPattern, useLiteralSearch, opts, gitignore, result, 0, streamEnc, &streamMu)
			}
		} else {
			err = searchFile(w, path, re, pattern, literalPattern, useLiteralSearch, opts, result, streamEnc, &streamMu)
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

	// Output results
	if opts.JSONStream {
		// Write summary
		streamMu.Lock()
		//nolint:errchkjson // StreamSummary is a concrete type, not any
		_ = streamEnc.Encode(StreamMessage{
			Type: "summary",
			Data: StreamSummary{
				TotalFiles:   result.TotalFiles,
				TotalMatches: result.TotalMatch,
			},
		})

		streamMu.Unlock()
	} else if opts.JSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		return enc.Encode(result.Result)
	}

	return nil
}

// searchDirParallel performs parallel directory traversal and search
func searchDirParallel(w io.Writer, dir string, re *regexp.Regexp, pattern, literalPattern string, useLiteral bool, opts Options, gitignore *GitignoreSet, result *resultInternal, numWorkers int, streamEnc *json.Encoder, streamMu *sync.Mutex) error {
	// Collect all files to search
	var files []string

	err := collectFiles(dir, opts, gitignore, &files, 0)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	// Create work channel and result channel
	fileCh := make(chan string, numWorkers*2)
	resultCh := make(chan FileResult, numWorkers*2)
	errCh := make(chan error, numWorkers)

	// Start workers
	var wg sync.WaitGroup

	for range numWorkers {
		wg.Go(func() {
			for path := range fileCh {
				fr, err := searchFileSingle(path, re, pattern, literalPattern, useLiteral, opts)
				if err != nil {
					select {
					case errCh <- fmt.Errorf("%s: %w", path, err):
					default:
						// Drop error if channel is full
					}

					continue
				}

				if fr != nil && fr.Count > 0 {
					resultCh <- *fr
				}
			}
		})
	}

	// Start result collector goroutine
	var collectorWg sync.WaitGroup

	collectorWg.Go(func() {
		for fr := range resultCh {
			result.mu.Lock()
			result.Files = append(result.Files, fr)
			result.TotalFiles++
			result.TotalMatch += fr.Count
			result.mu.Unlock()

			// Output results
			if !opts.JSON && !opts.JSONStream {
				outputFileResult(w, fr, opts, re, pattern, useLiteral)
			} else if opts.JSONStream && streamEnc != nil {
				streamMu.Lock()

				_ = streamEnc.Encode(StreamMessage{Type: "begin", Data: StreamBegin{Path: fr.Path}})

				for _, m := range fr.Matches {
					_ = streamEnc.Encode(StreamMessage{
						Type: "match",
						Data: StreamMatch{
							Path:       m.Path,
							LineNumber: m.LineNumber,
							Column:     m.Column,
							Lines:      StreamLines{Text: m.Line},
							Match:      m.Match,
						},
					})
				}

				_ = streamEnc.Encode(StreamMessage{Type: "end", Data: StreamEnd{Path: fr.Path, MatchCount: fr.Count}})

				streamMu.Unlock()
			}
		}
	})

	// Send files to workers
	for _, f := range files {
		fileCh <- f
	}

	close(fileCh)

	// Wait for workers to finish
	wg.Wait()
	close(resultCh)

	// Wait for collector to finish
	collectorWg.Wait()
	close(errCh)

	// Report errors (non-fatal)
	if !opts.Quiet {
		for err := range errCh {
			_, _ = fmt.Fprintf(w, "rg: %v\n", err)
		}
	}

	return nil
}

// collectFiles recursively collects all searchable files
func collectFiles(dir string, opts Options, gitignore *GitignoreSet, files *[]string, depth int) error {
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
		if gitignore != nil && gitignore.ShouldIgnore(path, entry.IsDir()) {
			continue
		}

		if entry.IsDir() {
			if opts.FollowSymlinks || entry.Type()&os.ModeSymlink == 0 {
				_ = collectFiles(path, opts, gitignore, files, depth+1)
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

		*files = append(*files, path)
	}

	return nil
}

// errSkipBinary signals that a file was skipped because it's binary
var errSkipBinary = fmt.Errorf("binary file skipped")

// searchFileSingle searches a single file and returns results (used by parallel search)
func searchFileSingle(path string, re *regexp.Regexp, pattern, literalPattern string, useLiteral bool, opts Options) (*FileResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() { _ = file.Close() }()

	// Check if binary
	buf := make([]byte, 512)

	n, _ := file.Read(buf)
	if n > 0 && isBinary(buf[:n]) {
		return nil, errSkipBinary
	}

	// Reset file position
	if _, err := file.Seek(0, 0); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)

	var (
		lineNum    int
		matches    []Match
		matchCount int
	)

	caseInsensitive := opts.IgnoreCase || (opts.SmartCase && pattern == strings.ToLower(pattern))

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check for match using appropriate method
		var found bool

		var matchStart int

		if useLiteral {
			// Fast literal search
			searchLine := line
			if caseInsensitive {
				searchLine = strings.ToLower(line)
			}

			matchStart = strings.Index(searchLine, literalPattern)
			found = matchStart >= 0
		} else {
			// Regex search
			loc := re.FindStringIndex(line)
			found = loc != nil

			if found {
				matchStart = loc[0]
			}
		}

		if opts.InvertMatch {
			found = !found
		}

		if found {
			matchCount++

			if opts.MaxCount > 0 && matchCount > opts.MaxCount {
				break
			}

			if !opts.Count && !opts.FilesWithMatch && !opts.Quiet {
				match := Match{
					Path:       path,
					LineNumber: lineNum,
					Column:     matchStart + 1, // 1-indexed
					Line:       line,
				}

				if opts.OnlyMatching {
					if useLiteral {
						match.Match = line[matchStart : matchStart+len(literalPattern)]
					} else {
						match.Match = re.FindString(line)
					}
				}

				matches = append(matches, match)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if matchCount == 0 {
		return nil, nil //nolint:nilnil // nil result with nil error means no matches
	}

	return &FileResult{
		Path:    path,
		Matches: matches,
		Count:   matchCount,
	}, nil
}

// outputFileResult outputs results for a single file
func outputFileResult(w io.Writer, fr FileResult, opts Options, re *regexp.Regexp, pattern string, useLiteral bool) {
	if opts.FilesWithMatch {
		colorMode := ParseColorMode(opts.Color)
		useColor := ShouldUseColor(colorMode)
		scheme := DefaultScheme()
		_, _ = fmt.Fprintln(w, FormatPath(fr.Path, scheme, useColor))

		return
	}

	if opts.Count {
		colorMode := ParseColorMode(opts.Color)
		useColor := ShouldUseColor(colorMode)
		scheme := DefaultScheme()
		_, _ = fmt.Fprintf(w, "%s%s%d\n",
			FormatPath(fr.Path, scheme, useColor),
			FormatSeparator(":", scheme, useColor),
			fr.Count)

		return
	}

	for _, m := range fr.Matches {
		printLineWithColor(w, m.Path, m.LineNumber, m.Column, m.Line, opts, false, re, pattern, useLiteral)
	}
}

func searchDir(w io.Writer, dir string, re *regexp.Regexp, pattern, literalPattern string, useLiteral bool, opts Options, gitignore *GitignoreSet, result *resultInternal, depth int, streamEnc *json.Encoder, streamMu *sync.Mutex) error {
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
		if gitignore != nil && gitignore.ShouldIgnore(path, entry.IsDir()) {
			continue
		}

		if entry.IsDir() {
			if opts.FollowSymlinks || entry.Type()&os.ModeSymlink == 0 {
				if err := searchDir(w, path, re, pattern, literalPattern, useLiteral, opts, gitignore, result, depth+1, streamEnc, streamMu); err != nil {
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

		if err := searchFile(w, path, re, pattern, literalPattern, useLiteral, opts, result, streamEnc, streamMu); err != nil {
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

func searchFile(w io.Writer, path string, re *regexp.Regexp, pattern, literalPattern string, useLiteral bool, opts Options, result *resultInternal, streamEnc *json.Encoder, streamMu *sync.Mutex) error {
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

	type contextLine struct {
		lineNum int
		text    string
	}

	var (
		lineNum         int
		matches         []Match
		matchCount      int
		beforeLines     []contextLine
		afterNeeded     int
		lastPrintedLine int
	)

	beforeContext := opts.Before

	afterContext := opts.After
	if opts.Context > 0 {
		beforeContext = opts.Context
		afterContext = opts.Context
	}

	needsContext := beforeContext > 0 || afterContext > 0
	caseInsensitive := opts.IgnoreCase || (opts.SmartCase && pattern == strings.ToLower(pattern))

	// Stream begin message
	if opts.JSONStream && streamEnc != nil {
		streamMu.Lock()
		//nolint:errchkjson // StreamBegin is a concrete type
		_ = streamEnc.Encode(StreamMessage{Type: "begin", Data: StreamBegin{Path: path}})

		streamMu.Unlock()
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check for match using appropriate method
		var found bool

		var matchStart int

		if useLiteral {
			// Fast literal search
			searchLine := line
			if caseInsensitive {
				searchLine = strings.ToLower(line)
			}

			matchStart = strings.Index(searchLine, literalPattern)
			found = matchStart >= 0
		} else {
			// Regex search
			loc := re.FindStringIndex(line)
			found = loc != nil

			if found {
				matchStart = loc[0]
			}
		}

		if opts.InvertMatch {
			found = !found
		}

		if found {
			matchCount++

			result.mu.Lock()
			result.TotalMatch++
			result.mu.Unlock()

			if opts.MaxCount > 0 && matchCount > opts.MaxCount {
				break
			}

			if !opts.Count && !opts.FilesWithMatch && !opts.Quiet {
				match := Match{
					Path:       path,
					LineNumber: lineNum,
					Column:     matchStart + 1,
					Line:       line,
				}

				if opts.OnlyMatching {
					if useLiteral {
						match.Match = line[matchStart : matchStart+len(literalPattern)]
					} else {
						match.Match = re.FindString(line)
					}
				}

				matches = append(matches, match)

				// Stream output
				if opts.JSONStream && streamEnc != nil {
					streamMu.Lock()
					//nolint:errchkjson // StreamMatch is a concrete type
					_ = streamEnc.Encode(StreamMessage{
						Type: "match",
						Data: StreamMatch{
							Path:       match.Path,
							LineNumber: match.LineNumber,
							Column:     match.Column,
							Lines:      StreamLines{Text: match.Line},
							Match:      match.Match,
						},
					})

					streamMu.Unlock()
				}

				// Print context before
				if !opts.JSON && !opts.JSONStream && beforeContext > 0 && len(beforeLines) > 0 {
					// Print separator if there's a gap
					firstBeforeLine := beforeLines[0].lineNum
					if needsContext && lastPrintedLine > 0 && firstBeforeLine > lastPrintedLine+1 {
						printContextSeparator(w, opts)
					}

					for _, bl := range beforeLines {
						printLineWithColor(w, path, bl.lineNum, 0, bl.text, opts, true, re, pattern, useLiteral)
						lastPrintedLine = bl.lineNum
					}

					beforeLines = nil
				}

				// Print separator if there's a gap and no before context was printed
				if !opts.JSON && !opts.JSONStream && needsContext && lastPrintedLine > 0 && lineNum > lastPrintedLine+1 {
					printContextSeparator(w, opts)
				}

				if !opts.JSON && !opts.JSONStream {
					printLineWithColor(w, path, lineNum, matchStart+1, line, opts, false, re, pattern, useLiteral)
					lastPrintedLine = lineNum
				}

				afterNeeded = afterContext
			}
		} else {
			// Handle after context
			if afterNeeded > 0 && !opts.JSON && !opts.JSONStream && !opts.Count && !opts.FilesWithMatch {
				printLineWithColor(w, path, lineNum, 0, line, opts, true, re, pattern, useLiteral)
				lastPrintedLine = lineNum

				afterNeeded--
			}

			// Store for before context
			if beforeContext > 0 {
				beforeLines = append(beforeLines, contextLine{lineNum: lineNum, text: line})
				if len(beforeLines) > beforeContext {
					beforeLines = beforeLines[1:]
				}
			}
		}
	}

	if matchCount > 0 {
		result.mu.Lock()
		result.TotalFiles++

		fileResult := FileResult{
			Path:    path,
			Matches: matches,
			Count:   matchCount,
		}
		result.Files = append(result.Files, fileResult)
		result.mu.Unlock()

		if opts.FilesWithMatch && !opts.JSON && !opts.JSONStream {
			colorMode := ParseColorMode(opts.Color)
			useColor := ShouldUseColor(colorMode)
			scheme := DefaultScheme()
			_, _ = fmt.Fprintln(w, FormatPath(path, scheme, useColor))
		}

		if opts.Count && !opts.JSON && !opts.JSONStream {
			colorMode := ParseColorMode(opts.Color)
			useColor := ShouldUseColor(colorMode)
			scheme := DefaultScheme()
			_, _ = fmt.Fprintf(w, "%s%s%d\n",
				FormatPath(path, scheme, useColor),
				FormatSeparator(":", scheme, useColor),
				matchCount)
		}
	}

	// Stream end message
	if opts.JSONStream && streamEnc != nil {
		streamMu.Lock()
		//nolint:errchkjson // StreamEnd is a concrete type
		_ = streamEnc.Encode(StreamMessage{Type: "end", Data: StreamEnd{Path: path, MatchCount: matchCount}})

		streamMu.Unlock()
	}

	return scanner.Err()
}

func printContextSeparator(w io.Writer, opts Options) {
	colorMode := ParseColorMode(opts.Color)
	useColor := ShouldUseColor(colorMode)
	scheme := DefaultScheme()
	_, _ = fmt.Fprintln(w, FormatSeparator("--", scheme, useColor))
}

func printLineWithColor(w io.Writer, path string, lineNum, column int, line string, opts Options, isContext bool, re *regexp.Regexp, pattern string, useLiteral bool) {
	sep := ":"
	if isContext {
		sep = "-"
	}

	// Determine if we should use colors
	colorMode := ParseColorMode(opts.Color)
	useColor := ShouldUseColor(colorMode)
	scheme := DefaultScheme()

	// Apply custom color specs
	for _, spec := range opts.Colors {
		_ = ApplyColorSpec(&scheme, spec)
	}

	// Handle trim
	if opts.Trim {
		line = strings.TrimSpace(line)
	}

	// Handle replacement
	if opts.Replace != "" && !isContext && re != nil {
		line = re.ReplaceAllString(line, opts.Replace)
	}

	// Highlight matches
	highlightedLine := line

	if useColor && !isContext {
		caseInsensitive := opts.IgnoreCase || (opts.SmartCase && pattern == strings.ToLower(pattern))
		if useLiteral {
			highlightedLine = HighlightLiteralMatches(line, pattern, caseInsensitive, scheme, useColor)
		} else if re != nil {
			highlightedLine = HighlightMatches(line, re, scheme, useColor)
		}
	}

	// Build output
	if opts.NoHeading {
		pathStr := path
		if useColor {
			pathStr = FormatPath(path, scheme, useColor)
		}

		sepStr := sep
		if useColor {
			sepStr = FormatSeparator(sep, scheme, useColor)
		}

		if opts.LineNumber && lineNum > 0 {
			lineNumStr := fmt.Sprintf("%d", lineNum)
			if useColor {
				lineNumStr = FormatLineNumber(lineNum, scheme, useColor)
			}

			if opts.ShowColumn && column > 0 {
				colStr := fmt.Sprintf("%d", column)
				if useColor {
					colStr = FormatColumn(column, scheme, useColor)
				}

				_, _ = fmt.Fprintf(w, "%s%s%s%s%s%s%s\n", pathStr, sepStr, lineNumStr, sepStr, colStr, sepStr, highlightedLine)
			} else {
				_, _ = fmt.Fprintf(w, "%s%s%s%s%s\n", pathStr, sepStr, lineNumStr, sepStr, highlightedLine)
			}
		} else {
			_, _ = fmt.Fprintf(w, "%s%s%s\n", pathStr, sepStr, highlightedLine)
		}
	} else {
		sepStr := sep
		if useColor {
			sepStr = FormatSeparator(sep, scheme, useColor)
		}

		if opts.LineNumber && lineNum > 0 {
			lineNumStr := fmt.Sprintf("%d", lineNum)
			if useColor {
				lineNumStr = FormatLineNumber(lineNum, scheme, useColor)
			}

			if opts.ShowColumn && column > 0 {
				colStr := fmt.Sprintf("%d", column)
				if useColor {
					colStr = FormatColumn(column, scheme, useColor)
				}

				_, _ = fmt.Fprintf(w, "%s%s%s%s%s\n", lineNumStr, sepStr, colStr, sepStr, highlightedLine)
			} else {
				_, _ = fmt.Fprintf(w, "%s%s%s\n", lineNumStr, sepStr, highlightedLine)
			}
		} else {
			_, _ = fmt.Fprintln(w, highlightedLine)
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

func isBinary(data []byte) bool {
	return bytes.Contains(data, []byte{0})
}
