package attest_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/inovacc/omni/internal/cli/attest"
	"github.com/inovacc/omni/internal/cli/cmderr"
)

// fixtureKeyPassphrase is the passphrase the committed sign golden fixtures were
// generated with (see testing/golden/fixtures/sign/gen_fixtures.go).
const fixtureKeyPassphrase = "golden-fixture-passphrase"

const signFixtures = "../../../testing/golden/fixtures/sign"

func writeTemp(t *testing.T, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRunAttestThenVerifyRoundTrip(t *testing.T) {
	t.Setenv("OMNI_SIGN_PASSPHRASE", fixtureKeyPassphrase)
	artifact := writeTemp(t, "app.tar.gz", []byte("artifact-bytes"))
	out := writeTemp(t, "app.intoto.jsonl", nil)

	var w bytes.Buffer
	gen := attest.GenOptions{
		KeyPath:       filepath.Join(signFixtures, "test.key"),
		ArtifactPath:  artifact,
		PredicateType: "slsa-provenance",
		BuilderID:     "", // empty -> defaults to local builder.id
		OutPath:       out,
	}
	if err := attest.RunAttest(&w, gen); err != nil {
		t.Fatalf("RunAttest: %v", err)
	}

	ver := attest.VerifyOptions{
		KeyPath:      filepath.Join(signFixtures, "test.pub"),
		EnvelopePath: out,
		ArtifactPath: artifact,
	}
	if err := attest.RunVerify(&w, ver); err != nil {
		t.Fatalf("RunVerify(valid) = %v, want nil", err)
	}
}

func TestRunAttestToStdout(t *testing.T) {
	t.Setenv("OMNI_SIGN_PASSPHRASE", fixtureKeyPassphrase)
	artifact := writeTemp(t, "app", []byte("x"))
	var w bytes.Buffer
	err := attest.RunAttest(&w, attest.GenOptions{
		KeyPath:       filepath.Join(signFixtures, "test.key"),
		ArtifactPath:  artifact,
		PredicateType: "slsa-provenance",
		OutPath:       "", // stdout
	})
	if err != nil {
		t.Fatalf("RunAttest(stdout): %v", err)
	}
	if !bytes.Contains(w.Bytes(), []byte(`"payloadType"`)) {
		t.Fatalf("stdout envelope missing payloadType: %q", w.String())
	}
}

func TestRunAttestRejectsOverclaim(t *testing.T) {
	t.Setenv("OMNI_SIGN_PASSPHRASE", fixtureKeyPassphrase)
	artifact := writeTemp(t, "app", []byte("x"))
	var w bytes.Buffer
	err := attest.RunAttest(&w, attest.GenOptions{
		KeyPath:       filepath.Join(signFixtures, "test.key"),
		ArtifactPath:  artifact,
		PredicateType: "slsa-provenance",
		BuilderID:     "https://slsa.dev/fake-l3-platform",
		OutPath:       writeTemp(t, "x.jsonl", nil),
	})
	if !cmderr.IsInvalidInput(err) {
		t.Fatalf("overclaim builder.id: err = %v, want cmderr.ErrInvalidInput", err)
	}
}

func TestRunAttestUnsupportedPredicate(t *testing.T) {
	t.Setenv("OMNI_SIGN_PASSPHRASE", fixtureKeyPassphrase)
	artifact := writeTemp(t, "app", []byte("x"))
	var w bytes.Buffer
	err := attest.RunAttest(&w, attest.GenOptions{
		KeyPath:       filepath.Join(signFixtures, "test.key"),
		ArtifactPath:  artifact,
		PredicateType: "spdx",
		OutPath:       writeTemp(t, "x.jsonl", nil),
	})
	if !cmderr.IsUnsupported(err) {
		t.Fatalf("unsupported predicate: err = %v, want cmderr.ErrUnsupported", err)
	}
}

func TestRunAttestMissingKey(t *testing.T) {
	t.Setenv("OMNI_SIGN_PASSPHRASE", fixtureKeyPassphrase)
	artifact := writeTemp(t, "app", []byte("x"))
	var w bytes.Buffer
	err := attest.RunAttest(&w, attest.GenOptions{
		KeyPath:       filepath.Join(signFixtures, "does-not-exist.key"),
		ArtifactPath:  artifact,
		PredicateType: "slsa-provenance",
		OutPath:       writeTemp(t, "x.jsonl", nil),
	})
	if !cmderr.IsNotFound(err) {
		t.Fatalf("missing key: err = %v, want cmderr.ErrNotFound", err)
	}
}

func TestRunVerifyTamperedFailsConflict(t *testing.T) {
	t.Setenv("OMNI_SIGN_PASSPHRASE", fixtureKeyPassphrase)
	artifact := writeTemp(t, "app", []byte("x"))
	out := writeTemp(t, "x.jsonl", nil)
	var w bytes.Buffer
	if err := attest.RunAttest(&w, attest.GenOptions{
		KeyPath:       filepath.Join(signFixtures, "test.key"),
		ArtifactPath:  artifact,
		PredicateType: "slsa-provenance",
		OutPath:       out,
	}); err != nil {
		t.Fatalf("RunAttest: %v", err)
	}
	env, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	// Flip a byte inside the base64 payload string to break the signature binding.
	tampered := bytes.Replace(env, []byte(`"payload": "`), []byte(`"payload": "A`), 1)
	bad := writeTemp(t, "bad.jsonl", tampered)
	err = attest.RunVerify(&w, attest.VerifyOptions{
		KeyPath:      filepath.Join(signFixtures, "test.pub"),
		EnvelopePath: bad,
		ArtifactPath: artifact,
	})
	if !cmderr.IsConflict(err) && !cmderr.IsInvalidInput(err) {
		t.Fatalf("tampered envelope: err = %v, want Conflict or InvalidInput", err)
	}
}

func TestRunVerifyWrongKeyFailsConflict(t *testing.T) {
	t.Setenv("OMNI_SIGN_PASSPHRASE", fixtureKeyPassphrase)
	artifact := writeTemp(t, "app", []byte("x"))
	out := writeTemp(t, "x.jsonl", nil)
	var w bytes.Buffer
	if err := attest.RunAttest(&w, attest.GenOptions{
		KeyPath:       filepath.Join(signFixtures, "test.key"),
		ArtifactPath:  artifact,
		PredicateType: "slsa-provenance",
		OutPath:       out,
	}); err != nil {
		t.Fatalf("RunAttest: %v", err)
	}
	err := attest.RunVerify(&w, attest.VerifyOptions{
		KeyPath:      filepath.Join(signFixtures, "wrong.pub"),
		EnvelopePath: out,
		ArtifactPath: artifact,
	})
	if !cmderr.IsConflict(err) {
		t.Fatalf("wrong key: err = %v, want cmderr.ErrConflict", err)
	}
}

func TestRunVerifyDigestMismatchFailsConflict(t *testing.T) {
	t.Setenv("OMNI_SIGN_PASSPHRASE", fixtureKeyPassphrase)
	artifact := writeTemp(t, "app", []byte("x"))
	other := writeTemp(t, "other", []byte("different-artifact-bytes"))
	out := writeTemp(t, "x.jsonl", nil)
	var w bytes.Buffer
	if err := attest.RunAttest(&w, attest.GenOptions{
		KeyPath:       filepath.Join(signFixtures, "test.key"),
		ArtifactPath:  artifact,
		PredicateType: "slsa-provenance",
		OutPath:       out,
	}); err != nil {
		t.Fatalf("RunAttest: %v", err)
	}
	err := attest.RunVerify(&w, attest.VerifyOptions{
		KeyPath:      filepath.Join(signFixtures, "test.pub"),
		EnvelopePath: out,
		ArtifactPath: other, // sha256 will not match the subject
	})
	if !cmderr.IsConflict(err) {
		t.Fatalf("digest mismatch: err = %v, want cmderr.ErrConflict", err)
	}
}

func TestRunVerifyBadEnvelopeJSONFailsInvalidInput(t *testing.T) {
	t.Setenv("OMNI_SIGN_PASSPHRASE", fixtureKeyPassphrase)
	bad := writeTemp(t, "bad.jsonl", []byte("this is not json"))
	var w bytes.Buffer
	err := attest.RunVerify(&w, attest.VerifyOptions{
		KeyPath:      filepath.Join(signFixtures, "test.pub"),
		EnvelopePath: bad,
	})
	if !cmderr.IsInvalidInput(err) {
		t.Fatalf("bad envelope JSON: err = %v, want cmderr.ErrInvalidInput", err)
	}
}

func TestRunVerifyMissingEnvelopeFailsNotFound(t *testing.T) {
	var w bytes.Buffer
	err := attest.RunVerify(&w, attest.VerifyOptions{
		KeyPath:      filepath.Join(signFixtures, "test.pub"),
		EnvelopePath: filepath.Join(t.TempDir(), "nope.jsonl"),
	})
	if !cmderr.IsNotFound(err) {
		t.Fatalf("missing envelope: err = %v, want cmderr.ErrNotFound", err)
	}
}
