// Package vault provides HashiCorp Vault CLI functionality for omni.
package vault

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
)

// Client wraps the Vault API client.
type Client struct {
	client *api.Client
}

// Options for creating a Vault client.
type Options struct {
	Address   string // Vault server address (default: VAULT_ADDR or https://127.0.0.1:8200)
	Token     string // Auth token (default: VAULT_TOKEN)
	Namespace string // Vault namespace (default: VAULT_NAMESPACE)
	TLSSkip   bool   // Skip TLS verification
}

// New creates a new Vault client with the given options.
func New(opts Options) (*Client, error) {
	config := api.DefaultConfig()

	if opts.Address != "" {
		config.Address = opts.Address
	}

	if opts.TLSSkip {
		if err := config.ConfigureTLS(&api.TLSConfig{Insecure: true}); err != nil {
			return nil, fmt.Errorf("failed to configure TLS: %w", err)
		}
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	if opts.Token != "" {
		client.SetToken(opts.Token)
	}

	if opts.Namespace != "" {
		client.SetNamespace(opts.Namespace)
	}

	return &Client{client: client}, nil
}

// NewFromEnv creates a new Vault client using environment variables.
func NewFromEnv() (*Client, error) {
	return New(Options{})
}

// SetToken sets the authentication token.
func (c *Client) SetToken(token string) {
	c.client.SetToken(token)
}

// Token returns the current token.
func (c *Client) Token() string {
	return c.client.Token()
}

// Address returns the Vault server address.
func (c *Client) Address() string {
	return c.client.Address()
}

// Read reads a secret from the given path.
func (c *Client) Read(ctx context.Context, path string) (*api.Secret, error) {
	secret, err := c.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret: %w", err)
	}

	return secret, nil
}

// Write writes data to the given path.
func (c *Client) Write(ctx context.Context, path string, data map[string]any) (*api.Secret, error) {
	secret, err := c.client.Logical().WriteWithContext(ctx, path, data)
	if err != nil {
		return nil, fmt.Errorf("failed to write secret: %w", err)
	}

	return secret, nil
}

// List lists secrets at the given path.
func (c *Client) List(ctx context.Context, path string) (*api.Secret, error) {
	secret, err := c.client.Logical().ListWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	return secret, nil
}

// Delete deletes a secret at the given path.
func (c *Client) Delete(ctx context.Context, path string) (*api.Secret, error) {
	secret, err := c.client.Logical().DeleteWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to delete secret: %w", err)
	}

	return secret, nil
}

// LoginToken performs token-based login.
func (c *Client) LoginToken(token string) error {
	c.client.SetToken(token)

	// Verify the token by looking it up
	_, err := c.client.Auth().Token().LookupSelf()
	if err != nil {
		return fmt.Errorf("token verification failed: %w", err)
	}

	return nil
}

// LoginUserpass performs userpass authentication.
func (c *Client) LoginUserpass(ctx context.Context, username, password string, mount string) (*api.Secret, error) {
	if mount == "" {
		mount = "userpass"
	}

	path := fmt.Sprintf("auth/%s/login/%s", mount, username)
	data := map[string]any{
		"password": password,
	}

	secret, err := c.client.Logical().WriteWithContext(ctx, path, data)
	if err != nil {
		return nil, fmt.Errorf("userpass login failed: %w", err)
	}

	if secret != nil && secret.Auth != nil {
		c.client.SetToken(secret.Auth.ClientToken)
	}

	return secret, nil
}

// LoginAppRole performs AppRole authentication.
func (c *Client) LoginAppRole(ctx context.Context, roleID, secretID string, mount string) (*api.Secret, error) {
	if mount == "" {
		mount = "approle"
	}

	path := fmt.Sprintf("auth/%s/login", mount)
	data := map[string]any{
		"role_id":   roleID,
		"secret_id": secretID,
	}

	secret, err := c.client.Logical().WriteWithContext(ctx, path, data)
	if err != nil {
		return nil, fmt.Errorf("approle login failed: %w", err)
	}

	if secret != nil && secret.Auth != nil {
		c.client.SetToken(secret.Auth.ClientToken)
	}

	return secret, nil
}

// TokenLookupSelf looks up the current token.
func (c *Client) TokenLookupSelf() (*api.Secret, error) {
	return c.client.Auth().Token().LookupSelf()
}

// TokenRenewSelf renews the current token.
func (c *Client) TokenRenewSelf(increment int) (*api.Secret, error) {
	return c.client.Auth().Token().RenewSelf(increment)
}

// TokenRevokeSelf revokes the current token.
func (c *Client) TokenRevokeSelf() error {
	return c.client.Auth().Token().RevokeSelf("")
}

// Status returns the Vault server status.
func (c *Client) Status() (*api.SealStatusResponse, error) {
	return c.client.Sys().SealStatus()
}

// Health returns the Vault server health.
func (c *Client) Health() (*api.HealthResponse, error) {
	return c.client.Sys().Health()
}

// SaveToken saves the current token to the default token file.
func (c *Client) SaveToken() error {
	token := c.client.Token()
	if token == "" {
		return fmt.Errorf("no token to save")
	}

	tokenFile := getTokenFile()

	return os.WriteFile(tokenFile, []byte(token), 0600)
}

// LoadToken loads the token from the default token file.
func (c *Client) LoadToken() error {
	tokenFile := getTokenFile()

	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return fmt.Errorf("failed to read token file: %w", err)
	}

	c.client.SetToken(string(data))

	return nil
}

func getTokenFile() string {
	if home, err := os.UserHomeDir(); err == nil {
		return home + "/.vault-token"
	}

	return ".vault-token"
}
