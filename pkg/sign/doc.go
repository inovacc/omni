// Package sign implements a pure-Go, fail-closed, minisign-compatible Ed25519
// signing primitive: passphrase-protected key generation, prehashed detached
// signatures, and strict verification.
//
// The on-disk format is byte-compatible with the minisign tool: a two-line
// public key (*.pub), a two-line scrypt-encrypted secret key (*.key), and a
// four-line detached signature (*.minisig) carrying a trusted comment and a
// global signature over the signature plus that comment.
//
// New signatures use the prehashed "ED" algorithm (Ed25519 over a Blake2b-512
// digest of the data); the legacy raw "Ed" algorithm is accepted on read but
// never emitted. Verification is fail-closed: any parse error, key-ID mismatch,
// unknown algorithm, or failed data/global signature returns an error.
//
// Secret key material is held inside pkg/secret.Key so it never leaks into
// logs, errors, %v/%#v, or panics.
//
// This is a stable v1.0 API.
package sign
