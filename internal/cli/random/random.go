package random

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"strings"

	"github.com/inovacc/omni/internal/cli/output"
)

// RandomOptions configures the random command behavior
type RandomOptions struct {
	Count        int           // -n: number of values to generate
	Length       int           // -l: length of random strings
	Min          int64         // --min: minimum value for numbers
	Max          int64         // --max: maximum value for numbers
	Type         string        // -t: type (int, float, string, hex, alpha, alnum, bytes)
	Charset      string        // -c: custom character set
	Sep          string        // -s: separator between values
	OutputFormat output.Format // output format (text, json, table)
}

// RandomResult represents random output for JSON
type RandomResult struct {
	Type   string   `json:"type"`
	Values []string `json:"values"`
	Count  int      `json:"count"`
}

const (
	charsetAlpha    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsetDigits   = "0123456789"
	charsetAlnum    = charsetAlpha + charsetDigits
	charsetHex      = "0123456789abcdef"
	charsetSpecial  = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	charsetPassword = charsetAlnum + charsetSpecial
)

// RunRandom generates random values
func RunRandom(w io.Writer, opts RandomOptions) error {
	if opts.Count <= 0 {
		opts.Count = 1
	}

	if opts.Length <= 0 {
		opts.Length = 16
	}

	if opts.Sep == "" {
		opts.Sep = "\n"
	}

	if opts.Type == "" {
		opts.Type = "string"
	}

	var results []string

	for i := 0; i < opts.Count; i++ {
		var (
			result string
			err    error
		)

		switch strings.ToLower(opts.Type) {
		case "int", "integer", "number":
			result, err = randomInt(opts.Min, opts.Max)
		case "float", "decimal":
			result, err = randomFloat()
		case "string", "str":
			result, err = randomString(opts.Length, charsetAlnum)
		case "alpha", "letters":
			result, err = randomString(opts.Length, charsetAlpha)
		case "alnum", "alphanumeric":
			result, err = randomString(opts.Length, charsetAlnum)
		case "hex":
			result, err = randomString(opts.Length, charsetHex)
		case "password", "pass":
			result, err = randomString(opts.Length, charsetPassword)
		case "bytes", "binary":
			result, err = randomBytes(opts.Length)
		case "custom":
			if opts.Charset == "" {
				return fmt.Errorf("random: custom type requires -c charset")
			}

			result, err = randomString(opts.Length, opts.Charset)
		default:
			return fmt.Errorf("random: unknown type: %s", opts.Type)
		}

		if err != nil {
			return fmt.Errorf("random: %w", err)
		}

		results = append(results, result)
	}

	f := output.New(w, opts.OutputFormat)

	if f.IsJSON() {
		return f.Print(RandomResult{Type: opts.Type, Values: results, Count: len(results)})
	}

	_, _ = fmt.Fprintln(w, strings.Join(results, opts.Sep))

	return nil
}

func randomInt(minVal, maxVal int64) (string, error) {
	if maxVal <= minVal {
		maxVal = minVal + 100
	}

	rangeSize := maxVal - minVal

	n, err := rand.Int(rand.Reader, big.NewInt(rangeSize))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", n.Int64()+minVal), nil
}

func randomFloat() (string, error) {
	// Generate random float between 0 and 1
	n, err := rand.Int(rand.Reader, big.NewInt(1<<53))
	if err != nil {
		return "", err
	}

	f := float64(n.Int64()) / float64(1<<53)

	return fmt.Sprintf("%f", f), nil
}

func randomString(length int, charset string) (string, error) {
	if length <= 0 {
		length = 16
	}

	if charset == "" {
		charset = charsetAlnum
	}

	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}

		result[i] = charset[idx.Int64()]
	}

	return string(result), nil
}

func randomBytes(length int) (string, error) {
	if length <= 0 {
		length = 16
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Return as hex string
	return fmt.Sprintf("%x", bytes), nil
}

// RandomString generates a random alphanumeric string of given length
func RandomString(length int) string {
	s, _ := randomString(length, charsetAlnum)
	return s
}

// RandomHex generates a random hex string of given length
func RandomHex(length int) string {
	s, _ := randomString(length, charsetHex)
	return s
}

// RandomPassword generates a random password of given length
func RandomPassword(length int) string {
	s, _ := randomString(length, charsetPassword)
	return s
}

// RandomInt generates a random integer in [minVal, maxVal)
func RandomInt(minVal, maxVal int64) int64 {
	if maxVal <= minVal {
		return minVal
	}

	n, err := rand.Int(rand.Reader, big.NewInt(maxVal-minVal))
	if err != nil {
		return minVal
	}

	return n.Int64() + minVal
}

// RandomChoice returns a random element from the slice
func RandomChoice[T any](items []T) T {
	var zero T
	if len(items) == 0 {
		return zero
	}

	idx := RandomInt(0, int64(len(items)))

	return items[idx]
}

// Shuffle randomly shuffles a slice in place
func Shuffle[T any](items []T) {
	n := len(items)
	for i := n - 1; i > 0; i-- {
		j := int(RandomInt(0, int64(i+1)))
		items[i], items[j] = items[j], items[i]
	}
}
