package arch

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime"

	"github.com/inovacc/omni/internal/cli/uname"
)

// ArchOptions configures the arch command behavior
type ArchOptions struct {
	JSON bool // --json: output as JSON
}

// ArchResult represents arch output for JSON
type ArchResult struct {
	Architecture string `json:"architecture"`
	GOARCH       string `json:"goarch"`
}

// RunArch prints machine architecture
func RunArch(w io.Writer, opts ArchOptions) error {
	arch := uname.MapMachine(runtime.GOARCH)

	if opts.JSON {
		return json.NewEncoder(w).Encode(ArchResult{Architecture: arch, GOARCH: runtime.GOARCH})
	}

	_, _ = fmt.Fprintln(w, arch)

	return nil
}
