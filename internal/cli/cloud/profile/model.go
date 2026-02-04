// Package profile provides cloud profile management for AWS, Azure, and GCP.
package profile

import (
	"time"
)

// Provider represents a cloud provider type.
type Provider string

const (
	// ProviderAWS represents Amazon Web Services.
	ProviderAWS Provider = "aws"
	// ProviderAzure represents Microsoft Azure.
	ProviderAzure Provider = "azure"
	// ProviderGCP represents Google Cloud Platform.
	ProviderGCP Provider = "gcp"
)

// ValidProviders lists all valid cloud providers.
var ValidProviders = []Provider{ProviderAWS, ProviderAzure, ProviderGCP}

// IsValidProvider checks if the provider string is valid.
func IsValidProvider(p string) bool {
	for _, valid := range ValidProviders {
		if string(valid) == p {
			return true
		}
	}

	return false
}

// TokenStorage represents how credentials are stored.
type TokenStorage string

const (
	// TokenStorageEncrypted stores credentials with AES-256-GCM encryption.
	TokenStorageEncrypted TokenStorage = "encrypted"
	// TokenStorageOpen stores credentials without encryption (fallback).
	TokenStorageOpen TokenStorage = "open"
)

// CloudProfile represents a cloud provider profile without credentials.
type CloudProfile struct {
	Name         string       `json:"name"`
	Provider     Provider     `json:"provider"`
	Region       string       `json:"region,omitempty"`
	AccountID    string       `json:"account_id,omitempty"`
	RoleArn      string       `json:"role_arn,omitempty"`
	TokenStorage TokenStorage `json:"token_storage"`
	Default      bool         `json:"default"`
	CreatedAt    time.Time    `json:"created_at"`
	LastUsedAt   time.Time    `json:"last_used_at,omitempty"`
}

// AWSCredentials holds AWS access credentials.
type AWSCredentials struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	SessionToken    string `json:"session_token,omitempty"`
}

// AzureCredentials holds Azure service principal credentials.
type AzureCredentials struct {
	TenantID       string `json:"tenant_id"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	SubscriptionID string `json:"subscription_id"`
}

// GCPCredentials holds GCP service account credentials.
type GCPCredentials struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

// Credentials is an interface for credential types.
type Credentials interface {
	Provider() Provider
}

// Provider returns the provider type for AWS credentials.
func (c *AWSCredentials) Provider() Provider {
	return ProviderAWS
}

// Provider returns the provider type for Azure credentials.
func (c *AzureCredentials) Provider() Provider {
	return ProviderAzure
}

// Provider returns the provider type for GCP credentials.
func (c *GCPCredentials) Provider() Provider {
	return ProviderGCP
}
