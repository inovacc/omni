package loc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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
	Language string `json:"language"`
	Files    int    `json:"files"`
	Lines    int    `json:"lines"`
	Code     int    `json:"code"`
	Comments int    `json:"comments"`
	Blanks   int    `json:"blanks"`
}

// Result holds the complete LOC analysis result
type Result struct {
	Languages []LanguageStats `json:"languages"`
	Total     LanguageStats   `json:"total"`
}

// Language definition
type langDef struct {
	name       string
	extensions []string
	lineComment string
	blockStart  string
	blockEnd    string
}

var languages = []langDef{
	{"Go", []string{".go"}, "//", "/*", "*/"},
	{"Rust", []string{".rs"}, "//", "/*", "*/"},
	{"JavaScript", []string{".js", ".mjs", ".cjs"}, "//", "/*", "*/"},
	{"TypeScript", []string{".ts", ".tsx"}, "//", "/*", "*/"},
	{"Python", []string{".py", ".pyw"}, "#", `"""`, `"""`},
	{"Java", []string{".java"}, "//", "/*", "*/"},
	{"C", []string{".c", ".h"}, "//", "/*", "*/"},
	{"C++", []string{".cpp", ".cc", ".cxx", ".hpp", ".hxx"}, "//", "/*", "*/"},
	{"C#", []string{".cs"}, "//", "/*", "*/"},
	{"Ruby", []string{".rb"}, "#", "=begin", "=end"},
	{"PHP", []string{".php"}, "//", "/*", "*/"},
	{"Swift", []string{".swift"}, "//", "/*", "*/"},
	{"Kotlin", []string{".kt", ".kts"}, "//", "/*", "*/"},
	{"Scala", []string{".scala"}, "//", "/*", "*/"},
	{"Shell", []string{".sh", ".bash", ".zsh"}, "#", "", ""},
	{"PowerShell", []string{".ps1", ".psm1"}, "#", "<#", "#>"},
	{"Lua", []string{".lua"}, "--", "--[[", "]]"},
	{"Perl", []string{".pl", ".pm"}, "#", "=pod", "=cut"},
	{"R", []string{".r", ".R"}, "#", "", ""},
	{"SQL", []string{".sql"}, "--", "/*", "*/"},
	{"HTML", []string{".html", ".htm"}, "", "<!--", "-->"},
	{"CSS", []string{".css"}, "", "/*", "*/"},
	{"SCSS", []string{".scss", ".sass"}, "//", "/*", "*/"},
	{"XML", []string{".xml", ".xsl", ".xslt"}, "", "<!--", "-->"},
	{"JSON", []string{".json"}, "", "", ""},
	{"YAML", []string{".yaml", ".yml"}, "#", "", ""},
	{"TOML", []string{".toml"}, "#", "", ""},
	{"Markdown", []string{".md", ".markdown"}, "", "", ""},
	{"Makefile", []string{"Makefile", "makefile", ".mk"}, "#", "", ""},
	{"Dockerfile", []string{"Dockerfile"}, "#", "", ""},
	{"Protobuf", []string{".proto"}, "//", "/*", "*/"},
	{"GraphQL", []string{".graphql", ".gql"}, "#", "", ""},
	{"Terraform", []string{".tf", ".tfvars"}, "#", "/*", "*/"},
	{"Zig", []string{".zig"}, "//", "", ""},
	{"Elixir", []string{".ex", ".exs"}, "#", "", ""},
	{"Erlang", []string{".erl", ".hrl"}, "%", "", ""},
	{"Haskell", []string{".hs"}, "--", "{-", "-}"},
	{"OCaml", []string{".ml", ".mli"}, "", "(*", "*)"},
	{"F#", []string{".fs", ".fsi", ".fsx"}, "//", "(*", "*)"},
	{"Clojure", []string{".clj", ".cljs", ".cljc"}, ";", "", ""},
	{"Nim", []string{".nim"}, "#", "#[", "]#"},
	{"V", []string{".v"}, "//", "/*", "*/"},
	{"Assembly", []string{".asm", ".s", ".S"}, ";", "", ""},
}

var extToLang = make(map[string]*langDef)

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
			stats, err := countFile(path, lang)
			if err != nil {
				return nil // Skip files we can't read
			}

			// Aggregate
			if langStats[lang.name] == nil {
				langStats[lang.name] = &LanguageStats{Language: lang.name}
			}
			ls := langStats[lang.name]
			ls.Files++
			ls.Lines += stats.Lines
			ls.Code += stats.Code
			ls.Comments += stats.Comments
			ls.Blanks += stats.Blanks

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

func countFile(path string, lang *langDef) (*Stats, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	stats := &Stats{}
	scanner := bufio.NewScanner(f)
	inBlockComment := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		stats.Lines++

		if line == "" {
			stats.Blanks++
			continue
		}

		// Handle block comments
		if lang.blockStart != "" && lang.blockEnd != "" {
			if inBlockComment {
				stats.Comments++
				if strings.Contains(line, lang.blockEnd) {
					inBlockComment = false
				}
				continue
			}

			if strings.HasPrefix(line, lang.blockStart) {
				stats.Comments++
				if !strings.Contains(line, lang.blockEnd) ||
				   strings.Index(line, lang.blockEnd) < strings.Index(line, lang.blockStart)+len(lang.blockStart) {
					inBlockComment = true
				}
				continue
			}
		}

		// Handle line comments
		if lang.lineComment != "" && strings.HasPrefix(line, lang.lineComment) {
			stats.Comments++
			continue
		}

		stats.Code++
	}

	return stats, scanner.Err()
}

func printTable(w io.Writer, result Result) error {
	if len(result.Languages) == 0 {
		_, _ = fmt.Fprintln(w, "No files found.")
		return nil
	}

	// Print header
	_, _ = fmt.Fprintf(w, "%-15s %8s %10s %10s %10s %10s\n",
		"Language", "Files", "Lines", "Code", "Comments", "Blanks")
	_, _ = fmt.Fprintln(w, strings.Repeat("-", 67))

	// Print each language
	for _, ls := range result.Languages {
		_, _ = fmt.Fprintf(w, "%-15s %8d %10d %10d %10d %10d\n",
			truncate(ls.Language, 15), ls.Files, ls.Lines, ls.Code, ls.Comments, ls.Blanks)
	}

	// Print total
	_, _ = fmt.Fprintln(w, strings.Repeat("-", 67))
	_, _ = fmt.Fprintf(w, "%-15s %8d %10d %10d %10d %10d\n",
		"Total", result.Total.Files, result.Total.Lines, result.Total.Code, result.Total.Comments, result.Total.Blanks)

	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "."
}
