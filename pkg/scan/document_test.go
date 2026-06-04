package scan

import "testing"

func TestScanGatesOnFailOn(t *testing.T) {
	// two components: one vulnerable (high), one clean.
	comps := []component{
		{Pkg: "github.com/foo/bar", Version: "1.0.0"},
		{Pkg: "github.com/safe/lib", Version: "2.0.0"},
	}
	db := &DB{byPkg: map[string][]osvEntry{
		"github.com/foo/bar": {entry("GO-1", "github.com/foo/bar", nil,
			[]rng{{Type: "SEMVER", Events: []rngEvent{{Introduced: "0"}, {Fixed: "1.2.3"}}}},
			[]rawSeverity{{Type: "CVSS_V3", Score: "7.5"}})},
	}}

	rep, err := scanComponents(comps, db, Options{})
	if err != nil {
		t.Fatalf("scanComponents(no gate): %v", err)
	}
	if len(rep.Findings) != 1 || rep.Findings[0].ID != "GO-1" {
		t.Fatalf("findings = %+v, want 1 GO-1", rep.Findings)
	}
	if rep.Scanned != 2 {
		t.Errorf("Scanned = %d, want 2", rep.Scanned)
	}

	// gate at high -> the high finding trips ErrConflict.
	if _, err := scanComponents(comps, db, Options{FailOn: SeverityHigh}); err == nil {
		t.Error("FailOn high must return ErrConflict for a high finding")
	}
	// gate at critical -> high finding is below threshold -> no error.
	if _, err := scanComponents(comps, db, Options{FailOn: SeverityCritical}); err != nil {
		t.Errorf("FailOn critical with only a high finding = %v, want nil", err)
	}
}

func TestScanDeterministicOrder(t *testing.T) {
	comps := []component{{Pkg: "github.com/foo/bar", Version: "1.0.0"}}
	db := &DB{byPkg: map[string][]osvEntry{
		"github.com/foo/bar": {
			entry("GO-9", "github.com/foo/bar", []string{"1.0.0"}, nil, nil),
			entry("GO-1", "github.com/foo/bar", []string{"1.0.0"}, nil, nil),
		},
	}}
	rep, _ := scanComponents(comps, db, Options{})
	if len(rep.Findings) != 2 || rep.Findings[0].ID != "GO-1" || rep.Findings[1].ID != "GO-9" {
		t.Fatalf("findings not sorted by (pkg,id): %+v", rep.Findings)
	}
}
