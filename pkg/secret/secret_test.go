package secret_test

import (
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/inovacc/omni/pkg/secret"
)

func TestKeyRedactsEverywhere(t *testing.T) {
	raw := []byte("super-secret-ed25519-bytes")
	k := secret.New(raw)
	for name, got := range map[string]string{
		"String":   k.String(),
		"%v":       fmt.Sprintf("%v", k),
		"%s":       fmt.Sprintf("%s", k),
		"%#v":      fmt.Sprintf("%#v", k),
		"GoString": k.GoString(),
	} {
		if strings.Contains(got, "super-secret") {
			t.Errorf("%s leaked key material: %q", name, got)
		}
		if !strings.Contains(got, "REDACTED") {
			t.Errorf("%s = %q, want a REDACTED placeholder", name, got)
		}
	}
}

func TestKeySlogRedacts(t *testing.T) {
	var buf strings.Builder
	l := slog.New(slog.NewTextHandler(&buf, nil))
	l.Info("signing", "key", secret.New([]byte("super-secret")))
	if strings.Contains(buf.String(), "super-secret") {
		t.Errorf("slog leaked key material: %q", buf.String())
	}
}

func TestKeyBytesRoundTrip(t *testing.T) {
	raw := []byte{1, 2, 3, 4}
	k := secret.New(raw)
	if got := k.Bytes(); string(got) != string(raw) {
		t.Errorf("Bytes() = %v, want %v", got, raw)
	}
}
