package env

import (
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
