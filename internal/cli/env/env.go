package env

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// EnvOptions configures the env command behavior
type EnvOptions struct {
	NullTerminated bool   // -0: end each output line with NUL, not newline
	Unset          string // -u: remove variable from the environment (for display only)
	Ignore         bool   // -i: start with an empty environment
	JSON           bool   // --json: output in JSON format
}

// EnvVar represents an environment variable for JSON output
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// RunEnv prints environment variables
func RunEnv(w io.Writer, args []string, opts EnvOptions) error {
	var envVars []string

	if opts.Ignore {
		// Start with empty environment - only show explicitly set vars
		envVars = []string{}
	} else {
		envVars = os.Environ()
	}

	// Filter out unset variable if specified
	if opts.Unset != "" {
		filtered := make([]string, 0, len(envVars))

		prefix := opts.Unset + "="
		for _, env := range envVars {
			if !strings.HasPrefix(env, prefix) {
				filtered = append(filtered, env)
			}
		}

		envVars = filtered
	}

	// If args provided, print only those variables
	if len(args) > 0 {
		if opts.JSON {
			result := make([]EnvVar, 0, len(args))
			for _, name := range args {
				value := os.Getenv(name)
				if value != "" {
					result = append(result, EnvVar{Name: name, Value: value})
				}
			}

			return json.NewEncoder(w).Encode(result)
		}

		for _, name := range args {
			value := os.Getenv(name)
			if value != "" {
				if opts.NullTerminated {
					_, _ = fmt.Fprintf(w, "%s=%s\x00", name, value)
				} else {
					_, _ = fmt.Fprintf(w, "%s=%s\n", name, value)
				}
			}
		}

		return nil
	}

	// Sort for consistent output
	sort.Strings(envVars)

	// JSON output
	if opts.JSON {
		result := make([]EnvVar, 0, len(envVars))
		for _, env := range envVars {
			if idx := strings.Index(env, "="); idx > 0 {
				result = append(result, EnvVar{Name: env[:idx], Value: env[idx+1:]})
			}
		}

		return json.NewEncoder(w).Encode(result)
	}

	terminator := "\n"
	if opts.NullTerminated {
		terminator = "\x00"
	}

	for _, env := range envVars {
		_, _ = fmt.Fprint(w, env+terminator)
	}

	return nil
}

// GetEnv returns the value of an environment variable
func GetEnv(name string) string {
	return os.Getenv(name)
}

// LookupEnv returns the value of an environment variable and whether it was set
func LookupEnv(name string) (string, bool) {
	return os.LookupEnv(name)
}

// Environ returns all environment variables as a map
func Environ() map[string]string {
	env := make(map[string]string)

	for _, e := range os.Environ() {
		if idx := strings.Index(e, "="); idx > 0 {
			env[e[:idx]] = e[idx+1:]
		}
	}

	return env
}
