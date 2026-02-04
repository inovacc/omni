package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	profilesDir     = "profiles"
	profileExt      = ".json"
	credentialsExt  = ".enc"
	profilePerms    = 0644
	credentialPerms = 0600
	dirPerms        = 0755
)

// FileStore handles file-based storage for cloud profiles.
type FileStore struct {
	baseDir string
}

// NewFileStore creates a new FileStore with the given base directory.
func NewFileStore(baseDir string) *FileStore {
	return &FileStore{baseDir: baseDir}
}

// profileDir returns the directory for a provider's profiles.
func (s *FileStore) profileDir(provider Provider) string {
	return filepath.Join(s.baseDir, profilesDir, string(provider))
}

// profilePath returns the path for a profile's metadata file.
func (s *FileStore) profilePath(provider Provider, name string) string {
	return filepath.Join(s.profileDir(provider), name+profileExt)
}

// credentialsPath returns the path for a profile's encrypted credentials file.
func (s *FileStore) credentialsPath(provider Provider, name string) string {
	return filepath.Join(s.profileDir(provider), name+credentialsExt)
}

// ensureDir creates the directory for a provider if it doesn't exist.
func (s *FileStore) ensureDir(provider Provider) error {
	dir := s.profileDir(provider)
	if err := os.MkdirAll(dir, dirPerms); err != nil {
		return fmt.Errorf("creating profile directory: %w", err)
	}

	return nil
}

// SaveProfile saves a profile's metadata to disk.
func (s *FileStore) SaveProfile(profile *CloudProfile) error {
	if err := s.ensureDir(profile.Provider); err != nil {
		return err
	}

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling profile: %w", err)
	}

	path := s.profilePath(profile.Provider, profile.Name)
	if err := os.WriteFile(path, data, profilePerms); err != nil {
		return fmt.Errorf("writing profile file: %w", err)
	}

	return nil
}

// SaveCredentials saves encrypted credentials to disk.
func (s *FileStore) SaveCredentials(provider Provider, name string, encrypted []byte) error {
	if err := s.ensureDir(provider); err != nil {
		return err
	}

	path := s.credentialsPath(provider, name)
	if err := os.WriteFile(path, encrypted, credentialPerms); err != nil {
		return fmt.Errorf("writing credentials file: %w", err)
	}

	return nil
}

// LoadProfile loads a profile's metadata from disk.
func (s *FileStore) LoadProfile(provider Provider, name string) (*CloudProfile, error) {
	path := s.profilePath(provider, name)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("profile not found: %s/%s", provider, name)
		}

		return nil, fmt.Errorf("reading profile file: %w", err)
	}

	var profile CloudProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("parsing profile: %w", err)
	}

	return &profile, nil
}

// LoadCredentials loads encrypted credentials from disk.
func (s *FileStore) LoadCredentials(provider Provider, name string) ([]byte, error) {
	path := s.credentialsPath(provider, name)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("credentials not found: %s/%s", provider, name)
		}

		return nil, fmt.Errorf("reading credentials file: %w", err)
	}

	return data, nil
}

// ListProfiles returns all profile names for a provider.
func (s *FileStore) ListProfiles(provider Provider) ([]string, error) {
	dir := s.profileDir(provider)

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No profiles yet
		}

		return nil, fmt.Errorf("reading profile directory: %w", err)
	}

	var names []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if before, ok := strings.CutSuffix(entry.Name(), profileExt); ok {
			name := before
			names = append(names, name)
		}
	}

	return names, nil
}

// ListAllProfiles returns all profiles for a provider.
func (s *FileStore) ListAllProfiles(provider Provider) ([]*CloudProfile, error) {
	names, err := s.ListProfiles(provider)
	if err != nil {
		return nil, err
	}

	profiles := make([]*CloudProfile, 0, len(names))
	for _, name := range names {
		profile, err := s.LoadProfile(provider, name)
		if err != nil {
			continue // Skip invalid profiles
		}

		profiles = append(profiles, profile)
	}

	return profiles, nil
}

// DeleteProfile removes a profile and its credentials from disk.
func (s *FileStore) DeleteProfile(provider Provider, name string) error {
	profilePath := s.profilePath(provider, name)
	credsPath := s.credentialsPath(provider, name)

	// Remove profile file
	if err := os.Remove(profilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing profile file: %w", err)
	}

	// Remove credentials file
	if err := os.Remove(credsPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing credentials file: %w", err)
	}

	return nil
}

// ProfileExists checks if a profile exists.
func (s *FileStore) ProfileExists(provider Provider, name string) bool {
	path := s.profilePath(provider, name)
	_, err := os.Stat(path)

	return err == nil
}
