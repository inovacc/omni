package profile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/cloud/crypto"
)

func skipIfNoMachineID(t *testing.T) {
	t.Helper()

	if _, err := crypto.GetMachineID(); err != nil {
		t.Skipf("skipping: %v", err)
	}
}

func newTestService(t *testing.T) (*Service, string) {
	t.Helper()
	skipIfNoMachineID(t)

	tmpDir, err := os.MkdirTemp("", "omni-profile-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	svc, err := NewServiceWithDir(tmpDir)
	if err != nil {
		t.Fatalf("NewServiceWithDir failed: %v", err)
	}

	return svc, tmpDir
}

func TestService_AddAndGetProfile(t *testing.T) {
	svc, _ := newTestService(t)

	profile := &CloudProfile{
		Name:     "test-profile",
		Provider: ProviderAWS,
		Region:   "us-east-1",
	}

	creds := &AWSCredentials{
		AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
		SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}

	if err := svc.AddProfile(profile, creds); err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}

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
	svc, _ := newTestService(t)

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

	list, err := svc.ListProfiles(ProviderAWS)
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}

	if len(list) != len(profiles) {
		t.Errorf("list length: got %d, want %d", len(list), len(profiles))
	}
}

func TestService_DeleteProfile(t *testing.T) {
	svc, tmpDir := newTestService(t)

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

	if err := svc.DeleteProfile(ProviderAWS, "to-delete"); err != nil {
		t.Fatalf("DeleteProfile failed: %v", err)
	}

	_, err := svc.GetProfile(ProviderAWS, "to-delete")
	if err == nil {
		t.Error("profile should have been deleted")
	}

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
	svc, _ := newTestService(t)

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

	if err := svc.SetDefault(ProviderAWS, "profile2"); err != nil {
		t.Fatalf("SetDefault failed: %v", err)
	}

	defaultName := svc.GetDefault(ProviderAWS)
	if defaultName != "profile2" {
		t.Errorf("default: got %q, want %q", defaultName, "profile2")
	}

	p2, _ := svc.GetProfile(ProviderAWS, "profile2")
	if !p2.Default {
		t.Error("profile2 should have Default=true")
	}

	p1, _ := svc.GetProfile(ProviderAWS, "profile1")
	if p1.Default {
		t.Error("profile1 should have Default=false")
	}
}

func TestService_AddProfile_DuplicateName(t *testing.T) {
	svc, _ := newTestService(t)

	profile := &CloudProfile{
		Name:     "duplicate",
		Provider: ProviderAWS,
	}
	creds := &AWSCredentials{
		AccessKeyID:     "AKIATEST",
		SecretAccessKey: "secret",
	}

	if err := svc.AddProfile(profile, creds); err != nil {
		t.Fatalf("first AddProfile failed: %v", err)
	}

	err := svc.AddProfile(profile, creds)
	if err == nil {
		t.Error("expected error for duplicate profile name")
	}
}

func TestService_ProviderMismatch(t *testing.T) {
	svc, _ := newTestService(t)

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

	err := svc.AddProfile(profile, creds)
	if err == nil {
		t.Error("expected error for provider mismatch")
	}

	if err != nil && !strings.Contains(err.Error(), "mismatch") {
		t.Errorf("error should mention mismatch: %v", err)
	}
}
