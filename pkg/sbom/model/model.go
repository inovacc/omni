package model

import (
	"sort"
	"strings"
)

// Kind describes whether a Component is the root subject or a dependency.
type Kind int

const (
	// KindRoot is the main module / binary the SBOM describes.
	KindRoot Kind = iota
	// KindLibrary is a dependency module.
	KindLibrary
	// KindToolchain is the Go toolchain (binary SBOMs only).
	KindToolchain
)

// Component is one node in the SBOM: a Go module, the root, or the toolchain.
type Component struct {
	Path            string // module path (effective path after replace)
	Version         string // module version ("" if unknown)
	Kind            Kind
	OriginalPath    string // pre-replace module path, "" if not replaced
	OriginalVersion string // pre-replace version, "" if not replaced
}

// SBOM is the source-format-agnostic representation produced by collectors and
// consumed by the SPDX/CycloneDX emitters.
type SBOM struct {
	Name       string      // document/subject name
	Root       Component   // the described subject
	Components []Component // dependencies (+ toolchain), excludes Root
}

// Normalize sorts Components deterministically by (Path, Version) and removes
// exact (Path,Version) duplicates so emitted output is stable.
func (s *SBOM) Normalize() {
	sort.Slice(s.Components, func(i, j int) bool {
		if s.Components[i].Path != s.Components[j].Path {
			return s.Components[i].Path < s.Components[j].Path
		}
		return s.Components[i].Version < s.Components[j].Version
	})
	out := s.Components[:0]
	var prevP, prevV string
	first := true
	for _, c := range s.Components {
		if !first && c.Path == prevP && c.Version == prevV {
			continue
		}
		out = append(out, c)
		prevP, prevV, first = c.Path, c.Version, false
	}
	s.Components = out
}

// Slug converts a module path into an SPDX-ID-safe token: every rune that is
// not a letter, digit, or '.' becomes '-'.
func Slug(path string) string {
	var b strings.Builder
	for _, r := range path {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	return b.String()
}
