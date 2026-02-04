package crypto

import (
	"bytes"
	"testing"
)

func TestDeriveProfileKey(t *testing.T) {
	masterKey := []byte("12345678901234567890123456789012")

	tests := []struct {
		name     string
		provider string
		profile  string
	}{
		{"aws_prod", "aws", "prod"},
		{"aws_dev", "aws", "dev"},
		{"azure_prod", "azure", "prod"},
		{"gcp_prod", "gcp", "prod"},
	}

	keys := make(map[string][]byte)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := DeriveProfileKey(masterKey, tt.provider, tt.profile)

			if len(key) != KeySize {
				t.Errorf("key size: got %d, want %d", len(key), KeySize)
			}

			// Same inputs should produce same output
			key2 := DeriveProfileKey(masterKey, tt.provider, tt.profile)
			if !bytes.Equal(key, key2) {
				t.Error("same inputs should produce same key")
			}

			keys[tt.name] = key
		})
	}

	// Different inputs should produce different keys
	for name1, key1 := range keys {
		for name2, key2 := range keys {
			if name1 != name2 && bytes.Equal(key1, key2) {
				t.Errorf("different inputs (%s, %s) should produce different keys", name1, name2)
			}
		}
	}
}

func TestDeriveProfileKey_DifferentMasterKeys(t *testing.T) {
	masterKey1 := []byte("12345678901234567890123456789012")
	masterKey2 := []byte("abcdefghijklmnopqrstuvwxyz123456")

	key1 := DeriveProfileKey(masterKey1, "aws", "prod")
	key2 := DeriveProfileKey(masterKey2, "aws", "prod")

	if bytes.Equal(key1, key2) {
		t.Error("different master keys should produce different derived keys")
	}
}
