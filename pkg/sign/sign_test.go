package sign_test

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/inovacc/omni/pkg/sign"
)

// missingGlobalLine returns a copy of a .minisig with the 4th line (global
// signature) removed, so a fail-closed verifier must reject it.
func missingGlobalLine(minisig []byte) []byte {
	lines := strings.Split(strings.TrimRight(string(minisig), "\n"), "\n")
	if len(lines) < 4 {
		return minisig
	}
	lines = lines[:3]
	return []byte(strings.Join(lines, "\n") + "\n")
}

// sigLine2Algo decodes the base64 of line 2 of a .minisig and returns its
// leading 2 algorithm bytes.
func sigLine2Algo(t *testing.T, minisig []byte) [2]byte {
	t.Helper()
	lines := strings.Split(strings.TrimRight(string(minisig), "\n"), "\n")
	if len(lines) < 2 {
		t.Fatalf("minisig must have >=2 lines, got %d: %q", len(lines), minisig)
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(lines[1]))
	if err != nil {
		t.Fatalf("decode line 2: %v", err)
	}
	if len(raw) < 2 {
		t.Fatalf("line 2 too short: %d bytes", len(raw))
	}
	return [2]byte{raw[0], raw[1]}
}

// lowCost is the scrypt cost used by every test. The default SENSITIVE cost
// (N=2^20, ~1 GiB) would OOM/hang the test runner, so tests MUST use this.
func lowCost() sign.Option { return sign.WithScryptParams(1<<15, 8, 1) }

func TestGenerateKeyPairRoundTrip(t *testing.T) {
	kp, err := sign.GenerateKeyPair("test-passphrase", lowCost())
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	pubText := kp.PublicKey.MarshalText()
	skText, err := kp.SecretKey.MarshalText("test-passphrase", lowCost())
	if err != nil {
		t.Fatalf("MarshalText sk: %v", err)
	}

	pub2, err := sign.ParsePublicKey(pubText)
	if err != nil || pub2.KeyID != kp.PublicKey.KeyID {
		t.Fatalf("public round-trip: %v", err)
	}
	sk2, err := sign.ParseSecretKey(skText, "test-passphrase")
	if err != nil {
		t.Fatalf("secret round-trip: %v", err)
	}
	if sk2.KeyID != kp.SecretKey.KeyID {
		t.Fatalf("secret key id mismatch: got %x want %x", sk2.KeyID, kp.SecretKey.KeyID)
	}
	if _, err := sign.ParseSecretKey(skText, "WRONG"); err == nil {
		t.Error("wrong passphrase must fail (checksum mismatch)")
	}
}

// TestSecretKeyRoundTripAcrossCosts guards against the limit<->param mapping
// rejecting a correct passphrase at low scrypt costs (regression: small N once
// round-tripped opslimit/memlimit into a different p, corrupting the XOR
// stream so the correct passphrase failed the checksum).
func TestSecretKeyRoundTripAcrossCosts(t *testing.T) {
	for _, n := range []int{1 << 4, 1 << 9, 1 << 10, 1 << 12, 1 << 15} {
		kp, err := sign.GenerateKeyPair("pw", sign.WithScryptParams(n, 8, 1))
		if err != nil {
			t.Fatalf("N=%d GenerateKeyPair: %v", n, err)
		}
		txt, err := kp.SecretKey.MarshalText("pw", sign.WithScryptParams(n, 8, 1))
		if err != nil {
			t.Fatalf("N=%d MarshalText: %v", n, err)
		}
		if _, err := sign.ParseSecretKey(txt, "pw"); err != nil {
			t.Errorf("N=%d: correct passphrase rejected on round-trip: %v", n, err)
		}
		if _, err := sign.ParseSecretKey(txt, "nope"); err == nil {
			t.Errorf("N=%d: wrong passphrase accepted", n)
		}
	}
}

func TestPublicKeyTextFormat(t *testing.T) {
	kp, err := sign.GenerateKeyPair("p", lowCost())
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	got := string(kp.PublicKey.MarshalText())
	lines := strings.Split(strings.TrimRight(got, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("public key should be 2 lines, got %d: %q", len(lines), got)
	}
	if !strings.HasPrefix(lines[0], "untrusted comment: ") {
		t.Errorf("first line must be an untrusted comment, got %q", lines[0])
	}
}

func TestSignProducesPrehashedMinisig(t *testing.T) {
	kp, err := sign.GenerateKeyPair("p", lowCost())
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	data := []byte("artifact bytes")
	sig, err := sign.Sign(data, kp.SecretKey, sign.WithTrustedComment("ts:1"))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	// A .minisig is 4 lines: comment, sig, trusted comment, global sig.
	lines := strings.Split(strings.TrimRight(string(sig), "\n"), "\n")
	if len(lines) != 4 {
		t.Fatalf("minisig must have 4 lines, got %d: %q", len(lines), sig)
	}
	if !strings.HasPrefix(lines[0], "untrusted comment: ") {
		t.Errorf("line 1 must be an untrusted comment, got %q", lines[0])
	}
	if !strings.HasPrefix(lines[2], "trusted comment: ") {
		t.Errorf("line 3 must be a trusted comment, got %q", lines[2])
	}
	if got := strings.TrimPrefix(lines[2], "trusted comment: "); got != "ts:1" {
		t.Errorf("trusted comment = %q, want %q", got, "ts:1")
	}

	// Default new signatures MUST use the prehashed algorithm "ED".
	if algo := sigLine2Algo(t, sig); algo != [2]byte{'E', 'D'} {
		t.Errorf("sig algo = %q, want %q (prehashed)", algo, [2]byte{'E', 'D'})
	}
}

func TestVerifyValid(t *testing.T) {
	kp, err := sign.GenerateKeyPair("p", lowCost())
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	data := []byte("artifact bytes")
	sig, err := sign.Sign(data, kp.SecretKey, sign.WithTrustedComment("ts:1"))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if err := sign.Verify(data, sig, kp.PublicKey); err != nil {
		t.Fatalf("Verify(valid) = %v, want nil", err)
	}
}

// flipByte returns a copy of a .minisig with one byte of the base64 signature
// payload (line 2) flipped, so the decoded signature no longer matches.
func flipByte(minisig []byte) []byte {
	lines := strings.Split(strings.TrimRight(string(minisig), "\n"), "\n")
	if len(lines) < 2 {
		return minisig
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(lines[1]))
	if err != nil || len(raw) == 0 {
		return minisig
	}
	// Flip a byte well inside the signature (past sig_alg+key_id).
	idx := len(raw) - 1
	raw[idx] ^= 0xff
	lines[1] = base64.StdEncoding.EncodeToString(raw)
	return []byte(strings.Join(lines, "\n") + "\n")
}

// withAlgo returns a copy of a .minisig whose line-2 sig_alg bytes are replaced
// with algo, leaving everything else intact.
func withAlgo(minisig []byte, algo [2]byte) []byte {
	lines := strings.Split(strings.TrimRight(string(minisig), "\n"), "\n")
	if len(lines) < 2 {
		return minisig
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(lines[1]))
	if err != nil || len(raw) < 2 {
		return minisig
	}
	raw[0], raw[1] = algo[0], algo[1]
	lines[1] = base64.StdEncoding.EncodeToString(raw)
	return []byte(strings.Join(lines, "\n") + "\n")
}

// flipTrustedComment returns a copy of a .minisig with its trusted comment text
// altered, which must invalidate the global signature.
func flipTrustedComment(minisig []byte) []byte {
	lines := strings.Split(strings.TrimRight(string(minisig), "\n"), "\n")
	if len(lines) < 4 {
		return minisig
	}
	lines[2] = "trusted comment: tampered"
	return []byte(strings.Join(lines, "\n") + "\n")
}

func TestVerifyFailsClosed(t *testing.T) {
	kp, err := sign.GenerateKeyPair("p", lowCost())
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	other, err := sign.GenerateKeyPair("p", lowCost())
	if err != nil {
		t.Fatalf("GenerateKeyPair other: %v", err)
	}
	data := []byte("payload")
	good, err := sign.Sign(data, kp.SecretKey, sign.WithTrustedComment("t"))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	cases := map[string]func() error{
		"wrong key":            func() error { return sign.Verify(data, good, other.PublicKey) },
		"tampered payload":     func() error { return sign.Verify([]byte("payloaX"), good, kp.PublicKey) },
		"tampered sig":         func() error { return sign.Verify(data, flipByte(good), kp.PublicKey) },
		"tampered trusted cmt": func() error { return sign.Verify(data, flipTrustedComment(good), kp.PublicKey) },
		"empty sig":            func() error { return sign.Verify(data, []byte{}, kp.PublicKey) },
		"garbage sig":          func() error { return sign.Verify(data, []byte("not base64!!!"), kp.PublicKey) },
		"bad algo":             func() error { return sign.Verify(data, withAlgo(good, [2]byte{'X', 'x'}), kp.PublicKey) },
		"truncated sig payload": func() error {
			return sign.Verify(data, []byte("untrusted comment: x\nQQ==\ntrusted comment: t\nQQ==\n"), kp.PublicKey)
		},
		"missing global sigline": func() error { return sign.Verify(data, missingGlobalLine(good), kp.PublicKey) },
	}
	for name, fn := range cases {
		if err := fn(); err == nil {
			t.Errorf("%s: Verify = nil, want error (fail-closed)", name)
		}
	}
}

// TestVerifyErrorsAreVerificationSentinel ensures every fail-closed path returns
// an error that errors.Is(err, sign.ErrVerification) — the CLI maps that
// sentinel to cmderr.ErrConflict.
func TestVerifyErrorsAreVerificationSentinel(t *testing.T) {
	kp, err := sign.GenerateKeyPair("p", lowCost())
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	other, err := sign.GenerateKeyPair("p", lowCost())
	if err != nil {
		t.Fatalf("GenerateKeyPair other: %v", err)
	}
	data := []byte("payload")
	good, err := sign.Sign(data, kp.SecretKey, sign.WithTrustedComment("t"))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	for name, gotErr := range map[string]error{
		"wrong key":        sign.Verify(data, good, other.PublicKey),
		"tampered payload": sign.Verify([]byte("payloaX"), good, kp.PublicKey),
		"tampered sig":     sign.Verify(data, flipByte(good), kp.PublicKey),
		"garbage sig":      sign.Verify(data, []byte("not base64!!!"), kp.PublicKey),
		"bad algo":         sign.Verify(data, withAlgo(good, [2]byte{'X', 'x'}), kp.PublicKey),
	} {
		if gotErr == nil {
			t.Errorf("%s: want error, got nil", name)
			continue
		}
		if !errors.Is(gotErr, sign.ErrVerification) {
			t.Errorf("%s: errors.Is(err, ErrVerification) = false, err = %v", name, gotErr)
		}
	}
}

// --- Adversarial fail-closed helpers (Task 6 security review) ---

// decodeLine2 returns the decoded base64 payload of line 2 of a .minisig.
func decodeLine2(t *testing.T, minisig []byte) []byte {
	t.Helper()
	lines := strings.Split(strings.TrimRight(string(minisig), "\n"), "\n")
	if len(lines) < 2 {
		t.Fatalf("minisig must have >=2 lines, got %d", len(lines))
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(lines[1]))
	if err != nil {
		t.Fatalf("decode line 2: %v", err)
	}
	return raw
}

// withLine replaces line index i (0-based) of a .minisig with s and re-joins.
func withLine(minisig []byte, i int, s string) []byte {
	lines := strings.Split(strings.TrimRight(string(minisig), "\n"), "\n")
	if i < 0 || i >= len(lines) {
		return minisig
	}
	lines[i] = s
	return []byte(strings.Join(lines, "\n") + "\n")
}

// emptyGlobalSigLine returns a copy with line 4 (global sig) emptied. The 4-line
// shape is preserved on disk but splitLines drops the trailing empty line, so a
// fail-closed verifier must still reject (too few lines / zero-length global).
func emptyGlobalSigLine(minisig []byte) []byte {
	return withLine(minisig, 3, "")
}

// zeroSignaturePayload returns a copy whose line-2 detached signature (64 bytes
// after sig_alg+key_id) is all zeros, keeping sig_alg and key_id intact.
func zeroSignaturePayload(t *testing.T, minisig []byte) []byte {
	raw := decodeLine2(t, minisig)
	for i := 2 + 8; i < len(raw); i++ {
		raw[i] = 0
	}
	return withLine(minisig, 1, base64.StdEncoding.EncodeToString(raw))
}

// zeroGlobalSigLine returns a copy whose line-4 global signature is 64 zero
// bytes (valid length, invalid crypto).
func zeroGlobalSigLine(minisig []byte) []byte {
	return withLine(minisig, 3, base64.StdEncoding.EncodeToString(make([]byte, 64)))
}

// newlineInBase64 returns a copy with a newline injected into the middle of the
// line-2 base64. A naive line splitter would mis-index the file; a fail-closed
// verifier must reject (shifted/short lines).
func newlineInBase64(minisig []byte) []byte {
	lines := strings.Split(strings.TrimRight(string(minisig), "\n"), "\n")
	if len(lines) < 2 || len(lines[1]) < 4 {
		return minisig
	}
	mid := len(lines[1]) / 2
	lines[1] = lines[1][:mid] + "\n" + lines[1][mid:]
	return []byte(strings.Join(lines, "\n") + "\n")
}

// overLongSigPayload returns a copy whose line-2 payload has 8 extra trailing
// bytes (wrong length for sig_alg||key_id||64-byte-sig).
func overLongSigPayload(t *testing.T, minisig []byte) []byte {
	raw := decodeLine2(t, minisig)
	raw = append(raw, make([]byte, 8)...)
	return withLine(minisig, 1, base64.StdEncoding.EncodeToString(raw))
}

// swapGlobalForDataSig returns a copy whose line-4 global signature is replaced
// with the line-2 detached signature (a valid 64-byte Ed25519 sig, but over the
// wrong message). Guards against accepting any well-formed 64-byte blob.
func swapGlobalForDataSig(t *testing.T, minisig []byte) []byte {
	raw := decodeLine2(t, minisig)
	dataSig := raw[2+8:]
	return withLine(minisig, 3, base64.StdEncoding.EncodeToString(dataSig))
}

func TestVerifyAdversarialFailsClosed(t *testing.T) {
	kp, err := sign.GenerateKeyPair("p", lowCost())
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	other, err := sign.GenerateKeyPair("p", lowCost())
	if err != nil {
		t.Fatalf("GenerateKeyPair other: %v", err)
	}
	data := []byte("payload-under-attack")
	good, err := sign.Sign(data, kp.SecretKey, sign.WithTrustedComment("release v1.2.3"))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	// A signature over completely different data (same key) must not verify
	// against `data`.
	otherData, err := sign.Sign([]byte("different artifact"), kp.SecretKey, sign.WithTrustedComment("release v1.2.3"))
	if err != nil {
		t.Fatalf("Sign otherData: %v", err)
	}

	// Positive control: the untampered signature MUST verify, otherwise every
	// negative case below would pass for the wrong reason.
	if err := sign.Verify(data, good, kp.PublicKey); err != nil {
		t.Fatalf("positive control: Verify(good) = %v, want nil", err)
	}
	// Guard against no-op mutation helpers giving false PASSes: the mutated
	// bytes for the trickier cases must actually differ from `good`.
	for name, mutated := range map[string][]byte{
		"flipTrustedComment":   flipTrustedComment(good),
		"swapGlobalForDataSig": swapGlobalForDataSig(t, good),
		"zeroGlobalSigLine":    zeroGlobalSigLine(good),
		"newlineInBase64":      newlineInBase64(good),
		"zeroSignaturePayload": zeroSignaturePayload(t, good),
		"overLongSigPayload":   overLongSigPayload(t, good),
		"relabel Ed":           withAlgo(good, [2]byte{'E', 'd'}),
	} {
		if string(mutated) == string(good) {
			t.Fatalf("helper %s did not mutate the signature (would be a false PASS)", name)
		}
	}

	cases := map[string]func() error{
		// key_id mismatch must be rejected before any crypto.
		"key_id mismatch": func() error { return sign.Verify(data, good, other.PublicKey) },
		// missing / empty / zero global signature.
		"empty global sig line":  func() error { return sign.Verify(data, emptyGlobalSigLine(good), kp.PublicKey) },
		"missing global sigline": func() error { return sign.Verify(data, missingGlobalLine(good), kp.PublicKey) },
		"zero global sig":        func() error { return sign.Verify(data, zeroGlobalSigLine(good), kp.PublicKey) },
		// trusted comment swapped while the data signature stays valid.
		"trusted comment tampered": func() error { return sign.Verify(data, flipTrustedComment(good), kp.PublicKey) },
		// global sig replaced by the (valid-length, wrong-message) data sig.
		"global sig = data sig": func() error { return sign.Verify(data, swapGlobalForDataSig(t, good), kp.PublicKey) },
		// legacy "Ed" vs prehashed "ED" confusion: a prehashed sig relabeled raw.
		"prehashed relabeled as legacy Ed": func() error { return sign.Verify(data, withAlgo(good, [2]byte{'E', 'd'}), kp.PublicKey) },
		"unknown lowercase algo":           func() error { return sign.Verify(data, withAlgo(good, [2]byte{'e', 'd'}), kp.PublicKey) },
		// truncated / over-long signature buffers.
		"truncated line-2 payload": func() error {
			return sign.Verify(data, []byte("untrusted comment: x\nQQ==\ntrusted comment: t\nQQ==\n"), kp.PublicKey)
		},
		"over-long line-2 payload": func() error { return sign.Verify(data, overLongSigPayload(t, good), kp.PublicKey) },
		// all-zero detached signature.
		"zero data signature": func() error { return sign.Verify(data, zeroSignaturePayload(t, good), kp.PublicKey) },
		// base64 with an embedded newline (line-shifting attack).
		"newline in base64": func() error { return sign.Verify(data, newlineInBase64(good), kp.PublicKey) },
		// nil / empty / wrong-length public key.
		"nil public key":   func() error { return sign.Verify(data, good, sign.PublicKey{KeyID: kp.PublicKey.KeyID, Pub: nil}) },
		"empty public key": func() error { return sign.Verify(data, good, sign.PublicKey{KeyID: kp.PublicKey.KeyID, Pub: []byte{}}) },
		"short public key": func() error {
			return sign.Verify(data, good, sign.PublicKey{KeyID: kp.PublicKey.KeyID, Pub: make([]byte, 16)})
		},
		"zero public key": func() error {
			return sign.Verify(data, good, sign.PublicKey{KeyID: kp.PublicKey.KeyID, Pub: make([]byte, 32)})
		},
		// signature for different data verified against this data.
		"signature for different data": func() error { return sign.Verify(data, otherData, kp.PublicKey) },
		// empty / garbage / nil inputs must not panic and must error.
		"empty sig":   func() error { return sign.Verify(data, []byte{}, kp.PublicKey) },
		"nil sig":     func() error { return sign.Verify(data, nil, kp.PublicKey) },
		"garbage sig": func() error { return sign.Verify(data, []byte("not base64!!!"), kp.PublicKey) },
	}

	for name, fn := range cases {
		name, fn := name, fn
		t.Run(name, func(t *testing.T) {
			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("Verify panicked (fail-open via panic): %v", r)
					}
				}()
				err = fn()
			}()
			if err == nil {
				t.Errorf("%s: Verify = nil, want non-nil error (fail-closed)", name)
				return
			}
			if !errors.Is(err, sign.ErrVerification) {
				t.Errorf("%s: errors.Is(err, ErrVerification) = false, err = %v", name, err)
			}
		})
	}
}

func TestSecretKeyNeverLeaksMaterial(t *testing.T) {
	kp, err := sign.GenerateKeyPair("p", lowCost())
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	// The SecretKey's secret.Key wrapper must redact under %v/%#v.
	for _, format := range []string{"%v", "%s", "%#v"} {
		got := fmt.Sprintf(format, kp.SecretKey)
		if !strings.Contains(got, "REDACTED") {
			t.Errorf("SecretKey %s = %q, want a REDACTED placeholder", format, got)
		}
	}
}
