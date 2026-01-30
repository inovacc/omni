package json2struct

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"

	"github.com/inovacc/omni/internal/cli/input"
)

// Options configures the json2struct command behavior
type Options struct {
	Name       string // struct name (default: "Root")
	Package    string // package name (default: "main")
	Inline     bool   // inline nested structs
	OmitEmpty  bool   // add omitempty to all fields
	UsePointer bool   // use pointers for optional fields
}

// RunJSON2Struct converts JSON to Go struct definition
func RunJSON2Struct(w io.Writer, r io.Reader, args []string, opts Options) error {
	sources, err := input.Open(args, r)
	if err != nil {
		return fmt.Errorf("json2struct: %w", err)
	}
	defer input.CloseAll(sources)

	if opts.Name == "" {
		opts.Name = "Root"
	}

	if opts.Package == "" {
		opts.Package = "main"
	}

	for _, src := range sources {
		data, err := io.ReadAll(src.Reader)
		if err != nil {
			return fmt.Errorf("json2struct: %w", err)
		}

		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("json2struct: invalid JSON: %w", err)
		}

		gen := &generator{
			opts:    opts,
			structs: make(map[string]string),
		}

		mainType := gen.generateType(opts.Name, v)

		// Output package declaration
		_, _ = fmt.Fprintf(w, "package %s\n\n", opts.Package)

		// Output structs in order (main struct first, then nested)
		if opts.Inline {
			_, _ = fmt.Fprintf(w, "type %s %s\n", opts.Name, mainType)
		} else {
			// Collect and sort struct names for deterministic output
			var names []string
			for name := range gen.structs {
				names = append(names, name)
			}

			sort.Strings(names)

			// Output main struct first
			if def, ok := gen.structs[opts.Name]; ok {
				_, _ = fmt.Fprintf(w, "type %s %s\n", opts.Name, def)
				delete(gen.structs, opts.Name)
			}

			// Output remaining structs
			for _, name := range names {
				if def, ok := gen.structs[name]; ok {
					_, _ = fmt.Fprintf(w, "\ntype %s %s\n", name, def)
				}
			}
		}
	}

	return nil
}

type generator struct {
	opts    Options
	structs map[string]string
	counter int
}

func (g *generator) generateType(name string, v any) string {
	switch val := v.(type) {
	case nil:
		return "any"
	case bool:
		return "bool"
	case float64:
		// Check if it's actually an integer
		if val == float64(int64(val)) {
			return "int"
		}

		return "float64"
	case string:
		return "string"
	case []any:
		if len(val) == 0 {
			return "[]any"
		}

		elemType := g.generateType(singularize(name), val[0])

		return "[]" + elemType
	case map[string]any:
		return g.generateStruct(name, val)
	default:
		return "any"
	}
}

func (g *generator) generateStruct(name string, obj map[string]any) string {
	if len(obj) == 0 {
		return "map[string]any"
	}

	var fields []string

	// Sort keys for deterministic output
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, key := range keys {
		val := obj[key]
		fieldName := toGoName(key)
		fieldType := g.generateType(fieldName, val)

		// Handle nested structs
		if _, isMap := val.(map[string]any); isMap && !g.opts.Inline {
			// Register as separate struct
			g.structs[fieldName] = fieldType
			fieldType = fieldName
		}

		// Build tag
		tag := fmt.Sprintf("`json:\"%s", key)
		if g.opts.OmitEmpty {
			tag += ",omitempty"
		}

		tag += "\"`"

		fields = append(fields, fmt.Sprintf("\t%s %s %s", fieldName, fieldType, tag))
	}

	structDef := "struct {\n" + strings.Join(fields, "\n") + "\n}"

	if !g.opts.Inline {
		g.structs[name] = structDef
	}

	return structDef
}

// toGoName converts a JSON key to a valid Go identifier
func toGoName(s string) string {
	if s == "" {
		return "Field"
	}

	// Split by common separators
	var (
		words   []string
		current strings.Builder
	)

	for i, r := range s {
		if r == '_' || r == '-' || r == ' ' || r == '.' {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		} else if i > 0 && unicode.IsUpper(r) && !unicode.IsUpper(rune(s[i-1])) {
			// camelCase boundary
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}

			current.WriteRune(r)
		} else {
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		words = append(words, current.String())
	}

	// Title case each word
	var result strings.Builder

	for _, word := range words {
		if len(word) > 0 {
			// Handle common acronyms
			upper := strings.ToUpper(word)
			if isAcronym(upper) {
				result.WriteString(upper)
			} else {
				result.WriteString(strings.ToUpper(string(word[0])))
				result.WriteString(strings.ToLower(word[1:]))
			}
		}
	}

	name := result.String()
	if name == "" {
		return "Field"
	}

	// Ensure starts with letter
	if !unicode.IsLetter(rune(name[0])) {
		name = "F" + name
	}

	return name
}

func isAcronym(s string) bool {
	acronyms := map[string]bool{
		"ID": true, "URL": true, "URI": true, "API": true,
		"HTTP": true, "HTTPS": true, "HTML": true, "JSON": true,
		"XML": true, "SQL": true, "SSH": true, "TCP": true,
		"UDP": true, "IP": true, "DNS": true, "UUID": true,
	}

	return acronyms[s]
}

func singularize(s string) string {
	if strings.HasSuffix(s, "ies") {
		return s[:len(s)-3] + "y"
	}

	if strings.HasSuffix(s, "es") {
		return s[:len(s)-2]
	}

	if strings.HasSuffix(s, "s") && len(s) > 1 {
		return s[:len(s)-1]
	}

	return s + "Item"
}
