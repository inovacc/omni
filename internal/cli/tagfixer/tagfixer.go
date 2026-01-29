// Package tagfixer provides Go struct tag standardization and fixing.
package tagfixer

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// CaseType represents the case convention for struct tags
type CaseType string

const (
	CaseCamel  CaseType = "camel"  // camelCase
	CasePascal CaseType = "pascal" // PascalCase
	CaseSnake  CaseType = "snake"  // snake_case
	CaseKebab  CaseType = "kebab"  // kebab-case
)

// Options configures tagfixer behavior
type Options struct {
	Path      string   // File or directory path
	Case      CaseType // Target case type
	Tags      []string // Tag types to fix (json, yaml, xml, etc.)
	DryRun    bool     // Preview changes without writing
	Recursive bool     // Process directories recursively
	Analyze   bool     // Analyze mode - generate report only
	Verbose   bool     // Verbose output
	JSON      bool     // Output as JSON
}

// FileResult represents the result of processing a file
type FileResult struct {
	Path     string      `json:"path"`
	Modified bool        `json:"modified"`
	Changes  []TagChange `json:"changes,omitempty"`
	Error    string      `json:"error,omitempty"`
}

// TagChange represents a single tag change
type TagChange struct {
	Struct   string `json:"struct"`
	Field    string `json:"field"`
	Tag      string `json:"tag"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}

// AnalysisResult represents the analysis of a codebase
type AnalysisResult struct {
	TotalFiles   int                 `json:"total_files"`
	TotalStructs int                 `json:"total_structs"`
	TotalFields  int                 `json:"total_fields"`
	TagStats     map[string]TagStats `json:"tag_stats"`
	CaseStats    map[string]int      `json:"case_stats"`
	Consistency  float64             `json:"consistency_score"`
	Recommended  CaseType            `json:"recommended_case"`
	Files        []FileAnalysis      `json:"files,omitempty"`
}

// TagStats holds statistics for a tag type
type TagStats struct {
	Count      int            `json:"count"`
	CaseCounts map[string]int `json:"case_counts"`
}

// FileAnalysis represents analysis of a single file
type FileAnalysis struct {
	Path    string           `json:"path"`
	Structs []StructAnalysis `json:"structs"`
}

// StructAnalysis represents analysis of a struct
type StructAnalysis struct {
	Name   string          `json:"name"`
	Fields []FieldAnalysis `json:"fields"`
}

// FieldAnalysis represents analysis of a field
type FieldAnalysis struct {
	Name string            `json:"name"`
	Tags map[string]string `json:"tags"`
}

// Result represents the overall result
type Result struct {
	Files   []FileResult `json:"files"`
	Summary Summary      `json:"summary"`
}

// Summary provides an overview of changes
type Summary struct {
	FilesProcessed int `json:"files_processed"`
	FilesModified  int `json:"files_modified"`
	TotalChanges   int `json:"total_changes"`
	Errors         int `json:"errors"`
}

// RunTagFixer executes the tagfixer command
func RunTagFixer(w io.Writer, opts Options) error {
	if opts.Case == "" {
		opts.Case = CaseCamel
	}

	if len(opts.Tags) == 0 {
		opts.Tags = []string{"json"}
	}

	if opts.Path == "" {
		opts.Path = "."
	}

	if opts.Analyze {
		return runAnalyze(w, opts)
	}

	return runFix(w, opts)
}

func runAnalyze(w io.Writer, opts Options) error {
	files, err := collectGoFiles(opts.Path, opts.Recursive)
	if err != nil {
		return err
	}

	analysis := &AnalysisResult{
		TagStats:  make(map[string]TagStats),
		CaseStats: make(map[string]int),
	}

	for _, file := range files {
		fileAnalysis, err := analyzeFile(file, opts.Tags)
		if err != nil {
			if opts.Verbose {
				_, _ = fmt.Fprintf(w, "Error analyzing %s: %v\n", file, err)
			}

			continue
		}

		analysis.TotalFiles++
		analysis.TotalStructs += len(fileAnalysis.Structs)

		for _, s := range fileAnalysis.Structs {
			analysis.TotalFields += len(s.Fields)

			for _, f := range s.Fields {
				for tagName, tagValue := range f.Tags {
					stats, ok := analysis.TagStats[tagName]
					if !ok {
						stats = TagStats{CaseCounts: make(map[string]int)}
					}

					stats.Count++

					caseType := detectCase(tagValue)
					stats.CaseCounts[caseType]++
					analysis.CaseStats[caseType]++

					analysis.TagStats[tagName] = stats
				}
			}
		}

		if opts.Verbose {
			analysis.Files = append(analysis.Files, *fileAnalysis)
		}
	}

	// Calculate consistency score and recommendation
	analysis.Consistency, analysis.Recommended = calculateConsistency(analysis.CaseStats)

	if opts.JSON {
		return json.NewEncoder(w).Encode(analysis)
	}

	// Text output
	_, _ = fmt.Fprintf(w, "Analysis Report\n")
	_, _ = fmt.Fprintf(w, "===============\n\n")
	_, _ = fmt.Fprintf(w, "Files analyzed:   %d\n", analysis.TotalFiles)
	_, _ = fmt.Fprintf(w, "Structs found:    %d\n", analysis.TotalStructs)
	_, _ = fmt.Fprintf(w, "Fields with tags: %d\n", analysis.TotalFields)
	_, _ = fmt.Fprintf(w, "\nTag Statistics:\n")

	for tag, stats := range analysis.TagStats {
		_, _ = fmt.Fprintf(w, "  %s: %d fields\n", tag, stats.Count)
		for caseType, count := range stats.CaseCounts {
			_, _ = fmt.Fprintf(w, "    - %s: %d\n", caseType, count)
		}
	}

	_, _ = fmt.Fprintf(w, "\nCase Distribution:\n")
	for caseType, count := range analysis.CaseStats {
		_, _ = fmt.Fprintf(w, "  %s: %d\n", caseType, count)
	}

	_, _ = fmt.Fprintf(w, "\nConsistency Score: %.1f%%\n", analysis.Consistency*100)
	_, _ = fmt.Fprintf(w, "Recommended Case:  %s\n", analysis.Recommended)

	return nil
}

func runFix(w io.Writer, opts Options) error {
	files, err := collectGoFiles(opts.Path, opts.Recursive)
	if err != nil {
		return err
	}

	result := &Result{}

	for _, file := range files {
		fileResult := processFile(file, opts)
		result.Files = append(result.Files, fileResult)

		result.Summary.FilesProcessed++
		if fileResult.Modified {
			result.Summary.FilesModified++
		}

		result.Summary.TotalChanges += len(fileResult.Changes)
		if fileResult.Error != "" {
			result.Summary.Errors++
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(result)
	}

	// Text output
	for _, fr := range result.Files {
		if fr.Error != "" {
			_, _ = fmt.Fprintf(w, "Error: %s: %s\n", fr.Path, fr.Error)
			continue
		}

		if len(fr.Changes) > 0 {
			_, _ = fmt.Fprintf(w, "%s:\n", fr.Path)
			for _, change := range fr.Changes {
				_, _ = fmt.Fprintf(w, "  %s.%s [%s]: %s -> %s\n",
					change.Struct, change.Field, change.Tag,
					change.OldValue, change.NewValue)
			}
		} else if opts.Verbose {
			_, _ = fmt.Fprintf(w, "%s: no changes\n", fr.Path)
		}
	}

	if !opts.JSON {
		_, _ = fmt.Fprintf(w, "\nSummary: %d files processed, %d modified, %d changes\n",
			result.Summary.FilesProcessed, result.Summary.FilesModified, result.Summary.TotalChanges)
		if opts.DryRun {
			_, _ = fmt.Fprintf(w, "(dry-run mode - no files were modified)\n")
		}
	}

	return nil
}

func collectGoFiles(path string, recursive bool) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			return []string{path}, nil
		}

		return nil, nil
	}

	var files []string

	walkFn := func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Skip vendor, .git, etc.
			name := d.Name()
			if name == "vendor" || name == ".git" || name == "node_modules" {
				return filepath.SkipDir
			}

			if !recursive && p != path {
				return filepath.SkipDir
			}

			return nil
		}

		if strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") {
			files = append(files, p)
		}

		return nil
	}

	if err := filepath.WalkDir(path, walkFn); err != nil {
		return nil, err
	}

	return files, nil
}

func processFile(path string, opts Options) FileResult {
	result := FileResult{Path: path}

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	modified := false

	ast.Inspect(file, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		structName := typeSpec.Name.Name

		for _, field := range structType.Fields.List {
			if field.Tag == nil || len(field.Names) == 0 {
				continue
			}

			fieldName := field.Names[0].Name
			oldTag := field.Tag.Value

			newTag, changes := fixTag(oldTag, fieldName, structName, opts.Tags, opts.Case)

			if newTag != oldTag {
				if !opts.DryRun {
					field.Tag.Value = newTag
					modified = true
				}

				result.Changes = append(result.Changes, changes...)
			}
		}

		return true
	})

	if modified && !opts.DryRun {
		// Write back to file
		f, err := os.Create(path)
		if err != nil {
			result.Error = err.Error()
			return result
		}

		defer func() { _ = f.Close() }()

		if err := format.Node(f, fset, file); err != nil {
			result.Error = err.Error()
			return result
		}

		result.Modified = true
	} else if len(result.Changes) > 0 {
		result.Modified = true // Would be modified
	}

	return result
}

func analyzeFile(path string, tags []string) (*FileAnalysis, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	analysis := &FileAnalysis{Path: path}

	ast.Inspect(file, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		structAnalysis := StructAnalysis{Name: typeSpec.Name.Name}

		for _, field := range structType.Fields.List {
			if field.Tag == nil || len(field.Names) == 0 {
				continue
			}

			fieldAnalysis := FieldAnalysis{
				Name: field.Names[0].Name,
				Tags: make(map[string]string),
			}

			tagValue := strings.Trim(field.Tag.Value, "`")
			for _, tagName := range tags {
				if value := extractTagValue(tagValue, tagName); value != "" {
					// Extract just the name part (before comma)
					parts := strings.Split(value, ",")
					fieldAnalysis.Tags[tagName] = parts[0]
				}
			}

			if len(fieldAnalysis.Tags) > 0 {
				structAnalysis.Fields = append(structAnalysis.Fields, fieldAnalysis)
			}
		}

		if len(structAnalysis.Fields) > 0 {
			analysis.Structs = append(analysis.Structs, structAnalysis)
		}

		return true
	})

	return analysis, nil
}

var tagRegex = regexp.MustCompile(`(\w+):"([^"]*)"`)

func fixTag(tag, fieldName, structName string, targetTags []string, targetCase CaseType) (string, []TagChange) {
	var changes []TagChange

	newTag := tag

	for _, tagName := range targetTags {
		oldValue := extractTagValue(strings.Trim(tag, "`"), tagName)
		if oldValue == "" {
			continue
		}

		// Parse the tag value (may have options like omitempty)
		parts := strings.Split(oldValue, ",")
		name := parts[0]

		options := ""
		if len(parts) > 1 {
			options = "," + strings.Join(parts[1:], ",")
		}

		// Skip special values
		if name == "-" || name == "" {
			continue
		}

		// Convert to target case
		newName := convertCase(fieldName, targetCase)

		if name != newName {
			// Build new tag value
			newValue := newName + options

			// Replace in tag string
			oldPattern := fmt.Sprintf(`%s:"%s"`, tagName, regexp.QuoteMeta(oldValue))
			newPattern := fmt.Sprintf(`%s:"%s"`, tagName, newValue)
			re := regexp.MustCompile(oldPattern)
			newTag = re.ReplaceAllString(newTag, newPattern)

			changes = append(changes, TagChange{
				Struct:   structName,
				Field:    fieldName,
				Tag:      tagName,
				OldValue: name,
				NewValue: newName,
			})
		}
	}

	return newTag, changes
}

func extractTagValue(tag, key string) string {
	for _, match := range tagRegex.FindAllStringSubmatch(tag, -1) {
		if len(match) == 3 && match[1] == key {
			return match[2]
		}
	}

	return ""
}

func convertCase(name string, targetCase CaseType) string {
	// First, split into words
	words := splitIntoWords(name)

	switch targetCase {
	case CaseCamel:
		return toCamelCase(words)
	case CasePascal:
		return toPascalCase(words)
	case CaseSnake:
		return toSnakeCase(words)
	case CaseKebab:
		return toKebabCase(words)
	default:
		return name
	}
}

func splitIntoWords(s string) []string {
	// Handle snake_case and kebab-case
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")

	// Handle PascalCase and camelCase, including acronyms like HTTP, ID
	var (
		words   []string
		current strings.Builder
	)

	runes := []rune(s)

	for i, r := range runes {
		if r == ' ' {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}

			continue
		}

		if unicode.IsUpper(r) && i > 0 {
			prev := runes[i-1]
			// Split when transitioning from lowercase to uppercase
			if unicode.IsLower(prev) {
				if current.Len() > 0 {
					words = append(words, current.String())
					current.Reset()
				}
			} else if unicode.IsUpper(prev) && i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
				// Split before the start of a new word after an acronym (e.g., HTTPServer -> HTTP + Server)
				if current.Len() > 0 {
					words = append(words, current.String())
					current.Reset()
				}
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		words = append(words, current.String())
	}

	return words
}

func toCamelCase(words []string) string {
	if len(words) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString(strings.ToLower(words[0]))

	for _, word := range words[1:] {
		if len(word) > 0 {
			result.WriteString(strings.ToUpper(string(word[0])) + strings.ToLower(word[1:]))
		}
	}

	return result.String()
}

func toPascalCase(words []string) string {
	var result strings.Builder

	for _, word := range words {
		if len(word) > 0 {
			result.WriteString(strings.ToUpper(string(word[0])) + strings.ToLower(word[1:]))
		}
	}

	return result.String()
}

func toSnakeCase(words []string) string {
	result := make([]string, 0, len(words))
	for _, word := range words {
		result = append(result, strings.ToLower(word))
	}

	return strings.Join(result, "_")
}

func toKebabCase(words []string) string {
	result := make([]string, 0, len(words))
	for _, word := range words {
		result = append(result, strings.ToLower(word))
	}

	return strings.Join(result, "-")
}

func detectCase(s string) string {
	if strings.Contains(s, "_") {
		return "snake_case"
	}

	if strings.Contains(s, "-") {
		return "kebab-case"
	}

	if len(s) > 0 && unicode.IsUpper(rune(s[0])) {
		return "PascalCase"
	}

	if len(s) > 0 && unicode.IsLower(rune(s[0])) {
		for _, r := range s[1:] {
			if unicode.IsUpper(r) {
				return "camelCase"
			}
		}

		return "lowercase"
	}

	return "unknown"
}

func calculateConsistency(caseStats map[string]int) (float64, CaseType) {
	if len(caseStats) == 0 {
		return 1.0, CaseCamel
	}

	total := 0
	maxCount := 0

	var maxCase string

	for caseType, count := range caseStats {
		total += count
		if count > maxCount {
			maxCount = count
			maxCase = caseType
		}
	}

	if total == 0 {
		return 1.0, CaseCamel
	}

	consistency := float64(maxCount) / float64(total)

	// Map detected case to CaseType
	recommended := CaseCamel

	switch maxCase {
	case "camelCase":
		recommended = CaseCamel
	case "PascalCase":
		recommended = CasePascal
	case "snake_case":
		recommended = CaseSnake
	case "kebab-case":
		recommended = CaseKebab
	}

	return consistency, recommended
}

// Library functions for direct use

// ConvertToCase converts a string to the specified case
func ConvertToCase(s string, targetCase CaseType) string {
	return convertCase(s, targetCase)
}

// ToCamelCase converts a string to camelCase
func ToCamelCase(s string) string {
	return convertCase(s, CaseCamel)
}

// ToPascalCase converts a string to PascalCase
func ToPascalCase(s string) string {
	return convertCase(s, CasePascal)
}

// ToSnakeCase converts a string to snake_case
func ToSnakeCase(s string) string {
	return convertCase(s, CaseSnake)
}

// ToKebabCase converts a string to kebab-case
func ToKebabCase(s string) string {
	return convertCase(s, CaseKebab)
}

// DetectCase detects the case type of a string
func DetectCase(s string) string {
	return detectCase(s)
}

// AnalyzePath analyzes Go files at the given path
func AnalyzePath(path string, tags []string, recursive bool) (*AnalysisResult, error) {
	files, err := collectGoFiles(path, recursive)
	if err != nil {
		return nil, err
	}

	analysis := &AnalysisResult{
		TagStats:  make(map[string]TagStats),
		CaseStats: make(map[string]int),
	}

	for _, file := range files {
		fileAnalysis, err := analyzeFile(file, tags)
		if err != nil {
			continue
		}

		analysis.TotalFiles++
		analysis.TotalStructs += len(fileAnalysis.Structs)

		for _, s := range fileAnalysis.Structs {
			analysis.TotalFields += len(s.Fields)

			for _, f := range s.Fields {
				for tagName, tagValue := range f.Tags {
					stats, ok := analysis.TagStats[tagName]
					if !ok {
						stats = TagStats{CaseCounts: make(map[string]int)}
					}

					stats.Count++

					caseType := detectCase(tagValue)
					stats.CaseCounts[caseType]++
					analysis.CaseStats[caseType]++

					analysis.TagStats[tagName] = stats
				}
			}
		}

		analysis.Files = append(analysis.Files, *fileAnalysis)
	}

	analysis.Consistency, analysis.Recommended = calculateConsistency(analysis.CaseStats)

	return analysis, nil
}

// ValidCaseTypes returns all valid case type options
func ValidCaseTypes() []CaseType {
	return []CaseType{CaseCamel, CasePascal, CaseSnake, CaseKebab}
}

// ParseCaseType parses a string into a CaseType
func ParseCaseType(s string) (CaseType, error) {
	switch strings.ToLower(s) {
	case "camel", "camelcase":
		return CaseCamel, nil
	case "pascal", "pascalcase":
		return CasePascal, nil
	case "snake", "snakecase", "snake_case":
		return CaseSnake, nil
	case "kebab", "kebabcase", "kebab-case":
		return CaseKebab, nil
	default:
		return "", fmt.Errorf("unknown case type: %s (valid: camel, pascal, snake, kebab)", s)
	}
}

// GetSortedTags returns commonly used struct tags sorted by usage frequency
func GetSortedTags() []string {
	return []string{"json", "yaml", "xml", "toml", "bson", "db", "form", "query", "mapstructure"}
}

// ListStructTags lists all struct tags found in a Go file
func ListStructTags(path string) ([]string, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	tagSet := make(map[string]struct{})

	ast.Inspect(file, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		for _, field := range structType.Fields.List {
			if field.Tag == nil {
				continue
			}

			tagValue := strings.Trim(field.Tag.Value, "`")
			for _, match := range tagRegex.FindAllStringSubmatch(tagValue, -1) {
				if len(match) >= 2 {
					tagSet[match[1]] = struct{}{}
				}
			}
		}

		return true
	})

	var tags []string
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	sort.Strings(tags)

	return tags, nil
}
