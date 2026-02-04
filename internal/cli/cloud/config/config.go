// Package config handles global cloud configuration.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configFileName = "config.json"
	configPerms    = 0644
	dirPerms       = 0755
)

// DefaultProfiles maps providers to their default profile names.
type DefaultProfiles map[string]string

// Config holds the global cloud configuration.
type Config struct {
	DefaultProfiles DefaultProfiles `json:"default_profiles"`
}

// ConfigService manages the global configuration.
type ConfigService struct {
	baseDir string
	config  *Config
}

// NewConfigService creates a new ConfigService.
func NewConfigService(baseDir string) *ConfigService {
	return &ConfigService{
		baseDir: baseDir,
		config:  &Config{DefaultProfiles: make(DefaultProfiles)},
	}
}

// configPath returns the path to the config file.
func (s *ConfigService) configPath() string {
	return filepath.Join(s.baseDir, configFileName)
}

// Load reads the configuration from disk.
func (s *ConfigService) Load() error {
	path := s.configPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Initialize with empty config
			s.config = &Config{DefaultProfiles: make(DefaultProfiles)}
			return nil
		}

		return fmt.Errorf("reading config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("parsing config file: %w", err)
	}

	if config.DefaultProfiles == nil {
		config.DefaultProfiles = make(DefaultProfiles)
	}

	s.config = &config

	return nil
}

// Save writes the configuration to disk.
func (s *ConfigService) Save() error {
	// Ensure directory exists
	if err := os.MkdirAll(s.baseDir, dirPerms); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	path := s.configPath()
	if err := os.WriteFile(path, data, configPerms); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// GetDefaultProfile returns the default profile name for a provider.
func (s *ConfigService) GetDefaultProfile(provider string) string {
	if s.config == nil || s.config.DefaultProfiles == nil {
		return ""
	}

	return s.config.DefaultProfiles[provider]
}

// SetDefaultProfile sets the default profile for a provider.
func (s *ConfigService) SetDefaultProfile(provider, name string) error {
	if s.config == nil {
		s.config = &Config{DefaultProfiles: make(DefaultProfiles)}
	}

	if s.config.DefaultProfiles == nil {
		s.config.DefaultProfiles = make(DefaultProfiles)
	}

	s.config.DefaultProfiles[provider] = name

	return s.Save()
}

// ClearDefaultProfile removes the default profile for a provider.
func (s *ConfigService) ClearDefaultProfile(provider string) error {
	if s.config == nil || s.config.DefaultProfiles == nil {
		return nil
	}

	delete(s.config.DefaultProfiles, provider)

	return s.Save()
}
