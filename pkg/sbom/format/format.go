package format

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"sort"

	"github.com/inovacc/omni/pkg/sbom/model"
	"github.com/inovacc/omni/pkg/sbom/purl"
)

// Kind selects the output document format.
type Kind int

const (
	// SPDX selects SPDX 2.3 JSON.
	SPDX Kind = iota
	// CycloneDX selects CycloneDX 1.5 JSON.
	CycloneDX
)

const defaultDate = "1970-01-01T00:00:00Z"

// errUnknownFormat is returned by Encode when given an unrecognized Kind.
var errUnknownFormat = errors.New("unknown sbom format")

// Options tunes document generation. SourceDate (RFC-3339) is the fixed
// creation timestamp; empty means the epoch default (keeps output deterministic
// with no flag). OmniVersion labels the generating tool.
type Options struct {
	OmniVersion string
	SourceDate  string
}

// entry is the internal, fully-resolved view of one component.
type entry struct {
	c    model.Component
	purl string
	slug string
}

// Document is the STABLE cross-package boundary type. pkg/scan/ (Phase 6)
// depends only on this type, never on pkg/sbom/model. It carries a resolved,
// sorted, format-agnostic snapshot ready for deterministic emission.
type Document struct {
	name        string
	created     string
	omniVersion string
	root        entry
	entries     []entry // sorted by purl/slug; excludes root
	contentHash string  // 16 hex chars over sorted purls
}

// From resolves a model.SBOM into a Document (purls + slugs computed, slices
// sorted, content hash derived). The input is normalized defensively.
func From(s *model.SBOM, opt Options) *Document {
	s.Normalize()
	created := opt.SourceDate
	if created == "" {
		created = defaultDate
	}
	d := &Document{
		name:        s.Name,
		created:     created,
		omniVersion: opt.OmniVersion,
		root:        toEntry(s.Root),
	}
	slugs := map[string]int{}
	dedupeSlug(&d.root, slugs)
	purls := make([]string, 0, len(s.Components))
	for _, c := range s.Components {
		e := toEntry(c)
		dedupeSlug(&e, slugs)
		d.entries = append(d.entries, e)
		purls = append(purls, e.purl)
	}
	sort.Slice(d.entries, func(i, j int) bool { return d.entries[i].purl < d.entries[j].purl })
	sort.Strings(purls)
	h := sha256.New()
	for _, p := range purls {
		_, _ = h.Write([]byte(p))
		_, _ = h.Write([]byte{'\n'})
	}
	d.contentHash = hex.EncodeToString(h.Sum(nil))[:16]
	return d
}

func toEntry(c model.Component) entry {
	return entry{c: c, purl: purl.ForModule(c.Path, c.Version), slug: model.Slug(c.Path)}
}

// dedupeSlug ensures slugs are unique by appending -<n> on collision.
func dedupeSlug(e *entry, seen map[string]int) {
	base := e.slug
	if n, ok := seen[base]; ok {
		seen[base] = n + 1
		e.slug = base + "-" + itoa(n+1)
	} else {
		seen[base] = 0
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// Encode writes the document in the requested format to w (deterministic bytes,
// trailing newline).
func (d *Document) Encode(w io.Writer, k Kind) error {
	switch k {
	case SPDX:
		return d.encodeSPDX(w)
	case CycloneDX:
		return d.encodeCycloneDX(w)
	default:
		return errUnknownFormat
	}
}
