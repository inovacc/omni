package task

import (
	"fmt"
	maps0 "maps"
	"os"
	"regexp"
	"strings"
)

// VarResolver handles variable expansion in task commands
type VarResolver struct {
	vars     map[string]any
	envVars  map[string]string
	taskVars map[string]any
}

// NewVarResolver creates a new variable resolver
func NewVarResolver(global, task map[string]any, env map[string]string) *VarResolver {
	return &VarResolver{
		vars:     global,
		envVars:  env,
		taskVars: task,
	}
}

// templateVarPattern matches {{.VAR}} and {{ .VAR }}
var templateVarPattern = regexp.MustCompile(`\{\{\s*\.([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`)

// envVarPattern matches $VAR and ${VAR}
var envVarPattern = regexp.MustCompile(`\$\{?([a-zA-Z_][a-zA-Z0-9_]*)\}?`)

// Expand expands all variables in a string
func (r *VarResolver) Expand(s string) string {
	// First expand template variables {{.VAR}}
	s = templateVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		// Extract variable name
		parts := templateVarPattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		name := parts[1]

		return r.resolveVar(name)
	})

	// Then expand env variables $VAR
	s = envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		// Extract variable name
		parts := envVarPattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		name := parts[1]

		// Check taskfile env first
		if val, ok := r.envVars[name]; ok {
			return val
		}

		// Fall back to OS environment
		return os.Getenv(name)
	})

	return s
}

// resolveVar resolves a single variable by name
func (r *VarResolver) resolveVar(name string) string {
	// Task-level vars take precedence
	if val, ok := r.taskVars[name]; ok {
		return formatVar(val)
	}

	// Then global vars
	if val, ok := r.vars[name]; ok {
		return formatVar(val)
	}

	// Then env vars
	if val, ok := r.envVars[name]; ok {
		return val
	}

	// Finally OS environment
	if val := os.Getenv(name); val != "" {
		return val
	}

	// Return empty if not found
	return ""
}

// formatVar converts a variable value to string
func formatVar(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case int, int64, int32:
		return fmt.Sprintf("%d", val)
	case float64, float32:
		return fmt.Sprintf("%v", val)
	case bool:
		if val {
			return "true"
		}

		return "false"
	case []any:
		// Join slice elements
		var parts []string
		for _, elem := range val {
			parts = append(parts, formatVar(elem))
		}

		return strings.Join(parts, " ")
	case map[string]any:
		// For maps, just return empty (complex types not supported in cmd expansion)
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

// MergeVars merges multiple variable maps (later takes precedence)
func MergeVars(maps ...map[string]any) map[string]any {
	result := make(map[string]any)

	for _, m := range maps {
		maps0.Copy(result, m)
	}

	return result
}

// EvaluateDynamicVar evaluates a dynamic variable (sh command result)
// For omni, we only support static values, not shell execution
func EvaluateDynamicVar(v any) (string, error) {
	switch val := v.(type) {
	case string:
		return val, nil
	case map[string]any:
		// Check for sh: field
		if sh, ok := val["sh"]; ok {
			return "", fmt.Errorf("dynamic shell variables not supported (omni cannot exec external commands): sh: %v", sh)
		}
		// Check for ref: field (reference to another var)
		if ref, ok := val["ref"]; ok {
			return formatVar(ref), nil
		}

		return "", fmt.Errorf("unsupported variable format: %v", val)
	default:
		return formatVar(val), nil
	}
}
