package scan

import (
	"errors"
	"testing"

	"github.com/inovacc/omni/internal/cli/cmderr"
)

// TestScanFailOnTripsOnUnknownSeverity guards the
// verifysoundness/scan-failon-unknown-severity-bypass finding (HARDENING
// 2026-06-04, HIGH): a matched vulnerability whose OSV record carries no usable
// CVSS data ("unknown") must NOT silently pass the --fail-on CI gate. A large
// fraction of Go advisories ship without a CVSS vector, so a fail-open gate ships
// known-vulnerable deps with exit 0.
func TestScanFailOnTripsOnUnknownSeverity(t *testing.T) {
	comps := []component{{Pkg: "github.com/foo/bar", Version: "1.0.0"}}
	// versions match, NO severity[] -> severityLabel == SeverityUnknown -> "unknown".
	db := &DB{byPkg: map[string][]osvEntry{
		"github.com/foo/bar": {entry("GO-UNK", "github.com/foo/bar", []string{"1.0.0"}, nil, nil)},
	}}

	rep, err := scanComponents(comps, db, Options{})
	if err != nil {
		t.Fatalf("scanComponents(no gate): %v", err)
	}
	if len(rep.Findings) != 1 || rep.Findings[0].Severity != "unknown" {
		t.Fatalf("setup: want 1 unknown-severity finding, got %+v", rep.Findings)
	}

	// An active gate at ANY level must fail closed on the unscored-but-matched vuln.
	for _, level := range []Severity{SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical} {
		_, gateErr := scanComponents(comps, db, Options{FailOn: level})
		if !errors.Is(gateErr, cmderr.ErrConflict) {
			t.Errorf("FailOn %s with an unknown-severity finding = %v, want cmderr.ErrConflict (fail-closed)", level, gateErr)
		}
	}

	// No gate (FailOn unset) must still NOT error — unknown findings are reported, not gated.
	if _, gateErr := scanComponents(comps, db, Options{}); gateErr != nil {
		t.Errorf("no gate must not error on unknown finding, got %v", gateErr)
	}
}
