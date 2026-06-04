package attest

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
)

func TestSignEnvelope(t *testing.T) {
	st := NewStatement(
		[]ResourceDescriptor{SubjectFromBytes("app", []byte("x"))},
		Provenance{
			BuildDefinition: BuildDefinition{BuildType: BuildTypeLocal, ExternalParameters: map[string]any{}},
			RunDetails:      RunDetails{Builder: Builder{ID: BuilderIDLocal}},
		},
	)
	var gotPAE []byte
	signer := func(pae []byte) ([]byte, string, error) {
		gotPAE = append([]byte(nil), pae...)
		return []byte("SIGBYTES"), "deadbeef", nil
	}
	env, err := SignStatement(st, signer)
	if err != nil {
		t.Fatalf("SignStatement: %v", err)
	}
	if env.PayloadType != PayloadTypeInToto {
		t.Fatalf("payloadType = %q", env.PayloadType)
	}
	body, err := base64.StdEncoding.DecodeString(env.Payload)
	if err != nil {
		t.Fatalf("payload not std-base64: %v", err)
	}
	var rt Statement
	if err := json.Unmarshal(body, &rt); err != nil {
		t.Fatalf("payload not statement JSON: %v", err)
	}
	wantPAE := PAE(PayloadTypeInToto, body)
	if string(gotPAE) != string(wantPAE) {
		t.Fatalf("signer received wrong PAE:\n got=%q\nwant=%q", gotPAE, wantPAE)
	}
	if len(env.Signatures) != 1 || env.Signatures[0].KeyID != "deadbeef" {
		t.Fatalf("signatures = %+v", env.Signatures)
	}
	if env.Signatures[0].Sig != base64.StdEncoding.EncodeToString([]byte("SIGBYTES")) {
		t.Fatalf("sig not std-base64 of signer output: %q", env.Signatures[0].Sig)
	}
}

func TestVerifyEnvelopeFailsClosed(t *testing.T) {
	st := NewStatement(
		[]ResourceDescriptor{SubjectFromBytes("app", []byte("x"))},
		Provenance{
			BuildDefinition: BuildDefinition{BuildType: BuildTypeLocal, ExternalParameters: map[string]any{}},
			RunDetails:      RunDetails{Builder: Builder{ID: BuilderIDLocal}},
		},
	)
	// A verifier that accepts ONLY this exact byte string.
	accept := func(pae, sig []byte, keyid string) error {
		if string(sig) == "GOOD" {
			return nil
		}
		return ErrVerification
	}
	signer := func(pae []byte) ([]byte, string, error) { return []byte("GOOD"), "kid", nil }
	good, err := SignStatement(st, signer)
	if err != nil {
		t.Fatalf("SignStatement: %v", err)
	}

	// Positive: a well-formed, correctly-signed envelope verifies and returns
	// the parsed statement.
	gotSt, err := VerifyEnvelope(good, accept)
	if err != nil {
		t.Fatalf("VerifyEnvelope(valid) = %v, want nil", err)
	}
	if gotSt.Type != StatementType || gotSt.PredicateType != PredicateTypeSLSAProvenance {
		t.Fatalf("VerifyEnvelope(valid) returned wrong statement: %+v", gotSt)
	}

	bad := func(mut func(*Envelope)) Envelope { e := good; mut(&e); return e }
	cases := map[string]Envelope{
		"wrong payloadType": bad(func(e *Envelope) { e.PayloadType = "application/json" }),
		"no signatures":     bad(func(e *Envelope) { e.Signatures = nil }),
		"bad payload b64":   bad(func(e *Envelope) { e.Payload = "!!!not base64!!!" }),
		"bad sig b64":       bad(func(e *Envelope) { e.Signatures = []EnvelopeSignature{{Sig: "!!!"}} }),
		"tampered payload":  bad(func(e *Envelope) { e.Payload = base64.StdEncoding.EncodeToString([]byte(`{"_type":"x"}`)) }),
		"sig rejected": bad(func(e *Envelope) {
			e.Signatures = []EnvelopeSignature{{Sig: base64.StdEncoding.EncodeToString([]byte("BAD"))}}
		}),
		"malformed statement": bad(func(e *Envelope) {
			e.Payload = base64.StdEncoding.EncodeToString([]byte("not json at all"))
		}),
	}
	for name, env := range cases {
		gotSt, err := VerifyEnvelope(env, accept)
		if err == nil {
			t.Errorf("%s: VerifyEnvelope = nil, want error (fail-closed)", name)
		}
		if gotSt.Type != "" || gotSt.Subject != nil || gotSt.PredicateType != "" || gotSt.Predicate != nil {
			t.Errorf("%s: VerifyEnvelope returned non-zero statement on failure: %+v", name, gotSt)
		}
		if err != nil && !errors.Is(err, ErrVerification) {
			t.Errorf("%s: error %v is not ErrVerification", name, err)
		}
	}

	// "tampered payload" must fail at the signature step (the verifier never
	// saw GOOD over this PAE), not silently re-parse a different payload.
	if _, err := VerifyEnvelope(cases["tampered payload"], accept); !errors.Is(err, ErrVerification) {
		t.Fatalf("tampered payload: err = %v, want ErrVerification", err)
	}
}
