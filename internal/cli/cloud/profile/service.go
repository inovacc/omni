package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/inovacc/omni/internal/cli/cloud/config"
	"github.com/inovacc/omni/internal/cli/cloud/crypto"
)

const (
	omniDir = ".omni"
)

// Service provides profile management operations.
type Service struct {
	store     *FileStore
	config    *config.ConfigService
	baseDir   string
	masterKey []byte
}

// NewService creates a new profile service with the default base directory.
func NewService() (*Service, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, omniDir)

	return NewServiceWithDir(baseDir)
}

// NewServiceWithDir creates a new profile service with a custom base directory.
func NewServiceWithDir(baseDir string) (*Service, error) {
	// Ensure base directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("creating base directory: %w", err)
	}

	// Initialize master key
	masterKey, err := crypto.GetOrCreateMasterKey(baseDir)
	if err != nil {
		return nil, fmt.Errorf("initializing master key: %w", err)
	}

	// Initialize config service
	configSvc := config.NewConfigService(baseDir)
	if err := configSvc.Load(); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	return &Service{
		store:     NewFileStore(baseDir),
		config:    configSvc,
		baseDir:   baseDir,
		masterKey: masterKey,
	}, nil
}

// AddProfile creates a new profile with encrypted credentials.
func (s *Service) AddProfile(profile *CloudProfile, creds Credentials) error {
	// Validate provider match
	if creds.Provider() != profile.Provider {
		return fmt.Errorf("credential provider mismatch: got %s, want %s", creds.Provider(), profile.Provider)
	}

	// Check for existing profile
	if s.store.ProfileExists(profile.Provider, profile.Name) {
		return fmt.Errorf("profile already exists: %s/%s", profile.Provider, profile.Name)
	}

	// Set timestamps
	profile.CreatedAt = time.Now()
	profile.TokenStorage = TokenStorageEncrypted

	// Serialize credentials
	credsJSON, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("marshaling credentials: %w", err)
	}

	// Derive profile-specific key
	profileKey := crypto.DeriveProfileKey(s.masterKey, string(profile.Provider), profile.Name)

	// Encrypt credentials
	encrypted, err := crypto.EncryptWithKey(credsJSON, profileKey)
	if err != nil {
		return fmt.Errorf("encrypting credentials: %w", err)
	}

	// Save profile metadata
	if err := s.store.SaveProfile(profile); err != nil {
		return err
	}

	// Save encrypted credentials
	if err := s.store.SaveCredentials(profile.Provider, profile.Name, encrypted); err != nil {
		// Clean up profile if credentials fail to save
		_ = s.store.DeleteProfile(profile.Provider, profile.Name)
		return err
	}

	// Set as default if it's the first profile for this provider
	names, _ := s.store.ListProfiles(profile.Provider)
	if len(names) == 1 || profile.Default {
		if err := s.SetDefault(profile.Provider, profile.Name); err != nil {
			return fmt.Errorf("setting as default: %w", err)
		}
	}

	return nil
}

// GetProfile retrieves a profile's metadata.
func (s *Service) GetProfile(provider Provider, name string) (*CloudProfile, error) {
	return s.store.LoadProfile(provider, name)
}

// GetCredentials retrieves and decrypts credentials for a profile.
func (s *Service) GetCredentials(provider Provider, name string) (Credentials, error) {
	// Load profile to verify it exists
	profile, err := s.store.LoadProfile(provider, name)
	if err != nil {
		return nil, err
	}

	// Load encrypted credentials
	encrypted, err := s.store.LoadCredentials(provider, name)
	if err != nil {
		return nil, err
	}

	// Derive profile-specific key
	profileKey := crypto.DeriveProfileKey(s.masterKey, string(provider), name)

	// Decrypt credentials
	decrypted, err := crypto.DecryptWithKey(encrypted, profileKey)
	if err != nil {
		return nil, fmt.Errorf("decrypting credentials: %w", err)
	}

	// Unmarshal based on provider
	var creds Credentials

	switch profile.Provider {
	case ProviderAWS:
		var awsCreds AWSCredentials
		if err := json.Unmarshal(decrypted, &awsCreds); err != nil {
			return nil, fmt.Errorf("parsing AWS credentials: %w", err)
		}

		creds = &awsCreds

	case ProviderAzure:
		var azureCreds AzureCredentials
		if err := json.Unmarshal(decrypted, &azureCreds); err != nil {
			return nil, fmt.Errorf("parsing Azure credentials: %w", err)
		}

		creds = &azureCreds

	case ProviderGCP:
		var gcpCreds GCPCredentials
		if err := json.Unmarshal(decrypted, &gcpCreds); err != nil {
			return nil, fmt.Errorf("parsing GCP credentials: %w", err)
		}

		creds = &gcpCreds

	default:
		return nil, fmt.Errorf("unknown provider: %s", profile.Provider)
	}

	// Update last used timestamp
	profile.LastUsedAt = time.Now()
	_ = s.store.SaveProfile(profile)

	return creds, nil
}

// GetAWSCredentials retrieves AWS credentials for a profile.
func (s *Service) GetAWSCredentials(name string) (*AWSCredentials, error) {
	creds, err := s.GetCredentials(ProviderAWS, name)
	if err != nil {
		return nil, err
	}

	awsCreds, ok := creds.(*AWSCredentials)
	if !ok {
		return nil, fmt.Errorf("profile is not an AWS profile")
	}

	return awsCreds, nil
}

// ListProfiles returns all profiles for a provider.
func (s *Service) ListProfiles(provider Provider) ([]*CloudProfile, error) {
	return s.store.ListAllProfiles(provider)
}

// ListAllProviderProfiles returns all profiles across all providers.
func (s *Service) ListAllProviderProfiles() (map[Provider][]*CloudProfile, error) {
	result := make(map[Provider][]*CloudProfile)

	for _, provider := range ValidProviders {
		profiles, err := s.store.ListAllProfiles(provider)
		if err != nil {
			continue
		}

		if len(profiles) > 0 {
			result[provider] = profiles
		}
	}

	return result, nil
}

// DeleteProfile removes a profile and its credentials.
func (s *Service) DeleteProfile(provider Provider, name string) error {
	// Check if it's the default profile
	defaultName := s.config.GetDefaultProfile(string(provider))
	if defaultName == name {
		// Clear the default
		if err := s.config.ClearDefaultProfile(string(provider)); err != nil {
			return fmt.Errorf("clearing default profile: %w", err)
		}
	}

	return s.store.DeleteProfile(provider, name)
}

// SetDefault sets a profile as the default for its provider.
func (s *Service) SetDefault(provider Provider, name string) error {
	// Verify profile exists
	if !s.store.ProfileExists(provider, name) {
		return fmt.Errorf("profile not found: %s/%s", provider, name)
	}

	// Update old default profile
	oldDefault := s.config.GetDefaultProfile(string(provider))
	if oldDefault != "" && oldDefault != name {
		oldProfile, err := s.store.LoadProfile(provider, oldDefault)
		if err == nil {
			oldProfile.Default = false
			_ = s.store.SaveProfile(oldProfile)
		}
	}

	// Update new default profile
	profile, err := s.store.LoadProfile(provider, name)
	if err != nil {
		return err
	}

	profile.Default = true
	if err := s.store.SaveProfile(profile); err != nil {
		return err
	}

	// Update config
	return s.config.SetDefaultProfile(string(provider), name)
}

// GetDefault returns the default profile name for a provider.
func (s *Service) GetDefault(provider Provider) string {
	return s.config.GetDefaultProfile(string(provider))
}

// GetDefaultProfile returns the default profile for a provider.
func (s *Service) GetDefaultProfile(provider Provider) (*CloudProfile, error) {
	name := s.GetDefault(provider)
	if name == "" {
		return nil, fmt.Errorf("no default profile for %s", provider)
	}

	return s.store.LoadProfile(provider, name)
}
