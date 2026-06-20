package encoding

import (
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcutil/base58"
)

// Base64Encode encodes data to base64
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64Decode decodes base64 data
func Base64Decode(s string) ([]byte, error) {
	// Remove whitespace
	clean := stripWhitespace(s)

	decoded, err := base64.StdEncoding.DecodeString(clean)
	if err != nil {
		return nil, fmt.Errorf("base64: invalid input: %w", err)
	}

	return decoded, nil
}

// Base32Encode encodes data to base32
func Base32Encode(data []byte) string {
	return base32.StdEncoding.EncodeToString(data)
}

// Base32Decode decodes base32 data
func Base32Decode(s string) ([]byte, error) {
	clean := stripWhitespace(s)

	decoded, err := base32.StdEncoding.DecodeString(clean)
	if err != nil {
		return nil, fmt.Errorf("base32: invalid input: %w", err)
	}

	return decoded, nil
}

// Base58Encode encodes data to base58 (Bitcoin alphabet)
func Base58Encode(data []byte) string {
	return base58.Encode(data)
}

// Base58Decode decodes base58 data (Bitcoin alphabet).
//
// Deprecated: Use Base58DecodeStrict, which reports invalid input instead of
// silently returning an empty slice. Will be removed after 2026-07-19.
func Base58Decode(s string) []byte {
	return base58.Decode(s)
}

// base58Alphabet is the Bitcoin base58 alphabet (excludes 0, O, I, l).
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// Base58DecodeStrict decodes base58 (Bitcoin alphabet) and reports invalid
// input, unlike the deprecated Base58Decode which silently returns an empty
// slice. An empty (or all-whitespace) input decodes to an empty slice with no
// error.
func Base58DecodeStrict(s string) ([]byte, error) {
	clean := stripWhitespace(s)
	for _, r := range clean {
		if !strings.ContainsRune(base58Alphabet, r) {
			return nil, fmt.Errorf("base58: invalid input: illegal character %q", r)
		}
	}
	return base58.Decode(clean), nil
}

// WrapString wraps a string at the specified width, inserting newlines.
func WrapString(s string, width int) string {
	if width <= 0 || len(s) <= width {
		return s
	}

	var result strings.Builder

	for i := 0; i < len(s); i += width {
		end := min(i+width, len(s))

		result.WriteString(s[i:end])

		if end < len(s) {
			result.WriteString("\n")
		}
	}

	return result.String()
}

func stripWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
			return -1
		}

		return r
	}, s)
}
