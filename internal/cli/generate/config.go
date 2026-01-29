package generate

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// CobraConfig represents the configuration file structure for cobra generator.
// Compatible with cobra-cli's ~/.cobra.yaml format with extensions.
type CobraConfig struct {
	Author     string `yaml:"author"`
	License    string `yaml:"license"`
	UseViper   bool   `yaml:"useViper"`
	UseService bool   `yaml:"useService"`
	Full       bool   `yaml:"full"`
	Year       int    `yaml:"year,omitempty"`
}

// configPaths returns the list of paths to search for config files
func configPaths() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	return []string{
		filepath.Join(home, ".cobra.yaml"),         // cobra-cli compatible
		filepath.Join(home, ".cobra.yml"),          // alternative extension
		filepath.Join(home, ".omni", "cobra.yaml"), // omni-specific location
		filepath.Join(home, ".omni", "cobra.yml"),
	}
}

// LoadCobraConfig loads the cobra configuration from the first found config file.
// Search order: ~/.cobra.yaml, ~/.cobra.yml, ~/.omni/cobra.yaml, ~/.omni/cobra.yml
func LoadCobraConfig() (*CobraConfig, string, error) {
	for _, path := range configPaths() {
		if _, err := os.Stat(path); err == nil {
			cfg, err := loadConfigFromFile(path)
			if err != nil {
				return nil, path, err
			}

			return cfg, path, nil
		}
	}

	// No config file found, return empty config
	return &CobraConfig{}, "", nil
}

// loadConfigFromFile reads and parses a YAML config file
func loadConfigFromFile(path string) (*CobraConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg CobraConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// MergeWithFlags merges config file values with command-line flags.
// Command-line flags take precedence over config file values.
func (c *CobraConfig) MergeWithFlags(opts *CobraInitOptions, flagsSet map[string]bool) {
	// Only apply config values if the corresponding flag was not explicitly set
	if !flagsSet["author"] && c.Author != "" {
		opts.Author = c.Author
	}

	if !flagsSet["license"] && c.License != "" {
		opts.License = c.License
	}

	if !flagsSet["viper"] && c.UseViper {
		opts.UseViper = c.UseViper
	}

	if !flagsSet["service"] && c.UseService {
		opts.UseService = c.UseService
	}

	if !flagsSet["full"] && c.Full {
		opts.Full = c.Full
	}
}

// DefaultConfigPath returns the default path for creating a new config file
func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".cobra.yaml")
}

// WriteDefaultConfig writes a default configuration file to the specified path
func WriteDefaultConfig(path string, cfg *CobraConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	// Add header comment
	header := `# Cobra Generator Configuration
# This file is used by 'omni generate cobra' command.
# Compatible with cobra-cli's configuration format.
#
# Available options:
#   author: Your Name <email@example.com>
#   license: MIT | Apache-2.0 | BSD-3
#   useViper: true | false
#   useService: true | false (use inovacc/config service pattern)
#   full: true | false (generate full project with CI/CD)
#

`

	return os.WriteFile(path, []byte(header+string(data)), 0644)
}
