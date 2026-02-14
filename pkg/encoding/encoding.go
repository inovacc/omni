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

// Base58Decode decodes base58 data (Bitcoin alphabet)
func Base58Decode(s string) []byte {
	return base58.Decode(s)
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
