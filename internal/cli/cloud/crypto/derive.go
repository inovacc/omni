package crypto

import (
	"crypto/sha256"
)

// DeriveProfileKey derives a profile-specific encryption key from the master key.
// Uses SHA256(masterKey || provider:name) to create isolated keys per profile.
func DeriveProfileKey(masterKey []byte, provider, name string) []byte {
	h := sha256.New()
	h.Write(masterKey)
	h.Write([]byte(provider + ":" + name))

	return h.Sum(nil)
}
