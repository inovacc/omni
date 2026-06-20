package hmac

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

func TestRunHMAC(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		stdin   string
		opts    HMACOptions
		want    string
		wantErr error
	}{
		{
			// Well-known Wikipedia HMAC-SHA256 vector: key "key", pangram message.
			name: "sha256 arg known vector",
			args: []string{"The quick brown fox jumps over the lazy dog"},
			opts: HMACOptions{Algorithm: "sha256", Key: "key"},
			want: "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8\n",
		},
		{
			name:  "sha256 stdin matches arg",
			stdin: "The quick brown fox jumps over the lazy dog",
			opts:  HMACOptions{Algorithm: "sha256", Key: "key"},
			want:  "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8\n",
		},
		{
			name: "default algorithm is sha256",
			args: []string{"The quick brown fox jumps over the lazy dog"},
			opts: HMACOptions{Key: "key"},
			want: "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8\n",
		},
		{
			name:    "empty key",
			args:    []string{"msg"},
			opts:    HMACOptions{Algorithm: "sha256", Key: ""},
			wantErr: cmderr.ErrInvalidInput,
		},
		{
			name:    "unknown algorithm",
			args:    []string{"msg"},
			opts:    HMACOptions{Algorithm: "md4", Key: "secret"},
			wantErr: cmderr.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := RunHMAC(&buf, strings.NewReader(tt.stdin), tt.args, tt.opts)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("RunHMAC() error = %v, want errors.Is(..., %v)", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("RunHMAC() unexpected error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("RunHMAC() = %q, want %q", got, tt.want)
			}
		})
	}
}
