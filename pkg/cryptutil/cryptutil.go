package cryptutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

const (
	SaltSize  = 16
	KeySize   = 32 // AES-256
	NonceSize = 12 // GCM nonce size
	// DefaultIter is the default PBKDF2-HMAC-SHA256 iteration count.
	//
	// NOTE: the iteration count is NOT stored in the ciphertext envelope
	// (layout is salt + nonce + ciphertext), so decryption must use the same
	// count the data was sealed with. Raising this default would make every
	// existing ciphertext that relied on the default undecryptable, so it must
	// not be bumped until the envelope persists the iteration count. See
	// docs/quality/HARDENING.md finding crypto-02.
	DefaultIter = 100000
	// MinIter is the minimum PBKDF2 iteration count the package will use.
	// Iteration counts below this floor are clamped up to it so a caller cannot
	// silently derive a trivially brute-forceable key (e.g. WithIterations(1)).
	// It is intentionally equal to DefaultIter so the floor never weakens an
	// existing default-cost ciphertext round-trip. See finding crypto-01.
	MinIter = DefaultIter
)

// Options configures encryption/decryption behavior
type Options struct {
	Iterations int  // PBKDF2 iterations (default 100000)
	Base64     bool // Base64 encode/decode the output/input
}

// Option is a functional option for encryption/decryption
type Option func(*Options)

// WithIterations sets the PBKDF2 iteration count
func WithIterations(n int) Option {
	return func(o *Options) { o.Iterations = n }
}

// WithBase64 enables base64 encoding/decoding
func WithBase64() Option {
	return func(o *Options) { o.Base64 = true }
}

func applyOptions(opts []Option) Options {
	o := Options{Iterations: DefaultIter}
	for _, opt := range opts {
		opt(&o)
	}

	// Enforce a minimum iteration floor. Non-positive values fall back to the
	// default; any positive value below the floor (e.g. WithIterations(1)) is
	// clamped up so a weak KDF cost cannot be selected silently. Because the
	// iteration count is not stored in the envelope, this clamp is applied
	// symmetrically by both Encrypt and Decrypt, keeping round-trips consistent.
	if o.Iterations < MinIter {
		o.Iterations = MinIter
	}

	return o
}

// Encrypt encrypts plaintext using AES-256-GCM with PBKDF2 key derivation.
// Returns salt + nonce + ciphertext as raw bytes, or base64-encoded if WithBase64 is set.
func Encrypt(plaintext []byte, password string, opts ...Option) ([]byte, error) {
	o := applyOptions(opts)

	// Generate salt
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("cryptutil: failed to generate salt: %w", err)
	}

	// Derive key using PBKDF2
	key := pbkdf2.Key([]byte(password), salt, o.Iterations, KeySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("cryptutil: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cryptutil: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("cryptutil: failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Combine: salt + nonce + ciphertext
	output := make([]byte, 0, SaltSize+NonceSize+len(ciphertext))
	output = append(output, salt...)
	output = append(output, nonce...)
	output = append(output, ciphertext...)

	if o.Base64 {
		encoded := base64.StdEncoding.EncodeToString(output)
		return []byte(encoded), nil
	}

	return output, nil
}

// Decrypt decrypts data encrypted by Encrypt using AES-256-GCM with PBKDF2 key derivation.
// Input should be salt + nonce + ciphertext (raw bytes), or base64-encoded if WithBase64 is set.
func Decrypt(data []byte, password string, opts ...Option) ([]byte, error) {
	o := applyOptions(opts)

	input := data

	// Decode base64 if needed
	if o.Base64 {
		var err error

		input, err = base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			return nil, fmt.Errorf("cryptutil: invalid base64: %w", err)
		}
	}

	// Validate minimum length
	minLen := SaltSize + NonceSize + 16 // 16 = minimum GCM tag
	if len(input) < minLen {
		return nil, fmt.Errorf("cryptutil: input too short")
	}

	// Extract salt, nonce, ciphertext
	salt := input[:SaltSize]
	nonce := input[SaltSize : SaltSize+NonceSize]
	ciphertext := input[SaltSize+NonceSize:]

	// Derive key using PBKDF2
	key := pbkdf2.Key([]byte(password), salt, o.Iterations, KeySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("cryptutil: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cryptutil: %w", err)
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("cryptutil: authentication failed (wrong password?)")
	}

	return plaintext, nil
}

// GenerateKey generates a random encryption key of the specified size and returns it base64-encoded.
func GenerateKey(size int) (string, error) {
	if size <= 0 {
		size = 32
	}

	key := make([]byte, size)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("cryptutil: failed to generate key: %w", err)
	}

	return base64.StdEncoding.EncodeToString(key), nil
}

// DeriveKey derives an encryption key from a password and salt using PBKDF2.
// Iteration counts below MinIter are clamped up to it so a caller cannot derive
// a key with a trivially low work factor (see finding crypto-01).
func DeriveKey(password string, salt []byte, iterations, keySize int) []byte {
	if iterations < MinIter {
		iterations = MinIter
	}

	if keySize <= 0 {
		keySize = KeySize
	}

	return pbkdf2.Key([]byte(password), salt, iterations, keySize, sha256.New)
}
