package sign

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/scrypt"

	"github.com/inovacc/omni/pkg/secret"
)

// ErrVerification is the sentinel returned by Verify (and key parsing on a
// checksum mismatch) for any fail-closed condition. Callers map it to an exit
// failure; the CLI wraps it in cmderr.ErrConflict.
var ErrVerification = errors.New("sign: verification failed")

// ErrMalformed is returned when a key or signature cannot be parsed (bad
// base64, short buffer, unknown algorithm). The CLI maps it to ErrInvalidInput.
var ErrMalformed = errors.New("sign: malformed key or signature")

// PublicKey is a minisign-compatible Ed25519 public key.
type PublicKey struct {
	KeyID [8]byte
	Pub   ed25519.PublicKey
}

// SecretKey is a minisign-compatible Ed25519 secret key. The decrypted private
// key material is held inside a secret.Key so it never leaks into logs/errors.
type SecretKey struct {
	KeyID [8]byte
	key   secret.Key // decrypted ed25519 private key (64 bytes)
}

// String implements fmt.Stringer so a SecretKey never reveals key material.
func (s SecretKey) String() string { return s.key.String() }

// GoString implements fmt.GoStringer so %#v never reveals key material.
func (s SecretKey) GoString() string {
	return fmt.Sprintf("sign.SecretKey{KeyID:%x, key:%s}", s.KeyID, s.key.GoString())
}

// LogValue implements slog.LogValuer so structured logs never reveal material.
func (s SecretKey) LogValue() slog.Value { return s.key.LogValue() }

// KeyPair bundles a freshly generated public/secret key pair.
type KeyPair struct {
	PublicKey PublicKey
	SecretKey SecretKey
}

// Options configures signing-key generation and signing behavior.
type Options struct {
	// ScryptN, ScryptR, ScryptP are the scrypt cost parameters used to encrypt
	// the secret key. Defaults are the libsodium SENSITIVE profile.
	ScryptN int
	ScryptR int
	ScryptP int
	// TrustedComment is the trusted comment embedded in (and signed by) a
	// signature's global signature.
	TrustedComment string
	// UntrustedComment is the first-line comment written to key/signature files.
	UntrustedComment string
}

// Option is a functional option for key generation and signing.
type Option func(*Options)

// WithScryptParams overrides the scrypt cost parameters. Tests MUST use a low
// cost (e.g. WithScryptParams(1<<15, 8, 1)); the default SENSITIVE cost needs
// roughly 1 GiB of RAM and several seconds.
func WithScryptParams(n, r, p int) Option {
	return func(o *Options) { o.ScryptN, o.ScryptR, o.ScryptP = n, r, p }
}

// WithTrustedComment sets the trusted comment embedded in a signature.
func WithTrustedComment(c string) Option {
	return func(o *Options) { o.TrustedComment = c }
}

// WithUntrustedComment sets the first-line untrusted comment.
func WithUntrustedComment(c string) Option {
	return func(o *Options) { o.UntrustedComment = c }
}

func applyOptions(opts []Option) Options {
	o := Options{ScryptN: scryptN, ScryptR: scryptR, ScryptP: scryptP}
	for _, opt := range opts {
		opt(&o)
	}
	if o.ScryptN <= 1 {
		o.ScryptN = scryptN
	}
	if o.ScryptR <= 0 {
		o.ScryptR = scryptR
	}
	if o.ScryptP <= 0 {
		o.ScryptP = scryptP
	}
	return o
}

// GenerateKeyPair creates a new passphrase-protected Ed25519 key pair. The
// passphrase encrypts the secret key at rest. By default the libsodium
// SENSITIVE scrypt cost is used (~1 GiB RAM); pass WithScryptParams for tests.
func GenerateKeyPair(passphrase string, opts ...Option) (KeyPair, error) {
	_ = applyOptions(opts) // cost is consumed at MarshalText time; validate here.

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return KeyPair{}, fmt.Errorf("sign: generate ed25519 key: %w", err)
	}

	var keyID [8]byte
	if _, err := rand.Read(keyID[:]); err != nil {
		return KeyPair{}, fmt.Errorf("sign: generate key id: %w", err)
	}

	kp := KeyPair{
		PublicKey: PublicKey{KeyID: keyID, Pub: pub},
		SecretKey: SecretKey{KeyID: keyID, key: secret.New([]byte(priv))},
	}
	return kp, nil
}

// keynumChecksum computes Blake2b-256("Ed" || key_id || ed25519_secret).
func keynumChecksum(keyID [8]byte, priv []byte) [checksumSize]byte {
	h, _ := blake2b.New256(nil)
	_, _ = h.Write(sigAlgEd[:])
	_, _ = h.Write(keyID[:])
	_, _ = h.Write(priv)
	var out [checksumSize]byte
	copy(out[:], h.Sum(nil))
	return out
}

// MarshalText encodes the public key as a two-line minisign *.pub file.
func (p PublicKey) MarshalText() []byte {
	body := make([]byte, 0, 2+keyIDSize+ed25519PublicKeySize)
	body = append(body, sigAlgEd[:]...)
	body = append(body, p.KeyID[:]...)
	body = append(body, p.Pub...)

	comment := defaultPublicComment
	var sb strings.Builder
	sb.WriteString(comment)
	sb.WriteByte('\n')
	sb.WriteString(base64.StdEncoding.EncodeToString(body))
	sb.WriteByte('\n')
	return []byte(sb.String())
}

// ParsePublicKey decodes a minisign *.pub file into a PublicKey.
func ParsePublicKey(text []byte) (PublicKey, error) {
	lines := splitLines(text)
	if len(lines) < 2 {
		return PublicKey{}, fmt.Errorf("%w: public key needs 2 lines", ErrMalformed)
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(lines[1]))
	if err != nil {
		return PublicKey{}, fmt.Errorf("%w: public key base64: %v", ErrMalformed, err)
	}
	if len(raw) != 2+keyIDSize+ed25519PublicKeySize {
		return PublicKey{}, fmt.Errorf("%w: public key length %d", ErrMalformed, len(raw))
	}
	var alg [2]byte
	copy(alg[:], raw[:2])
	if alg != sigAlgEd {
		return PublicKey{}, fmt.Errorf("%w: unsupported public key algorithm %q", ErrMalformed, alg)
	}
	var p PublicKey
	copy(p.KeyID[:], raw[2:2+keyIDSize])
	p.Pub = make(ed25519.PublicKey, ed25519PublicKeySize)
	copy(p.Pub, raw[2+keyIDSize:])
	return p, nil
}

// MarshalText encodes the secret key as a two-line minisign *.key file,
// encrypting keynum_sk with a scrypt-derived XOR stream from the passphrase.
func (s SecretKey) MarshalText(passphrase string, opts ...Option) ([]byte, error) {
	o := applyOptions(opts)

	priv := s.key.Bytes()
	if len(priv) != ed25519SecretKeySize {
		return nil, fmt.Errorf("%w: secret key length %d", ErrMalformed, len(priv))
	}

	var salt [saltSize]byte
	if _, err := rand.Read(salt[:]); err != nil {
		return nil, fmt.Errorf("sign: generate salt: %w", err)
	}

	// Cleartext keynum_sk = key_id || ed25519_secret || checksum.
	cksum := keynumChecksum(s.KeyID, priv)
	keynum := make([]byte, 0, keynumSKSize)
	keynum = append(keynum, s.KeyID[:]...)
	keynum = append(keynum, priv...)
	keynum = append(keynum, cksum[:]...)

	// The header records libsodium opslimit/memlimit, NOT the raw (N, r, p);
	// ParseSecretKey recovers (N, r, p) from those limits via
	// scryptParamsFromLimits. To guarantee the decrypt stream matches the
	// encrypt stream for every cost (the limit<->param mapping is not a perfect
	// inverse at very low N — see scryptParamsFromLimits), derive the stream
	// from the SAME canonical params the parser will recover, not the caller's
	// raw request. For SENSITIVE and any realistic cost this is a no-op.
	opsLimit, memLimit := limitsFromScryptParams(o.ScryptN, o.ScryptR, o.ScryptP)
	encN, encR, encP := scryptParamsFromLimits(opsLimit, memLimit)

	stream, err := scrypt.Key([]byte(passphrase), salt[:], encN, encR, encP, keynumSKSize)
	if err != nil {
		return nil, fmt.Errorf("sign: scrypt: %w", err)
	}
	encrypted := make([]byte, keynumSKSize)
	subtle.XORBytes(encrypted, keynum, stream)

	// Header: sig_alg || kdf_alg || cksum_alg || salt || opslimit || memlimit || encrypted_keynum_sk
	body := make([]byte, 0, 2+2+2+saltSize+8+8+keynumSKSize)
	body = append(body, sigAlgEd[:]...)
	body = append(body, kdfScrypt[:]...)
	body = append(body, cksumBlake2b[:]...)
	body = append(body, salt[:]...)
	var ops, mem [8]byte
	binary.LittleEndian.PutUint64(ops[:], opsLimit)
	binary.LittleEndian.PutUint64(mem[:], memLimit)
	body = append(body, ops[:]...)
	body = append(body, mem[:]...)
	body = append(body, encrypted...)

	comment := o.UntrustedComment
	if comment == "" {
		comment = defaultUntrustedComment
	}
	var sb strings.Builder
	sb.WriteString(comment)
	sb.WriteByte('\n')
	sb.WriteString(base64.StdEncoding.EncodeToString(body))
	sb.WriteByte('\n')
	return []byte(sb.String()), nil
}

// ParseSecretKey decodes a minisign *.key file, decrypting keynum_sk with the
// passphrase and rejecting a checksum mismatch (wrong passphrase) with
// ErrVerification.
func ParseSecretKey(text []byte, passphrase string, opts ...Option) (SecretKey, error) {
	_ = opts // scrypt params come from the file header, not the caller.

	lines := splitLines(text)
	if len(lines) < 2 {
		return SecretKey{}, fmt.Errorf("%w: secret key needs 2 lines", ErrMalformed)
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(lines[1]))
	if err != nil {
		return SecretKey{}, fmt.Errorf("%w: secret key base64: %v", ErrMalformed, err)
	}
	const headerSize = 2 + 2 + 2 + saltSize + 8 + 8
	if len(raw) != headerSize+keynumSKSize {
		return SecretKey{}, fmt.Errorf("%w: secret key length %d", ErrMalformed, len(raw))
	}

	off := 0
	var sigAlg, kdfAlg, cksumAlg [2]byte
	copy(sigAlg[:], raw[off:off+2])
	off += 2
	copy(kdfAlg[:], raw[off:off+2])
	off += 2
	copy(cksumAlg[:], raw[off:off+2])
	off += 2
	if sigAlg != sigAlgEd || cksumAlg != cksumBlake2b {
		return SecretKey{}, fmt.Errorf("%w: unsupported secret key algorithms", ErrMalformed)
	}

	var salt [saltSize]byte
	copy(salt[:], raw[off:off+saltSize])
	off += saltSize
	opsLimit := binary.LittleEndian.Uint64(raw[off : off+8])
	off += 8
	memLimit := binary.LittleEndian.Uint64(raw[off : off+8])
	off += 8
	encrypted := raw[off:]

	var n, r, p int
	switch kdfAlg {
	case kdfScrypt:
		// Derive the scrypt cost from the libsodium opslimit/memlimit recorded
		// in the header (matches minisign), so any low-cost or SENSITIVE key
		// round-trips without the caller restating the cost.
		n, r, p = scryptParamsFromLimits(opsLimit, memLimit)
	case kdfNone:
		// Unencrypted key: stream is all-zero, so keynum == encrypted.
		n, r, p = 0, 0, 0
	default:
		return SecretKey{}, fmt.Errorf("%w: unsupported secret key KDF", ErrMalformed)
	}

	keynum := make([]byte, keynumSKSize)
	if kdfAlg == kdfScrypt {
		stream, err := scrypt.Key([]byte(passphrase), salt[:], n, r, p, keynumSKSize)
		if err != nil {
			return SecretKey{}, fmt.Errorf("sign: scrypt: %w", err)
		}
		subtle.XORBytes(keynum, encrypted, stream)
	} else {
		copy(keynum, encrypted)
	}

	var keyID [8]byte
	copy(keyID[:], keynum[:keyIDSize])
	priv := make([]byte, ed25519SecretKeySize)
	copy(priv, keynum[keyIDSize:keyIDSize+ed25519SecretKeySize])
	var gotCksum [checksumSize]byte
	copy(gotCksum[:], keynum[keyIDSize+ed25519SecretKeySize:])

	wantCksum := keynumChecksum(keyID, priv)
	if subtle.ConstantTimeCompare(gotCksum[:], wantCksum[:]) != 1 {
		return SecretKey{}, fmt.Errorf("%w: secret key checksum mismatch (wrong passphrase?)", ErrVerification)
	}

	return SecretKey{KeyID: keyID, key: secret.New(priv)}, nil
}

// Sign produces a detached minisign signature over data using the prehashed
// Ed25519 scheme ("ED"): the signature covers Blake2b-512(data). The returned
// bytes are the full four-line *.minisig file: an untrusted comment, the
// base64 signature payload (sig_alg || key_id || signature), a trusted-comment
// line, and the base64 global signature over signature || trusted_comment.
//
// WithTrustedComment sets the trusted comment (authenticated by the global
// signature); WithUntrustedComment sets the first-line comment (not signed).
func Sign(data []byte, sk SecretKey, opts ...Option) ([]byte, error) {
	o := applyOptions(opts)

	priv := sk.key.Bytes()
	if len(priv) != ed25519SecretKeySize {
		return nil, fmt.Errorf("%w: secret key length %d", ErrMalformed, len(priv))
	}
	edPriv := ed25519.PrivateKey(priv)

	// Prehash: sign Blake2b-512(data), not the raw data ("ED" algorithm).
	h := blake2b.Sum512(data)
	signature := ed25519.Sign(edPriv, h[:])

	// Line-2 payload: sig_alg || key_id || signature.
	body := make([]byte, 0, 2+keyIDSize+ed25519SignatureSize)
	body = append(body, sigAlgPrehashed[:]...)
	body = append(body, sk.KeyID[:]...)
	body = append(body, signature...)

	trustedComment := o.TrustedComment

	// Global signature: ed25519.Sign(sk, signature || trusted_comment_bytes).
	globalMsg := make([]byte, 0, ed25519SignatureSize+len(trustedComment))
	globalMsg = append(globalMsg, signature...)
	globalMsg = append(globalMsg, []byte(trustedComment)...)
	globalSig := ed25519.Sign(edPriv, globalMsg)

	untrusted := o.UntrustedComment
	if untrusted == "" {
		untrusted = defaultUntrustedComment
	}

	var sb strings.Builder
	sb.WriteString(untrusted)
	sb.WriteByte('\n')
	sb.WriteString(base64.StdEncoding.EncodeToString(body))
	sb.WriteByte('\n')
	sb.WriteString(trustedCommentPrefix)
	sb.WriteString(trustedComment)
	sb.WriteByte('\n')
	sb.WriteString(base64.StdEncoding.EncodeToString(globalSig))
	sb.WriteByte('\n')
	return []byte(sb.String()), nil
}

// parsedSignature is the decoded content of a minisign *.minisig file.
type parsedSignature struct {
	sigAlg         [2]byte
	keyID          [8]byte
	signature      []byte // 64-byte detached signature over (prehashed) data
	trustedComment string // trusted-comment text, sans prefix and trailing newline
	globalSig      []byte // 64-byte signature over signature || trusted_comment
}

// parseSignature decodes a four-line minisign *.minisig file. Every malformed
// condition (too few lines, bad base64, short buffer, missing trusted-comment
// prefix) returns ErrVerification so the caller fails closed.
func parseSignature(sig []byte) (parsedSignature, error) {
	var ps parsedSignature
	lines := splitLines(sig)
	if len(lines) < 4 {
		return ps, fmt.Errorf("%w: signature needs 4 lines, got %d", ErrVerification, len(lines))
	}

	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(lines[1]))
	if err != nil {
		return ps, fmt.Errorf("%w: signature base64: %v", ErrVerification, err)
	}
	if len(raw) != 2+keyIDSize+ed25519SignatureSize {
		return ps, fmt.Errorf("%w: signature payload length %d", ErrVerification, len(raw))
	}
	copy(ps.sigAlg[:], raw[:2])
	copy(ps.keyID[:], raw[2:2+keyIDSize])
	ps.signature = make([]byte, ed25519SignatureSize)
	copy(ps.signature, raw[2+keyIDSize:])

	if !strings.HasPrefix(lines[2], trustedCommentPrefix) {
		return ps, fmt.Errorf("%w: missing trusted comment prefix", ErrVerification)
	}
	ps.trustedComment = strings.TrimPrefix(lines[2], trustedCommentPrefix)

	gsig, err := base64.StdEncoding.DecodeString(strings.TrimSpace(lines[3]))
	if err != nil {
		return ps, fmt.Errorf("%w: global signature base64: %v", ErrVerification, err)
	}
	if len(gsig) != ed25519SignatureSize {
		return ps, fmt.Errorf("%w: global signature length %d", ErrVerification, len(gsig))
	}
	ps.globalSig = gsig
	return ps, nil
}

// Verify checks a detached minisign signature against data and a public key. It
// is fail-closed: it returns nil ONLY when every check passes, and an error
// wrapping ErrVerification on ANY failure (parse error, key_id mismatch, bad
// base64, short buffer, unsupported algorithm, failed data signature, or failed
// global signature). The CLI maps ErrVerification to cmderr.ErrConflict.
//
// Verification sequence (Pitfall 2): (1) parse the signature; (2) reject if the
// signature key_id differs from the public key's key_id BEFORE any crypto; (3)
// dispatch on sig_alg — "ED" verifies over Blake2b-512(data) (prehashed), "Ed"
// verifies over raw data (legacy, read-only), any other algorithm is rejected;
// (4) ALSO verify the global signature over signature || trusted_comment.
func Verify(data []byte, sig []byte, pub PublicKey) error {
	ps, err := parseSignature(sig)
	if err != nil {
		return err
	}

	// (2) key_id must match before any crypto.
	if subtle.ConstantTimeCompare(ps.keyID[:], pub.KeyID[:]) != 1 {
		return fmt.Errorf("%w: signature key id %x does not match public key %x", ErrVerification, ps.keyID, pub.KeyID)
	}

	if len(pub.Pub) != ed25519PublicKeySize {
		return fmt.Errorf("%w: public key length %d", ErrVerification, len(pub.Pub))
	}

	// (3) dispatch on signature algorithm.
	var dataMsg []byte
	switch ps.sigAlg {
	case sigAlgPrehashed:
		h := blake2b.Sum512(data)
		dataMsg = h[:]
	case sigAlgEd:
		dataMsg = data
	default:
		return fmt.Errorf("%w: unsupported signature algorithm %q", ErrVerification, ps.sigAlg)
	}
	if !ed25519.Verify(pub.Pub, dataMsg, ps.signature) {
		return fmt.Errorf("%w: data signature mismatch", ErrVerification)
	}

	// (4) verify the global signature over signature || trusted_comment.
	globalMsg := make([]byte, 0, len(ps.signature)+len(ps.trustedComment))
	globalMsg = append(globalMsg, ps.signature...)
	globalMsg = append(globalMsg, []byte(ps.trustedComment)...)
	if !ed25519.Verify(pub.Pub, globalMsg, ps.globalSig) {
		return fmt.Errorf("%w: global (trusted comment) signature mismatch", ErrVerification)
	}

	return nil
}

// splitLines splits text on newlines, trimming a trailing CR for CRLF files,
// and dropping a single trailing empty line.
func splitLines(text []byte) []string {
	raw := strings.Split(string(text), "\n")
	out := make([]string, 0, len(raw))
	for _, line := range raw {
		out = append(out, strings.TrimRight(line, "\r"))
	}
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return out
}
