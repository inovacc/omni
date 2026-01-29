package jwt

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Options configures the jwt decode command behavior
type Options struct {
	JSON    bool // --json: output as JSON
	Header  bool // --header: show only header
	Payload bool // --payload: show only payload
	Raw     bool // --raw: output raw JSON without formatting
}

// DecodedJWT represents a decoded JWT token
type DecodedJWT struct {
	Header    map[string]any `json:"header"`
	Payload   map[string]any `json:"payload"`
	Signature string         `json:"signature"`
	Valid     bool           `json:"valid"`     // structure is valid (not cryptographic)
	Expired   *bool          `json:"expired"`   // nil if no exp claim
	ExpiresAt *string        `json:"expiresAt"` // human readable expiration
}

// RunDecode decodes a JWT token
func RunDecode(w io.Writer, args []string, opts Options) error {
	input, err := getInput(args)
	if err != nil {
		return err
	}

	decoded, err := decodeJWT(input)
	if err != nil {
		return err
	}

	if opts.JSON {
		return outputJSON(w, decoded, opts)
	}

	return outputText(w, decoded, opts)
}

func decodeJWT(token string) (*DecodedJWT, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("jwt: invalid token format (expected 3 parts, got %d)", len(parts))
	}

	result := &DecodedJWT{
		Signature: parts[2],
		Valid:     true,
	}

	// Decode header
	headerBytes, err := base64URLDecode(parts[0])
	if err != nil {
		return nil, fmt.Errorf("jwt: failed to decode header: %w", err)
	}

	if err := json.Unmarshal(headerBytes, &result.Header); err != nil {
		return nil, fmt.Errorf("jwt: failed to parse header JSON: %w", err)
	}

	// Decode payload
	payloadBytes, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("jwt: failed to decode payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &result.Payload); err != nil {
		return nil, fmt.Errorf("jwt: failed to parse payload JSON: %w", err)
	}

	// Check expiration
	if exp, ok := result.Payload["exp"]; ok {
		if expFloat, ok := exp.(float64); ok {
			expTime := time.Unix(int64(expFloat), 0)
			expired := time.Now().After(expTime)
			result.Expired = &expired
			expStr := expTime.Format(time.RFC3339)
			result.ExpiresAt = &expStr
		}
	}

	return result, nil
}

func base64URLDecode(s string) ([]byte, error) {
	// Add padding if necessary
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}

	return base64.URLEncoding.DecodeString(s)
}

func outputJSON(w io.Writer, decoded *DecodedJWT, opts Options) error {
	var output any

	if opts.Header && !opts.Payload {
		output = decoded.Header
	} else if opts.Payload && !opts.Header {
		output = decoded.Payload
	} else {
		output = decoded
	}

	enc := json.NewEncoder(w)
	if !opts.Raw {
		enc.SetIndent("", "  ")
	}

	return enc.Encode(output)
}

func outputText(w io.Writer, decoded *DecodedJWT, opts Options) error {
	showHeader := !opts.Payload || opts.Header
	showPayload := !opts.Header || opts.Payload

	if showHeader {
		_, _ = fmt.Fprintln(w, "=== Header ===")

		headerJSON, err := json.MarshalIndent(decoded.Header, "", "  ")
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintln(w, string(headerJSON))
	}

	if showHeader && showPayload {
		_, _ = fmt.Fprintln(w)
	}

	if showPayload {
		_, _ = fmt.Fprintln(w, "=== Payload ===")

		payloadJSON, err := json.MarshalIndent(decoded.Payload, "", "  ")
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintln(w, string(payloadJSON))

		// Show expiration info
		if decoded.ExpiresAt != nil {
			_, _ = fmt.Fprintln(w)

			if *decoded.Expired {
				_, _ = fmt.Fprintf(w, "⚠ Token EXPIRED at %s\n", *decoded.ExpiresAt)
			} else {
				_, _ = fmt.Fprintf(w, "✓ Token expires at %s\n", *decoded.ExpiresAt)
			}
		}
	}

	return nil
}

// getInput reads input from args or stdin
func getInput(args []string) (string, error) {
	if len(args) > 0 {
		// Check if it's a file
		if _, err := os.Stat(args[0]); err == nil {
			content, err := os.ReadFile(args[0])
			if err != nil {
				return "", fmt.Errorf("jwt: %w", err)
			}

			return strings.TrimSpace(string(content)), nil
		}

		// Treat as literal string
		return strings.Join(args, " "), nil
	}

	// Read from stdin
	scanner := bufio.NewScanner(os.Stdin)

	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("jwt: %w", err)
	}

	return strings.Join(lines, ""), nil
}
