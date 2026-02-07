package dotenv

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

// ShellType represents a shell type for export format
type ShellType string

const (
	ShellAuto       ShellType = "auto"
	ShellBash       ShellType = "bash"
	ShellZsh        ShellType = "zsh"
	ShellFish       ShellType = "fish"
	ShellPowerShell ShellType = "powershell"
	ShellCmd        ShellType = "cmd"
	ShellNuShell    ShellType = "nushell"
)

// DotenvOptions configures the dotenv command behavior
type DotenvOptions struct {
	Files    []string  // .env files to load
	Export   bool      // -e: output as export statements
	Quiet    bool      // -q: suppress warnings
	Override bool      // -o: override existing environment variables
	Expand   bool      // -x: expand variables in values
	Shell    ShellType // -s: target shell for export format
}

// EnvVar represents an environment variable
type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// DetectShell attempts to detect the current shell type from environment
func DetectShell() ShellType {
	// Check SHELL environment variable (Unix)
	shell := os.Getenv("SHELL")
	if shell != "" {
		shell = strings.ToLower(shell)
		switch {
		case strings.Contains(shell, "bash"):
			return ShellBash
		case strings.Contains(shell, "zsh"):
			return ShellZsh
		case strings.Contains(shell, "fish"):
			return ShellFish
		case strings.Contains(shell, "nu"):
			return ShellNuShell
		}
	}

	// Check PSModulePath for PowerShell (works on all platforms)
	if os.Getenv("PSModulePath") != "" {
		return ShellPowerShell
	}

	// Check ComSpec for Windows CMD
	if runtime.GOOS == "windows" {
		comspec := strings.ToLower(os.Getenv("ComSpec"))
		if strings.Contains(comspec, "cmd.exe") {
			// But if we're in PowerShell, PSModulePath would be set
			// So if we get here, we're likely in CMD
			return ShellCmd
		}
	}

	// Default based on OS
	if runtime.GOOS == "windows" {
		return ShellPowerShell
	}

	return ShellBash
}

// FormatExport formats an environment variable for the specified shell
func FormatExport(key, value string, shell ShellType) string {
	switch shell {
	case ShellPowerShell:
		// PowerShell: $env:KEY = "value"
		// Escape double quotes and backticks in value
		escaped := strings.ReplaceAll(value, "`", "``")
		escaped = strings.ReplaceAll(escaped, "\"", "`\"")
		escaped = strings.ReplaceAll(escaped, "$", "`$")

		return fmt.Sprintf("$env:%s = \"%s\"", key, escaped)

	case ShellCmd:
		// CMD: set KEY=value (no quotes around value for set command)
		// Escape special characters
		return fmt.Sprintf("set %s=%s", key, value)

	case ShellFish:
		// Fish: set -gx KEY "value"
		escaped := strings.ReplaceAll(value, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		escaped = strings.ReplaceAll(escaped, "$", "\\$")

		return fmt.Sprintf("set -gx %s \"%s\"", key, escaped)

	case ShellNuShell:
		// NuShell: $env.KEY = "value"
		escaped := strings.ReplaceAll(value, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")

		return fmt.Sprintf("$env.%s = \"%s\"", key, escaped)

	case ShellBash, ShellZsh, ShellAuto:
		fallthrough
	default:
		// Bash/Zsh: export KEY="value"
		escaped := strings.ReplaceAll(value, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		escaped = strings.ReplaceAll(escaped, "$", "\\$")
		escaped = strings.ReplaceAll(escaped, "`", "\\`")

		return fmt.Sprintf("export %s=\"%s\"", key, escaped)
	}
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

	// Determine shell type
	shell := opts.Shell
	if shell == "" || shell == ShellAuto {
		shell = DetectShell()
	}

	// Output or set variables
	for _, v := range allVars {
		if opts.Export {
			_, _ = fmt.Fprintln(w, FormatExport(v.Key, v.Value, shell))
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
