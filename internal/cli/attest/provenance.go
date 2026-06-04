package attest

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/attest"
)

// buildProvenance constructs the SLSA Provenance v1 predicate and returns it
// plus the subject name (the artifact's base name). It always enforces the
// ADR-0009 builder.id allowlist:
//
//   - With --predicate <file>, the file is trusted for shape but its
//     runDetails.builder.id is still validated against the allowlist.
//   - With --from-env on a GitHub-Actions runner, the release builder.id /
//     buildType and GITHUB_* externalParameters are used.
//   - Otherwise the local builder.id / buildType (the weaker, unverified tier)
//     is used.
//
// Any builder.id outside the allowlist (an SLSA overclaim) is refused with
// cmderr.ErrInvalidInput.
func buildProvenance(opts GenOptions, artifactPath string) (attest.Provenance, string, error) {
	subjName := filepath.Base(artifactPath)

	// Explicit predicate file: trust its shape but still validate builder.id.
	if opts.PredicatePath != "" {
		raw, err := readFile(opts.PredicatePath, cmderr.ErrNotFound, cmderr.ErrPermission)
		if err != nil {
			return attest.Provenance{}, "", err
		}
		var prov attest.Provenance
		if err := json.Unmarshal(raw, &prov); err != nil {
			return attest.Provenance{}, "", cmderr.Wrap(cmderr.ErrInvalidInput, "parse predicate JSON")
		}
		if err := attest.ValidateBuilderID(prov.RunDetails.Builder.ID); err != nil {
			return attest.Provenance{}, "", cmderr.Wrap(cmderr.ErrInvalidInput, "predicate builder.id (SLSA overclaim refused)")
		}
		if prov.BuildDefinition.ExternalParameters == nil {
			prov.BuildDefinition.ExternalParameters = map[string]any{}
		}
		return prov, subjName, nil
	}

	builderID := opts.BuilderID
	buildType := attest.BuildTypeLocal
	ext := map[string]any{"artifact": subjName}
	var meta *attest.BuildMetadata

	if opts.FromEnv && os.Getenv("GITHUB_ACTIONS") == "true" {
		builderID = attest.BuilderIDRelease
		buildType = attest.BuildTypeGHA
		ext = map[string]any{
			"workflow":   os.Getenv("GITHUB_WORKFLOW"),
			"repository": os.Getenv("GITHUB_REPOSITORY"),
			"ref":        os.Getenv("GITHUB_REF"),
			"sha":        os.Getenv("GITHUB_SHA"),
		}
		meta = &attest.BuildMetadata{
			InvocationID: os.Getenv("GITHUB_RUN_ID"),
			StartedOn:    time.Now().UTC().Format(time.RFC3339),
		}
	}
	if builderID == "" {
		builderID = attest.BuilderIDLocal
	}

	prov, err := attest.BuildProvenancePredicate(builderID, buildType, ext, meta)
	if err != nil {
		if errors.Is(err, attest.ErrOverclaim) {
			return attest.Provenance{}, "", cmderr.Wrap(cmderr.ErrInvalidInput, "builder.id (SLSA overclaim refused)")
		}
		return attest.Provenance{}, "", cmderr.Wrap(cmderr.ErrInvalidInput, "build provenance predicate")
	}
	return prov, subjName, nil
}
