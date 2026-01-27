package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// DotenvOptions configures the dotenv command behavior
type DotenvOptions struct {
	Files    []string // .env files to load
	Export   bool     // -e: output as export statements
	Quiet    bool     // -q: suppress warnings
	Override bool     // -o: override existing environment variables
	Expand   bool     // -x: expand variables in values
}

// EnvVar represents an environment variable
type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RunDotenv loads environment variables from .env files
func RunDotenv(w io.Writer, args []string, opts DotenvOptions) error {
	files := opts.Files
	if len(files) == 0 {
		files = args
	}

	if len(files) == 0 {
		files = []string{".env"}
	}

	var allVars []EnvVar

	for _, file := range files {
		vars, err := ParseDotenvFile(file, opts)
		if err != nil {
			if !opts.Quiet {
				_, _ = fmt.Fprintf(os.Stderr, "dotenv: %s: %v\n", file, err)
			}

			continue
		}

		allVars = append(allVars, vars...)
	}

	// Output or set variables
	for _, v := range allVars {
		if opts.Export {
			_, _ = fmt.Fprintf(w, "export %s=%q\n", v.Key, v.Value)
		} else {
			_, _ = fmt.Fprintf(w, "%s=%s\n", v.Key, v.Value)
		}
	}

	return nil
}

// ParseDotenvFile parses a .env file and returns the variables
func ParseDotenvFile(path string, opts DotenvOptions) ([]EnvVar, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = f.Close()
	}()

	return ParseDotenv(f, opts)
}

// ParseDotenv parses .env content from a reader
func ParseDotenv(r io.Reader, opts DotenvOptions) ([]EnvVar, error) {
	var vars []EnvVar

	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip empty lines and comments
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value
		key, value, err := parseDotenvLine(line)
		if err != nil {
			continue
		}

		// Expand variables if requested
		if opts.Expand {
			value = expandEnvVars(value, vars)
		}

		vars = append(vars, EnvVar{Key: key, Value: value})
	}

	return vars, scanner.Err()
}

func parseDotenvLine(line string) (string, string, error) {
	// Handle export prefix
	if after, ok := strings.CutPrefix(line, "export "); ok {
		line = after
		line = strings.TrimSpace(line)
	}

	// Find the = separator
	before, after, ok := strings.Cut(line, "=")
	if !ok {
		return "", "", fmt.Errorf("invalid line: no '=' found")
	}

	key := strings.TrimSpace(before)
	value := after

	// Validate key
	if key == "" {
		return "", "", fmt.Errorf("empty key")
	}

	// Parse value (handle quotes)
	value = parseValue(value)

	return key, value, nil
}

func parseValue(value string) string {
	value = strings.TrimSpace(value)

	// Handle quoted values
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'') {
			quote := value[0]
			value = value[1 : len(value)-1]

			// Handle escape sequences in double quotes
			if quote == '"' {
				value = unescapeDoubleQuoted(value)
			}

			return value
		}
	}

	// Unquoted value - remove inline comments
	if idx := strings.Index(value, " #"); idx != -1 {
		value = strings.TrimSpace(value[:idx])
	}

	return value
}

func unescapeDoubleQuoted(s string) string {
	replacer := strings.NewReplacer(
		`\\`, `\`,
		`\"`, `"`,
		`\n`, "\n",
		`\r`, "\r",
		`\t`, "\t",
	)

	return replacer.Replace(s)
}

func expandEnvVars(value string, vars []EnvVar) string {
	// Build a map of current variables
	varMap := make(map[string]string)
	for _, v := range vars {
		varMap[v.Key] = v.Value
	}

	// Expand ${VAR} and $VAR patterns
	result := os.Expand(value, func(key string) string {
		// First check our parsed vars
		if v, ok := varMap[key]; ok {
			return v
		}
		// Then check environment
		return os.Getenv(key)
	})

	return result
}

// LoadDotenv loads .env file(s) into the current process environment
func LoadDotenv(files ...string) error {
	if len(files) == 0 {
		files = []string{".env"}
	}

	opts := DotenvOptions{Override: false, Expand: true}

	for _, file := range files {
		vars, err := ParseDotenvFile(file, opts)
		if err != nil {
			// Silently skip missing .env files
			if os.IsNotExist(err) {
				continue
			}

			return err
		}

		for _, v := range vars {
			// Only set if not already set (unless Override is true)
			if _, exists := os.LookupEnv(v.Key); !exists || opts.Override {
				if err := os.Setenv(v.Key, v.Value); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// LoadDotenvOverride loads .env file(s) and overrides existing vars
func LoadDotenvOverride(files ...string) error {
	if len(files) == 0 {
		files = []string{".env"}
	}

	opts := DotenvOptions{Override: true, Expand: true}

	for _, file := range files {
		vars, err := ParseDotenvFile(file, opts)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return err
		}

		for _, v := range vars {
			if err := os.Setenv(v.Key, v.Value); err != nil {
				return err
			}
		}
	}

	return nil
}

// MustLoadDotenv loads .env file(s) and panics on error
func MustLoadDotenv(files ...string) {
	if err := LoadDotenv(files...); err != nil {
		panic(fmt.Sprintf("dotenv: %v", err))
	}
}
