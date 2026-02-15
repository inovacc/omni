package crypt

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/cryptutil"
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

// RunEncrypt encrypts data using AES-256-GCM
func RunEncrypt(w io.Writer, args []string, opts CryptOptions) error {
	password, err := getPassword(opts)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	// Read input
	var input []byte
	if len(args) == 0 || args[0] == "-" {
		input, err = io.ReadAll(os.Stdin)
	} else {
		input, err = os.ReadFile(args[0])
	}

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("encrypt: %s", args[0]))
		}
		return fmt.Errorf("encrypt: %w", err)
	}

	// Build options for cryptutil
	var cryptOpts []cryptutil.Option
	if opts.Iterations > 0 {
		cryptOpts = append(cryptOpts, cryptutil.WithIterations(opts.Iterations))
	}

	if opts.Base64 || opts.Armor {
		cryptOpts = append(cryptOpts, cryptutil.WithBase64())
	}

	// Encrypt using pkg/cryptutil
	output, err := cryptutil.Encrypt(input, password, cryptOpts...)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

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
		_, _ = fmt.Fprintln(outWriter, string(output))
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

	// Read input
	var input []byte
	if len(args) == 0 || args[0] == "-" {
		input, err = io.ReadAll(os.Stdin)
	} else {
		input, err = os.ReadFile(args[0])
	}

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cmderr.Wrap(cmderr.ErrNotFound, fmt.Sprintf("decrypt: %s", args[0]))
		}
		return fmt.Errorf("decrypt: %w", err)
	}

	// Build options for cryptutil
	var cryptOpts []cryptutil.Option
	if opts.Iterations > 0 {
		cryptOpts = append(cryptOpts, cryptutil.WithIterations(opts.Iterations))
	}

	if opts.Base64 || opts.Armor {
		cryptOpts = append(cryptOpts, cryptutil.WithBase64())
	}

	// Trim whitespace from base64 input
	if opts.Base64 || opts.Armor {
		input = []byte(strings.TrimSpace(string(input)))
	}

	// Decrypt using pkg/cryptutil
	plaintext, err := cryptutil.Decrypt(input, password, cryptOpts...)
	if err != nil {
		return fmt.Errorf("decrypt: %w", err)
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

	return "", cmderr.Wrap(cmderr.ErrInvalidInput, "no password provided (use -p, -P, -k, or omni_PASSWORD)")
}

// GenerateKey generates a random encryption key
func GenerateKey(w io.Writer, size int) error {
	key, err := cryptutil.GenerateKey(size)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(w, key)

	return nil
}
