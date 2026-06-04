package scan

import (
	"sort"
	"time"

	"github.com/inovacc/omni/internal/cli/cmderr"
	"github.com/inovacc/omni/pkg/sbom/format"
)

// component is the normalized (module path, version) pair the matcher consumes.
type component struct {
	Pkg     string
	Version string
}

// Options configures a scan.
type Options struct {
	// FailOn trips cmderr.ErrConflict when any finding's severity is >= FailOn.
	// SeverityUnknown (the zero value) disables the gate.
	FailOn Severity
	// Reachability is reserved for the future contrib reachability path; it is
	// ignored in v1.0 (scan source returns ErrUnsupported per ADR-0008).
	Reachability bool
}

// componentsOf adapts the SBOM boundary type into normalized components. THIS is
// the single point of coupling to pkg/sbom/format: only "golang"-ecosystem
// components with a non-empty module path and version are scanned.
func componentsOf(doc *format.Document) []component {
	if doc == nil {
		return nil
	}
	comps := doc.Components()
	out := make([]component, 0, len(comps))
	for _, c := range comps {
		if c.Ecosystem != "golang" {
			continue // non-Go ecosystems are not matchable against the Go OSV DB
		}
		pkg := modulePathFromComponent(c)
		if pkg == "" || c.Version == "" {
			continue
		}
		out = append(out, component{Pkg: pkg, Version: c.Version})
	}
	return out
}

// modulePathFromComponent prefers the Go module path carried on the component;
// it derives the path from the component's purl when Name is empty.
func modulePathFromComponent(c format.Component) string {
	if c.Name != "" {
		return c.Name
	}
	return modulePathFromPURL(c.PURL) // strips "pkg:golang/" prefix and any @version
}

// Scan scans an SBOM Document against db and applies opts. It adapts the
// Document's components through the single pkg/sbom/format coupling point
// (componentsOf) and delegates to scanComponents for matching and gating.
func Scan(doc *format.Document, db *DB, opts Options) (Report, error) {
	return scanComponents(componentsOf(doc), db, opts)
}

// scanComponents matches every component against db, builds a deterministically
// ordered Report, and applies the --fail-on gate (findings at or above
// opts.FailOn trip cmderr.ErrConflict).
func scanComponents(comps []component, db *DB, opts Options) (Report, error) {
	var findings []Finding
	for _, c := range comps {
		for _, e := range db.entriesFor(c.Pkg) {
			if f, ok := matchEntry(e, c.Pkg, c.Version); ok {
				findings = append(findings, f)
			}
		}
	}
	sortFindings(findings)
	rep := Report{Findings: findings, Scanned: len(comps), DBAge: db.Age().Round(time.Second).String()}

	if opts.FailOn > SeverityUnknown {
		for _, f := range findings {
			sev, _ := ParseSeverity(f.Severity)
			if sev >= opts.FailOn {
				return rep, cmderr.Wrap(cmderr.ErrConflict,
					"vulnerabilities found at or above --fail-on threshold")
			}
		}
	}
	return rep, nil
}

// sortFindings orders findings deterministically by (Package, ID).
func sortFindings(findings []Finding) {
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Package != findings[j].Package {
			return findings[i].Package < findings[j].Package
		}
		return findings[i].ID < findings[j].ID
	})
}
