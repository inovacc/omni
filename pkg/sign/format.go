package sign

// Minisign on-disk format constants. All algorithm identifiers are two ASCII
// bytes. See ADR-0005 and the byte-exact format section of the Phase 04 plan.
var (
	// sigAlgEd is the legacy raw Ed25519 algorithm ("Ed"): ed25519.Sign(sk, data).
	// Accepted on read for backwards compatibility; never emitted by default.
	sigAlgEd = [2]byte{'E', 'd'}
	// sigAlgPrehashed is the prehashed Ed25519 algorithm ("ED"):
	// ed25519.Sign(sk, Blake2b-512(data)). Default for new signatures.
	sigAlgPrehashed = [2]byte{'E', 'D'}
	// kdfScrypt identifies the scrypt KDF ("Sc") used to encrypt the secret key.
	kdfScrypt = [2]byte{'S', 'c'}
	// kdfNone is written when the secret key is not passphrase-protected (-W).
	kdfNone = [2]byte{0x00, 0x00}
	// cksumBlake2b identifies the Blake2b checksum algorithm ("B2").
	cksumBlake2b = [2]byte{'B', '2'}
)

const (
	// scryptN, scryptR, scryptP are the scrypt cost parameters corresponding to
	// the libsodium SENSITIVE profile (opslimit=33554432, memlimit=1 GiB).
	scryptN = 1 << 20
	scryptR = 8
	scryptP = 1

	// opsLimitSensitive and memLimitSensitive are the libsodium SENSITIVE limits
	// recorded in the secret-key header (8-byte little-endian unsigned).
	opsLimitSensitive uint64 = 33554432
	memLimitSensitive uint64 = 1 << 30

	// keyIDSize is the length of a minisign key ID in bytes.
	keyIDSize = 8
	// ed25519PublicKeySize is the length of an Ed25519 public key in bytes.
	ed25519PublicKeySize = 32
	// ed25519SecretKeySize is the length of an Ed25519 private key in bytes.
	ed25519SecretKeySize = 64
	// ed25519SignatureSize is the length of an Ed25519 signature in bytes.
	ed25519SignatureSize = 64
	// checksumSize is the length of the Blake2b-256 secret-key checksum.
	checksumSize = 32
	// saltSize is the length of the scrypt salt stored in the secret key.
	saltSize = 32

	// keynumSKSize is the cleartext keynum_sk: key_id || ed25519_secret || checksum.
	keynumSKSize = keyIDSize + ed25519SecretKeySize + checksumSize

	// defaultUntrustedComment is the comment written to public/secret key files.
	defaultUntrustedComment = "untrusted comment: signature from omni secret key"
	// defaultPublicComment is the comment written to public key files.
	defaultPublicComment = "untrusted comment: omni public key"
	// untrustedCommentPrefix is the prefix on the first line of all key/sig files.
	untrustedCommentPrefix = "untrusted comment: "
	// trustedCommentPrefix is the prefix on the trusted-comment line of a signature.
	trustedCommentPrefix = "trusted comment: "
)

// scryptParamsFromLimits derives scrypt (N, r, p) from the libsodium
// opslimit/memlimit pair stored in a minisign secret-key header. It mirrors
// libsodium's crypto_pwhash_scryptsalsa208sha256_ll selection so a header
// written by minisign (or by limitsFromScryptParams below) round-trips to the
// same cost parameters. r is fixed at 8, matching minisign.
func scryptParamsFromLimits(opsLimit, memLimit uint64) (n, r, p int) {
	r = 8
	if opsLimit < 32768 {
		opsLimit = 32768
	}
	var logN uint
	if opsLimit < memLimit/32 {
		p = 1
		maxN := opsLimit / (uint64(r) * 4)
		for logN = 1; logN < 63; logN++ {
			if uint64(1)<<logN > maxN/2 {
				break
			}
		}
	} else {
		maxN := memLimit / (uint64(r) * 128)
		for logN = 1; logN < 63; logN++ {
			if uint64(1)<<logN > maxN/2 {
				break
			}
		}
		pl := (opsLimit / 4) / (uint64(r) * (uint64(1) << logN))
		switch {
		case pl > 0x3fffffff:
			p = 0x3fffffff
		case pl < 1:
			p = 1
		default:
			p = int(pl)
		}
	}
	n = 1 << logN
	return n, r, p
}

// limitsFromScryptParams produces an opslimit/memlimit pair that
// scryptParamsFromLimits maps back to the same (n, r, p). For the libsodium
// SENSITIVE profile (N=2^20, r=8, p=1) it yields exactly opsLimitSensitive /
// memLimitSensitive, keeping headers byte-compatible with minisign.
func limitsFromScryptParams(n, r, p int) (opsLimit, memLimit uint64) {
	rn := uint64(r) * uint64(n)
	// memlimit must satisfy n <= memlimit/(r*128)/... ; use libsodium's relation
	// memlimit = N * r * 128 (the minimum that selects this N via the mem branch
	// equivalent), and opslimit = 4 * p * r * N so the ops branch is consistent.
	memLimit = rn * 128
	opsLimit = uint64(p) * rn * 4
	if opsLimit < 32768 {
		opsLimit = 32768
	}
	return opsLimit, memLimit
}
