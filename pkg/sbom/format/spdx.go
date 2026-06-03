package format

import (
	"encoding/json"
	"io"
)

// spdxDoc is the minimal SPDX-2.3 document. Field declaration order is the
// on-disk JSON key order (Go marshals struct fields in declaration order).
type spdxDoc struct {
	SPDXVersion       string             `json:"spdxVersion"`
	DataLicense       string             `json:"dataLicense"`
	SPDXID            string             `json:"SPDXID"`
	Name              string             `json:"name"`
	DocumentNamespace string             `json:"documentNamespace"`
	CreationInfo      spdxCreation       `json:"creationInfo"`
	Packages          []spdxPackage      `json:"packages"`
	Relationships     []spdxRelationship `json:"relationships"`
}

type spdxCreation struct {
	Created  string   `json:"created"`
	Creators []string `json:"creators"`
}

type spdxExtRef struct {
	ReferenceCategory string `json:"referenceCategory"`
	ReferenceType     string `json:"referenceType"`
	ReferenceLocator  string `json:"referenceLocator"`
}

type spdxPackage struct {
	Name             string       `json:"name"`
	SPDXID           string       `json:"SPDXID"`
	VersionInfo      string       `json:"versionInfo,omitempty"`
	DownloadLocation string       `json:"downloadLocation"`
	FilesAnalyzed    bool         `json:"filesAnalyzed"`
	LicenseConcluded string       `json:"licenseConcluded"`
	LicenseDeclared  string       `json:"licenseDeclared"`
	CopyrightText    string       `json:"copyrightText"`
	ExternalRefs     []spdxExtRef `json:"externalRefs,omitempty"`
}

type spdxRelationship struct {
	SPDXElementID      string `json:"spdxElementId"`
	RelatedSPDXElement string `json:"relatedSpdxElement"`
	RelationshipType   string `json:"relationshipType"`
}

// writeJSON emits v as deterministic, indented JSON with a trailing newline.
func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v) // Encode appends a trailing newline.
}

func (d *Document) encodeSPDX(w io.Writer) error {
	doc := spdxDoc{
		SPDXVersion:       "SPDX-2.3",
		DataLicense:       "CC0-1.0",
		SPDXID:            "SPDXRef-DOCUMENT",
		Name:              d.name,
		DocumentNamespace: "https://spdx.org/spdxdocs/omni/" + d.name + "-" + d.contentHash,
		CreationInfo: spdxCreation{
			Created:  d.created,
			Creators: []string{"Tool: omni-sbom-" + d.omniVersion},
		},
	}

	doc.Packages = append(doc.Packages, spdxPackageFor(d.root))
	for _, e := range d.entries {
		doc.Packages = append(doc.Packages, spdxPackageFor(e))
	}

	rootID := "SPDXRef-Package-" + d.root.slug
	doc.Relationships = append(doc.Relationships, spdxRelationship{
		SPDXElementID:      "SPDXRef-DOCUMENT",
		RelatedSPDXElement: rootID,
		RelationshipType:   "DESCRIBES",
	})
	for _, e := range d.entries {
		doc.Relationships = append(doc.Relationships, spdxRelationship{
			SPDXElementID:      rootID,
			RelatedSPDXElement: "SPDXRef-Package-" + e.slug,
			RelationshipType:   "DEPENDS_ON",
		})
	}

	return writeJSON(w, doc)
}

// spdxPackageFor builds the SPDX package record for one resolved entry.
func spdxPackageFor(e entry) spdxPackage {
	p := spdxPackage{
		Name:             e.c.Path,
		SPDXID:           "SPDXRef-Package-" + e.slug,
		VersionInfo:      e.c.Version,
		DownloadLocation: "NOASSERTION",
		FilesAnalyzed:    false,
		LicenseConcluded: "NOASSERTION",
		LicenseDeclared:  "NOASSERTION",
		CopyrightText:    "NOASSERTION",
	}
	if e.purl != "" {
		p.ExternalRefs = []spdxExtRef{{
			ReferenceCategory: "PACKAGE-MANAGER",
			ReferenceType:     "purl",
			ReferenceLocator:  e.purl,
		}}
	}
	return p
}
