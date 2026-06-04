package scan

// osvEntry is the subset of an OSV record (https://ossf.github.io/osv-schema/)
// that omni's matcher consumes. Entries are byte-passthrough from upstream:
// unknown fields are tolerated (encoding/json ignores them) so future OSV minor
// versions do not break scan.
type osvEntry struct {
	ID       string        `json:"id"`
	Summary  string        `json:"summary"`
	Details  string        `json:"details"`
	Modified string        `json:"modified"`
	Severity []rawSeverity `json:"severity"`
	Affected []osvAffected `json:"affected"`
}

// osvAffected describes one affected package and the version ranges/versions
// that are vulnerable. An affected-level severity, when present, overrides the
// top-level entry severity.
type osvAffected struct {
	Package           osvPackage       `json:"package"`
	Severity          []rawSeverity    `json:"severity"`
	Ranges            []rng            `json:"ranges"`
	Versions          []string         `json:"versions"`
	EcosystemSpecific osvEcosystemSpec `json:"ecosystem_specific"`
}

// osvPackage identifies the affected package within its ecosystem.
type osvPackage struct {
	Ecosystem string `json:"ecosystem"`
	Name      string `json:"name"`
	PURL      string `json:"purl"`
}

// rng is a version range. Only Type=="SEMVER" ranges are interval-walked;
// "ECOSYSTEM" and "GIT" ranges fall back to exact versions[] membership.
type rng struct {
	Type   string     `json:"type"`
	Events []rngEvent `json:"events"`
}

// rngEvent is one boundary in a range timeline. Introduced opens an interval
// (empty or "0" means genesis); Fixed closes it exclusively ([introduced, fixed));
// LastAffected closes it inclusively ([introduced, last_affected]).
type rngEvent struct {
	Introduced   string `json:"introduced"`
	Fixed        string `json:"fixed"`
	LastAffected string `json:"last_affected"`
}

// osvEcosystemSpec carries Go-vuln symbol data used only by the reachability
// path (deferred per ADR-0008); unused by the v1.0 version-range matcher.
type osvEcosystemSpec struct {
	Imports []osvImport `json:"imports"`
}

// osvImport is one imported package plus the vulnerable symbols within it.
type osvImport struct {
	Path    string   `json:"path"`
	Symbols []string `json:"symbols"`
}
