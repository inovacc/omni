package cli

import (
	"crypto/rand"
	"fmt"
	"io"
	"strings"
)

// UUIDOptions configures the uuid command behavior
type UUIDOptions struct {
	Count    int  // -n: generate N UUIDs
	Upper    bool // -u: output in uppercase
	NoDashes bool // -x: output without dashes
	Version  int  // -v: UUID version (4 = random, default)
}

// RunUUID generates random UUIDs
func RunUUID(w io.Writer, opts UUIDOptions) error {
	if opts.Count <= 0 {
		opts.Count = 1
	}

	for i := 0; i < opts.Count; i++ {
		uuid, err := generateUUIDv4()
		if err != nil {
			return fmt.Errorf("uuid: %w", err)
		}

		if opts.NoDashes {
			uuid = strings.ReplaceAll(uuid, "-", "")
		}

		if opts.Upper {
			uuid = strings.ToUpper(uuid)
		}

		_, _ = fmt.Fprintln(w, uuid)
	}

	return nil
}

// generateUUIDv4 generates a random UUID version 4
func generateUUIDv4() (string, error) {
	uuid := make([]byte, 16)

	_, err := rand.Read(uuid)
	if err != nil {
		return "", err
	}

	// Set version (4) and variant (RFC 4122)
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant RFC 4122

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4],
		uuid[4:6],
		uuid[6:8],
		uuid[8:10],
		uuid[10:16]), nil
}

// NewUUID returns a new random UUID string
func NewUUID() string {
	uuid, err := generateUUIDv4()
	if err != nil {
		// Fallback to empty string on error (shouldn't happen with crypto/rand)
		return ""
	}
	return uuid
}

// MustNewUUID returns a new random UUID string, panics on error
func MustNewUUID() string {
	uuid, err := generateUUIDv4()
	if err != nil {
		panic(fmt.Sprintf("uuid: %v", err))
	}
	return uuid
}

// IsValidUUID checks if a string is a valid UUID format
func IsValidUUID(s string) bool {
	// Remove dashes for validation
	s = strings.ReplaceAll(s, "-", "")

	if len(s) != 32 {
		return false
	}

	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}

	return true
}
