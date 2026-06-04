package scan

// Severity is an ordered vulnerability severity label.
type Severity int

const (
	// SeverityUnknown means no usable CVSS data was found; it is always reported
	// but sorts below SeverityLow for --fail-on gating.
	SeverityUnknown Severity = iota
	// SeverityNone is the CVSS "none" band (base score 0.0).
	SeverityNone
	// SeverityLow is the CVSS "low" band (base score 0.1–3.9).
	SeverityLow
	// SeverityMedium is the CVSS "medium" band (base score 4.0–6.9).
	SeverityMedium
	// SeverityHigh is the CVSS "high" band (base score 7.0–8.9).
	SeverityHigh
	// SeverityCritical is the CVSS "critical" band (base score 9.0–10.0).
	SeverityCritical
)

// String returns the lowercase label (e.g. "high").
func (s Severity) String() string {
	switch s {
	case SeverityNone:
		return "none"
	case SeverityLow:
		return "low"
	case SeverityMedium:
		return "medium"
	case SeverityHigh:
		return "high"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// ParseSeverity maps a label to a Severity; unknown text yields (SeverityUnknown, false).
func ParseSeverity(label string) (Severity, bool) {
	switch label {
	case "none":
		return SeverityNone, true
	case "low":
		return SeverityLow, true
	case "medium":
		return SeverityMedium, true
	case "high":
		return SeverityHigh, true
	case "critical":
		return SeverityCritical, true
	default:
		return SeverityUnknown, false
	}
}

// Finding is one vulnerability matched against one component.
type Finding struct {
	ID           string `json:"id"`                  // OSV id, e.g. GO-2023-1234
	Package      string `json:"package"`             // Go module path
	Version      string `json:"version"`             // affected version present in the SBOM
	FixedVersion string `json:"fixed_version"`       // smallest fixed > Version, or "" if none
	Severity     string `json:"severity"`            // Severity.String()
	Summary      string `json:"summary"`             // OSV summary
	Reachable    *bool  `json:"reachable,omitempty"` // set only by the reachability path
}

// Report is the full scan result, sorted deterministically by (Package, ID).
type Report struct {
	Findings []Finding `json:"findings"`
	Scanned  int       `json:"scanned"` // components examined
	DBAge    string    `json:"db_age"`  // human duration since manifest.generated
}
