package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

const (
	profilesDir    = "profiles"
	profileExt     = ".json"
	credentialsExt = ".enc"
	// profilePerms is 0600: profile metadata describes per-user cloud accounts
	// (account_id, role_arn, region) and must not be world-readable on shared hosts.
	profilePerms    = 0600
	credentialPerms = 0600
	// dirPerms is 0700: the profiles tree is per-user state and other local users
	// must not be able to enumerate which profiles exist.
	dirPerms = 0700
)

// ErrInvalidName indicates a profile or provider name that is not a single safe
// path element. Wrapping cmderr.ErrInvalidInput maps it to exit code 2.
var ErrInvalidName = cmderr.Wrap(cmderr.ErrInvalidInput, "cloud profile: name must be a single path element without separators or '..'")

// validNamePattern restricts profile and provider names to a single safe path
// element so they cannot traverse outside the profile store when joined into a path.
var validNamePattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// validateName rejects names that are empty, equal to "." or "..", contain a path
// separator, or contain characters outside the safe set. This prevents path
// traversal (CWE-22) when name/provider are joined into the store path.
func validateName(name string) error {
	if name == "" || name == "." || name == ".." {
		return ErrInvalidName
	}

	if strings.ContainsRune(name, '/') || strings.ContainsRune(name, os.PathSeparator) {
		return ErrInvalidName
	}

	if !validNamePattern.MatchString(name) {
		return ErrInvalidName
	}

	return nil
}

// validatePathArgs validates both the provider and profile name components that
// are joined into the on-disk path.
func validatePathArgs(provider Provider, name string) error {
	if err := validateName(string(provider)); err != nil {
		return err
	}

	return validateName(name)
}

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
	if err := validatePathArgs(profile.Provider, profile.Name); err != nil {
		return err
	}

	if err := s.ensureDir(profile.Provider); err != nil {
		return err
	}

	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling profile: %w", err)
	}

	path := s.profilePath(profile.Provider, profile.Name)

	return atomicWriteFile(path, data, profilePerms)
}

// atomicWriteFile writes data to path atomically: it writes to a temporary file
// in the SAME directory (so the final os.Rename stays on one filesystem and is
// atomic), fsyncs and closes it, then renames it over path. A crash mid-write
// leaves either the old file intact or the new file fully in place, never a
// truncated/corrupt file. The temp file is created and chmod'd to perms (0600)
// before any data is written so the profile metadata is never momentarily
// world-readable, and it is removed if any step before the rename fails.
func atomicWriteFile(path string, data []byte, perms os.FileMode) error {
	dir := filepath.Dir(path)

	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp profile file: %w", err)
	}

	tmpName := tmp.Name()

	// Ensure the temp file is cleaned up on any failure before the rename
	// succeeds. After a successful rename tmpName no longer exists, so the
	// Remove is a harmless no-op.
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}

	// os.CreateTemp creates the file with 0600, but chmod explicitly so the
	// 0600 invariant holds regardless of umask or platform defaults.
	if err := tmp.Chmod(perms); err != nil {
		cleanup()
		return fmt.Errorf("setting temp profile permissions: %w", err)
	}

	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("writing temp profile file: %w", err)
	}

	if err := tmp.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("syncing temp profile file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("closing temp profile file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("renaming temp profile file into place: %w", err)
	}

	return nil
}

// SaveCredentials saves encrypted credentials to disk.
func (s *FileStore) SaveCredentials(provider Provider, name string, encrypted []byte) error {
	if err := validatePathArgs(provider, name); err != nil {
		return err
	}

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
	if err := validatePathArgs(provider, name); err != nil {
		return nil, err
	}

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
	if err := validatePathArgs(provider, name); err != nil {
		return nil, err
	}

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
	if err := validatePathArgs(provider, name); err != nil {
		return err
	}

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
	if err := validatePathArgs(provider, name); err != nil {
		return false
	}

	path := s.profilePath(provider, name)
	_, err := os.Stat(path)

	return err == nil
}
