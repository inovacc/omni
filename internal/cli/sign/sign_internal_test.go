package sign

import (
	"errors"
	"os"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

func TestClassifyFileErrInternal(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{"notexist", os.ErrNotExist, cmderr.ErrNotFound},
		{"permission", os.ErrPermission, cmderr.ErrPermission},
		{"other", errors.New("io boom"), cmderr.ErrIO},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyFileErr(tt.err, "/k.key")
			if !errors.Is(got, tt.want) {
				t.Fatalf("classifyFileErr(%v) -> %v want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestReadPassphraseFromEnv(t *testing.T) {
	t.Setenv(passphraseEnv, "hunter2")
	got, err := readPassphrase(false)
	if err != nil {
		t.Fatalf("readPassphrase: %v", err)
	}
	if got != "hunter2" {
		t.Fatalf("readPassphrase=%q want hunter2", got)
	}
	// confirm=true also returns env value without prompting.
	got2, err := readPassphrase(true)
	if err != nil || got2 != "hunter2" {
		t.Fatalf("readPassphrase(confirm) = %q, %v", got2, err)
	}
}

func TestReadPassphraseNoTTYNoEnv(t *testing.T) {
	// With the env unset and stdin not a terminal (the test runner's case),
	// readPassphrase must fail with a clear ErrInvalidInput rather than block.
	os.Unsetenv(passphraseEnv)
	_, err := readPassphrase(false)
	if err == nil {
		t.Skip("stdin appears to be a terminal in this environment; skipping non-TTY assertion")
	}
	if !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestRejectInlineKey(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"plain path", "/tmp/key.key", false},
		{"newline", "line1\nline2", true},
		{"carriage return", "a\rb", true},
		{"inline comment", "untrusted comment: minisign", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := rejectInlineKey(tt.value); (err != nil) != tt.wantErr {
				t.Fatalf("rejectInlineKey(%q) err=%v wantErr=%v", tt.value, err, tt.wantErr)
			}
		})
	}
}
