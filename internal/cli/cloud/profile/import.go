package profile

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ImportOptions configures the import behavior.
type ImportOptions struct {
	SourceProfile string // Source profile name (for AWS)
	TargetName    string // Target omni profile name
	SetDefault    bool   // Set as default after import
}

// AWSImporter handles importing AWS credentials.
type AWSImporter struct {
	homeDir string
}

// NewAWSImporter creates a new AWS importer.
func NewAWSImporter() (*AWSImporter, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home directory: %w", err)
	}

	return &AWSImporter{homeDir: home}, nil
}

// ListProfiles returns available AWS profiles from ~/.aws/credentials.
func (i *AWSImporter) ListProfiles() ([]string, error) {
	credsPath := filepath.Join(i.homeDir, ".aws", "credentials")
	return i.parseINIProfiles(credsPath)
}

// Import imports an AWS profile.
func (i *AWSImporter) Import(opts ImportOptions) (*CloudProfile, *AWSCredentials, error) {
	sourceProfile := opts.SourceProfile
	if sourceProfile == "" {
		sourceProfile = "default"
	}

	targetName := opts.TargetName
	if targetName == "" {
		targetName = sourceProfile
	}

	// Parse credentials file
	credsPath := filepath.Join(i.homeDir, ".aws", "credentials")

	creds, err := i.parseAWSCredentials(credsPath, sourceProfile)
	if err != nil {
		return nil, nil, err
	}

	// Parse config file for region
	configPath := filepath.Join(i.homeDir, ".aws", "config")
	region := i.parseAWSConfigRegion(configPath, sourceProfile)

	profile := &CloudProfile{
		Name:     targetName,
		Provider: ProviderAWS,
		Region:   region,
		Default:  opts.SetDefault,
	}

	return profile, creds, nil
}

func (i *AWSImporter) parseINIProfiles(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("AWS credentials file not found: %s", path)
		}

		return nil, err
	}

	defer func() { _ = file.Close() }()

	var profiles []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			profile := strings.TrimPrefix(strings.TrimSuffix(line, "]"), "[")
			profiles = append(profiles, profile)
		}
	}

	return profiles, scanner.Err()
}

func (i *AWSImporter) parseAWSCredentials(path, profileName string) (*AWSCredentials, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("AWS credentials file not found: %s", path)
		}

		return nil, err
	}

	defer func() { _ = file.Close() }()

	creds := &AWSCredentials{}
	inProfile := false
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Check for profile header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			profile := strings.TrimPrefix(strings.TrimSuffix(line, "]"), "[")
			inProfile = (profile == profileName)

			continue
		}

		if !inProfile {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "aws_access_key_id":
			creds.AccessKeyID = value
		case "aws_secret_access_key":
			creds.SecretAccessKey = value
		case "aws_session_token":
			creds.SessionToken = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if creds.AccessKeyID == "" || creds.SecretAccessKey == "" {
		return nil, fmt.Errorf("profile '%s' not found or missing credentials in %s", profileName, path)
	}

	return creds, nil
}

func (i *AWSImporter) parseAWSConfigRegion(path, profileName string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}

	defer func() { _ = file.Close() }()

	// In config file, non-default profiles are prefixed with "profile "
	targetSection := profileName
	if profileName != "default" {
		targetSection = "profile " + profileName
	}

	inProfile := false
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := strings.TrimPrefix(strings.TrimSuffix(line, "]"), "[")
			inProfile = (section == targetSection)

			continue
		}

		if !inProfile {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "region" {
			return value
		}
	}

	return ""
}

// GCPImporter handles importing GCP credentials.
type GCPImporter struct {
	homeDir string
}

// NewGCPImporter creates a new GCP importer.
func NewGCPImporter() (*GCPImporter, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home directory: %w", err)
	}

	return &GCPImporter{homeDir: home}, nil
}

// ListSources returns available GCP credential sources.
func (i *GCPImporter) ListSources() []string {
	var sources []string

	// Check for ADC
	adcPath := i.getADCPath()
	if _, err := os.Stat(adcPath); err == nil {
		sources = append(sources, "application_default_credentials")
	}

	// Check for service account in common locations
	saPath := filepath.Join(i.homeDir, ".config", "gcloud", "service_account.json")
	if _, err := os.Stat(saPath); err == nil {
		sources = append(sources, "service_account")
	}

	return sources
}

// Import imports GCP credentials.
func (i *GCPImporter) Import(opts ImportOptions) (*CloudProfile, *GCPCredentials, error) {
	source := opts.SourceProfile
	if source == "" {
		source = "application_default_credentials"
	}

	targetName := opts.TargetName
	if targetName == "" {
		targetName = "default"
	}

	var credPath string

	switch source {
	case "application_default_credentials", "adc":
		credPath = i.getADCPath()
	case "service_account":
		credPath = filepath.Join(i.homeDir, ".config", "gcloud", "service_account.json")
	default:
		// Treat as a file path
		credPath = source
	}

	data, err := os.ReadFile(credPath)
	if err != nil {
		return nil, nil, fmt.Errorf("reading credentials: %w", err)
	}

	// Check if it's ADC format (has refresh_token) or service account format
	var rawCreds map[string]any
	if err := json.Unmarshal(data, &rawCreds); err != nil {
		return nil, nil, fmt.Errorf("parsing credentials: %w", err)
	}

	// ADC has "type": "authorized_user", service account has "type": "service_account"
	credType, _ := rawCreds["type"].(string)

	if credType == "authorized_user" {
		return nil, nil, fmt.Errorf("application default credentials (authorized_user) cannot be migrated; use a service account key instead")
	}

	if credType != "service_account" {
		return nil, nil, fmt.Errorf("unsupported credential type: %s", credType)
	}

	var creds GCPCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, nil, fmt.Errorf("parsing service account: %w", err)
	}

	profile := &CloudProfile{
		Name:      targetName,
		Provider:  ProviderGCP,
		AccountID: creds.ProjectID,
		Default:   opts.SetDefault,
	}

	return profile, &creds, nil
}

func (i *GCPImporter) getADCPath() string {
	// Check GOOGLE_APPLICATION_CREDENTIALS first
	if path := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); path != "" {
		return path
	}

	// Default ADC location
	if os.Getenv("APPDATA") != "" {
		// Windows
		return filepath.Join(os.Getenv("APPDATA"), "gcloud", "application_default_credentials.json")
	}

	// Unix
	return filepath.Join(i.homeDir, ".config", "gcloud", "application_default_credentials.json")
}

// AzureImporter handles importing Azure credentials.
type AzureImporter struct {
	homeDir string
}

// NewAzureImporter creates a new Azure importer.
func NewAzureImporter() (*AzureImporter, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home directory: %w", err)
	}

	return &AzureImporter{homeDir: home}, nil
}

// ListSubscriptions returns available Azure subscriptions from profile.
func (i *AzureImporter) ListSubscriptions() ([]AzureSubscriptionInfo, error) {
	profilePath := filepath.Join(i.homeDir, ".azure", "azureProfile.json")

	data, err := os.ReadFile(profilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("Azure profile not found: %s (run 'az login' first)", profilePath)
		}

		return nil, err
	}

	var profile azureProfileFile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("parsing Azure profile: %w", err)
	}

	var subs []AzureSubscriptionInfo
	for _, sub := range profile.Subscriptions {
		subs = append(subs, AzureSubscriptionInfo{
			ID:        sub.ID,
			Name:      sub.Name,
			TenantID:  sub.TenantID,
			IsDefault: sub.IsDefault,
		})
	}

	return subs, nil
}

// AzureSubscriptionInfo contains subscription details.
type AzureSubscriptionInfo struct {
	ID        string
	Name      string
	TenantID  string
	IsDefault bool
}

type azureProfileFile struct {
	Subscriptions []struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		TenantID  string `json:"tenantId"`
		IsDefault bool   `json:"isDefault"`
	} `json:"subscriptions"`
}

// Import imports Azure credentials from a service principal file or prompts for manual entry.
// Note: Azure CLI credentials cannot be directly migrated as they use token-based auth.
func (i *AzureImporter) Import(opts ImportOptions) (*CloudProfile, *AzureCredentials, error) {
	// Check for service principal JSON file
	spPath := opts.SourceProfile
	if spPath == "" {
		// Try common locations
		locations := []string{
			filepath.Join(i.homeDir, ".azure", "service_principal.json"),
			filepath.Join(i.homeDir, ".azure", "sp.json"),
		}
		for _, loc := range locations {
			if _, err := os.Stat(loc); err == nil {
				spPath = loc
				break
			}
		}
	}

	if spPath == "" {
		return nil, nil, fmt.Errorf("no service principal file found; Azure CLI credentials cannot be migrated directly\n\nCreate a service principal file with:\n  az ad sp create-for-rbac --name omni-sp --sdk-auth > ~/.azure/service_principal.json")
	}

	data, err := os.ReadFile(spPath)
	if err != nil {
		return nil, nil, fmt.Errorf("reading service principal file: %w", err)
	}

	// Azure SDK auth format
	var sp struct {
		ClientID       string `json:"clientId"`
		ClientSecret   string `json:"clientSecret"`
		TenantID       string `json:"tenantId"`
		SubscriptionID string `json:"subscriptionId"`
	}

	if err := json.Unmarshal(data, &sp); err != nil {
		return nil, nil, fmt.Errorf("parsing service principal: %w", err)
	}

	if sp.ClientID == "" || sp.ClientSecret == "" {
		return nil, nil, fmt.Errorf("invalid service principal file: missing clientId or clientSecret")
	}

	targetName := opts.TargetName
	if targetName == "" {
		targetName = "default"
	}

	profile := &CloudProfile{
		Name:      targetName,
		Provider:  ProviderAzure,
		AccountID: sp.SubscriptionID,
		Default:   opts.SetDefault,
	}

	creds := &AzureCredentials{
		TenantID:       sp.TenantID,
		ClientID:       sp.ClientID,
		ClientSecret:   sp.ClientSecret,
		SubscriptionID: sp.SubscriptionID,
	}

	return profile, creds, nil
}
