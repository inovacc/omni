package scan

import "testing"

func entry(id, name string, versions []string, ranges []rng, sevs []rawSeverity) osvEntry {
	return osvEntry{
		ID:       id,
		Severity: sevs,
		Affected: []osvAffected{{
			Package:  osvPackage{Ecosystem: "Go", Name: name},
			Versions: versions,
			Ranges:   ranges,
		}},
	}
}

func TestMatchSemverRange(t *testing.T) {
	e := entry("GO-1", "github.com/foo/bar", nil,
		[]rng{{Type: "SEMVER", Events: []rngEvent{{Introduced: "0"}, {Fixed: "1.2.3"}}}},
		[]rawSeverity{{Type: "CVSS_V3", Score: "7.5"}})

	cases := map[string]struct {
		version string
		want    bool
		fixed   string
	}{
		"vulnerable below fixed": {"1.0.0", true, "1.2.3"},
		"safe at fixed":          {"1.2.3", false, ""},
		"safe above fixed":       {"2.0.0", false, ""},
		"vulnerable at zero":     {"0.0.1", true, "1.2.3"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			f, ok := matchEntry(e, "github.com/foo/bar", c.version)
			if ok != c.want {
				t.Fatalf("matchEntry ok = %v, want %v", ok, c.want)
			}
			if ok && f.FixedVersion != c.fixed {
				t.Errorf("FixedVersion = %q, want %q", f.FixedVersion, c.fixed)
			}
			if ok && f.Severity != "high" {
				t.Errorf("Severity = %q, want high", f.Severity)
			}
		})
	}
}

func TestMatchOpenEndedRange(t *testing.T) {
	e := entry("GO-2", "github.com/foo/bar", nil,
		[]rng{{Type: "SEMVER", Events: []rngEvent{{Introduced: "1.5.0"}}}}, nil)
	if _, ok := matchEntry(e, "github.com/foo/bar", "2.0.0"); !ok {
		t.Error("2.0.0 must be vulnerable (introduced 1.5.0, no fix)")
	}
	if _, ok := matchEntry(e, "github.com/foo/bar", "1.4.0"); ok {
		t.Error("1.4.0 must be safe (below introduced)")
	}
}

func TestMatchLastAffectedInclusive(t *testing.T) {
	// last_affected closes the interval INCLUSIVELY: the version equal to
	// last_affected is still vulnerable (unlike fixed, which is exclusive).
	e := entry("GO-5", "github.com/foo/bar", nil,
		[]rng{{Type: "SEMVER", Events: []rngEvent{{Introduced: "1.0.0"}, {LastAffected: "1.4.0"}}}}, nil)
	if _, ok := matchEntry(e, "github.com/foo/bar", "1.4.0"); !ok {
		t.Error("1.4.0 must be vulnerable (last_affected is inclusive)")
	}
	if _, ok := matchEntry(e, "github.com/foo/bar", "1.4.1"); ok {
		t.Error("1.4.1 must be safe (above last_affected)")
	}
	if _, ok := matchEntry(e, "github.com/foo/bar", "1.0.0"); !ok {
		t.Error("1.0.0 must be vulnerable (at introduced)")
	}
}

func TestMatchExactVersionsList(t *testing.T) {
	e := entry("GO-3", "github.com/foo/bar", []string{"1.0.0", "1.0.1"}, nil, nil)
	if _, ok := matchEntry(e, "github.com/foo/bar", "1.0.1"); !ok {
		t.Error("1.0.1 in versions list must match")
	}
	if _, ok := matchEntry(e, "github.com/foo/bar", "1.0.2"); ok {
		t.Error("1.0.2 not in versions list must not match")
	}
}

func TestMatchExactVersionsVPrefixAsymmetry(t *testing.T) {
	// OSV versions[] are typically bare ("1.2.3") while an SBOM purl version may
	// carry a leading "v" ("v1.2.3"). The exact-membership path must normalize
	// both operands like the SEMVER interval path, otherwise ECOSYSTEM/GIT
	// advisories silently MISS the vulnerable component.
	e := entry("GO-7", "github.com/foo/bar", []string{"1.2.3"}, nil, nil)
	if _, ok := matchEntry(e, "github.com/foo/bar", "v1.2.3"); !ok {
		t.Error(`OSV versions=["1.2.3"] vs SBOM "v1.2.3" must HIT`)
	}
	// And the reverse asymmetry: OSV with a "v" prefix vs bare SBOM version.
	e2 := entry("GO-8", "github.com/foo/bar", []string{"v1.2.3"}, nil, nil)
	if _, ok := matchEntry(e2, "github.com/foo/bar", "1.2.3"); !ok {
		t.Error(`OSV versions=["v1.2.3"] vs SBOM "1.2.3" must HIT`)
	}
	// Non-semver pseudo/tag versions still fall back to a raw equality compare.
	e3 := entry("GO-9", "github.com/foo/bar",
		[]string{"v0.0.0-20240101000000-abcdef123456"}, nil, nil)
	if _, ok := matchEntry(e3, "github.com/foo/bar", "v0.0.0-20240101000000-abcdef123456"); !ok {
		t.Error("non-semver pseudo-version must still match via raw equality")
	}
	if _, ok := matchEntry(e3, "github.com/foo/bar", "v0.0.0-20240101000000-000000000000"); ok {
		t.Error("different non-semver pseudo-version must not match")
	}
}

func TestMatchEcosystemRangeUsesVersionsOnly(t *testing.T) {
	// ECOSYSTEM ranges are NOT interval-interpreted; only exact versions[] count.
	e := entry("GO-6", "github.com/foo/bar", []string{"1.0.0"},
		[]rng{{Type: "ECOSYSTEM", Events: []rngEvent{{Introduced: "0"}, {Fixed: "9.9.9"}}}}, nil)
	if _, ok := matchEntry(e, "github.com/foo/bar", "1.0.0"); !ok {
		t.Error("1.0.0 in versions list must match despite ECOSYSTEM range")
	}
	if _, ok := matchEntry(e, "github.com/foo/bar", "2.0.0"); ok {
		t.Error("2.0.0 must not match: ECOSYSTEM ranges are not interval-walked")
	}
}

func TestMatchWrongPackageOrEcosystem(t *testing.T) {
	e := entry("GO-4", "github.com/foo/bar", []string{"1.0.0"}, nil, nil)
	if _, ok := matchEntry(e, "github.com/other/pkg", "1.0.0"); ok {
		t.Error("different package must not match")
	}
	e.Affected[0].Package.Ecosystem = "npm"
	if _, ok := matchEntry(e, "github.com/foo/bar", "1.0.0"); ok {
		t.Error("non-Go ecosystem must not match")
	}
}
