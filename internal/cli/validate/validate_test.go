package validate

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

func TestRunEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"valid simple", "user@example.com", nil},
		{"valid with name", "Barry Gibbs <bg@example.com>", nil},
		{"invalid no at", "not-an-email", cmderr.ErrConflict},
		{"invalid empty", "", cmderr.ErrConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := RunEmail(&buf, tt.input)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("RunEmail(%q) error = %v, want errors.Is(..., %v)", tt.input, err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("RunEmail(%q) unexpected error = %v", tt.input, err)
			}
			if !strings.Contains(buf.String(), "OK") {
				t.Errorf("RunEmail(%q) output = %q, want it to contain OK", tt.input, buf.String())
			}
		})
	}
}

func TestRunIP(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"valid ipv4", "192.168.0.1", nil},
		{"valid ipv6", "::1", nil},
		{"invalid octet", "999.1.1.1", cmderr.ErrConflict},
		{"invalid junk", "not-an-ip", cmderr.ErrConflict},
		{"invalid empty", "", cmderr.ErrConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := RunIP(&buf, tt.input)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("RunIP(%q) error = %v, want errors.Is(..., %v)", tt.input, err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("RunIP(%q) unexpected error = %v", tt.input, err)
			}
			if !strings.Contains(buf.String(), "OK") {
				t.Errorf("RunIP(%q) output = %q, want it to contain OK", tt.input, buf.String())
			}
		})
	}
}
