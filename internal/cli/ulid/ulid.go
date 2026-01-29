package ulid

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// ULID is a Universally Unique Lexicographically Sortable Identifier
// Format: 10 chars timestamp (48-bit ms) + 16 chars randomness (80-bit)
// Total: 26 characters in Crockford's Base32

const (
	// EncodedSize is the length of a ULID string
	encodedSize = 26
	// TimestampSize is the size of the timestamp component
	timestampSize = 6
	// RandomnessSize is the size of the randomness component
	randomnessSize = 10
)

// Crockford's Base32 alphabet (excludes I, L, O, U to avoid confusion)
const crockfordAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// Options configures the ulid command behavior
type Options struct {
	Count int  // -n: generate N ULIDs
	Lower bool // -l: output in lowercase
	JSON  bool // --json: output as JSON
}

// Result represents ulid output for JSON
type Result struct {
	ULIDs []string `json:"ulids"`
	Count int      `json:"count"`
}

// ULID represents a Universally Unique Lexicographically Sortable Identifier
type ULID [16]byte

// RunULID generates ULIDs
func RunULID(w io.Writer, opts Options) error {
	if opts.Count <= 0 {
		opts.Count = 1
	}

	var ulids []string

	for i := 0; i < opts.Count; i++ {
		ulid, err := New()
		if err != nil {
			return fmt.Errorf("ulid: %w", err)
		}

		encoded := ulid.String()
		if opts.Lower {
			encoded = strings.ToLower(encoded)
		}

		if opts.JSON {
			ulids = append(ulids, encoded)
		} else {
			_, _ = fmt.Fprintln(w, encoded)
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(Result{ULIDs: ulids, Count: len(ulids)})
	}

	return nil
}

// New generates a new ULID
func New() (ULID, error) {
	return NewWithTime(time.Now())
}

// NewWithTime generates a new ULID with the given timestamp
func NewWithTime(t time.Time) (ULID, error) {
	var ulid ULID

	// Get timestamp in milliseconds
	ms := uint64(t.UnixMilli())

	// Set timestamp bytes (big-endian, 48 bits = 6 bytes)
	ulid[0] = byte(ms >> 40)
	ulid[1] = byte(ms >> 32)
	ulid[2] = byte(ms >> 24)
	ulid[3] = byte(ms >> 16)
	ulid[4] = byte(ms >> 8)
	ulid[5] = byte(ms)

	// Fill randomness with random bytes (80 bits = 10 bytes)
	_, err := rand.Read(ulid[timestampSize:])
	if err != nil {
		return ulid, err
	}

	return ulid, nil
}

// String returns the Crockford's Base32 encoded ULID
func (u ULID) String() string {
	result := make([]byte, encodedSize)

	// Encode timestamp (first 10 characters from 6 bytes = 48 bits)
	result[0] = crockfordAlphabet[(u[0]&224)>>5]
	result[1] = crockfordAlphabet[u[0]&31]
	result[2] = crockfordAlphabet[(u[1]&248)>>3]
	result[3] = crockfordAlphabet[((u[1]&7)<<2)|((u[2]&192)>>6)]
	result[4] = crockfordAlphabet[(u[2]&62)>>1]
	result[5] = crockfordAlphabet[((u[2]&1)<<4)|((u[3]&240)>>4)]
	result[6] = crockfordAlphabet[((u[3]&15)<<1)|((u[4]&128)>>7)]
	result[7] = crockfordAlphabet[(u[4]&124)>>2]
	result[8] = crockfordAlphabet[((u[4]&3)<<3)|((u[5]&224)>>5)]
	result[9] = crockfordAlphabet[u[5]&31]

	// Encode randomness (last 16 characters from 10 bytes = 80 bits)
	result[10] = crockfordAlphabet[(u[6]&248)>>3]
	result[11] = crockfordAlphabet[((u[6]&7)<<2)|((u[7]&192)>>6)]
	result[12] = crockfordAlphabet[(u[7]&62)>>1]
	result[13] = crockfordAlphabet[((u[7]&1)<<4)|((u[8]&240)>>4)]
	result[14] = crockfordAlphabet[((u[8]&15)<<1)|((u[9]&128)>>7)]
	result[15] = crockfordAlphabet[(u[9]&124)>>2]
	result[16] = crockfordAlphabet[((u[9]&3)<<3)|((u[10]&224)>>5)]
	result[17] = crockfordAlphabet[u[10]&31]
	result[18] = crockfordAlphabet[(u[11]&248)>>3]
	result[19] = crockfordAlphabet[((u[11]&7)<<2)|((u[12]&192)>>6)]
	result[20] = crockfordAlphabet[(u[12]&62)>>1]
	result[21] = crockfordAlphabet[((u[12]&1)<<4)|((u[13]&240)>>4)]
	result[22] = crockfordAlphabet[((u[13]&15)<<1)|((u[14]&128)>>7)]
	result[23] = crockfordAlphabet[(u[14]&124)>>2]
	result[24] = crockfordAlphabet[((u[14]&3)<<3)|((u[15]&224)>>5)]
	result[25] = crockfordAlphabet[u[15]&31]

	return string(result)
}

// Timestamp returns the time the ULID was created
func (u ULID) Timestamp() time.Time {
	ms := uint64(u[0])<<40 | uint64(u[1])<<32 | uint64(u[2])<<24 |
		uint64(u[3])<<16 | uint64(u[4])<<8 | uint64(u[5])
	return time.UnixMilli(int64(ms))
}

// NewString returns a new ULID as a string
func NewString() string {
	ulid, err := New()
	if err != nil {
		return ""
	}
	return ulid.String()
}
