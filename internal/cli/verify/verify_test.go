package verify

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/sign"
)

// lowCost keeps scrypt cheap so the key-marshal step in these tests is fast.
// The default SENSITIVE cost needs ~1 GiB RAM and several seconds.
func lowCost() sign.Option { return sign.WithScryptParams(1<<15, 8, 1) }

// signedFixture writes an artifact, its detached *.minisig signature and the
// matching *.pub public key into dir, returning their paths. All key material
// is generated at test time; nothing is read from the network or disk.
func signedFixture(t *testing.T, dir string, data []byte) (artifact, sig, pub string) {
	t.Helper()
	kp, err := sign.GenerateKeyPair("pw", lowCost())
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	sigBytes, err := sign.Sign(data, kp.SecretKey, sign.WithTrustedComment("ts:1"))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	artifact = filepath.Join(dir, "artifact.bin")
	sig = artifact + ".minisig"
	pub = filepath.Join(dir, "minisign.pub")

	if err := os.WriteFile(artifact, data, 0o600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	if err := os.WriteFile(sig, sigBytes, 0o600); err != nil {
		t.Fatalf("write sig: %v", err)
	}
	if err := os.WriteFile(pub, kp.PublicKey.MarshalText(), 0o600); err != nil {
		t.Fatalf("write pub: %v", err)
	}
	return artifact, sig, pub
}

// TestRunVerifySuccess exercises the happy path end-to-end with a freshly
// generated key pair, asserting both nil error and the success banner.
func TestRunVerifySuccess(t *testing.T) {
	dir := t.TempDir()
	artifact, sig, pub := signedFixture(t, dir, []byte("hello omni"))

	var buf bytes.Buffer
	err := RunVerify(&buf, nil, []string{artifact}, VerifyOptions{PubPath: pub, SigPath: sig})
	if err != nil {
		t.Fatalf("RunVerify success path error = %v, want nil", err)
	}
	if !strings.Contains(buf.String(), "verified") {
		t.Errorf("output = %q, want it to mention 'verified'", buf.String())
	}
}

// TestRunVerifyDefaultSigPath asserts the default signature path of
// "<artifact>.minisig" is used when SigPath is empty.
func TestRunVerifyDefaultSigPath(t *testing.T) {
	dir := t.TempDir()
	artifact, _, pub := signedFixture(t, dir, []byte("default sig path"))

	var buf bytes.Buffer
	// SigPath intentionally left empty -> defaults to artifact+".minisig".
	err := RunVerify(&buf, nil, []string{artifact}, VerifyOptions{PubPath: pub})
	if err != nil {
		t.Fatalf("RunVerify default sig path error = %v, want nil", err)
	}
}

// TestRunVerifyTamperedArtifact asserts that a modified artifact fails closed
// and is classified as cmderr.ErrConflict (exit code 1).
func TestRunVerifyTamperedArtifact(t *testing.T) {
	dir := t.TempDir()
	artifact, sig, pub := signedFixture(t, dir, []byte("original bytes"))

	// Tamper with the artifact AFTER signing so the signature no longer matches.
	if err := os.WriteFile(artifact, []byte("tampered bytes"), 0o600); err != nil {
		t.Fatalf("tamper artifact: %v", err)
	}

	var buf bytes.Buffer
	err := RunVerify(&buf, nil, []string{artifact}, VerifyOptions{PubPath: pub, SigPath: sig})
	if err == nil {
		t.Fatal("RunVerify(tampered) = nil, want a verification mismatch error")
	}
	if !cmderr.IsConflict(err) {
		t.Fatalf("RunVerify(tampered) error = %v, want cmderr.ErrConflict", err)
	}
	if code := cmderr.ExitCodeFor(err); code != 1 {
		t.Fatalf("ExitCodeFor = %d, want 1 (ErrConflict)", code)
	}
}

// TestRunVerifyErrors is a table-driven sweep of the failure-classification
// branches in RunVerify that do not need a valid signed fixture.
func TestRunVerifyErrors(t *testing.T) {
	dir := t.TempDir()

	// A real, parseable public key for cases that get past the key-parse step.
	kp, err := sign.GenerateKeyPair("pw", lowCost())
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	goodPub := filepath.Join(dir, "good.pub")
	if err := os.WriteFile(goodPub, kp.PublicKey.MarshalText(), 0o600); err != nil {
		t.Fatalf("write good pub: %v", err)
	}

	// A malformed public-key file (not a minisign key).
	badPub := filepath.Join(dir, "bad.pub")
	if err := os.WriteFile(badPub, []byte("not a real key\n"), 0o600); err != nil {
		t.Fatalf("write bad pub: %v", err)
	}

	// An artifact + signature file that exist for the missing-artifact case.
	existingArtifact := filepath.Join(dir, "present.bin")
	if err := os.WriteFile(existingArtifact, []byte("x"), 0o600); err != nil {
		t.Fatalf("write existing artifact: %v", err)
	}

	missing := filepath.Join(dir, "does-not-exist.bin")

	tests := []struct {
		name    string
		args    []string
		opts    VerifyOptions
		wantIs  func(error) bool
		wantHas string
	}{
		{
			name:   "missing artifact arg",
			args:   nil,
			opts:   VerifyOptions{PubPath: goodPub},
			wantIs: cmderr.IsInvalidInput,
		},
		{
			name:   "missing key flag",
			args:   []string{existingArtifact},
			opts:   VerifyOptions{},
			wantIs: cmderr.IsInvalidInput,
		},
		{
			name:    "inline key material with newline",
			args:    []string{existingArtifact},
			opts:    VerifyOptions{PubPath: "untrusted comment: foo\nRWQ..."},
			wantIs:  cmderr.IsInvalidInput,
			wantHas: "file path",
		},
		{
			name:   "inline key material comment prefix",
			args:   []string{existingArtifact},
			opts:   VerifyOptions{PubPath: "untrusted comment: foo"},
			wantIs: cmderr.IsInvalidInput,
		},
		{
			name:   "public key file not found",
			args:   []string{existingArtifact},
			opts:   VerifyOptions{PubPath: filepath.Join(dir, "nope.pub")},
			wantIs: cmderr.IsNotFound,
		},
		{
			name:   "malformed public key",
			args:   []string{existingArtifact},
			opts:   VerifyOptions{PubPath: badPub},
			wantIs: cmderr.IsInvalidInput,
		},
		{
			name:   "signature file not found",
			args:   []string{existingArtifact},
			opts:   VerifyOptions{PubPath: goodPub, SigPath: filepath.Join(dir, "nope.minisig")},
			wantIs: cmderr.IsNotFound,
		},
		{
			name:   "artifact file not found",
			args:   []string{missing},
			opts:   VerifyOptions{PubPath: goodPub, SigPath: goodPub},
			wantIs: cmderr.IsNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := RunVerify(&buf, nil, tt.args, tt.opts)
			if err == nil {
				t.Fatalf("RunVerify(%+v) = nil, want error", tt.opts)
			}
			if !tt.wantIs(err) {
				t.Fatalf("RunVerify error = %v, classified wrong", err)
			}
			if tt.wantHas != "" && !strings.Contains(err.Error(), tt.wantHas) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantHas)
			}
		})
	}
}

// TestRejectInlineKey covers the unexported key-policy helper directly.
func TestRejectInlineKey(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"plain path", "/etc/keys/minisign.pub", false},
		{"relative path", "minisign.pub", false},
		{"newline", "line1\nline2", true},
		{"carriage return", "line1\rline2", true},
		{"comment prefix", "untrusted comment: minisign public key", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rejectInlineKey(tt.value)
			if tt.wantErr && err == nil {
				t.Fatalf("rejectInlineKey(%q) = nil, want error", tt.value)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("rejectInlineKey(%q) = %v, want nil", tt.value, err)
			}
			if tt.wantErr && !cmderr.IsInvalidInput(err) {
				t.Errorf("rejectInlineKey(%q) error = %v, want ErrInvalidInput", tt.value, err)
			}
		})
	}
}
