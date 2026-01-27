package cli

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

// CryptOptions configures the encrypt/decrypt command behavior
type CryptOptions struct {
	Password     string // -p: password for encryption
	PasswordFile string // -P: read password from file
	KeyFile      string // -k: use key file
	Salt         string // -s: salt for key derivation
	Iterations   int    // -i: PBKDF2 iterations (default 100000)
	Output       string // -o: output file
	Base64       bool   // -b: base64 encode/decode
	Armor        bool   // -a: ASCII armor output (same as -b)
}

const (
	saltSize    = 16
	keySize     = 32 // AES-256
	nonceSize   = 12 // GCM nonce size
	defaultIter = 100000
)

// RunEncrypt encrypts data using AES-256-GCM
func RunEncrypt(w io.Writer, args []string, opts CryptOptions) error {
	password, err := getPassword(opts)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if opts.Iterations <= 0 {
		opts.Iterations = defaultIter
	}

	// Read input
	var input []byte
	if len(args) == 0 || args[0] == "-" {
		input, err = io.ReadAll(os.Stdin)
	} else {
		input, err = os.ReadFile(args[0])
	}

	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	// Generate salt
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("encrypt: failed to generate salt: %w", err)
	}

	// Derive key using PBKDF2
	key := pbkdf2.Key([]byte(password), salt, opts.Iterations, keySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return fmt.Errorf("encrypt: failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nil, nonce, input, nil)

	// Combine: salt + nonce + ciphertext
	output := make([]byte, 0, saltSize+nonceSize+len(ciphertext))
	output = append(output, salt...)
	output = append(output, nonce...)
	output = append(output, ciphertext...)

	// Write output
	var outWriter = w

	if opts.Output != "" {
		f, err := os.Create(opts.Output)
		if err != nil {
			return fmt.Errorf("encrypt: %w", err)
		}

		defer func() {
			_ = f.Close()
		}()

		outWriter = f
	}

	if opts.Base64 || opts.Armor {
		_, _ = fmt.Fprintln(outWriter, base64.StdEncoding.EncodeToString(output))
	} else {
		_, _ = outWriter.Write(output)
	}

	return nil
}

// RunDecrypt decrypts data using AES-256-GCM
func RunDecrypt(w io.Writer, args []string, opts CryptOptions) error {
	password, err := getPassword(opts)
	if err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}

	if opts.Iterations <= 0 {
		opts.Iterations = defaultIter
	}

	// Read input
	var input []byte
	if len(args) == 0 || args[0] == "-" {
		input, err = io.ReadAll(os.Stdin)
	} else {
		input, err = os.ReadFile(args[0])
	}

	if err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}

	// Decode base64 if needed
	if opts.Base64 || opts.Armor {
		input, err = base64.StdEncoding.DecodeString(string(input))
		if err != nil {
			return fmt.Errorf("decrypt: invalid base64: %w", err)
		}
	}

	// Validate minimum length
	minLen := saltSize + nonceSize + 16 // 16 = minimum GCM tag
	if len(input) < minLen {
		return fmt.Errorf("decrypt: input too short")
	}

	// Extract salt, nonce, ciphertext
	salt := input[:saltSize]
	nonce := input[saltSize : saltSize+nonceSize]
	ciphertext := input[saltSize+nonceSize:]

	// Derive key using PBKDF2
	key := pbkdf2.Key([]byte(password), salt, opts.Iterations, keySize, sha256.New)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("decrypt: authentication failed (wrong password?)")
	}

	// Write output
	var outWriter = w

	if opts.Output != "" {
		f, err := os.Create(opts.Output)
		if err != nil {
			return fmt.Errorf("decrypt: %w", err)
		}

		defer func() {
			_ = f.Close()
		}()

		outWriter = f
	}

	_, _ = outWriter.Write(plaintext)

	return nil
}

func getPassword(opts CryptOptions) (string, error) {
	if opts.Password != "" {
		return opts.Password, nil
	}

	if opts.PasswordFile != "" {
		data, err := os.ReadFile(opts.PasswordFile)
		if err != nil {
			return "", fmt.Errorf("cannot read password file: %w", err)
		}
		// Trim newline
		password := string(data)
		if len(password) > 0 && password[len(password)-1] == '\n' {
			password = password[:len(password)-1]
		}

		return password, nil
	}

	if opts.KeyFile != "" {
		data, err := os.ReadFile(opts.KeyFile)
		if err != nil {
			return "", fmt.Errorf("cannot read key file: %w", err)
		}
		// Use file content as password
		return string(data), nil
	}

	// Check environment variable
	if envPass := os.Getenv("omni_PASSWORD"); envPass != "" {
		return envPass, nil
	}

	return "", fmt.Errorf("no password provided (use -p, -P, -k, or omni_PASSWORD)")
}

// GenerateKey generates a random encryption key
func GenerateKey(w io.Writer, size int) error {
	if size <= 0 {
		size = 32
	}

	key := make([]byte, size)
	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	_, _ = fmt.Fprintln(w, base64.StdEncoding.EncodeToString(key))

	return nil
}
