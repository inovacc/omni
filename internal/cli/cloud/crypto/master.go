package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/pbkdf2"
)

const (
	masterKeyFile  = "master.key"
	saltSize       = 16
	pbkdf2Iter     = 100000
	masterKeyPerms = 0600
)

// MasterKeyData holds the encrypted master key and its salt.
type MasterKeyData struct {
	Salt      []byte `json:"salt"`
	Encrypted []byte `json:"encrypted"`
}

// GetOrCreateMasterKey retrieves or creates the master encryption key.
// The master key is encrypted using a key derived from the machine ID.
func GetOrCreateMasterKey(baseDir string) ([]byte, error) {
	keyPath := filepath.Join(baseDir, masterKeyFile)

	// Try to load existing key
	if _, err := os.Stat(keyPath); err == nil {
		return LoadMasterKey(baseDir)
	}

	// Generate new master key
	masterKey, err := GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("generating master key: %w", err)
	}

	// Get machine ID for encryption
	machineID, err := GetMachineID()
	if err != nil {
		return nil, fmt.Errorf("getting machine ID: %w", err)
	}

	// Generate salt for PBKDF2
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generating salt: %w", err)
	}

	// Derive encryption key from machine ID
	derivedKey := pbkdf2.Key([]byte(machineID), salt, pbkdf2Iter, KeySize, sha256.New)

	// Encrypt master key
	encrypted, err := EncryptWithKey(masterKey, derivedKey)
	if err != nil {
		return nil, fmt.Errorf("encrypting master key: %w", err)
	}

	// Save to file
	data := MasterKeyData{
		Salt:      salt,
		Encrypted: encrypted,
	}

	if err := saveMasterKeyData(keyPath, &data); err != nil {
		return nil, err
	}

	return masterKey, nil
}

// LoadMasterKey loads and decrypts the master key from disk.
func LoadMasterKey(baseDir string) ([]byte, error) {
	keyPath := filepath.Join(baseDir, masterKeyFile)

	data, err := loadMasterKeyData(keyPath)
	if err != nil {
		return nil, err
	}

	// Get machine ID for decryption
	machineID, err := GetMachineID()
	if err != nil {
		return nil, fmt.Errorf("getting machine ID: %w", err)
	}

	// Derive decryption key from machine ID
	derivedKey := pbkdf2.Key([]byte(machineID), data.Salt, pbkdf2Iter, KeySize, sha256.New)

	// Decrypt master key
	masterKey, err := DecryptWithKey(data.Encrypted, derivedKey)
	if err != nil {
		return nil, fmt.Errorf("decrypting master key (machine changed?): %w", err)
	}

	return masterKey, nil
}

func saveMasterKeyData(path string, data *MasterKeyData) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling master key data: %w", err)
	}

	if err := os.WriteFile(path, jsonData, masterKeyPerms); err != nil {
		return fmt.Errorf("writing master key file: %w", err)
	}

	return nil
}

func loadMasterKeyData(path string) (*MasterKeyData, error) {
	jsonData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading master key file: %w", err)
	}

	var data MasterKeyData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("parsing master key file: %w", err)
	}

	return &data, nil
}
