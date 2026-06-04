package scan

import (
	"net/url"
	"strings"

	"golang.org/x/mod/semver"
)

// modulePathFromPURL extracts the Go module path from a "pkg:golang/<module>"
// package-URL: it strips the "pkg:golang/" prefix, drops any "@version" suffix,
// and URL-decodes the path. It returns "" for any purl whose type is not golang.
func modulePathFromPURL(s string) string {
	const prefix = "pkg:golang/"
	if !strings.HasPrefix(s, prefix) {
		return ""
	}
	body := s[len(prefix):]
	if at := strings.IndexByte(body, '@'); at >= 0 {
		body = body[:at]
	}
	if dec, err := url.PathUnescape(body); err == nil {
		return dec
	}
	return body
}

// sv normalizes an SBOM/OSV version to an x/mod/semver-comparable string by
// prepending a leading "v" when absent. Empty input yields empty output.
func sv(v string) string {
	if v == "" {
		return ""
	}
	if v[0] == 'v' {
		return v
	}
	return "v" + v
}

// matchEntry reports whether (pkg, version) is hit by entry e, and builds the
// Finding. Per ADR-0008: the affected ecosystem must be "Go" and its name must
// equal pkg; then exact versions[] membership OR an open SEMVER interval makes
// the version vulnerable. An affected-level severity overrides the top-level one.
func matchEntry(e osvEntry, pkg, version string) (Finding, bool) {
	for _, a := range e.Affected {
		if a.Package.Ecosystem != "Go" || a.Package.Name != pkg {
			continue
		}
		hit, fixed := affectedHit(a, version)
		if !hit {
			continue
		}
		sev := e.Severity
		if len(a.Severity) > 0 {
			sev = a.Severity // affected-level severity overrides top-level
		}
		return Finding{
			ID:           e.ID,
			Package:      pkg,
			Version:      version,
			FixedVersion: fixed,
			Severity:     severityLabel(sev).String(),
			Summary:      e.Summary,
		}, true
	}
	return Finding{}, false
}

// affectedHit decides whether version is vulnerable within a single affected
// block and returns the smallest applicable fix bound (or "").
func affectedHit(a osvAffected, version string) (bool, string) {
	for _, v := range a.Versions { // exact-membership shortcut (also covers ECOSYSTEM/GIT)
		if v == version {
			return true, smallestFixAbove(a, version)
		}
	}
	for _, r := range a.Ranges {
		if r.Type != "SEMVER" {
			continue // ECOSYSTEM and GIT ranges handled by exact versions only
		}
		if inOpenInterval(r.Events, version) {
			return true, smallestFixAbove(a, version)
		}
	}
	return false, ""
}

// inOpenInterval walks ordered events: introduced opens an interval (empty/"0"
// is genesis), fixed closes it exclusively ([introduced, fixed)), last_affected
// closes it inclusively ([introduced, last_affected]). An introduced with no
// later closing event means "all versions >= introduced".
func inOpenInterval(events []rngEvent, version string) bool {
	open := false
	cur := sv(version)
	for _, ev := range events {
		switch {
		case ev.Introduced != "":
			lo := ev.Introduced
			if lo == "0" {
				open = true
				continue
			}
			open = semver.Compare(cur, sv(lo)) >= 0
		case ev.Fixed != "":
			if open && semver.Compare(cur, sv(ev.Fixed)) < 0 {
				return true
			}
			open = false
		case ev.LastAffected != "":
			if open && semver.Compare(cur, sv(ev.LastAffected)) <= 0 {
				return true
			}
			open = false
		}
	}
	return open // introduced with no later closing event => all >= introduced
}

// smallestFixAbove returns the smallest "fixed" bound strictly greater than
// version, or "" if none. last_affected events do not contribute a fix bound.
func smallestFixAbove(a osvAffected, version string) string {
	best := ""
	cur := sv(version)
	for _, r := range a.Ranges {
		for _, ev := range r.Events {
			if ev.Fixed == "" {
				continue
			}
			if semver.Compare(sv(ev.Fixed), cur) > 0 {
				if best == "" || semver.Compare(sv(ev.Fixed), sv(best)) < 0 {
					best = ev.Fixed
				}
			}
		}
	}
	return best
}
