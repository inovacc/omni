package loc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Options configures the loc command behavior
type Options struct {
	Exclude []string // directories to exclude
	Hidden  bool     // include hidden files
	JSON    bool     // output as JSON
}

// Stats holds statistics for a single file
type Stats struct {
	Lines    int `json:"lines"`
	Code     int `json:"code"`
	Comments int `json:"comments"`
	Blanks   int `json:"blanks"`
}

// LanguageStats holds aggregated statistics for a language
type LanguageStats struct {
	Language string                   `json:"language"`
	Files    int                      `json:"files"`
	Lines    int                      `json:"lines"`
	Code     int                      `json:"code"`
	Comments int                      `json:"comments"`
	Blanks   int                      `json:"blanks"`
	Children map[string]*LanguageStats `json:"children,omitempty"` // Embedded languages (for Markdown)
}

// Result holds the complete LOC analysis result
type Result struct {
	Languages []LanguageStats `json:"languages"`
	Total     LanguageStats   `json:"total"`
}

// Language definition with string awareness
type langDef struct {
	name        string
	extensions  []string
	lineComment []string            // line comment starters (e.g., "//", "#")
	blockStart  string              // block comment start
	blockEnd    string              // block comment end
	quotes      [][2]string         // string quote pairs (start, end)
	literate    bool                // literate mode (like Markdown)
	nested      bool                // supports nested block comments
}

var languages = []langDef{
	{
		name:        "Go",
		extensions:  []string{".go"},
		lineComment: []string{"//"},
		blockStart:  "/*",
		blockEnd:    "*/",
		quotes:      [][2]string{{`"`, `"`}, {"`", "`"}},
	},
	{
		name:        "Rust",
		extensions:  []string{".rs"},
		lineComment: []string{"//"},
		blockStart:  "/*",
		blockEnd:    "*/",
		quotes:      [][2]string{{`"`, `"`}},
		nested:      true,
	},
	{
		name:        "JavaScript",
		extensions:  []string{".js", ".mjs", ".cjs", ".jsx"},
		lineComment: []string{"//"},
		blockStart:  "/*",
		blockEnd:    "*/",
		quotes:      [][2]string{{`"`, `"`}, {"'", "'"}, {"`", "`"}},
	},
	{
		name:        "TypeScript",
		extensions:  []string{".ts", ".tsx"},
		lineComment: []string{"//"},
		blockStart:  "/*",
		blockEnd:    "*/",
		quotes:      [][2]string{{`"`, `"`}, {"'", "'"}, {"`", "`"}},
	},
	{
		name:        "Python",
		extensions:  []string{".py", ".pyw"},
		lineComment: []string{"#"},
		blockStart:  `"""`,
		blockEnd:    `"""`,
		quotes:      [][2]string{{`"`, `"`}, {"'", "'"}},
	},
	{
		name:        "Java",
		extensions:  []string{".java"},
		lineComment: []string{"//"},
		blockStart:  "/*",
		blockEnd:    "*/",
		quotes:      [][2]string{{`"`, `"`}},
	},
	{
		name:        "C",
		extensions:  []string{".c", ".h"},
		lineComment: []string{"//"},
		blockStart:  "/*",
		blockEnd:    "*/",
		quotes:      [][2]string{{`"`, `"`}, {"'", "'"}},
	},
	{
		name:        "C++",
		extensions:  []string{".cpp", ".cc", ".cxx", ".hpp", ".hxx", ".hh"},
		lineComment: []string{"//"},
		blockStart:  "/*",
		blockEnd:    "*/",
		quotes:      [][2]string{{`"`, `"`}, {"'", "'"}},
	},
	{
		name:        "C#",
		extensions:  []string{".cs"},
		lineComment: []string{"//"},
		blockStart:  "/*",
		blockEnd:    "*/",
		quotes:      [][2]string{{`"`, `"`}, {"@\"", `"`}},
	},
	{
		name:        "Ruby",
		extensions:  []string{".rb"},
		lineComment: []string{"#"},
		blockStart:  "=begin",
		blockEnd:    "=end",
		quotes:      [][2]string{{`"`, `"`}, {"'", "'"}},
	},
	{
		name:        "PHP",
		extensions:  []string{".php"},
		lineComment: []string{"//", "#"},
		blockStart:  "/*",
		blockEnd:    "*/",
		quotes:      [][2]string{{`"`, `"`}, {"'", "'"}},
	},
	{
		name:        "Swift",
		extensions:  []string{".swift"},
		lineComment: []string{"//"},
		blockStart:  "/*",
		blockEnd:    "*/",
		quotes:      [][2]string{{`"`, `"`}},
		nested:      true,
	},
	{
		name:        "Kotlin",
		extensions:  []string{".kt", ".kts"},
		lineComment: []string{"//"},
		blockStart:  "/*",
		blockEnd:    "*/",
		quotes:      [][2]string{{`"`, `"`}, {`"""`, `"""`}},
		nested:      true,
	},
	{
		name:        "Scala",
		extensions:  []string{".scala"},
		lineComment: []string{"//"},
		blockStart:  "/*",
		blockEnd:    "*/",
		quotes:      [][2]string{{`"`, `"`}, {`"""`, `"""`}},
		nested:      true,
	},
	{
		name:        "Shell",
		extensions:  []string{".sh", ".bash", ".zsh"},
		lineComment: []string{"#"},
		quotes:      [][2]string{{`"`, `"`}, {"'", "'"}},
	},
	{
		name:        "PowerShell",
		extensions:  []string{".ps1", ".psm1"},
		lineComment: []string{"#"},
		blockStart:  "<#",
		blockEnd:    "#>",
		quotes:      [][2]string{{`"`, `"`}, {"'", "'"}},
	},
	{
		name:        "Lua",
		extensions:  []string{".lua"},
		lineComment: []string{"--"},
		blockStart:  "--[[",
		blockEnd:    "]]",
		quotes:      [][2]string{{`"`, `"`}, {"'", "'"}},
	},
	{
		name:        "Perl",
		extensions:  []string{".pl", ".pm"},
		lineComment: []string{"#"},
		blockStart:  "=pod",
		blockEnd:    "=cut",
		quotes:      [][2]string{{`"`, `"`}, {"'", "'"}},
	},
	{
		name:        "R",
		extensions:  []string{".r", ".R"},
		lineComment: []string{"#"},
		quotes:      [][2]string{{`"`, `"`}, {"'", "'"}},
	},
	{
		name:        "SQL",
		extensions:  []string{".sql"},
		lineComment: []string{"--"},
		blockStart:  "/*",
		blockEnd:    "*/",
		quotes:      [][2]string{{"'", "'"}},
	},
	{
		name:        "HTML",
		extensions:  []string{".html", ".htm"},
		blockStart:  "<!--",
		blockEnd:    "-->",
		quotes:      [][2]string{{`"`, `"`}, {"'", "'"}},
	},
	{
		name:        "CSS",
		extensions:  []string{".css"},
		blockStart:  "/*",
		blockEnd:    "*/",
		quotes:      [][2]string{{`"`, `"`}, {"'", "'"}},
	},
	{
		name:        "SCSS",
		extensions:  []string{".scss", ".sass"},
		lineComment: []string{"//"},
		blockStart:  "/*",
		blockEnd:    "*/",
		quotes:      [][2]string{{`"`, `"`}, {"'", "'"}},
	},
	{
		name:       "XML",
		extensions: []string{".xml", ".xsl", ".xslt", ".xsd", ".svg"},
		blockStart: "<!--",
		blockEnd:   "-->",
		quotes:     [][2]string{{`"`, `"`}, {"'", "'"}},
	},
	{
		name:       "JSON",
		extensions: []string{".json"},
		quotes:     [][2]string{{`"`, `"`}},
	},
	{
		name:        "YAML",
		extensions:  []string{".yaml", ".yml"},
		lineComment: []string{"#"},
		quotes:      [][2]string{{`"`, `"`}, {"'", "'"}},
	},
	{
		name:        "TOML",
		extensions:  []string{".toml"},
		lineComment: []string{"#"},
		quotes:      [][2]string{{`"`, `"`}, {"'", "'"}, {`"""`, `"""`}, {"'''", "'''"}},
	},
	{
		name:       "Markdown",
		extensions: []string{".md", ".markdown"},
		literate:   true, // All content is comments, code blocks extracted
	},
	{
		name:        "Makefile",
		extensions:  []string{"Makefile", "makefile", ".mk"},
		lineComment: []string{"#"},
	},
	{
		name:        "Dockerfile",
		extensions:  []string{"Dockerfile"},
		lineComment: []string{"#"},
	},
	{
		name:        "Protobuf",
		extensions:  []string{".proto"},
		lineComment: []string{"//"},
		blockStart:  "/*",
		blockEnd:    "*/",
		quotes:      [][2]string{{`"`, `"`}},
	},
	{
		name:        "GraphQL",
		extensions:  []string{".graphql", ".gql"},
		lineComment: []string{"#"},
		quotes:      [][2]string{{`"`, `"`}, {`"""`, `"""`}},
	},
	{
		name:        "Terraform",
		extensions:  []string{".tf", ".tfvars"},
		lineComment: []string{"#", "//"},
		blockStart:  "/*",
		blockEnd:    "*/",
		quotes:      [][2]string{{`"`, `"`}},
	},
	{
		name:        "Zig",
		extensions:  []string{".zig"},
		lineComment: []string{"//"},
		quotes:      [][2]string{{`"`, `"`}},
	},
	{
		name:        "Elixir",
		extensions:  []string{".ex", ".exs"},
		lineComment: []string{"#"},
		quotes:      [][2]string{{`"`, `"`}, {`"""`, `"""`}},
	},
	{
		name:        "Erlang",
		extensions:  []string{".erl", ".hrl"},
		lineComment: []string{"%"},
		quotes:      [][2]string{{`"`, `"`}},
	},
	{
		name:        "Haskell",
		extensions:  []string{".hs"},
		lineComment: []string{"--"},
		blockStart:  "{-",
		blockEnd:    "-}",
		quotes:      [][2]string{{`"`, `"`}},
		nested:      true,
	},
	{
		name:       "OCaml",
		extensions: []string{".ml", ".mli"},
		blockStart: "(*",
		blockEnd:   "*)",
		quotes:     [][2]string{{`"`, `"`}},
		nested:     true,
	},
	{
		name:        "F#",
		extensions:  []string{".fs", ".fsi", ".fsx"},
		lineComment: []string{"//"},
		blockStart:  "(*",
		blockEnd:    "*)",
		quotes:      [][2]string{{`"`, `"`}},
	},
	{
		name:        "Clojure",
		extensions:  []string{".clj", ".cljs", ".cljc"},
		lineComment: []string{";"},
		quotes:      [][2]string{{`"`, `"`}},
	},
	{
		name:        "Nim",
		extensions:  []string{".nim"},
		lineComment: []string{"#"},
		blockStart:  "#[",
		blockEnd:    "]#",
		quotes:      [][2]string{{`"`, `"`}},
	},
	{
		name:        "V",
		extensions:  []string{".v", ".vv"},
		lineComment: []string{"//"},
		blockStart:  "/*",
		blockEnd:    "*/",
		quotes:      [][2]string{{`"`, `"`}, {"'", "'"}},
	},
	{
		name:        "Assembly",
		extensions:  []string{".asm"},
		lineComment: []string{";"},
	},
	{
		name:        "AssemblyGAS",
		extensions:  []string{".s", ".S"},
		lineComment: []string{"//", "#"},
		blockStart:  "/*",
		blockEnd:    "*/",
	},
}

var extToLang = make(map[string]*langDef)

// codeBlockRegex matches markdown code fences with language
var codeBlockStartRegex = regexp.MustCompile("^```(\\w+)?\\s*$")
var codeBlockEndRegex = regexp.MustCompile("^```\\s*$")

func init() {
	for i := range languages {
		for _, ext := range languages[i].extensions {
			extToLang[ext] = &languages[i]
		}
	}
}

// RunLoc counts lines of code
func RunLoc(w io.Writer, args []string, opts Options) error {
	if len(args) == 0 {
		args = []string{"."}
	}

	// Build exclude map
	excludeMap := make(map[string]bool)
	defaultExcludes := []string{".git", "node_modules", "vendor", "__pycache__", ".idea", ".vscode", "target", "build", "dist"}
	for _, e := range defaultExcludes {
		excludeMap[e] = true
	}
	for _, e := range opts.Exclude {
		excludeMap[e] = true
	}

	// Aggregate stats by language
	langStats := make(map[string]*LanguageStats)

	for _, root := range args {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Skip files we can't access
			}

			name := d.Name()

			// Skip hidden files/dirs unless requested
			if !opts.Hidden && strings.HasPrefix(name, ".") && name != "." {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Skip excluded directories
			if d.IsDir() {
				if excludeMap[name] {
					return filepath.SkipDir
				}
				return nil
			}

			// Find language by extension
			ext := filepath.Ext(name)
			lang := extToLang[ext]

			// Try full filename for things like Makefile, Dockerfile
			if lang == nil {
				lang = extToLang[name]
			}

			if lang == nil {
				return nil // Unknown file type
			}

			// Count lines
			fileStats, err := countFile(path, lang)
			if err != nil {
				return nil // Skip files we can't read
			}

			// Aggregate main language stats
			if langStats[lang.name] == nil {
				langStats[lang.name] = &LanguageStats{Language: lang.name}
			}
			ls := langStats[lang.name]
			ls.Files++
			ls.Lines += fileStats.main.Lines
			ls.Code += fileStats.main.Code
			ls.Comments += fileStats.main.Comments
			ls.Blanks += fileStats.main.Blanks

			// Aggregate embedded language stats as children (like tokei)
			for embLang, embStats := range fileStats.embedded {
				if ls.Children == nil {
					ls.Children = make(map[string]*LanguageStats)
				}
				if ls.Children[embLang] == nil {
					ls.Children[embLang] = &LanguageStats{Language: embLang}
				}
				els := ls.Children[embLang]
				els.Lines += embStats.Lines
				els.Code += embStats.Code
				els.Comments += embStats.Comments
				els.Blanks += embStats.Blanks
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("loc: %w", err)
		}
	}

	// Build result
	result := Result{}
	for _, ls := range langStats {
		result.Languages = append(result.Languages, *ls)
		result.Total.Files += ls.Files
		result.Total.Lines += ls.Lines
		result.Total.Code += ls.Code
		result.Total.Comments += ls.Comments
		result.Total.Blanks += ls.Blanks
		// Add embedded language stats to totals
		for _, child := range ls.Children {
			result.Total.Lines += child.Lines
			result.Total.Code += child.Code
			result.Total.Comments += child.Comments
			result.Total.Blanks += child.Blanks
		}
	}
	result.Total.Language = "Total"

	// Sort by code lines (descending)
	sort.Slice(result.Languages, func(i, j int) bool {
		return result.Languages[i].Code > result.Languages[j].Code
	})

	if opts.JSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	return printTable(w, result)
}

// fileStats holds stats for main language and embedded languages
type fileStats struct {
	main     Stats
	embedded map[string]Stats
}

func countFile(path string, lang *langDef) (*fileStats, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	result := &fileStats{
		embedded: make(map[string]Stats),
	}

	scanner := bufio.NewScanner(f)
	// Increase buffer size for files with very long lines (like minified XML/JSON)
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024) // 10MB max line

	if lang.literate {
		countLiterate(scanner, lang, result)
	} else {
		countStandard(scanner, lang, result)
	}

	return result, scanner.Err()
}

// countLiterate handles literate languages like Markdown
// Like tokei: code block content goes to embedded stats, not main
func countLiterate(scanner *bufio.Scanner, lang *langDef, result *fileStats) {
	inCodeBlock := false
	var codeBlockLang string

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Check for code block start
		if !inCodeBlock {
			result.main.Lines++
			if trimmed == "" {
				result.main.Blanks++
				continue
			}
			if matches := codeBlockStartRegex.FindStringSubmatch(trimmed); matches != nil {
				result.main.Comments++ // The ``` line is a comment
				inCodeBlock = true
				if len(matches) > 1 && matches[1] != "" {
					codeBlockLang = detectLanguageName(matches[1])
				} else {
					codeBlockLang = ""
				}
				continue
			}
			// Regular markdown content = comment
			result.main.Comments++
			continue
		}

		// Check for code block end
		if codeBlockEndRegex.MatchString(trimmed) {
			result.main.Lines++
			result.main.Comments++ // The closing ``` is a comment
			inCodeBlock = false
			codeBlockLang = ""
			continue
		}

		// Inside code block with known language - count in embedded stats
		// Blank lines inside code blocks count in BOTH main.Blanks AND embedded.Blanks (like tokei)
		if codeBlockLang != "" {
			embLang := getLangByName(codeBlockLang)
			embStats := result.embedded[codeBlockLang]
			embStats.Lines++
			if trimmed == "" {
				embStats.Blanks++
				result.main.Lines++  // Blank lines count in main too
				result.main.Blanks++
			} else if embLang != nil && isLineComment(trimmed, embLang) {
				embStats.Comments++
			} else {
				embStats.Code++
			}
			result.embedded[codeBlockLang] = embStats
		} else {
			// Unknown code blocks - count as main (prose)
			result.main.Lines++
			if trimmed == "" {
				result.main.Blanks++
			} else {
				result.main.Comments++
			}
		}
	}
}

// countStandard handles normal programming languages with state machine
func countStandard(scanner *bufio.Scanner, lang *langDef, result *fileStats) {
	inBlockComment := false
	blockDepth := 0

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		result.main.Lines++

		if trimmed == "" {
			result.main.Blanks++
			continue
		}

		// Use state machine to parse the line
		lineType := parseLine(trimmed, lang, &inBlockComment, &blockDepth)

		switch lineType {
		case lineCode:
			result.main.Code++
		case lineComment:
			result.main.Comments++
		}
	}
}

type lineClass int

const (
	lineCode lineClass = iota
	lineComment
)

// parseLine determines if a line is code or comment using a state machine
// that handles strings, block comments, and line comments
func parseLine(line string, lang *langDef, inBlockComment *bool, blockDepth *int) lineClass {
	// If we're continuing a block comment from previous line
	if *inBlockComment {
		// Look for block end
		if lang.blockEnd != "" {
			if idx := strings.Index(line, lang.blockEnd); idx >= 0 {
				if lang.nested {
					*blockDepth--
					if *blockDepth == 0 {
						*inBlockComment = false
					}
				} else {
					*inBlockComment = false
				}
				// Check if there's code after the comment ends
				afterComment := strings.TrimSpace(line[idx+len(lang.blockEnd):])
				if afterComment != "" && !startsWithLineComment(afterComment, lang) {
					return lineCode
				}
			}
		}
		return lineComment
	}

	// State machine for parsing the line
	inString := false
	var stringEnd string
	hasCode := false

	for i := 0; i < len(line); {
		remaining := line[i:]

		// If in string, look for string end
		if inString {
			if strings.HasPrefix(remaining, stringEnd) {
				// Check for escape
				if i > 0 && line[i-1] == '\\' {
					i++
					continue
				}
				inString = false
				i += len(stringEnd)
				continue
			}
			i++
			continue
		}

		// Check for string start
		for _, q := range lang.quotes {
			if strings.HasPrefix(remaining, q[0]) {
				inString = true
				stringEnd = q[1]
				hasCode = true
				i += len(q[0])
				break
			}
		}
		if inString {
			continue
		}

		// Check for block comment start
		if lang.blockStart != "" && strings.HasPrefix(remaining, lang.blockStart) {
			// If we had code before this, it's a code line
			if hasCode {
				// Find block end on same line
				afterStart := remaining[len(lang.blockStart):]
				if endIdx := strings.Index(afterStart, lang.blockEnd); endIdx >= 0 {
					// Block comment ends on same line, continue parsing
					i += len(lang.blockStart) + endIdx + len(lang.blockEnd)
					continue
				}
				*inBlockComment = true
				if lang.nested {
					*blockDepth = 1
				}
				return lineCode
			}

			// Check if block comment ends on same line
			afterStart := remaining[len(lang.blockStart):]
			if endIdx := strings.Index(afterStart, lang.blockEnd); endIdx >= 0 {
				// Entire comment on one line, check what's after
				afterComment := strings.TrimSpace(afterStart[endIdx+len(lang.blockEnd):])
				if afterComment == "" {
					return lineComment
				}
				// There's content after the comment
				i += len(lang.blockStart) + endIdx + len(lang.blockEnd)
				continue
			}

			// Block comment continues to next line
			*inBlockComment = true
			if lang.nested {
				*blockDepth = 1
			}
			return lineComment
		}

		// Check for line comment
		for _, lc := range lang.lineComment {
			if strings.HasPrefix(remaining, lc) {
				// If we had code before this comment, it's code
				if hasCode {
					return lineCode
				}
				return lineComment
			}
		}

		// Regular character - it's code
		if !isWhitespace(remaining[0]) {
			hasCode = true
		}
		i++
	}

	if hasCode {
		return lineCode
	}
	return lineComment
}

func startsWithLineComment(line string, lang *langDef) bool {
	for _, lc := range lang.lineComment {
		if strings.HasPrefix(line, lc) {
			return true
		}
	}
	return false
}

// isLineComment checks if a line is purely a comment (for embedded code parsing)
func isLineComment(line string, lang *langDef) bool {
	// Check line comments
	for _, lc := range lang.lineComment {
		if strings.HasPrefix(line, lc) {
			return true
		}
	}
	return false
}

// getLangByName returns a language definition by its display name
func getLangByName(name string) *langDef {
	for i := range languages {
		if languages[i].name == name {
			return &languages[i]
		}
	}
	return nil
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\r' || b == '\n'
}

// detectLanguageName maps code block language hints to our language names
func detectLanguageName(hint string) string {
	hint = strings.ToLower(hint)
	mapping := map[string]string{
		"go":         "Go",
		"golang":     "Go",
		"rust":       "Rust",
		"rs":         "Rust",
		"javascript": "JavaScript",
		"js":         "JavaScript",
		"typescript": "TypeScript",
		"ts":         "TypeScript",
		"python":     "Python",
		"py":         "Python",
		"java":       "Java",
		"c":          "C",
		"cpp":        "C++",
		"c++":        "C++",
		"csharp":     "C#",
		"cs":         "C#",
		"ruby":       "Ruby",
		"rb":         "Ruby",
		"php":        "PHP",
		"swift":      "Swift",
		"kotlin":     "Kotlin",
		"kt":         "Kotlin",
		"scala":      "Scala",
		"shell":      "Shell",
		"bash":       "Shell",
		"sh":         "Shell",
		"zsh":        "Shell",
		"powershell": "PowerShell",
		"ps1":        "PowerShell",
		"lua":        "Lua",
		"perl":       "Perl",
		"r":          "R",
		"sql":        "SQL",
		"html":       "HTML",
		"css":        "CSS",
		"scss":       "SCSS",
		"sass":       "SCSS",
		"xml":        "XML",
		"json":       "JSON",
		"yaml":       "YAML",
		"yml":        "YAML",
		"toml":       "TOML",
		"makefile":   "Makefile",
		"make":       "Makefile",
		"dockerfile": "Dockerfile",
		"docker":     "Dockerfile",
		"protobuf":   "Protobuf",
		"proto":      "Protobuf",
		"graphql":    "GraphQL",
		"gql":        "GraphQL",
		"terraform":  "Terraform",
		"tf":         "Terraform",
		"hcl":        "Terraform",
		"zig":        "Zig",
		"elixir":     "Elixir",
		"ex":         "Elixir",
		"erlang":     "Erlang",
		"erl":        "Erlang",
		"haskell":    "Haskell",
		"hs":         "Haskell",
		"ocaml":      "OCaml",
		"ml":         "OCaml",
		"fsharp":     "F#",
		"fs":         "F#",
		"clojure":    "Clojure",
		"clj":        "Clojure",
		"nim":        "Nim",
		"v":          "V",
		"asm":        "Assembly",
		"assembly":   "Assembly",
		"nasm":       "Assembly",
	}

	if name, ok := mapping[hint]; ok {
		return name
	}
	return ""
}

func printTable(w io.Writer, result Result) error {
	if len(result.Languages) == 0 {
		_, _ = fmt.Fprintln(w, "No files found.")
		return nil
	}

	// Print header
	_, _ = fmt.Fprintf(w, "%-18s %8s %10s %10s %10s %10s\n",
		"Language", "Files", "Lines", "Code", "Comments", "Blanks")
	_, _ = fmt.Fprintln(w, strings.Repeat("-", 79))

	// Print each language
	for _, ls := range result.Languages {
		_, _ = fmt.Fprintf(w, " %-17s %8d %10d %10d %10d %10d\n",
			truncate(ls.Language, 17), ls.Files, ls.Lines, ls.Code, ls.Comments, ls.Blanks)

		// Print embedded languages as children (like tokei)
		if len(ls.Children) > 0 {
			// Sort children by code count
			childNames := make([]string, 0, len(ls.Children))
			for name := range ls.Children {
				childNames = append(childNames, name)
			}
			sort.Slice(childNames, func(i, j int) bool {
				return ls.Children[childNames[i]].Code > ls.Children[childNames[j]].Code
			})

			for _, childName := range childNames {
				child := ls.Children[childName]
				_, _ = fmt.Fprintf(w, " |- %-14s %8s %10d %10d %10d %10d\n",
					truncate(child.Language, 14), "", child.Lines, child.Code, child.Comments, child.Blanks)
			}

			// Print subtotal for parent + children
			totalLines := ls.Lines
			totalCode := ls.Code
			totalComments := ls.Comments
			totalBlanks := ls.Blanks
			for _, child := range ls.Children {
				totalLines += child.Lines
				totalCode += child.Code
				totalComments += child.Comments
				totalBlanks += child.Blanks
			}
			_, _ = fmt.Fprintf(w, " (Total)          %8s %10d %10d %10d %10d\n",
				"", totalLines, totalCode, totalComments, totalBlanks)
		}
	}

	// Print total
	_, _ = fmt.Fprintln(w, strings.Repeat("=", 79))
	_, _ = fmt.Fprintf(w, " %-17s %8d %10d %10d %10d %10d\n",
		"Total", result.Total.Files, result.Total.Lines, result.Total.Code, result.Total.Comments, result.Total.Blanks)

	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "."
}
