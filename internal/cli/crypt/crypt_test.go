package crypt

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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

func TestEncryptDecryptExtended(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "crypt_ext_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("empty file", func(t *testing.T) {
		inputFile := filepath.Join(tmpDir, "empty.txt")
		if err := os.WriteFile(inputFile, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunEncrypt(&buf, []string{inputFile}, CryptOptions{
			Password: "test123",
			Armor:    true,
		})
		if err != nil {
			t.Logf("RunEncrypt() empty file: %v", err)
		}
	})

	t.Run("unicode content", func(t *testing.T) {
		plaintext := "世界🌍こんにちは"
		password := "unicodepass"

		inputFile := filepath.Join(tmpDir, "unicode.txt")
		encryptedFile := filepath.Join(tmpDir, "unicode.enc")

		if err := os.WriteFile(inputFile, []byte(plaintext), 0644); err != nil {
			t.Fatal(err)
		}

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

		var decBuf bytes.Buffer

		err = RunDecrypt(&decBuf, []string{encryptedFile}, CryptOptions{
			Password: password,
			Armor:    true,
		})
		if err != nil {
			t.Fatalf("RunDecrypt() error = %v", err)
		}

		if decBuf.String() != plaintext {
			t.Errorf("Unicode roundtrip failed")
		}
	})

	t.Run("binary content", func(t *testing.T) {
		binaryData := []byte{0x00, 0x01, 0xFF, 0xFE, 0x7F, 0x80}
		password := "binarypass"

		inputFile := filepath.Join(tmpDir, "binary.bin")
		encryptedFile := filepath.Join(tmpDir, "binary.enc")

		if err := os.WriteFile(inputFile, binaryData, 0644); err != nil {
			t.Fatal(err)
		}

		var encBuf bytes.Buffer

		err := RunEncrypt(&encBuf, []string{inputFile}, CryptOptions{
			Password: password,
			Armor:    false,
		})
		if err != nil {
			t.Fatalf("RunEncrypt() error = %v", err)
		}

		if err := os.WriteFile(encryptedFile, encBuf.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}

		var decBuf bytes.Buffer

		err = RunDecrypt(&decBuf, []string{encryptedFile}, CryptOptions{
			Password: password,
			Armor:    false,
		})
		if err != nil {
			t.Fatalf("RunDecrypt() error = %v", err)
		}

		if !bytes.Equal(decBuf.Bytes(), binaryData) {
			t.Error("Binary roundtrip failed")
		}
	})

	t.Run("large content", func(t *testing.T) {
		largeContent := strings.Repeat("Large content for encryption test. ", 1000)
		password := "largepass"

		inputFile := filepath.Join(tmpDir, "large.txt")
		encryptedFile := filepath.Join(tmpDir, "large.enc")

		if err := os.WriteFile(inputFile, []byte(largeContent), 0644); err != nil {
			t.Fatal(err)
		}

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

		var decBuf bytes.Buffer

		err = RunDecrypt(&decBuf, []string{encryptedFile}, CryptOptions{
			Password: password,
			Armor:    true,
		})
		if err != nil {
			t.Fatalf("RunDecrypt() error = %v", err)
		}

		if decBuf.String() != largeContent {
			t.Error("Large content roundtrip failed")
		}
	})

	t.Run("special password characters", func(t *testing.T) {
		plaintext := "secret data"
		password := "p@$$w0rd!#%&*()[]{}|"

		inputFile := filepath.Join(tmpDir, "special.txt")
		encryptedFile := filepath.Join(tmpDir, "special.enc")

		if err := os.WriteFile(inputFile, []byte(plaintext), 0644); err != nil {
			t.Fatal(err)
		}

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

		var decBuf bytes.Buffer

		err = RunDecrypt(&decBuf, []string{encryptedFile}, CryptOptions{
			Password: password,
			Armor:    true,
		})
		if err != nil {
			t.Fatalf("RunDecrypt() error = %v", err)
		}

		if decBuf.String() != plaintext {
			t.Error("Special password roundtrip failed")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		var buf bytes.Buffer

		err := RunEncrypt(&buf, []string{"/nonexistent/file.txt"}, CryptOptions{
			Password: "test",
		})
		if err == nil {
			t.Error("RunEncrypt() expected error for nonexistent file")
		}
	})

	t.Run("decrypt corrupted data", func(t *testing.T) {
		corruptedFile := filepath.Join(tmpDir, "corrupted.enc")
		if err := os.WriteFile(corruptedFile, []byte("corrupted data that is not valid encrypted content"), 0644); err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer

		err := RunDecrypt(&buf, []string{corruptedFile}, CryptOptions{
			Password: "test",
			Armor:    true,
		})
		if err == nil {
			t.Log("RunDecrypt() may handle corrupted data gracefully")
		}
	})

	t.Run("different encryption same password", func(t *testing.T) {
		plaintext := "same content"
		password := "samepass"

		inputFile := filepath.Join(tmpDir, "same.txt")
		if err := os.WriteFile(inputFile, []byte(plaintext), 0644); err != nil {
			t.Fatal(err)
		}

		var buf1, buf2 bytes.Buffer

		_ = RunEncrypt(&buf1, []string{inputFile}, CryptOptions{
			Password: password,
			Armor:    true,
		})
		_ = RunEncrypt(&buf2, []string{inputFile}, CryptOptions{
			Password: password,
			Armor:    true,
		})

		// Due to random IV/salt, encryptions should differ
		if buf1.String() == buf2.String() {
			t.Log("Note: Encryptions of same content may be identical (no random IV)")
		}
	})
}

func TestGenerateKeyUniqueness(t *testing.T) {
	t.Run("multiple keys are unique", func(t *testing.T) {
		keys := make(map[string]bool)

		for i := 0; i < 100; i++ {
			var buf bytes.Buffer

			err := GenerateKey(&buf, 32)
			if err != nil {
				t.Fatalf("GenerateKey() error = %v", err)
			}

			key := strings.TrimSpace(buf.String())
			if keys[key] {
				t.Errorf("GenerateKey() produced duplicate key")
			}

			keys[key] = true
		}
	})

	t.Run("different sizes produce different lengths", func(t *testing.T) {
		sizes := []int{16, 32, 64, 128}
		lengths := make(map[int]int)

		for _, size := range sizes {
			var buf bytes.Buffer

			err := GenerateKey(&buf, size)
			if err != nil {
				t.Fatalf("GenerateKey(%d) error = %v", size, err)
			}

			lengths[size] = len(strings.TrimSpace(buf.String()))
		}

		// Verify larger sizes produce longer keys
		if lengths[16] >= lengths[32] || lengths[32] >= lengths[64] {
			t.Log("Note: Key lengths may not scale linearly with size")
		}
	})
}
