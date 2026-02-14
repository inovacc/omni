package arch

import (
	"fmt"
	"io"
	"runtime"

	"github.com/inovacc/omni/internal/cli/output"
	"github.com/inovacc/omni/internal/cli/uname"
)

// ArchOptions configures the arch command behavior
type ArchOptions struct {
	OutputFormat output.Format // output format
}

// ArchResult represents arch output for JSON
type ArchResult struct {
	Architecture string `json:"architecture"`
	GOARCH       string `json:"goarch"`
}

// RunArch prints machine architecture
func RunArch(w io.Writer, opts ArchOptions) error {
	arch := uname.MapMachine(runtime.GOARCH)

	f := output.New(w, opts.OutputFormat)
	if f.IsJSON() {
		return f.Print(ArchResult{Architecture: arch, GOARCH: runtime.GOARCH})
	}

	_, _ = fmt.Fprintln(w, arch)

	return nil
}
