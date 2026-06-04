package format

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/inovacc/omni/pkg/sbom/purl"
)

// Component is the exported, format-agnostic READ view of one SBOM component,
// for consumers (pkg/scan) that read a Document. Name is the Go module path,
// Version its module version (may be empty), PURL the full package-URL, Ecosystem
// the purl type (e.g. "golang").
type Component struct {
	Name      string
	Version   string
	PURL      string
	Ecosystem string
}

// Components returns every component carried by the Document (root included),
// in the Document's deterministic order. The read side of the stable boundary;
// pkg/scan depends on this, never on pkg/sbom/model. Entries with no purl
// (e.g. the synthetic root of a Parse-built Document) are skipped.
func (d *Document) Components() []Component {
	out := make([]Component, 0, len(d.entries)+1)
	for _, e := range append([]entry{d.root}, d.entries...) {
		if e.purl == "" {
			continue // a Parse-built Document has no root purl
		}
		mod, ver, ok := purl.Parse(e.purl)
		eco := ""
		if ok {
			eco = "golang"
		}
		out = append(out, Component{Name: mod, Version: ver, PURL: e.purl, Ecosystem: eco})
	}
	return out
}

// Parse reads an SPDX-2.3 or CycloneDX-1.5 JSON SBOM and returns a Document whose
// Components() yields its Go components. Detection: "spdxVersion" => SPDX;
// "bomFormat":"CycloneDX" or "specVersion" => CycloneDX. Unknown fields are ignored
// so third-party SBOMs (e.g. syft output) parse cleanly. Only entry.purl is
// populated — the read side needs nothing else.
func Parse(r io.Reader) (*Document, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var probe struct {
		SPDXVersion string `json:"spdxVersion"`
		BOMFormat   string `json:"bomFormat"`
		SpecVersion string `json:"specVersion"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, fmt.Errorf("sbom: not valid JSON: %w", err)
	}
	switch {
	case probe.SPDXVersion != "":
		return parseSPDX(raw)
	case probe.BOMFormat == "CycloneDX" || probe.SpecVersion != "":
		return parseCycloneDX(raw)
	default:
		return nil, fmt.Errorf("sbom: unrecognized format (no spdxVersion/bomFormat/specVersion)")
	}
}

func parseSPDX(raw []byte) (*Document, error) {
	var doc struct {
		Name     string `json:"name"`
		Packages []struct {
			ExternalRefs []struct {
				ReferenceType    string `json:"referenceType"`
				ReferenceLocator string `json:"referenceLocator"`
			} `json:"externalRefs"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("sbom: bad SPDX: %w", err)
	}
	d := &Document{name: doc.Name}
	for _, p := range doc.Packages {
		for _, ref := range p.ExternalRefs {
			if ref.ReferenceType == "purl" && ref.ReferenceLocator != "" {
				d.entries = append(d.entries, entry{purl: ref.ReferenceLocator})
			}
		}
	}
	return d, nil
}

func parseCycloneDX(raw []byte) (*Document, error) {
	var bom struct {
		Metadata struct {
			Component struct {
				Name string `json:"name"`
			} `json:"component"`
		} `json:"metadata"`
		Components []struct {
			PURL string `json:"purl"`
		} `json:"components"`
	}
	if err := json.Unmarshal(raw, &bom); err != nil {
		return nil, fmt.Errorf("sbom: bad CycloneDX: %w", err)
	}
	d := &Document{name: bom.Metadata.Component.Name}
	for _, c := range bom.Components {
		if c.PURL != "" {
			d.entries = append(d.entries, entry{purl: c.PURL})
		}
	}
	return d, nil
}
