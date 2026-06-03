package sign_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
	clisign "github.com/inovacc/omni/internal/cli/sign"
	"github.com/inovacc/omni/internal/cli/verify"
	"github.com/inovacc/omni/pkg/sign"
)

// lowCost mirrors the SENSITIVE-cost guard: every test MUST use low scrypt cost.
const lowN, lowR, lowP = 1 << 15, 8, 1

// writeFixtureKeys generates a low-cost keypair and writes the .pub/.key files,
// returning their paths. It NEVER uses default-cost GenerateKeyPair.
func writeFixtureKeys(t *testing.T, dir, passphrase string) (pubPath, keyPath string) {
	t.Helper()
	kp, err := sign.GenerateKeyPair(passphrase, sign.WithScryptParams(lowN, lowR, lowP))
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	pubText := kp.PublicKey.MarshalText()
	keyText, err := kp.SecretKey.MarshalText(passphrase, sign.WithScryptParams(lowN, lowR, lowP))
	if err != nil {
		t.Fatalf("MarshalText sk: %v", err)
	}
	pubPath = filepath.Join(dir, "test.pub")
	keyPath = filepath.Join(dir, "test.key")
	if err := os.WriteFile(pubPath, pubText, 0o644); err != nil {
		t.Fatalf("write pub: %v", err)
	}
	if err := os.WriteFile(keyPath, keyText, 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}
	return pubPath, keyPath
}

func TestRunKeygenWritesFiles(t *testing.T) {
	dir := t.TempDir()
	pubPath := filepath.Join(dir, "k.pub")
	keyPath := filepath.Join(dir, "k.key")
	t.Setenv("OMNI_SIGN_PASSPHRASE", "pw")

	var out bytes.Buffer
	err := clisign.RunKeygen(&out, clisign.KeygenOptions{
		PubPath:   pubPath,
		KeyPath:   keyPath,
		ScryptN:   lowN,
		ScryptR:   lowR,
		ScryptP:   lowP,
		LowCostOK: true,
	})
	if err != nil {
		t.Fatalf("RunKeygen: %v", err)
	}
	if _, err := os.Stat(pubPath); err != nil {
		t.Errorf("public key not written: %v", err)
	}
	if _, err := os.Stat(keyPath); err != nil {
		t.Errorf("secret key not written: %v", err)
	}
	// The generated key must be parseable with the passphrase.
	keyText, _ := os.ReadFile(keyPath)
	if _, err := sign.ParseSecretKey(keyText, "pw"); err != nil {
		t.Errorf("generated key not parseable: %v", err)
	}
}

func TestRunSignWritesMinisigAndVerifies(t *testing.T) {
	dir := t.TempDir()
	pubPath, keyPath := writeFixtureKeys(t, dir, "pw")
	t.Setenv("OMNI_SIGN_PASSPHRASE", "pw")

	dataPath := filepath.Join(dir, "artifact.bin")
	if err := os.WriteFile(dataPath, []byte("artifact bytes"), 0o644); err != nil {
		t.Fatalf("write data: %v", err)
	}

	var out bytes.Buffer
	sigPath := dataPath + ".minisig"
	err := clisign.RunSign(&out, nil, []string{dataPath}, clisign.SignOptions{
		KeyPath:        keyPath,
		SigPath:        sigPath,
		TrustedComment: "ts:1",
	})
	if err != nil {
		t.Fatalf("RunSign: %v", err)
	}
	if _, err := os.Stat(sigPath); err != nil {
		t.Fatalf("signature not written: %v", err)
	}

	// RunVerify must return nil for the good signature.
	var vout bytes.Buffer
	verr := verify.RunVerify(&vout, nil, []string{dataPath}, verify.VerifyOptions{
		PubPath: pubPath,
		SigPath: sigPath,
	})
	if verr != nil {
		t.Fatalf("RunVerify(valid) = %v, want nil", verr)
	}
}

func TestRunVerifyTamperedIsConflict(t *testing.T) {
	dir := t.TempDir()
	pubPath, keyPath := writeFixtureKeys(t, dir, "pw")
	t.Setenv("OMNI_SIGN_PASSPHRASE", "pw")

	dataPath := filepath.Join(dir, "artifact.bin")
	if err := os.WriteFile(dataPath, []byte("artifact bytes"), 0o644); err != nil {
		t.Fatalf("write data: %v", err)
	}
	sigPath := dataPath + ".minisig"
	var out bytes.Buffer
	if err := clisign.RunSign(&out, nil, []string{dataPath}, clisign.SignOptions{
		KeyPath: keyPath,
		SigPath: sigPath,
	}); err != nil {
		t.Fatalf("RunSign: %v", err)
	}

	// Tamper with the payload after signing.
	if err := os.WriteFile(dataPath, []byte("artifact byteX"), 0o644); err != nil {
		t.Fatalf("tamper: %v", err)
	}

	var vout bytes.Buffer
	verr := verify.RunVerify(&vout, nil, []string{dataPath}, verify.VerifyOptions{
		PubPath: pubPath,
		SigPath: sigPath,
	})
	if verr == nil {
		t.Fatal("RunVerify(tampered) = nil, want ErrConflict")
	}
	if !cmderr.IsConflict(verr) {
		t.Errorf("RunVerify(tampered) error = %v, want cmderr.ErrConflict", verr)
	}
}

func TestRunVerifyBadKeyIsInvalidInput(t *testing.T) {
	dir := t.TempDir()
	_, keyPath := writeFixtureKeys(t, dir, "pw")
	t.Setenv("OMNI_SIGN_PASSPHRASE", "pw")

	dataPath := filepath.Join(dir, "artifact.bin")
	if err := os.WriteFile(dataPath, []byte("artifact bytes"), 0o644); err != nil {
		t.Fatalf("write data: %v", err)
	}
	sigPath := dataPath + ".minisig"
	var out bytes.Buffer
	if err := clisign.RunSign(&out, nil, []string{dataPath}, clisign.SignOptions{
		KeyPath: keyPath,
		SigPath: sigPath,
	}); err != nil {
		t.Fatalf("RunSign: %v", err)
	}

	// A malformed public key must classify as ErrInvalidInput.
	badPub := filepath.Join(dir, "bad.pub")
	if err := os.WriteFile(badPub, []byte("untrusted comment: bad\nnot-valid-base64!!!\n"), 0o644); err != nil {
		t.Fatalf("write bad pub: %v", err)
	}

	var vout bytes.Buffer
	verr := verify.RunVerify(&vout, nil, []string{dataPath}, verify.VerifyOptions{
		PubPath: badPub,
		SigPath: sigPath,
	})
	if verr == nil {
		t.Fatal("RunVerify(bad key) = nil, want ErrInvalidInput")
	}
	if !cmderr.IsInvalidInput(verr) {
		t.Errorf("RunVerify(bad key) error = %v, want cmderr.ErrInvalidInput", verr)
	}
}

func TestRunVerifyMissingKeyIsNotFound(t *testing.T) {
	dir := t.TempDir()
	pubPath, keyPath := writeFixtureKeys(t, dir, "pw")
	_ = pubPath
	t.Setenv("OMNI_SIGN_PASSPHRASE", "pw")

	dataPath := filepath.Join(dir, "artifact.bin")
	if err := os.WriteFile(dataPath, []byte("artifact bytes"), 0o644); err != nil {
		t.Fatalf("write data: %v", err)
	}
	sigPath := dataPath + ".minisig"
	var out bytes.Buffer
	if err := clisign.RunSign(&out, nil, []string{dataPath}, clisign.SignOptions{
		KeyPath: keyPath,
		SigPath: sigPath,
	}); err != nil {
		t.Fatalf("RunSign: %v", err)
	}

	var vout bytes.Buffer
	verr := verify.RunVerify(&vout, nil, []string{dataPath}, verify.VerifyOptions{
		PubPath: filepath.Join(dir, "does-not-exist.pub"),
		SigPath: sigPath,
	})
	if verr == nil {
		t.Fatal("RunVerify(missing key) = nil, want ErrNotFound")
	}
	if !cmderr.IsNotFound(verr) {
		t.Errorf("RunVerify(missing key) error = %v, want cmderr.ErrNotFound", verr)
	}
}

func TestRunSignRejectsInlineKeyMaterial(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("OMNI_SIGN_PASSPHRASE", "pw")
	dataPath := filepath.Join(dir, "artifact.bin")
	if err := os.WriteFile(dataPath, []byte("data"), 0o644); err != nil {
		t.Fatalf("write data: %v", err)
	}

	// A --key value that is not a real file path (looks like inline key material)
	// must be rejected as invalid input, never used as key material.
	var out bytes.Buffer
	err := clisign.RunSign(&out, nil, []string{dataPath}, clisign.SignOptions{
		KeyPath: "untrusted comment: x\nRWQAAAAAAAAAAAAA",
		SigPath: filepath.Join(dir, "out.minisig"),
	})
	if err == nil {
		t.Fatal("RunSign(inline key) = nil, want error")
	}
	if !cmderr.IsInvalidInput(err) && !cmderr.IsNotFound(err) {
		t.Errorf("RunSign(inline key) error = %v, want ErrInvalidInput/ErrNotFound", err)
	}
}
