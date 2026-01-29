package nanoid

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
)

// Default alphabet for NanoID (URL-safe)
const defaultAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz-"

// Default length for NanoID
const defaultLength = 21

// Options configures the nanoid command behavior
type Options struct {
	Count    int    // -n: generate N NanoIDs
	Length   int    // -l: length of NanoID (default 21)
	Alphabet string // -a: custom alphabet
	JSON     bool   // --json: output as JSON
}

// Result represents nanoid output for JSON
type Result struct {
	NanoIDs []string `json:"nanoids"`
	Count   int      `json:"count"`
}

// RunNanoID generates NanoIDs
func RunNanoID(w io.Writer, opts Options) error {
	if opts.Count <= 0 {
		opts.Count = 1
	}

	if opts.Length <= 0 {
		opts.Length = defaultLength
	}

	alphabet := opts.Alphabet
	if alphabet == "" {
		alphabet = defaultAlphabet
	}

	var nanoids []string

	for i := 0; i < opts.Count; i++ {
		nanoid, err := Generate(alphabet, opts.Length)
		if err != nil {
			return fmt.Errorf("nanoid: %w", err)
		}

		if opts.JSON {
			nanoids = append(nanoids, nanoid)
		} else {
			_, _ = fmt.Fprintln(w, nanoid)
		}
	}

	if opts.JSON {
		return json.NewEncoder(w).Encode(Result{NanoIDs: nanoids, Count: len(nanoids)})
	}

	return nil
}

// Generate creates a NanoID with custom alphabet and length
func Generate(alphabet string, length int) (string, error) {
	if len(alphabet) == 0 {
		alphabet = defaultAlphabet
	}

	if length <= 0 {
		length = defaultLength
	}

	alphabetLen := big.NewInt(int64(len(alphabet)))
	result := make([]byte, length)

	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, alphabetLen)
		if err != nil {
			return "", err
		}
		result[i] = alphabet[idx.Int64()]
	}

	return string(result), nil
}

// New generates a new NanoID with default settings
func New() (string, error) {
	return Generate(defaultAlphabet, defaultLength)
}

// NewString returns a new NanoID as a string, empty on error
func NewString() string {
	nanoid, err := New()
	if err != nil {
		return ""
	}
	return nanoid
}

// MustNew generates a new NanoID and panics on error
func MustNew() string {
	nanoid, err := New()
	if err != nil {
		panic(fmt.Sprintf("nanoid: %v", err))
	}
	return nanoid
}
