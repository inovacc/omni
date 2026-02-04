package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestService_AddAndGetProfile(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "omni-profile-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create service
	svc, err := NewServiceWithDir(tmpDir)
	if err != nil {
		t.Fatalf("NewServiceWithDir failed: %v", err)
	}

	// Create profile
	profile := &CloudProfile{
		Name:     "test-profile",
		Provider: ProviderAWS,
		Region:   "us-east-1",
	}

	creds := &AWSCredentials{
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}

	// Add profile
	if err := svc.AddProfile(profile, creds); err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}

	// Get profile
	loaded, err := svc.GetProfile(ProviderAWS, "test-profile")
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}

	if loaded.Name != "test-profile" {
		t.Errorf("name: got %q, want %q", loaded.Name, "test-profile")
	}

	if loaded.Provider != ProviderAWS {
		t.Errorf("provider: got %q, want %q", loaded.Provider, ProviderAWS)
	}

	if loaded.Region != "us-east-1" {
		t.Errorf("region: got %q, want %q", loaded.Region, "us-east-1")
	}

	if loaded.TokenStorage != TokenStorageEncrypted {
		t.Errorf("token_storage: got %q, want %q", loaded.TokenStorage, TokenStorageEncrypted)
	}

	// Get credentials
	loadedCreds, err := svc.GetCredentials(ProviderAWS, "test-profile")
	if err != nil {
		t.Fatalf("GetCredentials failed: %v", err)
	}

	awsCreds, ok := loadedCreds.(*AWSCredentials)
	if !ok {
		t.Fatal("expected AWSCredentials type")
	}

	if awsCreds.AccessKeyID != creds.AccessKeyID {
		t.Errorf("access_key_id: got %q, want %q", awsCreds.AccessKeyID, creds.AccessKeyID)
	}

	if awsCreds.SecretAccessKey != creds.SecretAccessKey {
		t.Errorf("secret_access_key mismatch")
	}
}

func TestService_ListProfiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "omni-profile-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	svc, err := NewServiceWithDir(tmpDir)
	if err != nil {
		t.Fatalf("NewServiceWithDir failed: %v", err)
	}

	// Add multiple profiles
	profiles := []string{"prod", "dev", "staging"}
	for _, name := range profiles {
		profile := &CloudProfile{
			Name:     name,
			Provider: ProviderAWS,
		}

		creds := &AWSCredentials{
			AccessKeyID:     "AKIA" + name,
			SecretAccessKey: "secret-" + name,
		}
		if err := svc.AddProfile(profile, creds); err != nil {
			t.Fatalf("AddProfile(%s) failed: %v", name, err)
		}
	}

	// List profiles
	list, err := svc.ListProfiles(ProviderAWS)
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}

	if len(list) != len(profiles) {
		t.Errorf("list length: got %d, want %d", len(list), len(profiles))
	}
}

func TestService_DeleteProfile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "omni-profile-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	svc, err := NewServiceWithDir(tmpDir)
	if err != nil {
		t.Fatalf("NewServiceWithDir failed: %v", err)
	}

	// Add profile
	profile := &CloudProfile{
		Name:     "to-delete",
		Provider: ProviderAWS,
	}

	creds := &AWSCredentials{
		AccessKeyID:     "AKIATEST",
		SecretAccessKey: "secret",
	}
	if err := svc.AddProfile(profile, creds); err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}

	// Delete profile
	if err := svc.DeleteProfile(ProviderAWS, "to-delete"); err != nil {
		t.Fatalf("DeleteProfile failed: %v", err)
	}

	// Verify deleted
	_, err = svc.GetProfile(ProviderAWS, "to-delete")
	if err == nil {
		t.Error("profile should have been deleted")
	}

	// Verify files are gone
	profilePath := filepath.Join(tmpDir, "profiles", "aws", "to-delete.json")
	credsPath := filepath.Join(tmpDir, "profiles", "aws", "to-delete.enc")

	if _, err := os.Stat(profilePath); !os.IsNotExist(err) {
		t.Error("profile file should have been deleted")
	}

	if _, err := os.Stat(credsPath); !os.IsNotExist(err) {
		t.Error("credentials file should have been deleted")
	}
}

func TestService_SetDefault(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "omni-profile-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	svc, err := NewServiceWithDir(tmpDir)
	if err != nil {
		t.Fatalf("NewServiceWithDir failed: %v", err)
	}

	// Add two profiles
	for _, name := range []string{"profile1", "profile2"} {
		profile := &CloudProfile{
			Name:     name,
			Provider: ProviderAWS,
		}

		creds := &AWSCredentials{
			AccessKeyID:     "AKIA" + name,
			SecretAccessKey: "secret",
		}
		if err := svc.AddProfile(profile, creds); err != nil {
			t.Fatalf("AddProfile(%s) failed: %v", name, err)
		}
	}

	// Set profile2 as default
	if err := svc.SetDefault(ProviderAWS, "profile2"); err != nil {
		t.Fatalf("SetDefault failed: %v", err)
	}

	// Verify default
	defaultName := svc.GetDefault(ProviderAWS)
	if defaultName != "profile2" {
		t.Errorf("default: got %q, want %q", defaultName, "profile2")
	}

	// Verify profile2 has Default=true
	p2, _ := svc.GetProfile(ProviderAWS, "profile2")
	if !p2.Default {
		t.Error("profile2 should have Default=true")
	}

	// Verify profile1 has Default=false
	p1, _ := svc.GetProfile(ProviderAWS, "profile1")
	if p1.Default {
		t.Error("profile1 should have Default=false")
	}
}

func TestService_AddProfile_DuplicateName(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "omni-profile-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	svc, err := NewServiceWithDir(tmpDir)
	if err != nil {
		t.Fatalf("NewServiceWithDir failed: %v", err)
	}

	profile := &CloudProfile{
		Name:     "duplicate",
		Provider: ProviderAWS,
	}
	creds := &AWSCredentials{
		AccessKeyID:     "AKIATEST",
		SecretAccessKey: "secret",
	}

	// First add should succeed
	if err := svc.AddProfile(profile, creds); err != nil {
		t.Fatalf("first AddProfile failed: %v", err)
	}

	// Second add should fail
	err = svc.AddProfile(profile, creds)
	if err == nil {
		t.Error("expected error for duplicate profile name")
	}
}

func TestService_ProviderMismatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "omni-profile-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	svc, err := NewServiceWithDir(tmpDir)
	if err != nil {
		t.Fatalf("NewServiceWithDir failed: %v", err)
	}

	// AWS profile with Azure credentials should fail
	profile := &CloudProfile{
		Name:     "mismatch",
		Provider: ProviderAWS,
	}
	creds := &AzureCredentials{
		TenantID:       "tenant",
		ClientID:       "client",
		ClientSecret:   "secret",
		SubscriptionID: "sub",
	}

	err = svc.AddProfile(profile, creds)
	if err == nil {
		t.Error("expected error for provider mismatch")
	}
}
