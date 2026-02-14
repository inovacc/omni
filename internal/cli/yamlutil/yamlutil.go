package yamlutil

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/output"
	"gopkg.in/yaml.v3"
)

// ValidateOptions configures the yaml validate command behavior
type ValidateOptions struct {
	OutputFormat output.Format // Output format
	Strict       bool          // --strict: fail on unknown fields
}

// ValidateResult represents the output for JSON mode
type ValidateResult struct {
	File    string `json:"file,omitempty"`
	Valid   bool   `json:"valid"`
	Error   string `json:"error,omitempty"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
	Message string `json:"message,omitempty"`
}

// RunValidate validates YAML input
func RunValidate(w io.Writer, args []string, opts ValidateOptions) error {
	if len(args) == 0 {
		// Read from stdin
		return validateReader(w, os.Stdin, "<stdin>", opts)
	}

	var hasError bool

	for _, arg := range args {
		// Check if it's a file
		if info, err := os.Stat(arg); err == nil && !info.IsDir() {
			f, err := os.Open(arg)
			if err != nil {
				return fmt.Errorf("yaml validate: %w", err)
			}

			err = validateReader(w, f, arg, opts)
			_ = f.Close()

			if err != nil {
				hasError = true
			}
		} else {
			// Treat as literal YAML string
			err := validateReader(w, strings.NewReader(arg), "<input>", opts)
			if err != nil {
				hasError = true
			}
		}
	}

	if hasError {
		return fmt.Errorf("yaml validation failed")
	}

	return nil
}

func validateReader(w io.Writer, r io.Reader, name string, opts ValidateOptions) error {
	content, err := io.ReadAll(r)
	if err != nil {
		return outputResult(w, ValidateResult{
			File:  name,
			Valid: false,
			Error: err.Error(),
		}, opts)
	}

	var data any

	decoder := yaml.NewDecoder(strings.NewReader(string(content)))
	if opts.Strict {
		decoder.KnownFields(true)
	}

	// Try to decode all documents in the YAML
	docCount := 0

	for {
		err = decoder.Decode(&data)
		if err == io.EOF {
			break
		}

		if err != nil {
			result := ValidateResult{
				File:  name,
				Valid: false,
				Error: err.Error(),
			}

			// Try to extract line/column info from yaml.TypeError
			if typeErr, ok := err.(*yaml.TypeError); ok {
				result.Message = strings.Join(typeErr.Errors, "; ")
			}

			return outputResult(w, result, opts)
		}

		docCount++
	}

	result := ValidateResult{
		File:    name,
		Valid:   true,
		Message: fmt.Sprintf("valid YAML (%d document(s))", docCount),
	}

	return outputResult(w, result, opts)
}

func outputResult(w io.Writer, result ValidateResult, opts ValidateOptions) error {
	f := output.New(w, opts.OutputFormat)
	if f.IsJSON() {
		return f.Print(result)
	}

	if result.Valid {
		_, _ = fmt.Fprintf(w, "%s: %s\n", result.File, result.Message)
	} else {
		_, _ = fmt.Fprintf(w, "%s: invalid YAML - %s\n", result.File, result.Error)
	}

	if !result.Valid {
		return fmt.Errorf("validation failed")
	}

	return nil
}

// FormatOptions configures the yaml format command behavior
type FormatOptions struct {
	Indent      int  // indentation width
	JSON        bool // output as JSON instead
	SortKeys    bool // sort keys alphabetically
	RemoveEmpty bool // remove empty/null values
	InPlace     bool // modify file in place
	K8s         bool // use Kubernetes key ordering
}

// RunFormat formats YAML input
func RunFormat(w io.Writer, args []string, opts FormatOptions) error {
	input, filename, err := getInputWithFilename(args)
	if err != nil {
		return err
	}

	// Parse all documents
	docs, err := parseMultiDoc(input)
	if err != nil {
		return fmt.Errorf("yaml format: %w", err)
	}

	// Process each document
	for i, doc := range docs {
		if opts.RemoveEmpty {
			doc = removeEmptyValues(doc)
		}

		if opts.SortKeys {
			doc = sortKeys(doc)
		}

		if opts.K8s {
			doc = sortK8sKeys(doc)
		}

		docs[i] = doc
	}

	// Handle in-place editing
	if opts.InPlace && filename != "" {
		f, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("yaml format: %w", err)
		}

		defer func() { _ = f.Close() }()

		w = f
	}

	// Output
	if opts.JSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", strings.Repeat(" ", opts.Indent))

		if len(docs) == 1 {
			return enc.Encode(docs[0])
		}

		return enc.Encode(docs)
	}

	enc := yaml.NewEncoder(w)
	enc.SetIndent(opts.Indent)

	for _, doc := range docs {
		if err := enc.Encode(doc); err != nil {
			return fmt.Errorf("yaml format: %w", err)
		}
	}

	return enc.Close()
}

// K8sFormatOptions configures the yaml k8s command behavior
type K8sFormatOptions struct {
	Indent      int  // indentation width
	RemoveEmpty bool // remove empty/null values
	InPlace     bool // modify file in place
}

// RunK8sFormat formats YAML with Kubernetes conventions
func RunK8sFormat(w io.Writer, args []string, opts K8sFormatOptions) error {
	fmtOpts := FormatOptions{
		Indent:      opts.Indent,
		RemoveEmpty: opts.RemoveEmpty,
		InPlace:     opts.InPlace,
		K8s:         true,
	}

	return RunFormat(w, args, fmtOpts)
}

// parseMultiDoc parses a YAML string with multiple documents
func parseMultiDoc(input string) ([]any, error) {
	var docs []any

	decoder := yaml.NewDecoder(strings.NewReader(input))

	for {
		var doc any

		err := decoder.Decode(&doc)
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		if doc != nil {
			docs = append(docs, doc)
		}
	}

	return docs, nil
}

// sortKeys recursively sorts map keys alphabetically
func sortKeys(v any) any {
	switch val := v.(type) {
	case map[string]any:
		sorted := make(map[string]any)
		for k, v := range val {
			sorted[k] = sortKeys(v)
		}

		return sorted
	case []any:
		result := make([]any, len(val))
		for i, item := range val {
			result[i] = sortKeys(item)
		}

		return result
	default:
		return v
	}
}

// k8sKeyOrder defines the standard Kubernetes key ordering
var k8sKeyOrder = []string{
	"apiVersion", "kind", "metadata", "spec", "status",
	"data", "stringData", "binaryData",
	"rules", "roleRef", "subjects",
	"template", "containers", "initContainers",
	"volumes", "volumeMounts",
	"ports", "env", "envFrom",
	"resources", "limits", "requests",
	"selector", "matchLabels", "matchExpressions",
	"replicas", "strategy",
}

// k8sMetadataOrder defines metadata field ordering
var k8sMetadataOrder = []string{
	"name", "namespace", "labels", "annotations",
	"generateName", "uid", "resourceVersion",
	"generation", "creationTimestamp", "deletionTimestamp",
	"ownerReferences", "finalizers",
}

// sortK8sKeys sorts keys according to Kubernetes conventions
func sortK8sKeys(v any) any {
	return sortK8sKeysWithOrder(v, k8sKeyOrder)
}

// sortK8sKeysWithOrder sorts keys with the given order preference
func sortK8sKeysWithOrder(v any, keyOrder []string) any {
	switch val := v.(type) {
	case map[string]any:
		// Process values recursively with appropriate ordering
		processed := make(map[string]any)

		for k, v := range val {
			if k == "metadata" {
				// Use metadata-specific ordering for metadata field
				processed[k] = sortK8sKeysWithOrder(v, k8sMetadataOrder)
			} else {
				processed[k] = sortK8sKeysWithOrder(v, k8sKeyOrder)
			}
		}

		return createOrderedMap(processed, keyOrder)
	case []any:
		items := make([]any, len(val))
		for i, item := range val {
			items[i] = sortK8sKeysWithOrder(item, k8sKeyOrder)
		}

		return items
	default:
		return v
	}
}

// OrderedMap preserves key order in YAML output
type OrderedMap struct {
	Keys   []string
	Values map[string]any
}

// MarshalYAML implements yaml.Marshaler for ordered output
func (o OrderedMap) MarshalYAML() (any, error) {
	node := &yaml.Node{
		Kind: yaml.MappingNode,
	}

	for _, k := range o.Keys {
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: k}

		var valueNode yaml.Node
		if err := valueNode.Encode(o.Values[k]); err != nil {
			return nil, err
		}

		node.Content = append(node.Content, keyNode, &valueNode)
	}

	return node, nil
}

// createOrderedMap creates an OrderedMap with specified key ordering
func createOrderedMap(m map[string]any, keyOrder []string) OrderedMap {
	result := OrderedMap{
		Values: m,
	}

	// Build ordered keys list
	added := make(map[string]bool)

	// Add keys in specified order first
	for _, k := range keyOrder {
		if _, exists := m[k]; exists {
			result.Keys = append(result.Keys, k)
			added[k] = true
		}
	}

	// Add remaining keys alphabetically
	var remaining []string

	for k := range m {
		if !added[k] {
			remaining = append(remaining, k)
		}
	}

	sortStrings(remaining)
	result.Keys = append(result.Keys, remaining...)

	return result
}

// sortStrings sorts a string slice in place
func sortStrings(s []string) {
	for i := 0; i < len(s)-1; i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// removeEmptyValues recursively removes nil and empty values
func removeEmptyValues(v any) any {
	switch val := v.(type) {
	case map[string]any:
		result := make(map[string]any)

		for k, v := range val {
			cleaned := removeEmptyValues(v)
			if !isEmpty(cleaned) {
				result[k] = cleaned
			}
		}

		if len(result) == 0 {
			return nil
		}

		return result
	case []any:
		var result []any

		for _, item := range val {
			cleaned := removeEmptyValues(item)
			if !isEmpty(cleaned) {
				result = append(result, cleaned)
			}
		}

		if len(result) == 0 {
			return nil
		}

		return result
	default:
		return v
	}
}

// isEmpty checks if a value is considered empty
func isEmpty(v any) bool {
	if v == nil {
		return true
	}

	switch val := v.(type) {
	case string:
		return val == ""
	case map[string]any:
		return len(val) == 0
	case []any:
		return len(val) == 0
	}

	return false
}

// getInputWithFilename reads input and returns the filename if from file
func getInputWithFilename(args []string) (string, string, error) {
	if len(args) > 0 {
		// Check if it's a file
		if _, err := os.Stat(args[0]); err == nil {
			content, err := os.ReadFile(args[0])
			if err != nil {
				return "", "", fmt.Errorf("yaml: %w", err)
			}

			return string(content), args[0], nil
		}
		// Treat as literal string
		return strings.Join(args, " "), "", nil
	}

	// Read from stdin
	scanner := bufio.NewScanner(os.Stdin)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("yaml: %w", err)
	}

	return strings.Join(lines, "\n"), "", nil
}
