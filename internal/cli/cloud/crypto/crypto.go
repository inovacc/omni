// Package crypto provides AES-256-GCM encryption for cloud profile credentials.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

const (
	// KeySize is the AES-256 key size in bytes.
	KeySize = 32
	// NonceSize is the GCM nonce size in bytes.
	NonceSize = 12
	// TagSize is the GCM authentication tag size in bytes.
	TagSize = 16
	// MinEncryptedSize is the minimum size of encrypted data (nonce + tag).
	MinEncryptedSize = NonceSize + TagSize

	// PrefixEncrypted indicates encrypted data.
	PrefixEncrypted = "ENC:"
	// PrefixOpen indicates unencrypted data (fallback mode).
	PrefixOpen = "OPEN:"
)

// EncryptWithKey encrypts data using AES-256-GCM with the provided key.
// Returns: [12-byte nonce][ciphertext][16-byte GCM tag]
func EncryptWithKey(data, key []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("invalid key size: got %d, want %d", len(key), KeySize)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generating nonce: %w", err)
	}

	// Seal appends the ciphertext and tag to nonce
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	return ciphertext, nil
}

// DecryptWithKey decrypts data using AES-256-GCM with the provided key.
// Expects input format: [12-byte nonce][ciphertext][16-byte GCM tag]
func DecryptWithKey(encrypted, key []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("invalid key size: got %d, want %d", len(key), KeySize)
	}

	if len(encrypted) < MinEncryptedSize {
		return nil, fmt.Errorf("encrypted data too short: got %d, want at least %d", len(encrypted), MinEncryptedSize)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	nonce := encrypted[:NonceSize]
	ciphertext := encrypted[NonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (wrong key?): %w", err)
	}

	return plaintext, nil
}

// GenerateKey generates a random 32-byte encryption key.
func GenerateKey() ([]byte, error) {
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generating key: %w", err)
	}

	return key, nil
}
