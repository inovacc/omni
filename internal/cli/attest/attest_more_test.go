package attest_test

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/inovacc/omni/internal/cli/attest"
	"github.com/inovacc/omni/internal/cli/cmderr"
)

func TestRunVerifyReaderRoundTrip(t *testing.T) {
	t.Setenv("OMNI_SIGN_PASSPHRASE", fixtureKeyPassphrase)
	artifact := writeTemp(t, "app.bin", []byte("artifact-bytes"))

	// Generate a signed envelope to stdout (a buffer).
	var env bytes.Buffer
	if err := attest.RunAttest(&env, attest.GenOptions{
		KeyPath:       filepath.Join(signFixtures, "test.key"),
		ArtifactPath:  artifact,
		PredicateType: "slsa-provenance",
	}); err != nil {
		t.Fatalf("RunAttest: %v", err)
	}

	// Verify it through the reader entry point.
	var out bytes.Buffer
	err := attest.RunVerifyReader(&out, bytes.NewReader(env.Bytes()), []string{filepath.Join(signFixtures, "test.pub")})
	if err != nil {
		t.Fatalf("RunVerifyReader(valid) = %v, want nil", err)
	}
	if !strings.Contains(out.String(), "OK:") {
		t.Fatalf("expected OK summary, got %q", out.String())
	}
}

func TestRunVerifyReaderMissingArg(t *testing.T) {
	err := attest.RunVerifyReader(&bytes.Buffer{}, strings.NewReader("{}"), nil)
	if !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for missing key arg, got %v", err)
	}
	err = attest.RunVerifyReader(&bytes.Buffer{}, strings.NewReader("{}"), []string{""})
	if !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for empty key arg, got %v", err)
	}
}

func TestRunVerifyReaderMissingKeyFile(t *testing.T) {
	err := attest.RunVerifyReader(&bytes.Buffer{}, strings.NewReader("{}"), []string{filepath.Join(t.TempDir(), "absent.pub")})
	if !errors.Is(err, cmderr.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for missing pubkey, got %v", err)
	}
}

func TestRunVerifyReaderBadEnvelope(t *testing.T) {
	t.Setenv("OMNI_SIGN_PASSPHRASE", fixtureKeyPassphrase)
	err := attest.RunVerifyReader(&bytes.Buffer{}, strings.NewReader("not json"), []string{filepath.Join(signFixtures, "test.pub")})
	if !errors.Is(err, cmderr.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for bad envelope JSON, got %v", err)
	}
}

// builderIDLocal / builderIDRelease / buildTypeLocal mirror the ADR-0009
// allowlist constants in pkg/attest. They are duplicated here (rather than
// imported) so this external test exercises buildProvenance via the public
// RunAttest entry point with a realistic predicate file.
const (
	builderIDLocal   = "https://github.com/inovacc/omni/attest/local@v1"
	builderIDFakeL3  = "https://slsa.dev/fake-l3-platform"
	buildTypeLocalID = "https://github.com/inovacc/omni/attest/local-buildtype/v1"
)

// writePredicate writes a minimal SLSA Provenance v1 predicate JSON with the
// given builder.id and returns its path.
func writePredicate(t *testing.T, builderID string) string {
	t.Helper()
	doc := `{
  "buildDefinition": {
    "buildType": "` + buildTypeLocalID + `",
    "externalParameters": {"artifact": "app"}
  },
  "runDetails": {
    "builder": {"id": "` + builderID + `"}
  }
}`
	return writeTemp(t, "predicate.json", []byte(doc))
}

// TestRunAttestWithPredicateFile exercises buildProvenance's --predicate branch
// with a valid (allowlisted) builder.id; the resulting envelope must verify.
func TestRunAttestWithPredicateFile(t *testing.T) {
	t.Setenv("OMNI_SIGN_PASSPHRASE", fixtureKeyPassphrase)
	artifact := writeTemp(t, "app.bin", []byte("xyz"))
	pred := writePredicate(t, builderIDLocal)
	out := writeTemp(t, "env.jsonl", nil)

	var w bytes.Buffer
	err := attest.RunAttest(&w, attest.GenOptions{
		KeyPath:       filepath.Join(signFixtures, "test.key"),
		ArtifactPath:  artifact,
		PredicateType: "slsa-provenance",
		PredicatePath: pred,
		OutPath:       out,
	})
	if err != nil {
		t.Fatalf("RunAttest(predicate file) = %v, want nil", err)
	}

	// The generated envelope must verify with the matching public key.
	if verr := attest.RunVerify(&w, attest.VerifyOptions{
		KeyPath:      filepath.Join(signFixtures, "test.pub"),
		EnvelopePath: out,
		ArtifactPath: artifact,
	}); verr != nil {
		t.Fatalf("RunVerify(predicate envelope) = %v, want nil", verr)
	}
}

// TestRunAttestPredicateFileOverclaim ensures a --predicate file whose
// builder.id is outside the ADR-0009 allowlist is refused as ErrInvalidInput.
func TestRunAttestPredicateFileOverclaim(t *testing.T) {
	t.Setenv("OMNI_SIGN_PASSPHRASE", fixtureKeyPassphrase)
	artifact := writeTemp(t, "app.bin", []byte("xyz"))
	pred := writePredicate(t, builderIDFakeL3)

	var w bytes.Buffer
	err := attest.RunAttest(&w, attest.GenOptions{
		KeyPath:       filepath.Join(signFixtures, "test.key"),
		ArtifactPath:  artifact,
		PredicateType: "slsa-provenance",
		PredicatePath: pred,
		OutPath:       writeTemp(t, "env.jsonl", nil),
	})
	if !cmderr.IsInvalidInput(err) {
		t.Fatalf("predicate overclaim: err = %v, want ErrInvalidInput", err)
	}
}

// TestRunAttestPredicateFileBadJSON covers the malformed-predicate branch.
func TestRunAttestPredicateFileBadJSON(t *testing.T) {
	t.Setenv("OMNI_SIGN_PASSPHRASE", fixtureKeyPassphrase)
	artifact := writeTemp(t, "app.bin", []byte("xyz"))
	pred := writeTemp(t, "predicate.json", []byte("{not valid json"))

	var w bytes.Buffer
	err := attest.RunAttest(&w, attest.GenOptions{
		KeyPath:       filepath.Join(signFixtures, "test.key"),
		ArtifactPath:  artifact,
		PredicateType: "slsa-provenance",
		PredicatePath: pred,
		OutPath:       writeTemp(t, "env.jsonl", nil),
	})
	if !cmderr.IsInvalidInput(err) {
		t.Fatalf("bad predicate JSON: err = %v, want ErrInvalidInput", err)
	}
}

// TestRunAttestPredicateFileMissing covers the readFile failure inside the
// --predicate branch of buildProvenance.
func TestRunAttestPredicateFileMissing(t *testing.T) {
	t.Setenv("OMNI_SIGN_PASSPHRASE", fixtureKeyPassphrase)
	artifact := writeTemp(t, "app.bin", []byte("xyz"))

	var w bytes.Buffer
	err := attest.RunAttest(&w, attest.GenOptions{
		KeyPath:       filepath.Join(signFixtures, "test.key"),
		ArtifactPath:  artifact,
		PredicateType: "slsa-provenance",
		PredicatePath: filepath.Join(t.TempDir(), "absent-predicate.json"),
		OutPath:       writeTemp(t, "env.jsonl", nil),
	})
	if !cmderr.IsNotFound(err) {
		t.Fatalf("missing predicate file: err = %v, want ErrNotFound", err)
	}
}

// TestRunAttestFromEnvGHA exercises buildProvenance's --from-env path on a
// simulated GitHub-Actions runner (GITHUB_ACTIONS=true). The release builder.id
// and GITHUB_* external parameters are used; the envelope must still verify.
func TestRunAttestFromEnvGHA(t *testing.T) {
	t.Setenv("OMNI_SIGN_PASSPHRASE", fixtureKeyPassphrase)
	t.Setenv("GITHUB_ACTIONS", "true")
	t.Setenv("GITHUB_WORKFLOW", "release")
	t.Setenv("GITHUB_REPOSITORY", "inovacc/omni")
	t.Setenv("GITHUB_REF", "refs/heads/main")
	t.Setenv("GITHUB_SHA", "deadbeef")
	t.Setenv("GITHUB_RUN_ID", "12345")

	artifact := writeTemp(t, "app.bin", []byte("xyz"))
	out := writeTemp(t, "env.jsonl", nil)

	var w bytes.Buffer
	err := attest.RunAttest(&w, attest.GenOptions{
		KeyPath:       filepath.Join(signFixtures, "test.key"),
		ArtifactPath:  artifact,
		PredicateType: "slsa-provenance",
		FromEnv:       true,
		OutPath:       out,
	})
	if err != nil {
		t.Fatalf("RunAttest(--from-env GHA) = %v, want nil", err)
	}
	if verr := attest.RunVerify(&w, attest.VerifyOptions{
		KeyPath:      filepath.Join(signFixtures, "test.pub"),
		EnvelopePath: out,
		ArtifactPath: artifact,
	}); verr != nil {
		t.Fatalf("RunVerify(--from-env envelope) = %v, want nil", verr)
	}
}
