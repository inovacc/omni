package vault

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

// KVClient wraps the Vault KV v2 API.
type KVClient struct {
	client *api.Client
	mount  string
}

// KVOptions for KV operations.
type KVOptions struct {
	Mount   string // KV mount path (default: "secret")
	Version int    // Version to read (0 = latest)
}

// NewKV creates a new KV client.
func (c *Client) NewKV(mount string) *KVClient {
	if mount == "" {
		mount = "secret"
	}

	return &KVClient{
		client: c.client,
		mount:  mount,
	}
}

// Get retrieves a secret from the KV store.
func (kv *KVClient) Get(ctx context.Context, path string, version int) (*api.KVSecret, error) {
	kvv2 := kv.client.KVv2(kv.mount)

	var secret *api.KVSecret

	var err error

	if version > 0 {
		secret, err = kvv2.GetVersion(ctx, path, version)
	} else {
		secret, err = kvv2.Get(ctx, path)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	return secret, nil
}

// Put writes a secret to the KV store.
func (kv *KVClient) Put(ctx context.Context, path string, data map[string]any) (*api.KVSecret, error) {
	kvv2 := kv.client.KVv2(kv.mount)

	secret, err := kvv2.Put(ctx, path, data)
	if err != nil {
		return nil, fmt.Errorf("failed to put secret: %w", err)
	}

	return secret, nil
}

// Patch merges data into an existing secret.
func (kv *KVClient) Patch(ctx context.Context, path string, data map[string]any) (*api.KVSecret, error) {
	kvv2 := kv.client.KVv2(kv.mount)

	secret, err := kvv2.Patch(ctx, path, data)
	if err != nil {
		return nil, fmt.Errorf("failed to patch secret: %w", err)
	}

	return secret, nil
}

// Delete deletes a secret (soft delete in v2).
func (kv *KVClient) Delete(ctx context.Context, path string) error {
	kvv2 := kv.client.KVv2(kv.mount)

	err := kvv2.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}

// DeleteVersions deletes specific versions of a secret.
func (kv *KVClient) DeleteVersions(ctx context.Context, path string, versions []int) error {
	kvv2 := kv.client.KVv2(kv.mount)

	err := kvv2.DeleteVersions(ctx, path, versions)
	if err != nil {
		return fmt.Errorf("failed to delete versions: %w", err)
	}

	return nil
}

// Undelete restores deleted versions of a secret.
func (kv *KVClient) Undelete(ctx context.Context, path string, versions []int) error {
	kvv2 := kv.client.KVv2(kv.mount)

	err := kvv2.Undelete(ctx, path, versions)
	if err != nil {
		return fmt.Errorf("failed to undelete secret: %w", err)
	}

	return nil
}

// Destroy permanently destroys versions of a secret.
func (kv *KVClient) Destroy(ctx context.Context, path string, versions []int) error {
	kvv2 := kv.client.KVv2(kv.mount)

	err := kvv2.Destroy(ctx, path, versions)
	if err != nil {
		return fmt.Errorf("failed to destroy secret: %w", err)
	}

	return nil
}

// List lists secrets at the given path.
func (kv *KVClient) List(ctx context.Context, path string) ([]string, error) {
	// KV v2 stores metadata at mount/metadata/path
	listPath := fmt.Sprintf("%s/metadata/%s", kv.mount, path)

	secret, err := kv.client.Logical().ListWithContext(ctx, listPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	if secret == nil || secret.Data == nil {
		return nil, nil
	}

	keysRaw, ok := secret.Data["keys"]
	if !ok {
		return nil, nil
	}

	keysSlice, ok := keysRaw.([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected keys format")
	}

	keys := make([]string, 0, len(keysSlice))

	for _, k := range keysSlice {
		if s, ok := k.(string); ok {
			keys = append(keys, s)
		}
	}

	return keys, nil
}

// GetMetadata retrieves metadata for a secret.
func (kv *KVClient) GetMetadata(ctx context.Context, path string) (*api.KVMetadata, error) {
	kvv2 := kv.client.KVv2(kv.mount)

	metadata, err := kvv2.GetMetadata(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	return metadata, nil
}

// PutMetadata updates metadata for a secret.
func (kv *KVClient) PutMetadata(ctx context.Context, path string, metadata api.KVMetadataPutInput) error {
	kvv2 := kv.client.KVv2(kv.mount)

	err := kvv2.PutMetadata(ctx, path, metadata)
	if err != nil {
		return fmt.Errorf("failed to put metadata: %w", err)
	}

	return nil
}

// DeleteMetadata permanently deletes all versions and metadata for a secret.
func (kv *KVClient) DeleteMetadata(ctx context.Context, path string) error {
	kvv2 := kv.client.KVv2(kv.mount)

	err := kvv2.DeleteMetadata(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	return nil
}

// Rollback rolls back to a previous version of a secret.
func (kv *KVClient) Rollback(ctx context.Context, path string, version int) (*api.KVSecret, error) {
	kvv2 := kv.client.KVv2(kv.mount)

	secret, err := kvv2.Rollback(ctx, path, version)
	if err != nil {
		return nil, fmt.Errorf("failed to rollback secret: %w", err)
	}

	return secret, nil
}
