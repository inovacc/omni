package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "crypt_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("encrypt and decrypt file", func(t *testing.T) {
		plaintext := "Hello, World! This is a secret message."
		password := "supersecretpassword123"

		inputFile := filepath.Join(tmpDir, "plain.txt")
		encryptedFile := filepath.Join(tmpDir, "encrypted.txt")

		if err := os.WriteFile(inputFile, []byte(plaintext), 0644); err != nil {
			t.Fatal(err)
		}

		// Encrypt
		var encBuf bytes.Buffer

		err := RunEncrypt(&encBuf, []string{inputFile}, CryptOptions{
			Password: password,
			Armor:    true,
		})
		if err != nil {
			t.Fatalf("RunEncrypt() error = %v", err)
		}

		encrypted := encBuf.String()
		if encrypted == plaintext {
			t.Error("RunEncrypt() output should not equal plaintext")
		}

		// Write encrypted content to file
		if err := os.WriteFile(encryptedFile, []byte(encrypted), 0644); err != nil {
			t.Fatal(err)
		}

		// Decrypt
		var decBuf bytes.Buffer

		err = RunDecrypt(&decBuf, []string{encryptedFile}, CryptOptions{
			Password: password,
			Armor:    true,
		})
		if err != nil {
			t.Fatalf("RunDecrypt() error = %v", err)
		}

		decrypted := decBuf.String()
		if decrypted != plaintext {
			t.Errorf("RunDecrypt() = %v, want %v", decrypted, plaintext)
		}
	})

	t.Run("wrong password fails", func(t *testing.T) {
		plaintext := "secret data"
		password := "correctpassword"
		wrongPassword := "wrongpassword"

		inputFile := filepath.Join(tmpDir, "secret.txt")
		encryptedFile := filepath.Join(tmpDir, "secret.enc")

		if err := os.WriteFile(inputFile, []byte(plaintext), 0644); err != nil {
			t.Fatal(err)
		}

		// Encrypt
		var encBuf bytes.Buffer

		err := RunEncrypt(&encBuf, []string{inputFile}, CryptOptions{
			Password: password,
			Armor:    true,
		})
		if err != nil {
			t.Fatalf("RunEncrypt() error = %v", err)
		}

		if err := os.WriteFile(encryptedFile, encBuf.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}

		// Decrypt with wrong password
		var decBuf bytes.Buffer

		err = RunDecrypt(&decBuf, []string{encryptedFile}, CryptOptions{
			Password: wrongPassword,
			Armor:    true,
		})
		if err == nil {
			t.Error("RunDecrypt() expected error with wrong password")
		}
	})

	t.Run("binary mode", func(t *testing.T) {
		plaintext := "binary test data"
		password := "binarypassword"

		inputFile := filepath.Join(tmpDir, "binary.txt")
		encryptedFile := filepath.Join(tmpDir, "binary.enc")

		if err := os.WriteFile(inputFile, []byte(plaintext), 0644); err != nil {
			t.Fatal(err)
		}

		// Encrypt in binary mode
		var encBuf bytes.Buffer

		err := RunEncrypt(&encBuf, []string{inputFile}, CryptOptions{
			Password: password,
			Armor:    false,
		})
		if err != nil {
			t.Fatalf("RunEncrypt() binary error = %v", err)
		}

		if err := os.WriteFile(encryptedFile, encBuf.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}

		// Decrypt
		var decBuf bytes.Buffer

		err = RunDecrypt(&decBuf, []string{encryptedFile}, CryptOptions{
			Password: password,
			Armor:    false,
		})
		if err != nil {
			t.Fatalf("RunDecrypt() binary error = %v", err)
		}

		if decBuf.String() != plaintext {
			t.Errorf("RunDecrypt() = %v, want %v", decBuf.String(), plaintext)
		}
	})

	t.Run("no password", func(t *testing.T) {
		inputFile := filepath.Join(tmpDir, "nopass.txt")
		if err := os.WriteFile(inputFile, []byte("data"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunEncrypt(&buf, []string{inputFile}, CryptOptions{
			Password: "",
		})
		if err == nil {
			t.Error("RunEncrypt() expected error with empty password")
		}
	})
}

func TestGenerateKey(t *testing.T) {
	t.Run("default size", func(t *testing.T) {
		var buf bytes.Buffer

		err := GenerateKey(&buf, 0)
		if err != nil {
			t.Fatalf("GenerateKey() error = %v", err)
		}

		// Base64 encoded 32 bytes = 44 characters
		key := buf.String()
		if len(key) < 40 {
			t.Errorf("GenerateKey() key too short: %v", key)
		}
	})

	t.Run("custom size", func(t *testing.T) {
		var buf bytes.Buffer

		err := GenerateKey(&buf, 64)
		if err != nil {
			t.Fatalf("GenerateKey() error = %v", err)
		}

		key := buf.String()
		if len(key) < 80 {
			t.Errorf("GenerateKey() key too short for 64 bytes: %v", key)
		}
	})
}
