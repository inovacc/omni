package ksuid

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// KSUID is a K-Sortable Unique IDentifier
// Format: 4 bytes timestamp (seconds since epoch) + 16 bytes random payload
// Encoded as 27-character base62 string

const (
	// Epoch is the KSUID epoch (May 13, 2014)
	epoch = 1400000000
	// PayloadSize is the size of the random payload
	payloadSize = 16
	// TimestampSize is the size of the timestamp
	timestampSize = 4
	// TotalSize is the total size of a KSUID
	totalSize = timestampSize + payloadSize
	// EncodedSize is the size of the base62 encoded KSUID
	encodedSize = 27
)

// Options configures the ksuid command behavior
type Options struct {
	Count int  // -n: generate N KSUIDs
	JSON  bool // --json: output as JSON
}

// Result represents ksuid output for JSON
type Result struct {
	KSUIDs []string `json:"ksuids"`
	Count  int      `json:"count"`
}

// KSUID represents a K-Sortable Unique IDentifier
type KSUID [totalSize]byte

// RunKSUID generates KSUIDs
func RunKSUID(w io.Writer, opts Options) error {
	if opts.Count <= 0 {
		opts.Count = 1
	}

	var ksuids []string

	for i := 0; i < opts.Count; i++ {
		ksuid, err := New()
		if err != nil {
			return fmt.Errorf("ksuid: %w", err)
		}

		encoded := ksuid.String()
		if opts.JSON {
			ksuids = append(ksuids, encoded)
		} else {
			_, _ = fmt.Fprintln(w, encoded)
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(Result{KSUIDs: ksuids, Count: len(ksuids)})
	}

	return nil
}

// New generates a new KSUID
func New() (KSUID, error) {
	var ksuid KSUID

	// Get timestamp (seconds since KSUID epoch)
	timestamp := uint32(time.Now().Unix() - epoch)

	// Set timestamp bytes (big-endian)
	ksuid[0] = byte(timestamp >> 24)
	ksuid[1] = byte(timestamp >> 16)
	ksuid[2] = byte(timestamp >> 8)
	ksuid[3] = byte(timestamp)

	// Fill payload with random bytes
	_, err := rand.Read(ksuid[timestampSize:])
	if err != nil {
		return ksuid, err
	}

	return ksuid, nil
}

// String returns the base62 encoded KSUID
func (k KSUID) String() string {
	return base62Encode(k[:])
}

// Timestamp returns the time the KSUID was created
func (k KSUID) Timestamp() time.Time {
	ts := uint32(k[0])<<24 | uint32(k[1])<<16 | uint32(k[2])<<8 | uint32(k[3])
	return time.Unix(int64(ts)+epoch, 0)
}

// base62 encoding without external dependencies
const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func base62Encode(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// Convert bytes to a big integer representation
	// and then to base62
	result := make([]byte, encodedSize)
	for i := range result {
		result[i] = '0'
	}

	// Process each byte
	for _, b := range data {
		carry := int(b)
		for j := encodedSize - 1; j >= 0; j-- {
			carry += 256 * int(result[j]-'0')
			if result[j] >= 'A' && result[j] <= 'Z' {
				carry += 256 * (int(result[j]-'A') + 10 - int('0'))
			} else if result[j] >= 'a' && result[j] <= 'z' {
				carry += 256 * (int(result[j]-'a') + 36 - int('0'))
			}
			result[j] = base62Chars[carry%62]
			carry /= 62
		}
	}

	return string(result)
}

// NewString returns a new KSUID as a string
func NewString() string {
	ksuid, err := New()
	if err != nil {
		return ""
	}
	return ksuid.String()
}
