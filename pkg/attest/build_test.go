package attest

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
)

func TestSubjectFromBytes(t *testing.T) {
	data := []byte("artifact bytes")
	rd := SubjectFromBytes("app.tar.gz", data)
	sum := sha256.Sum256(data)
	if rd.Name != "app.tar.gz" {
		t.Fatalf("name = %q", rd.Name)
	}
	if rd.Digest["sha256"] != hex.EncodeToString(sum[:]) {
		t.Fatalf("digest = %q", rd.Digest["sha256"])
	}
}

func TestNewStatementSLSA(t *testing.T) {
	prov := Provenance{
		BuildDefinition: BuildDefinition{
			BuildType:          BuildTypeLocal,
			ExternalParameters: map[string]any{"artifact": "app.tar.gz"},
		},
		RunDetails: RunDetails{Builder: Builder{ID: BuilderIDLocal}},
	}
	st := NewStatement([]ResourceDescriptor{SubjectFromBytes("app.tar.gz", []byte("x"))}, prov)
	if st.Type != StatementType || st.PredicateType != PredicateTypeSLSAProvenance {
		t.Fatalf("fixed type fields wrong: %+v", st)
	}
	if st.Predicate == nil {
		t.Fatal("predicate not set")
	}
	if len(st.Subject) != 1 {
		t.Fatalf("subject len = %d", len(st.Subject))
	}
}

func TestValidateBuilderIDRejectsOverclaim(t *testing.T) {
	if err := ValidateBuilderID("https://slsa.dev/some-l3-platform"); err == nil {
		t.Fatal("ValidateBuilderID must reject a builder.id not in the ADR-0009 allowlist")
	}
	if err := ValidateBuilderID(BuilderIDLocal); err != nil {
		t.Fatalf("ValidateBuilderID(local) = %v, want nil", err)
	}
	if err := ValidateBuilderID(BuilderIDRelease); err != nil {
		t.Fatalf("ValidateBuilderID(release) = %v, want nil", err)
	}
}

func TestValidateBuilderIDWrapsErrOverclaim(t *testing.T) {
	err := ValidateBuilderID("")
	if err == nil {
		t.Fatal("empty builder.id must be rejected")
	}
	if !errors.Is(err, ErrOverclaim) {
		t.Fatalf("err = %v, want errors.Is(err, ErrOverclaim)", err)
	}
}

func TestBuildProvenanceLocal(t *testing.T) {
	prov, err := BuildProvenancePredicate(BuilderIDLocal, BuildTypeLocal, map[string]any{"artifact": "app"}, nil)
	if err != nil {
		t.Fatalf("BuildProvenancePredicate(local) = %v, want nil", err)
	}
	if prov.RunDetails.Builder.ID != BuilderIDLocal {
		t.Fatalf("builder.id = %q", prov.RunDetails.Builder.ID)
	}
	if prov.BuildDefinition.BuildType != BuildTypeLocal {
		t.Fatalf("buildType = %q", prov.BuildDefinition.BuildType)
	}
	if prov.BuildDefinition.ExternalParameters["artifact"] != "app" {
		t.Fatalf("externalParameters = %+v", prov.BuildDefinition.ExternalParameters)
	}
}

func TestBuildProvenancePredicateRejectsOverclaim(t *testing.T) {
	_, err := BuildProvenancePredicate("https://slsa.dev/fake-l3", BuildTypeLocal, map[string]any{}, nil)
	if !errors.Is(err, ErrOverclaim) {
		t.Fatalf("err = %v, want errors.Is(err, ErrOverclaim)", err)
	}
}

func TestBuildProvenanceNilExternalParameters(t *testing.T) {
	// externalParameters is REQUIRED for SLSA Build L1; a nil map must become a
	// non-nil empty object so the marshaled predicate carries the field.
	prov, err := BuildProvenancePredicate(BuilderIDLocal, BuildTypeLocal, nil, nil)
	if err != nil {
		t.Fatalf("BuildProvenancePredicate(nil ext) = %v", err)
	}
	if prov.BuildDefinition.ExternalParameters == nil {
		t.Fatal("externalParameters must be non-nil (required field)")
	}
}
