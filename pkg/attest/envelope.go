package attest

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

// ErrVerification is returned by VerifyEnvelope for any fail-closed condition:
// a non-in-toto payloadType, no signatures, undecodable payload/signature
// base64, a signature that no verifier accepts, or a payload that does not parse
// into a valid in-toto Statement. The CLI maps it to cmderr.ErrConflict. Every
// failure mode returns this sentinel and a zero Statement; VerifyEnvelope never
// silently passes.
var ErrVerification = errors.New("attest: envelope verification failed")

// Signer signs the DSSE PAE bytes and returns (signatureBytes, keyidHex). The
// CLI wires this to pkg/sign (Ed25519); tests may inject a stub. The signature
// bytes are base64-std encoded into the envelope's sig field, and keyidHex is
// placed verbatim into keyid.
type Signer func(pae []byte) (sig []byte, keyidHex string, err error)

// MarshalStatement serializes a Statement to its canonical JSON SERIALIZED_BODY.
// Field order follows struct declaration order (encoding/json is deterministic),
// which is what gets signed via PAE and base64-encoded into the envelope payload.
func MarshalStatement(st Statement) ([]byte, error) {
	b, err := json.Marshal(st)
	if err != nil {
		return nil, fmt.Errorf("attest: marshal statement: %w", err)
	}
	return b, nil
}

// SignStatement marshals st, computes the DSSE PAE over the in-toto payload
// type, invokes signer, and returns a DSSE JSON envelope. The SAME bytes that
// are signed are base64-encoded into Payload (DSSE binding requirement): the
// signature covers PAE(PayloadTypeInToto, body), while Payload carries
// base64-std(body) so a verifier can re-derive the identical PAE input.
func SignStatement(st Statement, signer Signer) (Envelope, error) {
	body, err := MarshalStatement(st)
	if err != nil {
		return Envelope{}, err
	}
	pae := PAE(PayloadTypeInToto, body)
	sig, keyid, err := signer(pae)
	if err != nil {
		return Envelope{}, fmt.Errorf("attest: sign: %w", err)
	}
	return Envelope{
		Payload:     base64.StdEncoding.EncodeToString(body),
		PayloadType: PayloadTypeInToto,
		Signatures:  []EnvelopeSignature{{KeyID: keyid, Sig: base64.StdEncoding.EncodeToString(sig)}},
	}, nil
}

// Verifier checks that sig is a valid signature over pae for the key hinted by
// keyid. It returns nil ONLY on a cryptographically valid signature; any other
// outcome (including a wrong key or a malformed signature) is a non-nil error.
// The CLI wires this to pkg/sign.Verify; tests may inject a stub.
type Verifier func(pae, sig []byte, keyid string) error

// VerifyEnvelope verifies a DSSE envelope fail-closed and returns the parsed
// Statement on success. In order it: (1) rejects a non-in-toto payloadType,
// (2) requires at least one signature, (3) base64-decodes the payload,
// (4) re-derives PAE(payloadType, payload-bytes) and accepts only if some
// signature base64-decodes AND verify reports it valid over that exact PAE,
// (5) parses the decoded payload into a Statement and checks the fixed _type.
// ANY failure returns ErrVerification and a zero Statement — never a silent
// pass. Per the DSSE binding rule, the returned Statement is parsed from the
// SAME bytes that were verified; the envelope is never re-parsed afterward.
func VerifyEnvelope(env Envelope, verify Verifier) (Statement, error) {
	if env.PayloadType != PayloadTypeInToto {
		return Statement{}, fmt.Errorf("%w: unexpected payloadType %q", ErrVerification, env.PayloadType)
	}
	if len(env.Signatures) == 0 {
		return Statement{}, fmt.Errorf("%w: no signatures", ErrVerification)
	}
	body, err := base64.StdEncoding.DecodeString(env.Payload)
	if err != nil {
		return Statement{}, fmt.Errorf("%w: payload base64: %v", ErrVerification, err)
	}
	pae := PAE(env.PayloadType, body)
	ok := false
	for _, s := range env.Signatures {
		sig, err := base64.StdEncoding.DecodeString(s.Sig)
		if err != nil {
			continue
		}
		if verify(pae, sig, s.KeyID) == nil {
			ok = true
			break
		}
	}
	if !ok {
		return Statement{}, fmt.Errorf("%w: no valid signature", ErrVerification)
	}
	var st Statement
	if err := json.Unmarshal(body, &st); err != nil {
		return Statement{}, fmt.Errorf("%w: malformed statement: %v", ErrVerification, err)
	}
	if st.Type != StatementType {
		return Statement{}, fmt.Errorf("%w: unexpected _type %q", ErrVerification, st.Type)
	}
	return st, nil
}
