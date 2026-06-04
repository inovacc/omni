// Package attest is a pure-Go reimplementation of the in-toto Statement v1
// envelope, the SLSA Provenance v1 predicate, and the DSSE Pre-Authentication
// Encoding (PAE), built on encoding/json, encoding/base64, and crypto/sha256.
//
// It deliberately avoids the official in-toto/sigstore SDKs (which pull the
// Rekor/go-tuf/protobuf-specs tree into go.mod via MVS — see ADR-0007); the wire
// formats are implemented byte-exactly against their published specs and pinned
// to reference test vectors. Signatures are produced and verified with the
// in-repo pkg/sign Ed25519 primitive, so the same key infrastructure
// (omni sign keygen) underpins the whole supply-chain toolchain.
//
// The honest SLSA Build level is conveyed solely by an ADR-0009-pinned
// builder.id allowlist; no numeric slsaLevel field is emitted and the generator
// refuses any builder.id outside the allowlist.
//
// Experimental: this package's API is not yet stable and may change without
// notice before omni v1.0.
package attest
