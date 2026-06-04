package attest

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

// Builder-id / build-type constants pinned by ADR-0009. The generator may emit
// ONLY these builder IDs; anything else is an SLSA overclaim and is refused.
const (
	// BuilderIDRelease is the dogfooded GitHub-Actions release path (SLSA Build
	// L2). It is emitted only when --from-env detects the GitHub-Actions runner.
	BuilderIDRelease = "https://github.com/inovacc/omni/.github/workflows/release.yml@refs/heads/main"
	// BuilderIDLocal is the local/non-CI fallback (a LOWER, unverified tier,
	// effectively L1) used for developer-machine generation.
	BuilderIDLocal = "https://github.com/inovacc/omni/attest/local@v1"
	// BuildTypeGHA is the SLSA buildType for the GitHub-Actions workflow build.
	BuildTypeGHA = "https://slsa-framework.github.io/github-actions-buildtypes/workflow/v1"
	// BuildTypeLocal is the buildType for local (developer-machine) generation.
	BuildTypeLocal = "https://github.com/inovacc/omni/attest/local-buildtype/v1"
)

// ErrOverclaim is returned when a requested builder.id is not in the ADR-0009
// allowlist (i.e. would claim a higher SLSA level than omni honestly achieves).
// The CLI maps it to cmderr.ErrInvalidInput.
var ErrOverclaim = errors.New("attest: builder.id not permitted by ADR-0009 (SLSA overclaim)")

// allowedBuilderIDs is the ADR-0009 allowlist.
var allowedBuilderIDs = map[string]bool{
	BuilderIDRelease: true,
	BuilderIDLocal:   true,
}

// ValidateBuilderID returns ErrOverclaim unless id is an ADR-0009-pinned value.
func ValidateBuilderID(id string) error {
	if !allowedBuilderIDs[id] {
		return fmt.Errorf("%w: %q", ErrOverclaim, id)
	}
	return nil
}

// SubjectFromBytes builds an in-toto subject descriptor with a sha256 digest
// (lowercase hex) computed over data.
func SubjectFromBytes(name string, data []byte) ResourceDescriptor {
	sum := sha256.Sum256(data)
	return ResourceDescriptor{Name: name, Digest: map[string]string{"sha256": hex.EncodeToString(sum[:])}}
}

// BuildProvenancePredicate assembles a SLSA Provenance v1 predicate from a
// builder.id, buildType, and externalParameters. It enforces the ADR-0009
// builder.id allowlist (returning ErrOverclaim for any disallowed id) and
// guarantees a non-nil externalParameters object, which is REQUIRED for SLSA
// Build L1. A nil meta omits runDetails.metadata.
func BuildProvenancePredicate(builderID, buildType string, externalParameters map[string]any, meta *BuildMetadata) (Provenance, error) {
	if err := ValidateBuilderID(builderID); err != nil {
		return Provenance{}, err
	}
	if externalParameters == nil {
		externalParameters = map[string]any{}
	}
	return Provenance{
		BuildDefinition: BuildDefinition{
			BuildType:          buildType,
			ExternalParameters: externalParameters,
		},
		RunDetails: RunDetails{
			Builder:  Builder{ID: builderID},
			Metadata: meta,
		},
	}, nil
}

// NewStatement wraps subjects + a SLSA provenance predicate into an in-toto
// Statement v1 with the fixed _type and predicateType URIs.
func NewStatement(subject []ResourceDescriptor, prov Provenance) Statement {
	return Statement{
		Type:          StatementType,
		Subject:       subject,
		PredicateType: PredicateTypeSLSAProvenance,
		Predicate:     prov,
	}
}
