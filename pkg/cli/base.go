package cli

import (
	"bufio"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"os"
	"strings"
)

// BaseOptions configures the base encode/decode command behavior
type BaseOptions struct {
	Decode        bool // -d: decode data
	Wrap          int  // -w: wrap encoded lines after N characters (0 = no wrap)
	IgnoreGarbage bool // -i: ignore non-alphabet characters when decoding
}

// Base58 alphabet (Bitcoin style)
const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// RunBase64 encodes or decodes base64 data
func RunBase64(w io.Writer, args []string, opts BaseOptions) error {
	if opts.Wrap == 0 {
		opts.Wrap = 76 // default wrap for base64
	}

	var input io.Reader
	if len(args) == 0 || args[0] == "-" {
		input = os.Stdin
	} else {
		f, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("base64: %w", err)
		}

		defer func() {
			_ = f.Close()
		}()

		input = f
	}

	data, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("base64: %w", err)
	}

	if opts.Decode {
		// Remove whitespace for decoding
		cleanData := strings.Map(func(r rune) rune {
			if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
				return -1
			}

			return r
		}, string(data))

		decoded, err := base64.StdEncoding.DecodeString(cleanData)
		if err != nil {
			return fmt.Errorf("base64: invalid input: %w", err)
		}

		_, _ = w.Write(decoded)
	} else {
		encoded := base64.StdEncoding.EncodeToString(data)
		if opts.Wrap > 0 {
			encoded = wrapString(encoded, opts.Wrap)
		}

		_, _ = fmt.Fprintln(w, encoded)
	}

	return nil
}

// RunBase32 encodes or decodes base32 data
func RunBase32(w io.Writer, args []string, opts BaseOptions) error {
	if opts.Wrap == 0 {
		opts.Wrap = 76
	}

	var input io.Reader
	if len(args) == 0 || args[0] == "-" {
		input = os.Stdin
	} else {
		f, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("base32: %w", err)
		}

		defer func() {
			_ = f.Close()
		}()

		input = f
	}

	data, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("base32: %w", err)
	}

	if opts.Decode {
		cleanData := strings.Map(func(r rune) rune {
			if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
				return -1
			}

			return r
		}, string(data))

		decoded, err := base32.StdEncoding.DecodeString(cleanData)
		if err != nil {
			return fmt.Errorf("base32: invalid input: %w", err)
		}

		_, _ = w.Write(decoded)
	} else {
		encoded := base32.StdEncoding.EncodeToString(data)
		if opts.Wrap > 0 {
			encoded = wrapString(encoded, opts.Wrap)
		}

		_, _ = fmt.Fprintln(w, encoded)
	}

	return nil
}

// RunBase58 encodes or decodes base58 data (Bitcoin alphabet)
func RunBase58(w io.Writer, args []string, opts BaseOptions) error {
	var input io.Reader
	if len(args) == 0 || args[0] == "-" {
		input = os.Stdin
	} else {
		f, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("base58: %w", err)
		}

		defer func() {
			_ = f.Close()
		}()

		input = f
	}

	// Read line by line for base58 (typically used for short strings)
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()
		if opts.Decode {
			decoded, err := base58Decode(line)
			if err != nil {
				return fmt.Errorf("base58: invalid input: %w", err)
			}

			_, _ = w.Write(decoded)
			_, _ = fmt.Fprintln(w)
		} else {
			encoded := base58Encode([]byte(line))
			_, _ = fmt.Fprintln(w, encoded)
		}
	}

	return scanner.Err()
}

func base58Encode(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// Count leading zeros
	var leadingZeros int

	for _, b := range data {
		if b == 0 {
			leadingZeros++
		} else {
			break
		}
	}

	// Convert to big integer
	num := new(big.Int).SetBytes(data)
	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := new(big.Int)

	var encoded []byte

	for num.Cmp(zero) > 0 {
		num.DivMod(num, base, mod)
		encoded = append([]byte{base58Alphabet[mod.Int64()]}, encoded...)
	}

	// Add leading '1's for each leading zero byte
	for i := 0; i < leadingZeros; i++ {
		encoded = append([]byte{'1'}, encoded...)
	}

	return string(encoded)
}

func base58Decode(s string) ([]byte, error) {
	if len(s) == 0 {
		return []byte{}, nil
	}

	// Count leading '1's
	var leadingOnes int

	for _, c := range s {
		if c == '1' {
			leadingOnes++
		} else {
			break
		}
	}

	// Convert from base58
	num := big.NewInt(0)
	base := big.NewInt(58)

	for _, c := range s {
		idx := strings.IndexRune(base58Alphabet, c)
		if idx == -1 {
			return nil, fmt.Errorf("invalid base58 character: %c", c)
		}

		num.Mul(num, base)
		num.Add(num, big.NewInt(int64(idx)))
	}

	decoded := num.Bytes()

	// Add leading zero bytes
	result := make([]byte, leadingOnes+len(decoded))
	copy(result[leadingOnes:], decoded)

	return result, nil
}

func wrapString(s string, width int) string {
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
