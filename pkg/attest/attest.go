package attest

// Fixed wire constants (see the in-toto/DSSE/SLSA specs).
const (
	// StatementType is the fixed in-toto Statement v1 _type URI.
	StatementType = "https://in-toto.io/Statement/v1"
	// PredicateTypeSLSAProvenance is the fixed SLSA Provenance v1 predicateType URI.
	PredicateTypeSLSAProvenance = "https://slsa.dev/provenance/v1"
	// PayloadTypeInToto is the fixed DSSE payloadType for in-toto JSON statements.
	PayloadTypeInToto = "application/vnd.in-toto+json"
)

// ResourceDescriptor is the in-toto subject/dependency descriptor (digest-bearing subset).
type ResourceDescriptor struct {
	Name   string            `json:"name,omitempty"`
	URI    string            `json:"uri,omitempty"`
	Digest map[string]string `json:"digest,omitempty"`
}

// Statement is the in-toto Statement v1 envelope payload.
type Statement struct {
	Type          string               `json:"_type"`
	Subject       []ResourceDescriptor `json:"subject"`
	PredicateType string               `json:"predicateType"`
	Predicate     any                  `json:"predicate"`
}

// Provenance is the SLSA Provenance v1 predicate.
type Provenance struct {
	BuildDefinition BuildDefinition `json:"buildDefinition"`
	RunDetails      RunDetails      `json:"runDetails"`
}

// BuildDefinition describes the build inputs.
type BuildDefinition struct {
	BuildType            string               `json:"buildType"`
	ExternalParameters   map[string]any       `json:"externalParameters"`
	InternalParameters   map[string]any       `json:"internalParameters,omitempty"`
	ResolvedDependencies []ResourceDescriptor `json:"resolvedDependencies,omitempty"`
}

// RunDetails describes a single execution of the build.
type RunDetails struct {
	Builder    Builder              `json:"builder"`
	Metadata   *BuildMetadata       `json:"metadata,omitempty"`
	Byproducts []ResourceDescriptor `json:"byproducts,omitempty"`
}

// Builder identifies the trusted build platform. Builder.ID is the sole determiner of SLSA level.
type Builder struct {
	ID                  string               `json:"id"`
	Version             map[string]string    `json:"version,omitempty"`
	BuilderDependencies []ResourceDescriptor `json:"builderDependencies,omitempty"`
}

// BuildMetadata captures invocation timing.
type BuildMetadata struct {
	InvocationID string `json:"invocationId,omitempty"`
	StartedOn    string `json:"startedOn,omitempty"`
	FinishedOn   string `json:"finishedOn,omitempty"`
}

// Envelope is the DSSE JSON envelope.
type Envelope struct {
	Payload     string              `json:"payload"`
	PayloadType string              `json:"payloadType"`
	Signatures  []EnvelopeSignature `json:"signatures"`
}

// EnvelopeSignature is one signature over PAE(payloadType, payload-bytes).
type EnvelopeSignature struct {
	KeyID string `json:"keyid,omitempty"`
	Sig   string `json:"sig"`
}
