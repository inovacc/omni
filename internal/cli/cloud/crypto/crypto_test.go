package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"short", []byte("hello")},
		{"medium", []byte("The quick brown fox jumps over the lazy dog")},
		{"with_nulls", []byte("data\x00with\x00nulls")},
		{"binary", []byte{0x00, 0x01, 0x02, 0xff, 0xfe, 0xfd}},
	}

	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := EncryptWithKey(tt.data, key)
			if err != nil {
				t.Fatalf("EncryptWithKey failed: %v", err)
			}

			decrypted, err := DecryptWithKey(encrypted, key)
			if err != nil {
				t.Fatalf("DecryptWithKey failed: %v", err)
			}

			if !bytes.Equal(decrypted, tt.data) {
				t.Errorf("roundtrip failed: got %v, want %v", decrypted, tt.data)
			}
		})
	}
}

func TestEncryptWithKey_InvalidKeySize(t *testing.T) {
	tests := []struct {
		name    string
		keySize int
	}{
		{"too_short", 16},
		{"too_long", 64},
		{"empty", 0},
	}

	data := []byte("test data")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keySize)

			_, err := EncryptWithKey(data, key)
			if err == nil {
				t.Error("expected error for invalid key size")
			}
		})
	}
}

func TestDecryptWithKey_InvalidKeySize(t *testing.T) {
	key, _ := GenerateKey()
	data := []byte("test data")
	encrypted, _ := EncryptWithKey(data, key)

	badKey := make([]byte, 16)

	_, err := DecryptWithKey(encrypted, badKey)
	if err == nil {
		t.Error("expected error for invalid key size")
	}
}

func TestDecryptWithKey_TooShort(t *testing.T) {
	key, _ := GenerateKey()

	_, err := DecryptWithKey([]byte("short"), key)
	if err == nil {
		t.Error("expected error for short encrypted data")
	}
}

func TestDecryptWithKey_WrongKey(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()

	data := []byte("test data")
	encrypted, _ := EncryptWithKey(data, key1)

	_, err := DecryptWithKey(encrypted, key2)
	if err == nil {
		t.Error("expected error for wrong key")
	}
}

func TestDecryptWithKey_Tampered(t *testing.T) {
	key, _ := GenerateKey()
	data := []byte("test data")
	encrypted, _ := EncryptWithKey(data, key)

	// Tamper with the ciphertext
	encrypted[len(encrypted)-1] ^= 0xff

	_, err := DecryptWithKey(encrypted, key)
	if err == nil {
		t.Error("expected error for tampered data")
	}
}

func TestGenerateKey(t *testing.T) {
	key1, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}

	if len(key1) != KeySize {
		t.Errorf("key size: got %d, want %d", len(key1), KeySize)
	}

	// Keys should be unique
	key2, _ := GenerateKey()
	if bytes.Equal(key1, key2) {
		t.Error("generated keys should be unique")
	}
}

func TestEncryptWithKey_UniqueNonce(t *testing.T) {
	key, _ := GenerateKey()
	data := []byte("same data")

	enc1, _ := EncryptWithKey(data, key)
	enc2, _ := EncryptWithKey(data, key)

	if bytes.Equal(enc1, enc2) {
		t.Error("encryptions of same data should be different (unique nonce)")
	}
}
