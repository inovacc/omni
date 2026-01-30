package buf

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Options configures common buf options
type Options struct {
	ErrorFormat string   // Output format: text, json, github-actions
	ExcludePath []string // Paths to exclude
	Path        []string // Specific paths to process
	Config      string   // Custom config file path
}

// LintOptions configures buf lint
type LintOptions struct {
	Options
}

// FormatOptions configures buf format
type FormatOptions struct {
	Write    bool // Rewrite files in place
	Diff     bool // Display diff
	ExitCode bool // Exit with non-zero if files unformatted
}

// BuildOptions configures buf build
type BuildOptions struct {
	Options

	Output string // Output file path
}

// BreakingOptions configures buf breaking
type BreakingOptions struct {
	Options

	Against        string // Source to compare against
	ExcludeImports bool   // Don't check imported files
}

// GenerateOptions configures buf generate
type GenerateOptions struct {
	Template       string // Alternate buf.gen.yaml location
	Output         string // Base output directory
	IncludeImports bool   // Include imported files
}

// LintResult represents a lint issue
type LintResult struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// FormatResult represents format output
type FormatResult struct {
	File      string `json:"file"`
	Formatted bool   `json:"formatted"`
	Diff      string `json:"diff,omitempty"`
	BytesDiff int    `json:"bytes_diff,omitempty"`
}

// BuildResult represents build output
type BuildResult struct {
	Success bool     `json:"success"`
	Files   []string `json:"files"`
	Errors  []string `json:"errors,omitempty"`
}

// Config represents buf.yaml configuration
type Config struct {
	Version  string         `yaml:"version"`
	Name     string         `yaml:"name,omitempty"`
	Deps     []string       `yaml:"deps,omitempty"`
	Lint     LintConfig     `yaml:"lint,omitempty"`
	Breaking BreakingConfig `yaml:"breaking,omitempty"`
}

// LintConfig represents lint configuration
type LintConfig struct {
	Use        []string            `yaml:"use,omitempty"`
	Except     []string            `yaml:"except,omitempty"`
	Ignore     []string            `yaml:"ignore,omitempty"`
	IgnoreOnly map[string][]string `yaml:"ignore_only,omitempty"`
}

// BreakingConfig represents breaking change configuration
type BreakingConfig struct {
	Use    []string `yaml:"use,omitempty"`
	Except []string `yaml:"except,omitempty"`
	Ignore []string `yaml:"ignore,omitempty"`
}

// GenerateConfig represents buf.gen.yaml configuration
type GenerateConfig struct {
	Version string         `yaml:"version"`
	Plugins []PluginConfig `yaml:"plugins"`
	Clean   bool           `yaml:"clean,omitempty"`
	Managed ManagedConfig  `yaml:"managed,omitempty"`
}

// PluginConfig represents a plugin configuration
type PluginConfig struct {
	Remote string   `yaml:"remote,omitempty"`
	Local  string   `yaml:"local,omitempty"`
	Out    string   `yaml:"out"`
	Opt    []string `yaml:"opt,omitempty"`
}

// ManagedConfig represents managed mode configuration
type ManagedConfig struct {
	Enabled         bool            `yaml:"enabled"`
	GoPackagePrefix GoPackagePrefix `yaml:"go_package_prefix,omitempty"`
}

// GoPackagePrefix represents go_package_prefix configuration
type GoPackagePrefix struct {
	Default string `yaml:"default"`
}

// LoadConfig loads buf.yaml from a directory
func LoadConfig(dir string) (*Config, error) {
	configPath := filepath.Join(dir, "buf.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config
			return &Config{
				Version: "v1",
				Lint: LintConfig{
					Use: []string{"STANDARD"},
				},
			}, nil
		}

		return nil, fmt.Errorf("failed to read buf.yaml: %w", err)
	}

	var config Config

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse buf.yaml: %w", err)
	}

	return &config, nil
}

// LoadGenerateConfig loads buf.gen.yaml from a directory
func LoadGenerateConfig(dir string) (*GenerateConfig, error) {
	configPath := filepath.Join(dir, "buf.gen.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read buf.gen.yaml: %w", err)
	}

	var config GenerateConfig

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse buf.gen.yaml: %w", err)
	}

	return &config, nil
}

// FindProtoFiles finds all .proto files in a directory
func FindProtoFiles(dir string, excludePaths []string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Check if directory should be excluded
			relPath, _ := filepath.Rel(dir, path)
			for _, exclude := range excludePaths {
				if strings.HasPrefix(relPath, exclude) || relPath == exclude {
					return filepath.SkipDir
				}
			}

			return nil
		}

		// Check if file is a .proto file
		if !strings.HasSuffix(path, ".proto") {
			return nil
		}

		// Check if file should be excluded
		relPath, _ := filepath.Rel(dir, path)
		for _, exclude := range excludePaths {
			if strings.HasPrefix(relPath, exclude) || relPath == exclude {
				return nil
			}
		}

		files = append(files, path)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

// OutputResults outputs results in the specified format
func OutputResults(w io.Writer, results []LintResult, format string) error {
	switch format {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")

		return enc.Encode(results)

	case "github-actions":
		for _, r := range results {
			_, _ = fmt.Fprintf(w, "::error file=%s,line=%d,col=%d::%s: %s\n",
				r.File, r.Line, r.Column, r.Rule, r.Message)
		}

		return nil

	default: // text
		for _, r := range results {
			_, _ = fmt.Fprintf(w, "%s:%d:%d: %s: %s\n",
				r.File, r.Line, r.Column, r.Rule, r.Message)
		}

		return nil
	}
}
