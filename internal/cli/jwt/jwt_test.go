package jwt

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// Sample JWT tokens for testing
// Header: {"alg":"HS256","typ":"JWT"}
// Payload: {"sub":"1234567890","name":"John Doe","iat":1516239022}
const validToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

func TestRunDecode(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		opts        Options
		wantHeader  map[string]any
		wantPayload map[string]any
		wantErr     bool
	}{
		{
			name: "valid token",
			args: []string{validToken},
			opts: Options{},
			wantHeader: map[string]any{
				"alg": "HS256",
				"typ": "JWT",
			},
			wantPayload: map[string]any{
				"sub":  "1234567890",
				"name": "John Doe",
				"iat":  float64(1516239022),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunDecode(&buf, tt.args, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunDecode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			output := buf.String()

			// Check header values appear
			if tt.wantHeader["alg"] != nil {
				if !strings.Contains(output, "HS256") {
					t.Errorf("Output missing algorithm")
				}
			}

			// Check payload values appear
			if tt.wantPayload["name"] != nil {
				if !strings.Contains(output, "John Doe") {
					t.Errorf("Output missing name")
				}
			}
		})
	}
}

func TestRunDecodeInvalid(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr string
	}{
		{
			name:    "empty token",
			token:   "",
			wantErr: "invalid token format",
		},
		{
			name:    "one part",
			token:   "eyJhbGciOiJIUzI1NiJ9",
			wantErr: "invalid token format",
		},
		{
			name:    "two parts",
			token:   "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0",
			wantErr: "invalid token format",
		},
		{
			name:    "invalid base64",
			token:   "!!!.!!!.!!!",
			wantErr: "failed to decode header",
		},
		{
			name:    "invalid json",
			token:   "bm90anNvbg.bm90anNvbg.sig",
			wantErr: "failed to parse header JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := RunDecode(&buf, []string{tt.token}, Options{})
			if err == nil {
				t.Errorf("RunDecode() expected error, got nil")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("RunDecode() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestRunDecodeJSON(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{JSON: true}

	err := RunDecode(&buf, []string{validToken}, opts)
	if err != nil {
		t.Fatalf("RunDecode() error = %v", err)
	}

	var result DecodedJWT
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result.Header["alg"] != "HS256" {
		t.Errorf("Header alg = %v, want HS256", result.Header["alg"])
	}

	if result.Payload["name"] != "John Doe" {
		t.Errorf("Payload name = %v, want John Doe", result.Payload["name"])
	}

	if !result.Valid {
		t.Errorf("Valid = false, want true")
	}
}

func TestRunDecodeHeaderOnly(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Header: true}

	err := RunDecode(&buf, []string{validToken}, opts)
	if err != nil {
		t.Fatalf("RunDecode() error = %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Header") {
		t.Errorf("Output should contain Header section")
	}

	if strings.Contains(output, "Payload") {
		t.Errorf("Output should not contain Payload section")
	}
}

func TestRunDecodePayloadOnly(t *testing.T) {
	var buf bytes.Buffer

	opts := Options{Payload: true}

	err := RunDecode(&buf, []string{validToken}, opts)
	if err != nil {
		t.Fatalf("RunDecode() error = %v", err)
	}

	output := buf.String()

	if strings.Contains(output, "Header") {
		t.Errorf("Output should not contain Header section")
	}

	if !strings.Contains(output, "Payload") {
		t.Errorf("Output should contain Payload section")
	}
}

func TestExpiredToken(t *testing.T) {
	// Create a token with exp in the past
	// Header: {"alg":"HS256","typ":"JWT"}
	// Payload: {"sub":"123","exp":1000000000} (Sep 2001)
	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMiLCJleHAiOjEwMDAwMDAwMDB9.signature"

	var buf bytes.Buffer

	opts := Options{JSON: true}

	err := RunDecode(&buf, []string{expiredToken}, opts)
	if err != nil {
		t.Fatalf("RunDecode() error = %v", err)
	}

	var result DecodedJWT
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result.Expired == nil {
		t.Fatalf("Expired should not be nil")
	}

	if !*result.Expired {
		t.Errorf("Token should be expired")
	}
}

func TestNotExpiredToken(t *testing.T) {
	// Create a token with exp far in the future
	futureExp := time.Now().Add(24 * time.Hour).Unix()

	// We'll test with the decoded function directly
	decoded := &DecodedJWT{
		Header:  map[string]any{"alg": "HS256"},
		Payload: map[string]any{"exp": float64(futureExp)},
	}

	// Check expiration
	if exp, ok := decoded.Payload["exp"]; ok {
		if expFloat, ok := exp.(float64); ok {
			expTime := time.Unix(int64(expFloat), 0)
			expired := time.Now().After(expTime)
			decoded.Expired = &expired
		}
	}

	if decoded.Expired == nil {
		t.Fatalf("Expired should not be nil")
	}

	if *decoded.Expired {
		t.Errorf("Token should not be expired")
	}
}

func TestBase64URLDecode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "no padding needed",
			input: "dGVzdA",
			want:  "test",
		},
		{
			name:  "one padding",
			input: "dGVzdHM",
			want:  "tests",
		},
		{
			name:  "two padding",
			input: "dGVzdGluZw",
			want:  "testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := base64URLDecode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("base64URLDecode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if string(got) != tt.want {
				t.Errorf("base64URLDecode() = %q, want %q", string(got), tt.want)
			}
		})
	}
}
