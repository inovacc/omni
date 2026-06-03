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

	// masterPassphraseEnv is an optional user-controlled secret mixed into the
	// master-key KDF. When set during key creation, the machine ID alone is no
	// longer sufficient to recover the master key. It is opt-in: if unset, the
	// derivation falls back to the legacy machine-ID-only behavior so existing
	// master.key files remain decryptable.
	masterPassphraseEnv = "OMNI_MASTER_PASSPHRASE"
)

// MasterKeyData holds the encrypted master key and its salt.
type MasterKeyData struct {
	Salt      []byte `json:"salt"`
	Encrypted []byte `json:"encrypted"`
	// WithPassphrase records whether a user passphrase (OMNI_MASTER_PASSPHRASE)
	// was mixed into the KDF when this key was created. It is omitted for legacy
	// machine-ID-only files so their on-disk format is unchanged, and absence is
	// interpreted as false (machine-ID-only derivation).
	WithPassphrase bool `json:"with_passphrase,omitempty"`
}

// deriveMasterEncKey derives the key-encryption key from the machine ID and,
// when withPassphrase is true, the user-supplied OMNI_MASTER_PASSPHRASE secret.
// A NUL byte separates the two inputs so distinct (machineID, passphrase) pairs
// cannot collide. When withPassphrase is true but the env var is empty, an error
// is returned so a key created with a passphrase cannot be silently downgraded.
func deriveMasterEncKey(machineID string, salt []byte, withPassphrase bool) ([]byte, error) {
	secret := machineID
	if withPassphrase {
		passphrase := os.Getenv(masterPassphraseEnv)
		if passphrase == "" {
			return nil, fmt.Errorf("master key requires %s but it is not set", masterPassphraseEnv)
		}
		secret = machineID + "\x00" + passphrase
	}

	return pbkdf2.Key([]byte(secret), salt, pbkdf2Iter, KeySize, sha256.New), nil
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

	// Mix in the optional user passphrase when present so machine ID alone is
	// not sufficient to recover the master key.
	withPassphrase := os.Getenv(masterPassphraseEnv) != ""

	// Derive encryption key from machine ID (and passphrase, if set)
	derivedKey, err := deriveMasterEncKey(machineID, salt, withPassphrase)
	if err != nil {
		return nil, fmt.Errorf("deriving master encryption key: %w", err)
	}

	// Encrypt master key
	encrypted, err := EncryptWithKey(masterKey, derivedKey)
	if err != nil {
		return nil, fmt.Errorf("encrypting master key: %w", err)
	}

	// Save to file
	data := MasterKeyData{
		Salt:           salt,
		Encrypted:      encrypted,
		WithPassphrase: withPassphrase,
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

	// Derive decryption key from machine ID (and passphrase, if the key was
	// created with one). Legacy files omit the marker, so this stays
	// machine-ID-only and decrypts exactly as before.
	derivedKey, err := deriveMasterEncKey(machineID, data.Salt, data.WithPassphrase)
	if err != nil {
		return nil, fmt.Errorf("deriving master encryption key: %w", err)
	}

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
