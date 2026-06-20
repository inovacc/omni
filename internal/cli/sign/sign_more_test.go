package sign_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
	clisign "github.com/inovacc/omni/internal/cli/sign"
	"github.com/inovacc/omni/pkg/sign"
)

// writeLowCostKeys generates a low-cost keypair on disk and returns its paths.
func writeLowCostKeys(t *testing.T, dir, pass string) (pub, key string) {
	t.Helper()
	kp, err := sign.GenerateKeyPair(pass, sign.WithScryptParams(1<<15, 8, 1))
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	keyText, err := kp.SecretKey.MarshalText(pass, sign.WithScryptParams(1<<15, 8, 1))
	if err != nil {
		t.Fatalf("MarshalText: %v", err)
	}
	pub = filepath.Join(dir, "k.pub")
	key = filepath.Join(dir, "k.key")
	if err := os.WriteFile(pub, kp.PublicKey.MarshalText(), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(key, keyText, 0o600); err != nil {
		t.Fatal(err)
	}
	return pub, key
}

func TestRunSignStdin(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("OMNI_SIGN_PASSPHRASE", "pw")
	_, key := writeLowCostKeys(t, dir, "pw")

	sigPath := filepath.Join(dir, "out.minisig")
	var out bytes.Buffer
	err := clisign.RunSign(&out, strings.NewReader("piped data"), []string{"-"}, clisign.SignOptions{
		KeyPath: key,
		SigPath: sigPath,
	})
	if err != nil {
		t.Fatalf("RunSign stdin: %v", err)
	}
	if _, err := os.Stat(sigPath); err != nil {
		t.Fatalf("expected signature file: %v", err)
	}
}

func TestRunSignStdinMissingSig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("OMNI_SIGN_PASSPHRASE", "pw")
	_, key := writeLowCostKeys(t, dir, "pw")
	err := clisign.RunSign(&bytes.Buffer{}, strings.NewReader("x"), []string{"-"}, clisign.SignOptions{KeyPath: key})
	if err == nil {
		t.Fatal("expected error: --sig required for stdin")
	}
}

func TestRunSignNilReader(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("OMNI_SIGN_PASSPHRASE", "pw")
	_, key := writeLowCostKeys(t, dir, "pw")
	err := clisign.RunSign(&bytes.Buffer{}, nil, nil, clisign.SignOptions{KeyPath: key, SigPath: filepath.Join(dir, "s")})
	if err == nil {
		t.Fatal("expected error: no input")
	}
}

func TestRunSignNoKey(t *testing.T) {
	err := clisign.RunSign(&bytes.Buffer{}, nil, []string{"f"}, clisign.SignOptions{})
	if err == nil || !errorsContains(err, "key") {
		t.Fatalf("expected --key required error, got %v", err)
	}
}

func TestRunSignMissingArtifact(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("OMNI_SIGN_PASSPHRASE", "pw")
	_, key := writeLowCostKeys(t, dir, "pw")
	err := clisign.RunSign(&bytes.Buffer{}, nil, []string{filepath.Join(dir, "absent.bin")}, clisign.SignOptions{KeyPath: key})
	if err == nil {
		t.Fatal("expected not-found error for missing artifact")
	}
	if !errors.Is(err, cmderr.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRunSignWithComments(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("OMNI_SIGN_PASSPHRASE", "pw")
	_, key := writeLowCostKeys(t, dir, "pw")
	artifact := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(artifact, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := clisign.RunSign(&out, nil, []string{artifact}, clisign.SignOptions{
		KeyPath:          key,
		TrustedComment:   "trusted",
		UntrustedComment: "untrusted",
	})
	if err != nil {
		t.Fatalf("RunSign with comments: %v", err)
	}
	if _, err := os.Stat(artifact + ".minisig"); err != nil {
		t.Fatalf("expected default .minisig: %v", err)
	}
}

func errorsContains(err error, substr string) bool {
	return err != nil && strings.Contains(err.Error(), substr)
}
