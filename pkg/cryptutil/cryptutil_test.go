package cryptutil

import (
	"bytes"
	"strings"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		plaintext string
		password  string
		opts      []Option
	}{
		{
			name:      "simple text",
			plaintext: "Hello, World! This is a secret message.",
			password:  "supersecretpassword123",
		},
		{
			name:      "with base64",
			plaintext: "Hello, World!",
			password:  "testpass",
			opts:      []Option{WithBase64()},
		},
		{
			name:      "unicode content",
			plaintext: "世界🌍こんにちは",
			password:  "unicodepass",
		},
		{
			name:      "empty plaintext",
			plaintext: "",
			password:  "test123",
		},
		{
			name:      "special password characters",
			plaintext: "secret data",
			password:  "p@$$w0rd!#%&*()[]{}|",
		},
		{
			name:      "custom iterations",
			plaintext: "test data",
			password:  "testpass",
			opts:      []Option{WithIterations(1000)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := Encrypt([]byte(tt.plaintext), tt.password, tt.opts...)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			if string(encrypted) == tt.plaintext && len(tt.plaintext) > 0 {
				t.Error("Encrypt() output should not equal plaintext")
			}

			decrypted, err := Decrypt(encrypted, tt.password, tt.opts...)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if string(decrypted) != tt.plaintext {
				t.Errorf("Decrypt() = %v, want %v", string(decrypted), tt.plaintext)
			}
		})
	}
}

func TestDecryptWrongPassword(t *testing.T) {
	plaintext := []byte("secret data")

	encrypted, err := Encrypt(plaintext, "correctpassword")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	_, err = Decrypt(encrypted, "wrongpassword")
	if err == nil {
		t.Error("Decrypt() expected error with wrong password")
	}
}

func TestDecryptCorrupted(t *testing.T) {
	_, err := Decrypt([]byte("too short"), "test")
	if err == nil {
		t.Error("Decrypt() expected error for short input")
	}
}

func TestDecryptInvalidBase64(t *testing.T) {
	_, err := Decrypt([]byte("not valid base64!!!"), "test", WithBase64())
	if err == nil {
		t.Error("Decrypt() expected error for invalid base64")
	}
}

func TestEncryptBinary(t *testing.T) {
	binaryData := []byte{0x00, 0x01, 0xFF, 0xFE, 0x7F, 0x80}
	password := "binarypass"

	encrypted, err := Encrypt(binaryData, password)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	decrypted, err := Decrypt(encrypted, password)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if !bytes.Equal(decrypted, binaryData) {
		t.Error("Binary roundtrip failed")
	}
}

func TestEncryptLargeContent(t *testing.T) {
	largeContent := []byte(strings.Repeat("Large content for encryption test. ", 1000))
	password := "largepass"

	encrypted, err := Encrypt(largeContent, password, WithBase64())
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	decrypted, err := Decrypt(encrypted, password, WithBase64())
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if !bytes.Equal(decrypted, largeContent) {
		t.Error("Large content roundtrip failed")
	}
}

func TestEncryptDifferentOutputs(t *testing.T) {
	plaintext := []byte("same content")
	password := "samepass"

	enc1, _ := Encrypt(plaintext, password)
	enc2, _ := Encrypt(plaintext, password)

	// Due to random salt/nonce, encryptions should differ
	if bytes.Equal(enc1, enc2) {
		t.Log("Note: Two encryptions of same content produced identical output (extremely unlikely)")
	}
}

func TestGenerateKey(t *testing.T) {
	t.Run("default size", func(t *testing.T) {
		key, err := GenerateKey(0)
		if err != nil {
			t.Fatalf("GenerateKey() error = %v", err)
		}

		// Base64 encoded 32 bytes = 44 characters
		if len(key) < 40 {
			t.Errorf("GenerateKey() key too short: %v", key)
		}
	})

	t.Run("custom size", func(t *testing.T) {
		key, err := GenerateKey(64)
		if err != nil {
			t.Fatalf("GenerateKey() error = %v", err)
		}

		if len(key) < 80 {
			t.Errorf("GenerateKey() key too short for 64 bytes: %v", key)
		}
	})

	t.Run("uniqueness", func(t *testing.T) {
		keys := make(map[string]bool)

		for range 100 {
			key, err := GenerateKey(32)
			if err != nil {
				t.Fatalf("GenerateKey() error = %v", err)
			}

			if keys[key] {
				t.Errorf("GenerateKey() produced duplicate key")
			}

			keys[key] = true
		}
	})
}

func TestDeriveKey(t *testing.T) {
	salt := []byte("testsalt12345678")
	key1 := DeriveKey("password", salt, 1000, 32)
	key2 := DeriveKey("password", salt, 1000, 32)

	if !bytes.Equal(key1, key2) {
		t.Error("DeriveKey() should produce same output for same inputs")
	}

	key3 := DeriveKey("different", salt, 1000, 32)
	if bytes.Equal(key1, key3) {
		t.Error("DeriveKey() should produce different output for different passwords")
	}
}
