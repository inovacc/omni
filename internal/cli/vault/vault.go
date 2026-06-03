// Package vault provides HashiCorp Vault CLI functionality for omni.
package vault

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hashicorp/vault/api"
	"github.com/inovacc/omni/internal/cli/cmderr"
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

// classifyVaultError maps a Vault API error to a cmderr sentinel.
func classifyVaultError(err error, op string) error {
	if err == nil {
		return nil
	}

	var respErr *api.ResponseError
	if errors.As(err, &respErr) {
		switch respErr.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("vault: %s: %v", op, err))
		case http.StatusNotFound:
			return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("vault: %s: %v", op, err))
		case http.StatusBadRequest:
			return cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("vault: %s: %v", op, err))
		default:
			return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("vault: %s: %v", op, err))
		}
	}

	return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("vault: %s: %v", op, err))
}

// New creates a new Vault client with the given options.
func New(opts Options) (*Client, error) {
	config := api.DefaultConfig()

	if opts.Address != "" {
		config.Address = opts.Address
	}

	if opts.TLSSkip {
		// Disabling TLS verification means Vault tokens and secrets traverse an
		// unverified channel, exposing them to MITM interception. Make this loud
		// so an accidentally-inherited insecure profile is visible in logs/stderr.
		slog.Warn("vault: TLS certificate verification disabled (Insecure); tokens and secrets are sent over an unverified channel — do not use against production Vault")
		_, _ = fmt.Fprintln(os.Stderr, "vault: warning: TLS certificate verification disabled (--tls-skip); do not use against production Vault")

		if err := config.ConfigureTLS(&api.TLSConfig{Insecure: true}); err != nil {
			return nil, cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("vault: failed to configure TLS: %v", err))
		}
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, cmderr.Wrap(cmderr.ErrInvalidInput, fmt.Sprintf("vault: failed to create client: %v", err))
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
		return nil, classifyVaultError(err, "read "+path)
	}

	if secret == nil {
		return nil, cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("vault: secret not found at path: %s", path))
	}

	return secret, nil
}

// Write writes data to the given path.
func (c *Client) Write(ctx context.Context, path string, data map[string]any) (*api.Secret, error) {
	secret, err := c.client.Logical().WriteWithContext(ctx, path, data)
	if err != nil {
		return nil, classifyVaultError(err, "write "+path)
	}

	return secret, nil
}

// List lists secrets at the given path.
func (c *Client) List(ctx context.Context, path string) (*api.Secret, error) {
	secret, err := c.client.Logical().ListWithContext(ctx, path)
	if err != nil {
		return nil, classifyVaultError(err, "list "+path)
	}

	return secret, nil
}

// Delete deletes a secret at the given path.
func (c *Client) Delete(ctx context.Context, path string) (*api.Secret, error) {
	secret, err := c.client.Logical().DeleteWithContext(ctx, path)
	if err != nil {
		return nil, classifyVaultError(err, "delete "+path)
	}

	return secret, nil
}

// LoginToken performs token-based login.
func (c *Client) LoginToken(token string) error {
	c.client.SetToken(token)

	// Verify the token by looking it up
	_, err := c.client.Auth().Token().LookupSelf()
	if err != nil {
		return classifyVaultError(err, "token verification")
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
		return nil, classifyVaultError(err, "userpass login")
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
		return nil, classifyVaultError(err, "approle login")
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
		return cmderr.Wrap(cmderr.ErrInvalidInput, "vault: no token to save")
	}

	tokenFile := getTokenFile()

	if err := os.WriteFile(tokenFile, []byte(token), 0600); err != nil {
		if errors.Is(err, os.ErrPermission) {
			return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("vault: save token: %v", err))
		}
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("vault: save token: %v", err))
	}

	return nil
}

// LoadToken loads the token from the default token file.
func (c *Client) LoadToken() error {
	tokenFile := getTokenFile()

	data, err := os.ReadFile(tokenFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("vault: token file not found: %s", tokenFile))
		}
		if errors.Is(err, os.ErrPermission) {
			return cmderr.Wrap(cmderr.ErrPermission, fmt.Sprintf("vault: load token: %v", err))
		}
		return cmderr.Wrap(cmderr.ErrIO, fmt.Sprintf("vault: load token: %v", err))
	}

	c.client.SetToken(string(data))

	return nil
}

func getTokenFile() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".vault-token")
	}

	return ".vault-token"
}
