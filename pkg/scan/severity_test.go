package scan

import "testing"

func TestSeverityLabelBands(t *testing.T) {
	cases := []struct {
		name string
		sev  []rawSeverity
		want Severity
	}{
		{"critical numeric", []rawSeverity{{Type: "CVSS_V3", Score: "9.8"}}, SeverityCritical},
		{"high vector v3", []rawSeverity{{Type: "CVSS_V3", Score: "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H"}}, SeverityCritical},
		{"medium numeric", []rawSeverity{{Type: "CVSS_V3", Score: "5.3"}}, SeverityMedium},
		{"low numeric", []rawSeverity{{Type: "CVSS_V3", Score: "3.1"}}, SeverityLow},
		{"none numeric", []rawSeverity{{Type: "CVSS_V3", Score: "0.0"}}, SeverityNone},
		{"prefers v4 over v2", []rawSeverity{{Type: "CVSS_V2", Score: "5.0"}, {Type: "CVSS_V4", Score: "9.3"}}, SeverityCritical},
		{"empty -> unknown", nil, SeverityUnknown},
		{"unparseable -> unknown", []rawSeverity{{Type: "Ubuntu", Score: "high"}}, SeverityUnknown},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := severityLabel(c.sev); got != c.want {
				t.Errorf("severityLabel(%v) = %v, want %v", c.sev, got, c.want)
			}
		})
	}
}

func TestSeverityOrderingForGating(t *testing.T) {
	if !(SeverityUnknown < SeverityLow && SeverityLow < SeverityMedium &&
		SeverityMedium < SeverityHigh && SeverityHigh < SeverityCritical) {
		t.Fatal("severity ordering broken — --fail-on gating depends on it")
	}
}
