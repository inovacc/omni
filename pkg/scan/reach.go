package scan

import "github.com/inovacc/omni/internal/cli/cmderr"

// ScanSource performs reachability-aware scanning of a Go source tree — reporting
// only vulnerabilities whose vulnerable symbol is actually called.
//
// DEFERRED (ADR-0008): reachability requires golang.org/x/vuln, which (a) execs
// `go list` via golang.org/x/tools/go/packages (violating omni's no-exec rule) and
// (b) pulls golang.org/x/vuln + golang.org/x/tools into the main go.mod via MVS even
// behind a build tag (violating the lean-go.mod rule, ADR-0007). It is therefore
// unavailable in v1.0 and returns ErrUnsupported. The future home is a self-contained
// contrib/govulncheck-scan module (see docs/BACKLOG.md). The params are accepted for
// signature stability with that future implementation.
func ScanSource(_ string, _ *DB, _ Options) (Report, error) {
	return Report{}, cmderr.Wrap(cmderr.ErrUnsupported,
		"omni scan source (reachability) is not available in this build; see docs/BACKLOG.md (deferred per ADR-0008)")
}
