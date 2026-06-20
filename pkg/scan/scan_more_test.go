package scan

import (
	"math"
	"strings"
	"testing"

	"github.com/inovacc/omni/pkg/sbom/format"
)

func TestSeverityString(t *testing.T) {
	cases := map[Severity]string{
		SeverityUnknown:  "unknown",
		SeverityNone:     "none",
		SeverityLow:      "low",
		SeverityMedium:   "medium",
		SeverityHigh:     "high",
		SeverityCritical: "critical",
		Severity(99):     "unknown",
	}
	for s, want := range cases {
		if got := s.String(); got != want {
			t.Errorf("Severity(%d).String() = %q, want %q", s, got, want)
		}
	}
}

func TestParseSeverity(t *testing.T) {
	cases := []struct {
		in   string
		want Severity
		ok   bool
	}{
		{"none", SeverityNone, true},
		{"low", SeverityLow, true},
		{"medium", SeverityMedium, true},
		{"high", SeverityHigh, true},
		{"critical", SeverityCritical, true},
		{"bogus", SeverityUnknown, false},
		{"", SeverityUnknown, false},
	}
	for _, c := range cases {
		got, ok := ParseSeverity(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("ParseSeverity(%q) = (%v,%v), want (%v,%v)", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestModulePathFromPURL(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"pkg:golang/github.com/foo/bar@v1.2.3", "github.com/foo/bar"},
		{"pkg:golang/github.com/foo/bar", "github.com/foo/bar"},
		{"pkg:golang/golang.org%2Fx%2Fnet@v0.1.0", "golang.org/x/net"},
		{"pkg:npm/left-pad@1.0.0", ""}, // non-golang type
		{"not-a-purl", ""},
	}
	for _, c := range cases {
		if got := modulePathFromPURL(c.in); got != c.want {
			t.Errorf("modulePathFromPURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestModulePathFromComponent(t *testing.T) {
	// Name present -> used directly.
	if got := modulePathFromComponent(format.Component{Name: "github.com/a/b"}); got != "github.com/a/b" {
		t.Errorf("from Name = %q", got)
	}
	// Name empty -> derived from PURL.
	got := modulePathFromComponent(format.Component{PURL: "pkg:golang/github.com/c/d@v1.0.0"})
	if got != "github.com/c/d" {
		t.Errorf("from PURL = %q", got)
	}
}

const cycloneDXFixture = `{
  "bomFormat": "CycloneDX",
  "specVersion": "1.5",
  "metadata": {"component": {"name": "root"}},
  "components": [
    {"purl": "pkg:golang/github.com/foo/bar@v1.0.0"},
    {"purl": "pkg:golang/github.com/safe/lib@v2.0.0"},
    {"purl": "pkg:npm/ignored@1.0.0"}
  ]
}`

func TestComponentsOfAndScan(t *testing.T) {
	doc, err := format.Parse(strings.NewReader(cycloneDXFixture))
	if err != nil {
		t.Fatalf("format.Parse: %v", err)
	}
	comps := componentsOf(doc)
	// npm component is filtered out -> only the 2 golang ones remain.
	if len(comps) != 2 {
		t.Fatalf("componentsOf = %d comps, want 2: %+v", len(comps), comps)
	}

	db := &DB{byPkg: map[string][]osvEntry{
		"github.com/foo/bar": {entry("GO-7", "github.com/foo/bar", nil,
			[]rng{{Type: "SEMVER", Events: []rngEvent{{Introduced: "0"}, {Fixed: "1.2.0"}}}},
			[]rawSeverity{{Type: "CVSS_V3", Score: "7.5"}})},
	}}

	rep, err := Scan(doc, db, Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if rep.Scanned != 2 {
		t.Errorf("Scanned = %d, want 2", rep.Scanned)
	}
	if len(rep.Findings) != 1 || rep.Findings[0].ID != "GO-7" {
		t.Fatalf("findings = %+v", rep.Findings)
	}
}

func TestComponentsOfNil(t *testing.T) {
	if got := componentsOf(nil); got != nil {
		t.Errorf("componentsOf(nil) = %v, want nil", got)
	}
}

func TestCVSS31BaseScore(t *testing.T) {
	cases := []struct {
		name   string
		typ    string
		vector string
		want   float64
		ok     bool
	}{
		{"critical unchanged", "CVSS_V3", "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H", 9.8, true},
		{"scope changed", "CVSS_V3", "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H", 10.0, true},
		{"none impact", "CVSS_V3", "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:N/A:N", 0.0, true},
		{"missing metric", "CVSS_V3", "CVSS:3.1/AV:N/AC:L", 0, false},
		{"not v3 type", "CVSS_V4", "anything", 0, false},
		{"bad scope value omitted", "CVSS_V3", "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/C:H/I:H/A:H", 0, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := cvssBaseScore(c.typ, c.vector)
			if ok != c.ok {
				t.Fatalf("cvssBaseScore ok = %v, want %v", ok, c.ok)
			}
			if ok && math.Abs(got-c.want) > 0.05 {
				t.Errorf("cvssBaseScore = %v, want %v", got, c.want)
			}
		})
	}
}

func TestSeverityLabelFromVector(t *testing.T) {
	// No numeric score -> computed from vector -> critical band.
	got := severityLabel([]rawSeverity{{Type: "CVSS_V3", Score: "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H"}})
	if got != SeverityCritical {
		t.Errorf("severityLabel(vector) = %v, want critical", got)
	}
	// Unparseable score and unparseable vector -> unknown.
	if got := severityLabel([]rawSeverity{{Type: "CVSS_V3", Score: "garbage"}}); got != SeverityUnknown {
		t.Errorf("severityLabel(garbage) = %v, want unknown", got)
	}
}

func TestRoundup(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{4.0, 4.0},
		{4.02, 4.1},
		{0.0, 0.0},
	}
	for _, c := range cases {
		if got := roundup(c.in); math.Abs(got-c.want) > 0.001 {
			t.Errorf("roundup(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}
