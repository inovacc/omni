package format

import (
	"io"
	"sort"

	"github.com/inovacc/omni/pkg/sbom/model"
)

// cdxBOM is the minimal CycloneDX-1.5 BOM. Field declaration order is the
// on-disk JSON key order. No serialNumber field exists, so it is always
// omitted (it would otherwise be random and break determinism).
type cdxBOM struct {
	Schema       string          `json:"$schema"`
	BOMFormat    string          `json:"bomFormat"`
	SpecVersion  string          `json:"specVersion"`
	Version      int             `json:"version"`
	Metadata     cdxMetadata     `json:"metadata"`
	Components   []cdxComponent  `json:"components"`
	Dependencies []cdxDependency `json:"dependencies"`
}

type cdxMetadata struct {
	Timestamp string       `json:"timestamp"`
	Tools     cdxTools     `json:"tools"`
	Component cdxComponent `json:"component"`
}

type cdxTools struct {
	Components []cdxComponent `json:"components"`
}

type cdxComponent struct {
	BOMRef  string `json:"bom-ref,omitempty"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	PURL    string `json:"purl,omitempty"`
}

type cdxDependency struct {
	Ref       string   `json:"ref"`
	DependsOn []string `json:"dependsOn,omitempty"`
}

func (d *Document) encodeCycloneDX(w io.Writer) error {
	rootRef := d.root.purl
	if rootRef == "" {
		rootRef = d.root.slug
	}

	bom := cdxBOM{
		Schema:      "http://cyclonedx.org/schema/bom-1.5.schema.json",
		BOMFormat:   "CycloneDX",
		SpecVersion: "1.5",
		Version:     1,
		Metadata: cdxMetadata{
			Timestamp: d.created,
			Tools: cdxTools{
				Components: []cdxComponent{{
					Type:    "application",
					Name:    "omni",
					Version: d.omniVersion,
				}},
			},
			Component: cdxComponent{
				BOMRef:  rootRef,
				Type:    "application",
				Name:    d.root.c.Path,
				Version: d.root.c.Version,
				PURL:    d.root.purl,
			},
		},
	}

	deps := make([]string, 0, len(d.entries))
	for _, e := range d.entries {
		bom.Components = append(bom.Components, cdxComponentFor(e))
		deps = append(deps, e.purl)
	}
	sort.Strings(deps)

	bom.Dependencies = append(bom.Dependencies, cdxDependency{Ref: rootRef, DependsOn: deps})
	for _, e := range d.entries {
		bom.Dependencies = append(bom.Dependencies, cdxDependency{Ref: e.purl})
	}

	return writeJSON(w, bom)
}

// cdxComponentFor maps a resolved entry to its CycloneDX component record.
// The Go toolchain is emitted as an "application"; everything else "library".
func cdxComponentFor(e entry) cdxComponent {
	typ := "library"
	if e.c.Kind == model.KindToolchain {
		typ = "application"
	}
	return cdxComponent{
		BOMRef:  e.purl,
		Type:    typ,
		Name:    e.c.Path,
		Version: e.c.Version,
		PURL:    e.purl,
	}
}
